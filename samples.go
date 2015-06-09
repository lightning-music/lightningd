package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lightning/lightning"
	"golang.org/x/net/websocket"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
)

// samples manages the sample pool
type samples struct {
	// engine plays back samples
	engine lightning.Engine
	// pool is a map from name => path
	pool map[string]string
}

// response is a response object in the websocket API
type response struct {
	success bool   `json:"success"`
	message string `json:"message,omitempty"`
}

// writeJSON writes a response as JSON to an io.Writer
func (self response) writeJSON(w io.Writer) {
	enc := json.NewEncoder(w)
	err := enc.Encode(self)
	if err != nil {
		w.Write([]byte("Could not write response"))
	}
}

// writeJSON writes samples in json format to an io.Writer
func (self *samples) writeJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	samples := make([]string, len(self.pool))
	i := 0
	// key is sample name, value is path to file
	for name, _ := range self.pool {
		samples[i] = name
		i += 1
	}
	return enc.Encode(samples)
}

// list returns an http handler that lists samples
func (self *samples) list() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := self.writeJSON(w)
		if err != nil {
			// assume status code is not already sent
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
	}
}

// play returns an http handler that plays a sample
func (self *samples) play() websocket.Handler {
	return func(conn *websocket.Conn) {
		for {
			note, err := lightning.ReadNote(conn)
			if err == io.EOF {
				continue
			}
			if err != nil {
				response{false, err.Error()}.writeJSON(conn)
				return
			}
			if samplePath, exists := self.pool[note.Sample]; !exists {
				msg := fmt.Sprintf("sample %s does not exist", samplePath)
				response{false, msg}.writeJSON(conn)
				return
			} else {
				note.Sample = samplePath
			}
			err = self.engine.PlayNote(note)
			if err != nil {
				response{false, err.Error()}.writeJSON(conn)
			}
			response{true, "played " + note.Sample}.writeJSON(conn)
		}
	}
}

// readSamples reads samples from a directory
func (self *samples) readSamples(dir string) error {
	fh, eo := os.Open(dir)
	if eo != nil {
		return eo
	}
	// determine if it is a directory
	info, es := fh.Stat()
	if es != nil {
		return es
	}
	if !info.IsDir() {
		return errors.New(dir + " is not a directory")
	}
	fs, er := fh.Readdir(1024)
	if er == io.EOF {
		return errors.New("no samples in " + dir)
	}
	for _, f := range fs {
		if isSupported(f.Name()) {
			name := getName(f.Name())
			self.pool[name] = path.Join(dir, f.Name())
		}
	}
	return nil
}

// newSamples creates a new samples object
func newSamples(engine lightning.Engine) *samples {
	return &samples{engine, make(map[string]string, 0)}
}

// supportedExtensions is a whitelist of file extensions that we support
var supportedExtensions = []string{".wav", ".flac", ".aif", ".aiff"}

// determine if a file has a supported extension
func isSupported(f string) bool {
	for _, ext := range supportedExtensions {
		if strings.HasSuffix(f, ext) {
			return true
		}
	}
	return false
}

// getName gets the sample name from the path to the
// sample file. The sample name is the base name stripped of
// the file extension.
func getName(f string) string {
	base := path.Base(f)
	return base[0:strings.LastIndex(base, ".")]
}
