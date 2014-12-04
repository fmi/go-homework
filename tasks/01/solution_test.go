package main

import (
	"strconv"
	"testing"
)

func testOutput(t *testing.T, result []string, expected []string) {
	if expected == nil && len(result) != 0 {
		t.Fatalf("We expeced no output but got %+v", result)
	}

	if len(result) != len(expected) {
		t.Fatalf("The result has len %d we expected %d", len(result), len(expected))
	}
	for index, value := range result {
		if value != expected[index] {
			t.Errorf("Expected value on index %d was %s but got %s", index, expected[index], value)
		}
	}
}

func length(s string) string {
	return strconv.Itoa(len(s))
}

func sqrStr(s string) string {
	number, err := strconv.Atoi(s)
	if err != nil {
		return "-1"
	}
	return strconv.Itoa(number * number)
}

func isNumber(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}
func isNotNumber(s string) bool {
	return !isNumber(s)
}

func concat(s1 string, s2 string) string {
	return s1 + s2
}
func revConcatLen(s1 string, s2 string) string {
	return concat(length(s2), length(s1))
}

func TestMap(t *testing.T) {
	input := []string{"1", "two", "three", "four"}
	output := []string{"1", "3", "5", "4"}

	result := Map(input, length)

	testOutput(t, result, output)
}

func TestDoubleMap(t *testing.T) {
	input := []string{"1", "two", "three", "four"}
	output := []string{"1", "9", "25", "16"}
	result := Map(Map(input, length), sqrStr)

	testOutput(t, result, output)
}

func TestFilter(t *testing.T) {
	input := []string{"1", "two", "three", "four"}

	result := Filter(input, func(s string) bool {
		return false
	})

	testOutput(t, result, nil)
}

func TestFilterIsNumber(t *testing.T) {
	input := []string{"1", "two", "three", "four"}
	output := []string{"1"}

	result := Filter(input, isNumber)

	testOutput(t, result, output)
}

func TestReduceEmpty(t *testing.T) {
	result := Reduce([]string{}, concat)

	if result != "" {
		t.Fatalf("expected [%s] got [%s]", "", result)
	}
}

func TestReduceOneElement(t *testing.T) {
	output := "pesho"
	input := []string{output}
	result := Reduce(input, revConcatLen)

	if result != output {
		t.Fatalf("expected [%s] got [%s]", output, result)
	}
}

func TestReduce(t *testing.T) {
	input := []string{"1", "two", "three", "four"}
	output := "42" // by accident

	result := Reduce(input, revConcatLen)

	if result != output {
		t.Errorf("expected [%s] got [%s]", output, result)
	}
}

func TestReduceFilterMap(t *testing.T) {
	input := []string{"1", "two", "three", "four"}
	output := "154"

	result := Reduce(Filter(Map(input, length), func(s string) bool {
		return s != "3"
	}), concat)

	if result != output {
		t.Errorf("Expected [%s] got [%s]", output, result)
	}
}

func TestAny(t *testing.T) {
	input1 := []string{"1", "two", "three", "four"}
	input2 := []string{"one", "two", "three", "four"}

	if !Any(input1, isNumber) {
		t.Errorf("expected true got false")
	}

	if Any(input2, isNumber) {
		t.Errorf("expected false got true")
	}
}

func TestAll(t *testing.T) {
	input1 := []string{"1", "two", "three", "four"}
	input2 := []string{"one", "two", "three", "four"}

	if All(input1, isNotNumber) {
		t.Errorf("expected false got true")
	}

	if !All(input2, isNotNumber) {
		t.Errorf("expected true got false")
	}
}
