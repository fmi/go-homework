package main

import (
	"fmt"
	"testing"
	"time"
)

func TestSimpleScenario(t *testing.T) {
	theQuickBrownFox := func() string { return "jump!" }
	dogIsRested := false
	theLazyDog := func() string {
		if dogIsRested {
			return "woof"
		}
		time.Sleep(200 * time.Millisecond)
		dogIsRested = true
		return ""
	}

	done := make(chan struct{})
	defer close(done)

	go func() {
		select {
		case <-time.After(500 * time.Millisecond):
			t.Fatalf("Test did not finish on time")
		case <-done:
			// Everything is fine
		}
	}()

	results := ConcurrentRetryExecutor([]func() string{theQuickBrownFox, theLazyDog}, 3, 4)
	if res, ok := <-results; ok != true || res.index != 0 || res.result != "jump!" {
		t.Errorf("Expected the first result to be 'jump!' from the quick brown fox but received %+v", res)
	}
	if res, ok := <-results; ok != true || res.index != 1 || res.result != "" {
		t.Errorf("Expected the second result to be an error from the lazy dog but received %+v", res)
	}
	if res, ok := <-results; ok != true || res.index != 1 || res.result != "woof" {
		t.Errorf("Expected the third result to be 'woof' from the lazy dog but received %+v", res)
	}
	if res, ok := <-results; ok != false {
		t.Errorf("Expected the channel to be closed, instead received %+v", res)
	}
}

func ExampleConcurrentRetryExecutor() {
	first := func() string {
		time.Sleep(2 * time.Second)
		return "first"
	}
	second := func() string {
		time.Sleep(1 * time.Second)
		return "second"
	}
	third := func() string {
		time.Sleep(600 * time.Millisecond)
		return "" // always a failure :(
	}
	fourth := func() string {
		time.Sleep(700 * time.Millisecond)
		return "am I last?"
	}

	fmt.Println("Starting concurrent executor!")
	tasks := []func() string{first, second, third, fourth}
	results := ConcurrentRetryExecutor(tasks, 2, 3)
	for result := range results {
		if result.result == "" {
			fmt.Printf("Task %d returned an error!\n", result.index+1)
		} else {
			fmt.Printf("Task %d successfully returned '%s'\n", result.index+1, result.result)
		}
	}
	fmt.Println("All done!")
	// Output:
	// Starting concurrent executor!
	// Task 2 successfully returned 'second'
	// Task 3 returned an error!
	// Task 1 successfully returned 'first'
	// Task 3 returned an error!
	// Task 4 successfully returned 'am I last?'
	// Task 3 returned an error!
	// All done!
}
