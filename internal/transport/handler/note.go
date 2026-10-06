package handler

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type NotesHandler struct{}

func NewNotesHandler() *NotesHandler {
	notesHandler := NotesHandler{}
	return &notesHandler
}

func (h *NotesHandler) GetNotes(w http.ResponseWriter, r *http.Request) {

}

func (h *NotesHandler) GetNote(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	response := fmt.Sprintf("Note %s", id)
	fmt.Fprint(w, response)
}
