package main

import (
	"testing"
)

func TestWithReadmeExample(t *testing.T) {
	var expected uint64 = 2640
	found := SquareSumDifference(10)

	if found != expected {
		t.Errorf("Expected %d but found %d for n=10", expected, found)
	}
}
