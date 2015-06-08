package main

import (
	"github.com/lightning/lightning"
	"github.com/lightning/metro"
)

// sequencer provides a way to play a Pattern using timing
// events emitted from a Metro
type sequencer struct {
	PosChan    chan uint64
	PlayErrors chan error
	engine     lightning.Engine
	metro      metro.Metro
	pattern    *Pattern `json:"pattern"`
}

// newSequencer creates a Sequencer
func newSequencer(engine lightning.Engine, patternSize int, tempo float32) *sequencer {
	seq := new(sequencer)
	seq.PosChan = make(chan uint64)
	seq.PlayErrors = make(chan error)
	seq.engine = engine
	seq.pattern = NewPattern(patternSize)
	seq.metro = metro.New(tempo)

	go func() {
		for tick := range seq.metro.Ticks() {
			pos := tick % uint64(seq.pattern.Length)
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
func (this *sequencer) PlayNotesAt(pos uint64) error {
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
// sequencer's Pattern.
func (this *sequencer) NotesAt(pos uint64) []*lightning.Note {
	return this.pattern.NotesAt(pos)
}

// AddTo adds a note to the sequencer's pattern at pos.
func (this *sequencer) AddTo(pos uint64, note *lightning.Note) error {
	return this.pattern.AddTo(pos, note)
}

// AddTo adds a note to the sequencer's pattern at pos.
func (this *sequencer) RemoveFrom(pos uint64, note *lightning.Note) error {
	return this.pattern.RemoveFrom(pos, note)
}

// Clear removes all the notes at a given position
// in the sequencer's Pattern.
func (this *sequencer) Clear(pos uint64) error {
	return this.pattern.Clear(pos)
}

// Start plays the sequencer's Pattern.
func (this *sequencer) Start() error {
	return this.metro.Start()
}

// Stop playing the sequencer's Pattern.
func (this *sequencer) Stop() error {
	return this.metro.Stop()
}
