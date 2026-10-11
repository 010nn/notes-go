package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"notes-go/internal/service"
	"notes-go/internal/transport"

	"github.com/go-chi/chi/v5"
)

type NotesHandler struct {
	service *service.NoteService
}

func NewNotesHandler(service *service.NoteService) *NotesHandler {
	notesHandler := NotesHandler{
		service: service,
	}
	return &notesHandler
}

// DTO
type updateNoteRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Version int64  `json:"version"`
}

// Handlers

// GET /notes/{id}
func (h *NotesHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username, ok := transport.UserFromCtx(ctx)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	note, err := h.service.Get(r.Context(), username, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	transport.WriteJSON(w, http.StatusOK, note)
}

// GET /notes
func (h *NotesHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username, ok := transport.UserFromCtx(ctx)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	notes, err := h.service.List(r.Context(), username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	transport.WriteJSON(w, http.StatusOK, notes)
}

// POST /notes/{id}
func (h *NotesHandler) Update(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	ctx := r.Context()
	username, ok := transport.UserFromCtx(ctx)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	var req updateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	note, err := h.service.Update(r.Context(), username, id, req.Title, req.Content, req.Version)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
	}

	transport.WriteJSON(w, http.StatusOK, note)
}

// POST /notes/{id}
func (h *NotesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	ctx := r.Context()
	username, ok := transport.UserFromCtx(ctx)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(r.Context(), username, id); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	transport.WriteJSON(w, http.StatusOK, nil)
}
