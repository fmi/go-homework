package main

import (
	"fmt"
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

func ExamplePipeline_first() {
	if res, err := Pipeline(adder{50}, adder{60}).Execute(10); err != nil {
		fmt.Printf("The pipeline returned an error\n")
	} else {
		fmt.Printf("The pipeline returned %d\n", res)
	}
	// Output:
	// The pipeline returned 120
}

func ExamplePipeline_second() {
	p := Pipeline(adder{20}, adder{10}, adder{-50})
	if res, err := p.Execute(100); err != nil {
		fmt.Printf("The pipeline returned an error\n")
	} else {
		fmt.Printf("The pipeline returned %d\n", res)
	}
	// Output:
	// The pipeline returned an error
}

type lazyAdder struct {
	adder
	delay time.Duration
}

func (la lazyAdder) Execute(addend int) (int, error) {
	time.Sleep(la.delay * time.Millisecond)
	return la.adder.Execute(addend)
}

func TestFastest(t *testing.T) {
	f := Fastest(
		lazyAdder{adder{20}, 500},
		lazyAdder{adder{50}, 300},
		adder{41},
	)
	if result, err := f.Execute(1); err != nil {
		t.Errorf("Received an unexpected error %s", err)
	} else if result != 42 {
		t.Errorf("Expected to receive 42 but received %d", result)
	}
}

func TestTimed(t *testing.T) {
	_, e1 := Timed(lazyAdder{adder{20}, 50}, 2*time.Millisecond).Execute(2)
	if e1 == nil {
		t.Error("Expected an error in e1")
	}
	r2, e2 := Timed(lazyAdder{adder{20}, 50}, 300*time.Millisecond).Execute(2)
	if e2 != nil {
		t.Error("Did not expect an error in e2")
	}
	if r2 != 22 {
		t.Errorf("Expected to receive 22 in r2 but received %d", r2)
	}
}

func ExampleConcurrentMapReduce() {
	reduce := func(results []int) int {
		smallest := 128
		for _, v := range results {
			if v < smallest {
				smallest = v
			}
		}
		return smallest
	}

	mr := ConcurrentMapReduce(reduce, adder{30}, adder{50}, adder{20})
	if res, err := mr.Execute(5); err != nil {
		fmt.Printf("We got an error!\n")
	} else {
		fmt.Printf("The ConcurrentMapReduce returned %d\n", res)
	}
	// Output:
	// The ConcurrentMapReduce returned 25
}

func TestGreatestSearcherSimple(t *testing.T) {
	tasks := make(chan Task)
	gs := GreatestSearcher(2, tasks) // We accept two errors

	go func() {
		tasks <- adder{4}
		tasks <- lazyAdder{adder{22}, 20}
		tasks <- adder{125} // This should be the first acceptable error
		time.Sleep(50 * time.Millisecond)
		tasks <- adder{32} // This should be the winner
		// This should time out be the second acceptable error
		tasks <- Timed(lazyAdder{adder{100}, 2000}, 20*time.Millisecond)
		// If we uncomment this, the whole gs.Execute() should return an error
		// tasks <- adder{127} // third unacceptable error
		close(tasks)
	}()

	expResult := 42
	result, err := gs.Execute(10)
	if err != nil {
		t.Errorf("Received an unexpected error %s", err)
	} else if result != expResult {
		t.Errorf("Received result %d when expecting %d", result, expResult)
	}
}
