package main

import (
	"testing"
)

func TestSumOfSquares(t *testing.T) {
	square := func(a int) int {
		return a * a
	}

	found := mapSum(square, 1, 2, 3, 4, 5)
	expected := 55

	if found != expected {
		t.Errorf("Expected %d but found %d", expected, found)
	}
}
