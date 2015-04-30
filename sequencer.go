package main

import (
	"github.com/lightning/lightning"
	"github.com/lightning/metro"
)

// Sequencer provides a way to play a Pattern using timing
// events emitted from a Metro.
type Sequencer struct {
	PosChan    chan uint64
	PlayErrors chan error
	engine     lightning.Engine
	metro      metro.Metro
	pattern    *Pattern `json:"pattern"`
}

// NewSequencer creates a Sequencer
func NewSequencer(engine lightning.Engine, patternSize int, tempo float32) *Sequencer {
	seq := new(Sequencer)
	seq.PosChan = make(chan uint64)
	seq.PlayErrors = make(chan error)
	seq.engine = engine
	seq.pattern = NewPattern(patternSize)
	seq.metro = metro.New(tempo)

	go func() {
		for tick := range seq.metro.Ticks() {
			pos := uint64(tick)
			seq.PosChan <- pos
			err := seq.PlayNotesAt(tick)
			if err != nil {
				// Crash if sample playback errors are not handled!
				select {
				case seq.PlayErrors <- err:
				default:
					panic(err)
				}
			}
		}
	}()

	return seq
}

// Play plays all the notes stored at pos
func (this *Sequencer) PlayNotesAt(pos uint64) error {
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
func (this *Sequencer) NotesAt(pos uint64) []*lightning.Note {
	return this.pattern.NotesAt(pos)
}

// AddTo adds a note to the Sequencer's pattern at pos.
func (this *Sequencer) AddTo(pos uint64, note *lightning.Note) error {
	return this.pattern.AddTo(pos, note)
}

// AddTo adds a note to the Sequencer's pattern at pos.
func (this *Sequencer) RemoveFrom(pos uint64, note *lightning.Note) error {
	return this.pattern.RemoveFrom(pos, note)
}

// Clear removes all the notes at a given position
// in the Sequencer's Pattern.
func (this *Sequencer) Clear(pos uint64) error {
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
