package main

import (
	"encoding/json"
	"github.com/lightning/go"
)

// note contains the information to play a single note in a pattern
type note struct {
	Samp string `json:"sample"`
	Num  int32  `json:"number"`
	Vel  int32  `json:"velocity"`
}

func (this note) Sample() string {
	return this.Samp
}

func (this note) Number() int32 {
	return this.Num
}

func (this note) Velocity() int32 {
	return this.Vel
}

// create a new Note instance
func NewNote(sample string, number int32, velocity int32) lightning.Note {
	n := note{sample, number, velocity}
	return n
}

// parse a note from a json object
func ParseNote(ba []byte) (lightning.Note, error) {
	n := new(note)
	ed := json.Unmarshal(ba, n)
	if ed != nil {
		return nil, ed
	}
	return n, nil
}
