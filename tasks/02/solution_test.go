package main

import (
	"fmt"
	"runtime"
	"strconv"
	"testing"
	"time"
)

func withTimeout(t *testing.T, timeout time.Duration, action func()) {
	finished := make(chan bool)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Panicked during the test with message '%s'", r)
			}
			close(finished)
		}()
		action()
	}()

	select {
	case <-finished:
	case <-time.After(timeout):
		t.Fatalf("Test exceeded allowed time of %d milliseconds", timeout/time.Millisecond)
	}
	return
}

func resultChecker(t *testing.T, orderedLog <-chan string, expectedResults []string) chan struct{} {
	done := make(chan struct{})
	go func() {
		for num, expRes := range expectedResults {
			res, ok := <-orderedLog
			if !ok {
				t.Errorf("Channel was closed when we expected to receive entry #%d with contents '%s'", num, expRes)
			} else if res != expRes {
				t.Errorf("Received log entry #%d with contents '%s' when expecting '%s'", num, res, expRes)
			}
		}
		res, ok := <-orderedLog
		if ok {
			t.Errorf("Channel was stil open when. We expected it to only have %d entries but we received '%s'", len(expectedResults), res)
		}
		close(done)
	}()
	return done
}

func TestWithOneMessage(t *testing.T) {
	t.Parallel()
	withTimeout(t, 500*time.Millisecond, func() {
		logs := make(chan (chan string))
		orderedLog := OrderedLogDrainer(logs)
		log1 := make(chan string)
		logs <- log1
		log1 <- "aaa"
		done := resultChecker(t, orderedLog, []string{"1\taaa"})
		close(logs)
		close(log1)
		<-done
	})
}

func TestWithExample1(t *testing.T) {
	t.Parallel()
	expectedResult := []string{
		"1	test message 1 in first",
		"1	test message 2 in first",
		"1	test message 3 in first",
		"1	test message 4 in first",
		"2	test message 1 in second",
		"2	test message 2 in second",
		"2	test message 3 in second",
		"2	test message 4 in second",
		"3	test message 1 in third",
	}

	withTimeout(t, 500*time.Millisecond, func() {
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
		<-resultChecker(t, orderedLog, expectedResult)
	})
}

func TestWithExample2(t *testing.T) {
	t.Parallel()
	expectedResult := []string{
		"1	aaa",
		"1	ccc",
		"2	bbb",
		"2	ddd",
	}

	withTimeout(t, 500*time.Millisecond, func() {
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
		<-resultChecker(t, orderedLog, expectedResult)
	})
}

func TestWithNoLogs(t *testing.T) {
	t.Parallel()
	withTimeout(t, 500*time.Millisecond, func() {
		logs := make(chan (chan string))
		orderedLog := OrderedLogDrainer(logs)
		done := resultChecker(t, orderedLog, []string{})
		close(logs)
		<-done
	})
}

func TestWithTwoEmptyLogs(t *testing.T) {
	t.Parallel()
	withTimeout(t, 500*time.Millisecond, func() {
		logs := make(chan (chan string))
		orderedLog := OrderedLogDrainer(logs)
		log1 := make(chan string)
		logs <- log1
		log2 := make(chan string)
		logs <- log2
		log3 := make(chan string)
		logs <- log3
		done := resultChecker(t, orderedLog, []string{"3\taaa", "3\tbbb"})
		close(log2)
		log3 <- "aaa"
		close(logs)
		close(log1)
		log3 <- "bbb"
		close(log3)
		<-done
	})
}

func TestWithDelays(t *testing.T) {
	t.Parallel()
	withTimeout(t, 900*time.Millisecond, func() {
		logs := make(chan (chan string))
		orderedLog := OrderedLogDrainer(logs)
		done := make(chan struct{})
		go func() {
			withTimeout(t, 400*time.Millisecond, func() {
				firstResult := <-orderedLog
				expected := "1\tfirst"
				if firstResult != expected {
					t.Errorf("Invalid first result '%s', expected '%s'", firstResult, expected)
				}
			})
			close(done)
		}()

		log1 := make(chan string)
		logs <- log1
		log2 := make(chan string)
		logs <- log2
		log1 <- "first"
		time.Sleep(700 * time.Millisecond)
		log2 <- "second"
		close(logs)
		log2 <- "third"
		log1 <- "forth"
		close(log2)
		log1 <- "fifth"
		close(log1)
		<-resultChecker(t, orderedLog, []string{"1\tforth", "1\tfifth", "2\tsecond", "2\tthird"})
		<-done
	})
}

func TestTheLimits(t *testing.T) {
	t.Parallel()
	runtime.GOMAXPROCS(runtime.NumCPU())
	const reps = 100
	expectedResult := make([]string, 2*reps+1)

	withTimeout(t, 500*time.Millisecond, func() {
		logs := make(chan (chan string))
		orderedLog := OrderedLogDrainer(logs)

		log1 := make(chan string)
		logs <- log1
		log2 := make(chan string)
		logs <- log2
		log3 := make(chan string)
		logs <- log3
		close(logs)
		log3 <- "end!"
		close(log3)
		go func() {
			for i := 1; i <= reps; i++ {
				log1 <- fmt.Sprintf("Ni %.03d", i)
			}
			close(log1)
		}()
		go func() {
			for i := 1; i <= reps; i++ {
				log2 <- fmt.Sprintf("Ni %.03d", i)
			}
			close(log2)
		}()
		for i := 0; i < 2*reps; i++ {
			expectedResult[i] = fmt.Sprintf("%d\tNi %.03d", (i/100)+1, i%100+1)
		}
		expectedResult[2*reps] = "3\tend!"
		<-resultChecker(t, orderedLog, expectedResult)
	})
}

func TestWithManyLogs(t *testing.T) {
	t.Parallel()

	const logsCount = 150
	expectedResult := make([]string, logsCount)
	logs := make([]chan string, logsCount)
	logsCh := make(chan (chan string))

	withTimeout(t, 800*time.Millisecond, func() {
		orderedLog := OrderedLogDrainer(logsCh)

		for i := 0; i < logsCount; i++ {
			logs[i] = make(chan string)
			logsCh <- logs[i]
			logs[i] <- "test"
			expectedResult[i] = strconv.Itoa(i+1) + "\ttest"
		}

		for i := 0; i < logsCount; i++ {
			close(logs[i])
		}
		close(logsCh)

		<-resultChecker(t, orderedLog, expectedResult)
	})
}
