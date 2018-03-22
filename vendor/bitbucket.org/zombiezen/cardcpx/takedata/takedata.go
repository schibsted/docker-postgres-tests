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
	"sort"

	"bitbucket.org/zombiezen/cardcpx/natsort"
)

// ID is a unique identifier for a take.
type ID struct {
	Scene string `json:"scene"`
	Num   string `json:"num"`
}

func (id ID) String() string {
	return id.Scene + "-" + id.Num
}

// Less reports whether id1 is less than id2.
func (id1 ID) Less(id2 ID) bool {
	if id1.Scene == id2.Scene {
		return natsort.Less(id1.Num, id2.Num)
	} else {
		return natsort.Less(id1.Scene, id2.Scene)
	}
}

// Take represents a single video clip.
type Take struct {
	ID       ID     `json:"id"`
	ClipName string `json:"clipName,omitempty"`
	Select   bool   `json:"select"`
}

// Storage provides a database of takes.
// Calls to ListTakes and GetTake from multiple goroutines are safe,
// as long as they do not happen concurrently any other calls.
type Storage interface {
	// ListTakes returns a sorted list of all takes in the storage.
	// The returned takes should not be modified.
	ListTakes() ([]*Take, error)

	// GetTake retrieves a take with the given ID.
	// The returned takes should not be modified.
	GetTake(id ID) (*Take, error)

	// InsertTake adds a new take.
	InsertTake(take *Take) error

	// UpdateTake updates an existing take.
	UpdateTake(id ID, take *Take) error

	// DeleteTake deletes a single take.
	DeleteTake(id ID) error

	// Close flushes the storage to disk.
	Close() error
}

type memStorage struct {
	list []*Take
}

// NewMemoryStorage returns a new in-memory storage.
func NewMemoryStorage() Storage {
	const initCap = 200
	return &memStorage{
		list: make([]*Take, 0, initCap),
	}
}

func (ms *memStorage) ListTakes() ([]*Take, error) {
	takes := make([]*Take, len(ms.list))
	copy(takes, ms.list)
	return takes, nil
}

func (ms *memStorage) GetTake(id ID) (*Take, error) {
	i, ok := ms.search(id)
	if !ok {
		return nil, &NotFoundError{ID: id}
	}
	return ms.list[i], nil
}

func (ms *memStorage) InsertTake(take *Take) error {
	i, ok := ms.search(take.ID)
	if ok {
		return &ExistsError{ID: take.ID}
	}
	ms.list = append(ms.list, nil)
	copy(ms.list[i+1:], ms.list[i:])
	ms.store(i, take)
	return nil
}

func (ms *memStorage) UpdateTake(id ID, take *Take) error {
	i, ok := ms.search(id)
	if !ok {
		return &NotFoundError{ID: id}
	}
	if id == take.ID {
		ms.store(i, take)
		return nil
	}
	j, ok := ms.search(take.ID)
	switch {
	case ok:
		return &ExistsError{ID: take.ID}
	case i < j:
		copy(ms.list[i:j], ms.list[i+1:j])
		ms.store(j-1, take)
	default:
		copy(ms.list[j+1:i+1], ms.list[j:i+1])
		ms.store(j, take)
	}
	return nil
}

func (ms *memStorage) DeleteTake(id ID) error {
	i, ok := ms.search(id)
	if !ok {
		return &NotFoundError{ID: id}
	}
	n := len(ms.list)
	copy(ms.list[i:], ms.list[i+1:])
	ms.list[n-1] = nil
	ms.list = ms.list[:n-1]
	return nil
}

func (ms *memStorage) search(id ID) (i int, ok bool) {
	n := len(ms.list)
	i = sort.Search(n, func(i int) bool {
		return !ms.list[i].ID.Less(id)
	})
	return i, i < n && ms.list[i].ID == id
}

func (ms *memStorage) store(i int, take *Take) {
	p := new(Take)
	*p = *take
	ms.list[i] = p
}

func (ms *memStorage) Close() error {
	return nil
}

// NotFoundError is returned when a Storage method fails to find a take.
type NotFoundError struct {
	ID ID
}

func (e *NotFoundError) Error() string {
	return "takedata: ID " + e.ID.String() + " not found"
}

// ExistsError is returned when CreateTake is attempted with an existing ID.
type ExistsError struct {
	ID ID
}

func (e *ExistsError) Error() string {
	return "takedata: ID " + e.ID.String() + " already exists"
}
