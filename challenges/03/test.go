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


func listen(listener <-chan int, id int) chan int {
	newChannel := make(chan int)

	go func() {
		for {
			_ = <-listener
			newChannel <- id
		}
	}()

	return newChannel
}

func TestOnlyOneToRuleThemAll(t *testing.T) {

	b := NewBroadcaster()
	listener1 := listen(b.Register(), 1)
	b.Send() <- 100
	time.Sleep(102 * time.Millisecond)
	listener2 := listen(b.Register(), 2)
	var res int
	select {
	case res = <-listener1:
	case res = <-listener2:
	}

	if res != 1 {
		t.Errorf("Expected  first lestner to recieve the message", res)
	}
}
