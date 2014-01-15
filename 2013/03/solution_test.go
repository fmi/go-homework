package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestHeaders(t *testing.T) {
	mdParser := NewMarkdownParser(`
# One
##One # Two

Two
===

# # Three # Four
`)
	// let's not test whether the solution trims or not
	headers := []string{}
	for _, h := range mdParser.Headers() {
		headers = append(headers, strings.TrimSpace(h))
	}
	expected := []string{"One", "Two", "# Three # Four"}

	if !reflect.DeepEqual(headers, expected) {
		t.Errorf("Not equal:\n  %#v\n  %#v", headers, expected)
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
	expected := []string{"Dee-dee", "Boo-boo"}
	if !reflect.DeepEqual(subHeaders, expected) {
		t.Errorf("Not equal:\n  %#v\n  %#v", subHeaders, expected)
	}

	subHeaders = mdParser.SubHeadersOf("Bla-bla")
	expected = nil
	if !reflect.DeepEqual(subHeaders, expected) {
		t.Errorf("Not equal:\n  %#v\n  %#v", subHeaders, expected)
	}

	subHeaders = mdParser.SubHeadersOf("One-two")
	expected = []string{"Three-four"}
	if !reflect.DeepEqual(subHeaders, expected) {
		t.Errorf("Not equal:\n  %#v\n  %#v", subHeaders, expected)
	}
}

func TestNames(t *testing.T) {
	mdParser := NewMarkdownParser("Beginning Of Line. Аз съм Иван Петров. He Used to play games. His favourite browser is Firefox. That's Mozilla Firefox.")
	names := mdParser.Names()
	expected := []string{"Of Line", "Иван Петров", "Mozilla Firefox"}

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
	mdParser := NewMarkdownParser(`
# Path
Нещо
## Примери:
Още нещо
`)
	tableOfContents := mdParser.GenerateTableOfContents()
	splitContents := getSplitContents(tableOfContents)
	expected := []string{
		"1. Path",
		"1.1 Примери:",
	}

	if !reflect.DeepEqual(splitContents, expected) {
		t.Errorf("Not equal:\n  %#v\n  %#v", splitContents, expected)
	}

	mdParser = NewMarkdownParser(`
One
====
Two
====
## Three

# Four
Five
----
	`)
	tableOfContents = mdParser.GenerateTableOfContents()
	splitContents = getSplitContents(tableOfContents)
	expected = []string{
		"1. One",
		"2. Two",
		"2.1 Three",
		"3. Four",
		"3.1 Five",
	}

	if !reflect.DeepEqual(splitContents, expected) {
		t.Errorf("Not equal:\n  %#v\n  %#v", splitContents, expected)
	}

	mdParser = NewMarkdownParser(`
# One
## Two
### Three
#### Four
##### Five
###### Six
	`)
	tableOfContents = mdParser.GenerateTableOfContents()
	splitContents = getSplitContents(tableOfContents)
	expected = []string{
		"1. One",
		"1.1 Two",
		"1.1.1 Three",
		"1.1.1.1 Four",
		"1.1.1.1.1 Five",
		"1.1.1.1.1.1 Six",
	}

	if !reflect.DeepEqual(splitContents, expected) {
		t.Errorf("Not equal:\n  %#v\n  %#v", splitContents, expected)
	}
}
