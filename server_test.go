package main

import "fmt"
import "testing"
import "time"

func TestServer(t *testing.T) {
	port := 25870
	addr := fmt.Sprintf("localhost:%d", port)
	server, err := newServer(".")
	if err != nil {
		t.Fatal(err)
	}
	go server.listen(addr)
	time.Sleep(50 * time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	origin := "http://localhost/"
	c, err := newClient(origin, port)
	if err != nil {
		t.Fatal(err)
	}
	err = c.play()
	if err != nil {
		t.Fatal(err)
	}
	positions := make([]bool, 16)
	for pos := range c.PatternPosition {
		if pos == uint64(16) {
			break
		}
		positions[pos] = true
	}
	for i, received := range positions {
		if received != true {
			t.Fatalf("did not receive position %d", i)
		}
	}
	err = c.stop()
	if err != nil {
		t.Fatal(err)
	}
	server.close()
}
