package main

import (
	"encoding/json"
	"flag"
	"github.com/lightning/go"
	"io/ioutil"
	"log"
	"os"
	"path"
)

func main() {
	home := os.Getenv("HOME")
	defaultRoot := path.Join(home, "www")
	defaultAudio := path.Join(home, "audio")
	defaultCh1 := "system:playback_1"
	defaultCh2 := "system:playback_2"
	bind := flag.String("bind", "localhost:3428", "bind address")
	www := flag.String("www", defaultRoot, "web root")
	audio := flag.String("audio", defaultAudio, "audio sample directory")
	ch1 := flag.String("ch1", defaultCh1, "left channel JACK sink")
	ch2 := flag.String("ch2", defaultCh2, "right channel JACK sink")
	// parse cli flags
	flag.Parse()
	server, err := lightning.NewServer(*www, *audio)
	if err != nil {
		log.Fatal("could not create server: " + err.Error())
	}
	log.Printf("serving static content from %s\n", *www)
	log.Printf("binding to %s\n", *bind)
	log.Printf("playing audio samples from %s\n", *audio)
	log.Printf("connecting output1 to %s and output2 to %s\n", *ch1, *ch2)
	server.Connect(*ch1, *ch2)
	server.Listen(*bind)

	/* setup a pattern from a chunk of json */
	pat := lightning.NewPattern(0)
	content, err := ioutil.ReadFile("pat.json")
	if err != nil {
		log.Fatal("could not read pat.json")
	}
	err = json.Unmarshal(content, &pat)
	if err != nil {
		log.Fatal("could not parse pat.json")
	}
}
