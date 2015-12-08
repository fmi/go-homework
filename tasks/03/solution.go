package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
)

// haz some jam
type jam struct{}

type Library interface {

	// Добавя книга от json
	// Oтговаря с общия брой копия в библиотеката (не само наличните).
	// Aко са повече от 4 - връща грешка
	AddBookJSON(data []byte) (int, error)

	// Добавя книга от xml
	// Oтговаря с общия брой копия в библиотеката (не само наличните).
	// Ако са повече от 4 - връщаме грешка
	AddBookXML(data []byte) (int, error)

	// Ангажира свободен "библиотекар" да ни обработва заявките.
	// Библиотекарите са фиксиран брой - подават се като параметър на NewLibrary
	// Блокира ако всички библиотекари са заети.
	// Връщат се два канала:
	// първият е само за писане -  по него ще изпращаме заявките
	// вторият е само за четене - по него ще получаваме отговорите.
	// Ако затворим канала със заявките - освобождаваме библиотекаря.
	Hello() (chan<- LibraryRequest, <-chan LibraryResponse)
}

type LibraryRequest interface {
	// Тип на заявката:
	// 1 - Borrow book
	// 2 - Return book
	// 3 - Get availability information about book
	GetType() int

	// Връща isbn на книгата, за която се отнася Request-a
	GetISBN() string
}

type LibraryResponse interface {
	// Ако книгата съществува/налична е - обект имплементиращ Stringer (повече информация по-долу)
	// Aко книгата не съществува първият резултат е nil.
	// Връща се и подобаващa грешка (виж по-долу) - ако такава е възникнала.
	// Когато се е резултат на заявка от тип 2 (Return book) - не е нужно да я закачаме към отговора.
	GetBook() (fmt.Stringer, error)

	// available - Колко наличности от книгата имаме останали след изпълнението на заявката.
	// Тоест, ако сме имали 3 копия от Х и това е отговор на Take заявка - тук ще има 2.
	// registered - Колко копия от тази книга има регистрирани в библиотеката (макс 4).
	GetAvailability() (available int, registered int)
}

const (
	_ = iota
	TakeBook
	ReturnBook
	GetAvailability
)

type Book struct {
	ISBN    string  `xml:"isbn,attr" json:"isbn"`         // isbn на книгата
	Title   string  `xml:"title" json:"title"`            // заглавие на книгата
	Author  *Author `xml:"author" json:"author"`          // Обект от тип имплементираш Person интерфейса. Данни за автора.
	Genre   string  `xml:"genre" json:"genre"`            // Стил на книгата.
	Pages   int     `xml:"pages" json:"pages"`            // # страници
	Ratings []int   `xml:"ratings>rating" json:"ratings"` // слайс от оценки за книгата - цели числа между 0 и 5.
}

func (b *Book) String() string {
	return "[" + b.ISBN + "] " + b.Title + " от " + b.Author.FirstName + " " + b.Author.LastName
}

type Author struct {
	FirstName string `xml:"first_name" json:"first_name"`
	LastName  string `xml:"last_name" json:"last_name"`
}

type MyLibraryResponse struct {
	book         *Book
	err          error
	availability Availability
}

func (lr *MyLibraryResponse) GetBook() (fmt.Stringer, error) {
	return lr.book, lr.err
}

func (lr *MyLibraryResponse) SetBook(book *Book) {
	lr.book = book
}

func (lr *MyLibraryResponse) SetError(err error) {
	lr.err = err
}

func (lr *MyLibraryResponse) SetAvailability(av Availability) {
	lr.availability = av
}

func (lr *MyLibraryResponse) GetAvailability() (int, int) {
	return lr.availability.Available, lr.availability.Registered
}

type Availability struct {
	Available  int
	Registered int
}

type MyLibrary struct {
	librarians        chan jam
	books             map[string]*Book
	booksAvailability map[string]*Availability
	booksLock         chan jam
}

type Customer struct {
	requests  chan LibraryRequest
	responses chan LibraryResponse
}

func (l *MyLibrary) lockBooks() {
	<-l.booksLock
}

func (l *MyLibrary) unlockBooks() {
	l.booksLock <- jam{}
}

func (l *MyLibrary) addBook(b *Book) (int, error) {
	var av *Availability
	var ok bool
	l.lockBooks()
	defer l.unlockBooks()
	if av, ok = l.booksAvailability[b.ISBN]; ok {
		if av.Registered > 3 {
			err := errors.New("Има 4 копия на книга " + b.ISBN)
			return av.Registered, err
		} else {
			av.Registered++
			av.Available++
		}
	} else {
		av = &Availability{1, 1}
		l.booksAvailability[b.ISBN] = av
		l.books[b.ISBN] = b
	}
	return av.Registered, nil
}

func (l *MyLibrary) AddBookJSON(data []byte) (int, error) {
	var b Book
	if err := json.Unmarshal(data, &b); err != nil {
		return 0, err
	}
	return l.addBook(&b)
}

func (l *MyLibrary) AddBookXML(data []byte) (int, error) {
	var b Book
	if err := xml.Unmarshal(data, &b); err != nil {
		return 0, err
	}
	return l.addBook(&b)
}

func (l *MyLibrary) Hello() (chan<- LibraryRequest, <-chan LibraryResponse) {
	requests := make(chan LibraryRequest)
	responses := make(chan LibraryResponse)
	c := &Customer{requests, responses}
	<-l.librarians
	go func(customer *Customer) {
		l.handle(customer)
		l.librarians <- jam{}
	}(c)
	return requests, responses
}

func (l *MyLibrary) handle(customer *Customer) {
	requests := customer.requests
	responses := customer.responses
	var av *Availability
	var found bool
	for {
		select {
		case request, ok := <-requests:
			if !ok {
				return
			}
			l.lockBooks()

			isbn := request.GetISBN()
			resp := &MyLibraryResponse{}
			if av, found = l.booksAvailability[isbn]; !found {
				resp.SetError(errors.New("Непозната книга " + isbn))
				responses <- resp
				l.unlockBooks()
				continue
			}
			resp.SetAvailability(*av)

			switch request.GetType() {
			case TakeBook:
				if av.Available > 0 {
					if book, ok := l.books[isbn]; ok {
						resp.SetBook(book)
						av.Available--
						resp.SetAvailability(*av)
					} else {
						// This should not happen, but we are prepared!
						resp.SetError(errors.New("Непозната книга " + isbn))
					}
				} else {
					resp.SetError(errors.New("Няма наличност на книга " + isbn))
				}

			case ReturnBook:
				if av.Available >= av.Registered {
					resp.SetError(errors.New("Всички копия са налични " + isbn))
				} else {
					if book, ok := l.books[isbn]; ok {
						resp.SetBook(book)
						av.Available++
						resp.SetAvailability(*av)
					} else {
						// This should not happen, but we are prepared!
						resp.SetError(errors.New("Непозната книга " + isbn))
					}
				}

			case GetAvailability:
				if book, ok := l.books[isbn]; ok {
					resp.SetBook(book)
				} else {
					// This should not happen, but we are prepared!
					resp.SetError(errors.New("Непозната книга " + isbn))
				}
			}
			l.unlockBooks()
			responses <- resp
		}
	}
}

func NewLibrary(librarians int) Library {
	l := &MyLibrary{}
	l.librarians = make(chan jam, librarians)
	l.books = make(map[string]*Book)
	l.booksAvailability = make(map[string]*Availability)
	l.booksLock = make(chan jam, 1)

	// give 'em some jam
	for i := 0; i < librarians; i++ {
		l.librarians <- jam{}
	}

	// just unlock the first book operation
	go func() {
		l.booksLock <- jam{}
	}()

	return l
}
