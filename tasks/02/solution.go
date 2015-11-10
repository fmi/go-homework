package main

import "strconv"

func drainLog(num int, log, result chan string, lastIsDone chan struct{}) chan struct{} {
	thisIsDone := make(chan struct{})
	buffer := make(chan string, 100)

	go func() {
		for logEntry := range log {
			buffer <- strconv.Itoa(num) + "\t" + logEntry
		}
		close(buffer)
	}()

	go func() {
		<-lastIsDone
		for modifiedLogEntry := range buffer {
			result <- modifiedLogEntry
		}
		close(thisIsDone)
	}()

	return thisIsDone
}

// OrderedLogDrainer takes care to drain all incoming logs and output
// them in an ordered manner.
func OrderedLogDrainer(logs chan (chan string)) chan string {
	result := make(chan string)

	go func() {
		i := 0
		lastIsDone := make(chan struct{})
		close(lastIsDone)
		for log := range logs {
			i++
			lastIsDone = drainLog(i, log, result, lastIsDone)
		}
		<-lastIsDone
		close(result)
	}()

	return result
}
