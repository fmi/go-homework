package main

import (
	"container/heap"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

var BIG_BANG = time.Unix(0, 0)

type Elem struct {
	Key string
	Val interface{}
}

type EpireTime struct {
	Key     string
	Expires time.Time
}

type ExpireHeap []EpireTime

func (h ExpireHeap) Len() int {
	return len(h)
}

func (h ExpireHeap) Less(i, j int) bool {
	return h[i].Expires.Before(h[j].Expires)
}

func (h ExpireHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *ExpireHeap) Push(x interface{}) {
	*h = append(*h, x.(EpireTime))
}

func (h *ExpireHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

type ExpireMap struct {
	stopChan chan struct{}
	wg       sync.WaitGroup

	setRequest           chan *Elem
	getRequest           chan string
	getResponse          chan *interface{}
	deleteRequest        chan string
	containsRequest      chan string
	containsResponse     chan bool
	cleanupRequest       chan struct{}
	sizeRequest          chan struct{}
	sizeResponse         chan int
	atomicRequestInt     chan *Elem
	atomicResponseInt    chan error
	atomicRequestString  chan *Elem
	atomicResponseString chan error

	newExpireTime         chan *Elem
	expiresRequest        chan string
	expiresResponse       chan time.Time
	removeExpiresRequest  chan string
	cleanupExpiresRequest chan struct{}

	expiresChan chan string
}

func NewExpireMap() (em *ExpireMap) {
	em = &ExpireMap{}

	em.stopChan = make(chan struct{})
	em.setRequest = make(chan *Elem)
	em.getRequest = make(chan string)
	em.getResponse = make(chan *interface{})
	em.deleteRequest = make(chan string)
	em.containsRequest = make(chan string)
	em.containsResponse = make(chan bool)
	em.cleanupRequest = make(chan struct{})
	em.sizeRequest = make(chan struct{})
	em.sizeResponse = make(chan int)
	em.atomicRequestInt = make(chan *Elem)
	em.atomicResponseInt = make(chan error)
	em.atomicRequestString = make(chan *Elem)
	em.atomicResponseString = make(chan error)

	em.newExpireTime = make(chan *Elem)
	em.expiresRequest = make(chan string)
	em.expiresResponse = make(chan time.Time)
	em.removeExpiresRequest = make(chan string)
	em.cleanupExpiresRequest = make(chan struct{})

	em.expiresChan = make(chan string)

	em.wg.Add(1)
	go em.storageHandler()
	em.wg.Add(1)
	go em.expiresHandler()

	return
}

func (em *ExpireMap) storageHandler() {
	defer em.wg.Done()
	cache := make(map[string]interface{})

	for {
		select {
		case <-em.stopChan:
			return

		case key := <-em.getRequest:
			found, ok := cache[key]
			if !ok {
				em.getResponse <- nil
			} else {
				em.getResponse <- &found
			}

		case elem := <-em.setRequest:
			cache[elem.Key] = elem.Val

		case <-em.cleanupRequest:
			cache = make(map[string]interface{})

		case key := <-em.containsRequest:
			_, ok := cache[key]
			em.containsResponse <- ok

		case key := <-em.deleteRequest:
			delete(cache, key)

		case <-em.sizeRequest:
			em.sizeResponse <- len(cache)

		case elem := <-em.atomicRequestInt:

			do, ok := elem.Val.(func(i int) int)

			if !ok {
				em.atomicResponseInt <- fmt.Errorf("Atomic request with wrong func")
				continue
			}

			num, found := cache[elem.Key]

			if !found {
				em.atomicResponseInt <- fmt.Errorf("No such key: %s", elem.Key)
				continue
			}

			switch num.(type) {
			case int:
				cache[elem.Key] = do(num.(int))
			case string:
				val, err := strconv.Atoi(num.(string))
				if err != nil {
					em.atomicResponseInt <- err
					continue
				}
				cache[elem.Key] = fmt.Sprintf("%d", do(val))
			default:
				em.atomicResponseInt <- fmt.Errorf("I dont know how to " +
					"increment/decrement this type")
				continue
			}

			em.atomicResponseInt <- nil

		case elem := <-em.atomicRequestString:

			do, ok := elem.Val.(func(i string) string)

			if !ok {
				em.atomicResponseString <- fmt.Errorf("Atomic request with wrong func")
				continue
			}

			str, found := cache[elem.Key]

			if !found {
				em.atomicResponseString <- fmt.Errorf("No such key: %s", elem.Key)
				continue
			}

			switch str.(type) {
			case string:
				cache[elem.Key] = do(str.(string))
			default:
				em.atomicResponseString <- fmt.Errorf("I dont know what to " +
					"do with this type")
				continue
			}

			em.atomicResponseString <- nil

		}
	}

	fmt.Println("Storage handler expited abnormally")
}

func (em *ExpireMap) expiresHandler() {
	defer em.wg.Done()

	expiresDict := make(map[string]time.Time)
	expires := &ExpireHeap{}
	heap.Init(expires)

	const MaxUint64 = ^uint64(0)
	const MaxInt64 = int64(MaxUint64 >> 1)

	for {

		var nextExpire *EpireTime = nil
		nextExpireDuration := time.Unix(MaxInt64, 0).Sub(time.Now())

		if expires.Len() > 0 {
			nextExpire = &((*expires)[0])
			nextExpireDuration = nextExpire.Expires.Sub(time.Now())
		}

		select {
		case <-em.stopChan:
			return

		case elem := <-em.newExpireTime:
			expireTime, _ := elem.Val.(time.Time)
			heap.Push(expires, EpireTime{Key: elem.Key, Expires: expireTime})
			expiresDict[elem.Key] = expireTime

		case key := <-em.expiresRequest:
			val, ok := expiresDict[key]
			if ok {
				em.expiresResponse <- val
			} else {
				em.expiresResponse <- BIG_BANG
			}

		case key := <-em.removeExpiresRequest:
			for index, elem := range []EpireTime(*expires) {
				if key != elem.Key {
					continue
				}

				heap.Remove(expires, index)
				delete(expiresDict, elem.Key)
				break
			}

		case <-em.cleanupExpiresRequest:
			expiresDict = make(map[string]time.Time)
			expires = &ExpireHeap{}
			heap.Init(expires)

		case <-time.After(nextExpireDuration):
			if nextExpire == nil {
				continue
			}
			em.deleteRequest <- nextExpire.Key
			delete(expiresDict, nextExpire.Key)

			select {
			case em.expiresChan <- nextExpire.Key:
				//ok
			default:
				//no one is reading
			}

			heap.Remove(expires, 0)
		}
	}
}

func (em *ExpireMap) Set(key string, value interface{}, expire time.Duration) {
	em.newExpireTime <- &Elem{Key: key, Val: time.Now().Add(expire)}
	em.setRequest <- &Elem{Key: key, Val: value}
}

func (em *ExpireMap) Get(key string) (interface{}, bool) {
	em.getRequest <- key
	val := <-em.getResponse
	if val == nil {
		return nil, false
	}
	return *val, true
}

func (em *ExpireMap) GetString(key string) (string, bool) {

	if v, found := em.Get(key); found {
		switch val := v.(type) {
		case string:
			return val, true
		default:
			return "", false
		}
	}

	return "", false
}

func (em *ExpireMap) GetBool(key string) (bool, bool) {

	if v, found := em.Get(key); found {
		switch val := v.(type) {
		case bool:
			return val, true
		default:
			return false, false
		}
	}

	return false, false
}

func (em *ExpireMap) GetFloat64(key string) (float64, bool) {

	if v, found := em.Get(key); found {
		switch val := v.(type) {
		case float64:
			return val, true
		default:
			return 0.0, false
		}
	}

	return 0.0, false
}

func (em *ExpireMap) GetInt(key string) (int, bool) {
	if v, found := em.Get(key); found {
		switch val := v.(type) {
		case int:
			return val, true
		default:
			return 0, false
		}
	}

	return 0, false
}

func (em *ExpireMap) Delete(key string) {
	em.removeExpiresRequest <- key
	em.deleteRequest <- key
}

func (em *ExpireMap) Contains(key string) bool {
	em.containsRequest <- key
	return <-em.containsResponse
}

func (em *ExpireMap) Size() int {
	em.sizeRequest <- struct{}{}
	return <-em.sizeResponse
}

func (em *ExpireMap) Cleanup() {
	em.cleanupRequest <- struct{}{}
	em.cleanupExpiresRequest <- struct{}{}
}

func (em *ExpireMap) Increment(key string) error {
	em.atomicRequestInt <- &Elem{Key: key, Val: func(i int) int {
		return i + 1
	}}
	return <-em.atomicResponseInt
}

func (em *ExpireMap) Decrement(key string) error {
	em.atomicRequestInt <- &Elem{Key: key, Val: func(i int) int {
		return i - 1
	}}
	return <-em.atomicResponseInt
}

func (em *ExpireMap) ExpiredChan() <-chan string {
	return em.expiresChan
}

func (em *ExpireMap) Expires(key string) (time.Time, bool) {
	em.expiresRequest <- key
	val := <-em.expiresResponse
	if val.Equal(BIG_BANG) {
		return val, false
	}
	return val, true
}

func (em *ExpireMap) ToUpper(key string) error {
	em.atomicRequestString <- &Elem{Key: key, Val: func(str string) string {
		return strings.ToUpper(str)
	}}
	return <-em.atomicResponseString
}

func (em *ExpireMap) ToLower(key string) error {
	em.atomicRequestString <- &Elem{Key: key, Val: func(str string) string {
		return strings.ToLower(str)
	}}
	return <-em.atomicResponseString
}

func (em *ExpireMap) Destroy() {

	close(em.stopChan)
	em.wg.Wait()

	close(em.expiresChan)
	close(em.setRequest)
	close(em.getRequest)
	close(em.getResponse)
	close(em.deleteRequest)
	close(em.containsRequest)
	close(em.containsResponse)
	close(em.cleanupRequest)
	close(em.sizeRequest)
	close(em.sizeResponse)
	close(em.newExpireTime)
	close(em.expiresRequest)
	close(em.expiresResponse)
	close(em.removeExpiresRequest)
	close(em.cleanupExpiresRequest)
	close(em.atomicRequestInt)
	close(em.atomicResponseInt)
	close(em.atomicRequestString)
	close(em.atomicResponseString)
}
