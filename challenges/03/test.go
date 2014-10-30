package main

import (
	"testing"
	"time"
)

func TestWithTwoListeners(t *testing.T) {
	b := NewBroadcaster()
	listener1 := b.Register()
	listener2 := b.Register()
	b.Send() <- 42
	result1 := <-listener1
	result2 := <-listener2

	if result1 != result2 {
		t.Errorf("Expected both listeners to recieve the same number, but they recieved %v, %v", result1, result2)
	}
}

func TestLateToTheParty(t *testing.T) {
	var result int
	b := NewBroadcaster()
	b.Send() <- 42

	time.Sleep(10 * time.Millisecond)
	listener := b.Register()

	select {
	case res := <-listener:
		result = res
	case _ = <-time.After(100 * time.Millisecond):
		result = 0
	}

	if result != 0 {
		t.Errorf("Expected listeners registered later to not recieve the message, but the listener recieved %v", result)
	}
}
