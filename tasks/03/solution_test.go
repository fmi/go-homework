package main

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"runtime"
	"runtime/pprof"
	"sort"
	"testing"
	"time"
)

type adder struct {
	augend int
}

func (a adder) Execute(addend int) (int, error) {
	result := a.augend + addend
	if result > 127 {
		return 0, fmt.Errorf("Result %d exceeds the adder threshold", a)
	}
	return result, nil
}

type lazyAdder struct {
	adder
	delay time.Duration
}

func (la lazyAdder) Execute(addend int) (int, error) {
	time.Sleep(la.delay * time.Millisecond)
	return la.adder.Execute(addend)
}

type fTask func(int) (int, error)

func (f fTask) Execute(i int) (int, error) {
	return f(i)
}

func errorTask(msg string) Task {
	return fTask(func(int) (int, error) {
		return 0, errors.New(msg)
	})
}

func TestPipelineCorrect(t *testing.T) {
	if res, err := Pipeline(adder{10}).Execute(-10); err != nil || res != 0 {
		t.Errorf("Expected Pipeline to return 0 but got (%d, %s)",
			res, err)
	}

	if res, err := Pipeline(adder{50}, adder{60}).Execute(10); err != nil || res != 120 {
		t.Errorf("Expected Pipeline to return 120 but got (%d, %s)",
			res, err)
	}
}

func TestPipelineErrors(t *testing.T) {
	if res, err := Pipeline().Execute(1); err == nil {
		t.Errorf("Expected error did not occur instead got %d", res)
	}

	if res, err := Pipeline(errorTask("oops")).Execute(10); err == nil {
		t.Errorf("Expected error did not occur instead got %d", res)
	}

	if res, err := Pipeline(adder{20}, adder{10}, adder{-50}).Execute(100); err == nil {
		t.Errorf("Expected error did not occur instead got %d", res)
	}
}

func TestFastestErrors(t *testing.T) {
	if res, err := Fastest().Execute(5); err == nil {
		t.Errorf("Expected error did not occur instead got %d", res)
	}

	if res, err := Fastest(lazyAdder{adder{20}, 50}, errorTask("oops")).Execute(5); err == nil {
		t.Errorf("Expected error did not occur instead got %d", res)
	}
}

func TestFastestSimple(t *testing.T) {
	f := Fastest(
		lazyAdder{adder{20}, 50},
		lazyAdder{adder{50}, 30},
		adder{41},
	)
	if result, err := f.Execute(1); err != nil {
		t.Errorf("Received an unexpected error %s", err)
	} else if result != 42 {
		t.Errorf("Expected to receive 42 but received %d", result)
	}

	// This is intentional so the test below will work when run together
	time.Sleep(100 * time.Millisecond)
}

// This test may fail when run together with the other tests
// (for example via `go test ./...`). Evans runs the tests one by one, so it
// should not be a problem for the specific task
func TestFastestWaitsForGoroutines(t *testing.T) {
	var (
		firstStarted     = make(chan struct{})
		firstReturn      = make(chan struct{})
		firstReturned    = make(chan struct{})
		secondReturn     = make(chan struct{})
		end              = make(chan struct{})
		routinesAtStart  = runtime.NumGoroutine()
		startStacktraces = &bytes.Buffer{}
	)
	pprof.Lookup("goroutine").WriteTo(startStacktraces, 1)

	f := Fastest(
		fTask(func(int) (int, error) {
			close(firstStarted)
			<-firstReturn
			return 1, nil
		}),
		fTask(func(int) (int, error) {
			<-firstStarted
			close(firstReturn)
			<-secondReturn
			return 5, nil
		}),
	)
	go func() {
		routinesAtStart++ // for this one
		<-firstReturned
		var nowRoutines = runtime.NumGoroutine()
		if nowRoutines < routinesAtStart+1 {
			t.Errorf("Expected that there will be atleast 1 more goroutines than at the start(%d) after one of two fastest task finishes but got %d", routinesAtStart, nowRoutines)
		}
		close(secondReturn)
		time.Sleep(200 * time.Millisecond) // should be more than enough

		nowRoutines = runtime.NumGoroutine()
		if nowRoutines != routinesAtStart {
			endStacktraces := &bytes.Buffer{}
			pprof.Lookup("goroutine").WriteTo(endStacktraces, 1)

			t.Errorf("Expected that there will be as many goroutines as at the start(%d) after all tasks in Fastest have finishes but got %d"+
				"\n\nBEFORE:\n%s\n\nAFTER:\n%s\n",
				routinesAtStart, nowRoutines, startStacktraces.String(), endStacktraces.String())
		}
		close(end)

	}()
	if result, err := f.Execute(1); err != nil {
		t.Errorf("Received an unexpected error %s", err)
	} else if result != 1 {
		t.Errorf("Expected to receive 1 but received %d", result)
	}
	close(firstReturned)
	<-end
}

func TestTimedSimple(t *testing.T) {
	if res, err := Timed(errorTask("oops"), 100*time.Millisecond).Execute(5); err == nil {
		t.Errorf("Expected error did not occur instead got %d", res)
	}

	if res, err := Timed(lazyAdder{adder{20}, 50}, 1*time.Millisecond).Execute(5); err == nil {
		t.Errorf("Expected error did not occur instead got %d", res)
	}

	if res, err := Timed(lazyAdder{adder{20}, 30}, 100*time.Millisecond).Execute(2); err != nil {
		t.Error("Did not expect an error in err")
	} else if res != 22 {
		t.Errorf("Expected to receive 22 in res but received %d", res)
	}

	// This is intentional so the test below will work when run together
	time.Sleep(100 * time.Millisecond)
}

// This test may fail when run together with the other tests
// (for example via `go test ./...`). Evans runs the tests one by one, so it
// should not be a problem for the specific task
func TestTimedDoesntLeaveGoroutineHanging(t *testing.T) {
	var (
		started          = make(chan struct{})
		returning        = make(chan struct{})
		returned         = make(chan struct{})
		end              = make(chan struct{})
		routinesAtStart  = runtime.NumGoroutine()
		startStacktraces = &bytes.Buffer{}
	)
	pprof.Lookup("goroutine").WriteTo(startStacktraces, 1)

	f := Timed(
		fTask(func(int) (int, error) {
			close(started)
			<-returning
			return 1, nil
		}), 100*time.Millisecond)
	go func() {
		routinesAtStart++ // for this one
		<-started
		var nowRoutines = runtime.NumGoroutine()
		if nowRoutines != routinesAtStart+1 {
			t.Errorf("Expected that there will be 1 more goroutines than at the start(%d) after Timed one has started got %d", routinesAtStart, nowRoutines)
		}
		<-returned
		close(returning)
		time.Sleep(100 * time.Millisecond) // should be enough time
		nowRoutines = runtime.NumGoroutine()
		if nowRoutines != routinesAtStart {
			endStacktraces := &bytes.Buffer{}
			pprof.Lookup("goroutine").WriteTo(endStacktraces, 1)

			t.Errorf("Expected that there will be as many goroutines as at the start(%d) after Timed task has finished after it has timeouted but got %d"+
				"\n\nBEFORE:\n%s\n\nAFTER:\n%s\n",
				routinesAtStart, nowRoutines, startStacktraces.String(), endStacktraces.String())
		}
		close(end)

	}()
	if result, err := f.Execute(1); err == nil {
		t.Errorf("Expected error receibed %d", result)
	}
	close(returned)
	<-end
}

func reduceToLeast(results []int) int {
	smallest := 128
	for _, v := range results {
		if v < smallest {
			smallest = v
		}
	}
	return smallest
}

func TestConcurrentMapReduceFails(t *testing.T) {
	if res, err := ConcurrentMapReduce(reduceToLeast).Execute(1); err == nil {
		t.Errorf("Expected error did not occur instead got %d", res)
	}

	if res, err := ConcurrentMapReduce(reduceToLeast, adder{11}, errorTask("oops")).Execute(1); err == nil {
		t.Errorf("Expected error did not occur instead got %d", res)
	}
}

func TestConcurrentMapReduceSimple(t *testing.T) {
	cmr := ConcurrentMapReduce(reduceToLeast, adder{11}, adder{22}, adder{33})
	if res, err := cmr.Execute(44); err != nil {
		t.Errorf("Did not expected error but got %s", err)
	} else if res != 55 {
		t.Errorf("Expected result to be 55 but is %d", res)
	}
}

func greatestSearcherStuffer(tasks chan Task, fourthOops bool) {
	time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
	tasks <- adder{4}
	tasks <- lazyAdder{adder{22}, 20}
	tasks <- adder{127} // This is the first acceptable error
	time.Sleep(time.Duration(rand.Intn(30)) * time.Millisecond)
	tasks <- adder{32}           // This should be the winner
	tasks <- errorTask("Oops 2") // This is the second acceptable error
	tasks <- adder{5}
	time.Sleep(time.Duration(rand.Intn(30)) * time.Millisecond)
	tasks <- errorTask("Oops 3") // This is the third acceptable error
	tasks <- adder{-1}
	if fourthOops {
		tasks <- errorTask("Oops 4!") // This be the unacceptable error
	}
	close(tasks)
}

func TestGreatestSearcherSimple(t *testing.T) {
	tasks := make(chan Task)
	go func() {
		tasks <- adder{-2}
		tasks <- errorTask("oops")
		tasks <- adder{1}
		close(tasks)
	}()
	gs := GreatestSearcher(1, tasks)

	expResult := 2
	if result, err := gs.Execute(1); err != nil {
		t.Errorf("Received an unexpected error %s", err)
	} else if result != expResult {
		t.Errorf("Received result %d when expecting %d", result, expResult)
	}
}

func TestGreatestSearcherComplex(t *testing.T) {
	tasks := make(chan Task)
	gs := GreatestSearcher(3, tasks) // We accept 3 errors
	go greatestSearcherStuffer(tasks, false)

	expResult := 42
	if result, err := gs.Execute(10); err != nil {
		t.Errorf("Received an unexpected error %s", err)
	} else if result != expResult {
		t.Errorf("Received result %d when expecting %d", result, expResult)
	}
}

func TestGreatestSearcherErrors(t *testing.T) {
	t.Run("like the example", func(t *testing.T) {
		tasks := make(chan Task)
		gs := GreatestSearcher(3, tasks) // We accept 3 errors
		go greatestSearcherStuffer(tasks, true)

		if res, err := gs.Execute(10); err == nil {
			t.Errorf("Expected error did not occur instead got %d", res)
		}
	})
	t.Run("close immediately", func(t *testing.T) {
		tasks := make(chan Task)
		close(tasks)
		gs := GreatestSearcher(1, tasks)

		if res, err := gs.Execute(10); err == nil {
			t.Errorf("Expected error did not occur instead got %d", res)
		}
	})

	t.Run("only failure", func(t *testing.T) {
		tasks := make(chan Task)
		go func() {
			tasks <- errorTask("oops")
			close(tasks)
		}()
		gs := GreatestSearcher(1, tasks)

		if res, err := gs.Execute(10); err == nil {
			t.Errorf("Expected error did not occur instead got %d", res)
		}
	})
}

func TestThemAll(t *testing.T) {
	tasks := make(chan Task)
	go func() {
		tasks <- adder{10}
		time.Sleep(50 * time.Millisecond)
		tasks <- errorTask("oops")
		tasks <- adder{60}
		time.Sleep(50 * time.Millisecond)
		close(tasks)
	}()
	median := func(results []int) int {
		sort.Ints(results)
		return results[len(results)/2]
	}

	res, err := Pipeline(
		adder{10},
		lazyAdder{adder{20}, 10},
		Fastest(
			adder{30},
			lazyAdder{adder{300}, 30000},
			lazyAdder{adder{100}, 10000},
		),
		Timed(lazyAdder{adder{40}, 10}, 100*time.Millisecond),
		Timed(
			ConcurrentMapReduce(
				median,
				lazyAdder{adder{50}, 60},
				adder{1},
				lazyAdder{adder{50}, 60},
				GreatestSearcher(1, tasks),
			),
			200*time.Millisecond,
		),
	).Execute(-75)

	expResult := 75
	if err != nil {
		t.Errorf("Received an unexpected error %s", err)
	} else if res != expResult {
		t.Errorf("Received %d when expecting %d", res, expResult)
	}
}
