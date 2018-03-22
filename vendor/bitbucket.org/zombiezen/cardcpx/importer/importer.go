/*
	Copyright 2014 Google Inc. All rights reserved.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

		http://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package importer

import (
	"encoding/json"
	"errors"
	"flag"
	"io"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"bitbucket.org/zombiezen/cardcpx/video"
)

var estimatedMiBps = flag.Int64("estImportRate", 60, "estimated import rate (in MiB/s)")

func estimateTime(n, written int64, elapsed time.Duration) time.Duration {
	return time.Duration(float64(n) * float64(elapsed) / float64(written))
}

// ErrBadSubdir is returned from the status when attempting to import into a bad subdirectory.
var ErrBadSubdir = errors.New("cardcpx: importer: invalid subdirectory")

// An Importer copies files from a video source to video storage.
type Importer struct {
	s       video.Storage
	ncopied int64
	elapsed time.Duration

	m       sync.Mutex
	status  Status
	updates chan *Status
}

// New returns an importer that copies to storage s.
func New(s video.Storage) *Importer {
	return &Importer{
		s:       s,
		ncopied: *estimatedMiBps * (2 << 20),
		elapsed: 1 * time.Second,
	}
}

func (imp *Importer) init(clips []*video.Clip) chan<- *Status {
	imp.m.Lock()
	defer imp.m.Unlock()
	imp.updates = make(chan *Status)
	imp.status = Status{
		Active:  true,
		Start:   time.Now(),
		Pending: clips,
		Results: make([]Result, 0, len(clips)),
	}
	for _, clip := range clips {
		imp.status.BytesTotal += clip.TotalSize
	}
	imp.status.ETA = imp.status.Start.Add(estimateTime(imp.status.BytesTotal, imp.ncopied, imp.elapsed))
	return imp.updates
}

// Status returns the current status of the importer.  This method is
// safe to call from multiple goroutines.
func (imp *Importer) Status() *Status {
	imp.m.Lock()
	defer imp.m.Unlock()
	if imp.updates == nil {
		return imp.newStatus()
	}
	return <-imp.updates
}

// Import copies the files specified by clips from src into its storage.
func (imp *Importer) Import(src video.Source, subdir string, clips []*video.Clip) *Status {
	updates := imp.init(clips)
	defer func() {
		imp.m.Lock()
		defer imp.m.Unlock()
		imp.updates = nil
	}()
	for _, clip := range clips {
		result := Result{
			Clip:  clip,
			Start: time.Now(),
		}
		var clipWritten int64
		for _, file := range clip.Paths {
			var fileWritten int64
			fileWritten, result.Error = imp.copy(src, subdir, file, updates)
			clipWritten += fileWritten
			if result.Error != nil {
				imp.status.BytesTotal -= clip.TotalSize - clipWritten
				break
			}
		}
		result.End = time.Now()
		imp.ncopied += clipWritten
		imp.elapsed += result.End.Sub(result.Start)
		imp.status.ETA = result.End.Add(estimateTime(imp.status.BytesTotal-imp.status.BytesCopied, imp.ncopied, imp.elapsed))
		imp.status.Pending = imp.status.Pending[1:]
		imp.status.Results = append(imp.status.Results, result)
		select {
		case updates <- imp.newStatus():
		default:
		}
	}
	imp.status.Active = false
	return imp.newStatus()
}

func (imp *Importer) newStatus() *Status {
	st := new(Status)
	// TODO(light): is this copy sufficient?
	*st = imp.status
	return st
}

func (imp *Importer) copy(src video.Source, subdir, file string, updates chan<- *Status) (int64, error) {
	r, err := src.Open(file)
	if err != nil {
		return 0, err
	}
	defer r.Close()
	outfile := file
	if subdir != "" {
		subdir = filepath.Clean(subdir)
		if filepath.IsAbs(subdir) || subdir == ".." || strings.HasPrefix(subdir, ".."+string(filepath.Separator)) {
			return 0, ErrBadSubdir
		}
		outfile = filepath.Join(subdir, outfile)
	}
	w, err := imp.s.Store(outfile)
	if err != nil {
		return 0, err
	}
	defer func() {
		cerr := w.Close()
		if err != nil {
			err = cerr
		}
	}()

	// Start copying
	progressChan := make(chan copyProgress, 1)
	go copyWithProgress(w, r, progressChan)
	progress := <-progressChan
	imp.status.BytesCopied += progress.written

	st := imp.newStatus()
	for {
		select {
		case p, ok := <-progressChan:
			if !ok {
				return progress.written, progress.err
			}
			imp.status.BytesCopied += p.written - progress.written
			st.BytesCopied = imp.status.BytesCopied
			progress = p
		case imp.updates <- st:
			st = imp.newStatus()
		}
	}
}

// copyWithProgress copies from src to dst and sends updates on ch.
// ch is closed when the copy is finished.  There will be at least one
// value sent on ch.
func copyWithProgress(dst io.Writer, src io.Reader, ch chan<- copyProgress) {
	var written int64
	var err error
	buf := make([]byte, 32*1024)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er == io.EOF {
			break
		}
		if er != nil {
			err = er
			break
		}

		select {
		case ch <- copyProgress{written: written}:
		default:
			// don't hang on sending progress
		}
	}
	ch <- copyProgress{written: written, err: err}
	close(ch)
}

type copyProgress struct {
	written int64
	err     error
}

// Status stores the state of an importer.
type Status struct {
	Active      bool          `json:"active"`
	BytesCopied int64         `json:"bytesCopied"`
	BytesTotal  int64         `json:"bytesTotal"`
	Start       time.Time     `json:"start"`
	ETA         time.Time     `json:"eta"`
	Pending     []*video.Clip `json:"pending"`
	Results     []Result      `json:"results"`
}

// Result stores the final result of a clip's import.
type Result struct {
	Clip  *video.Clip
	Error error
	Start time.Time
	End   time.Time
}

func (r *Result) MarshalJSON() ([]byte, error) {
	var s struct {
		Clip  *video.Clip `json:"clip"`
		Error string      `json:"error,omitempty"`
		Start time.Time   `json:"start"`
		End   time.Time   `json:"end"`
	}
	s.Clip = r.Clip
	if r.Error != nil {
		s.Error = r.Error.Error()
	}
	s.Start = r.Start
	s.End = r.End
	return json.Marshal(&s)
}
