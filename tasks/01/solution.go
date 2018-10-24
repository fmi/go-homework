package main

import (
	"strings"
)

func Repeater(s, sep string) func(int) string {
	return func(n int) string {
		return strings.TrimSuffix(strings.Repeat(s+sep, n), sep)
	}
}

func MapReducer(mapper func(int) int, reducer func(int, int) int, initial int) func(...int) int {
	return func(args ...int) int {
		acc := initial
		for i := 0; i < len(args); i++ {
			acc = reducer(acc, mapper(args[i]))
		}
		return acc
	}
}

func Generator(gen func(int) int, initial int) func() int {
	next := initial
	return func() int {
		current := next
		next = gen(current)
		return current
	}
}
