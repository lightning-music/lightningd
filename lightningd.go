package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
)

const (
	// DefaultAddr is the default address lightning will listen at
	DefaultAddr = "localhost:3428"
	// DefaultWWW is the default location of web assets for the lightning web ui
	// see github.com/lightning/lightning/{linux,darwin}.mk
	// for default www directories
	DefaultWWW = "/usr/local/share/lightning/www"
	// Default names of JACK system outputs
	DefaultCh1 = "system:playback_1"
	DefaultCh2 = "system:playback_2"
)

func main() {
	bind := flag.String("bind", DefaultAddr, "bind address")
	www := flag.String("www", DefaultWWW, "web root")
	ch1 := flag.String("ch1", DefaultCh1, "left channel JACK sink")
	ch2 := flag.String("ch2", DefaultCh2, "right channel JACK sink")
	// parse cli flags
	flag.Parse()
	server, err := NewServer(*www)
	if err != nil {
		log.Fatal("could not create server: " + err.Error())
	}
	log.Printf("serving static content from %s\n", *www)
	log.Printf("binding to %s\n", *bind)
	log.Printf("connecting audio output1 to %s and output2 to %s\n", *ch1, *ch2)
	server.Connect(*ch1, *ch2)
	server.Listen(*bind)

	/* setup a pattern from a chunk of json */
	pat := NewPattern(0)
	content, err := ioutil.ReadFile("pat.json")
	if err != nil {
		log.Fatal("could not read pat.json")
	}
	err = json.Unmarshal(content, &pat)
	if err != nil {
		log.Fatal("could not parse pat.json")
	}
}
