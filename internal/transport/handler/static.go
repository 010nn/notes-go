package handler

import (
	"net/http"
	"os"
)

type StaticHandler struct {
}

func NewStaticHandler() *StaticHandler {
	staticHandler := StaticHandler{}
	return &staticHandler
}

func (h *StaticHandler) Notes(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/html/notes.html")
}

func (h *StaticHandler) Auth(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/html/auth.html")
}

type NeuteredFileSystem struct {
	fs http.FileSystem
}

func NewNeuteredFileSystem(path string) *NeuteredFileSystem {
	return &NeuteredFileSystem{
		fs: http.Dir(path),
	}
}

func (nfs NeuteredFileSystem) Open(path string) (http.File, error) {
	f, err := nfs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if s.IsDir() {
		return nil, os.ErrNotExist
	}

	return f, nil
}
