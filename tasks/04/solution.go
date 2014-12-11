package main

import (
	//"sync"
	"errors"
	//"fmt"
	"time"
	//"strconv"
	"io/ioutil"
	"net/http"
	//"strings"
)

func SeekAndDestroy(callback func(string) bool, chunkedUrlsToCheck <-chan []string, workersCount int) (string, error) {
	if workersCount == 0 {
		return "", errors.New("gg")
	}

	sem := make(chan struct{}, workersCount)
	resChan := make(chan string, 13)
	errorChan := make(chan error, 17)

	go func() {
		go func() {
			timer := time.NewTimer(time.Second * 15)
			<-timer.C
			// cancel the function
			errorChan <- errors.New("gg")
		}()
		for chunk := range chunkedUrlsToCheck {
			for i := range chunk {
				go func() {
					sem <- struct{}{}
					defer func() {
						<-sem
					}()

					url := chunk[i]

					// http request
					timeout := time.Duration(3 * time.Second)
					client := http.Client{
						Timeout: timeout,
					}
					response, err := client.Get(url)
					if err != nil {
						//errorChan <- err
						return
					}
					defer response.Body.Close()
					contents, err := ioutil.ReadAll(response.Body)
					if err != nil {
						//
						return
					}

					if callback(string(contents)) {
						resChan <- url
						return
					}
				}()
			}
		}

		// channel is closed, throw error
		errorChan <- errors.New("gg")
	}()

	select {
	case result := <-resChan:
		return result, nil
	case err := <-errorChan:
		return "", err
	}
}
