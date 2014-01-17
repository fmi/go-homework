package main

import (
	"testing"
	"reflect"
)

func TestAllStuck(t *testing.T) {
	set := make(map[[2][2]int]bool)
	for x := 0; x < 4; x++ {
		for y := 0; y < 4; y++ {
			var stuck = [2][2]int { { x, y }, { -2, -2 } }
			set[stuck] = true
		}
	}

	var allStuck = [4][4]rune {
		{'X', 'X', 'X', 'X'},
		{'X', 'X', 'X', 'X'},
		{'X', 'X', 'X', 'X'},
		{'X', 'X', 'X', 'X'},
	}
	resultSet := make(map[[2][2]int]bool)
	for _, value := range playMall(allStuck) {
		resultSet[value] = true
	}

	if !reflect.DeepEqual(resultSet, set) {
		t.Fail()
	}
}

func TestSingleDweller(t *testing.T) {
	var singleDweller = [4][4]rune{
		{'-', '-', '-', '-'},
		{'-', '-', '-', '-'},
		{'-', '-', '-', '-'},
		{'-', '-', '-', 'X'},
	}
	result := playMall(singleDweller)

	j := 0
	for i := 103; i > 3; i-- {
		start := i % 4
		end   := (i - 1) % 4
		var currentElement = [2][2]int { { start, start }, { end, end } }
		if !reflect.DeepEqual(result[j], currentElement) {
			t.Fail()
		}
		j++
	}

	if j != len(result) - 1 {
		t.Fail()
	}

	var currentElement = [2][2]int { { 3, 3 }, { -1, -1 } }
	if !reflect.DeepEqual(result[j], currentElement) {
		t.Fail()
	}
}
