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
	"encoding/csv"
	"errors"
	"fmt"
	"io"
)

// A csvStorage stores takes into a CSV file.
type csvStorage struct {
	w     *csv.Writer
	werr  error
	cache Storage
}

// NewCSVStorage creates a CSV-file-backed storage from rw.
// rw is first read to get the initial contents, then mutations are
// written afterward.  No attempt is made to deduplicate records.
func NewCSVStorage(rw io.ReadWriter) (Storage, error) {
	cs := &csvStorage{
		w:     csv.NewWriter(rw),
		cache: NewMemoryStorage(),
	}
	if err := readCSV(cs.cache, rw); err != nil {
		return nil, err
	}
	return cs, nil
}

func readCSV(s Storage, r io.Reader) error {
	c := csv.NewReader(r)
	c.FieldsPerRecord = numCSVCols
	for {
		record, err := c.Read()
		if err == io.EOF {
			// TODO(light): partial record available?
			return nil
		} else if err != nil {
			return err
		}
		var row csvRow
		if err := row.read(record); err != nil {
			// TODO(light): line number?
			return err
		}
		if row.IsDelete {
			if err := s.DeleteTake(row.ID); err != nil {
				return err
			}
		} else if err := updateOrCreate(s, &row.Take); err != nil {
			return err
		}
	}
}

func updateOrCreate(s Storage, take *Take) error {
	if err := s.UpdateTake(take.ID, take); err != nil {
		if _, ok := err.(*NotFoundError); !ok {
			return err
		}
	}
	return s.InsertTake(take)
}

func (cs *csvStorage) ListTakes() ([]*Take, error) {
	return cs.cache.ListTakes()
}

func (cs *csvStorage) GetTake(id ID) (*Take, error) {
	return cs.cache.GetTake(id)
}

func (cs *csvStorage) InsertTake(take *Take) error {
	return cs.mutate(func() (row *csvRow, undo func(), err error) {
		err = cs.cache.InsertTake(take)
		if err != nil {
			return
		}
		row = &csvRow{Take: *take}
		undo = func() { cs.cache.DeleteTake(take.ID) }
		return
	})
}

func (cs *csvStorage) UpdateTake(id ID, take *Take) error {
	return cs.mutate(func() (row *csvRow, undo func(), err error) {
		old, err := cs.cache.GetTake(id)
		if err != nil {
			return
		}
		err = cs.cache.UpdateTake(id, take)
		if err != nil {
			return
		}
		row = &csvRow{Take: *take, IsDelete: true}
		undo = func() { cs.cache.UpdateTake(take.ID, old) }
		return
	})
}

func (cs *csvStorage) DeleteTake(id ID) error {
	return cs.mutate(func() (row *csvRow, undo func(), err error) {
		take, err := cs.cache.GetTake(id)
		if err != nil {
			return
		}
		err = cs.cache.DeleteTake(id)
		if err != nil {
			return
		}
		row = &csvRow{Take: Take{ID: id}, IsDelete: true}
		undo = func() { cs.cache.InsertTake(take) }
		return
	})
}

func (cs *csvStorage) mutate(f mutateFunc) error {
	if cs.werr != nil {
		return cs.werr
	}
	row, undo, err := f()
	if err != nil {
		return err
	}
	if err := cs.w.Write(row.record()); err != nil {
		undo()
		cs.werr = err
		return err
	}
	cs.w.Flush()
	if err := cs.w.Error(); err != nil {
		undo()
		cs.werr = err
		return err
	}
	return nil
}

type mutateFunc func() (row *csvRow, undo func(), err error)

func (cs *csvStorage) Close() error {
	if cs.werr != nil {
		return cs.werr
	}
	cs.w.Flush()
	cs.werr = errClosed
	return cs.w.Error()
}

const numCSVCols = 5

type csvRow struct {
	Take
	IsDelete bool
}

func (row *csvRow) read(record []string) error {
	if len(record) != numCSVCols {
		return csvParseError{numColumnsError{len(record), numCSVCols}}
	}
	row.ID.Scene = record[0]
	row.ID.Num = record[1]
	row.ClipName = record[2]
	var err error
	row.Select, err = string2bool(record[3])
	if err != nil {
		return csvParseError{err}
	}
	row.IsDelete, err = string2bool(record[4])
	if err != nil {
		return csvParseError{err}
	}
	return nil
}

func (row *csvRow) record() []string {
	return []string{
		row.ID.Scene,
		row.ID.Num,
		row.ClipName,
		bool2string(row.Select),
		bool2string(row.IsDelete),
	}
}

func bool2string(b bool) string {
	if b {
		return "TRUE"
	} else {
		return "FALSE"
	}
}

func string2bool(s string) (bool, error) {
	switch s {
	case "1", "y", "Y", "true", "TRUE":
		return true, nil
	case "0", "n", "N", "false", "FALSE":
		return false, nil
	default:
		return false, fmt.Errorf("%q is not a boolean value", s)
	}
}

var errClosed = errors.New("takedata: write on closed storage")

type csvParseError struct {
	Err error
}

func (e csvParseError) Error() string {
	return "parsing take row: " + e.Error()
}

type numColumnsError struct {
	Found, Want int
}

func (e numColumnsError) Error() string {
	return fmt.Sprintf("number of columns in record = %d, want %d", e.Found, e.Want)
}
