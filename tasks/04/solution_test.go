package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"testing/iotest"
	"time"
)

const noValidURLsInsideTest = "no valid urls"

const nonExistingURLInsideTest = "http://some.non.existing.domain.at.nowhere/pesho"

func parseRangeInsideTest(s string) (start, end int) {
	fmt.Sscanf(s, "bytes=%d-%d", &start, &end)
	return
}

func responseRangeHeaderValueInsideTest(start, end, size int) string {
	return fmt.Sprintf("bytes %d-%d/%d", start, end, size)
}

// Test nothing to return
func TestNothingIsReturned(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add("Content-Length", "0")
		w.WriteHeader(200)
	}))
	defer s.Close()

	r := DownloadFile(context.TODO(), []string{s.URL + "/pesho"})
	var buf [512]byte
	n, err := io.ReadFull(r, buf[:])
	if n != 0 {
		t.Errorf("Expected to read 0 bytes from empty download but got %d", n)
	}

	if err != io.EOF {
		t.Errorf("Expected to get error '%s', but got '%s'", io.EOF, err)
	}
}

func checkResponseInsideTest(t *testing.T, expected, got []byte) {
	if !bytes.Equal(expected, got) {
		t.Errorf("Expected result was '%s' but got '%s'", hex.EncodeToString(expected), hex.EncodeToString(got))
	}
}

// Test simple case with one url
func TestSingleURLWithReturn(t *testing.T) {
	var resp = []byte("This IS the most epic of all response")
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodHead:
			w.Header().Add("Content-Length", strconv.Itoa(len(resp)))
			w.WriteHeader(200)
		case http.MethodGet:
			var currentResp = resp
			var statusCode = 200
			if _, ok := req.Header["Range"]; ok {
				requestRange := req.Header.Get("Range")
				start, end := parseRangeInsideTest(requestRange)
				currentResp = resp[start : end+1]
				w.Header().Add("Content-Range", responseRangeHeaderValueInsideTest(start, end, len(resp)))
				statusCode = 206
			}
			w.Header().Add("Content-Length", strconv.Itoa(len(currentResp)))
			w.WriteHeader(statusCode)
			n, err := w.Write(currentResp)
			if n != len(currentResp) {
				t.Errorf("Wrote %d not %d as expected", n, len(currentResp))
			}
			if err != nil {
				t.Errorf("Got error while writing response")
			}
		}
	}))
	defer s.Close()

	r := DownloadFile(context.Background(), []string{s.URL + "/pesho"})
	var buf bytes.Buffer
	n, err := buf.ReadFrom(r)
	if int(n) != len(resp) {
		t.Errorf("Expected to read %d bytes from simple download but got %d", len(resp), n)
	}

	if err != nil {
		t.Errorf("Expected to get no error, but got '%s'", err)
	}

	checkResponseInsideTest(t, resp, buf.Bytes()[:n])
}

// Test simple case with one url and wait to return any bytes until DownloadFile returns the reader
func TestSingleURLBlockUntilDownloadFileReturns(t *testing.T) {
	var resp = []byte("This IS the most epic of all response")
	var block = make(chan struct{})
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		<-block
		switch req.Method {
		case http.MethodHead:
			w.Header().Add("Content-Length", strconv.Itoa(len(resp)))
			w.WriteHeader(200)
		case http.MethodGet:
			var currentResp = resp
			var statusCode = 200
			if _, ok := req.Header["Range"]; ok {
				requestRange := req.Header.Get("Range")
				start, end := parseRangeInsideTest(requestRange)
				currentResp = resp[start : end+1]
				w.Header().Add("Content-Range", responseRangeHeaderValueInsideTest(start, end, len(resp)))
				statusCode = 206
			}
			w.Header().Add("Content-Length", strconv.Itoa(len(currentResp)))
			w.WriteHeader(statusCode)
			n, err := w.Write(currentResp)
			if n != len(currentResp) {
				t.Errorf("Wrote %d not %d as expected", n, len(currentResp))
			}
			if err != nil {
				t.Errorf("Got error while writing response")
			}
		}
	}))
	defer s.Close()

	r := DownloadFile(context.Background(), []string{s.URL + "/pesho"})
	close(block)
	var buf bytes.Buffer
	n, err := buf.ReadFrom(r)
	if int(n) != len(resp) {
		t.Errorf("Expected to read %d bytes from simple download but got %d", len(resp), n)
	}

	if err != nil {
		t.Errorf("Expected to get no error, but got '%s'", err)
	}

	checkResponseInsideTest(t, resp, buf.Bytes()[:n])
}

// Test simple case with one url which will return half the bytes it was requested the first time
// and cancel the context after that while returning empty result
func TestSingleURLCancelContextAfterHalfBytesWereServed(t *testing.T) {
	var resp = []byte("This IS the most epic of all response")
	var byteCount int32
	var ctx, cancelFunc = context.WithCancel(context.Background())
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodHead:
			w.Header().Add("Content-Length", strconv.Itoa(len(resp)))
			w.WriteHeader(200)
		case http.MethodGet:
			if atomic.LoadInt32(&byteCount) > 0 {
				cancelFunc()
				return
			}
			var currentResp = resp
			var statusCode = 200
			if _, ok := req.Header["Range"]; ok {
				requestRange := req.Header.Get("Range")
				start, end := parseRangeInsideTest(requestRange)
				currentResp = resp[start : end+1]
				w.Header().Add("Content-Range", responseRangeHeaderValueInsideTest(start, end, len(resp)))
				statusCode = 206
			}
			currentResp = currentResp[:len(currentResp)/2]
			atomic.AddInt32(&byteCount, int32(len(currentResp)))

			w.Header().Add("Content-Length", strconv.Itoa(len(currentResp)))
			w.WriteHeader(statusCode)
			n, err := w.Write(currentResp)
			if n != len(currentResp) {
				t.Errorf("Wrote %d not %d as expected", n, len(currentResp))
			}
			if err != nil {
				t.Errorf("Got error while writing response")
			}
		}
	}))
	defer s.Close()

	r := DownloadFile(ctx, []string{s.URL + "/pesho"})
	var buf bytes.Buffer
	n, err := buf.ReadFrom(r)
	if int32(n) != byteCount {
		t.Errorf("Expected to read %d bytes from simple download but got %d", byteCount, n)
	}

	if err != context.Canceled {
		t.Errorf("Expected to get error %s, but got '%s'", context.Canceled, err)
	}

	checkResponseInsideTest(t, resp[:byteCount], buf.Bytes()[:n])
}

// Test no valid urls
func TestNoValidUrls(t *testing.T) {
	r := DownloadFile(context.Background(), []string{nonExistingURLInsideTest})
	var buf bytes.Buffer
	n, err := buf.ReadFrom(r)
	if int(n) != 0 {
		t.Errorf("Expected to read 0 bytes from invalid url but got %d", n)
	}

	if err == nil {
		t.Errorf("Expected to get error, but got none")
	}
}

// Test only first 10 bytes
func TestReturnOnly10Bytes(t *testing.T) {
	var resp = []byte("This IS the most epic of all response")
	var bytesToReturn = 10
	var called int32
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodHead:
			w.Header().Add("Content-Length", strconv.Itoa(len(resp)))
			w.WriteHeader(200)
		case http.MethodGet:
			if atomic.CompareAndSwapInt32(&called, 0, 1) {
				var currentResp = resp
				var statusCode = 200
				if _, ok := req.Header["Range"]; ok {
					requestRange := req.Header.Get("Range")
					start, end := parseRangeInsideTest(requestRange)
					if start != 0 {
						t.Errorf("Expected to get one request from 0 to end got start %d", start)
					}
					currentResp = resp[start : end+1]
					w.Header().Add("Content-Range", responseRangeHeaderValueInsideTest(start, end, len(resp)))
					statusCode = 206
				}
				w.Header().Add("Content-Length", strconv.Itoa(len(currentResp)))
				w.WriteHeader(statusCode)
				// write only some bytes on purpose
				n, err := w.Write(currentResp[:bytesToReturn])
				if n != bytesToReturn {
					t.Errorf("Wrote %d not %d as expected", n, bytesToReturn)
				}
				if err != nil {
					t.Errorf("Got error while writing response")
				}
				w.(http.Flusher).Flush()
			} else {
				w.WriteHeader(500)
			}
		}
	}))
	defer s.Close()

	r := DownloadFile(context.Background(), []string{s.URL + "/pesho"})
	var buf bytes.Buffer
	n, err := buf.ReadFrom(r)
	if int(n) != bytesToReturn {
		t.Errorf("Expected to read %d bytes from simple download but got %d", len(resp), n)
	}

	if err == nil || err.Error() != noValidURLsInsideTest {
		t.Errorf("Expected to get  error with message '%s', but got '%s'", noValidURLsInsideTest, err)
	}

	checkResponseInsideTest(t, resp[:bytesToReturn], buf.Bytes()[:n])
}

// Test simple case with two urls
func TestTwoUrlsEverythingAsExpected(t *testing.T) {
	var resp = []byte("This IS the most epic of all response")
	var byteCount1, byteCount2 int32
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodHead:
			w.Header().Add("Content-Length", strconv.Itoa(len(resp)))
			w.WriteHeader(200)
		case http.MethodGet:
			var statusCode = 200
			var currentResp = resp
			if _, ok := req.Header["Range"]; ok {
				requestRange := req.Header.Get("Range")
				start, end := parseRangeInsideTest(requestRange)
				currentResp = resp[start : end+1]
				w.Header().Add("Content-Range", responseRangeHeaderValueInsideTest(start, end, len(resp)))
				statusCode = 206
			}
			switch req.URL.Path {
			case "/pesho":
				atomic.AddInt32(&byteCount1, int32(len(currentResp)))
			case "/pesho2":
				atomic.AddInt32(&byteCount2, int32(len(currentResp)))
			}

			w.Header().Add("Content-Length", strconv.Itoa(len(currentResp)))
			w.WriteHeader(statusCode)
			n, err := w.Write(currentResp)
			if n != len(currentResp) {
				t.Errorf("Wrote %d not %d as expected", n, len(currentResp))
			}
			if err != nil {
				t.Errorf("Got error while writing response")
			}
		}
	}))
	defer s.Close()

	r := DownloadFile(context.Background(), []string{s.URL + "/pesho", s.URL + "/pesho2"})
	var buf bytes.Buffer
	n, err := buf.ReadFrom(r)
	if int(n) != len(resp) {
		t.Errorf("Expected to read %d bytes from simple download but got %d", len(resp), n)
	}

	if err != nil {
		t.Errorf("Expected to get no error, but got '%s'", err)
	}

	checkResponseInsideTest(t, resp, buf.Bytes()[:n])
	if int64(byteCount1+byteCount2) != n {
		t.Errorf("It was expected that the downloaded amount from both urls "+
			" will be equal to the size of the request but while it was %d, "+
			"from url[0] was downloaded %d and from url[1] %d", n, byteCount1, byteCount2)
	}

	if abs32InsideTest(byteCount1-byteCount2) > 1 {
		t.Errorf("Both urls should have been used equally but %d were downloaded "+
			"from url[0] and %d from url[1]", byteCount1, byteCount2)
	}
}

// Test simple case with two urls with nil context
func TestTwoUrlsWithNilContext(t *testing.T) {
	var resp = []byte("This IS the most epic of all response")
	var byteCount1, byteCount2 int32
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodHead:
			w.Header().Add("Content-Length", strconv.Itoa(len(resp)))
			w.WriteHeader(200)
		case http.MethodGet:
			var statusCode = 200
			var currentResp = resp
			if _, ok := req.Header["Range"]; ok {
				requestRange := req.Header.Get("Range")
				start, end := parseRangeInsideTest(requestRange)
				currentResp = resp[start : end+1]
				w.Header().Add("Content-Range", responseRangeHeaderValueInsideTest(start, end, len(resp)))
				statusCode = 206
			}
			switch req.URL.Path {
			case "/pesho":
				atomic.AddInt32(&byteCount1, int32(len(currentResp)))
			case "/pesho2":
				atomic.AddInt32(&byteCount2, int32(len(currentResp)))
			}

			w.Header().Add("Content-Length", strconv.Itoa(len(currentResp)))
			w.WriteHeader(statusCode)
			n, err := w.Write(currentResp)
			if n != len(currentResp) {
				t.Errorf("Wrote %d not %d as expected", n, len(currentResp))
			}
			if err != nil {
				t.Errorf("Got error while writing response")
			}
		}
	}))
	defer s.Close()

	r := DownloadFile(nil, []string{s.URL + "/pesho", s.URL + "/pesho2"})
	var buf bytes.Buffer
	n, err := buf.ReadFrom(r)
	if int(n) != len(resp) {
		t.Errorf("Expected to read %d bytes from simple download but got %d", len(resp), n)
	}

	if err != nil {
		t.Errorf("Expected to get no error, but got '%s'", err)
	}

	checkResponseInsideTest(t, resp, buf.Bytes()[:n])
	if int64(byteCount1+byteCount2) != n {
		t.Errorf("It was expected that the downloaded amount from both urls "+
			" will be equal to the size of the request but while it was %d, "+
			"from url[0] was downloaded %d and from url[1] %d", n, byteCount1, byteCount2)
	}

	if abs32InsideTest(byteCount1-byteCount2) > 1 {
		t.Errorf("Both urls should have been used equally but %d were downloaded "+
			"from url[0] and %d from url[1]", byteCount1, byteCount2)
	}
}

// Test simple case with two urls one of which doesn't respond from the start
func TestTwoUrlsWithOneBroken(t *testing.T) {
	var resp = []byte("This IS the most epic of all response")
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodHead:
			w.Header().Add("Content-Length", strconv.Itoa(len(resp)))
			w.WriteHeader(200)
		case http.MethodGet:
			var statusCode = 200
			var currentResp = resp
			if _, ok := req.Header["Range"]; ok {
				requestRange := req.Header.Get("Range")
				start, end := parseRangeInsideTest(requestRange)
				currentResp = resp[start : end+1]
				w.Header().Add("Content-Range", responseRangeHeaderValueInsideTest(start, end, len(resp)))
				statusCode = 206
			}
			w.Header().Add("Content-Length", strconv.Itoa(len(currentResp)))
			w.WriteHeader(statusCode)
			n, err := w.Write(currentResp)
			if n != len(currentResp) {
				t.Errorf("Wrote %d not %d as expected", n, len(currentResp))
			}
			if err != nil {
				t.Errorf("Got error while writing response")
			}
		}
	}))
	defer s.Close()

	r := DownloadFile(context.Background(), []string{s.URL + "/pesho", nonExistingURLInsideTest})
	var buf bytes.Buffer
	n, err := buf.ReadFrom(r)
	if int(n) != len(resp) {
		t.Errorf("Expected to read %d bytes from simple download but got %d", len(resp), n)
	}

	if err != nil {
		t.Errorf("Expected to get no error, but got '%s'", err)
	}

	checkResponseInsideTest(t, resp, buf.Bytes()[:n])
}

// Test simple case with two urls one of which doesn't respond from the start
func TestTwoUrlsWithTheOtherOneBroken(t *testing.T) {
	var resp = []byte("This IS the most epic of all response")
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodHead:
			w.Header().Add("Content-Length", strconv.Itoa(len(resp)))
			w.WriteHeader(200)
		case http.MethodGet:
			var statusCode = 200
			var currentResp = resp
			if _, ok := req.Header["Range"]; ok {
				requestRange := req.Header.Get("Range")
				start, end := parseRangeInsideTest(requestRange)
				currentResp = resp[start : end+1]
				w.Header().Add("Content-Range", responseRangeHeaderValueInsideTest(start, end, len(resp)))
				statusCode = 206
			}
			w.Header().Add("Content-Length", strconv.Itoa(len(currentResp)))
			w.WriteHeader(statusCode)
			n, err := w.Write(currentResp)
			if n != len(currentResp) {
				t.Errorf("Wrote %d not %d as expected", n, len(currentResp))
			}
			if err != nil {
				t.Errorf("Got error while writing response")
			}
		}
	}))
	defer s.Close()

	r := DownloadFile(context.Background(), []string{nonExistingURLInsideTest, s.URL + "/pesho", ""})
	var buf bytes.Buffer
	n, err := buf.ReadFrom(r)
	if int(n) != len(resp) {
		t.Errorf("Expected to read %d bytes from simple download but got %d", len(resp), n)
	}

	if err != nil {
		t.Errorf("Expected to get no error, but got '%s'", err)
	}

	checkResponseInsideTest(t, resp, buf.Bytes()[:n])
}

// Test simple case with two urls one of which stops responding after the first 5 bytes
func TestTwoUrlsOneStopRespondingAfter5bytes(t *testing.T) {
	var resp = []byte("This IS the most epic of all response")
	var returnMaxBytes = 5
	var byteCount1, byteCount2 int32
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodHead:
			w.Header().Add("Content-Length", strconv.Itoa(len(resp)))
			w.WriteHeader(200)
		case http.MethodGet:
			var statusCode = 200
			var currentResp = resp
			if _, ok := req.Header["Range"]; ok {
				requestRange := req.Header.Get("Range")
				start, end := parseRangeInsideTest(requestRange)
				currentResp = resp[start : end+1]
				w.Header().Add("Content-Range", responseRangeHeaderValueInsideTest(start, end, len(resp)))
				statusCode = 206
			}
			switch req.URL.Path {
			case "/pesho":
				currentResp = currentResp[:returnMaxBytes]
				if !atomic.CompareAndSwapInt32(&byteCount1, 0, int32(len(currentResp))) {
					w.WriteHeader(500)
					return
				}
			case "/pesho2":
				atomic.AddInt32(&byteCount2, int32(len(currentResp)))
			}

			w.Header().Add("Content-Length", strconv.Itoa(len(currentResp)))
			w.WriteHeader(statusCode)
			n, err := w.Write(currentResp)
			if n != len(currentResp) {
				t.Errorf("Wrote %d not %d as expected", n, len(currentResp))
			}
			if err != nil {
				t.Errorf("Got error while writing response")
			}
		}
	}))

	defer s.Close()

	r := DownloadFile(context.Background(), []string{s.URL + "/pesho", s.URL + "/pesho2"})
	var buf bytes.Buffer
	n, err := buf.ReadFrom(r)
	if int(n) != len(resp) {
		t.Errorf("Expected to read %d bytes from simple download but got %d", len(resp), n)
	}

	if err != nil {
		t.Errorf("Expected to get no error, but got '%s'", err)
	}

	checkResponseInsideTest(t, resp, buf.Bytes()[:n])

	if int64(byteCount1+byteCount2) != n {
		t.Errorf("It was expected that the downloaded amount from both urls "+
			" will be equal to the size of the request but while it was %d, "+
			"from url[0] was downloaded %d and from url[1] %d", n, byteCount1, byteCount2)
	}

	if int(byteCount1) != returnMaxBytes {
		t.Errorf("It was expected that url[0] would serve %d but have served %d ",
			returnMaxBytes, byteCount1)
	}

	if int(byteCount2) != len(resp)-returnMaxBytes {
		t.Errorf("It was expected that url[1] would serve %d but have served %d ",
			len(resp)-returnMaxBytes, byteCount2)
	}
}

// Test case with three urls, one of which returns only 2 bytes, another returns 1 byte at a time
func TestThreeUrlsOneOfWhichReturns1byteAtATimeOneOfWhichBreaksAfter2bytes(t *testing.T) {
	var resp = []byte("This IS the most epic of all response")
	var returnMaxBytes = 2
	var byteCount1, byteCount2, byteCount3 int32
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodHead:
			w.Header().Add("Content-Length", strconv.Itoa(len(resp)))
			w.WriteHeader(200)
		case http.MethodGet:
			var statusCode = 200
			var currentResp = resp
			if _, ok := req.Header["Range"]; ok {
				requestRange := req.Header.Get("Range")
				start, end := parseRangeInsideTest(requestRange)
				currentResp = resp[start : end+1]
				w.Header().Add("Content-Range", responseRangeHeaderValueInsideTest(start, end, len(resp)))
				statusCode = 206
			}
			switch req.URL.Path {
			case "/pesho":
				currentResp = currentResp[:returnMaxBytes]
				if !atomic.CompareAndSwapInt32(&byteCount1, 0, int32(len(currentResp))) {
					w.WriteHeader(500)
					return
				}
			case "/pesho2":
				atomic.AddInt32(&byteCount2, int32(len(currentResp)))
			case "/pesho3":
				currentResp = currentResp[:1]
				atomic.AddInt32(&byteCount3, int32(len(currentResp)))
			}

			w.Header().Add("Content-Length", strconv.Itoa(len(currentResp)))
			w.WriteHeader(statusCode)
			n, err := w.Write(currentResp)
			if n != len(currentResp) {
				t.Errorf("Wrote %d not %d as expected", n, len(currentResp))
			}
			if err != nil {
				t.Errorf("Got error while writing response")
			}
		}
	}))

	defer s.Close()

	r := DownloadFile(context.Background(),
		[]string{
			s.URL + "/pesho",
			s.URL + "/pesho2",
			s.URL + "/pesho3",
		})
	var buf bytes.Buffer
	n, err := buf.ReadFrom(r)
	if int(n) != len(resp) {
		t.Errorf("Expected to read %d bytes from simple download but got %d", len(resp), n)
	}

	if err != nil {
		t.Errorf("Expected to get no error, but got '%s'", err)
	}

	checkResponseInsideTest(t, resp, buf.Bytes()[:n])

	if int64(byteCount1+byteCount2+byteCount3) != n {
		t.Errorf("It was expected that the downloaded amount from both urls "+
			" will be equal to the size of the request but while it was %d, "+
			"from url[0] = %d, url[1] = %d, url[2] = %d", n, byteCount1, byteCount2, byteCount3)
	}

	if int(byteCount1) != returnMaxBytes {
		t.Errorf("It was expected that url[0] would serve %d but have served %d ",
			returnMaxBytes, byteCount1)
	}

	if int(byteCount2+byteCount3) != len(resp)-returnMaxBytes {
		t.Errorf("It was expected that url[1] +url[2] would serve %d but have served %d ",
			len(resp)-returnMaxBytes, byteCount2)
	}

	if abs32InsideTest(byteCount2-byteCount3) > 1 {
		t.Errorf("Both urls should have been used equally but %d were downloaded "+
			"from url[1] and %d from url[2]", byteCount2, byteCount3)
	}
}

// Test simple case with one url which returns 1 byte each time
func TestSingleURL1ByteARequest(t *testing.T) {
	var resp = []byte("This IS the most epic of all response")
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodHead:
			w.Header().Add("Content-Length", strconv.Itoa(len(resp)))
			w.WriteHeader(200)
		case http.MethodGet:
			var statusCode = 200
			var currentResp = resp
			if _, ok := req.Header["Range"]; ok {
				requestRange := req.Header.Get("Range")
				start, end := parseRangeInsideTest(requestRange)
				currentResp = resp[start : end+1]
				w.Header().Add("Content-Range", responseRangeHeaderValueInsideTest(start, end, len(resp)))
				statusCode = 206
			}
			currentResp = currentResp[:1]
			w.Header().Add("Content-Length", strconv.Itoa(len(currentResp)))
			w.WriteHeader(statusCode)
			n, err := w.Write(currentResp)
			if n != len(currentResp) {
				t.Errorf("Wrote %d not %d as expected", n, len(currentResp))
			}
			if err != nil {
				t.Errorf("Got error while writing response")
			}
		}
	}))
	defer s.Close()

	r := DownloadFile(context.Background(), []string{s.URL + "/pesho"})
	var buf bytes.Buffer
	n, err := buf.ReadFrom(r)
	if int(n) != len(resp) {
		t.Errorf("Expected to read %d bytes from simple download but got %d", len(resp), n)
	}

	if err != nil {
		t.Errorf("Expected to get no error, but got '%s'", err)
	}

	checkResponseInsideTest(t, resp, buf.Bytes()[:n])
}

// Test simple case with two urls but reading 1 byte at a time from the received reader
func TestTwoUrlsWithOneByteReader(t *testing.T) {
	var resp = []byte("This IS the most epic of all response")
	var byteCount1, byteCount2 int32
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodHead:
			w.Header().Add("Content-Length", strconv.Itoa(len(resp)))
			w.WriteHeader(200)
		case http.MethodGet:
			var statusCode = 200
			var currentResp = resp
			if _, ok := req.Header["Range"]; ok {
				requestRange := req.Header.Get("Range")
				start, end := parseRangeInsideTest(requestRange)
				currentResp = resp[start : end+1]
				w.Header().Add("Content-Range", responseRangeHeaderValueInsideTest(start, end, len(resp)))
				statusCode = 206
			}
			switch req.URL.Path {
			case "/pesho":
				atomic.AddInt32(&byteCount1, int32(len(currentResp)))
			case "/pesho2":
				atomic.AddInt32(&byteCount2, int32(len(currentResp)))
			}

			w.Header().Add("Content-Length", strconv.Itoa(len(currentResp)))
			w.WriteHeader(statusCode)
			n, err := w.Write(currentResp)
			if n != len(currentResp) {
				t.Errorf("Wrote %d not %d as expected", n, len(currentResp))
			}
			if err != nil {
				t.Errorf("Got error while writing response")
			}
		}
	}))
	defer s.Close()

	r := DownloadFile(context.Background(), []string{s.URL + "/pesho", s.URL + "/pesho2"})
	r = iotest.OneByteReader(r)
	var buf bytes.Buffer
	n, err := buf.ReadFrom(r)
	if int(n) != len(resp) {
		t.Errorf("Expected to read %d bytes from simple download but got %d", len(resp), n)
	}

	if err != nil {
		t.Errorf("Expected to get no error, but got '%s'", err)
	}

	checkResponseInsideTest(t, resp, buf.Bytes()[:n])

	if int64(byteCount1+byteCount2) != n {
		t.Errorf("It was expected that the downloaded amount from both urls "+
			" will be equal to the size of the request but while it was %d, "+
			"from url[0] was downloaded %d and from url[1] %d", n, byteCount1, byteCount2)
	}

	if abs32InsideTest(byteCount1-byteCount2) > 1 {
		t.Errorf("Both urls should have been used equally but %d were downloaded "+
			"from url[0] and %d from url[1]", byteCount1, byteCount2)
	}
}

// Test Lingchi - Death By A Thousand Cuts - a thousand urls all returnting a single byte
func TestLingchi(t *testing.T) {
	const thousand = 500 // it's important to not forget how much is a thousand
	var resp = []byte("This IS the most epic of all response")
	for {
		resp = append(resp, resp...)
		if len(resp) > thousand {
			resp = resp[:thousand]
			break
		}
	}
	var byteCounts = [thousand]int32{}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodHead:
			w.Header().Add("Content-Length", strconv.Itoa(len(resp)))
			w.WriteHeader(200)
		case http.MethodGet:
			var statusCode = 200
			var currentResp = resp
			if _, ok := req.Header["Range"]; ok {
				requestRange := req.Header.Get("Range")
				start, end := parseRangeInsideTest(requestRange)
				currentResp = resp[start : end+1]
				w.Header().Add("Content-Range", responseRangeHeaderValueInsideTest(start, end, len(resp)))
				statusCode = 206
			}
			index, err := strconv.Atoi(req.URL.Path[1:])
			if err != nil {
				t.Errorf("Coudn't parse index out of %s", req.URL.Path)
				return
			}
			atomic.AddInt32(&byteCounts[index], int32(len(currentResp)))

			w.Header().Add("Content-Length", strconv.Itoa(len(currentResp)))
			w.WriteHeader(statusCode)
			n, err := w.Write(currentResp)
			if n != len(currentResp) {
				t.Errorf("Wrote %d not %d as expected", n, len(currentResp))
			}
			if err != nil {
				t.Errorf("Got error while writing response")
			}
		}
	}))
	defer s.Close()

	var urls = make([]string, thousand)
	for i := 0; i < thousand; i++ {
		urls[i] = s.URL + "/" + strconv.Itoa(i)
	}
	r := DownloadFile(context.Background(), urls)
	r = iotest.OneByteReader(r)
	var buf bytes.Buffer
	n, err := buf.ReadFrom(r)
	if int(n) != len(resp) {
		t.Errorf("Expected to read %d bytes from simple download but got %d", len(resp), n)
	}

	if err != nil {
		t.Errorf("Expected to get no error, but got '%s'", err)
	}

	checkResponseInsideTest(t, resp, buf.Bytes()[:n])

	for index, byteCount := range byteCounts {
		if byteCount != 1 {
			t.Errorf("From url with index %d were received %d bytes was expected 1", index, byteCount)
		}
	}
}

// Test Small Lingchi with max connections - Death By A Thousand Cuts - a thousand(hundred) urls all
// returnting a single byte but waiting a while first and with max of 20 connections
func TestSlowLingchiWithMaxConnections(t *testing.T) {
	const thousand = 100 // it's important to not forget how much is a thousand
	var resp = []byte("This IS the most epic of all response")
	for {
		resp = append(resp, resp...)
		if len(resp) > thousand {
			resp = resp[:thousand]
			break
		}
	}
	const maxConnections = 20
	var ctx = context.WithValue(context.Background(), "max-connections", maxConnections)
	var byteCounts = [thousand]int32{}
	var concurrentRequest int32
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if oldValue := atomic.AddInt32(&concurrentRequest, 1); oldValue > maxConnections {
			t.Errorf("Got new request while already having %d", oldValue)
		}
		defer atomic.AddInt32(&concurrentRequest, -1)
		time.Sleep(time.Millisecond * 10)
		switch req.Method {
		case http.MethodHead:
			w.Header().Add("Content-Length", strconv.Itoa(len(resp)))
			w.WriteHeader(200)
		case http.MethodGet:
			var statusCode = 200
			var currentResp = resp
			if _, ok := req.Header["Range"]; ok {
				requestRange := req.Header.Get("Range")
				start, end := parseRangeInsideTest(requestRange)
				currentResp = resp[start : end+1]
				w.Header().Add("Content-Range", responseRangeHeaderValueInsideTest(start, end, len(resp)))
				statusCode = 206
			}
			index, err := strconv.Atoi(req.URL.Path[1:])
			if err != nil {
				t.Errorf("Coudn't parse index out of %s", req.URL.Path)
				return
			}
			atomic.AddInt32(&byteCounts[index], int32(len(currentResp)))

			w.Header().Add("Content-Length", strconv.Itoa(len(currentResp)))
			w.WriteHeader(statusCode)
			n, err := w.Write(currentResp)
			if n != len(currentResp) {
				t.Errorf("Wrote %d not %d as expected", n, len(currentResp))
			}
			if err != nil {
				t.Errorf("Got error while writing response")
			}
		}
	}))
	defer s.Close()

	var urls = make([]string, thousand)
	for i := 0; i < thousand; i++ {
		urls[i] = s.URL + "/" + strconv.Itoa(i)
	}
	r := DownloadFile(ctx, urls)
	r = iotest.OneByteReader(r)
	var buf bytes.Buffer
	n, err := buf.ReadFrom(r)
	if int(n) != len(resp) {
		t.Errorf("Expected to read %d bytes from simple download but got %d", len(resp), n)
	}

	if err != nil {
		t.Errorf("Expected to get no error, but got '%s'", err)
	}

	checkResponseInsideTest(t, resp, buf.Bytes()[:n])

	for index, byteCount := range byteCounts {
		if byteCount != 1 {
			t.Errorf("From url with index %d were received %d bytes was expected 1", index, byteCount)
		}
	}
}

// Test Small Lingchi with max connections - Death By A Thousand Cuts - a thousand(hundred) urls all
// returnting a single byte but waiting a while first and with max of 20 connections and errors all
// around
func TestSlowLingchiWithBothMaxConnectionsAndALotOfErrors(t *testing.T) {
	const thousand = 100 // it's important to not forget how much is a thousand
	const returningUrls = 5
	var resp = []byte("This IS the most epic of all response")
	for {
		resp = append(resp, resp...)
		if len(resp) > thousand {
			resp = resp[:thousand]
			break
		}
	}
	const maxConnections = 20
	var ctx = context.WithValue(context.Background(), "max-connections", maxConnections)
	var byteCounts = [thousand]int32{}
	var currentlyGetting = [thousand]int32{}
	var concurrentRequest int32
	var shouldReturn = func(index int) bool {
		return index%(thousand/returningUrls) == 0
	}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if oldValue := atomic.AddInt32(&concurrentRequest, 1); oldValue > maxConnections {
			t.Errorf("Got new request while already having %d", oldValue)
		}
		defer atomic.AddInt32(&concurrentRequest, -1)
		index, err := strconv.Atoi(req.URL.Path[1:])
		if err != nil {
			t.Errorf("Coudn't parse index out of %s", req.URL.Path)
			return
		}
		if atomic.CompareAndSwapInt32(&currentlyGetting[index], 0, 1) {
			t.Errorf("url with index %d already in request when a new one came", index)
		}
		defer atomic.StoreInt32(&currentlyGetting[index], 0)
		time.Sleep(time.Millisecond * 10)

		switch req.Method {
		case http.MethodHead:
			w.Header().Add("Content-Length", strconv.Itoa(len(resp)))
			w.WriteHeader(200)
		case http.MethodGet:

			if !shouldReturn(index) {
				w.WriteHeader(500)
				return
			}

			var statusCode = 200
			var currentResp = resp
			if _, ok := req.Header["Range"]; ok {
				requestRange := req.Header.Get("Range")
				start, end := parseRangeInsideTest(requestRange)
				currentResp = resp[start : end+1]
				w.Header().Add("Content-Range", responseRangeHeaderValueInsideTest(start, end, len(resp)))
				statusCode = 206
			}
			atomic.AddInt32(&byteCounts[index], int32(len(currentResp)))

			w.Header().Add("Content-Length", strconv.Itoa(len(currentResp)))
			w.WriteHeader(statusCode)
			n, err := w.Write(currentResp)
			if n != len(currentResp) {
				t.Errorf("Wrote %d not %d as expected", n, len(currentResp))
			}
			if err != nil {
				t.Errorf("Got error while writing response")
			}
		}
	}))
	defer s.Close()

	var urls = make([]string, thousand)
	for i := 0; i < thousand; i++ {
		urls[i] = s.URL + "/" + strconv.Itoa(i)
	}
	r := DownloadFile(ctx, urls)
	r = iotest.OneByteReader(r)
	var buf bytes.Buffer
	n, err := buf.ReadFrom(r)
	if int(n) != len(resp) {
		t.Errorf("Expected to read %d bytes from simple download but got %d", len(resp), n)
	}

	if err != nil {
		t.Errorf("Expected to get no error, but got '%s'", err)
	}

	checkResponseInsideTest(t, resp, buf.Bytes()[:n])

	for index, byteCount := range byteCounts {
		var expectedByteCount int32 = 0
		if shouldReturn(index) {
			expectedByteCount = int32(len(resp) / returningUrls)
		}
		if byteCount != expectedByteCount {
			t.Errorf("From url with index %d were received %d bytes was expected %d", index, byteCount, expectedByteCount)
		}
	}
}
func abs32InsideTest(a int32) int32 {
	if a > 0 {
		return a
	}
	return -a
}
