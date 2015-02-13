package main

import (
	lightning "github.com/lightning/go"
)

// Sequencer provides a way to play a Pattern using timing
// events emitted from a Metro.
type Sequencer struct {
	PosChan    chan Pos
	PlayErrors chan error
	engine     lightning.Engine
	metro      *Metro
	pattern    *Pattern `json:"pattern"`
}

// NewSequencer creates a Sequencer
func NewSequencer(engine lightning.Engine, patternSize int, tempo Tempo, bardiv string) *Sequencer {
	seq := new(Sequencer)
	seq.PosChan = make(chan Pos, 32)
	seq.PlayErrors = make(chan error)
	seq.engine = engine
	seq.pattern = NewPattern(patternSize)
	seq.metro = NewMetro(tempo, bardiv, func(pos Pos) {
		seq.PosChan <- pos
		err := seq.PlayNotesAt(pos)
		if err != nil {
			// Crash if sample playback errors are not handled!
			select {
			case seq.PlayErrors <- err:
			default:
				panic(err)
			}
		}
	})
	return seq
}

// Play plays all the notes stored at pos
func (this *Sequencer) PlayNotesAt(pos Pos) error {
	var err error
	for _, note := range this.pattern.NotesAt(pos) {
		if note != nil {
			err = this.engine.PlayNote(note)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// NotesAt returns a slice representing the notes
// that are stored at a particular position in the
// Sequencer's Pattern.
func (this *Sequencer) NotesAt(pos Pos) []*lightning.Note {
	return this.pattern.NotesAt(pos)
}

// AddTo adds a note to the Sequencer's pattern at pos.
func (this *Sequencer) AddTo(pos Pos, note *lightning.Note) error {
	return this.pattern.AddTo(pos, note)
}

// AddTo adds a note to the Sequencer's pattern at pos.
func (this *Sequencer) RemoveFrom(pos Pos, note int32) error {
	return this.pattern.RemoveFrom(pos, note)
}

// Clear removes all the notes at a given position
// in the Sequencer's Pattern.
func (this *Sequencer) Clear(pos Pos) error {
	return this.pattern.Clear(pos)
}

// Start plays the Sequencer's Pattern.
func (this *Sequencer) Start() error {
	return this.metro.Start()
}

// Stop playing the Sequencer's Pattern.
func (this *Sequencer) Stop() {
	this.metro.Stop()
}
