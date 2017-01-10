package main

import (
	"bytes"
	"container/heap"
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"sync"
)

var errNoValidUrls = errors.New("no valid urls")
var errHostReturnedNon206 = errors.New("host returned non 206")

type errorReader [1]error

func newErrorReader(err error) errorReader             { return errorReader([1]error{err}) }
func (r errorReader) Read(b []byte) (n int, err error) { return 0, r[0] }
func (r errorReader) Close() (err error)               { return r[0] }

func head(ctx context.Context, urlString string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodHead, urlString, nil)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req.WithContext(ctx))
}

func get(ctx context.Context, urlString, rangeHeader string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, urlString, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Range", rangeHeader)
	return http.DefaultClient.Do(req.WithContext(ctx))
}

type pieceStruct struct {
	start, end int
	isNext     chan struct{}
	resultChan chan pieceResult
}

func (p pieceStruct) size() int      { return p.end - p.start + 1 }
func (p pieceStruct) String() string { return p.toRange() }

func (p pieceStruct) toRange() string {
	return "bytes=" + strconv.Itoa(p.start) + "-" + strconv.Itoa(p.end)
}

func generatePieces(pieceSize, length int) []*pieceStruct {
	var size = length / pieceSize
	if length%pieceSize != 0 {
		size++
	}
	var result = make([]*pieceStruct, size)
	for i := 0; i < size; i++ {
		result[i] = &pieceStruct{
			start:      i * pieceSize,
			end:        i*pieceSize + pieceSize - 1,
			resultChan: make(chan pieceResult),
			isNext:     make(chan struct{}),
		}
	}
	result[size-1].end = length - 1
	return result
}

func sendPieces(urlSource *urlMuxer, pieces []*pieceStruct) {
	for _, piece := range pieces {
		urlSource.queueForURL(urlSource.getURL(piece.size()), piece)
	}
}

type readComposer struct {
	ctx             context.Context
	cancelFunc      context.CancelFunc
	size, readSoFar int
	currentReader   io.ReadCloser
	currentlyRead   int // read from the current reader
	lastError       error
	ch              chan io.ReadCloser
	chR             chan readResult
}

func newReadComposer(ctx context.Context) *readComposer {
	ctx, cancelFunc := context.WithCancel(ctx)
	return &readComposer{
		ctx:        ctx,
		cancelFunc: cancelFunc,
		ch:         make(chan io.ReadCloser),
		chR:        make(chan readResult),
	}
}

func (r *readComposer) loop(u *urlMuxer, size int) {
	defer r.cancelFunc()
	defer u.stop()
	var urlCount = u.len()
	var pieceLength = size / urlCount
	if size%urlCount != 0 {
		pieceLength++
	}
	r.size = size
	var pieces = generatePieces(pieceLength, r.size)
	close(pieces[0].isNext)
	go sendPieces(u, pieces)
	for {
		select {
		case <-r.ctx.Done():
			r.sendErrorReader(newErrorReader(r.ctx.Err()))
			return
		case result, ok := <-pieces[0].resultChan:
			if !ok {
				// maybe the urls are down to none
				var urlCount = u.len()
				if urlCount == 0 {
					r.sendErrorReader(newErrorReader(errNoValidUrls))
					return
				}
				// redistribute the piece
				var newPieces = r.breakIntoPieces(pieces[0], urlCount)
				pieces = append(newPieces, pieces[1:]...)
				close(pieces[0].isNext)
				go sendPieces(u, newPieces)
				continue
			}
			r.ch <- result.reader
			rResult := <-r.chR
			result.markAsDone()
			if rResult.n < pieces[0].size() {
				// Read less than required ... gonna have to try again
				pieces[0].start += rResult.n
				u.queueForURL(result.from, pieces[0])
				continue
			}

			pieces = pieces[1:]
			if len(pieces) == 0 {
				return // Done
			}
			close(pieces[0].isNext)
		}
	}
}

type readResult struct {
	n   int
	err error
}

func (r *readComposer) breakIntoPieces(p *pieceStruct, count int) []*pieceStruct {
	var newPieceSize = p.size() / count
	if p.size()%count != 0 {
		newPieceSize++
	}
	var pieces = generatePieces(newPieceSize, p.size())
	var resultPieces = make([]*pieceStruct, len(pieces))
	for i, newP := range pieces {
		newP.start += p.start
		newP.end += p.start
		resultPieces[i] = newP
	}
	return resultPieces
}

type pieceResult struct {
	from   string
	reader io.ReadCloser
	done   chan struct{}
}

func (p pieceResult) markAsDone() {
	if p.reader != nil {
		_ = p.reader.Close()
	}
	if p.done != nil {
		close(p.done)
	}
}

func (r *readComposer) Read(b []byte) (n int, err error) {
	if r.currentReader == nil {
		var ok bool
		if r.currentReader, ok = <-r.ch; !ok {
			return 0, r.lastError
		}
		r.currentlyRead = 0
	}
	n, err = r.currentReader.Read(b)
	r.readSoFar += n
	r.currentlyRead += n
	if err != nil {
		r.lastError = err
		r.chR <- readResult{n: r.currentlyRead, err: err}
		r.currentReader = nil
	}
	if (err == io.EOF || err == io.ErrUnexpectedEOF) && r.readSoFar < r.size {
		err = nil
	}
	return n, err
}

func (r *readComposer) sendErrorReader(reader io.ReadCloser) {
	r.ch <- reader
	close(r.ch)
	<-r.chR
}

// DownloadFile download file from multiple urls
func DownloadFile(ctx context.Context, urls []string) io.Reader {
	if ctx == nil {
		ctx = context.Background()
	}

	var maxConnections = len(urls)

	if v := ctx.Value("max-connections"); v != nil {
		maxConnections = v.(int)
	}

	var reader = newReadComposer(ctx)

	go func() {
		var contentLength = -1
		for index, u := range urls {
			resp, err := head(ctx, u)
			if err != nil {
				continue
			}
			contentLength, err = strconv.Atoi(resp.Header.Get("Content-Length"))
			if err != nil {
				continue
			}
			urls = urls[index:]
			break
		}

		if contentLength == -1 {
			reader.sendErrorReader(newErrorReader(errNoValidUrls))
			return
		}

		if contentLength == 0 {
			reader.sendErrorReader(newErrorReader(io.EOF))
			return
		}

		var urlSource = newURLMuxer(maxConnections, urls)
		go reader.loop(urlSource, contentLength)
	}()
	return reader
}

func downloadRoutine(ctx context.Context, urlSource *urlMuxer, sem chan struct{}, pieces <-chan *pieceStruct, url string, priorityCh chan chan struct{}) {
	var (
		queue            pieceHeap                                     // queue for already received requests with lower priority
		downloadedPieces = make(map[*pieceStruct]*bufferedPieceReader) // map of the responses to already received requests with lower priority
		piece            *pieceStruct
		ok               bool
	)
	defer func() {
		// signal that the pieces were not downloaded
		for queue.Len() > 0 {
			piece := queue.Pop().(*pieceStruct)
			close(piece.resultChan)
		}
		for piece := range pieces {
			close(piece.resultChan)
		}
	}()

	for {
		// check queue
		if len(queue) > 0 {
			piece = queue.Pop().(*pieceStruct)
		} else {
			// get new piece if queue is empty
			if piece, ok = <-pieces; !ok {
				return
			}
		}

		for {
			select {
			case sem <- struct{}{}: // take a token
			case <-piece.isNext:
				// priority order take someone place
				select {
				case sem <- struct{}{}: // double checking that sem is full
				default:
					ch := make(chan struct{})
					select {
					case sem <- struct{}{}:
					case priorityCh <- ch:
						// Another thread will give us priority
						<-ch // got it
					}
				}
			case newPiece, ok := <-pieces:
				if !ok { // pieces was closed we are done
					return
				}
				if newPiece.start > piece.start {
					// if the piece is not before the current one we should first finish up
					queue.Add(newPiece)
					continue
				}
				// We have a piece that should be served before the one we have already started
				// We could kill the connection and try again later for the same piece but the
				// tests aren't ready for this so instead we are gonna buffer the whole response
				queue.Add(piece)
				piece = newPiece
				continue
			}
			// break has meaning in select, but continue doesn't so we continue when we don't
			// want to break and get to this break otherwise
			break
		}

		var (
			result = pieceResult{from: url, done: make(chan struct{})}
			err    error
		)
		if bufferedReader, ok := downloadedPieces[piece]; ok {
			result.reader = bufferedReader
			delete(downloadedPieces, piece)
		} else {
			var resp *http.Response
			resp, err = get(ctx, url, piece.toRange())
			if err == nil {
				// we can't handle anything else
				if resp.StatusCode != 206 {
					err = errHostReturnedNon206
					_ = resp.Body.Close()
				} else {
					result.reader = resp.Body
				}
			}
		}

		if err != nil { // we are done
			close(piece.resultChan)
			// different goroutine in order to not deadlock with queueForURL which wants to
			// write to pieces channel
			go urlSource.removeURL(url)
			<-sem
			return
		}
		for {
			select {
			case piece.resultChan <- result:
				select {
				case <-result.done:
				case <-ctx.Done():
				}
				<-sem
			case ch := <-priorityCh:
				// the downloader with the first piece wants priority
				downloadedPieces[piece] = bufferBody(result.reader)
				queue.Add(piece)
				close(ch)
			case newPiece, ok := <-pieces:
				if !ok { // pieces was closed we are done
					<-sem
					return
				}
				queue.Add(newPiece)
				if newPiece.start > piece.start {
					// if the piece is not before the current one we should first finish up
					continue
				}
				// We have a piece that should be served before the one we have already started
				// We could kill the connection and try again later for the same piece but the
				// tests aren't ready for this so instead we are gonna buffer the whole response
				// buffer the current piece and add it to the queue
				downloadedPieces[piece] = bufferBody(result.reader)
				queue.Add(piece)
				<-sem
			}
			break
		}
	}
}

type bufferedPieceReader struct {
	b   *bytes.Buffer
	err error
}

func bufferBody(reader io.ReadCloser) *bufferedPieceReader {
	var bpr = new(bufferedPieceReader)
	bpr.b = bytes.NewBuffer(nil)
	_, bpr.err = bpr.b.ReadFrom(reader)
	_ = reader.Close()
	return bpr
}

func (b *bufferedPieceReader) Read(a []byte) (n int, err error) {
	n, err = b.b.Read(a)
	if err == io.EOF && b.err != nil {
		err = b.err
	}
	return
}

func (b bufferedPieceReader) Close() error {
	b.b.Reset()
	return b.err
}

type pieceHeap []*pieceStruct

func (h pieceHeap) Len() int            { return len(h) }
func (h pieceHeap) Less(i, j int) bool  { return h[i].start > h[j].start }
func (h pieceHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *pieceHeap) Push(x interface{}) { *h = append(*h, x.(*pieceStruct)) }
func (h *pieceHeap) Add(x interface{}) {
	h.Push(x)
	heap.Init(h)
}

func (h *pieceHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

type urlHeap []urlHeapElement

type urlHeapElement struct {
	url   string
	bytes int
	ch    chan *pieceStruct
}

func (h urlHeap) Len() int            { return len(h) }
func (h urlHeap) Less(i, j int) bool  { return h[i].bytes < h[j].bytes }
func (h urlHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *urlHeap) Push(x interface{}) { *h = append(*h, x.(urlHeapElement)) }

func (h *urlHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// urlMuxer takes care of the following
// - make worker threads where no more than a certain amounts are actually working
// - log how much amount each url should serve
// - mux piece requests between them based on amount served
type urlMuxer struct {
	sync.Mutex
	heap       urlHeap
	cancelFunc context.CancelFunc
}

func newURLMuxer(maxConnections int, urls []string) *urlMuxer {
	// here we use a different context in order to not have racecondition betweens the original one
	// canceling and no valid urls.
	var (
		ctx, cancelFunc     = context.WithCancel(context.Background())
		sem                 = make(chan struct{}, maxConnections) // handle multiple of everything
		heap                = make(urlHeap, len(urls))
		result              = &urlMuxer{heap: heap, cancelFunc: cancelFunc}
		takePriorityChannel = make(chan chan struct{})
	)
	for i, url := range urls {
		heap[i] = urlHeapElement{url: url, ch: make(chan *pieceStruct)}
		go downloadRoutine(ctx, result, sem, heap[i].ch, url, takePriorityChannel)
	}
	return result
}

func (u *urlMuxer) len() (length int) {
	u.Lock()
	length = len(u.heap)
	u.Unlock()
	return length
}

func (u *urlMuxer) stop() {
	u.Lock()
	for _, element := range u.heap {
		close(element.ch)
	}
	u.heap = nil
	u.cancelFunc()
	u.Unlock()
}

func (u *urlMuxer) getURL(bytes int) string {
	u.Lock()
	if len(u.heap) == 0 {
		u.Unlock()
		return ""
	}
	url := u.heap[0].url
	u.heap[0].bytes += bytes
	heap.Init(&u.heap)
	u.Unlock()
	return url
}

func (u *urlMuxer) queueForURL(url string, piece *pieceStruct) {
	u.Lock()
	for _, element := range u.heap {
		if element.url == url {
			element.ch <- piece
			u.Unlock()
			return
		}
	}
	close(piece.resultChan)

	u.Unlock()
}

func (u *urlMuxer) removeURL(url string) {
	u.Lock()
	for index, element := range u.heap {
		if element.url == url {
			u.heap = append(u.heap[:index], u.heap[index+1:]...)
			heap.Init(&u.heap)
			close(element.ch)
			break
		}
	}
	u.Unlock()
}
