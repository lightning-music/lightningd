package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/lightning/go/api"
	"github.com/lightning/go/seq"
	"io/ioutil"
	"log"
	"os"
	"path"
)

func main() {
	home := os.Getenv("HOME")
	defaultRoot := path.Join(home, "www")
	defaultAudio := path.Join(home, "audio")
	bind := flag.String("bind", "localhost:3428", "bind address")
	www := flag.String("www", defaultRoot, "web root")
	audio := flag.String("audio", defaultAudio, "audio sample directory")
	help := flag.Bool("help", false, "print help message")
	flag.Parse()
	if *help {
		printHelp()
		return
	}
	server, err := api.NewServer(*www, *audio)
	if err != nil {
		log.Fatal("could not create server: " + err.Error())
	}
	log.Printf("serving static content from %s\n", *www)
	log.Printf("binding to %s\n", *bind)
	log.Printf("playing audio samples from %s\n", *audio)
	server.Listen(*bind)

	/* setup a pattern from a chunk of json */
	pat := seq.NewPattern(0)
	content, err := ioutil.ReadFile("pat.json")
	if err != nil {
		log.Fatal("could not read pat.json")
	}
	err = json.Unmarshal(content, &pat)
	if err != nil {
		log.Fatal("could not parse pat.json")
	}
}

func printHelp() {
	fmt.Println("Usage:")
	fmt.Println("lightningd [OPTIONS]")
	fmt.Println("")
	fmt.Println("OPTIONS:")
	fmt.Println("  [-audio AUDIO_SAMPLES_DIR] location of audio samples (default=$HOME/audio)")
	fmt.Println("  [-www WEB_ROOT] location of web assets (default=$HOME/www)")
	fmt.Println("  [-help] print a help message")
}
