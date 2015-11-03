package main

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

func ExtractColumn(logContents string, column uint8) string {

	// A simple input check. No wrong input is expected but it is always wise not to ignore
	// the possibility altogether.
	if column >= 3 {
		return ""
	}

	// A buffer in which will be stored the output string.
	var output bytes.Buffer

	// Making an io.Reader out of the input string.
	logReader := strings.NewReader(logContents)

	// bufio.Scanner makes possible reading the input line by line. Basically every
	// Scan call tries to parse the next line and returns whether there was a line
	// or the input has finished.
	scanner := bufio.NewScanner(logReader)

	for scanner.Scan() {
		line := scanner.Text()

		// Note the SplitN is called rather than Split. Using SplitN makes sure there will not
		// be any problems with spaces which are found after the initial 3 spaces, defined
		// by the input format.
		splitted := strings.SplitN(line, " ", 4)

		if len(splitted) < 4 {
			// The log line is malformed somehow. Skipping it. A well formed line must have
			// at least 3 saces: one in the Date column, one between the date and the IP and
			// one after the IP, just before the text.
			continue
		}

		// We have to extract only the asked column now.
		if column == 0 {
			// Note that the SplitN call has splitted the datetime into two different strings
			// since there is a space between the two parts of the column. Remember,
			// the format was "<date> <time>".
			output.WriteString(fmt.Sprintf("%s %s\n", splitted[0], splitted[1]))
		} else {
			// The WriteString call appends its argument to the buffer. Making sure to add the
			// new line at the end of the output line.
			// Note that indexes in splitted are one off because of the space in the date-time
			// column. See the previous comment.
			output.WriteString(fmt.Sprintf("%s\n", splitted[int(column)+1]))
		}
	}

	// Unsafely ignoring the scanner.Err() result.

	// Buffer is not an actual string, but something else. Its method String returns a
	// string represetation of itself.
	return output.String()
}
