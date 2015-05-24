package main

import (
	"encoding/json"
	"github.com/lightning/lightning"
	"golang.org/x/net/websocket"
	"io"
	"log"
	"net/http"
	"path"
)

const (
	PATTERN_LENGTH = 4096
	PATTERN_DIV    = "1/4"
)

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (self *Response) WriteJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(self)
}

type Server interface {
	Connect(ch1 string, ch2 string) error
	Listen(addr string) error
}

type posMessage struct {
	Position uint64 `json:"position"`
}

func (self posMessage) WriteJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(self)
}

type simp struct {
	engine    lightning.Engine
	sequencer *Sequencer
}

func (this *simp) Connect(ch1 string, ch2 string) error {
	return this.engine.Connect(ch1, ch2)
}

func (this *simp) Listen(addr string) error {
	return http.ListenAndServe(addr, nil)
}

func (this *simp) AddTo(pos uint64, note *lightning.Note) error {
	return this.sequencer.AddTo(pos, note)
}

func (this *simp) RemoveFrom(pos uint64, note *lightning.Note) error {
	return this.sequencer.RemoveFrom(pos, note)
}

// samplePlay exposes a websocket endpoint for playing a sample
func (this *simp) samplePlay() websocket.Handler {
	return func(conn *websocket.Conn) {
		for {
			var res Response
			note, re := lightning.ReadNote(conn)
			if re != nil {
				panic(re)
			}
			ep := this.engine.PlayNote(note)
			if ep != nil {
				panic(ep)
			}
			res = Response{"ok", "played " + note.Sample}
			ew := res.WriteJSON(conn)
			if ew != nil {
				panic(ew)
			}
		}
	}
}

// patternPlay generates an endpoint for starting pattern
func (this *simp) patternPlay() websocket.Handler {
	return func(conn *websocket.Conn) {
		msg := make([]byte, 0)
		for {
			_, err := conn.Read(msg)
			if err != nil {
				panic(err)
			}
			log.Println("starting sequencer")
			this.sequencer.Start()
		}
	}
}

// generate endpoint for stopping pattern
func (this *simp) patternStop() websocket.Handler {
	return func(conn *websocket.Conn) {
		msg := make([]byte, 0)
		for {
			_, err := conn.Read(msg)
			if err != nil {
				panic(err)
			}
			log.Println("stopping sequencer")
			this.sequencer.Stop()
		}
	}
}

// generate endpoint for adding notes to a pattern
func (this *simp) noteAdd() websocket.Handler {
	return func(conn *websocket.Conn) {
		msg := make([]byte, 0)
		for {
			var res Response
			_, err := conn.Read(msg)
			if err != nil {
				panic(err)
			}
			pes, er := ReadPatternEdits(conn)
			if er != nil {
				panic(er)
			}
			for _, pe := range pes {
				err := this.AddTo(pe.Pos, pe.Note)
				if err != nil {
					log.Println("could not add note: " + err.Error())
					return
				}
			}
			res = Response{"ok", "note added"}
			ew := res.WriteJSON(conn)
			if ew != nil {
				panic(ew)
			}
		}
	}
}

// generate endpoint for removing notes from a pattern
func (this *simp) noteRemove() websocket.Handler {
	return func(conn *websocket.Conn) {
		msg := make([]byte, 0)
		for {
			_, err := conn.Read(msg)
			if err != nil {
				panic(err)
			}
			var res Response
			pes, er := ReadPatternEdits(conn)
			if er != nil {
				panic(er)
			}
			for _, pe := range pes {
				err = this.RemoveFrom(pe.Pos, pe.Note)
				if err != nil {
					log.Println("could not remove note: " + err.Error())
					return
				}
			}
			res = Response{"ok", "note removed"}
			ew := res.WriteJSON(conn)
			if ew != nil {
				panic(ew)
			}
		}
	}
}

// generate endpoint for sending pattern position
func (this *simp) patternPosition() websocket.Handler {
	return func(conn *websocket.Conn) {
		// get messages and call handler
		for pos := range this.sequencer.PosChan {
			log.Printf("sending position %d\n", pos)
			// broadcast position
			err := posMessage{pos}.WriteJSON(conn)
			if err != nil {
				panic(err)
			}
		}
	}
}

func NewServer(webRoot string) (Server, error) {
	// our pattern has 16384 sixteenth notes,
	// which means we have 1024 bars available
	// initialize tempo to 120 bpm (a typical
	// starting point for sequencers)
	engine := lightning.NewEngine()
	// add audio root to dir search list
	audioRoot := path.Join(webRoot, "assets/audio")
	ead := engine.AddDir(audioRoot)
	if ead != nil {
		return nil, ead
	}
	// initialize sequencer
	seq := NewSequencer(engine, PATTERN_LENGTH, 120)
	// initialize server
	srv := &simp{
		engine,
		seq,
	}
	// api handler
	api, ea := NewApi(audioRoot)
	if ea != nil {
		log.Println("could not create api: " + ea.Error())
		return nil, ea
	}
	// setup handlers under default ServeMux
	fh := http.FileServer(http.Dir(webRoot))
	log.Println("setting up api endpoints")
	// static file server
	http.Handle("/", fh)
	// ReST endpoints
	http.HandleFunc("/samples", api.ListSamples())
	// websocket endpoints
	http.Handle("/sample/play", srv.samplePlay())
	http.Handle("/note/add", srv.noteAdd())
	http.Handle("/note/remove", srv.noteRemove())
	http.Handle("/pattern/play", srv.patternPlay())
	http.Handle("/pattern/stop", srv.patternStop())
	http.Handle("/pattern/position", srv.patternPosition())
	log.Println("done setting up api endpoints")
	return srv, nil
}
