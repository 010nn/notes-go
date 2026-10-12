package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"notes-go/internal/service"
	"notes-go/internal/transport"
	"time"
	"uuid"

	"github.com/go-chi/chi/v5"
)

// === DTO ===
type updateNoteRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Version int64  `json:"version"`
}

type noteVersionResponse struct {
	ID        uuid.UUID `json:"id"`
	UpdatedAt time.Time `json:"updatedAt"`
	Version   int64     `json:"version"`
}

type noteResponse struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	UpdatedAt time.Time `json:"updatedAt"`
	CreatedAt time.Time `json:"createdAt"`
	Version   int64     `json:"version"`
}

// === Handlers ===

type NotesHandler struct {
	service *service.NoteService
}

func NewNotesHandler(service *service.NoteService) *NotesHandler {
	notesHandler := NotesHandler{
		service: service,
	}
	return &notesHandler
}

// GET /notes/{id}
func (h *NotesHandler) Get(w http.ResponseWriter, r *http.Request) {
	// ctx := r.Context()
	// username, ok := transport.UserFromCtx(ctx)
	// if !ok {
	// 	http.Error(w, "unauthorized", http.StatusUnauthorized)
	// 	return
	// }

	// idStr := chi.URLParam(r, "id")
	// if idStr == "" {
	// 	http.Error(w, "id is required", http.StatusBadRequest)
	// 	return
	// }

	// id, err := uuid.Parse(idStr)
	// if err != nil {
	// 	http.Error(w, "invalid id", http.StatusBadRequest)
	// }

	// note, err := h.service.Get(r.Context(), username, id)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusUnauthorized)
	// 	return
	// }

	// transport.WriteJSON(w, http.StatusOK, note)
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

	notesResponse := make([]noteResponse, 0, len(notes))
	for _, note := range notes {
		notesResponse = append(notesResponse, noteResponse{
			ID:        note.ID,
			Title:     note.Title,
			Content:   note.Content,
			UpdatedAt: note.UpdatedAt,
			CreatedAt: note.CreatedAt,
			Version:   note.Version,
		})
	}

	transport.WriteJSON(w, http.StatusOK, notesResponse)
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

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
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
		return
	}

	noteVersionResponse := noteVersionResponse{
		ID:        note.ID,
		UpdatedAt: note.UpdatedAt,
		Version:   note.Version,
	}

	transport.WriteJSON(w, http.StatusOK, noteVersionResponse)
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

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
	}

	if err := h.service.Delete(r.Context(), username, id); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	transport.WriteJSON(w, http.StatusOK, nil)
}
