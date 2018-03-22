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

// Package video provides operations on video files.
package video

import (
	"io"
	"os"
	"path/filepath"

	"bitbucket.org/zombiezen/cardcpx/multiwriter"
)

// Storage is a write-only blob store.  filepath paths are used.
type Storage interface {
	Store(path string) (io.WriteCloser, error)
}

type fsStorage struct {
	dir string
}

// DirectoryStorage returns storage that writes clips to a flat directory.
func DirectoryStorage(dir string) Storage {
	return &fsStorage{dir}
}

func (fs *fsStorage) Store(path string) (io.WriteCloser, error) {
	fullPath := filepath.Join(fs.dir, path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0777); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE|os.O_EXCL, 0666)
	if f == nil {
		return nil, err
	}
	return syncFile{f}, err
}

// syncFile is a file that calls Sync() on Close().
type syncFile struct {
	*os.File
}

func (f syncFile) Close() error {
	syncErr := f.File.Sync()
	err := f.File.Close()
	if syncErr != nil {
		return syncErr
	} else if err != nil {
		return err
	} else {
		return nil
	}
}

type multiStorage []Storage

// MultiStorage returns storage that writes to the provided replicas.
func MultiStorage(s ...Storage) Storage {
	return multiStorage(s)
}

func (ms multiStorage) Store(path string) (io.WriteCloser, error) {
	w := make([]io.Writer, 0, len(ms))
	c := make([]io.Closer, 0, len(ms))
	for _, s := range ms {
		wc, err := s.Store(path)
		if err != nil {
			for _, cc := range c {
				// TODO(light): log error
				cc.Close()
			}
			return nil, err
		}
		w = append(w, wc)
		c = append(c, wc)
	}
	return multiWriteCloser{multiwriter.New(w...), c}, nil
}

type multiWriteCloser struct {
	io.Writer
	c []io.Closer
}

func (mwc multiWriteCloser) Close() error {
	var firstErr error
	for _, c := range mwc.c {
		err := c.Close()
		if firstErr == nil && err != nil {
			firstErr = err
		}
	}
	return firstErr
}
