package main

import "testing"

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

func okSetResult(result interface{}, err error) {
}
