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

package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"bitbucket.org/zombiezen/cardcpx/httputil"
	"bitbucket.org/zombiezen/cardcpx/importer"
	"bitbucket.org/zombiezen/cardcpx/takedata"
	"bitbucket.org/zombiezen/cardcpx/video"
	"bitbucket.org/zombiezen/webapp"
	"github.com/gorilla/mux"
)

var (
	address       = flag.String("address", "localhost:8080", "address to listen on")
	localhostOnly = flag.Bool("localhostOnly", true, "only allow connections from localhost, same user")
	takeFile      = flag.String("takeFile", "takes.csv", "file to read and write takes")
	uiDir         = flag.String("uiDir", "ui", "directory with Angular app")
	storageDirs   = flag.String("storageDirs", "", "list of directories to import to. If empty, imports will be disabled.")
)

var takeStorage struct {
	takedata.Storage
	sync.RWMutex
}

var (
	videoStorage video.Storage
	imp          *importer.Importer
	impChan      chan importJob
)

func main() {
	flag.Parse()

	tf, err := os.OpenFile(*takeFile, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalln("Open take storage:", err)
	}
	defer tf.Close()
	takeStorage.Storage, err = takedata.NewCSVStorage(tf)
	if err != nil {
		log.Fatalln("Take storage:", err)
	}
	defer takeStorage.Close()
	videoStorage = openVideoStorage()
	if videoStorage != nil {
		imp = importer.New(videoStorage)
		impChan = make(chan importJob)
		go handleImports()
	}

	r := mux.NewRouter()
	r.Handle("/", http.RedirectHandler("/ui/", http.StatusMovedPermanently))
	r.Handle("/source", myHandler(source)).Methods("GET")
	r.Handle("/import", myHandler(getImportStatus)).Methods("GET")
	r.Handle("/import", myHandler(startImport)).Methods("POST")
	ts := r.PathPrefix("/take").Subrouter()
	ts.Handle("/", myHandler(listTake)).Methods("GET")
	ts.Handle("/", myHandler(insertTake)).Methods("POST")
	ts.Handle("/takes.csv", myHandler(csvExport)).Methods("GET")
	ts.Handle("/{scene}/{num}", myHandler(getTake)).Methods("GET")
	ts.Handle("/{scene}/{num}", myHandler(updateTake)).Methods("PUT")
	ts.Handle("/{scene}/{num}", myHandler(deleteTake)).Methods("DELETE")

	us := r.PathPrefix("/ui").Subrouter()
	us.HandleFunc("/importer", serveEntryPoint)
	us.HandleFunc("/take/{scene}/{num}", serveEntryPoint)
	us.Handle("/{_:(.*)}", http.StripPrefix("/ui", http.FileServer(http.Dir(*uiDir))))

	h := http.Handler(r)
	if *localhostOnly {
		h = &Auth{h, httputil.IsLocalhost}
	}
	log.Println("listening at", *address)
	if err := http.ListenAndServe(*address, h); err != nil {
		log.Fatalln("Listen:", err)
	}
}

func openVideoStorage() video.Storage {
	if *storageDirs == "" {
		return nil
	}
	dirs := strings.Split(*storageDirs, string(filepath.ListSeparator))
	s := make([]video.Storage, len(dirs))
	for i, path := range dirs {
		log.Printf("Opening %s for storage", path)
		s[i] = video.DirectoryStorage(path)
	}
	return video.MultiStorage(s...)
}

func handleImports() {
	for job := range impChan {
		log.Printf("Importing %d clips from %s (%d bytes)", len(job.Items), job.Path, job.Size())
		st := imp.Import(job.Src, job.Subdirectory, job.Clips())
		for _, result := range st.Results {
			if result.Error != nil {
				log.Printf("Import failed for %s: %v", result.Clip.Name, result.Error)
			}
		}
		log.Printf("Finished importing %d clips from %s (%d bytes copied)", len(job.Items), job.Path, st.BytesCopied)
		addTakesFromImport(&job, st)
	}
}

func addTakesFromImport(job *importJob, st *importer.Status) {
	takeStorage.Lock()
	defer takeStorage.Unlock()
	var added int
	for _, result := range st.Results {
		if result.Error != nil {
			continue
		}
		item := job.ClipItem(result.Clip)
		if item == nil {
			log.Printf("Missing clip item for %s", result.Clip.Name)
			continue
		}
		err := takeStorage.InsertTake(item.Take())
		if err != nil {
			log.Printf("Could not add %s to take storage: %v", result.Clip.Name, err)
			continue
		}
		added++
	}
	log.Printf("%d takes added", added)
}

func serveEntryPoint(w http.ResponseWriter, req *http.Request) {
	http.ServeFile(w, req, filepath.Join(*uiDir, "index.html"))
}

func listTake(w http.ResponseWriter, req *http.Request) error {
	takeStorage.RLock()
	defer takeStorage.RUnlock()

	takes, err := takeStorage.ListTakes()
	if err != nil {
		return err
	}
	return webapp.JSONResponse(w, takes)
}

func csvExport(w http.ResponseWriter, req *http.Request) error {
	takeStorage.RLock()
	defer takeStorage.RUnlock()

	takes, err := takeStorage.ListTakes()
	if err != nil {
		return err
	}
	w.Header().Set(webapp.HeaderContentType, "text/csv; charset=utf-8")
	cw := csv.NewWriter(w)
	if err := cw.Write([]string{"Scene", "Take", "Clip Name", "Select"}); err != nil {
		return err
	}
	for _, take := range takes {
		selectString := "FALSE"
		if take.Select {
			selectString = "TRUE"
		}
		err := cw.Write([]string{
			take.ID.Scene,
			take.ID.Num,
			take.ClipName,
			selectString,
		})
		if err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

func insertTake(w http.ResponseWriter, req *http.Request) error {
	takeStorage.Lock()
	defer takeStorage.Unlock()

	var take takedata.Take
	if err := json.NewDecoder(req.Body).Decode(&take); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil
	}
	if err := takeStorage.InsertTake(&take); err != nil {
		// TODO(light): handle ExistsError
		return err
	}
	return webapp.JSONResponse(w, &take)
}

func getTake(w http.ResponseWriter, req *http.Request) error {
	takeStorage.RLock()
	defer takeStorage.RUnlock()

	v := mux.Vars(req)
	take, err := takeStorage.GetTake(takedata.ID{
		Scene: v["scene"],
		Num:   v["num"],
	})
	if _, ok := err.(*takedata.NotFoundError); ok {
		http.Error(w, "take does not exist", http.StatusNotFound)
		return nil
	} else if err != nil {
		return err
	}
	return webapp.JSONResponse(w, take)
}

func updateTake(w http.ResponseWriter, req *http.Request) error {
	takeStorage.Lock()
	defer takeStorage.Unlock()

	var take takedata.Take
	if err := json.NewDecoder(req.Body).Decode(&take); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil
	}
	v := mux.Vars(req)
	id := takedata.ID{
		Scene: v["scene"],
		Num:   v["num"],
	}
	if err := takeStorage.UpdateTake(id, &take); err != nil {
		if _, ok := err.(*takedata.NotFoundError); !ok {
			return err
		}
		if id != take.ID {
			// tried to do a move, but resource doesn't exist
			http.Error(w, "take does not exist", http.StatusNotFound)
			return nil
		}
		// Upsert
		if err := takeStorage.InsertTake(&take); err != nil {
			return err
		}
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func deleteTake(w http.ResponseWriter, req *http.Request) error {
	takeStorage.Lock()
	defer takeStorage.Unlock()

	v := mux.Vars(req)
	id := takedata.ID{
		Scene: v["scene"],
		Num:   v["num"],
	}
	err := takeStorage.DeleteTake(id)
	if _, ok := err.(*takedata.NotFoundError); ok {
		http.Error(w, "take does not exist", http.StatusNotFound)
		return nil
	} else if err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func source(w http.ResponseWriter, req *http.Request) error {
	path := req.FormValue("path")
	if path == "" {
		return webapp.NotFound
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	src, err := video.DirectorySource(path)
	if err != nil {
		return err
	}
	clips, err := src.List()
	if err != nil {
		return err
	}
	return webapp.JSONResponse(w, clips)
}

func getImportStatus(w http.ResponseWriter, req *http.Request) error {
	if imp == nil {
		return webapp.JSONResponse(w, &importStatus{
			Enabled: false,
		})
	}
	st := imp.Status()
	return webapp.JSONResponse(w, &importStatus{
		Enabled: true,
		Status:  *st,
	})
}

func startImport(w http.ResponseWriter, req *http.Request) error {
	if impChan == nil {
		return webapp.JSONResponse(w, &importResponse{
			Code:         importResponseDisabled,
			ErrorMessage: "imports are disabled",
		})
	}
	var job importJob
	var err error
	if err = json.NewDecoder(req.Body).Decode(&job); err != nil {
		return webapp.JSONResponse(w, &importResponse{
			Code:         importResponseBadRequest,
			ErrorMessage: "could not parse job: " + err.Error(),
		})
	}
	if job.Src, err = video.DirectorySource(job.Path); err != nil {
		return webapp.JSONResponse(w, &importResponse{
			Code:         importResponseBadSource,
			ErrorMessage: "could not open source: " + err.Error(),
		})
	}
	select {
	case impChan <- job:
		return webapp.JSONResponse(w, &importResponse{
			Code: importResponseSuccess,
		})
	default:
		return webapp.JSONResponse(w, &importResponse{
			Code:         importResponseActive,
			ErrorMessage: "existing import in progress",
		})
	}
}

type importJob struct {
	Path         string       `json:"path"`
	Src          video.Source `json:"-"`
	Subdirectory string       `json:"subdirectory"`
	Items        []importItem `json:"items"`
}

func (job *importJob) Size() (size int64) {
	for i := range job.Items {
		size += job.Items[i].Clip.TotalSize
	}
	return
}

func (job *importJob) Clips() []*video.Clip {
	clips := make([]*video.Clip, len(job.Items))
	for i := range job.Items {
		clips[i] = &job.Items[i].Clip
	}
	return clips
}

func (job *importJob) ClipItem(clip *video.Clip) *importItem {
	for i := range job.Items {
		if &job.Items[i].Clip == clip {
			return &job.Items[i]
		}
	}
	return nil
}

type importItem struct {
	Clip   video.Clip `json:"clip"`
	Scene  string     `json:"scene"`
	Num    string     `json:"num"`
	Select bool       `json:"select"`
}

func (item *importItem) Take() *takedata.Take {
	return &takedata.Take{
		ID:       takedata.ID{Scene: item.Scene, Num: item.Num},
		ClipName: item.Clip.Name,
		Select:   item.Select,
	}
}

type importStatus struct {
	Enabled bool `json:"enabled"`
	importer.Status
}

type importResponse struct {
	Code         int    `json:"code,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}

// Import response error codes
const (
	importResponseSuccess    = 200
	importResponseBadRequest = 400
	importResponseDisabled   = 401
	importResponseActive     = 402
	importResponseBadSource  = 403
)

type myHandler func(http.ResponseWriter, *http.Request) error

func (h myHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	stats := webapp.NewResponseStats(res)
	err := h(stats, req)
	if err != nil {
		log.Println("Error:", err)
		if stats.StatusCode() == 0 {
			// response not yet written
			if webapp.IsNotFound(err) {
				http.NotFound(res, req)
			} else {
				http.Error(res, err.Error(), http.StatusInternalServerError)
			}
		}
	}
}

type Auth struct {
	http.Handler
	f func(*http.Request) bool
}

func (auth *Auth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if auth.f(r) {
		auth.Handler.ServeHTTP(w, r)
	} else {
		http.Error(w, "not authorized", http.StatusUnauthorized)
	}
}
