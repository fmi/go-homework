package main

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

const DELTA = 10 * time.Millisecond

func TestSettingAndGettingAKey(t *testing.T) {
	defer func() {
		if v := recover(); v != nil {
			t.Errorf("Panic: %s", v)
		}
	}()
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

func TestGettingTypes(t *testing.T) {
	mp := NewExpireMap()
	defer mp.Destroy()

	var expectedInt int = -32
	mp.Set("int", expectedInt, 7*time.Second)

	var expectedFloat64 float64 = 111.23
	mp.Set("float64", expectedFloat64, 7*time.Second)

	var expectedBool bool = true
	mp.Set("bool", expectedBool, 7*time.Second)

	foundInt, _ := mp.GetInt("int")
	if foundInt != expectedInt {
		t.Errorf("Expected int %d but found %d", expectedInt, foundInt)
	}

	eps := 1e-6

	foundFloat, _ := mp.GetFloat64("float64")

	if foundFloat < expectedFloat64-eps ||
		foundFloat > expectedFloat64+eps {
		t.Errorf("Expected float32 %f but found %f", expectedFloat64, foundFloat)
	}

	foundBool, _ := mp.GetBool("bool")

	if foundBool != expectedBool {
		t.Errorf("Expected bool %b but found %b", expectedBool, foundBool)
	}
}

func TestSizes(t *testing.T) {
	defer func() {
		if v := recover(); v != nil {
			t.Errorf("Panic: %s", v)
		}
	}()

	mp := NewExpireMap()
	defer mp.Destroy()

	size := mp.Size()

	if size != 0 {
		t.Errorf("Expected size 0 but got %d", size)
	}

	expected := 10

	for i := 0; i < expected; i++ {
		key := fmt.Sprintf("key-%d", i)
		val := fmt.Sprintf("value-%d", i)
		mp.Set(key, val, time.Second)
	}

	size = mp.Size()

	if size != expected {
		t.Errorf("Expected size %d but got %d", expected, size)
	}

	time.Sleep(time.Second + DELTA)
	ping := make(chan struct{})

	go func() {
		select {
		case <-ping:
			// ok
		case <-time.After(DELTA * 5):
			t.Fatalf("Size did not return in time. Possibly there is a deadlock.")
		}
	}()

	size = mp.Size()
	close(ping)

	if size != 0 {
		t.Errorf("Expected size 0 but got %d", size)
	}
}

func TestMultipleSetsOnTheSameKey(t *testing.T) {
	mp := NewExpireMap()
	defer mp.Destroy()

	mp.Set("foo", 5, time.Second)
	mp.Set("foo", "larodi", 100*time.Millisecond)

	val, ok := mp.GetString("foo")

	if ok == false {
		t.Errorf("Getting a key returned an error when it shouldn't have")
	}

	if val != "larodi" {
		t.Errorf("Getting larodi failed")
	}

	time.Sleep(100*time.Millisecond + DELTA)

	_, ok = mp.GetString("foo")

	if ok {
		t.Errorf("ExpireMap had a key when it should have expired")
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

func TestContainsMethod(t *testing.T) {

	mp := NewExpireMap()
	defer mp.Destroy()

	mp.Set("foo", "super-foo", 3*time.Second)
	mp.Set("bar", "super-bar", 3*time.Second)

	if mp.Contains("foo") == false {
		t.Errorf("Contains: false negative with foo")
	}

	if mp.Contains("baba") {
		t.Errorf("Contains: false positive with baba")
	}

}

func TestDeleteMethod(t *testing.T) {
	mp := NewExpireMap()
	defer mp.Destroy()

	mp.Set("foo", "bar", 15*time.Second)
	mp.Set("baba", "larodi", 15*time.Second)

	mp.Delete("foo")

	if _, ok := mp.GetString("foo"); ok {
		t.Error("Key foo did not get deleted")
	}

	if _, ok := mp.Expires("foo"); ok {
		t.Error("Key foo expires did not get deleted")
	}

	found, ok := mp.GetString("baba")

	if ok == false {
		t.Errorf("Did not find baba after deleting foo")
	}

	expected := "larodi"

	if found != expected {
		t.Errorf("Expected %s but found %s", expected, found)
	}
}

func TestSimpleIncrementAndDecrementCalls(t *testing.T) {

	defer func() {
		if err := recover(); err != nil {
			t.Errorf("Fail simple increment and decrement calls:\n%s\n", err)
		}
	}()
	mp := NewExpireMap()
	defer mp.Destroy()

	mp.Set("foo", "bar", 15*time.Second)
	mp.Set("number", "794", 15*time.Second)
	mp.Set("negative", "-5", 15*time.Second)

	if err := mp.Increment("number"); err != nil {
		t.Errorf("Incrementing returned error %s", err)
	}

	found, ok := mp.GetString("number")

	if ok == false {
		t.Errorf("Getting the number returned an error")
	}

	if found != "795" {
		t.Errorf("Expected to get 795 but got %s", found)
	}

	if err := mp.Decrement("number"); err != nil {
		t.Errorf("Decrementing returned error %s", err)
	}

	found, ok = mp.GetString("number")

	if ok == false {
		t.Errorf("Getting the number returned an error")
	}

	if found != "794" {
		t.Errorf("Expected to get 794 but got %s", found)
	}

	err := mp.Increment("foo")

	if err == nil {
		t.Errorf("Incrementing non-number value did not return error")
	}

	err = mp.Increment("not-here")

	if err == nil {
		t.Errorf("Incrementing not present key did not return error")
	}

	err = mp.Decrement("foo")

	if err == nil {
		t.Errorf("Decrementing non-number value did not return error")
	}

	err = mp.Decrement("not-here")

	if err == nil {
		t.Errorf("Decrementing not present key did not return error")
	}

	err = mp.Increment("negative")

	if err != nil {
		t.Errorf("Incrementing negative number faild")
	}

	number, ok := mp.GetString("negative")

	if ok == false {
		t.Errorf("Getting the negative number faild")
	}

	if number != "-4" {
		t.Errorf("Incrementing the negative number failed")
	}

	mp.Set("integer", 3, 15*time.Second)
	mp.Set("float", 5.0, 15*time.Second)
	mp.Increment("integer")

	foundInt, _ := mp.GetInt("integer")

	if foundInt != 4 {
		t.Error("Incrementing integer failed")
	}

	err = mp.Increment("float")

	if err == nil {
		t.Error("Incrementing float did not return an error")
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

func TestIncAndDecInManyRoutines(t *testing.T) {
	mp := NewExpireMap()
	defer mp.Destroy()

	expected := "666"
	mp.Set("number", expected, 40*time.Second)

	var wg sync.WaitGroup

	passes := 1000
	routines := 4
	tries := 3

	for i := 0; i < tries; i++ {
		for i := 0; i < routines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < passes; i++ {
					mp.Increment("number")
				}
			}()
		}

		for i := 0; i < routines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < passes; i++ {
					mp.Decrement("number")
				}
			}()
		}

		wg.Wait()

		found, ok := mp.GetString("number")

		if ok == false {
			t.Fatalf("Getting the number faild")
		}

		if found != expected {
			t.Fatalf("Expected %s but found %s", expected, found)
		}
	}
}

func TestStoringNil(t *testing.T) {
	mp := NewExpireMap()
	defer mp.Destroy()

	mp.Set("nil-is-here", nil, 5*time.Second)

	n, ok := mp.Get("nil-is-here")

	if !ok {
		t.Error("Nil was not found in the ExpireMap")
	}

	if n != nil {
		t.Error("Returned value was not nil")
	}

}

type StrucForTest struct {
	field int
}

func TestStoringArbitraryStrunct(t *testing.T) {
	mp := NewExpireMap()
	defer mp.Destroy()

	expected := StrucForTest{42}

	mp.Set("struct-here", expected, 15*time.Second)

	found, ok := mp.Get("struct-here")

	if !ok {
		t.Error("The struct was not found in dictionary")
	}

	_, ok = found.(StrucForTest)

	if !ok {
		t.Error("The returned value was not of type StructForTest")
	}

	if found != expected {
		t.Errorf("The found struct was not equal to the one added. "+
			"Expected %v but got %v", expected, found)
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

func TestToUpperExample(t *testing.T) {
	em := NewExpireMap()
	val := "creeper on the roof"
	em.Set("example", val, 10*time.Minute)
	if err := em.ToUpper("example"); err != nil {
		t.Errorf("ToUpping a string returned error: %s", err)
	}
	upped, _ := em.GetString("example")

	expected := "CREEPER ON THE ROOF"
	if upped != expected {
		t.Errorf("Reversing `%s` did not return `%s` but `%s`", val,
			expected, upped)
	}

	em.Destroy()
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

func TestExpiredChannel(t *testing.T) {
	cache := NewExpireMap()
	defer cache.Destroy()
	expires := cache.ExpiredChan()

	cache.Set("long-key", struct{}{}, 1*time.Second)

	expireDuration := 50 * time.Millisecond

	added := time.Now()
	cache.Set("key", struct{}{}, expireDuration)

	select {
	case key := <-expires:
		if key != "key" {
			t.Fatal("Wrong key expired")
		}

		now := time.Now()
		if now.Before(added.Add(expireDuration)) {
			t.Error("The key expired too early")
		}
		if now.After(added.Add(expireDuration + DELTA)) {
			t.Error("The key expired too late")
		}
	case <-time.After(expireDuration + DELTA*2):
		t.Fatal("Expired key was not read from the channel on time")
	}

}

func TestExpiredChanWhenNoOneIsReading(t *testing.T) {
	cache := NewExpireMap()
	defer cache.Destroy()
	expires := cache.ExpiredChan()

	expireDuration := 120 * time.Millisecond
	shorterDuration := 40 * time.Millisecond
	added := time.Now()
	cache.Set("long-key", added, expireDuration)
	cache.Set("key", added, shorterDuration)

	time.Sleep(shorterDuration + 2*DELTA)

	select {
	case key := <-expires:
		if key != "long-key" {
			t.Fatal("Wrong key expired")
		}

		now := time.Now()
		if now.Before(added.Add(expireDuration)) {
			t.Error("The key expired too early")
		}
		if now.After(added.Add(expireDuration + DELTA)) {
			t.Error("The key expired too late")
		}
	case <-time.After(expireDuration + DELTA*2):
		t.Fatal("Expired key was not read from the channel on time")
	}
}

func TestExpiredChanDoesNotReturnDeletedKeys(t *testing.T) {
	cache := NewExpireMap()
	defer cache.Destroy()
	expires := cache.ExpiredChan()

	expireDuration := 100 * time.Millisecond

	cache.Set("long-key", time.Now(), expireDuration)
	cache.Set("key", time.Now(), 50*time.Millisecond)

	cache.Delete("key")

	var found string

	select {
	case found = <-expires:
	case <-time.After(expireDuration + 2*DELTA):
		t.Fatal("Expired key was notread from the channel on time")
	}

	if found == "key" {
		t.Error("Expires chan returned deleted key")
	}

	if found != "long-key" {
		t.Error("Expire chan did not return the expected key: long-key")
	}
}

func TestExpiredChanDoesNotReturnCleanedupKeys(t *testing.T) {
	cache := NewExpireMap()
	defer cache.Destroy()
	expires := cache.ExpiredChan()

	expireDuration := 100 * time.Millisecond

	cache.Set("long-key", time.Now(), expireDuration)
	cache.Set("key", time.Now(), 50*time.Millisecond)

	cache.Cleanup()

	select {
	case found := <-expires:
		t.Errorf("Expire chan returned key %s after it was cleaned up", found)
	case <-time.After(expireDuration + 2*DELTA):
		// ok, no expiration after cleanup
	}
}

func TestConcurrentOperations(t *testing.T) {
	cache := NewExpireMap()
	defer cache.Destroy()

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {

		syncChan := make(chan struct{})
		cache.Set("integer", 1, 15*time.Second)
		cache.Set("string", "tiger", 15*time.Second)

		// Starting up functions which will change the value during the concurrent
		// operations
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-syncChan
			time.Sleep(5 * time.Nanosecond)
			cache.Set("integer", int(1500000), 15*time.Second)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			<-syncChan
			time.Sleep(7 * time.Nanosecond)
			cache.Set("string", "a б", 15*time.Second)
		}()

		// Starting routines which will do work on the keys conrurrently
		for i := 0; i < 10; i++ {

			// This goroutine will increment the integer key many times as fast
			// as it could possibly do
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-syncChan
				for i := 0; i < 1000; i++ {
					cache.Increment("integer")
				}
				time.Sleep(5 * time.Nanosecond)
				for i := 0; i < 1000; i++ {
					cache.Increment("integer")
				}
			}()

			// This goroutine will ToUpper the string key many times as fast
			// as it could possibly do
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-syncChan
				for i := 0; i < 1000; i++ {
					cache.ToUpper("string")
				}
				time.Sleep(7 * time.Nanosecond)
				for i := 0; i < 1000; i++ {
					cache.ToUpper("string")
				}
			}()

			key := fmt.Sprintf("key-%d", i)
			value := fmt.Sprintf("value-%d", i)
			cache.Set(key, value, 15*time.Minute)

			// Staring a goroutine which will just Get a key repeatedly to
			// make sure any of the other operations does not remove the
			// whole storage at some point.
			wg.Add(1)
			go func(key, value string) {
				defer wg.Done()
				<-syncChan
				for i := 0; i < 1000; i++ {
					found, ok := cache.GetString(key)
					if !ok {
						t.Errorf("Key %s was not found in the cache", key)
					}

					if found != value {
						t.Errorf("Expected value %s but found %s", value, found)
					}
				}
			}(key, value)
		}

		close(syncChan)
		wg.Wait()

		found, ok := cache.GetInt("integer")

		if !ok {
			t.Fatal("Expected integer key to be in the map but it was not")
		}

		// If `found` is less than 1500000 it means some of the Increments has
		// overwritten the value set in the middle of execution.
		if found < 1500000 {
			t.Fatal("Expected something bigger than 1.5e6 but found", found)
		}

		foundString, ok := cache.GetString("string")

		if !ok {
			t.Fatal("Expected integer key to be in the map but it was not")
		}

		// If `found` is different it means some of the ToUppers has
		// overwritten the value set in the middle of execution.
		if foundString != "A Б" {
			t.Fatalf("Expected `A Б` but found `%s`", foundString)
		}
	}
}

func TestMultipleMapsWithDestroyAndCleanupCalls(t *testing.T) {
	mp := NewExpireMap()
	defer mp.Destroy()

	toBeDestroied := NewExpireMap()

	toBeDestroied.Set("key", "value", 15*time.Second)
	toBeDestroied.Set("key2", "value", 15*time.Second)
	toBeDestroied.Set("key3", "value", 15*time.Second)

	mp.Set("key", "value-mp", 14*time.Second)
	mp.Set("key2", "value-mp", 14*time.Second)
	mp.Set("key3", "value-mp", 14*time.Second)

	val, ok := toBeDestroied.GetString("key")

	if !ok {
		t.Error("'key' was not found in the toBeDestroied map")
	}

	if val != "value" {
		t.Errorf("Expected 'value' but got '%s'", val)
	}

	toBeDestroied.Cleanup()

	if mp.Size() != 3 {
		t.Error("Expected size 3 but got", mp.Size())
	}

	toBeDestroied.Destroy()

	if mp.Size() != 3 {
		t.Error("Expected size 3 but got", mp.Size())
	}

	val, ok = mp.GetString("key")

	if !ok {
		t.Error("Getting a value after Destroy of another map failed")
	}

	if val != "value-mp" {
		t.Errorf("Expected 'value-mp' but got '%s'", val)
	}

	mp.Set("yellow", "blue", 3*time.Second)

	colour, _ := mp.GetString("yellow")

	if colour != "blue" {
		t.Error("Getting a value from the map failed after destroy of " +
			"completely different instance")
	}
}

func TestDestroyMethodClosesExpireChannel(t *testing.T) {
	mp := NewExpireMap()

	expChn := mp.ExpiredChan()
	var wg sync.WaitGroup
	ping := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case val := <-expChn:
			if val != "" {
				t.Errorf("Expire channel was not closed properly. " +
					"It returned non-zero value")
			}
			ping <- struct{}{}
		case <-ping:
			return
		}

	}()

	mp.Destroy()

	select {
	case <-ping:
		// ok
	case <-time.After(DELTA * 5):
		t.Errorf("Expire channel was not closed in time")
		ping <- struct{}{}
	}

	close(ping)
	wg.Wait()
}

func TestExpiresMethod(t *testing.T) {
	cache := NewExpireMap()
	defer cache.Destroy()

	keys := make(map[string]time.Duration)

	keys["key1"] = 10 * time.Second
	keys["key2"] = 15 * time.Second
	keys["key3"] = 20 * time.Second

	for key, duration := range keys {
		cache.Set(key, time.Now(), duration)
	}

	for key, duration := range keys {
		expires, found := cache.Expires(key)
		if !found {
			t.Error("Did not find expires for key", key)
		}

		added, found := cache.Get(key)
		if !found {
			t.Errorf("Getting '%s' failed", key)
		}

		expected := added.(time.Time).Add(duration)

		var diff time.Duration

		if expected.After(expires) {
			diff = expected.Sub(expires)
		} else {
			diff = expires.Sub(expected)
		}

		if diff > 2*DELTA {
			t.Errorf("Expected expire time %s but got %s", expected, expires)
		}
	}

	_, found := cache.Expires("not-there")

	if found {
		t.Error("Expires returnd ok for key which is not there")
	}

}

func TestMultipleKeysWithTheSameExpireTime(t *testing.T) {
	cache := NewExpireMap()
	defer cache.Destroy()

	duration := 250 * time.Millisecond
	keys := []string{}
	for i := 0; i < 15; i++ {
		keys = append(keys, fmt.Sprintf("same-time-key-%d", i))
	}

	inFuture := time.Now().Add(duration)

	for _, key := range keys {
		cache.Set(key, inFuture, inFuture.Sub(time.Now()))
	}

	time.Sleep(inFuture.Sub(time.Now()) + DELTA)

	for _, key := range keys {
		_, ok := cache.Get(key)
		if ok {
			t.Errorf("Key %s did not expire", key)
		}

		_, ok = cache.Expires(key)
		if ok {
			t.Errorf("Key %s had expire time even though it has expired", key)
		}
	}
}

func TestStringMethods(t *testing.T) {
	em := NewExpireMap()
	defer em.Destroy()

	key := "string-key"

	em.Set(key, "Some Value Тук", 5*time.Minute)

	if err := em.ToUpper(key); err != nil {
		t.Error("ToUpper returned an error")
	}

	expected := "SOME VALUE ТУК"
	found, ok := em.GetString(key)

	if !ok {
		t.Errorf("Was not able to find key %s", key)
	}

	if found != expected {
		t.Errorf("ToUpper did not return %s but %s", expected, found)
	}

	if err := em.ToLower(key); err != nil {
		t.Error("ToLower returned an error")
	}

	expected = "some value тук"
	found, ok = em.GetString(key)

	if !ok {
		t.Errorf("Was not able to find key %s", key)
	}

	if found != expected {
		t.Errorf("ToLower did not return %s but %s", expected, found)
	}

	em.Set("not-string", 5.44, 5*time.Second)

	for _, key := range []string{"not-there", "not-string"} {

		if err := em.ToUpper(key); err == nil {
			t.Error("ToUpper for key did not return an error")
		}

		if err := em.ToLower(key); err == nil {
			t.Error("ToLower for key did not return an error")
		}
	}
}
