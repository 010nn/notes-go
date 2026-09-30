package handler

import "net/http"

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
