package main

import "testing"

type Editor interface {
	// Insert text starting from given position.
	Insert(position uint, text string) Editor

	// Delete length items from offset.
	Delete(offset, length uint) Editor

	// Undo reverts latest change.
	Undo() Editor

	// Redo re-applies latest undone change.
	Redo() Editor

	// String returns complete representation of what a file looks
	// like after all manipulations.
	String() string
}

func TestSample(t *testing.T) {
	var f = NewEditor("foobar")

	t.Run("origin", func(t *testing.T) {
		compare(t, "foobar", f.String())
	})

	t.Run("insert", func(t *testing.T) {
		compare(t, "fobazobar", f.Insert(2, "baz").String())
	})

	t.Run("append", func(t *testing.T) {
		compare(t, "foobarbaz", f.Insert(6, "baz").String())
	})

	t.Run("delete", func(t *testing.T) {
		compare(t, "far", f.Delete(1, 3).String())
	})
}
