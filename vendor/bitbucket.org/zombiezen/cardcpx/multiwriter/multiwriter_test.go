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

package multiwriter

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"
	"time"
)

const testString = "Hello, World!"

func TestWrite(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	w := New(&buf1, &buf2)
	n, err := w.Write([]byte(testString))
	if err != nil {
		t.Error("error:", err)
	}
	if n != len(testString) {
		t.Errorf("n = %d, want %d", n, len(testString))
	}
	if s := buf1.String(); s != testString {
		t.Errorf("buf1.String() = %q, want %q", s, testString)
	}
	if s := buf2.String(); s != testString {
		t.Errorf("buf2.String() = %q, want %q", s, testString)
	}
}

func BenchmarkSingleWrite(b *testing.B) {
	bench(b, benchWriter())
}

func BenchmarkStdlibWrite(b *testing.B) {
	bench(b, io.MultiWriter(benchWriter(), benchWriter()))
}

func BenchmarkWrite(b *testing.B) {
	bench(b, New(benchWriter(), benchWriter()))
}

const benchChunkSize = 32 * 1024

func bench(b *testing.B, w io.Writer) {
	buf := make([]byte, benchChunkSize)
	b.SetBytes(benchChunkSize)
	for i := 0; i < b.N; i++ {
		w.Write(buf)
	}
}

func benchWriter() io.Writer {
	return delayWriter{ioutil.Discard}
}

type delayWriter struct {
	w io.Writer
}

// diskMBps is the expected rate of disk writes in mebibytes per second.
const diskMBps = 128

func (dw delayWriter) Write(p []byte) (int, error) {
	time.Sleep(time.Duration(len(p)) * time.Second / (2 << 20 * diskMBps))
	return dw.w.Write(p)
}
