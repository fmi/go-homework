package main

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// This function accepts chunks of urls through the in channel, buffers them
// and passes them one by one to the out channel
func makeBufferedChan(in <-chan []string, initialSize int, doneSignal chan bool) chan string {
	out := make(chan string)

	go func() {
		defer close(out)
		bufferedUrls := make([]string, 0, initialSize)
		initialSlice := bufferedUrls[:]

		// Get a group of Urls and then try to pass them ony by one to "out" channgel.
		// If more are received meanwhile, add them to the buffer slice :)

		for {
			// Ensure that we have some urls in the buffer
			select {
			case <-doneSignal:
				return
			case chunkOfUrls, ok := <-in:
				if !ok {
					close(doneSignal)
					return
				}
				// Buffer the newly received Urls
				bufferedUrls = append(bufferedUrls, chunkOfUrls...)
			}

		innerloop:
			for {
				select {
				case <-doneSignal:
					return
				case anotherChunkOfUrls, ok := <-in: // More urls are received before they can be processed by the receiver
					if !ok {
						close(doneSignal)
						return
					}
					// Buffer the newly received Urls
					bufferedUrls = append(bufferedUrls, anotherChunkOfUrls...)

				case out <- bufferedUrls[0]: // Receiver consumed the first buffered Url

					bufferedUrls = bufferedUrls[1:]

					// If no more urls are in the buffer, go back to the beginning to fill up the tank
					if len(bufferedUrls) == 0 {
						bufferedUrls = initialSlice[:] // Clumsy way to somewhat limit memory leaks

						break innerloop
					}
				}
			}
		}
	}()

	return out
}

// WebCrawler is used to create concurrent HTTP/HTTPS crawlers, looking for a particular match
type WebCrawler struct {
	Callback       func(string) bool // The function used to test results
	ResultChan     chan string       // This channel is used for passing the result by one of the future workers
	Urls           chan string       // This channel is used buffer and get the coming urls one by one
	RequestTimeout time.Duration     // The timeout for a single http request
	DoneSignal     chan bool         // By closing this channel, you signal all internal goroutines that they should exit. Also, it can be used to detect whether the crawler exited prematurely.
	freeWorkers    chan bool         // Used for limiting the amount of concurrent workers (write: add a worker, read: remove a worker: blocks when there are no more free workers left)
}

// Start starts the crawler. Once started, urls will be processed
// by a new goroutine and distributed to a free worker, if there is one
func (wc *WebCrawler) Start() {
	defer close(wc.freeWorkers)
	for {
		select {
		case wc.freeWorkers <- false: // Wait for a free worker
			// Free worker found, reserve it and wait for future work...
			select {
			case urlToCheck, ok := <-wc.Urls:
				if !ok {
					<-wc.freeWorkers
					return
				}
				// Work was found for the reserved worker! Launch a goroutine to process the url!
				go wc.checkWebsite(urlToCheck)
			case <-wc.DoneSignal: // Exit if crawler is done
				<-wc.freeWorkers
				return
			}
		case <-wc.DoneSignal: // Exit if crawler is done
			return
		}
	}
}

func (wc *WebCrawler) checkWebsite(urlToCheckRaw string) {
	//fmt.Println("Crawlin", urlToCheckRaw, "..")

	// Always free the poor worker after their work is done :)
	defer func() {
		//fmt.Println("Worker for", urlToCheckRaw, "is free!!!")
		<-wc.freeWorkers
	}()

	urlToCheck, err := url.Parse(urlToCheckRaw)
	if err != nil {
		return
	}

	tr := &http.Transport{}
	client := &http.Client{Transport: tr}
	request := &http.Request{URL: urlToCheck}

	httpErrChan := make(chan bool)
	httpSuccessChan := make(chan bool)

	go func() {
		response, err := client.Do(request)
		if err != nil {
			close(httpErrChan)
			return
		}
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			close(httpErrChan)
			return
		}

		responseBody, err := ioutil.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			close(httpErrChan)
			return
		}
		if wc.Callback(string(responseBody)) {
			close(httpSuccessChan)
		} else {
			close(httpErrChan)
		}
	}()

	// Either the request timeouts, returns an error, returns a success
	// or the global done event happens (global timeout or someone else found the result)
	select {
	case <-time.After(wc.RequestTimeout):
		//fmt.Println("Request for", urlToCheckRaw, "timeouts :(")
		tr.CancelRequest(request)
	case <-wc.DoneSignal:
		//fmt.Println("Global done signal reached for", urlToCheckRaw)
		tr.CancelRequest(request)
	case <-httpErrChan:
		//fmt.Println("Error when downloading", urlToCheckRaw)
		// Do nothing
	case <-httpSuccessChan:
		//fmt.Println("THIS IS THE RESULT:", urlToCheckRaw)
		wc.ResultChan <- urlToCheckRaw
	}
}

// MakeWebCrawler instantiates a new web crawler. For convenience, it also provides
// a buffer for all incoming chunks of urls.
func MakeWebCrawler(callback func(string) bool, chunkedUrlsToCheck <-chan []string, workersCount int) (crawler WebCrawler, err error) {
	if callback == nil {
		err = errors.New("You should supply a valid callback!")
		return
	}

	if chunkedUrlsToCheck == nil {
		err = errors.New("You should initialize the channel!")
		return
	}

	if workersCount <= 0 {
		err = errors.New("Workers count was negative or zero!")
		return
	}

	doneSignal := make(chan bool)

	crawler = WebCrawler{
		Callback:       callback,
		ResultChan:     make(chan string),
		DoneSignal:     doneSignal,
		Urls:           makeBufferedChan(chunkedUrlsToCheck, 10/workersCount, doneSignal),
		RequestTimeout: 3 * time.Second,
		freeWorkers:    make(chan bool, workersCount),
	}

	return
}

// SeekAndDestroy concurrently finds the first url in `urls` that returns positive `callback` (with `workerscount` parallel workers)
func SeekAndDestroy(callback func(string) bool, urls <-chan []string, workersCount int) (result string, err error) {

	crawler, err := MakeWebCrawler(callback, urls, workersCount)
	if err != nil {
		return
	}

	go crawler.Start()

	// Everything is set, now we wait for the result or for timeout
	select {
	case <-time.After(15 * time.Second):
		err = errors.New("Global timeout was reached, aborting all operations...")
	case <-crawler.DoneSignal:
		err = errors.New("The crawler exited prematurely, probably because the input urls channel was closed...")
	case result = <-crawler.ResultChan:
		close(crawler.DoneSignal)
	}

	return
}
