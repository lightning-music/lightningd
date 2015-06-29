package main

import (
	"encoding/json"
	"fmt"
	"github.com/lightning/lightning"
	"golang.org/x/net/websocket"
	"io"
	"net/http"
	"strconv"
)

const (
	// our pattern has 16384 sixteenth notes,
	// which means we have 1024 bars available.
	patternLength  = 4096
	sequencerStop  = 0
	sequencerStart = 1
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

// readMessages reads messages for the websocket endpoint
// and sends them on a channel. errors are sent on the provided error
// channel. if an error occurs, the method returns
func (self *server) readMessages(conn *websocket.Conn, c chan interface{}, e chan error) {
	var msg interface{}
	dec := json.NewDecoder(conn)
	for err := dec.Decode(&msg); true; err = dec.Decode(&msg) {
		if err != nil {
			e <-err
			break
		}
		c <-msg
	}
}

// sequencerEndpoint creates a websocket handler for the /sequencer endpoint
func (self *server) sequencerEndpoint(conn *websocket.Conn) {
	var err error
	mc := make(chan interface{})
	ec := make(chan error)
	go self.readMessages(conn, mc, ec)
	for {
		select {
		case err := <-ec:
			if err == io.EOF {
				// the client closed the connection
				goto CloseConnection
			}
			if err != nil {
				panic(err)
			}
		case msg := <-mc:
			if s, isString := msg.(string); isString {
				// start or stop
				if s == "start" {
					err = self.seq.Start()
					if err != nil {
						panic(err)
					}
				} else if s == "stop" {
					err = self.seq.Stop()
					if err != nil {
						panic(err)
					}
				} else {
					panic(fmt.Errorf("unrecognized sequencer command %s", s))
				}
			} else if f, isFloat := msg.(float64); isFloat {
				// tempo
				self.seq.SetTempo(float32(f))
			}
		case pos := <-self.seq.PosChan:
			_, err = conn.Write([]byte(strconv.FormatUint(pos, 10)))
			if err == io.EOF {
				goto CloseConnection
			}
			if err != nil {
				panic(err)
			}
		}
	}
CloseConnection:
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
	// http endpoints
	http.HandleFunc("/samples", srv.samples.list())
	// websocket endpoints
	http.Handle("/sample/play", srv.samples.play())
	http.Handle("/sequencer", websocket.Handler(srv.sequencerEndpoint))
	// http.Handle("/note/remove", srv.noteRemove())
	// http.Handle("/pattern/play", srv.patternPlay())
	// http.Handle("/pattern/stop", srv.patternStop())
	// http.Handle("/pattern/position", srv.patternPosition())
	/* /sequencer */
	return srv, nil
}
