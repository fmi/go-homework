package main

import (
	"testing"
)

func TestSample(t *testing.T) {
	t.Run("origin", func(t *testing.T) {
		f := NewEditor("foobar")
		compare(t, "foobar", f.String())
	})

	t.Run("insert", func(t *testing.T) {
		f := NewEditor("foobar")
		compare(t, "fobazobar", f.Insert(2, "baz").String())
	})

	t.Run("insertInPiece", func(t *testing.T) {
		f := NewEditor("foo")
		compare(t, "foobar", f.Insert(3, "bar").String())
		compare(t, "foobazbar", f.Insert(4, "azb").String())
	})

	t.Run("append", func(t *testing.T) {
		f := NewEditor("foobar")
		compare(t, "foobarbaz", f.Insert(6, "baz").String())
	})

	t.Run("delete", func(t *testing.T) {
		f := NewEditor("foobar")
		compare(t, "far", f.Delete(1, 3).String())
	})

	t.Run("deleteFull", func(t *testing.T) {
		f := NewEditor("foobar")
		compare(t, "", f.Delete(0, 6).String())
	})

	t.Run("insert_deleteFirst", func(t *testing.T) {
		f := NewEditor("foobar")
		compare(t, "foobarbaz", f.Insert(6, "baz").String())
		compare(t, "baz", f.Delete(0, 6).String())
	})

	t.Run("insert_deletePartial", func(t *testing.T) {
		f := NewEditor("foobar")
		compare(t, "foobarbaz", f.Insert(6, "baz").String())
		compare(t, "foo", f.Delete(3, 6).String())
	})

	t.Run("insert_deleteMultiplePieces", func(t *testing.T) {
		f := NewEditor("bar")
		compare(t, "bazbarfoofoo",
			f.Insert(0, "baz").
				Insert(100, "foo").
				Insert(100, "foo").
				String())
		compare(t, "boo", f.Delete(1, 9).String())
	})

	t.Run("undo", func(t *testing.T) {
		f := NewEditor("bar")
		compare(t, "bar",
			f.Insert(0, "foo").
				Undo().
				Undo().
				String())
	})

	t.Run("redo", func(t *testing.T) {
		f := NewEditor("bar")
		compare(t, "foobar",
			f.Insert(0, "foo").
				Undo().
				Redo().
				Redo().
				String())
	})

	t.Run("null_undo", func(t *testing.T) {
		f := NewEditor("foobar")
		compare(t, "foobar", f.Undo().String())
	})

	t.Run("null_redo", func(t *testing.T) {
		f := NewEditor("foobar")
		compare(t, "foobar", f.Redo().String())
	})

	t.Run("multiple_undo", func(t *testing.T) {
		f := NewEditor("foobar")
		compare(t, "foobar",
			f.Insert(0, "baz").
				Insert(0, "boo").
				Undo().
				Undo().
				String())
	})

	t.Run("multiple_redo", func(t *testing.T) {
		f := NewEditor("foobar")
		compare(t, "boobazfoobar",
			f.Insert(0, "baz").
				Insert(0, "boo").
				Undo().
				Undo().
				Redo().
				Redo().
				String())
	})

	t.Run("delete_long", func(t *testing.T) {
		f := NewEditor("foobar")
		compare(t, "foobar", f.Delete(100, 10).String())
		compare(t, "foo", f.Delete(3, 100).String())
	})
}

func compare(t *testing.T, exp, got string) {
	if got != exp {
		t.Errorf("Expect: %q; got %q", exp, got)
	}
}
