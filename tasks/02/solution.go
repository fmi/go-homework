package main

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
