package main

import (
	"reflect"
	"strings"
	"testing"
)

// TODO: maybe we also want to test "# *foo*" headings? Would we test that they
// parse the markdown?
func TestHeaders(t *testing.T) {
	mdParser := NewMarkdownParser(`
# One
##One # Two

Two
===

# # Three # Four
`)
	headers := mdParser.Headers()
	if len(headers) < 3 {
		t.Errorf("Length is less than 3: %#v", headers)
		return
	}

	results := []string{"One", "Two", "# Three # Four"}
	for i := 0; i < len(results); i++ {
		if headers[i] != results[i] {
			t.Errorf("%s is missing: %#v", results[i], headers)
		}
	}
}

func TestSubHeadersOf(t *testing.T) {
	mdParser := NewMarkdownParser(`
# Doo-doo

## Dee-dee
### Dum-dum
## Boo-boo

# Bla-bla

One-two
=======

Three-four
-------
`)
	subHeaders := mdParser.SubHeadersOf("Doo-doo")

	if len(subHeaders) < 2 {
		t.Errorf("Length is less than 2: %#v", subHeaders)
		return
	}

	expecteds := []string{"Dee-dee", "Boo-boo"}
	for i, expected := range expecteds {
		if subHeaders[i] != expected {
			t.Errorf("%s is missing: %#v", expected, subHeaders)
		}
	}

	subHeaders = mdParser.SubHeadersOf("Bla-bla")
	if len(subHeaders) != 0 {
		t.Errorf("Length is not 0: %#v", subHeaders)
		return
	}

	subHeaders = mdParser.SubHeadersOf("One-two")

	if len(subHeaders) != 1 {
		t.Errorf("Length is less than 2: %#v", subHeaders)
		return
	}

	expectations := []string{"Three-four"}
	for i, expected := range expectations {
		if subHeaders[i] != expected {
			t.Errorf("%s is missing: %#v", expected, subHeaders)
		}
	}
}

func TestNames(t *testing.T) {
	mdParser := NewMarkdownParser("Beginning Of Line. Аз съм Иван Петров. He Used to play games. His favourite browser is Firefox. That's Mozilla Firefox.")
	names := mdParser.Names()
	expected := []string{"Beginning Of", "Иван Петров", "Mozilla Firefox"}

	if !reflect.DeepEqual(names, expected) {
		t.Errorf("Not equal:\n  %#v\n  %#v", names, expected)
	}
}

func TestPhoneNumbers(t *testing.T) {
	mdParser := NewMarkdownParser("sometext 0889123456 alabala, 0 (889) 123 baba - 456, +45-(31), foobar")
	numbers := mdParser.PhoneNumbers()
	expected := []string{"0889123456", "0 (889) 123", "456", "+45-(31)"}

	for i, x := range numbers {
		numbers[i] = strings.TrimSpace(x)
	}

	if !reflect.DeepEqual(numbers, expected) {
		t.Errorf("Not equal:\n  %#v\n  %#v", numbers, expected)
	}
}

func TestLinks(t *testing.T) {
	mdParser := NewMarkdownParser("sometext http://somelink.com:230 ignore this 123 https://www.google.bg/search?q=4531&ie=utf-8&oe=utf-8&rls=org.mozilla:en-US:official&client=%20firefox-a&gws_rd=asd&ei=some#somefragment endoflink junk")
	links := mdParser.Links()
	expected := []string{
		"http://somelink.com:230",
		"https://www.google.bg/search?q=4531&ie=utf-8&oe=utf-8&rls=org.mozilla:en-US:official&client=%20firefox-a&gws_rd=asd&ei=some#somefragment",
	}
	if !reflect.DeepEqual(links, expected) {
		t.Errorf("Not equal:\n  %#v\n  %#v", links, expected)
	}
}

func TestEmails(t *testing.T) {
	mdParser := NewMarkdownParser("ignore validMail12@foobar.com sometext@ _invalidmail@google.com toolongmailhereaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa@gmail.com 12mail@gmail.com ")
	emails := mdParser.Emails()
	expected := []string{"validMail12@foobar.com", "12mail@gmail.com"}
	if !reflect.DeepEqual(emails, expected) {
		t.Errorf("Not equal:\n  %#v\n  %#v", emails, expected)
	}
}

func getSplitContents(tableOfContents string) []string {
	splitContents := strings.Split(tableOfContents, "\n")
	lengthMinusOne := len(splitContents) - 1
	if splitContents[lengthMinusOne] == "" {
		splitContents = splitContents[:lengthMinusOne]
	}

	for i, x := range splitContents {
		splitContents[i] = strings.TrimSpace(x)
	}

	return splitContents
}

func TestTableOfContents(t *testing.T) {
	println("Pending: TestTableOfContents")
	return

	mdParser := NewMarkdownParser("")
	tableOfContents := mdParser.GenerateTableOfContents()
	splitContents := getSplitContents(tableOfContents)
	if !reflect.DeepEqual(splitContents, []string{"1. Path", "1.1 Примери:"}) {
		t.Fail()
	}

	mdParser = NewMarkdownParser("")
	tableOfContents = mdParser.GenerateTableOfContents()
	splitContents = getSplitContents(tableOfContents)
	if len(splitContents) != 7 || splitContents[3] != "1.3 Colors" {
		t.Fail()
	}

	mdParser = NewMarkdownParser("")
	tableOfContents = mdParser.GenerateTableOfContents()
	splitContents = getSplitContents(tableOfContents)
	if len(splitContents) != 11 || splitContents[9] != "1.1.8 `func (mp *MarkdownParser) GenerateTableOfContents() string`" {
		t.Fail()
	}
}
