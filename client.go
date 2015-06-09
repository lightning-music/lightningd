package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/websocket"
	"io"
)

type client struct {
	PatternPosition chan uint64
	patternPlay     *websocket.Conn
	patternStop     *websocket.Conn
}

func (self *client) play() error {
	msg := make([]byte, 4)
	_, err := self.patternPlay.Write(msg)
	return err
}

func (self *client) stop() error {
	msg := make([]byte, 4)
	_, err := self.patternStop.Write(msg)
	return err
}

func (self *client) receivePosition(origin, host string, port int, cher chan error) {
	posUrl := fmt.Sprintf("ws://%s:%d/pattern/position", host, port)
	conn, err := websocket.Dial(posUrl, "", origin)
	if err != nil {
		cher <-err
		return
	}
	cher <-nil
	var pos uint64
	msg := make([]byte, 8)
	for {
		bytesRead, err := conn.Read(msg)
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
		self.PatternPosition <-pos
	}
}

func newClient(origin string, port int) (*client, error) {
	var err error
	cher := make(chan error)
	c := new(client)
	c.PatternPosition = make(chan uint64)
	host := "localhost"
	playUrl := fmt.Sprintf("ws://%s:%d/pattern/play", host, port)
	stopUrl := fmt.Sprintf("ws://%s:%d/pattern/stop", host, port)
	c.patternPlay, err = websocket.Dial(playUrl, "", origin)
	if err != nil {
		return nil, err
	}
	c.patternStop, err = websocket.Dial(stopUrl, "", origin)
	if err != nil {
		return nil, err
	}
	go c.receivePosition(origin, host, port, cher)
	err = <-cher
	if err != nil {
		return nil, err
	}
	return c, nil
}
