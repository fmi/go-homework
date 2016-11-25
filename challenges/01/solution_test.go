package main

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func ExamplePubSub() {
	ps := NewPubSub()
	a := ps.Subscribe()
	b := ps.Subscribe()
	c := ps.Subscribe()
	go func() {
		ps.Publish() <- "wat"
		ps.Publish() <- ("wat" + <-c)
	}()
	fmt.Printf("A recieved %s, B recieved %s and we ignore C!\n", <-a, <-b)
	fmt.Printf("A recieved %s, B recieved %s and C received %s\n", <-a, <-b, <-c)
	// Output:
	// A recieved wat, B recieved wat and we ignore C!
	// A recieved watwat, B recieved watwat and C received watwat
}

func TestWithTwoSubscribers(t *testing.T) {
	ps := NewPubSub()
	s1 := ps.Subscribe()
	s2 := ps.Subscribe()
	ps.Publish() <- "test"
	result1 := <-s1
	result2 := <-s2

	if result1 != result2 {
		t.Errorf("Expected both subscribers to recieve the same string, but they recieved %s, %s", result1, result2)
	}
}

func TestMultipleMessagesWithoutSubscribers(t *testing.T) {
	ps := NewPubSub()

	for i := 0; i < 100; i++ {
		ps.Publish() <- "a"
	}
}

func TestLateSubscriber(t *testing.T) {
	ps := NewPubSub()
	ps.Publish() <- "test"

	time.Sleep(50 * time.Millisecond)

	select {
	case res := <-ps.Subscribe():
		t.Errorf("Expected later subscribers not to receive missed messages but recieved %v", res)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestWithConccurentSubscribers(t *testing.T) {
	testsCount := 200
	for j := 0; j < testsCount; j++ {
		limit := 20
		ps := NewPubSub()
		done := make(chan struct{}, limit)
		running := make(chan struct{}, limit)
		for i := 0; i < limit; i++ {
			go func(index int) {
				me := ps.Subscribe()
				running <- struct{}{}
				result := <-me
				if result != "Batman" {
					t.Errorf("Listener %d expected to recieve Batman but recieved %s!", index, result)
				}
				done <- struct{}{}
			}(i)
		}

		for i := 0; i < limit; i++ {
			<-running
		}

		ps.Publish() <- "Batman"

		for i := 0; i < limit; i++ {
			<-done
		}
	}
}

func TestMultipleMessagesOneSubscriber(t *testing.T) {
	ps := NewPubSub()
	me := ps.Subscribe()

	go func() {
		for i := 0; i < 100; i++ {
			ps.Publish() <- strings.Repeat("Na", i)
		}
	}()

	for i := 0; i < 100; i++ {
		result := <-me
		expResult := strings.Repeat("Na", i)
		if expResult != result {
			t.Errorf("Expected to recieve %v, but recieved %v.", expResult, result)
			break
		}
	}
}

func TestCorrectConcurrency(t *testing.T) {
	limit := 10
	for i := 0; i < limit; i++ {
		t.Run(fmt.Sprintf("%d of %d", i, limit), func(t *testing.T) {
			t.Parallel()
			ps := NewPubSub()
			first := ps.Subscribe()
			ps.Publish() <- "test"
			select {
			case <-time.After(10 * time.Millisecond):
				t.Errorf("Expected the first subscriber to receive the message")
			case <-first:
			}
			second := ps.Subscribe()

			select {
			case res := <-second:
				t.Errorf("Expected later subscribers not to receive missed messages but recieved %v", res)
			case <-time.After(10 * time.Millisecond):
			}
		})
	}

}
