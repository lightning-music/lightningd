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
	pat := NewPattern(1)
	note := lightning.NewNote("audio/file.flac", 72, 96)
	pat.AddTo(0, note)
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
