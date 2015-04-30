package main

import "github.com/lightning/lightning"
import "testing"

func TestSequencer(t *testing.T) {
	engine := lightning.NewEngine()
	seq := NewSequencer(engine, 128, 480)

	err := seq.Start()
	if err != nil {
		t.Fatal(err)
	}

	for pos := range seq.PosChan {
		if pos > 16 {
			break
		}
	}

	err = seq.Stop()
	if err != nil {
		t.Fatal(err)
	}
}
