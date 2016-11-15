package main

// ConcurrentRetryExecutor is function. Shut up golint
func ConcurrentRetryExecutor(tasks []func() string, concurrentLimit int, retryLimit int) <-chan struct {
	index  int
	result string
} {
	var resultCh = make(chan struct {
		index  int
		result string
	})
	var limiter = make(chan struct{}, concurrentLimit)

	handleTask := func(index int, task func() string) {
		for i := 0; i < retryLimit; i++ {
			response := struct {
				index  int
				result string
			}{index, task()}
			resultCh <- response
			if response.result != "" {
				break
			}
		}
		<-limiter
	}

	go func() {
		for index, task := range tasks {
			limiter <- struct{}{}
			go handleTask(index, task)
		}
		for i := 0; i < concurrentLimit; i++ {
			limiter <- struct{}{}
		}
		close(resultCh)
	}()

	return resultCh
}
