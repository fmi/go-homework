package main

import (
	"runtime"
	"testing"
	"runtime"
	"time"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

const DELTA = 10 * time.Millisecond

func TestSettingAndGettingAKey(t *testing.T) {
	mp := NewExpireMap()
	defer mp.Destroy()

	key := "foo"
	expected := "bar"

	mp.Set(key, expected, 15*time.Second)
	found, ok := mp.Get(key)

	if ok == false {
		t.Errorf("Getting the foo key failed")
	}

	if found != expected {
		t.Error("Found key was different than the expected")
	}

	_, ok = mp.Get("not-there")

	if ok {
		t.Error("Found key which shouldn't have been there")
	}
}

func TestGettingExpiredValues(t *testing.T) {
	mp := NewExpireMap()
	defer mp.Destroy()

	duration := 150 * time.Millisecond
	mp.Set("foo", "super-bar", duration)
	mp.Set("bar", "super-foo", 300*time.Millisecond)

	time.Sleep(duration + DELTA)

	if _, ok := mp.GetString("foo"); ok {
		t.Errorf("Getting expired key did not return error")
	}

	expected := "super-foo"
	found, ok := mp.GetString("bar")

	if ok == false {
		t.Errorf("Getting non-expired key returned error")
	}

	if expected != found {
		t.Errorf("Expected %s but found %s", expected, found)
	}
}

func TestCleaningUpTheMap(t *testing.T) {
	mp := NewExpireMap()
	defer mp.Destroy()

	mp.Set("foo", "bar", 15*time.Second)
	mp.Set("number", "794", 15*time.Second)

	mp.Cleanup()

	if mp.Size() != 0 {
		t.Errorf("Cleaning up ExpireMap did not work")
	}
}

func TestTheExampleInReadme(t *testing.T) {
	cache := NewExpireMap()
	defer cache.Destroy()

	cache.Set("foo", "bar", 15*time.Second)
	cache.Set("spam", "4", 25*time.Minute)
	cache.Set("eggs", 9000.01, 3*time.Minute)

	err := cache.Increment("spam")

	if err != nil {
		t.Error("Sadly, incrementing the spam did not succeed")
	}

	if val, ok := cache.Get("spam"); ok {
		if val != "5" {
			t.Error("spam was not 5")
		}
	} else {
		t.Error("No spam. Have some eggs instead?")
	}

	if eggs, ok := cache.GetFloat64("eggs"); !ok || eggs <= 9000 {
		t.Error("We did not have as many eggs as expected.",
			"Have you considered our spam offers?")
	}
}

func TestExpiredChanExample(t *testing.T) {
	em := NewExpireMap()
	defer em.Destroy()

	em.Set("key1", "val1", 50*time.Millisecond)
	em.Set("key2", "val2", 100*time.Millisecond)

	expires := em.ExpiredChan()

	for i := 0; i < 2; i++ {
		select {
		case <-expires:
			// nothing to do
		case <-time.After(50*time.Millisecond + DELTA*2):
			t.Fatal("Expired key was not read from the channel on time")
		}
	}
}
