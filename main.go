package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"github.com/gorilla/mux"
)

type ByModTime []os.FileInfo

func (fis ByModTime) Len() int {
	return len(fis)
}

func (fis ByModTime) Swap(i, j int) {
	fis[i], fis[j] = fis[j], fis[i]
}

func (fis ByModTime) Less(i, j int) bool {
	//return fis[i].ModTime().Before(fis[j].ModTime())
	return fis[i].ModTime().After(fis[j].ModTime())
}

type Filedata struct {
	Uid  string
	Name string
	Tag  string
	Date string
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

var tpl *template.Template

func init() {
	tpl = template.Must(template.ParseGlob("templates/*"))
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", index)
	r.HandleFunc("/{key}", index)
	http.Handle("/", r)
	http.HandleFunc("/save/", uploadsHandler)
	http.Handle("/neo/", http.StripPrefix("/neo/", http.FileServer(http.Dir("./neo"))))
	http.Handle("/gallery/", http.StripPrefix("/gallery/", http.FileServer(http.Dir("./gallery"))))
	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.ListenAndServe(":8082", nil)
}

func index(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	key, ok := vars["key"]
	if ok {
		if key == "viewer.html" {
			tpl.ExecuteTemplate(w, "viewer.html", nil)
			return
		} else if key == "index.html" {

		} else {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}

	file, err := os.Open("gallery")
	checkError(err)
	defer file.Close()

	fis, _ := file.Readdir(-1)
	sort.Sort(ByModTime(fis))

	list := make([]Filedata, len(fis))

	idx := 0

	for _, fi := range fis {
		list[idx].Name = fi.Name()
		list[idx].Date = fi.ModTime().String()
		idx++
		//fmt.Println(fi.Name())
	}
	tpl.ExecuteTemplate(w, "index.html", list)
}

type FileRes struct {
	Name string `json:"name"`
}

func uploadsHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		mf, _, err := req.FormFile("nf")
		if err != nil {
			fmt.Println(err)
		}
		defer mf.Close()
		// create sha for file name
		h := sha1.New()
		io.Copy(h, mf)
		fname := fmt.Sprintf("%x", h.Sum(nil)) + ".png"
		// create new file
		wd, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
		}
		path := filepath.Join(wd, "gallery", fname)
		nf, err := os.Create(path)
		if err != nil {
			fmt.Println(err)
		}
		defer nf.Close()
		mf.Seek(0, 0)
		io.Copy(nf, mf)

		// response file name : json
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fileres := FileRes{fname}
		enc := json.NewEncoder(w)
		enc.Encode(fileres)
	}
}
