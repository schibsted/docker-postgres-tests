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

// Package multiwriter provides an io.Writer that duplicates its writes to multiple writers concurrently.
package multiwriter

import (
	"io"
)

// multiWriter duplicates writes to multiple writers.
type multiWriter []io.Writer

// New returns an io.Writer that duplicates writes to all provided writers.
func New(w ...io.Writer) io.Writer {
	return multiWriter(w)
}

// Write writes p to all writers concurrently.  If any errors occur, the shortest write is returned.
func (mw multiWriter) Write(p []byte) (int, error) {
	done := make(chan result, len(mw))
	for _, w := range mw {
		go send(w, p, done)
	}
	endResult := result{n: len(p)}
	for _ = range mw {
		res := <-done
		if res.err != nil && (endResult.err == nil || res.n < endResult.n) {
			endResult = res
		}
	}
	return endResult.n, endResult.err
}

func send(w io.Writer, p []byte, done chan<- result) {
	var res result
	res.n, res.err = w.Write(p)
	if res.n < len(p) && res.err == nil {
		res.err = io.ErrShortWrite
	}
	done <- res
}

type result struct {
	n   int
	err error
}
