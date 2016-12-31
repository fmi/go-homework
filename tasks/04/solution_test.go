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
)

var noValidUrls = "no valid urls"

var nonExistingURL = "http://some.non.existing.domain.at.nowhere/pesho"

func parseRange(s string) (start, end int) {
	fmt.Sscanf(s, "bytes=%d-%d", &start, &end)
	return
}

func responseRangeHeaderValue(start, end, size int) string {
	return fmt.Sprintf("bytes %d-%d/%d", start, end, size)
}

func checkResponse(t *testing.T, expected, got []byte) {
	if !bytes.Equal(expected, got) {
		t.Errorf("Expected result was '%s' but got '%s'", hex.EncodeToString(expected), hex.EncodeToString(got))
	}
}

// Test nothing to return
func TestNothingIsReturned(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
				start, end := parseRange(requestRange)
				currentResp = resp[start : end+1]
				w.Header().Add("Content-Range", responseRangeHeaderValue(start, end, len(resp)))
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

	checkResponse(t, resp, buf.Bytes()[:n])
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
				start, end := parseRange(requestRange)
				currentResp = resp[start : end+1]
				w.Header().Add("Content-Range", responseRangeHeaderValue(start, end, len(resp)))
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

	checkResponse(t, resp, buf.Bytes()[:n])
}

// Test no valid urls
func TestNoValidUrls(t *testing.T) {
	r := DownloadFile(context.Background(), []string{nonExistingURL})
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
					start, end := parseRange(requestRange)
					if start != 0 {
						t.Errorf("Expected to get one request from 0 to end got start %d", start)
					}
					currentResp = resp[start : end+1]
					w.Header().Add("Content-Range", responseRangeHeaderValue(start, end, len(resp)))
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

	if err == nil || err.Error() != noValidUrls {
		t.Errorf("Expected to get  error with message '%s', but got '%s'", noValidUrls, err)
	}

	checkResponse(t, resp[:bytesToReturn], buf.Bytes()[:n])
}

// Test simple case with two urls
func TestTwoUrls(t *testing.T) {
	var resp = []byte("This IS the most epic of all response")
	var pesho, pesho2 int32
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
				start, end := parseRange(requestRange)
				currentResp = resp[start : end+1]
				w.Header().Add("Content-Range", responseRangeHeaderValue(start, end, len(resp)))
				statusCode = 206
			}
			switch req.URL.Path {
			case "/pesho":
				atomic.AddInt32(&pesho, int32(len(currentResp)))
			case "/pesho2":
				atomic.AddInt32(&pesho2, int32(len(currentResp)))
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

	checkResponse(t, resp, buf.Bytes()[:n])
	if int64(pesho+pesho2) != n {
		t.Errorf("It was expected that the downloaded amount from both urls "+
			" will be equal to the size of the request but while it was %d, "+
			"from url[0] was downloaded %d and from url[1] %d", n, pesho, pesho2)
	}

	if abs32(pesho-pesho2) > 1 {
		t.Errorf("Both urls should have been used equally but %d were downloaded "+
			"from url[0] and %d from url[1]", pesho, pesho2)
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
				start, end := parseRange(requestRange)
				currentResp = resp[start : end+1]
				w.Header().Add("Content-Range", responseRangeHeaderValue(start, end, len(resp)))
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

	r := DownloadFile(context.Background(), []string{s.URL + "/pesho", nonExistingURL})
	var buf bytes.Buffer
	n, err := buf.ReadFrom(r)
	if int(n) != len(resp) {
		t.Errorf("Expected to read %d bytes from simple download but got %d", len(resp), n)
	}

	if err != nil {
		t.Errorf("Expected to get no error, but got '%s'", err)
	}

	checkResponse(t, resp, buf.Bytes()[:n])
}

// Test simple case with two urls one of which stops responding after the first 5 bytes
func TestTwoUrlsOneStopRespondingAfter5bytes(t *testing.T) {
	var resp = []byte("This IS the most epic of all response")
	var returnMaxBytes = 5
	var pesho, pesho2 int32
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
				start, end := parseRange(requestRange)
				currentResp = resp[start : end+1]
				w.Header().Add("Content-Range", responseRangeHeaderValue(start, end, len(resp)))
				statusCode = 206
			}
			switch req.URL.Path {
			case "/pesho":
				currentResp = currentResp[:returnMaxBytes]
				if !atomic.CompareAndSwapInt32(&pesho, 0, int32(len(currentResp))) {
					w.WriteHeader(500)
					return
				}
			case "/pesho2":
				atomic.AddInt32(&pesho2, int32(len(currentResp)))
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

	checkResponse(t, resp, buf.Bytes()[:n])

	if int64(pesho+pesho2) != n {
		t.Errorf("It was expected that the downloaded amount from both urls "+
			" will be equal to the size of the request but while it was %d, "+
			"from url[0] was downloaded %d and from url[1] %d", n, pesho, pesho2)
	}

	if int(pesho) != returnMaxBytes {
		t.Errorf("It was expected that url[0] would serve %d but have served %d ",
			returnMaxBytes, pesho)
	}

	if int(pesho2) != len(resp)-returnMaxBytes {
		t.Errorf("It was expected that url[1] would serve %d but have served %d ",
			len(resp)-returnMaxBytes, pesho2)
	}
}

// Test case with three urls, one of which returns only 2 bytes, another returns 1 byte at a time
func TestThreeUrlsOneOfWhichReturns1byteAtATimeOneOfWhichBreaksAfter2bytes(t *testing.T) {
	var resp = []byte("This IS the most epic of all response")
	var returnMaxBytes = 2
	var pesho, pesho2, pesho3 int32
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
				start, end := parseRange(requestRange)
				currentResp = resp[start : end+1]
				w.Header().Add("Content-Range", responseRangeHeaderValue(start, end, len(resp)))
				statusCode = 206
			}
			switch req.URL.Path {
			case "/pesho":
				currentResp = currentResp[:returnMaxBytes]
				if !atomic.CompareAndSwapInt32(&pesho, 0, int32(len(currentResp))) {
					w.WriteHeader(500)
					return
				}
			case "/pesho2":
				atomic.AddInt32(&pesho2, int32(len(currentResp)))
			case "/pesho3":
				currentResp = currentResp[:1]
				atomic.AddInt32(&pesho3, int32(len(currentResp)))
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

	checkResponse(t, resp, buf.Bytes()[:n])

	if int64(pesho+pesho2+pesho3) != n {
		t.Errorf("It was expected that the downloaded amount from both urls "+
			" will be equal to the size of the request but while it was %d, "+
			"from url[0] = %d, url[1] = %d, url[2] = %d", n, pesho, pesho2, pesho3)
	}

	if int(pesho) != returnMaxBytes {
		t.Errorf("It was expected that url[0] would serve %d but have served %d ",
			returnMaxBytes, pesho)
	}

	if int(pesho2+pesho3) != len(resp)-returnMaxBytes {
		t.Errorf("It was expected that url[1] +url[2] would serve %d but have served %d ",
			len(resp)-returnMaxBytes, pesho2)
	}

	if abs32(pesho2-pesho3) > 1 {
		t.Errorf("Both urls should have been used equally but %d were downloaded "+
			"from url[1] and %d from url[2]", pesho2, pesho3)
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
				start, end := parseRange(requestRange)
				currentResp = resp[start : end+1]
				w.Header().Add("Content-Range", responseRangeHeaderValue(start, end, len(resp)))
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

	checkResponse(t, resp, buf.Bytes()[:n])
}

func abs32(a int32) int32 {
	if a > 0 {
		return a
	}
	return -a
}
