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
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"bitbucket.org/zombiezen/cardcpx/video"
)

// TODO(light): this doesn't test large files

func TestImport(t *testing.T) {
	src := &fakeSource{
		clips: []*video.Clip{
			{
				Name: "foo",
				Paths: []string{
					"foo/1",
					"foo/2",
				},
				TotalSize: int64(len(foo1Data) + len(foo2Data)),
			},
			{
				Name: "bar",
				Paths: []string{
					"bar/1",
					"bar/2",
				},
				TotalSize: int64(len(bar1Data) + len(bar2Data)),
			},
		},
		files: map[string]string{
			"foo/1": foo1Data,
			"foo/2": foo2Data,
			"bar/1": bar1Data,
			"bar/2": bar2Data,
		},
	}
	s := newFakeStorage()
	imp := New(s)

	st := imp.Import(src, "", src.clips)

	want := &Status{
		Active:      false,
		BytesCopied: int64(totalDataSize),
		BytesTotal:  int64(totalDataSize),
		Pending:     []*video.Clip{},
		Results: []Result{
			{
				Clip:  src.clips[0],
				Error: nil,
			},
			{
				Clip:  src.clips[1],
				Error: nil,
			},
		},
	}
	if !statusEq(st, want) {
		t.Errorf("Import(...) =\n%+v, want\n%+v", st, want)
	}
	st = imp.Status()
	if !statusEq(st, want) {
		t.Errorf("Status() =\n%+v, want\n%+v", st, want)
	}

	for path, want := range src.files {
		buf := s.files[path]
		if buf == nil {
			t.Errorf("did not write to %s", path)
		} else if s := buf.String(); s != want {
			t.Errorf("stored %q to %s, want %q", s, path, want)
		}
	}
}

func TestImport_Subdir(t *testing.T) {
	src := &fakeSource{
		clips: []*video.Clip{
			{
				Name: "foo",
				Paths: []string{
					"foo/1",
					"foo/2",
				},
				TotalSize: int64(len(foo1Data) + len(foo2Data)),
			},
			{
				Name: "bar",
				Paths: []string{
					"bar/1",
					"bar/2",
				},
				TotalSize: int64(len(bar1Data) + len(bar2Data)),
			},
		},
		files: map[string]string{
			"foo/1": foo1Data,
			"foo/2": foo2Data,
			"bar/1": bar1Data,
			"bar/2": bar2Data,
		},
	}
	s := newFakeStorage()
	imp := New(s)

	st := imp.Import(src, "SUB", src.clips)

	want := &Status{
		Active:      false,
		BytesCopied: int64(totalDataSize),
		BytesTotal:  int64(totalDataSize),
		Pending:     []*video.Clip{},
		Results: []Result{
			{
				Clip:  src.clips[0],
				Error: nil,
			},
			{
				Clip:  src.clips[1],
				Error: nil,
			},
		},
	}
	if !statusEq(st, want) {
		t.Errorf("Import(...) =\n%+v, want\n%+v", st, want)
	}

	for path, want := range src.files {
		outpath := filepath.Join("SUB", path)
		buf := s.files[outpath]
		if buf == nil {
			t.Errorf("did not write to %s", outpath)
		} else if s := buf.String(); s != want {
			t.Errorf("stored %q to %s, want %q", s, outpath, want)
		}
	}
}

func TestImport_SubdirParentsFail(t *testing.T) {
	src := &fakeSource{
		clips: []*video.Clip{
			{
				Name: "foo",
				Paths: []string{
					"foo/1",
				},
				TotalSize: int64(len(foo1Data)),
			},
		},
		files: map[string]string{
			"foo/1": foo1Data,
		},
	}
	want := &Status{
		Active:      false,
		BytesCopied: int64(0),
		BytesTotal:  int64(0),
		Pending:     []*video.Clip{},
		Results: []Result{
			{
				Clip:  src.clips[0],
				Error: ErrBadSubdir,
			},
		},
	}
	checks := []string{
		"..",
		"../..",
		"../foo",
		"../foo/..",
		"/foo/bar",
		"foo/../../..",
		"foo/../../../etc/passwd",
	}

	for _, subdir := range checks {
		st := New(newFakeStorage()).Import(src, subdir, src.clips)

		if !statusEq(st, want) {
			t.Errorf("New(...).Import(src, %q, src.clips) =\n%+v, want\n%+v", subdir, st, want)
		}
	}
}

func statusEq(s, t *Status) bool {
	if s.Active != t.Active {
		return false
	}
	if s.BytesCopied != t.BytesCopied {
		return false
	}
	if s.BytesTotal != t.BytesTotal {
		return false
	}
	if len(s.Pending) != len(t.Pending) {
		return false
	}
	for i := range s.Pending {
		if s.Pending[i] != t.Pending[i] {
			return false
		}
	}
	if len(s.Results) != len(t.Results) {
		return false
	}
	for i := range s.Results {
		if s.Results[i].Clip != t.Results[i].Clip {
			return false
		}
		if s.Results[i].Error != t.Results[i].Error {
			return false
		}
	}
	return true
}

func TestEstimateTime(t *testing.T) {
	tests := []struct {
		n, written int64
		elapsed    time.Duration
		want       time.Duration
	}{
		{0, 60, 1 * time.Second, 0},
		{30, 60, 1 * time.Second, 500 * time.Millisecond},
		{60, 60, 1 * time.Second, 1 * time.Second},
		{120, 60, 1 * time.Second, 2 * time.Second},
		{107600000000, 100000000, 1 * time.Second, 1076 * time.Second},
	}
	for _, test := range tests {
		d := estimateTime(test.n, test.written, test.elapsed)
		if d != test.want {
			t.Errorf("estimateTime(%d, %d, %v) = %v, want %v", test.n, test.written, test.elapsed, d, test.want)
		}
	}
}

const (
	foo1Data = "Hello, World!"
	foo2Data = "This is foo 2"
	bar1Data = "Yankee Hotel Foxtrot"
	bar2Data = "Whiskey Tango Foxtrot"
)

const totalDataSize = len(foo1Data) + len(foo2Data) + len(bar1Data) + len(bar2Data)

type fakeSource struct {
	clips []*video.Clip
	files map[string]string
}

func (src *fakeSource) List() ([]*video.Clip, error) {
	return src.clips, nil
}

func (src *fakeSource) Open(path string) (io.ReadCloser, error) {
	data, ok := src.files[path]
	if !ok {
		return nil, &os.PathError{Op: "open", Path: path, Err: os.ErrNotExist}
	}
	return newCloseBufferString(data), nil
}

type fakeStorage struct {
	files map[string]*closeBuffer
}

func newFakeStorage() *fakeStorage {
	return &fakeStorage{make(map[string]*closeBuffer)}
}

func (s *fakeStorage) Store(path string) (io.WriteCloser, error) {
	buf := new(closeBuffer)
	s.files[path] = buf
	return buf, nil
}

type closeBuffer struct {
	bytes.Buffer
}

func newCloseBufferString(s string) *closeBuffer {
	buf := new(closeBuffer)
	buf.WriteString(s)
	return buf
}

func (cb *closeBuffer) Close() error {
	return nil
}
