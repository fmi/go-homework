package main

import (
	"testing"
	"time"
)

type request struct {
	id         string
	cacheable  bool
	alreadyRan bool
	run        func() (interface{}, error)
	setResult  func(interface{}, error)
}

func (r *request) ID() string {
	return r.id
}

func (r *request) Run() (interface{}, error) {
	if r.alreadyRan {
		panic("Run after Run or SetResult")
	}
	r.alreadyRan = true
	return r.run()
}

func (r *request) Cacheable() bool {
	return r.cacheable
}

func (r *request) SetResult(result interface{}, err error) {
	if r.alreadyRan {
		panic("SetResult after Run or SetResult")
	}
	r.alreadyRan = true
	r.setResult(result, err)
}

func okSetResult(result interface{}, err error) {}

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
	var ran = make(chan struct{})
	r := &request{
		id:        "a",
		cacheable: true,
		run: func() (interface{}, error) {
			close(ran)
			return "result1", nil
		},
		setResult: okSetResult,
	}
	requester.AddRequest(r)
	<-ran
}

func TestNonCacheableRequests(t *testing.T) {
	var requester = NewRequester(10, 10)
	defer requester.Stop()
	var expected1, expected2 = "foo", "bar"
	var setted = make(chan struct{})
	r := &request{
		id:        "a",
		cacheable: false,
		run: func() (interface{}, error) {
			return expected1, nil
		},
		setResult: nil,
	}
	requester.AddRequest(r)
	r = &request{
		id:        "a",
		cacheable: false,
		run: func() (interface{}, error) {
			defer close(setted)
			return expected2, nil
		},
		setResult: nil,
	}
	requester.AddRequest(r)
	<-setted
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
	time.Sleep(100 * time.Millisecond)
	go func() {
		close(stopStarted)
		requester.Stop()
		close(stopFinished)
	}()
	<-stopStarted
	time.Sleep(100 * time.Millisecond)
	go requester.AddRequest(r6)
	go requester.AddRequest(r3)
	time.Sleep(100 * time.Millisecond)
	close(r1Continue)
	close(r4Continue)
	<-r1Finished
	<-r4Finished
	<-r2SetResult
	go requester.AddRequest(r7)
	time.Sleep(200 * time.Millisecond)
}
