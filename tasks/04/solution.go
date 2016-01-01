package main

import (
	"sync"
)

type Request interface {
	// Връща идентификатор за заявката. Ако две заявки имат еднакви идентификатори
	// то те са "равни".
	ID() string

	// Блокира докато изпълнява заявката.
	// Връща резултата или грешка ако изпълнението е неуспешно.
	// Резултата и грешката не трябва да бъдат подавани на SetResult
	// за текущата заявка - те са запазват вътрешно преди да бъдат върнати.
	Run() (result interface{}, err error)

	// Връща дали заявката е кешируерма.
	// Метода има неопределено поведение ако бъде бъде извикан преди `Run`.
	Cacheable() bool

	// Задава резултата на заявката.
	// Не трябва да се извиква за заявки, за които е бил извикан `Run`.
	SetResult(result interface{}, err error)
}

type Requester interface {
	// Добавя заявка за изпълнение и я изпълнява, ако това е необходимо, при първа възможност.
	AddRequest(request Request)

	// Спира 'Заявчика'. Това означава че изчаква всички заявки да завършат и връща резултати.
	// Нови заявки не трябва да бъдат започвани през това време.
	Stop()
}

func NewRequester(cacheSize int, throttleSize int) Requester {
	i := &impl{
		cb:       newCircularBuffer(cacheSize),
		requests: make(chan Request),
		stop:     make(chan struct{}),
		stopped:  make(chan struct{}),
	}
	go i.loop(throttleSize)
	return i
}

type impl struct {
	cb       *circularBuffer
	requests chan Request
	stop     chan struct{}
	stopped  chan struct{}
}

func (i *impl) AddRequest(request Request) {
	select {
	case <-i.stop:
	case i.requests <- request:
	}
}

func (i *impl) Stop() {
	close(i.stop)
	<-i.stopped
}

func (i *impl) loop(throttleSize int) {
	var (
		workers      = make(chan struct{}, throttleSize)
		finished     = make(chan *requestQueue)
		stopped      = make(chan *requestQueue)
		running      = make(map[string]*requestQueue)
		stopChannel  = i.stop
		countRunning = 0
	)

	defer close(i.stopped)

	var queueRequest = func(rq *requestQueue) {
		select {
		case workers <- struct{}{}:
			defer func() {
				<-workers
			}()
		}
		select {
		case <-i.stop:
			stopped <- rq
			return
		default:
		}

		result, err := rq.request.Run()
		rq.result = resultPair{
			result: result,
			err:    err,
		}
		finished <- rq
	}

	for countRunning != 0 || stopChannel != nil {
		select {
		case request := <-i.requests:
			if stopChannel == nil {
				continue
			}
			if pair, ok := i.cb.get(request.ID()); ok {
				request.SetResult(pair.result, pair.err)
				continue
			}

			if rq, ok := running[request.ID()]; ok {
				rq.append(request)
				continue
			}

			var rq = newRequestQueue(request)
			running[request.ID()] = rq
			countRunning++
			go queueRequest(rq)
		case rq := <-finished:
			countRunning--
			delete(running, rq.request.ID())
			rq.Lock()
			if rq.request.Cacheable() {
				i.cb.set(rq.request.ID(), rq.result)
				for _, request := range rq.sameIdRequests {
					request.SetResult(rq.result.result, rq.result.err)
				}
			} else {
				for _, request := range rq.sameIdRequests {
					countRunning++
					go queueRequest(newRequestQueue(request))
				}
			}
			rq.Unlock()
		case rq := <-stopped:
			countRunning--
			delete(running, rq.request.ID())
		case <-stopChannel:
			stopChannel = nil
		}
	}
}

func newRequestQueue(r Request) *requestQueue {
	return &requestQueue{request: r}
}

type requestQueue struct {
	sync.Mutex
	request        Request
	sameIdRequests []Request
	result         resultPair
}

func (rq *requestQueue) append(req Request) {
	rq.Lock()
	rq.sameIdRequests = append(rq.sameIdRequests, req)
	rq.Unlock()
}

type resultPair struct {
	result interface{}
	err    error
}

// Circular buffer implementation
type circularBuffer struct {
	ids    []string
	index  int
	buffer map[string]resultPair
	mutex  sync.RWMutex
}

func newCircularBuffer(size int) *circularBuffer {
	return &circularBuffer{
		ids:    make([]string, size),
		buffer: make(map[string]resultPair),
	}
}

func (cb *circularBuffer) get(key string) (resultPair, bool) {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	result, ok := cb.buffer[key]
	return result, ok
}

func (cb *circularBuffer) set(key string, result resultPair) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	index := cb.index % len(cb.ids)
	delete(cb.buffer, cb.ids[index])

	cb.ids[index] = key
	cb.buffer[key] = result
	cb.index++
}
