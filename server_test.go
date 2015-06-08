package main

// import "golang.org/x/net/websocket"
import "testing"

func TestServer(t *testing.T) {
	server, err := newServer(".")
	if err != nil {
		t.Fatal(err)
	}
	server.close()
}
