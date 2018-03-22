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

package takedata

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func TestnewEmptyCSVStorage(t *testing.T) {
	newStorageTest(t, newEmptyCSVStorage())
}

func TestCSVStorageInsertion(t *testing.T) {
	storageInsertionTest(t, newEmptyCSVStorage())
}

func TestCSVStorageUpdateInPlace(t *testing.T) {
	storageUpdateInPlaceTest(t, newEmptyCSVStorage())
}

func TestCSVStorageUpdateToExisting(t *testing.T) {
	storageUpdateToExistingTest(t, newEmptyCSVStorage())
}

func TestCSVStorageUpdateMove(t *testing.T) {
	storageUpdateMoveTest(t, newEmptyCSVStorage)
}

func TestCSVStorageDelete(t *testing.T) {
	storageDeleteTest(t, newEmptyCSVStorage())
}

func newEmptyCSVStorage() Storage {
	s, err := NewCSVStorage(new(closeBuffer))
	if err != nil {
		panic(err)
	}
	return s
}

const csvInitData = "1,2,foo,false,false\r\n"

func TestReadCSVStorage(t *testing.T) {
	buf := newCloseBufferString(csvInitData)
	s, err := NewCSVStorage(buf)
	if err != nil {
		t.Fatal("NewCSVStorage error:", err)
	}
	defer closeStorage(t, s)
	checkListTakes(t, s,
		&Take{ID: ID{Scene: "1", Num: "2"}, ClipName: "foo"},
	)
	checkGetTake(t, s, &Take{
		ID:       ID{Scene: "1", Num: "2"},
		ClipName: "foo",
	})
	bs := buf.String()
	if bs != "" {
		t.Errorf("CSV data = %q, want \"\"", bs)
	}
}

func TestWriteInsertCSVStorage(t *testing.T) {
	buf := newCloseBufferString(csvInitData)
	s, err := NewCSVStorage(buf)
	if err != nil {
		t.Fatal("NewCSVStorage error:", err)
	}
	defer closeStorage(t, s)
	take := &Take{ID{"1", "3"}, "bar", true}
	insertTakes(t, s, take)
	checkGetTake(t, s, take)
	bs := buf.String()
	const want = "1,3,bar,TRUE,FALSE\n"
	if bs != want {
		t.Errorf("CSV data = %q, want %q", bs, want)
	}
}

func TestWriteDeleteCSVStorage(t *testing.T) {
	buf := newCloseBufferString(csvInitData)
	s, err := NewCSVStorage(buf)
	if err != nil {
		t.Fatal("NewCSVStorage error:", err)
	}
	defer closeStorage(t, s)
	err = s.DeleteTake(ID{"1", "2"})
	if err != nil {
		t.Error("DeleteTake error:", err)
	}
	checkDeleted(t, s, ID{"1", "2"})
	bs := buf.String()

	// Go 1.4 changed empty cell quoting rules, so accept either.
	const (
		want12 = "1,2,\"\",FALSE,TRUE\n"
		want14 = "1,2,,FALSE,TRUE\n"
	)
	if !(bs == want12 || bs == want14) {
		t.Errorf("CSV data = %q, want %q or %q", bs, want12, want14)
	}
}

func TestCSVStorageWithFiles(t *testing.T) {
	f, err := ioutil.TempFile("", "takedata-test")
	if err != nil {
		t.Fatal("Temp file can't be created:", err)
	}
	name := f.Name()
	defer f.Close()
	defer os.Remove(name)
	if _, err := f.WriteString(csvInitData); err != nil {
		t.Fatal("Can't write temp file:", err)
	}
	if _, err := f.Seek(0, os.SEEK_SET); err != nil {
		t.Fatal("Can't seek temp file:", err)
	}

	s, err := NewCSVStorage(f)
	if err != nil {
		t.Error("NewCSVStorage error:", err)
	}
	checkGetTake(t, s, &Take{
		ID:       ID{Scene: "1", Num: "2"},
		ClipName: "foo",
	})
	insertTakes(t, s, &Take{ID{"1", "3"}, "bar", true})
	if err := s.Close(); err != nil {
		t.Error("Close() error:", err)
	}

	if _, err := f.Seek(0, os.SEEK_SET); err != nil {
		t.Fatal("Can't seek temp file:", err)
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		t.Error("Temp file read error:", err)
	}
	const wantData = csvInitData + "1,3,bar,TRUE,FALSE\n"
	if string(data) != wantData {
		t.Errorf("File contents = %q, want %q", data, wantData)
	}
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
