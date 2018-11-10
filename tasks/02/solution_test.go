package main

import "testing"

func TestExampleFromReadme(t *testing.T) {
	var f = NewEditor("A large span of text")

	f = f.Insert(16, "English ")
	compare(t, "A large span of English text", f.String())

	f = f.Delete(2, 6)
	compare(t, "A span of English text", f.String())

	f = f.Delete(10, 8)
	compare(t, "A span of text", f.String())
}

func TestOutOfBound(t *testing.T) {
	var f = NewEditor("A span of text")

	compare(t, f.String()+"!", f.Insert(150, "!").String())
	compare(t, f.String(), f.Delete(150, 20).String())
	compare(t, "A span of", f.Delete(9, 200).String())
}

func TestUndo(t *testing.T) {
	var f = NewEditor("A large span of text")
	compare(t, f.String(), f.Undo().String())

	f = f.Insert(16, "English ")
	compare(t, f.String(), f.Delete(2, 6).Undo().String())
}

func TestSeveralUndos(t *testing.T) {
	var f = NewEditor("A large span of text").
		Insert(16, "English ").
		Delete(2, 6).Delete(10, 8)

	compare(t, "A span of text", f.String())
	compare(t, "A large span of text", f.Undo().Undo().Undo().String())
}

func TestRedo(t *testing.T) {
	var f = NewEditor("A large span of text").
		Insert(16, "English ").Delete(2, 6)

	compare(t, f.String(), f.Undo().Redo().String())
}

func TestSeveralRedos(t *testing.T) {
	var f = NewEditor("A large span of text").
		Insert(16, "English ").Delete(2, 6).Delete(10, 8).
		Undo().Undo().Undo()

	compare(t, "A large span of text", f.String())
	compare(t, "A span of text", f.Redo().Redo().Redo().String())
}

func TestOpAfterUndoInvalidatesRedo(t *testing.T) {
	var f = NewEditor("A large span of text").
		Insert(16, "English ").Undo().Delete(0, 2)

	compare(t, f.String(), f.Redo().String())
}

func TestUnicode(t *testing.T) {
	var f = NewEditor("Жълтата дюля беше щастлива и замръзна като гьон.").
		Delete(49, 3).Insert(49, ", че пухът, който цъфна,")

	compare(t, "Жълтата дюля беше щастлива, че пухът, който цъфна, замръзна като гьон.", f.String())
}
