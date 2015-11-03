package main

import (
	"testing"
)

func TestWithTheExampleTest(t *testing.T) {
	logContents := `2015-08-23 12:37:03 8.8.8.8 As far as we can tell this is a DNS
2015-08-23 12:37:04 8.8.4.4 Yet another DNS, how quaint!
2015-08-23 12:37:05 208.122.23.23 There is definitely some trend here
`

	expected := `8.8.8.8
8.8.4.4
208.122.23.23
`

	test(t, expected, logContents, 1)
}

func TestDifferentColumns(t *testing.T) {
	logContents := `2015-08-23 12:37:03 8.8.8.8 As far as we can tell this is a DNS
2015-08-23 12:37:04 8.8.4.4 Yet another DNS, how quaint!
2015-08-23 12:37:05 208.122.23.23 There is definitely some trend here
1988-02-27 23:59:59 208.122.23.24 Ами други езици?
1987-01-31 23:59:59 208.122.23.24 日本に行きたい。変なひと
0172-03-25 06:55:10 127.0.0.1 Φραγκολεβαντίνικα
0001-01-01 00:00:00 255.255.255.255 Nothing particularly interesting happened
`

	columnTests := []struct {
		column   uint8
		expected string
	}{
		{0, `2015-08-23 12:37:03
2015-08-23 12:37:04
2015-08-23 12:37:05
1988-02-27 23:59:59
1987-01-31 23:59:59
0172-03-25 06:55:10
0001-01-01 00:00:00
`},
		{1, `8.8.8.8
8.8.4.4
208.122.23.23
208.122.23.24
208.122.23.24
127.0.0.1
255.255.255.255
`},
		{2, `As far as we can tell this is a DNS
Yet another DNS, how quaint!
There is definitely some trend here
Ами други езици?
日本に行きたい。変なひと
Φραγκολεβαντίνικα
Nothing particularly interesting happened
`},
	}

	for _, tt := range columnTests {
		test(t, tt.expected, logContents, tt.column)
	}
}

func TestWithEmptyLog(t *testing.T) {
	for i := 0; i < 3; i++ {
		test(t, "", "", uint8(i))
	}
}

func TestWithOneLiner(t *testing.T) {
	var (
		expected    = "127.0.0.1\n"
		logContents = "2015-08-23 12:37:05 127.0.0.1 There!\n"
	)
	test(t, expected, logContents, 1)
}

func TestSpacesAtTheStartOrEndOfALine(t *testing.T) {

	logContents := `2015-08-23 12:37:03 8.8.8.8 As far as we can tell this is a DNS
2015-08-23 12:37:04 8.8.4.4 Yet another DNS, how quaint!
2015-08-23 12:37:05 208.122.23.23 There is definitely some trend here
2015-08-22 12:37:06 208.122.244.221  Some spaces at the beginning of this line
0001-01-01 00:00:00 255.255.255.255 Nothing particularly interesting happened `

	expectedForColumn := []string{
		`2015-08-23 12:37:03
2015-08-23 12:37:04
2015-08-23 12:37:05
2015-08-22 12:37:06
0001-01-01 00:00:00
`,
		`8.8.8.8
8.8.4.4
208.122.23.23
208.122.244.221
255.255.255.255
`,
		`As far as we can tell this is a DNS
Yet another DNS, how quaint!
There is definitely some trend here
 Some spaces at the beginning of this line
Nothing particularly interesting happened 
`,
	}

	for column, expected := range expectedForColumn {
		test(t, expected, logContents, uint8(column))
	}

}

func TestNoNewLineAtEndOfInput(t *testing.T) {
	logContents := `2015-08-23 12:37:03 8.8.8.8 As far as we can tell this is a DNS
2015-08-23 12:37:05 208.122.23.23 There is definitely some trend here`

	expected := `As far as we can tell this is a DNS
There is definitely some trend here
`

	test(t, expected, logContents, 2)
}

func TestIPOrDateAtTheEndOfALine(t *testing.T) {
	logContents := `2015-08-23 12:37:03 8.8.8.8 2015-08-23 12:37:03
2015-08-23 12:37:05 208.122.23.23 8.8.4.4
`

	expectedByColumn := []string{
		`2015-08-23 12:37:03
2015-08-23 12:37:05
`,
		`8.8.8.8
208.122.23.23
`,
		`2015-08-23 12:37:03
8.8.4.4
`,
	}

	for column, expected := range expectedByColumn {
		test(t, expected, logContents, uint8(column))
	}

}

// I did not intend to write such a test, but I am forced now since I was made to
// say this explicitly: http://fmi.golang.bg/topics/69
func TestWithOnlyOneNewLine(t *testing.T) {
	logContents := "\n"
	expected := ""

	for i := 0; i < 3; i++ {
		test(t, expected, logContents, uint8(i))
	}
}

// The following tests are imported from students' repos, mostly
// https://gist.github.com/tdhris/

func TestExtractingIPs(t *testing.T) {
	logContents := `2015-08-23 12:37:03 8.8.8.8 As far as we can tell this is a DNS
2015-08-23 12:37:04 8.8.4.4 Yet another DNS, how quaint!
2015-08-23 12:37:05 208.122.23.23 There is definitely some trend here
`

	expected := `8.8.8.8
8.8.4.4
208.122.23.23
`

	test(t, expected, logContents, 1)
}

func TestExtractingTimes(t *testing.T) {
	logContents := `2015-08-23 12:37:03 8.8.8.8 As far as we can tell this is a DNS
2015-08-23 12:37:04 8.8.4.4 Yet another DNS, how quaint!
2015-08-23 12:37:05 208.122.23.23 There is definitely some trend here
`

	expected := `2015-08-23 12:37:03
2015-08-23 12:37:04
2015-08-23 12:37:05
`

	test(t, expected, logContents, 0)
}

func TestExtractingTexts(t *testing.T) {
	logContents := `2015-08-23 12:37:03 8.8.8.8 As far as we can tell this is a DNS
2015-08-23 12:37:04 8.8.4.4 Yet another DNS(8.8.4.4), how quaint!
2015-08-23 12:37:05 208.122.23.23 There is definitely some trend here
`

	expected := `As far as we can tell this is a DNS
Yet another DNS(8.8.4.4), how quaint!
There is definitely some trend here
`

	test(t, expected, logContents, 2)
}

func TestLogDoesNotEndInNewLine(t *testing.T) {
	logContents := `2015-08-23 12:37:03 8.8.8.8 As far as we can tell this is a DNS
2015-08-23 12:37:04 8.8.4.4 Yet another DNS, how quaint!
2015-08-23 12:37:05 208.122.23.23 There is definitely some trend here`

	expected := `8.8.8.8
8.8.4.4
208.122.23.23
`

	test(t, expected, logContents, 1)
}

func TestLogLogLineEndsInIP(t *testing.T) {
	logContents := `2015-08-23 12:37:03 8.8.8.8 As far as we can tell this is a DNS 8.8.3.8
2015-08-23 12:37:04 8.8.4.4 Yet another DNS, how quaint!
2015-08-23 12:37:05 208.122.23.23 There is definitely some trend here`

	expected := `8.8.8.8
8.8.4.4
208.122.23.23
`

	test(t, expected, logContents, 1)
}

func TestWithSpaces(t *testing.T) {
	logContents := `2015-08-23 12:37:03 8.8.8.8 spaces   tabs		end`

	expected := `spaces   tabs		end
`

	test(t, expected, logContents, 2)
}

func TestMoreLinesThanExample(t *testing.T) {
	logContents := `2015-08-23 12:37:03 8.8.8.8 As far as we can tell this is a DNS
2015-08-23 12:37:04 8.8.4.4 Yet another DNS, how quaint!
2015-08-23 12:37:05 208.122.23.23 There is definitely some trend here
2015-10-22 08:22:05 127.0.0.1 A campus crashes in the rainbow!
2015-10-22 08:22:06 127.0.0.1 Localhost?? Something is wrong here
2015-10-22 08:22:07 127.0.0.1 The amber libel flies a pope.
2015-10-22 08:22:07 42.42.42.42 Time is an illusion. Lunchtime doubly so.
`

	expected := `8.8.8.8
8.8.4.4
208.122.23.23
127.0.0.1
127.0.0.1
127.0.0.1
42.42.42.42
`

	test(t, expected, logContents, 1)
}

func test(t *testing.T, expected, logContents string, column uint8) {
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Errorf("There was a panic while testing: %s", recovered)
		}
	}()

	found := ExtractColumn(logContents, column)

	if found != expected {
		t.Errorf("Expected\n---\n%s\n---\nbut found\n---\n%s\n---\n", expected, found)
	}
}
