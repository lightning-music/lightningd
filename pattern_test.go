package main

import (
	"encoding/json"
	"github.com/bmizerany/assert"
	"github.com/lightning/go"
	"testing"
)

func TestPatternLength(t *testing.T) {
	pat := NewPattern(0)
	assert.Equal(t, len(pat.Notes), 0)
}

func TestPatternNotesAt(t *testing.T) {
	pat := NewPattern(4)

	err := pat.AddTo(0, lightning.NewNote("audio/file.flac", 60, 120))
	assert.Equal(t, err, nil)

	err = pat.AddTo(0, lightning.NewNote("audio/file.flac", 62, 120))
	assert.Equal(t, err, nil)

	err = pat.AddTo(0, lightning.NewNote("audio/file.flac", 64, 120))
	assert.Equal(t, err, nil)

	notes := pat.NotesAt(0)
	assert.Equal(t, len(notes), 3)
}

func TestPatternAddTo(t *testing.T) {
	// setup pattern
	pat := NewPattern(1)
	err := pat.AddTo(0, lightning.NewNote("audio/file.flac", 72, 96))
	if err != nil {
		t.Fatal(err)
	}
	err = pat.AddTo(0, lightning.NewNote("audio/file.flac", 76, 50))
	if err != nil {
		t.Fatal(err)
	}
	
	// try to add a note at a pos greater than pattern size - 1
	err = pat.AddTo(Pos(pat.Length+1), lightning.NewNote("file.wav", 59, 114))
	if err == nil {
		t.Fatalf("expected err when adding note but got nil")
	}
	assert.Equal(t, err.Error(), "pos (2) greater than pattern length (1)")

	firstNotes := pat.NotesAt(0)
	assert.Equal(t, len(firstNotes), 2)
	assert.Equal(t, firstNotes[0].Number, int32(72))
	assert.Equal(t, firstNotes[0].Velocity, int32(96))
	assert.Equal(t, firstNotes[1].Number, int32(76))
	assert.Equal(t, firstNotes[1].Velocity, int32(50))
}

func TestPatternRemoveFrom(t *testing.T) {
	// setup pattern
	pat := NewPattern(1)
	err := pat.AddTo(0, lightning.NewNote("audio/file.flac", 72, 96))
	if err != nil {
		t.Fatal(err)
	}
	err = pat.AddTo(0, lightning.NewNote("audio/file.flac", 58, 108))
	if err != nil {
		t.Fatal(err)
	}
	err = pat.RemoveFrom(0, int32(72))
	if err != nil {
		t.Fatal(err)
	}
	// look at the notes at pos 0
	firstNotes := pat.NotesAt(0)
	assert.Equal(t, len(firstNotes), 2)
	if firstNotes[0] != nil {
		t.Fatalf("failed to remove note")
	}
	assert.Equal(t, firstNotes[1].Number, int32(58))
	assert.Equal(t, firstNotes[1].Velocity, int32(108))
}

func TestPatternClear(t *testing.T) {
	// setup a pattern
	pat := NewPattern(2)
	err := pat.AddTo(0, lightning.NewNote("foo.ogg", 93, 72))
	if err != nil {
		t.Fatal(err)
	}
	err = pat.AddTo(0, lightning.NewNote("foo.ogg", 90, 83))
	if err != nil {
		t.Fatal(err)
	}
	// verify the notes we just added
	firstNotes := pat.NotesAt(0)
	assert.Equal(t, 2, len(firstNotes))
	assert.Equal(t, firstNotes[0].Sample, "foo.ogg")
	if num := firstNotes[0].Number; num != 93 {
		t.Fatalf("expected note number 93, but got %d", num)
	}
	if num := firstNotes[0].Velocity; num != 72 {
		t.Fatalf("expected note velocity 72, but got %d", num)
	}
	assert.Equal(t, firstNotes[0].Sample, "foo.ogg")
	if num := firstNotes[1].Number; num != 90 {
		t.Fatalf("expected note number 90, but got %d", num)
	}
	if num := firstNotes[1].Velocity; num != 83 {
		t.Fatalf("expected note velocity 83, but got %d", num)
	}
	// clear them
	err = pat.Clear(0)
	if err != nil {
		t.Fatal(err)
	}
	noNotes := pat.NotesAt(0)
	assert.Equal(t, len(noNotes), 0)
}

func TestPatternEncodeJson(t *testing.T) {
	pat := NewPattern(1)
	pat.AddTo(0, lightning.NewNote("audio/file.flac", 56, 101))
	expected := []byte(`{"length":1,"notes":[[{"sample":"audio/file.flac","number":56,"velocity":101}]]}`)
	bs, err := json.Marshal(pat)
	assert.Equal(t, err, nil)
	assert.Equal(t, bs, expected)
}

func TestPatternDecodeJson(t *testing.T) {
	expected := NewPattern(2)
	expected.AddTo(0, lightning.NewNote("audio/file1.flac", 55, 84))
	expected.AddTo(1, lightning.NewNote("audio/file2.flac", 54, 76))
	bs := []byte(`{"length":2,"notes":[[{"sample":"audio/file1.flac","number":55,"velocity":84}],[{"sample":"audio/file2.flac","number":54,"velocity":76}]]}`)
	pat := new(Pattern)
	err := json.Unmarshal(bs, &pat)
	if err != nil {
		t.Fatal(err)
	}
	// verify first notes
	firstNotes := pat.NotesAt(0)
	if len(firstNotes) != 1 {
		t.Fatalf("wrong number of notes (%d)", len(firstNotes))
	}
	fn := firstNotes[0]
	if fn.Sample != "audio/file1.flac" {
		t.Fatalf("wrong sample (%s)", fn.Sample)
	}
	if fn.Number != 55 {
		t.Fatalf("wrong note number (%d)", fn.Number)
	}
	if fn.Velocity != 84 {
		t.Fatalf("wrong note velocity (%d)", fn.Velocity)
	}
	// verify second notes
	secondNotes := pat.NotesAt(1)
	if len(secondNotes) != 1 {
		t.Fatalf("wrong number of notes (%d)", len(secondNotes))
	}
	sn := secondNotes[0]
	if sn.Sample != "audio/file2.flac" {
		t.Fatalf("wrong sample (%s)", sn.Sample)
	}
	if sn.Number != 54 {
		t.Fatalf("wrong note number (%d)", sn.Number)
	}
	if sn.Velocity != 76 {
		t.Fatalf("wrong note velocity (%d)", sn.Velocity)
	}
}
