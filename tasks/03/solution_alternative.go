package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
)

// 1 - Borrow book
// 2 - Return book
// 3 - Get availability information about book

const (
	Borrow = iota + 1
	Return
	Available
)

type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}

type Stringer interface {
	String() string
}

type Book struct {
	Title   string            `xml:"title" json:"title"`
	Isbn    string            `xml:"isbn,attr" json:"isbn"`
	Author  map[string]string `xml:"author>first_name" "author>last_name" json:"author"`
	Ratings [5]int            `xml:"ratings>rating" json:"ratings"`
}

func (b Book) String() string {
	return "[" + b.Isbn + "]" + " " + b.Title + " " + "от" + " " + b.Author["first_name"] + " " + b.Author["last_name"]
}

type Librarian struct {
	busy bool
}

type collectionInfo struct {
	Available  int
	Registered int
	Book       *Book
}

type Mylib struct {
	Collection map[string]*collectionInfo
	workers    []Librarian
}

type Request struct {
	Isbn string
	Type int
}

func (r *Request) GetType() int {
	return r.Type
}

func (r *Request) GetIsbn() string {
	return r.Isbn
}

type Response struct {
	available, registered int
	book                  Book
	bookError             error
}

func (r *Response) GetBook() (fmt.Stringer, error) {
	return r.book, r.bookError
}

func NewLibrary(librarians int) Library {
	new_lib := new(Mylib)
	var x = make([]Librarian, librarians)
	new_lib.workers = x
	new_lib.Collection = make(map[string]*collectionInfo)
	return new_lib
}

func (l *Mylib) AddBookJSON(data []byte) (int, error) {
	var count int
	var err error

	book := new(Book)

	if err = json.Unmarshal(data, &book); err != nil {
		return 0, err
	}

	info, ok := l.Collection[book.Isbn]
	if ok {
		count = info.Registered + 1
		if count > 4 {
			return 0, &errorString{"Има 4 копия на книга" + " " + book.Isbn}
		}
		info.Registered += 1
		info.Available += 1
	} else {
		count = 1
		l.Collection[book.Isbn] = &collectionInfo{
			Book:       book,
			Registered: 1,
			Available:  1,
		}
	}

	return count, nil
}

func (l *Mylib) AddBookXML(data []byte) (int, error) {
	b := new(Book)
	xml.Unmarshal(data, &b)
	mb, _ := json.Marshal(b)
	x, e := l.AddBookJSON(mb)
	return x, e
}

func (lib *Mylib) Hello() (chan<- LibraryRequest, <-chan LibraryResponse) {
	req := make(chan LibraryRequest)
	res := make(chan LibraryResponse)

	go func() {
		for req := range req {
			resp := &Response{}
			bookIsbn := req.GetISBN()
			bookInfo, ok := lib.Collection[bookIsbn]

			if !ok {
				resp.bookError = &errorString{"Непозната книга" + " " + bookIsbn}
				res <- resp
				continue
			}

			resp.registered = bookInfo.Registered
			resp.book = *bookInfo.Book

			switch req.GetType() {
			case Borrow:
				if bookInfo.Available < 1 {
					resp.bookError = &errorString{"Няма наличност на книга " + bookIsbn}
				} else {
					bookInfo.Available -= 1
				}
			case Return:
				if bookInfo.Available >= 4 {
					resp.bookError = &errorString{"Всички копия са налични " + bookIsbn}
				} else {
					bookInfo.Available += 1
				}
			case Available:
			default:
				panic("Unknown request type")
			}

			resp.available = bookInfo.Available

			res <- resp
		}
	}()

	return req, res
}

func (r *Response) GetAvailability() (available int, registered int) {
	return r.available, r.registered
}

type Library interface {
	AddBookJSON(data []byte) (int, error)
	AddBookXML(data []byte) (int, error)
	Hello() (chan<- LibraryRequest, <-chan LibraryResponse)
}

type LibraryRequest interface {
	GetType() int
	GetISBN() string
}

type LibraryResponse interface {
	GetBook() (fmt.Stringer, error)
	GetAvailability() (available int, registered int)
}
