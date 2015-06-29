package main

import (
	"encoding/json"
	"fmt"
	"github.com/lightning/lightning"
	"io"
)

// Pattern defines a pattern of notes.
// Notes that you add to a pattern at a position will overwrite
// any notes at that position with the same number.
type Pattern struct {
	Length int                 `json:"length"`
	Notes  [][]*lightning.Note `json:"notes"`
}

func (self *Pattern) indexTooLarge(pos uint64) error {
	str := "pos (%d) greater than pattern length (%d)"
	return fmt.Errorf(str, pos, self.Length)
}

// NotesAt returns a slice representing the notes
// that are stored at a particular position in a pattern.
// pos modulo the size of the pattern is the actual index into
// the pattern.
func (self *Pattern) NotesAt(pos uint64) []*lightning.Note {
	notes := len(self.Notes)
	return self.Notes[int(pos)%notes]
}

// AddTo adds a Note to the pattern at pos
func (self *Pattern) AddTo(pos uint64, note *lightning.Note) error {
	if pos >= uint64(self.Length) {
		return self.indexTooLarge(pos)
	}
	// try to insert to a nil position
	inserted := false
	notes := self.Notes[int(pos)]
	for i, n := range notes {
		if n == nil {
			inserted = true
			notes[i] = note
		}
	}
	// if there were no nil positions, append
	if !inserted {
		self.Notes[int(pos)] = append(notes, note)
	}
	return nil
}

// RemoveFrom removes a note from a particular position in a pattern
func (self *Pattern) RemoveFrom(pos uint64, note *lightning.Note) error {
	if pos >= uint64(self.Length) {
		return self.indexTooLarge(pos)
	}
	// remove a note with the same sample and same number, if one exists
	notes := self.Notes[int(pos)]
	for i, n := range notes {
		if n != nil && n.Number == note.Number && n.Sample == note.Sample {
			notes[i] = nil
		}
	}
	return nil
}

// Clear removes all the notes at a given position in the pattern
func (self *Pattern) Clear(pos uint64) error {
	if pos >= uint64(self.Length) {
		return self.indexTooLarge(pos)
	}
	self.Notes[pos] = make([]*lightning.Note, 0)
	return nil
}

// NewPattern creates a Pattern with the specified size
func NewPattern(size int) *Pattern {
	return &Pattern{
		size,
		make([][]*lightning.Note, size),
	}
}

// Event represents a single edit operation on a pattern
type Event struct {
	Pos  uint64          `json:"pos"`
	Note *lightning.Note `json:"note"`
}

func (self *Event) WriteJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(self)
}

func ReadEvent(r io.Reader) (*Event, error) {
	dec, pe := json.NewDecoder(r), new(Event)
	ed := dec.Decode(pe)
	if ed != nil {
		return nil, ed
	}
	return pe, nil
}

func ReadEvents(r io.Reader) ([]Event, error) {
	dec, pes := json.NewDecoder(r), make([]Event, 0)
	ed := dec.Decode(pes)
	if ed != nil {
		return nil, ed
	}
	return pes, nil
}
