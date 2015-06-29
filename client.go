package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/websocket"
	"io"
)

type client struct {
	PatternPosition chan uint64
	sequencer       *websocket.Conn
}

func (self *client) play() error {
	buf, err := json.Marshal("start")
	if err != nil {
		return err
	}
	_, err = self.sequencer.Write(buf)
	if err == nil {
		fmt.Println("sent sequencer start message")
	}
	return err
}

func (self *client) stop() error {
	buf, err := json.Marshal("stop")
	if err != nil {
		return err
	}
	_, err = self.sequencer.Write(buf)
	return err
}

func (self *client) receivePosition(origin, host string, port int) {
	var pos uint64
	msg := make([]byte, 8)
	for {
		bytesRead, err := self.sequencer.Read(msg)
		if err == io.EOF {
			continue
		}
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(msg[:bytesRead], &pos)
		if err != nil {
			panic(err)
		}
		fmt.Printf("received pos %v\n", pos)
		self.PatternPosition <- pos
	}
}

func newClient(origin string, port int) (*client, error) {
	var err error
	c := new(client)
	c.PatternPosition = make(chan uint64)
	host := "localhost"
	seqUrl := fmt.Sprintf("ws://%s:%d/sequencer", host, port)
	c.sequencer, err = websocket.Dial(seqUrl, "", origin)
	if err != nil {
		return nil, err
	}
	go c.receivePosition(origin, host, port)
	return c, nil
}
