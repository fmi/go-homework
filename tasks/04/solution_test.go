package main

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type countintRequest struct {
	*request
	counter uint32
}

func (c *countintRequest) Run() (interface{}, error) {
	atomic.AddUint32(&c.counter, 1)
	return c.request.Run()
}

type request struct {
	id         string
	cacheable  bool
	alreadyRan bool
	run        func() (interface{}, error)
	setResult  func(interface{}, error)
}

func (fr *request) ID() string {
	return fr.id
}

func (fr *request) Run() (interface{}, error) {
	if fr.alreadyRan {
		panic("Run after Run or SetResult")
	}
	fr.alreadyRan = true
	return fr.run()
}

func (fr *request) Cacheable() bool {
	return fr.cacheable
}

func (fr *request) SetResult(result interface{}, err error) {
	if fr.alreadyRan {
		panic("SetResult after Run or SetResult")
	}
	fr.alreadyRan = true
	fr.setResult(result, err)
}

func okSetResult(result interface{}, err error) {
}

func fatalRun(t *testing.T, msg string) func() (interface{}, error) {
	return func() (interface{}, error) {
		t.Fatal(msg)
		return nil, nil
	}
}

func fatalSetResult(t *testing.T, msg string) func(interface{}, error) {
	return func(_ interface{}, _ error) {
		t.Fatal(msg)
	}
}

func TestNewRequesterAndStop(t *testing.T) {
	var requester Requester = NewRequester(10, 10)
	defer requester.Stop()
	if requester == nil {
		t.Errorf("the returned requester is nil")
	}
}

func TestAddRequestRunsARequest(t *testing.T) {
	var requester = NewRequester(10, 10)
	defer requester.Stop()
	var r1Finished = make(chan struct{})
	r1 := &request{
		id:        "a",
		cacheable: true,
		run: func() (interface{}, error) {
			close(r1Finished)
			return "result1", nil
		},
		setResult: okSetResult,
	}
	go requester.AddRequest(r1)
	<-r1Finished
}

func TestNonCacheableRequests(t *testing.T) {
	var requester = NewRequester(10, 10)
	defer requester.Stop()
	var expected1, expected2 = "foo", "bar"
	var setted = make(chan struct{})
	var r1Finished = make(chan struct{})
	r1 := &request{
		id:        "a",
		cacheable: false,
		run: func() (interface{}, error) {
			close(r1Finished)
			return expected1, nil
		},
		setResult: fatalSetResult(t, "Should be Ran, not SetResulted"),
	}
	go requester.AddRequest(r1)
	r2 := &request{
		id:        "a",
		cacheable: false,
		run: func() (interface{}, error) {
			defer close(setted)
			return expected2, nil
		},
		setResult: fatalSetResult(t, "Should be Ran, not SetResulted"),
	}
	<-r1Finished
	go requester.AddRequest(r2)
	<-setted
}

func TestAddRequestRunsOnlyFirstRequest(t *testing.T) {
	var requester = NewRequester(10, 10)
	defer requester.Stop()
	var setted = make(chan struct{})
	var r1Finished = make(chan struct{})
	r1 := &request{
		id:        "a",
		cacheable: true,
		run: func() (interface{}, error) {
			close(r1Finished)
			return "result1", nil
		},
		setResult: fatalSetResult(t, "Should be Ran, not SetResulted"),
	}
	r2 := &request{
		id:        "a",
		cacheable: true,
		run:       fatalRun(t, "Equal request was ran"),
		setResult: func(result interface{}, err error) {
			defer close(setted)
			if result != "result1" {
				t.Errorf("wrong result %s was set for second of two equal requests", result)
			}
		},
	}

	go requester.AddRequest(r1)
	<-r1Finished
	go requester.AddRequest(r2)
	<-setted
}

func TestRequestsAreRunAsyncly(t *testing.T) {
	var requester = NewRequester(10, 2)
	defer requester.Stop()
	var expected1, expected2 = "foo", "bar"
	var r1Started = make(chan struct{})
	var r1Continue = make(chan struct{})
	var r1Finished = make(chan struct{})
	var r2Started = make(chan struct{})
	var r2Continue = make(chan struct{})
	var r2Finished = make(chan struct{})
	r1 := &request{
		id:        "a",
		cacheable: false,
		run: func() (interface{}, error) {
			defer close(r1Finished)
			close(r1Started)
			<-r1Continue
			return expected1, nil
		},
		setResult: fatalSetResult(t, "Should be Ran, not SetResulted"),
	}
	r2 := &request{
		id:        "b",
		cacheable: false,
		run: func() (interface{}, error) {
			defer close(r2Finished)
			close(r2Started)
			<-r2Continue
			return expected2, nil
		},
		setResult: fatalSetResult(t, "Should be Ran, not SetResulted"),
	}
	go requester.AddRequest(r1)
	<-r1Started
	go requester.AddRequest(r2)
	<-r2Started
	close(r2Continue)
	<-r2Finished
	close(r1Continue)
	<-r1Finished

}

func TestNoRunsAfterStop(t *testing.T) {
	var requester = NewRequester(2, 2)
	var r1Started = make(chan struct{})
	var r1Continue = make(chan struct{})
	var stopStarted = make(chan struct{})
	var stopFinished = make(chan struct{})
	r1 := &request{
		id:        "a",
		cacheable: true,
		run: func() (interface{}, error) {
			close(r1Started)
			<-r1Continue
			return "result1", nil
		},
		setResult: fatalSetResult(t, "Should be Ran, not SetResulted"),
	}
	r2 := &request{
		id:        "b",
		cacheable: true,
		run:       fatalRun(t, "request after stop was ran"),
	}

	go requester.AddRequest(r1)
	<-r1Started
	go func() {
		defer close(stopFinished)
		close(stopStarted)
		requester.Stop()
	}()
	<-stopStarted
	time.Sleep(20 * time.Millisecond)
	go requester.AddRequest(r2)
	close(r1Continue)
	<-stopFinished
	time.Sleep(20 * time.Millisecond)
}

func TestCacheSize(t *testing.T) {
	var requester = NewRequester(2, 2)
	defer requester.Stop()
	var finishRun = make(chan struct{})
	var runFinished = make(chan struct{})
	var resultSet = make(chan struct{})
	var resultFunc = func(_ interface{}, err error) {
		resultSet <- struct{}{}
	}
	var runFunc = func(result string) func() (interface{}, error) {
		return func() (interface{}, error) {
			<-finishRun
			defer func() {
				runFinished <- struct{}{}
			}()
			return result, nil
		}
	}
	r1 := &request{
		id:        "a",
		cacheable: true,
		run:       runFunc("result1"),
		setResult: fatalSetResult(t, "Should be Ran, not SetResulted"),
	}
	r2 := &request{
		id:        "b",
		cacheable: true,
		run:       runFunc("result2"),
		setResult: fatalSetResult(t, "Should be Ran, not SetResulted"),
	}
	r3 := &request{
		id:        "c",
		cacheable: true,
		run:       runFunc("result3"),
		setResult: fatalSetResult(t, "Should be Ran, not SetResulted"),
	}
	var (
		cr1 = &countintRequest{request: r1}
		cr2 = &countintRequest{request: r2}
		cr3 = &countintRequest{request: r3}
	)

	go requester.AddRequest(cr1)
	finishRun <- struct{}{}
	<-runFinished

	go requester.AddRequest(cr2)
	finishRun <- struct{}{}
	<-runFinished

	go requester.AddRequest(cr3)
	finishRun <- struct{}{}
	<-runFinished

	cr3.alreadyRan = false
	cr3.run = fatalRun(t, "cr3 should be cached and not runned")
	cr3.setResult = resultFunc
	go requester.AddRequest(cr3)
	<-resultSet

	cr2.alreadyRan = false
	cr2.run = fatalRun(t, "cr2 should be cached and not runned")
	cr2.setResult = resultFunc
	go requester.AddRequest(cr2)
	<-resultSet

	cr1.alreadyRan = false
	go requester.AddRequest(cr1)
	finishRun <- struct{}{}
	<-runFinished
	if cr1.counter != 2 {
		t.Errorf("it was expected %+v to be ran twice but it was %d", cr1, cr1.counter)
	}
	if cr2.counter != 1 {
		t.Errorf("it was expected %+v to be ran twice but it was %d", cr2, cr2.counter)
	}
	if cr3.counter != 1 {
		t.Errorf("it was expected %+v to be ran twice but it was %d", cr3, cr3.counter)
	}
}

func TestThrottleSize(t *testing.T) {
	var requester = NewRequester(2, 2)
	defer requester.Stop()
	var wg sync.WaitGroup
	var resultSet = make(chan struct{})
	var runFunc = func(result string) func() (interface{}, error) {
		return func() (interface{}, error) {
			wg.Done()
			defer wg.Add(1)
			defer func() {
				resultSet <- struct{}{}
			}()
			time.Sleep(20 * time.Millisecond)
			return result, nil
		}
	}

	wg.Add(2)
	r1 := &request{
		id:        "a",
		cacheable: true,
		run:       runFunc("result1"),
		setResult: fatalSetResult(t, "Should be Ran, not SetResulted"),
	}
	r2 := &request{
		id:        "b",
		cacheable: true,
		run:       runFunc("result2"),
		setResult: fatalSetResult(t, "Should be Ran, not SetResulted"),
	}
	r3 := &request{
		id:        "c",
		cacheable: true,
		run:       runFunc("result3"),
		setResult: fatalSetResult(t, "Should be Ran, not SetResulted"),
	}
	go requester.AddRequest(r1)
	go requester.AddRequest(r2)
	go requester.AddRequest(r3)
	<-resultSet
	<-resultSet
	<-resultSet
}

func TestNonCacheableRequestsFast(t *testing.T) {
	var requester = NewRequester(10, 10)
	defer requester.Stop()
	var expected1, expected2 = "foo", "bar"
	var r1Started = make(chan struct{})
	var r1Continue = make(chan struct{})
	var r2Started = make(chan struct{})
	r1 := &request{
		id:        "a",
		cacheable: false,
		run: func() (interface{}, error) {
			close(r1Started)
			<-r1Continue
			return expected1, nil
		},
		setResult: fatalSetResult(t, "Should be Ran, not SetResulted"),
	}
	go requester.AddRequest(r1)
	r2 := &request{
		id:        "a",
		cacheable: false,
		run: func() (interface{}, error) {
			close(r2Started)
			return expected2, nil
		},
		setResult: fatalSetResult(t, "Should be Ran, not SetResulted"),
	}
	<-r1Started
	go requester.AddRequest(r2)
	select {
	case <-r2Started:
		t.Fatalf("second requst has been started :(")
	case <-time.After(50 * time.Millisecond):
	}
	close(r1Continue)
	<-r2Started
}

func TestStopWithQueue(t *testing.T) {
	var requester = NewRequester(1, 1)
	var r1Started = make(chan struct{})
	var r1Continue = make(chan struct{})
	var stopStarted = make(chan struct{})
	var stopFinished = make(chan struct{})
	var setResult = make(chan struct{})
	r1 := &request{ // Should Run
		id:        "a",
		cacheable: true,
		run: func() (interface{}, error) {
			close(r1Started)
			<-r1Continue
			return "result1", nil
		},
		setResult: fatalSetResult(t, "Should be Ran, not SetResulted"),
	}

	r2 := &request{ // Should get result from fr
		id:        "a",
		cacheable: true,
		run:       fatalRun(t, "Should be SetResulted, not Ran"),
		setResult: func(_ interface{}, _ error) {
			close(setResult)
		},
	}

	r3 := &request{ // Should be skipped
		id:        "b",
		cacheable: true,
		run:       fatalRun(t, "Should not be Ran after Stop"),
		setResult: fatalSetResult(t, "Should not be Ran after Stop"),
	}

	r4 := &request{ // Should be skipped
		id:        "a",
		cacheable: true,
		run:       fatalRun(t, "Cached should not be Ran after Stop"),
		setResult: fatalSetResult(t, "Cached should be SetResulted,  after Stop"),
	}

	go requester.AddRequest(r1)
	<-r1Started
	go requester.AddRequest(r2)
	go requester.AddRequest(r3)
	time.Sleep(20 * time.Millisecond)
	go func() {
		close(stopStarted)
		requester.Stop()
		close(stopFinished)
	}()
	<-stopStarted
	time.Sleep(20 * time.Millisecond)
	go requester.AddRequest(r4)
	time.Sleep(20 * time.Millisecond)
	close(r1Continue)
	<-stopFinished
	<-setResult
	time.Sleep(40 * time.Millisecond)
}

func TestStopWaits(t *testing.T) {
	var (
		requester  = NewRequester(1, 1)
		r1Started  = make(chan struct{})
		r1Finished = make(chan struct{})
	)
	r := &request{ // Should Run
		id:        "a",
		cacheable: true,
		run: func() (interface{}, error) {
			close(r1Started)
			time.Sleep(30 * time.Millisecond)
			defer close(r1Finished)
			return "result1", nil
		},
		setResult: fatalSetResult(t, "Should be Ran, not SetResulted"),
	}
	go requester.AddRequest(r)
	<-r1Started
	requester.Stop()
	select {
	case <-r1Finished:
	default:
		t.Fatal("Stop finished before Ran was")
	}
}

func TestNonCacheableRequestsBeingWaitedByStop(t *testing.T) {
	var requester = NewRequester(10, 10)
	var expected1, expected2 = "foo", "bar"
	var r1Started = make(chan struct{})
	var r1Continue = make(chan struct{})
	var r1Finished = make(chan struct{})
	var r2Started = make(chan struct{})
	var r2Continue = make(chan struct{})
	var r2Finished = make(chan struct{})
	var stopFinished = make(chan struct{})
	r1 := &request{
		id:        "a",
		cacheable: false,
		run: func() (interface{}, error) {
			defer close(r1Finished)
			close(r1Started)
			<-r1Continue
			return expected1, nil
		},
		setResult: fatalSetResult(t, "Should be Ran, not SetResulted"),
	}
	go requester.AddRequest(r1)
	<-r1Started
	r2 := &request{
		id:        "a",
		cacheable: false,
		run: func() (interface{}, error) {
			defer close(r2Finished)
			close(r2Started)
			<-r2Continue
			return expected2, nil
		},
		setResult: fatalSetResult(t, "Should be Ran, not SetResulted"),
	}
	go requester.AddRequest(r2)
	close(r1Continue)
	<-r1Finished
	<-r2Started
	go func() {
		defer close(stopFinished)
		requester.Stop()
	}()
	time.Sleep(50 * time.Millisecond)
	select {
	case <-stopFinished:
		t.Fatal("Stop shouldn't have finished before non-cacheable")
	default:
	}
	close(r2Continue)
	<-r2Finished
}

func TestStopWithQueueFromForum(t *testing.T) {
	var (
		requester    = NewRequester(1000, 2)
		r1Started    = make(chan struct{})
		r1Continue   = make(chan struct{})
		r1Finished   = make(chan struct{})
		r2SetResult  = make(chan struct{})
		r4Started    = make(chan struct{})
		r4Continue   = make(chan struct{})
		r4Finished   = make(chan struct{})
		stopStarted  = make(chan struct{})
		stopFinished = make(chan struct{})
	)
	r1 := &request{ // Should Run
		id:        "a",
		cacheable: true,
		run: func() (interface{}, error) {
			close(r1Started)
			<-r1Continue
			defer close(r1Finished)
			return "result1", nil
		},
		setResult: fatalSetResult(t, "r4 should be runned not SetResulted"),
	}

	r2 := &request{
		id:        "a",
		cacheable: true,
		run:       fatalRun(t, "r2 should not be runned as it should be cached"),
		setResult: func(_ interface{}, _ error) {
			close(r2SetResult)
		},
	}

	r3 := &request{
		id:        "a",
		cacheable: false,
		run:       fatalRun(t, "r3 should not be runned as it should be cached"),
		setResult: fatalSetResult(t, "r3 should not be SetResulterd, as it was added after Stop"),
	}

	r4 := &request{
		id:        "b",
		cacheable: false,
		run: func() (interface{}, error) {
			close(r4Started)
			<-r4Continue
			defer close(r4Finished)
			return "result4", nil
		},
		setResult: fatalSetResult(t, "r4 should be runned not SetResulted"),
	}

	r5 := &request{
		id:        "b",
		cacheable: false,
		run:       fatalRun(t, "r5 should've not Run after Stop"),
		setResult: fatalSetResult(t, "r5 should've not been SetResulted as it's non cacheable"),
	}

	r6 := &request{
		id:        "c",
		cacheable: true,
		run:       fatalRun(t, "r6 should not be Run after Stop"),
		setResult: fatalSetResult(t, "r6 should not be SetResulted,  after Stop"),
	}

	r7 := &request{
		id:        "d",
		cacheable: true,
		run:       fatalRun(t, "r7 should not be Run after Stop"),
		setResult: fatalSetResult(t, "r7 should not be SetResulted, after Stop"),
	}

	go requester.AddRequest(r1)
	go requester.AddRequest(r4)
	<-r1Started
	<-r4Started
	go requester.AddRequest(r2)
	go requester.AddRequest(r5)
	time.Sleep(20 * time.Millisecond)
	go func() {
		close(stopStarted)
		requester.Stop()
		close(stopFinished)
	}()
	<-stopStarted
	time.Sleep(20 * time.Millisecond)
	go requester.AddRequest(r6)
	go requester.AddRequest(r3)
	time.Sleep(20 * time.Millisecond)
	close(r1Continue)
	close(r4Continue)
	<-r1Finished
	<-r4Finished
	<-r2SetResult
	go requester.AddRequest(r7)
	time.Sleep(50 * time.Millisecond)
}
