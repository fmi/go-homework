package main

import (
	"math"
	"testing"
)

func TestExamplesInReadme(t *testing.T) {

	t.Run("Repeater", func(t *testing.T) {
		actual := Repeater("foo", ":")(3)
		expected := "foo:foo:foo"

		if expected != actual {
			t.Errorf("Expected `%s` but got `%s`", expected, actual)
		}
	})

	t.Run("MapReducer", func(t *testing.T) {
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
	})

	t.Run("Generator", func(t *testing.T) {
		counter := Generator(
			func(v int) int { return v + 1 },
			0,
		)
		power := Generator(
			func(v int) int { return v * v },
			2,
		)

		actual := []int{
			counter(),
			counter(),
			power(),
			power(),
			counter(),
			power(),
			counter(),
			power(),
		}
		expected := []int{0, 1, 2, 4, 2, 16, 3, 256}

		for i := 0; i < len(expected); i++ {
			if expected[i] != actual[i] {
				t.Errorf("At index %d expected %d but got %d", i, expected[i], actual[i])
			}
		}
	})

}

func TestRepeater(t *testing.T) {
	tests := []struct {
		description string
		str         string
		sep         string
		n           int
		expected    string
	}{
		{
			description: "simple usage",
			str:         "gopher",
			sep:         ",",
			n:           5,
			expected:    "gopher,gopher,gopher,gopher,gopher",
		},
		{
			description: "empty repeated string",
			str:         "",
			sep:         ",",
			n:           5,
			expected:    ",,,,",
		},
		{
			description: "empty separator",
			str:         "gopher",
			sep:         "",
			n:           3,
			expected:    "gophergophergopher",
		},
		{
			description: "zero repeats",
			str:         "gopher",
			sep:         ":",
			n:           0,
			expected:    "",
		},
		{
			description: "one repeats",
			str:         "gopher",
			sep:         "-",
			n:           1,
			expected:    "gopher",
		},
		{
			description: "s and sep are the same",
			str:         "gopher",
			sep:         "gopher",
			n:           2,
			expected:    "gophergophergopher",
		},
		{
			description: "unicode",
			str:         "日本",
			sep:         "五",
			n:           5,
			expected:    "日本五日本五日本五日本五日本",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actual := Repeater(test.str, test.sep)(test.n)

			if test.expected != actual {
				t.Errorf("Expected `%s` but got `%s`", test.expected, actual)
			}
		})
	}
}

func TestRepeaterMultipleCalls(t *testing.T) {
	expected := "gopher,gopher,gopher"

	repeater := Repeater("gopher", ",")
	firstCall := repeater(3)
	secondCall := repeater(3)

	if firstCall != expected {
		t.Errorf("Expected `%s` but got `%s` after 1st call", expected, firstCall)
	}

	if secondCall != expected {
		t.Errorf("Expected `%s` but got `%s` after 2nd call", expected, secondCall)
	}
}

func TestGenerator(t *testing.T) {
	tests := []struct {
		description string
		gen         func(int) int
		initial     int
		expected    []int
	}{
		{
			description: "first call returns initial",
			gen:         func(n int) int { return n },
			initial:     2,
			expected:    []int{2},
		},
		{
			description: "pretty good random generator",
			gen:         func(_ int) int { return 4 },
			initial:     4,
			expected:    []int{4, 4, 4, 4, 4, 4},
		},
		{
			description: "powers of two",
			gen: func() func(int) int {
				last := 1
				return func(n int) int {
					last += n
					return last
				}
			}(),
			initial:  1,
			expected: []int{1, 2, 4, 8, 16, 32, 64, 128, 256},
		},
		{
			description: "fibbonaci",
			gen: func() func(int) int {
				last := 0
				return func(n int) int {
					next := last + n
					last = n
					return next
				}
			}(),
			initial:  1,
			expected: []int{1, 1, 2, 3, 5, 8, 13, 21, 34, 55},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			gen := Generator(test.gen, test.initial)

			for i, expected := range test.expected {
				actual := gen()
				if expected != actual {
					t.Errorf("At index %d expected %d but got %d", i, expected, actual)
				}
			}

		})
	}
}

func TestMapReducer(t *testing.T) {
	tests := []struct {
		description string
		mapFunc     func(int) int
		reduceFunc  func(int, int) int
		initial     int
		args        []int
		expected    int
	}{
		{
			description: "sqrtMap summer",
			mapFunc: func(n int) int {
				return int(math.Sqrt(float64(n)))
			},
			reduceFunc: func(a, b int) int {
				return a + b
			},
			initial:  0,
			args:     []int{1, 4, 9, 16, 25, 36, 49, 64},
			expected: 1 + 2 + 3 + 4 + 5 + 6 + 7 + 8,
		},
		{
			description: "biggest left shift",
			mapFunc: func(n int) int {
				return n << 1
			},
			reduceFunc: func(a, b int) int {
				if a > b {
					return a
				}
				return b
			},
			initial:  0,
			args:     []int{5, 21, 10, 25, 22, 18, 18, 19, 23},
			expected: 50,
		},
		{
			description: "left shift all the way",
			mapFunc: func(n int) int {
				return n
			},
			reduceFunc: func(a, b int) int {
				return a << uint(b)
			},
			initial:  42,
			args:     []int{1, 2, 3, 4, 5, 6},
			expected: 88080384,
		},
		{
			description: "initial is used first",
			mapFunc: func(n int) int {
				return n * n
			},
			reduceFunc: func(a, b int) int {
				return a * b
			},
			initial:  20,
			args:     []int{},
			expected: 20,
		},
		{
			description: "left foldedness",
			mapFunc: func(n int) int {
				return n * n
			},
			reduceFunc: func(a, b int) int {
				return b / a
			},
			initial:  1,
			args:     []int{2, 4, 8, 16, 32, 64},
			expected: 64,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			mreducer := MapReducer(test.mapFunc, test.reduceFunc, test.initial)
			actual := mreducer(test.args...)
			if test.expected != actual {
				t.Errorf("Expected %d but got %d", test.expected, actual)
			}
		})
	}
}

func TestMapReducerMultipleCalls(t *testing.T) {
	mreducer := MapReducer(
		func(n int) int {
			return int(math.Sqrt(float64(n)))
		},
		func(a, b int) int {
			return a + b
		},
		9,
	)
	args := []int{1, 4, 9, 16, 25, 36, 49, 64}
	expected := 1 + 2 + 3 + 4 + 5 + 6 + 7 + 8 + 9

	firstCall := mreducer(args...)
	secondCall := mreducer(args...)

	if expected != firstCall {
		t.Errorf("Expected %d but got %d after 1st call", expected, firstCall)
	}

	if expected != secondCall {
		t.Errorf("Expected %d but got %d after 2nd call", expected, secondCall)
	}
}
