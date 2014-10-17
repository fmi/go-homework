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
