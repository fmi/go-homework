package main

import (
	"testing"
)

func TestRepeaterSample(t *testing.T) {
	actual := Repeater("foo", ":")(3)
	expected := "foo:foo:foo"

	if expected != actual {
		t.Errorf("Expected `%s` but got `%s`", expected, actual)
	}
}

func TestMapReducerSample(t *testing.T) {
	sqrtSum := MapReducer(
		func(v int) int { return v * v },
		func(a, v int) int { return a + v },
		0,
	)

	actual := sqrtSum(1, 2, 3, 4)
	expected := 30

	if expected != actual {
		t.Errorf("Expected `%d` but got `%d`", expected, actual)
	}
}
