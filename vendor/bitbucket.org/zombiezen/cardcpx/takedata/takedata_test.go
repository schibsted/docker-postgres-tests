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
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestNewMemoryStorage(t *testing.T) {
	newStorageTest(t, NewMemoryStorage())
}

func newStorageTest(t *testing.T, s Storage) {
	defer closeStorage(t, s)
	checkListTakes(t, s)
	id := ID{"1", "2"}
	checkDeleted(t, s, id)
	take := &Take{ID: id, ClipName: "foo"}
	err := s.UpdateTake(id, take)
	if nf, ok := err.(*NotFoundError); !ok || nf.ID != id {
		t.Errorf("UpdateTake(%#v, %+v) error: %v", id, take, err)
	}
	checkDeleted(t, s, id)
}

func TestMemoryStorageInsertion(t *testing.T) {
	storageInsertionTest(t, NewMemoryStorage())
}

func storageInsertionTest(t *testing.T, s Storage) {
	defer closeStorage(t, s)
	insertTakes(t, s,
		&Take{ID: ID{Scene: "3", Num: "1"}, ClipName: "baz"},
		&Take{ID: ID{Scene: "1", Num: "1"}, ClipName: "foo"},
		&Take{ID: ID{Scene: "2", Num: "1"}, ClipName: "bar"},
	)
	checkListTakes(t, s,
		&Take{ID: ID{Scene: "1", Num: "1"}, ClipName: "foo"},
		&Take{ID: ID{Scene: "2", Num: "1"}, ClipName: "bar"},
		&Take{ID: ID{Scene: "3", Num: "1"}, ClipName: "baz"},
	)
	checkGetTake(t, s, &Take{
		ID:       ID{Scene: "1", Num: "1"},
		ClipName: "foo",
	})
	dupe := &Take{ID: ID{Scene: "1", Num: "1"}, ClipName: "bacon"}
	err := s.InsertTake(dupe)
	if ex, ok := err.(*ExistsError); !ok || ex.ID != dupe.ID {
		t.Errorf("InsertTake(%+v) error: %v", dupe, err)
	}
}

func TestMemoryStorageUpdateInPlace(t *testing.T) {
	storageUpdateInPlaceTest(t, NewMemoryStorage())
}

func storageUpdateInPlaceTest(t *testing.T, s Storage) {
	defer closeStorage(t, s)

	id := ID{"1", "1"}
	initTake := &Take{ID: id, ClipName: "foo"}
	upTake := &Take{ID: id, ClipName: "bacon"}

	insertTakes(t, s, initTake)
	err := s.UpdateTake(id, upTake)
	if err != nil {
		t.Errorf("UpdateTake(%#v, %+v) error: %v", id, upTake, err)
	}
	checkGetTake(t, s, upTake)
}

func TestMemoryStorageUpdateToExisting(t *testing.T) {
	storageUpdateToExistingTest(t, NewMemoryStorage())
}

func storageUpdateToExistingTest(t *testing.T, s Storage) {
	defer closeStorage(t, s)

	id1 := ID{"1", "1"}
	id2 := ID{"2", "1"}
	initTake1 := &Take{ID: id1, ClipName: "foo"}
	initTake2 := &Take{ID: id2, ClipName: "bar"}
	upTake := &Take{ID: id2, ClipName: "bacon"}

	insertTakes(t, s, initTake1, initTake2)
	err := s.UpdateTake(id1, upTake)
	if ex, ok := err.(*ExistsError); !ok || ex.ID != upTake.ID {
		t.Errorf("UpdateTake(%#v, %+v) error: %v", id1, upTake, err)
	}
	checkGetTake(t, s, initTake1)
	checkGetTake(t, s, initTake2)
}

func TestMemoryStorageUpdateMove(t *testing.T) {
	storageUpdateMoveTest(t, NewMemoryStorage)
}

func storageUpdateMoveTest(t *testing.T, newStorage func() Storage) {
	tests := []struct {
		init     []ID
		from, to ID
		post     []ID
	}{
		{
			init: []ID{{"1", ""}, {"2", ""}},
			from: ID{"1", ""},
			to:   ID{"3", ""},
			post: []ID{{"2", ""}, {"3", ""}},
		},
		{
			init: []ID{{"2", ""}, {"3", ""}},
			from: ID{"3", ""},
			to:   ID{"1", ""},
			post: []ID{{"1", ""}, {"2", ""}},
		},
		{
			init: []ID{{"2", ""}, {"4", ""}},
			from: ID{"4", ""},
			to:   ID{"3", ""},
			post: []ID{{"2", ""}, {"3", ""}},
		},
	}
	for _, test := range tests {
		s := newStorage()
		for _, id := range test.init {
			s.InsertTake(&Take{ID: id})
		}
		err := s.UpdateTake(test.from, &Take{ID: test.to})
		if err != nil {
			t.Errorf("init=%v UpdateTake(%v, %v) error: %v", test.init, test.from, test.to, err)
		}
		takes := make([]*Take, len(test.post))
		for i := range takes {
			takes[i] = &Take{ID: test.post[i]}
		}
		checkListTakes(t, s, takes...)
		closeStorage(t, s)
	}
}

func TestMemoryStorageDelete(t *testing.T) {
	storageDeleteTest(t, NewMemoryStorage())
}

func storageDeleteTest(t *testing.T, s Storage) {
	defer closeStorage(t, s)
	insertTakes(t, s,
		&Take{ID: ID{Scene: "3", Num: "1"}, ClipName: "baz"},
		&Take{ID: ID{Scene: "1", Num: "1"}, ClipName: "foo"},
		&Take{ID: ID{Scene: "2", Num: "1"}, ClipName: "bar"},
	)
	err := s.DeleteTake(ID{Scene: "2", Num: "1"})
	if err != nil {
		t.Error("DeleteTakes error:", err)
	}
	checkListTakes(t, s,
		&Take{ID: ID{Scene: "1", Num: "1"}, ClipName: "foo"},
		&Take{ID: ID{Scene: "3", Num: "1"}, ClipName: "baz"},
	)
	checkDeleted(t, s, ID{"2", "1"})
}

func insertTakes(t *testing.T, s Storage, takes ...*Take) {
	file, line := caller(1)
	for _, take := range takes {
		err := s.InsertTake(take)
		if err != nil {
			t.Errorf("%s:%d: InsertTake(%+v) error: %v", file, line, take, err)
		}
	}
}

func checkListTakes(t *testing.T, s Storage, want ...*Take) {
	file, line := caller(1)
	takes, err := s.ListTakes()
	if err != nil {
		t.Error("%s:%d: ListTakes() error:", file, line, err)
	}
	if !takesEqual(takes, want) {
		t.Errorf("%s:%d: ListTakes() =\n%s, want\n%s", file, line, takesString(takes), takesString(want))
	}
}

func checkGetTake(t *testing.T, s Storage, want *Take) {
	file, line := caller(1)
	take, err := s.GetTake(want.ID)
	if err != nil {
		t.Errorf("%s:%d: GetTake(%#v) error: %v", file, line, want.ID, err)
	}
	if !takeEqual(take, want) {
		t.Errorf("%s:%d: GetTake(%#v) = %+v, want %+v", file, line, want.ID, take, want)
	}
}

func checkDeleted(t *testing.T, s Storage, id ID) {
	file, line := caller(1)
	take, err := s.GetTake(id)
	if nf, ok := err.(*NotFoundError); !ok || nf.ID != id {
		t.Errorf("%s:%d: GetTake(%#v) error: %v", file, line, id, err)
	}
	if take != nil {
		t.Errorf("%s:%d: GetTake(%#v) = %+v, want nil", file, line, id, take)
	}
}

func caller(skip int) (file string, line int) {
	_, file, line, _ = runtime.Caller(skip + 1)
	file = filepath.Base(file)
	return
}

func takesEqual(t1, t2 []*Take) bool {
	if len(t1) != len(t2) {
		return false
	}
	for i := range t1 {
		if !takeEqual(t1[i], t2[i]) {
			return false
		}
	}
	return true
}

func takeEqual(take1, take2 *Take) bool {
	switch {
	case take1 == nil && take2 == nil:
		return true
	case take1 == nil || take2 == nil:
		return false
	default:
		return *take1 == *take2
	}
}

func takesString(t []*Take) string {
	s := make([]string, len(t))
	for i := range t {
		s[i] = fmt.Sprintf("%+v", t[i])
	}
	return "[" + strings.Join(s, " ") + "]"
}

func closeStorage(t *testing.T, storage Storage) {
	err := storage.Close()
	if err != nil {
		t.Error("Close() error:", err)
	}
}
