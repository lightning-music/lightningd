package main

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/lightning/go"
	"io"
	"log"
	"net/http"
	"path"
)

const (
	PATTERN_LENGTH = 4096
	PATTERN_DIV    = "1/4"
)

// function that handles websocket messages
type WebsocketHandler func(conn *websocket.Conn, messageType int, msg []byte)

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type Server interface {
	Connect(ch1 string, ch2 string) error
	Listen(addr string) error
}

type posMessage struct {
	Position Pos `json:"position"`
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

func (this *simp) AddTo(pos Pos, note *lightning.Note) error {
	return this.sequencer.AddTo(pos, note)
}

func (this *simp) RemoveFrom(pos Pos, note *lightning.Note) error {
	return this.sequencer.RemoveFrom(pos, note)
}

// generate the MetroFunc that wires the metro to
// the pattern and the audio engine
func genMetroFunc(s *simp) MetroFunc {
	return func(pos Pos) {
		notes := s.sequencer.NotesAt(pos % PATTERN_LENGTH)
		for _, note := range notes {
			s.engine.PlayNote(note)
		}
	}
}

// upgrade repeatedly calls a WebsocketHandler on each new
// incoming message
func (s *simp) upgrade(handler WebsocketHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// upgrade http connection
		var upgrader = websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("could not upgrade http conn to ws: " + err.Error())
			return
		}
		// get messages and call handler
		for {
			msgType, bs, err := conn.ReadMessage()

			if err != nil {
				if err == io.EOF {
					// if err is io.EOF, then it is likely the client
					// has closed the connection, in which case we should
					// close the connection on our end and start listening
					// for a new one.
					break
				} else {
					log.Fatal("could not read ws message: " + err.Error())
				}
			}

			handler(conn, msgType, bs)
		}
	}
}

func (this *simp) samplePlay() http.HandlerFunc {
	return this.upgrade(func(conn *websocket.Conn, msgType int, msg []byte) {
		var res Response
		note, decerr := lightning.DecodeNote(msg)
		if decerr != nil && len(msg) > 0 {
			fmtstr := "could not parse note from %s: %s\n"
			log.Printf(fmtstr, bytes.NewBuffer(msg).String(), decerr.Error())
			return
		}

		// log.Printf("playing %v\n", bytes.NewBuffer(msg).String())
		// note.Sample(), note.Number(), note.Velocity())

		ep := this.engine.PlayNote(note)
		if ep != nil {
			log.Println("could not play note: " + ep.Error())
			return
		}
		res = Response{"ok", "played " + note.Sample}
		resb, em := json.Marshal(res)
		if em != nil {
			log.Println("could not marshal response: " + em.Error())
			return
		}
		ew := conn.WriteMessage(msgType, resb)
		if ew != nil {
			log.Println("could not write ws message: " + ew.Error())
		}
	})
}

// generate endpoint for starting pattern
func (this *simp) patternPlay() http.HandlerFunc {
	return this.upgrade(func(conn *websocket.Conn, msgType int, msg []byte) {
		this.sequencer.Start()
	})
}

// generate endpoint for stopping pattern
func (this *simp) patternStop() http.HandlerFunc {
	return this.upgrade(func(conn *websocket.Conn, msgType int, msg []byte) {
		this.sequencer.Stop()
	})
}

// generate endpoint for adding notes to a pattern
func (this *simp) noteAdd() http.HandlerFunc {
	return this.upgrade(func(conn *websocket.Conn, msgType int, msg []byte) {
		var res Response
		pes := make([]PatternEdit, 0)
		eum := json.Unmarshal(msg, &pes)
		if eum != nil && len(msg) > 0 {
			log.Println("could not unmarshal request body: " + eum.Error())
			log.Printf("request body: %s\n", bytes.NewBuffer(msg).String())
			return
		}
		for _, pe := range pes {
			err := this.AddTo(pe.Pos, pe.Note)
			if err != nil {
				log.Println("could not add note: " + err.Error())
				return
			}
		}
		res = Response{"ok", "note added"}
		resb, ee := json.Marshal(res)
		if ee != nil {
			log.Println("could not encode response: " + ee.Error())
		}
		conn.WriteMessage(msgType, resb)
	})
}

// generate endpoint for removing notes from a pattern
func (this *simp) noteRemove() http.HandlerFunc {
	return this.upgrade(func(conn *websocket.Conn, msgType int, msg []byte) {
		var res Response
		pes := make([]PatternEdit, 0)
		eum := json.Unmarshal(msg, &pes)
		if eum != nil && len(msg) > 0 {
			log.Println("could not unmarshal request body: " + eum.Error())
			log.Printf("request body: %s\n", bytes.NewBuffer(msg).String())
			return
		}
		for _, pe := range pes {
			err := this.RemoveFrom(pe.Pos, pe.Note)
			if err != nil {
				log.Println("could not remove note: " + err.Error())
				return
			}
		}
		res = Response{"ok", "note removed"}
		resb, ee := json.Marshal(res)
		if ee != nil {
			log.Println("could not encode response: " + ee.Error())
		}
		conn.WriteMessage(msgType, resb)
	})
}

// generate endpoint for sending pattern position
func (this *simp) patternPosition() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// upgrade http connection
		var upgrader = websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("could not upgrade http conn to ws: " + err.Error())
			return
		}
		// get messages and call handler
		go func() {
			for {
				log.Println("waiting for PosChan message")
				pos := <-this.sequencer.PosChan
				log.Println("got PosChan message")
				// broadcast position
				conn.WriteJSON(posMessage{pos})
			}
		}()
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
	seq := NewSequencer(engine, PATTERN_LENGTH, Tempo(120), PATTERN_DIV)
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
	// static file server
	http.Handle("/", fh)
	// ReST endpoints
	http.HandleFunc("/samples", api.ListSamples())
	// websocket endpoints
	http.Handle("/sample/play", srv.samplePlay())
	http.HandleFunc("/note/add", srv.noteAdd())
	http.HandleFunc("/note/remove", srv.noteRemove())
	http.HandleFunc("/pattern/play", srv.patternPlay())
	http.HandleFunc("/pattern/stop", srv.patternStop())
	http.HandleFunc("/pattern/position", srv.patternPosition())
	return srv, nil
}
