package main

import (
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"
)

func thisShouldHappenInLessThan(period time.Duration, t *testing.T, errorMessage string, action func()) {
	finished := make(chan bool)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Paniced during the test with message '%s'", r)
			}
			close(finished)
		}()

		action()
	}()

	select {
	case <-finished:
	case <-time.After(period):
		t.Errorf("Test exceeded allowed time of %d seconds: %s", period/time.Second, errorMessage)
	}
	return
}

type testPage struct {
	Path         string
	ResponseCode int
	Latency      time.Duration
	Contents     string
}

func createTestServer(t *testing.T, chunksOrTimeouts ...interface{}) (baseAddr string, urlsChan chan []string, cleanupFunction func()) {

	// Prepare a transient local server
	listener, err := net.Listen("tcp4", "127.0.0.2:0") // Port 0 means that the OS will allocate a random free port
	if err != nil {
		panic("Could not listen to a local port!?")
	}

	baseAddr = fmt.Sprintf("http://%s/", listener.Addr())
	cleanupFunction = func() {
		listener.Close()
	}
	mux := http.NewServeMux()
	go http.Serve(listener, mux)

	handleTestPage := func(page testPage) string {
		pageAddr := baseAddr + page.Path

		mux.HandleFunc("/"+page.Path, func(w http.ResponseWriter, req *http.Request) {
			time.Sleep(page.Latency)
			w.WriteHeader(page.ResponseCode)
			fmt.Fprintf(w, page.Contents)
		})

		return pageAddr
	}

	// Create and then populate the urls channel
	urlsChan = make(chan []string)
	go func() {
		for _, nextMysteryItem := range chunksOrTimeouts {
			switch chunkOrTimeout := nextMysteryItem.(type) {

			case []interface{}: // That's a chunk
				urlsToPass := make([]string, len(chunkOrTimeout))

				for i, mysteryChunkPiece := range chunkOrTimeout {
					switch stringOrTestPage := mysteryChunkPiece.(type) {

					case string: // This is a simple string URL, pass it directly
						urlsToPass[i] = stringOrTestPage

					case testPage: // This is a custom web page, add it to the local test http server
						urlsToPass[i] = handleTestPage(stringOrTestPage)

					default:
						panic("How did you manage to mess this up...")
					}
				}

				thisShouldHappenInLessThan(1*time.Second, t, "sending new urls to the channel should not be blocked", func() {
					urlsChan <- urlsToPass
				})

			case time.Duration: // That's a timeout, duh...
				time.Sleep(chunkOrTimeout)

			default:
				panic("Breaking the tests... way to go, dude...")
			}
		}
	}()

	return
}

func alwaysMatch(string) bool {
	return true
}

func neverMatch(string) bool {
	return false
}

func getSubstrMatcher(needle string) func(string) bool {
	return func(haystack string) bool {
		return strings.Contains(haystack, needle)
	}
}

func getRegexMatcher(pattern string) func(string) bool {
	regexPattern := regexp.MustCompile(pattern)

	return func(contents string) bool {
		return regexPattern.MatchString(contents)
	}
}

func checkForWrongParameters(t *testing.T, checkDescription string, callback func(string) bool, urls <-chan []string, workersCount int) {
	timeoutMessage := fmt.Sprintf("parameter errors should be immediately returned (%s)", checkDescription)

	thisShouldHappenInLessThan(1*time.Second, t, timeoutMessage, func() {
		if _, err := SeekAndDestroy(callback, urls, workersCount); err == nil {
			t.Errorf("Function should return error when %s", checkDescription)
		}
	})
}

func TestWithNegativeWorkersCount(t *testing.T) {
	t.Parallel()
	checkForWrongParameters(t, "workersCount is negative", alwaysMatch, make(chan []string), -1)
}

func TestWithZeroWorkersCount(t *testing.T) {
	t.Parallel()
	checkForWrongParameters(t, "workersCount is zero", alwaysMatch, make(chan []string), 0)
}

func TestWithInvalidCallback(t *testing.T) {
	t.Parallel()
	checkForWrongParameters(t, "callback is nil", nil, make(chan []string), 1)
}

func TestWithNilChannel(t *testing.T) {
	t.Parallel()
	checkForWrongParameters(t, "channel is uninitialized", alwaysMatch, nil, 3)
}

func TestWithClosedChannelWhenStarting(t *testing.T) {
	t.Parallel()
	oops := make(chan []string)
	close(oops)
	checkForWrongParameters(t, "the urls channel was closed", alwaysMatch, oops, 2)
}

func TestWithClosedChannelMidway(t *testing.T) {
	t.Parallel()
	aboutToBeClosed := make(chan []string)

	go func() {
		time.Sleep(5 * time.Second)
		close(aboutToBeClosed)
	}()

	thisShouldHappenInLessThan(7*time.Second, t, "the urls channel was closed after 5 seconds", func() {

		_, err := SeekAndDestroy(alwaysMatch, aboutToBeClosed, 4)

		if err == nil {
			t.Errorf("Function should have returned an error when urls channel was closed")
		}
	})
}

func TestWhetherGlobalTimeoutIsHandled(t *testing.T) {
	t.Parallel()

	thisShouldHappenInLessThan(17*time.Second, t, "the global timeout should have happened in 15 seconds", func() {

		_, err := SeekAndDestroy(alwaysMatch, make(chan []string), 5)

		if err == nil {
			t.Errorf("Function should have returned an error when the global timeout was reached")
		}
	})
}

//TODO: test whether worker number exceeds limit

func TestWithLoremIpsum(t *testing.T) {
	t.Parallel()
	thisShouldHappenInLessThan(4*time.Second, t, "Connecting to localhost should be pretty fast...", func() {

		baseAddr, urlsChan, cleanupFunc := createTestServer(t,
			[]interface{}{
				testPage{Path: "testpage1", ResponseCode: 200, Latency: 0 * time.Second, Contents: "just a simple test page"},
				testPage{Path: "testpage2", ResponseCode: 200, Latency: 1 * time.Second, Contents: "another simple test page"},
			},
			[]interface{}{
				testPage{Path: "lorem_ipsum", ResponseCode: 200, Latency: 2 * time.Second, Contents: "How about some Lorem   Ipsum??"},
			},
		)
		defer cleanupFunc()
		expectedResult := baseAddr + "lorem_ipsum"

		res, err := SeekAndDestroy(getRegexMatcher("(?i)lorem\\s*ipsum"), urlsChan, 3)

		if err != nil {
			t.Errorf("Function should not return error when valid data is present")
		}

		if res != expectedResult {
			t.Errorf("Function returned '%s' when it should have returned '%s'", res, expectedResult)
		}
	})
}

func TestIfTimeoutAndErrorCodesAreHonoured(t *testing.T) {
	t.Parallel()
	thisShouldHappenInLessThan(10*time.Second, t, "This should have finished in approx. 8 seconds", func() {

		baseAddr, urlsChan, cleanupFunc := createTestServer(t,
			[]interface{}{
				testPage{Path: "page_over_3_seconds", ResponseCode: 200, Latency: 5 * time.Second, Contents: "Time travel. You just want to slap a hippie, but all you get is multiple Kowalskis."},
				testPage{Path: "page_with_error_code", ResponseCode: 400, Latency: 0 * time.Second, Contents: "Evasive maneuvers, boys!"},
				"this.is.just.a.simple.invalid.url....",
			},
			time.Duration(7*time.Second),
			[]interface{}{
				testPage{Path: "correct_page", ResponseCode: 200, Latency: 1 * time.Second, Contents: "Just smile and wave, boys. Smile and wave..."},
			},
		)
		defer cleanupFunc()
		expectedResult := baseAddr + "correct_page"

		res, err := SeekAndDestroy(alwaysMatch, urlsChan, 2)

		if err != nil {
			t.Errorf("Function should not return error when valid data is present")
		}

		if res != expectedResult {
			t.Errorf("Function returned '%s' when it should have returned '%s'", res, expectedResult)
		}
	})
}

func TestRaceCondition(t *testing.T) {
	t.Parallel()
	thisShouldHappenInLessThan(5*time.Second, t, "This should have finished in approx. 3 seconds", func() {

		baseAddr, urlsChan, cleanupFunc := createTestServer(t,
			[]interface{}{
				testPage{Path: "intro", ResponseCode: 200, Latency: 0 * time.Second, Contents: "High five, low five!"},
				testPage{Path: "slow_success", ResponseCode: 200, Latency: 2 * time.Second, Contents: "Down low, too slow!"},
				testPage{Path: "fast_success", ResponseCode: 200, Latency: 1 * time.Second, Contents: "I think our work here is done."},
			},
		)
		defer cleanupFunc()
		expectedResult := baseAddr + "fast_success"

		res, err := SeekAndDestroy(getRegexMatcher("Down|done"), urlsChan, 2)

		if err != nil {
			t.Errorf("Function should not return error when valid data is present")
		}

		if res != expectedResult {
			t.Errorf("Function returned '%s' when it should have returned '%s'", res, expectedResult)
		}
	})
}

func TestCloseChannelBeforeFinish(t *testing.T) {
	t.Parallel()
	thisShouldHappenInLessThan(3*time.Second, t, "This should have finished in approx. 1 second", func() {

		_, urlsChan, cleanupFunc := createTestServer(t,
			[]interface{}{
				testPage{Path: "irrelevant", ResponseCode: 200, Latency: 2 * time.Second, Contents: "Status report, Kowalski."},
			},
		)
		defer cleanupFunc()
		go func() {
			time.Sleep(1 * time.Second)
			close(urlsChan)
		}()

		_, err := SeekAndDestroy(alwaysMatch, urlsChan, 2)

		if err == nil {
			t.Errorf("Function should have returned an error, channel was closed before it could finish")
		}
	})
}

/*
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
*/
