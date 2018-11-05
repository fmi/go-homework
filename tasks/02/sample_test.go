package main

import "testing"

func TestSample(t *testing.T) {
	t.Run("origin", func(t *testing.T) {
		f := NewEditor("foobar")
		compare(t, "foobar", f.String())
	})

	t.Run("insert", func(t *testing.T) {
		f := NewEditor("foobar")
		compare(t, "fobazobar", f.Insert(2, "baz").String())
	})

	t.Run("insert_out_of_bounds_position", func(t *testing.T) {
		f := NewEditor("foo").Insert(453, ".")
		compare(t, "foo.", f.String())
	})

	t.Run("append", func(t *testing.T) {
		f := NewEditor("foobar")
		compare(t, "foobarbaz", f.Insert(6, "baz").String())
	})

	t.Run("insert_append_front_and_append_back", func(t *testing.T) {
		f := NewEditor("A large span of text")
		f.Insert(16, "an English ").Insert(2, "very ").Insert(36, " message.").Insert(0, "This is ")
		compare(t, "This is A very large span of an English text message.", f.String())
	})

	t.Run("delete", func(t *testing.T) {
		f := NewEditor("foobar")
		compare(t, "far", f.Delete(1, 3).String())
	})

	t.Run("delete_out_of_bounds_offset", func(t *testing.T) {
		f := NewEditor("foo").Delete(300, 1)
		compare(t, "foo", f.String())
	})

	t.Run("delete_out_of_bounds_length", func(t *testing.T) {
		f := NewEditor("foo").Delete(1, 3)
		compare(t, "f", f.String())
	})

	t.Run("delete_where_single_partial_piece_is_affected", func(t *testing.T) {
		f := NewEditor("A large span of text")
		f.Insert(16, "an English ").Insert(2, "very ").Insert(36, " message.").Insert(0, "This is ")
		f.Delete(12, 2)
		compare(t, "This is A ve large span of an English text message.", f.String())
	})

	t.Run("delete_where_single_whole_piece_is_affected", func(t *testing.T) {
		f := NewEditor("A large span of text")
		f.Insert(16, "an English ").Insert(2, "very ").Insert(36, " message.").Insert(0, "This is ")
		f.Delete(10, 5)
		compare(t, "This is A large span of an English text message.", f.String())
	})

	t.Run("delete_where_adjacent_pieces_are_affected", func(t *testing.T) {
		f := NewEditor("A large span of text")
		f.Insert(16, "an English ").Insert(2, "very ").Insert(36, " message.").Insert(0, "This is ")
		f.Delete(12, 8)
		compare(t, "This is A ve span of an English text message.", f.String())
	})

	t.Run("delete_where_multiple_pieces_are_affected", func(t *testing.T) {
		f := NewEditor("A span of text")
		f.Insert(10, "English ")
		f.Delete(1, 20)
		compare(t, "At", f.String())
	})

	t.Run("delete_where_multiple_pieces_are_affected_multiple_inserts", func(t *testing.T) {
		f := NewEditor("A large span of text")
		f.Insert(16, "an English ").Insert(2, "very ").Insert(36, " message.").Insert(0, "This is ")
		f.Delete(12, 27)
		compare(t, "This is A ve text message.", f.String())
	})

	t.Run("undo", func(t *testing.T) {
		f := NewEditor("A span of text")
		f.Insert(10, "English ").Insert(0, "This is ").Undo()
		compare(t, "A span of English text", f.String())
	})

	t.Run("undo_original", func(t *testing.T) {
		f := NewEditor("A span of text").Undo().Undo().Undo()
		compare(t, "A span of text", f.String())
	})

	t.Run("redo", func(t *testing.T) {
		f := NewEditor("A span of text")
		f.Insert(10, "English ").Insert(0, "This is ").Undo().Undo().Redo()
		compare(t, "A span of English text", f.String())
	})

	t.Run("redundant_redo", func(t *testing.T) {
		f := NewEditor("A span of text")
		f.Insert(10, "English ").Undo().Redo().Redo().Redo()
		compare(t, "A span of English text", f.String())
	})

	t.Run("edit_after_undo_should_invalidate_redo", func(t *testing.T) {
		f := NewEditor("A large span of text")
		f.Insert(16, "an English ").Insert(2, "very ").Insert(36, " message.").Insert(0, "This is ")
		f.Delete(12, 27)
		f.Undo().Undo().Undo().Undo().Undo()
		f.Insert(20, " message").Redo().Redo().Redo()
		compare(t, "A large span of text message", f.String())
	})
}

func compare(t *testing.T, exp, got string) {
	if got != exp {
		t.Errorf("Expect: %q; got %q", exp, got)
	}
}
