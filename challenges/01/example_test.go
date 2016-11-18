package main

import "fmt"

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
