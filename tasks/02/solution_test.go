package main

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

// generateTaskFunc generates a task func returns also a channel on which will be send when the task
// is called and a second one on which it will wait to return a value provided on it
func generateTaskFunc() (func() string, <-chan struct{}, chan<- string) {
	var called = make(chan struct{})
	var ret = make(chan string)
	var f = func() string {
		called <- struct{}{}
		return <-ret
	}

	return f, called, ret
}

func generateNTasks(count int) (
	tasks []func() string,
	called []<-chan struct{},
	ret []chan<- string,
) {
	tasks = make([]func() string, count)
	called = make([]<-chan struct{}, count)
	ret = make([]chan<- string, count)
	for i := 0; i < count; i++ {
		tasks[i], called[i], ret[i] = generateTaskFunc()
	}
	return
}

func generateIds(count int) (result []int) {
	for i := 0; i < count; i++ {
		result = append(result, i)
	}
	return
}

// chResult changes the struct provided if needed and returns true if the task is no longer expected
// to finish
func testBasic(t *testing.T, count, concurrentLimit, retryLimit int, chResult func(*struct {
	index  int
	result string
}) bool) {
	var (
		tasks, called, ret = generateNTasks(count)

		respCh          = ConcurrentRetryExecutor(tasks, concurrentLimit, retryLimit)
		currentlyInCall = 0
		alreadyReturned = 0
		unfinishedTasks = generateIds(count)
	)

	for alreadyReturned != count {
		for i := range unfinishedTasks {
			if i >= concurrentLimit {
				select {
				case <-called[unfinishedTasks[i]]:
					t.Errorf("task with id %d started executing when"+
						" already %d tasks were executing with limit %d",
						unfinishedTasks[i], currentlyInCall, concurrentLimit)
				default:
				}
			} else {
				<-called[unfinishedTasks[i]]
				currentlyInCall++
			}
		}
		var left = min(concurrentLimit, len(unfinishedTasks))
		for i := 0; left != 0; left-- {
			var (
				expectedResponse = struct {
					index  int
					result string
				}{
					index:  unfinishedTasks[i],
					result: strconv.Itoa(unfinishedTasks[i]),
				}
			)
			var isFinished = chResult(&expectedResponse)
			ret[expectedResponse.index] <- expectedResponse.result
			var resp = <-respCh
			checkResponse(t, resp, expectedResponse)
			if isFinished {
				alreadyReturned++
				unfinishedTasks = append(unfinishedTasks[:i], unfinishedTasks[i+1:]...)
			} else {
				i++
			}
		}
	}
	var resp, ok = <-respCh
	if ok != false {
		t.Errorf("Response channel not closed after everything was executed and returned %s", resp)
	}
}

func TestLessTaskThanConcurrency(t *testing.T) {
	testBasic(t, 2, 10, 10/2, func(*struct {
		index  int
		result string
	}) bool {
		return true
	})
}

func TestBasic1Concurrency(t *testing.T) {
	testBasic(t, 1024, 1, 10/2, func(*struct {
		index  int
		result string
	}) bool {
		return true
	})
}

func TestBasic10Concurrency(t *testing.T) {
	testBasic(t, 2048, 10, 10/2, func(*struct {
		index  int
		result string
	}) bool {
		return true
	})
}

func TestBasicFail1Concurrency(t *testing.T) {
	testBasicFail(t, 20, 1, 5)
}

func TestBasicFail10Concurrency(t *testing.T) {
	testBasicFail(t, 20, 10, 5)
}

func TestBasicFail1024Concurrency(t *testing.T) {
	testBasicFail(t, 5*1024, 1024, 3)
}

func testBasicFail(t *testing.T, count, concurrentLimit, retryLimit int) {
	var (
		fails = make([]int, count)
	)

	testBasic(t, count, concurrentLimit, retryLimit, func(resp *struct {
		index  int
		result string
	}) bool {
		if resp.index%3 != 1 {
			return true
		}
		if resp.index%5 == 0 && fails[resp.index] == (resp.index%retryLimit) {
			return true
		}
		resp.result = ""
		fails[resp.index]++
		if fails[resp.index] > retryLimit {
			t.Errorf("Failing task with id %d for the %d time which more than the max of %d",
				resp.index, fails[resp.index], retryLimit)
		}
		return fails[resp.index] >= retryLimit
	})
	for index, failed := range fails {
		if index%3 == 1 {
			var expectedFails = retryLimit
			if index%5 == 0 {
				expectedFails = (index % retryLimit)
			}
			if failed != expectedFails {
				t.Errorf("It was expected that at the end task"+
					" with index %d would have failed %d times but it has %d",
					index, expectedFails, failed)
			}
		}
	}
}

func TestPerformance(t *testing.T) {
	var (
		repeates        = 10
		size            = 10 * 1024
		concurrentLimit = 1024
		retryLimit      = 50
		tasks           = make([]func() string, 0, size)
		f               = func() string {
			return "perf"
		}
	)
	for i := 0; i < size; i++ {
		tasks = append(tasks, f)
	}
	for i := 0; i < repeates; i++ {
		var ch = ConcurrentRetryExecutor(tasks, concurrentLimit, retryLimit)

		for taskResult := range ch {
			if taskResult.result != "perf" {
				t.Errorf("Got %s should've gotten 'perf' as result", taskResult)
			}
		}
	}
}

func TestConcurrentCalls(t *testing.T) {
	var (
		concurrent      = 15
		repeates        = 15
		size            = 1024
		concurrentLimit = 10
		retryLimit      = 5
		waitCh          = make(chan struct{}, concurrent)
		tasks           = make([]func() string, 0, size)
		f               = func() string {
			return "perf"
		}
	)
	for i := 0; i < size; i++ {
		tasks = append(tasks, f)
	}

	for j := 0; j < concurrent; j++ {
		waitCh <- struct{}{}
		go func() {
			defer func() {
				<-waitCh
			}()

			for i := 0; i < repeates; i++ {
				var ch = ConcurrentRetryExecutor(tasks, concurrentLimit, retryLimit)

				for taskResult := range ch {
					if taskResult.result != "perf" {
						t.Errorf("Got %s should've gotten 'perf' as result", taskResult)
					}
				}
			}
		}()
	}
	for j := 0; j < concurrent; j++ {
		waitCh <- struct{}{}
	}
}

func checkResponse(t *testing.T, response, expectedResponse struct {
	index  int
	result string
}) {
	if response.index != expectedResponse.index {
		t.Errorf("index was expected to be %d was %d",
			expectedResponse.index, response.index)
	}
	if response.result != expectedResponse.result {
		t.Errorf("for index %d the expected result was %s but got %s",
			expectedResponse.index, expectedResponse.result, response.result)
	}
}

// min for na--
func min(l, r int) int {
	if l < r {
		return l
	}
	return r
}

func TestExampleSimpleScenario(t *testing.T) {
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

func ExampleConcurrentRetryExecutorFromTheExample() {
	first := func() string {
		time.Sleep(200 * time.Millisecond)
		return "first"
	}
	second := func() string {
		time.Sleep(60 * time.Millisecond)
		return "second"
	}
	third := func() string {
		time.Sleep(100 * time.Millisecond)
		return "" // always a failure :(
	}
	fourth := func() string {
		time.Sleep(80 * time.Millisecond)
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
