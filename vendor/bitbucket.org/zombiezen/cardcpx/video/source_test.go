/*
	Copyright 2015 Google Inc. All rights reserved.

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

package video

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const fileSize = 2 * 2 << 30

func TestDirectorySource_List(t *testing.T) {
	tests := []struct {
		root               string
		dirs               map[string][]string
		includeRedMagazine bool

		clips []*Clip
	}{
		{
			root: "/media/card",
			dirs: map[string][]string{
				"/media/card": {},
			},
			clips: []*Clip{},
		},
		{
			root: "/media/card",
			dirs: map[string][]string{
				"/media/card": {
					"a.mov",
					"b.mov",
				},
			},
			clips: []*Clip{
				{
					Name:      "a.mov",
					Paths:     []string{"a.mov"},
					TotalSize: fileSize,
				},
				{
					Name:      "b.mov",
					Paths:     []string{"b.mov"},
					TotalSize: fileSize,
				},
			},
		},
		{
			root: "/media/card",
			dirs: map[string][]string{
				"/media/card": {
					"a.mov",
					".DS_Store",
				},
			},
			clips: []*Clip{
				{
					Name:      "a.mov",
					Paths:     []string{"a.mov"},
					TotalSize: fileSize,
				},
			},
		},
		{
			root: "/media/card",
			dirs: map[string][]string{
				"/media/card": {
					"RedDirList.txt",
					"A002_0908FT.RDM",
					"digital_magazine.bin",
					"digital_magdynamic.bin",
				},
				"/media/card/A002_0908FT.RDM": {
					"A002_C001_0908R4.RDC",
				},
				"/media/card/A002_0908FT.RDM/A002_C001_0908R4.RDC": {
					"A002_C001_0908R4_001.R3D",
					"A002_C001_0908R4_002.R3D",
					"A002_C001_0908R4_003.R3D",
				},
			},
			clips: []*Clip{
				{
					Name: "A002_0908FT.RDM/A002_C001_0908R4.RDC",
					Paths: []string{
						"A002_0908FT.RDM/A002_C001_0908R4.RDC/A002_C001_0908R4_001.R3D",
						"A002_0908FT.RDM/A002_C001_0908R4.RDC/A002_C001_0908R4_002.R3D",
						"A002_0908FT.RDM/A002_C001_0908R4.RDC/A002_C001_0908R4_003.R3D",
					},
					TotalSize: fileSize * 3,
				},
			},
		},
		{
			root: "/media/card",
			dirs: map[string][]string{
				"/media/card": {
					"RedDirList.txt",
					"A002_0908FT.RDM",
					"digital_magazine.bin",
					"digital_magdynamic.bin",
				},
				"/media/card/A002_0908FT.RDM": {
					"A002_C001_0908R4.RDC",
				},
				"/media/card/A002_0908FT.RDM/A002_C001_0908R4.RDC": {
					"A002_C001_0908R4_001.R3D",
					"A002_C001_0908R4_002.R3D",
					"A002_C001_0908R4_003.R3D",
				},
			},
			includeRedMagazine: true,
			clips: []*Clip{
				{
					Name: "A002_0908FT.RDM digital_magazine",
					Paths: []string{
						"digital_magazine.bin",
						"digital_magdynamic.bin",
					},
					TotalSize: fileSize * 2,
				},
				{
					Name: "A002_0908FT.RDM/A002_C001_0908R4.RDC",
					Paths: []string{
						"A002_0908FT.RDM/A002_C001_0908R4.RDC/A002_C001_0908R4_001.R3D",
						"A002_0908FT.RDM/A002_C001_0908R4.RDC/A002_C001_0908R4_002.R3D",
						"A002_0908FT.RDM/A002_C001_0908R4.RDC/A002_C001_0908R4_003.R3D",
					},
					TotalSize: fileSize * 3,
				},
			},
		},
		{
			root: "/media/card",
			dirs: map[string][]string{
				"/media/card": {
					"RedDirList.txt",
					"A002_0908FT.RDM",
					"digital_magazine.bin",
					"digital_magdynamic.bin",
				},
				"/media/card/A002_0908FT.RDM": {
					"A002_C001_0908R4.RDC",
					"A002_C002_09081S.RDC",
					"A002_C003_0908ML.RDC",
					"A002_C004_0908JF.RDC",
					"A002_C005_0908EQ.RDC",
					"A002_C006_09081T.RDC",
					"A002_C007_0908ZA.RDC",
					"A002_C008_0908S1.RDC",
					"A002_C009_0908TN.RDC",
					"A002_C010_0908EV.RDC",
					"A002_C011_09081V.RDC",
					"A002_C012_0908TW.RDC",
					"A002_C013_0908KV.RDC",
				},
				"/media/card/A002_0908FT.RDM/A002_C001_0908R4.RDC": {
					"A002_C001_0908R4_001.R3D",
					"A002_C001_0908R4_002.R3D",
					"A002_C001_0908R4_003.R3D",
				},
				"/media/card/A002_0908FT.RDM/A002_C002_09081S.RDC": {
					"A002_C002_09081S_001.R3D",
					"A002_C002_09081S_002.R3D",
					"A002_C002_09081S_003.R3D",
					"A002_C002_09081S_004.R3D",
				},
				"/media/card/A002_0908FT.RDM/A002_C003_0908ML.RDC": {
					"A002_C003_0908ML_001.R3D",
					"A002_C003_0908ML_002.R3D",
				},
				"/media/card/A002_0908FT.RDM/A002_C004_0908JF.RDC": {
					"A002_C004_0908JF_001.R3D",
					"A002_C004_0908JF_002.R3D",
					"A002_C004_0908JF_003.R3D",
					"A002_C004_0908JF_004.R3D",
					"A002_C004_0908JF_005.R3D",
				},
				"/media/card/A002_0908FT.RDM/A002_C005_0908EQ.RDC": {
					"A002_C005_0908EQ_001.R3D",
					"A002_C005_0908EQ_002.R3D",
					"A002_C005_0908EQ_003.R3D",
					"A002_C005_0908EQ_004.R3D",
					"A002_C005_0908EQ_005.R3D",
				},
				"/media/card/A002_0908FT.RDM/A002_C006_09081T.RDC": {
					"A002_C006_09081T_001.R3D",
					"A002_C006_09081T_002.R3D",
					"A002_C006_09081T_003.R3D",
					"A002_C006_09081T_004.R3D",
					"A002_C006_09081T_005.R3D",
				},
				"/media/card/A002_0908FT.RDM/A002_C007_0908ZA.RDC": {
					"A002_C007_0908ZA_001.R3D",
				},
				"/media/card/A002_0908FT.RDM/A002_C008_0908S1.RDC": {
					"A002_C008_0908S1_001.R3D",
				},
				"/media/card/A002_0908FT.RDM/A002_C009_0908TN.RDC": {
					"A002_C009_0908TN_001.R3D",
					"A002_C009_0908TN_002.R3D",
				},
				"/media/card/A002_0908FT.RDM/A002_C010_0908EV.RDC": {
					"A002_C010_0908EV_001.R3D",
				},
				"/media/card/A002_0908FT.RDM/A002_C011_09081V.RDC": {
					"A002_C011_09081V_001.R3D",
					"A002_C011_09081V_002.R3D",
				},
				"/media/card/A002_0908FT.RDM/A002_C012_0908TW.RDC": {
					"A002_C012_0908TW_001.R3D",
					"A002_C012_0908TW_002.R3D",
				},
				"/media/card/A002_0908FT.RDM/A002_C013_0908KV.RDC": {
					"A002_C013_0908KV_001.R3D",
					"A002_C013_0908KV_002.R3D",
				},
			},
			clips: []*Clip{
				{
					Name: "A002_0908FT.RDM/A002_C001_0908R4.RDC",
					Paths: []string{
						"A002_0908FT.RDM/A002_C001_0908R4.RDC/A002_C001_0908R4_001.R3D",
						"A002_0908FT.RDM/A002_C001_0908R4.RDC/A002_C001_0908R4_002.R3D",
						"A002_0908FT.RDM/A002_C001_0908R4.RDC/A002_C001_0908R4_003.R3D",
					},
					TotalSize: fileSize * 3,
				},
				{
					Name: "A002_0908FT.RDM/A002_C002_09081S.RDC",
					Paths: []string{
						"A002_0908FT.RDM/A002_C002_09081S.RDC/A002_C002_09081S_001.R3D",
						"A002_0908FT.RDM/A002_C002_09081S.RDC/A002_C002_09081S_002.R3D",
						"A002_0908FT.RDM/A002_C002_09081S.RDC/A002_C002_09081S_003.R3D",
						"A002_0908FT.RDM/A002_C002_09081S.RDC/A002_C002_09081S_004.R3D",
					},
					TotalSize: fileSize * 4,
				},
				{
					Name: "A002_0908FT.RDM/A002_C003_0908ML.RDC",
					Paths: []string{
						"A002_0908FT.RDM/A002_C003_0908ML.RDC/A002_C003_0908ML_001.R3D",
						"A002_0908FT.RDM/A002_C003_0908ML.RDC/A002_C003_0908ML_002.R3D",
					},
					TotalSize: fileSize * 2,
				},
				{
					Name: "A002_0908FT.RDM/A002_C004_0908JF.RDC",
					Paths: []string{
						"A002_0908FT.RDM/A002_C004_0908JF.RDC/A002_C004_0908JF_001.R3D",
						"A002_0908FT.RDM/A002_C004_0908JF.RDC/A002_C004_0908JF_002.R3D",
						"A002_0908FT.RDM/A002_C004_0908JF.RDC/A002_C004_0908JF_003.R3D",
						"A002_0908FT.RDM/A002_C004_0908JF.RDC/A002_C004_0908JF_004.R3D",
						"A002_0908FT.RDM/A002_C004_0908JF.RDC/A002_C004_0908JF_005.R3D",
					},
					TotalSize: fileSize * 5,
				},
				{
					Name: "A002_0908FT.RDM/A002_C005_0908EQ.RDC",
					Paths: []string{
						"A002_0908FT.RDM/A002_C005_0908EQ.RDC/A002_C005_0908EQ_001.R3D",
						"A002_0908FT.RDM/A002_C005_0908EQ.RDC/A002_C005_0908EQ_002.R3D",
						"A002_0908FT.RDM/A002_C005_0908EQ.RDC/A002_C005_0908EQ_003.R3D",
						"A002_0908FT.RDM/A002_C005_0908EQ.RDC/A002_C005_0908EQ_004.R3D",
						"A002_0908FT.RDM/A002_C005_0908EQ.RDC/A002_C005_0908EQ_005.R3D",
					},
					TotalSize: fileSize * 5,
				},
				{
					Name: "A002_0908FT.RDM/A002_C006_09081T.RDC",
					Paths: []string{
						"A002_0908FT.RDM/A002_C006_09081T.RDC/A002_C006_09081T_001.R3D",
						"A002_0908FT.RDM/A002_C006_09081T.RDC/A002_C006_09081T_002.R3D",
						"A002_0908FT.RDM/A002_C006_09081T.RDC/A002_C006_09081T_003.R3D",
						"A002_0908FT.RDM/A002_C006_09081T.RDC/A002_C006_09081T_004.R3D",
						"A002_0908FT.RDM/A002_C006_09081T.RDC/A002_C006_09081T_005.R3D",
					},
					TotalSize: fileSize * 5,
				},
				{
					Name: "A002_0908FT.RDM/A002_C007_0908ZA.RDC",
					Paths: []string{
						"A002_0908FT.RDM/A002_C007_0908ZA.RDC/A002_C007_0908ZA_001.R3D",
					},
					TotalSize: fileSize,
				},
				{
					Name: "A002_0908FT.RDM/A002_C008_0908S1.RDC",
					Paths: []string{
						"A002_0908FT.RDM/A002_C008_0908S1.RDC/A002_C008_0908S1_001.R3D",
					},
					TotalSize: fileSize,
				},
				{
					Name: "A002_0908FT.RDM/A002_C009_0908TN.RDC",
					Paths: []string{
						"A002_0908FT.RDM/A002_C009_0908TN.RDC/A002_C009_0908TN_001.R3D",
						"A002_0908FT.RDM/A002_C009_0908TN.RDC/A002_C009_0908TN_002.R3D",
					},
					TotalSize: fileSize * 2,
				},
				{
					Name: "A002_0908FT.RDM/A002_C010_0908EV.RDC",
					Paths: []string{
						"A002_0908FT.RDM/A002_C010_0908EV.RDC/A002_C010_0908EV_001.R3D",
					},
					TotalSize: fileSize,
				},
				{
					Name: "A002_0908FT.RDM/A002_C011_09081V.RDC",
					Paths: []string{
						"A002_0908FT.RDM/A002_C011_09081V.RDC/A002_C011_09081V_001.R3D",
						"A002_0908FT.RDM/A002_C011_09081V.RDC/A002_C011_09081V_002.R3D",
					},
					TotalSize: fileSize * 2,
				},
				{
					Name: "A002_0908FT.RDM/A002_C012_0908TW.RDC",
					Paths: []string{
						"A002_0908FT.RDM/A002_C012_0908TW.RDC/A002_C012_0908TW_001.R3D",
						"A002_0908FT.RDM/A002_C012_0908TW.RDC/A002_C012_0908TW_002.R3D",
					},
					TotalSize: fileSize * 2,
				},
				{
					Name: "A002_0908FT.RDM/A002_C013_0908KV.RDC",
					Paths: []string{
						"A002_0908FT.RDM/A002_C013_0908KV.RDC/A002_C013_0908KV_001.R3D",
						"A002_0908FT.RDM/A002_C013_0908KV.RDC/A002_C013_0908KV_002.R3D",
					},
					TotalSize: fileSize * 2,
				},
			},
		},
	}

	for _, test := range tests {
		*includeRedMagazine = test.includeRedMagazine
		subject := fmt.Sprintf("newDirStructSource(%q, &fakeFilesystem{dirs: %v})",
			test.root, test.dirs)
		src, err := newDirStructSource(test.root, &fakeFilesystem{dirs: test.dirs})
		if err != nil {
			t.Errorf("%s: %v", subject, err)
			continue
		}

		clips, err := src.List()

		if err != nil {
			t.Errorf("%s.List(): %v", subject, err)
		}
		if !areClipListsEqual(clips, test.clips) {
			got, _ := json.MarshalIndent(clips, "", "  ")
			want, _ := json.MarshalIndent(test.clips, "", "  ")
			t.Errorf("%s.List() = %s; want %s", subject, got, want)
		}
	}
}

func areClipListsEqual(a, b []*Clip) bool {
	if len(a) != len(b) {
		return false
	}
	anames := make([]string, len(a))
	bnames := make([]string, len(b))
	for i := range a {
		anames[i] = a[i].Name
		bnames[i] = b[i].Name
	}
	if !areSameStrings(anames, bnames) {
		return false
	}
	for _, aClip := range a {
		bClip := findClip(b, aClip.Name)
		if bClip == nil || !areClipsEqual(aClip, bClip) {
			return false
		}
	}
	for _, bClip := range b {
		aClip := findClip(a, bClip.Name)
		if aClip == nil || !areClipsEqual(aClip, bClip) {
			return false
		}
	}
	return true
}

func findClip(clips []*Clip, name string) *Clip {
	for _, c := range clips {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func areClipsEqual(a, b *Clip) bool {
	return a.Name == b.Name && a.TotalSize == b.TotalSize && areSameStrings(a.Paths, b.Paths)
}

// areSameStrings reports whether r and s contain the same set of
// strings, regardless of order.
func areSameStrings(r, s []string) bool {
	if len(r) != len(s) {
		return false
	}
	for _, rr := range r {
		if !containsString(s, rr) {
			return false
		}
	}
	for _, ss := range s {
		if !containsString(r, ss) {
			return false
		}
	}
	return true
}

// containsString reports whether slice contains s.
func containsString(slice []string, s string) bool {
	for _, ss := range slice {
		if ss == s {
			return true
		}
	}
	return false
}

type fakeFilesystem struct {
	dirs map[string][]string
}

func (fs *fakeFilesystem) open(name string) (io.ReadCloser, error) {
	return nil, errors.New("fake filesystem")
}

func (fs *fakeFilesystem) readdirnames(name string) ([]string, error) {
	names, ok := fs.dirs[filepath.Clean(name)]
	if !ok {
		return nil, &os.PathError{
			Op:   "open",
			Path: name,
			Err:  os.ErrNotExist,
		}
	}
	return names, nil
}

func (fs *fakeFilesystem) walk(path string, f filepath.WalkFunc) error {
	info, err := fs.stat(path)
	if err != nil {
		return f(path, nil, err)
	}
	return fs.innerWalk(path, info, f)
}

func (fs *fakeFilesystem) innerWalk(path string, info os.FileInfo, f filepath.WalkFunc) error {
	err := f(path, info, nil)
	if err != nil {
		if info.IsDir() && err == filepath.SkipDir {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return nil
	}
	names, err := fs.readdirnames(path)
	if err != nil {
		return f(path, info, nil)
	}
	for _, name := range names {
		filename := filepath.Join(path, name)
		fileInfo, err := fs.stat(filename)
		if err != nil {
			if err := f(filename, fileInfo, err); err != nil && err != filepath.SkipDir {
				return err
			}
		} else {
			err = fs.innerWalk(filename, fileInfo, f)
			if err != nil {
				if !fileInfo.IsDir() || err != filepath.SkipDir {
					return err
				}
			}
		}
	}
	return nil
}

func (fs *fakeFilesystem) stat(path string) (os.FileInfo, error) {
	base := filepath.Base(path)
	if _, isDir := fs.dirs[path]; isDir {
		return fakeFileInfo{name: base, dir: true}, nil
	}
	if containsString(fs.dirs[filepath.Dir(path)], base) {
		return fakeFileInfo{name: base, size: fileSize}, nil
	}
	return nil, &os.PathError{
		Op:   "stat",
		Path: path,
		Err:  os.ErrNotExist,
	}
}

type fakeFileInfo struct {
	name string
	size int64
	dir  bool
}

func (info fakeFileInfo) Name() string {
	return info.name
}

func (info fakeFileInfo) Size() int64 {
	return info.size
}

func (info fakeFileInfo) IsDir() bool {
	return info.dir
}

func (info fakeFileInfo) Mode() os.FileMode {
	if info.dir {
		return os.ModeDir | 0755
	}
	return 0644
}

func (info fakeFileInfo) ModTime() time.Time {
	return time.Time{}
}

func (info fakeFileInfo) Sys() interface{} {
	return nil
}
