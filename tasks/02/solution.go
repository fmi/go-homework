package main

import (
	"strings"
)

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

type PieceTable struct {
	parent *PieceTable
	child  *PieceTable
	pieces []Piece
	origin []byte
	add    []byte
}

func NewEditor(text string) Editor {
	return &PieceTable{
		origin: []byte(text),
		pieces: []Piece{{
			origin: true,
			offset: 0,
			length: uint(len(text)),
		}},
	}
}

func (pt *PieceTable) Insert(position uint, text string) Editor {
	var (
		offset = uint(len(pt.add))
		l, r   = pt.split(position)
	)

	pt.add = append(pt.add, text...)
	pieces := make([]Piece, len(l.pieces)+len(r.pieces)+1)
	pieces = append(pieces, l.pieces...)
	pieces = append(pieces, Piece{false, offset, uint(len(text))})
	pieces = append(pieces, r.pieces...)
	return pt.slice(pieces)
}

func (pt *PieceTable) Delete(offset, length uint) Editor {
	l, _ := pt.split(offset)
	_, r := pt.split(offset + length)

	pieces := make([]Piece, len(l.pieces)+len(r.pieces))
	copy(pieces, l.pieces)
	copy(pieces[len(l.pieces):], r.pieces)

	return pt.slice(pieces)
}
func (pt *PieceTable) Undo() Editor {
	if pt.parent == nil {
		return pt
	}

	return pt.parent
}

func (pt *PieceTable) Redo() Editor {
	if pt.child == nil {
		return pt
	}
	return pt.child
}

func (pt PieceTable) String() string {
	var (
		s   strings.Builder
		buf []byte
	)

	for _, p := range pt.pieces {
		if p.origin {
			buf = pt.origin
		} else {
			buf = pt.add
		}

		s.Write(buf[p.offset : p.offset+p.length])
	}
	return s.String()
}

func (pt *PieceTable) split(offset uint) (*PieceTable, *PieceTable) {
	var pos uint

	for i, p := range pt.pieces {
		if pos == offset {
			return pt.slice(pt.pieces[:i]), pt.slice(pt.pieces[i:])
		}

		pos += p.length

		if pos > offset {
			l, r := p.split(offset - (pos - p.length))

			before := make([]Piece, len(pt.pieces[:i])+1)
			copy(before, pt.pieces[:i])
			before[len(before)-1] = l

			after := make([]Piece, 0, len(pt.pieces[i:]))
			after = append(after, r)
			if len(pt.pieces[i:]) > 1 {
				after = append(after, pt.pieces[i+1:]...)
			}

			return pt.slice(before), pt.slice(after)

		}
	}

	// Somewhere after the EOF.
	return pt.slice(pt.pieces), &PieceTable{}
}

func (pt *PieceTable) slice(pieces []Piece) *PieceTable {
	var child = &PieceTable{parent: pt, pieces: pieces, origin: pt.origin, add: pt.add}
	pt.child = child
	return child
}

// Piece is a record in a piece table.
type Piece struct {
	// origin is true if points to the original (read-only) buffer.
	origin bool

	// offset from the start of the used buffer.
	offset uint

	// length of given piece
	length uint
}

func (p Piece) split(pos uint) (Piece, Piece) {
	return Piece{origin: p.origin, offset: p.offset, length: pos},
		Piece{origin: p.origin, offset: p.offset + pos, length: p.length - pos}
}
