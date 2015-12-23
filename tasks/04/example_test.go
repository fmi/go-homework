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
	fr := &fakeRequest{
		id:        "fakeId",
		cacheable: true,
		run: func() (interface{}, error) {
			close(ran)
			return "result1", nil
		},
		setResult: okSetResult,
	}
	requester.AddRequest(fr)
	<-ran
}

func TestNonCacheableRequests(t *testing.T) {
	var requester = NewRequester(10, 10)
	defer requester.Stop()
	var expected1, expected2 = "foo", "bar"
	var setted = make(chan struct{})
	fr := &fakeRequest{
		id:        "fakeId",
		cacheable: false,
		run: func() (interface{}, error) {
			return expected1, nil
		},
		setResult: nil,
	}
	requester.AddRequest(fr)
	fr = &fakeRequest{
		id:        "fakeId",
		cacheable: false,
		run: func() (interface{}, error) {
			defer close(setted)
			return expected2, nil
		},
		setResult: nil,
	}
	requester.AddRequest(fr)
	<-setted
}

type fakeRequest struct {
	id         string
	cacheable  bool
	alreadyRan bool
	run        func() (interface{}, error)
	setResult  func(interface{}, error)
}

func (fr *fakeRequest) ID() string {
	return fr.id
}

func (fr *fakeRequest) Run() (interface{}, error) {
	if fr.alreadyRan {
		panic("Run after Run or SetResult")
	}
	fr.alreadyRan = true
	return fr.run()
}

func (fr *fakeRequest) Cacheable() bool {
	return fr.cacheable
}

func (fr *fakeRequest) SetResult(result interface{}, err error) {
	if fr.alreadyRan {
		panic("SetResult after Run or SetResult")
	}
	fr.alreadyRan = true
	fr.setResult(result, err)
}
