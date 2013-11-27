package main

import (
	"io/ioutil"
	"strings"
	"testing"
)

func loadTheReadme() string {
	content, err := ioutil.ReadFile("./README.md")
	if err != nil {
		return ""
	}
	return string(content)
}

func TestHeaders(t *testing.T) {
	mdParser := NewMarkdownParser(loadTheReadme())
	headers := mdParser.Headers()

	if headers[0] != "MarkdownParser" {
		t.Fail()
	}
}

func TestSubHeadersOf(t *testing.T) {
	mdParser := NewMarkdownParser(loadTheReadme())
	subHeaders := mdParser.SubHeadersOf("Пример")
	if len(subHeaders) != 0 {
		t.Fail()
	}
}

func TestTableOfContents(t *testing.T) {
	mdParser := NewMarkdownParser(loadTheReadme())
	tableOfContents := mdParser.GenerateTableOfContents()

	if strings.Split(tableOfContents, "\n")[4] != "1.1.3 `func (mp *MarkdownParser) SubHeadersOf(header string) []string`" {
		t.Fail()
	}
}
