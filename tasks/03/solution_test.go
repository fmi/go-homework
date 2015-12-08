package main

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"runtime"
	"testing"
)

const (
	_ = iota
	TestTakeBook
	TestReturnBook
	TestGetAvailability
)

var books = map[string]map[string]string{
	"anno": {
		"isbn":              "0954540018",
		"author":            "Anno Birkin",
		"author_first_name": "Anno",
		"author_last_name":  "Birkin",
		"title":             "Who Said the Race is Over?",
		"json": `{
					"isbn": "0954540018",
					"title": "Who Said the Race is Over?",
					"author": {
						"first_name": "Anno",
						"last_name": "Birkin"
					},
					"genre": "poetry",
					"pages": 80,
					"ratings": [5, 4, 4, 5, 3]
				}`,
		"xml": `
<book isbn="0954540018">
  <title>Who said the race is Over?</title>
  <author>
    <first_name>Anno</first_name>
    <last_name>Birkin</last_name>
  </author>
  <genre>poetry</genre>
  <pages>80</pages>
  <ratings>
    <rating>5</rating>
    <rating>4</rating>
    <rating>4</rating>
    <rating>5</rating>
    <rating>3</rating>
  </ratings>
</book>
`,
	},
}

type TestLibraryRequest struct {
	ISBN string
	Type int
}

func (lr *TestLibraryRequest) GetType() int {
	return lr.Type
}

func (lr *TestLibraryRequest) GetISBN() string {
	return lr.ISBN
}

func (lr *TestLibraryRequest) SetType(reqtype int) {
	lr.Type = reqtype
}

func (lr *TestLibraryRequest) SetISBN(isbn string) {
	lr.ISBN = isbn
}

func expected(t *testing.T, condition bool, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		v = append([]interface{}{filepath.Base(file), line}, v...)
		t.Errorf("[%s:%d] Expected:\n%v\nGot:\n%v\nContext: %v", v...)
	}
}

func seedBooks(l Library) {
	_, err := l.AddBookJSON([]byte(books["anno"]["json"]))
	for i := rand.Int() % 4; i >= 0; i-- {
		_, err = l.AddBookJSON([]byte(books["anno"]["json"]))
		if err != nil {
			break
		}
	}
}

func TestNewLibrary(t *testing.T) {
	var library Library = NewLibrary(2)
	if library == nil {
		t.Errorf("NewLibrary returned %#v", library)
	}
}

func TestAddBookJSON(t *testing.T) {
	l := NewLibrary(2)
	available, err := l.AddBookJSON([]byte(books["anno"]["json"]))
	expected(t, err == nil, nil, err, "Error when adding a valid book")
	expected(t, available == 1, 1, available, "first book added")

	available, err = l.AddBookJSON([]byte(books["anno"]["json"]))
	expected(t, available == 2, 2, available, "first book added 2nd time")

	available, err = l.AddBookJSON([]byte(books["anno"]["json"]))
	expected(t, available == 3, 3, available, "first book added 3rd time")

	available, err = l.AddBookJSON([]byte(books["anno"]["json"]))
	expected(t, available == 4, 4, available, "first book added 4th time")

	available, err = l.AddBookJSON([]byte(books["anno"]["json"]))
	expected(t, err != nil && err.Error() == "Има 4 копия на книга "+books["anno"]["isbn"], "Има 4 копия на книга "+books["anno"]["isbn"], nil, "Error when adding a book for 5th time")

	reqChan, respChan := l.Hello()

	reqChan <- &TestLibraryRequest{books["anno"]["isbn"], TestTakeBook}

	response := <-respChan

	_, err = response.GetBook()

	expected(t, err == nil, nil, err, "Error when getting a valid book")
}

func TestAddBookXML(t *testing.T) {
	l := NewLibrary(2)

	available, err := l.AddBookXML([]byte(books["anno"]["xml"]))
	expected(t, err == nil, nil, err, "Error when adding a valid book")
	expected(t, available == 1, 1, available, "first book added")

	available, err = l.AddBookXML([]byte(books["anno"]["xml"]))
	expected(t, available == 2, 2, available, "first book added 2nd time")

	available, err = l.AddBookXML([]byte(books["anno"]["xml"]))
	expected(t, available == 3, 3, available, "first book added 3rd time")

	available, err = l.AddBookXML([]byte(books["anno"]["xml"]))
	expected(t, available == 4, 4, available, "first book added 4th time")

	available, err = l.AddBookXML([]byte(books["anno"]["xml"]))
	expected(t, err != nil && err.Error() == "Има 4 копия на книга "+books["anno"]["isbn"], "Има 4 копия на книга "+books["anno"]["isbn"], nil, "Error when adding a book for 5th time")

	reqChan, respChan := l.Hello()

	reqChan <- &TestLibraryRequest{books["anno"]["isbn"], TestTakeBook}

	response := <-respChan

	_, err = response.GetBook()

	expected(t, err == nil, nil, err, "Error when getting a valid book")
	close(reqChan)
}

func TestAddBookCombined(t *testing.T) {
	l := NewLibrary(2)

	available, err := l.AddBookXML([]byte(books["anno"]["xml"]))
	expected(t, err == nil, nil, err, "Error when adding a valid book")
	expected(t, available == 1, 1, available, "first book added")

	available, err = l.AddBookJSON([]byte(books["anno"]["json"]))
	expected(t, err == nil, nil, err, "Error when adding a valid book")
	expected(t, available == 2, 2, available, "first book added 2nd time")

	available, err = l.AddBookXML([]byte(books["anno"]["xml"]))
	expected(t, available == 3, 3, available, "first book added 3rd time")
	expected(t, err == nil, nil, err, "Error when adding a valid book")

	available, err = l.AddBookJSON([]byte(books["anno"]["json"]))
	expected(t, available == 4, 4, available, "first book added 4th time")
	expected(t, err == nil, nil, err, "Error when adding a valid book")

	available, err = l.AddBookXML([]byte(books["anno"]["xml"]))
	expected(t, err.Error() == "Има 4 копия на книга "+books["anno"]["isbn"], nil, err, "Error when adding a book for 5th time")

	reqChan, respChan := l.Hello()

	reqChan <- &TestLibraryRequest{books["anno"]["isbn"], TestTakeBook}

	response := <-respChan

	_, err = response.GetBook()

	expected(t, err == nil, nil, err, "Error when getting a valid book")

	close(reqChan)
}

func TestBookAvailability(t *testing.T) {
	l := NewLibrary(2)

	available, err := l.AddBookJSON([]byte(books["anno"]["json"]))
	expected(t, err == nil, nil, err, "error when adding a valid book")
	expected(t, available == 1, 1, available, "first book added 1st time")

	available, err = l.AddBookJSON([]byte(books["anno"]["json"]))
	expected(t, err == nil, nil, err, "error when adding a valid book")
	expected(t, available == 2, 2, available, "first book added 2nd time")

	reqChan, respChan := l.Hello()

	reqChan <- &TestLibraryRequest{books["anno"]["isbn"], TestGetAvailability}

	response := <-respChan

	av, reg := response.GetAvailability()
	expected(t, reg == 2, 2, reg, "we have registered 2 copies")
	expected(t, av == 2, 2, av, "should have 2 copies available")

	close(reqChan)
}

func TestTakeBookRequest(t *testing.T) {
	var response LibraryResponse
	var err error
	l := NewLibrary(2)
	seedBooks(l)

	reqChan, respChan := l.Hello()

	reqChan <- &TestLibraryRequest{books["anno"]["isbn"], TestGetAvailability}

	response = <-respChan

	av, _ := response.GetAvailability()

	for i := 0; i < av; i++ {
		reqChan <- &TestLibraryRequest{books["anno"]["isbn"], TestTakeBook}

		response = <-respChan

		_, err = response.GetBook()
		av1, _ := response.GetAvailability()

		expected(t, av1 == av-i-1, av-i-1, av, " available books left...")
		expected(t, err == nil, nil, err, "Error when getting a valid book")
	}

	reqChan <- &TestLibraryRequest{books["anno"]["isbn"], TestTakeBook}

	response = <-respChan

	_, err = response.GetBook()

	if err == nil {
		t.Fatal("Expected error but got nil")
	}

	expected(t, err.Error() == "Няма наличност на книга "+books["anno"]["isbn"], "Няма наличност на книга "+books["anno"]["isbn"], err.Error(), "should get error when getting a book that is unavailable")

	close(reqChan)
}

func TestBookStringer(t *testing.T) {
	var response LibraryResponse
	var err error
	l := NewLibrary(2)
	seedBooks(l)

	reqChan, respChan := l.Hello()

	reqChan <- &TestLibraryRequest{books["anno"]["isbn"], TestTakeBook}

	response = <-respChan

	book, err := response.GetBook()

	if err != nil {
		t.Fatal("Expected nil but got error", err.Error())
	}

	expected(t, book.String() == fmt.Sprintf("[%s] %s от %s", books["anno"]["isbn"], books["anno"]["title"], books["anno"]["author"]), fmt.Sprintf("[%s] %s от %s", books["anno"]["isbn"], books["anno"]["title"], books["anno"]["author"]), book.String(), "Book as fmt.Stringer")

	close(reqChan)
}

func TestTakeMissingBook(t *testing.T) {
	var response LibraryResponse
	var err error
	l := NewLibrary(2)
	seedBooks(l)

	reqChan, respChan := l.Hello()

	reqChan <- &TestLibraryRequest{books["anno"]["isbn"] + "123", TestTakeBook}

	response = <-respChan

	_, err = response.GetBook()
	if err == nil {
		t.Fatal("Expected error but got nil")
	}

	expected(t, err.Error() == "Непозната книга "+books["anno"]["isbn"]+"123", "Непозната книга "+books["anno"]["isbn"]+"123", err.Error(), "Error when getting unregistered book")

	close(reqChan)
}

func TestReturnSomeBooks(t *testing.T) {
	var response LibraryResponse
	var err error
	l := NewLibrary(2)

	reqChan, respChan := l.Hello()

	for {
		if _, err := l.AddBookJSON([]byte(books["anno"]["json"])); err != nil {
			break
		}
	}

	//take 2
	reqChan <- &TestLibraryRequest{books["anno"]["isbn"], TestTakeBook}
	<-respChan
	reqChan <- &TestLibraryRequest{books["anno"]["isbn"], TestTakeBook}
	<-respChan

	reqChan <- &TestLibraryRequest{books["anno"]["isbn"], TestGetAvailability}

	response = <-respChan

	beforeReturningA, beforeReturningR := response.GetAvailability()

	toBeReturned := beforeReturningR - beforeReturningA

	for toBeReturned--; toBeReturned >= -1; toBeReturned-- {
		reqChan <- &TestLibraryRequest{books["anno"]["isbn"], TestReturnBook}

		response = <-respChan

		_, err = response.GetBook()

		av, reg := response.GetAvailability()
		if toBeReturned > -1 {
			expected(t, err == nil, nil, err, "Error when returning a valid book")
			expected(t, reg-av == toBeReturned, toBeReturned, reg-av, "Available should increment with one")
			if toBeReturned == 0 {
				expected(t, reg == av, av, reg, "all books to be returned")
			}
		} else {
			if err == nil {
				t.Fatal("Expected error got nil")
			}
			expected(t, err.Error() == "Всички копия са налични "+books["anno"]["isbn"], "Всички копия са налични "+books["anno"]["isbn"], err.Error(), "Error when returning too much copies")
		}
		expected(t, reg == beforeReturningR, beforeReturningR, reg, "Registered should not change when we return a book")
	}

	close(reqChan)
}

func TestReturnUnknownBook(t *testing.T) {
	var response LibraryResponse
	var err error
	l := NewLibrary(2)
	seedBooks(l)

	reqChan, respChan := l.Hello()

	reqChan <- &TestLibraryRequest{books["anno"]["isbn"] + "123", TestReturnBook}

	response = <-respChan
	_, err = response.GetBook()

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	expected(t, err.Error() == "Непозната книга "+books["anno"]["isbn"]+"123", "Непозната книга "+books["anno"]["isbn"]+"123", err.Error(), "should get error when we return unknown book")

	close(reqChan)
}
