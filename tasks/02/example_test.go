package main

import "fmt"

func ExampleWithTwoLogs() {
	logs := make(chan (chan string))
	orderedLog := OrderedLogDrainer(logs)

	log1 := make(chan string)
	logs <- log1
	log2 := make(chan string)
	logs <- log2
	close(logs)

	log1 <- "aaa"
	log2 <- "bbb"
	log1 <- "ccc"
	log2 <- "ddd"
	close(log1)
	close(log2)

	for logEntry := range orderedLog {
		fmt.Println(logEntry)
	}
	// Output:
	// 1	aaa
	// 1	ccc
	// 2	bbb
	// 2	ddd
}

func ExampleFromTaskDesctiption() {
	logs := make(chan (chan string))
	orderedLog := OrderedLogDrainer(logs)

	first := make(chan string)
	logs <- first
	second := make(chan string)
	logs <- second

	first <- "test message 1 in first"
	second <- "test message 1 in second"
	second <- "test message 2 in second"
	first <- "test message 2 in first"
	first <- "test message 3 in first"
	// Print the first message now just because we can
	fmt.Println(<-orderedLog)

	third := make(chan string)
	logs <- third

	third <- "test message 1 in third"
	first <- "test message 4 in first"
	close(first)
	second <- "test message 3 in second"
	close(third)
	close(logs)

	second <- "test message 4 in second"
	close(second)

	// Print all the rest of the messages
	for logEntry := range orderedLog {
		fmt.Println(logEntry)
	}
	// Output:
	// 1	test message 1 in first
	// 1	test message 2 in first
	// 1	test message 3 in first
	// 1	test message 4 in first
	// 2	test message 1 in second
	// 2	test message 2 in second
	// 2	test message 3 in second
	// 2	test message 4 in second
	// 3	test message 1 in third
}
