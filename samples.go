package main

import (
	"encoding/json"
	"errors"
	"github.com/lightning/lightning"
	"io"
	"net/http"
	"os"
	"strings"
)

// supportedExtensions is a whitelist of file extensions that we support
var supportedExtensions = []string{".wav", ".flac", ".aif", ".aiff"}

// samples manages the sample pool
type samples struct {
	// engine plays back samples
	engine lightning.Engine
	// pool is a map from name => path
	pool map[string]string
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
		}
	}
}

// play returns an http handler that plays a sample
func (self *samples) play() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		note, err := lightning.ReadNote(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if samplePath, exists := self.pool[note.Sample]; !exists {
			w.WriteHeader(http.StatusNotFound)
			return
		} else {
			note.Sample = samplePath
		}
		err = self.engine.PlayNote(note)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

// addDir reads samples from a directory
func (self *samples) addDir(dir string) error {
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
	var samples []string
	for _, f := range fs {
		if isSupported(f, supportedExtensions) {
			samples = append(samples, f.Name())
		}
	}
	return nil
}

// newSamples creates a new samples object
func newSamples(engine lightning.Engine) *samples {
	return &samples{engine, make(map[string]string, 0)}
}

// determine if a file has a supported extension
func isSupported(f os.FileInfo, exts []string) bool {
	is := false
	for _, ext := range exts {
		if strings.HasSuffix(f.Name(), ext) {
			is = true
		}
	}
	return is
}
