package main

import (
	"encoding/json"
	"github.com/lightning/lightning"
	"golang.org/x/net/websocket"
	"io"
	"log"
	"net/http"
)

const (
	// our pattern has 16384 sixteenth notes,
	// which means we have 1024 bars available.
	patternLength = 4096
)

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (self *Response) WriteJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(self)
}

type posMessage struct {
	Position uint64 `json:"position"`
}

func (self posMessage) WriteJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(self)
}

type server struct {
	engine lightning.Engine
	seq    *sequencer
}

func (self *server) Connect(ch1 string, ch2 string) error {
	return self.engine.Connect(ch1, ch2)
}

func (self *server) Listen(addr string) error {
	return http.ListenAndServe(addr, nil)
}

func (self *server) AddTo(pos uint64, note *lightning.Note) error {
	return self.seq.AddTo(pos, note)
}

func (self *server) RemoveFrom(pos uint64, note *lightning.Note) error {
	return self.seq.RemoveFrom(pos, note)
}

// samplePlay exposes a websocket endpoint for playing a sample
func (self *server) samplePlay() websocket.Handler {
	return func(conn *websocket.Conn) {
		for {
			var res Response
			note, re := lightning.ReadNote(conn)
			if re != nil {
				panic(re)
			}
			ep := self.engine.PlayNote(note)
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
func (self *server) patternPlay() websocket.Handler {
	return func(conn *websocket.Conn) {
		msg := make([]byte, 0)
		for {
			_, err := conn.Read(msg)
			if err != nil {
				panic(err)
			}
			log.Println("starting sequencer")
			self.seq.Start()
		}
	}
}

// generate endpoint for stopping pattern
func (self *server) patternStop() websocket.Handler {
	return func(conn *websocket.Conn) {
		msg := make([]byte, 0)
		for {
			_, err := conn.Read(msg)
			if err != nil {
				panic(err)
			}
			log.Println("stopping sequencer")
			self.seq.Stop()
		}
	}
}

// generate endpoint for adding notes to a pattern
func (self *server) noteAdd() websocket.Handler {
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
				err := self.AddTo(pe.Pos, pe.Note)
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
func (self *server) noteRemove() websocket.Handler {
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
				err = self.RemoveFrom(pe.Pos, pe.Note)
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

// patternPosition generate endpoint for sending pattern position
func (self *server) patternPosition() websocket.Handler {
	return func(conn *websocket.Conn) {
		// get messages and call handler
		for pos := range self.seq.PosChan {
			// broadcast position
			err := posMessage{pos}.WriteJSON(conn)
			if err != nil {
				panic(err)
			}
		}
	}
}

// close closes the audio engine
func (self *server) close() {
	self.engine.Close()
}

// newServer creates a websocket/rest server that manages the bulk
// of lightningd functionality
func newServer(www string) (*server, error) {
	srv := new(server)
	srv.engine = lightning.NewEngine()
	// initialize tempo to 120 bpm (a typical
	// starting point for sequencers)
	srv.seq = newSequencer(srv.engine, patternLength, 120)
	// initialize samples
	samples := newSamples(srv.engine)
	// setup handlers under default ServeMux
	fileServer := http.FileServer(http.Dir(www))
	// static file server
	http.Handle("/", fileServer)
	// rest endpoints
	http.HandleFunc("/samples", samples.list())
	// websocket endpoints
	http.Handle("/sample/play", samples.play())
	http.Handle("/note/add", srv.noteAdd())
	http.Handle("/note/remove", srv.noteRemove())
	http.Handle("/pattern/play", srv.patternPlay())
	http.Handle("/pattern/stop", srv.patternStop())
	http.Handle("/pattern/position", srv.patternPosition())
	return srv, nil
}
