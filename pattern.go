package main

import (
	"fmt"
	"github.com/lightning/lightning"
)

// Pattern defines a pattern of notes.
// Notes that you add to a pattern at a position will overwrite
// any notes at that position with the same number.
type Pattern struct {
	Length int                 `json:"length"`
	Notes  [][]*lightning.Note `json:"notes"`
}

type PatternEdit struct {
	Pos  uint64             `json:"pos"`
	Note *lightning.Note `json:"note"`
}

func (this *Pattern) indexTooLarge(pos uint64) error {
	str := "pos (%d) greater than pattern length (%d)"
	return fmt.Errorf(str, pos, this.Length)
}

// NotesAt returns a slice representing the notes
// that are stored at a particular position in a pattern.
// pos modulo the size of the pattern is the actual index into
// the pattern.
func (this *Pattern) NotesAt(pos uint64) []*lightning.Note {
	notes := len(this.Notes)
	return this.Notes[int(pos)%notes]
}

// AddTo adds a Note to the pattern at pos
func (this *Pattern) AddTo(pos uint64, note *lightning.Note) error {
	if pos >= uint64(this.Length) {
		return this.indexTooLarge(pos)
	}
	// try to insert to a nil position
	inserted := false
	notes := this.Notes[int(pos)]
	for i, n := range notes {
		if n == nil {
			inserted = true
			notes[i] = note
		}
	}
	// if there were no nil positions, append
	if !inserted {
		this.Notes[int(pos)] = append(notes, note)
	}
	return nil
}

// RemoveFrom removes a note from a particular position in a pattern
func (this *Pattern) RemoveFrom(pos uint64, note *lightning.Note) error {
	if pos >= uint64(this.Length) {
		return this.indexTooLarge(pos)
	}
	// remove a note with the same sample and same number, if one exists
	notes := this.Notes[int(pos)]
	for i, n := range notes {
		if n != nil && n.Number == note.Number && n.Sample == note.Sample {
			notes[i] = nil
		}
	}
	return nil
}

// Clear removes all the notes at a given position
// in the pattern.
func (this *Pattern) Clear(pos uint64) error {
	if pos >= uint64(this.Length) {
		return this.indexTooLarge(pos)
	}
	this.Notes[pos] = make([]*lightning.Note, 0)
	return nil
}

// NewPattern creates a Pattern with the specified size
func NewPattern(size int) *Pattern {
	return &Pattern{
		size,
		make([][]*lightning.Note, size),
	}
}
