package handler

import "net/http"

type StaticHandler struct {
}

func NewStaticHandler() *StaticHandler {
	staticHandler := StaticHandler{}
	return &staticHandler
}

func (h *StaticHandler) Notes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	http.ServeFile(w, r, "./static/html/notes.html")
}

func (h *StaticHandler) Auth(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/html/auth.html")
}
