package handler

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type NotesHandler struct{}

func NewNotesHandler() *NotesHandler {
	notesHandler := NotesHandler{}
	return &notesHandler
}

func (h *NotesHandler) GetNotes(w http.ResponseWriter, r *http.Request) {

}

func (h *NotesHandler) GetNote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	response := fmt.Sprintf("Note %s", id)
	fmt.Fprint(w, response)
}
