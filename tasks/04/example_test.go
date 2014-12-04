package main

import (
	"fmt"
	"strings"
	"time"
)

func ExampleWithExternalUrls() {

	urls := make(chan []string)
	go func() {
		urls <- []string{"http://www.abv.bg", "http://www.dir.bg"}
		time.Sleep(5 * time.Second)
		urls <- []string{"http://www.google.com", "invalid.url....", "http://en.wikipedia.org/wiki/Lorem_ipsum"}
	}()

	callback := func(contents string) bool {
		return strings.Contains(contents, "Lorem ipsum dolor sit amet")
	}

	result, _ := SeekAndDestroy(callback, urls, 3)

	fmt.Println(result)
	// Output: http://en.wikipedia.org/wiki/Lorem_ipsum
}

func ExampleWithTimeout() {

	urls := make(chan []string)
	go func() {
		urls <- []string{"http://www.dir.bg"}
		time.Sleep(15 * time.Second)
		urls <- []string{"http://en.wikipedia.org/wiki/Lorem_ipsum"}
	}()

	callback := func(contents string) bool {
		return strings.Contains(contents, "Lorem ipsum dolor sit amet")
	}

	_, err := SeekAndDestroy(callback, urls, 1)

	if err != nil {
		fmt.Println("An error occurred - probably a timeout :)")
	}

	// Output: An error occurred - probably a timeout :)
}
