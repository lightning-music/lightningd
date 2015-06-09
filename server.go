package main

import (
	"encoding/json"
	"fmt"
	"github.com/lightning/lightning"
	"golang.org/x/net/websocket"
	"io"
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

func (self *Response) writeJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(self)
}

type server struct {
	engine  lightning.Engine
	seq     *sequencer
	samples *samples
}

func (self *server) connect(ch1 string, ch2 string) error {
	return self.engine.Connect(ch1, ch2)
}

func (self *server) listen(addr string) error {
	return http.ListenAndServe(addr, nil)
}

func (self *server) readSamples(dir string) error {
	return self.samples.readSamples(dir)
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
			ew := res.writeJSON(conn)
			if ew != nil {
				panic(ew)
			}
		}
	}
}

// patternPlay generates an endpoint for starting pattern
func (self *server) patternPlay() websocket.Handler {
	return func(conn *websocket.Conn) {
		msg := make([]byte, 4)
		for {
			_, err := conn.Read(msg)
			if err == io.EOF {
				continue
			}
			if err != nil {
				panic(err)
			}
			self.seq.Start()
		}
	}
}

// generate endpoint for stopping pattern
func (self *server) patternStop() websocket.Handler {
	return func(conn *websocket.Conn) {
		msg := make([]byte, 4)
		for {
			_, err := conn.Read(msg)
			if err == io.EOF {
				continue
			}
			if err != nil {
				panic(err)
			}
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
			if err == io.EOF {
				continue
			}
			if err != nil {
				panic(err)
			}
			pes, er := ReadPatternEdits(conn)
			if er != nil {
				panic(er)
			}
			for _, pe := range pes {
				err := self.seq.AddTo(pe.Pos, pe.Note)
				if err != nil {
					panic(err)
				}
			}
			res = Response{"ok", "note added"}
			ew := res.writeJSON(conn)
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
			if err == io.EOF {
				continue
			}
			if err != nil {
				panic(err)
			}
			var res Response
			pes, er := ReadPatternEdits(conn)
			if er != nil {
				panic(er)
			}
			for _, pe := range pes {
				err = self.seq.RemoveFrom(pe.Pos, pe.Note)
				if err != nil {
					panic(err)
				}
			}
			res = Response{"ok", "note removed"}
			ew := res.writeJSON(conn)
			if ew != nil {
				panic(ew)
			}
		}
	}
}

// patternPosition generates a websocket endpoint for sending
// pattern position
func (self *server) patternPosition() websocket.Handler {
	return func(conn *websocket.Conn) {
		for pos := range self.seq.PosChan {
			// broadcast position
			msg, err := json.Marshal(pos)
			if err != nil {
				panic(err)
			}
			bytesWritten, err := conn.Write(msg)
			if err != nil {
				panic(err)
			}
			if bytesWritten != len(msg) {
				panic(fmt.Errorf("wrote %d out of %d bytes", bytesWritten, len(msg)))
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
	srv.samples = newSamples(srv.engine)
	// setup handlers under default ServeMux
	fileServer := http.FileServer(http.Dir(www))
	// static file server
	http.Handle("/", fileServer)
	// rest endpoints
	http.HandleFunc("/samples", srv.samples.list())
	// websocket endpoints
	http.Handle("/sample/play", srv.samples.play())
	http.Handle("/note/add", srv.noteAdd())
	http.Handle("/note/remove", srv.noteRemove())
	http.Handle("/pattern/play", srv.patternPlay())
	http.Handle("/pattern/stop", srv.patternStop())
	http.Handle("/pattern/position", srv.patternPosition())
	return srv, nil
}
