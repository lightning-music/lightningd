package main

import (
	"github.com/bmizerany/assert"
	"testing"
)

func TestNoteVelocity(t *testing.T) {
	note := NewNote("audio/file.flac", 60, 120)
	assert.Equal(t, note.Velocity(), int32(120))
}

func TestNoteNumber(t *testing.T) {
	note := NewNote("audio/file.flac", 60, 120)
	assert.Equal(t, note.Number(), int32(60))
}

func TestNoteSample(t *testing.T) {
	note := NewNote("audio/file.flac", 60, 120)
	assert.Equal(t, note.Sample(), "audio/file.flac")
}

func TestPatternLength(t *testing.T) {
	pat := NewPattern(0)
	assert.Equal(t, len(pat.Notes), 0)
}

func TestPatternNotesAt(t *testing.T) {
	pat := NewPattern(4)

	err := pat.AddTo(0, NewNote("audio/file.flac", 60, 120))
	assert.Equal(t, err, nil)

	err = pat.AddTo(0, NewNote("audio/file.flac", 62, 120))
	assert.Equal(t, err, nil)

	err = pat.AddTo(0, NewNote("audio/file.flac", 64, 120))
	assert.Equal(t, err, nil)

	notes := pat.NotesAt(0)
	assert.Equal(t, len(notes), 3)
}

func TestPatternAddTo(t *testing.T) {
	pat := NewPattern(1)
	note := NewNote("audio/file.flac", 72, 96)
	pat.AddTo(0, note)
}

// func TestNoteEncodeJson(t *testing.T) {
// 	expected := []byte(`{"sample":"audio/file.flac","number":64,"velocity":108}`)
// 	bs, err := json.Marshal(NewNote("audio/file.flac", 64, 108))
// 	assert.Equal(t, err, nil)
// 	assert.Equal(t, bs, expected)
// }

// func TestNoteDecodeJson(t *testing.T) {
// 	actual := new(Note)
// 	expected := NewNote("audio/file.flac", 58, 109)
// 	blob := []byte(`{"sample":"audio/file.flac","number":58,"velocity":109}`)
// 	err := json.Unmarshal(blob, actual)
// 	assert.Equal(t, err, nil)
// 	assert.Equal(t, &expected, actual)
// }

// func TestPatternEncodeJson(t *testing.T) {
// 	pat := NewPattern(1)
// 	pat.AddTo(0, NewNote("audio/file.flac", 56, 101))
// 	expected := []byte(`{"length":1,"notes":[[{"sample":"audio/file.flac","number":56,"velocity":101}]]}`)
// 	bs, err := json.Marshal(pat)
// 	assert.Equal(t, err, nil)
// 	assert.Equal(t, bs, expected)
// }

// func TestPatternDecodeJson(t *testing.T) {
// 	expected := NewPattern(2)
// 	expected.AddTo(0, NewNote("audio/file.flac", 55, 84))
// 	expected.AddTo(1, NewNote("audio/file2.flac", 54, 76))
// 	bs := []byte(`{"length":2,"notes":[[{"sample":"audio/file.flac","number":55,"velocity":84}],[{"sample":"audio/file2.flac","number":54,"velocity":76}]]}`)
// 	actual := new(Pattern)
// 	err := json.Unmarshal(bs, &actual)
// 	assert.Equal(t, err, nil)
// 	assert.Equal(t, &expected, actual)
// }
