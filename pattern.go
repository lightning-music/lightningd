package main

import (
	"fmt"
	"github.com/lightning/go"
)

// Tempo in bpm
type Tempo uint64

// Pattern defines a pattern of notes.
// Notes that you add to a pattern at a position will overwrite
// any notes at that position with the same number.
type Pattern struct {
	Length int                `json:"length"`
	Notes  [][]*lightning.Note `json:"notes"`
}

type PatternEdit struct {
	Pos  Pos            `json:"pos"`
	Note *lightning.Note `json:"note"`
}

func (this *Pattern) indexTooLarge(pos Pos) error {
	str := "pos (%d) greater than pattern length (%d)"
	return fmt.Errorf(str, pos, this.Length)
}

func (this *Pattern) indexNegative(pos Pos) error {
	return fmt.Errorf("pos (%d) less than 0", pos)
}

// NotesAt returns a slice representing the notes
// that are stored at a particular position in a pattern.
func (this *Pattern) NotesAt(pos Pos) []*lightning.Note {
	notes := len(this.Notes)
	return this.Notes[int(pos)%notes]
}

// AddTo adds a Note to the pattern at pos
func (this *Pattern) AddTo(pos Pos, note *lightning.Note) error {
	if int(pos) >= this.Length {
		return this.indexTooLarge(pos)
	}
	if int(pos) < 0 {
		return this.indexNegative(pos)
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
		this.Notes[int(pos)] = append(this.Notes[int(pos)], note)
	}
	return nil
}

// RemoveFrom removes a note from a particular position in a pattern
func (this *Pattern) RemoveFrom(pos Pos, note *lightning.Note) error {
	if int(pos) >= this.Length {
		return this.indexTooLarge(pos)
	}
	if int(pos) < 0 {
		return this.indexNegative(pos)
	}
	// remove a note with the same number, if one exists
	notes := this.Notes[int(pos)]
	for i, n := range notes {
		if n.Number == note.Number {
			notes[i] = nil
		}
	}
	return nil
}

// Clear removes all the notes at a given position
// in the pattern.
func (this *Pattern) Clear(pos Pos) error {
	if int(pos) >= this.Length {
		return this.indexTooLarge(pos)
	}
	if int(pos) < 0 {
		return this.indexNegative(pos)
	}
	this.Notes[pos] = make([]*lightning.Note, 0)
	return nil
}

// NewPattern creates a Pattern with the specified
// initial size.
func NewPattern(size int) *Pattern {
	return &Pattern{
		size,
		make([][]*lightning.Note, size),
	}
}
