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

func TestGeneratorSample(t *testing.T) {
	counter := Generator(
		func(v int) int { return v + 1 },
		0,
	)

	var counterResults [4]int
	for i := 0; i < 4; i++ {
		counterResults[i] = counter()
	}

	actual := counterResults
	expected := [4]int{0, 1, 2, 3}

	if expected != actual {
		t.Errorf("Expected `%d` but got `%d`", expected, actual)
	}

	power := Generator(
		func(v int) int { return v * v },
		2,
	)

	var powerResults [4]int
	for i := 0; i < 4; i++ {
		powerResults[i] = power()
	}

	actual = powerResults
	expected = [4]int{2, 4, 16, 256}

	if expected != actual {
		t.Errorf("Expected `%d` but got `%d`", expected, actual)
	}
}

func TestMapReducerSample(t *testing.T) {
	powerSum := MapReducer(
		func(v int) int { return v * v },
		func(a, v int) int { return a + v },
		0,
	)

	actual := powerSum(1, 2, 3, 4)
	expected := 30

	if expected != actual {
		t.Errorf("Expected `%d` but got `%d`", expected, actual)
	}
}
