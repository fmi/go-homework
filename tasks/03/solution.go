package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// Task is a simple interface for executable tasks and jobs
type Task interface {
	Execute(int) (int, error)
}

type pipeline []Task

// Pipeline executes the provided tasks sequentially
func Pipeline(tasks ...Task) Task {
	return pipeline(tasks)
}

func (p pipeline) Execute(initial int) (result int, err error) {
	if len(p) == 0 {
		return 0, errors.New("No tasks to execute")
	}

	result = initial
	for _, t := range p {
		result, err = t.Execute(result)
		if err != nil {
			return
		}
	}
	return
}

type fastest []Task

// Fastest executes the provided tasks concurrently and returns the first result
func Fastest(tasks ...Task) Task {
	return fastest(tasks)
}

func (f fastest) Execute(arg int) (result int, err error) {
	if len(f) == 0 {
		return 0, errors.New("No tasks to execute")
	}

	var once sync.Once
	var wg sync.WaitGroup
	wg.Add(1)

	for _, t := range f {
		go func(task Task) {
			r, e := task.Execute(arg)
			once.Do(func() {
				result = r
				err = e
				wg.Done()
			})
		}(t)
	}

	wg.Wait()
	return
}

type timed struct {
	task    Task
	timeout time.Duration
}

// Timed implements timeoutable tasks
func Timed(task Task, timeout time.Duration) Task {
	return &timed{task, timeout}
}

func (t timed) Execute(arg int) (int, error) {
	finish := make(chan struct{})
	var result int
	var err error

	go func() {
		result, err = t.task.Execute(arg)
		close(finish)
	}()

	select {
	case <-finish:
		return result, err
	case <-time.After(t.timeout):
		return 0, errors.New("Timeout")
	}
}

type concurrentMapReduce struct {
	reduce func(results []int) int
	tasks  []Task
}

// ConcurrentMapReduce does what the name says...
func ConcurrentMapReduce(reduce func(results []int) int, tasks ...Task) Task {
	return &concurrentMapReduce{reduce, tasks}
}

func (c concurrentMapReduce) Execute(arg int) (int, error) {
	if len(c.tasks) == 0 {
		return 0, errors.New("No tasks to execute")
	}

	results := make([]int, 0, len(c.tasks))
	finish := make(chan struct{})
	resultCh := make(chan int)
	errorCh := make(chan error)

	for _, t := range c.tasks {
		go func(task Task) {
			if res, err := task.Execute(arg); err != nil {
				select {
				case errorCh <- err:
				case <-finish:
				}
			} else {
				select {
				case resultCh <- res:
				case <-finish:
				}
			}
		}(t)
	}

	for {
		select {
		case err := <-errorCh:
			close(finish)
			return 0, err
		case res := <-resultCh:
			results = append(results, res)
			if len(results) == len(c.tasks) {
				return c.reduce(results), nil
			}
		}
	}
}

type greatestSearcher struct {
	errorLimit int
	tasks      <-chan Task
}

// GreatestSearcher greedily and concurrently executes the supplied tasks and
// returns the greatest result (if the error limit is not exceeded)
func GreatestSearcher(errorLimit int, tasks <-chan Task) Task {
	return &greatestSearcher{errorLimit, tasks}
}

func (g greatestSearcher) Execute(arg int) (int, error) {
	resultCh := make(chan int)
	errorCh := make(chan error)
	executeTask := func(task Task) {
		if res, err := task.Execute(arg); err != nil {
			errorCh <- err
		} else {
			resultCh <- res
		}
	}

	tasksIsOpen := true
	executingCount := 0
	errorCount := 0
	var greatest *int
	for tasksIsOpen {
		select {
		case newTask, ok := <-g.tasks:
			if ok {
				executingCount++
				go executeTask(newTask)
			} else {
				tasksIsOpen = false
			}
		case <-errorCh:
			executingCount--
			errorCount++
		case res := <-resultCh:
			executingCount--
			if greatest == nil || *greatest < res {
				greatest = &res
			}
		}
	}

	for executingCount > 0 {
		select {
		case <-errorCh:
			executingCount--
			errorCount++

		case res := <-resultCh:
			executingCount--
			if greatest == nil || *greatest < res {
				greatest = &res
			}
		}
	}

	if errorCount > g.errorLimit {
		return 0, fmt.Errorf("Encountered %d errors with a limit of %d", errorCount, g.errorLimit)
	}
	if greatest == nil {
		return 0, errors.New("No tasks finished successfully")
	}
	return *greatest, nil
}
