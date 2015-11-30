package main

import (
	"fmt"
	"math/rand"
)

const (
	_ = iota
	ExampleTestTakeBook
	ExampleTestReturnBook
	ExampleTestGetAvailability
)

var test_books = map[string]map[string]string{
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

type ExampleTestLibraryRequest struct {
	ISBN string
	Type int
}

func (lr *ExampleTestLibraryRequest) GetType() int {
	return lr.Type
}

func (lr *ExampleTestLibraryRequest) GetISBN() string {
	return lr.ISBN
}

func (lr *ExampleTestLibraryRequest) SetType(reqtype int) {
	lr.Type = reqtype
}

func (lr *ExampleTestLibraryRequest) SetISBN(isbn string) {
	lr.ISBN = isbn
}

func exampleSeedBooks(l Library) {
	_, err := l.AddBookJSON([]byte(test_books["anno"]["json"]))
	for i := rand.Int() % 4; i >= 0; i-- {
		_, err = l.AddBookJSON([]byte(test_books["anno"]["json"]))
		if err != nil {
			break
		}
	}
}
func ExampleAddBook() {
	l := NewLibrary(2)

	available, err := l.AddBookXML([]byte(test_books["anno"]["xml"]))
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Printf("Available: %d\n", available)
	}

	available, err = l.AddBookJSON([]byte(test_books["anno"]["json"]))
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Printf("Available: %d\n", available)
	}

	available, err = l.AddBookXML([]byte(test_books["anno"]["xml"]))
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Printf("Available: %d\n", available)
	}

	available, err = l.AddBookJSON([]byte(test_books["anno"]["json"]))
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Printf("Available: %d\n", available)
	}

	available, err = l.AddBookXML([]byte(test_books["anno"]["xml"]))
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	} else {
		fmt.Printf("Available: %d\n", available)
	}

	// Output:
	// Available: 1
	// Available: 2
	// Available: 3
	// Available: 4
	// Error: Има 4 копия на книга 0954540018
}

func ExampleTakeBook() {
	var response LibraryResponse
	var err error
	l := NewLibrary(2)
	exampleSeedBooks(l)

	reqChan, respChan := l.Hello()

	reqChan <- &ExampleTestLibraryRequest{test_books["anno"]["isbn"], ExampleTestTakeBook}

	response = <-respChan

	book, err := response.GetBook()

	if err != nil {
		fmt.Printf("Error: %s", err.Error())
		//t.Fatal("Expected nil but got error", err.Error())
	} else {
		fmt.Printf("Book: %s", book)
	}

	close(reqChan)
	// Output:
	// Book: [0954540018] Who Said the Race is Over? от Anno Birkin
}
