package main

import "testing"

func areEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func checkFail(t *testing.T, resultWords, expectedWords []string) {
	if !areEqual(resultWords, expectedWords) {
		t.Errorf("Expected words %v but result was %v", expectedWords, resultWords)
	}
	if resultWords == nil {
		t.Errorf("Function returned uninitialized slice!")
	}
}

func TestWithExampleData(t *testing.T) {
	text := "A Be Ce DE De a! oh? a oy! oY! OY!"

	checkFail(t, getMostCommonWords(text, 4), []string{})
	checkFail(t, getMostCommonWords(text, 3), []string{"oy!"})
	checkFail(t, getMostCommonWords(text, 2), []string{"a", "de", "oy!"})
	checkFail(t, getMostCommonWords(text, 1), []string{"a", "a!", "be", "ce", "de", "oh?", "oy!"})
	checkFail(t, getMostCommonWords(text, 0), []string{"a", "a!", "be", "ce", "de", "oh?", "oy!"})
}
