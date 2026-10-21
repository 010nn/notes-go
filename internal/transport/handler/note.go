package handler

import (
	"encoding/json"
	"net/http"
	"notes-go/internal/domain"
	"notes-go/internal/service"
	"notes-go/internal/transport"
	"notes-go/internal/transport/response"
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

// GET /notes
func (h *NotesHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username, ok := transport.UserFromCtx(ctx)
	if !ok {
		response.HandleError(w, r, domain.ErrUserNotFound)
		return
	}

	notes, err := h.service.List(r.Context(), username)
	if err != nil {
		response.HandleError(w, r, err)
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

	response.JSON(w, http.StatusOK, notesResponse)
}

// POST /notes/{id}
func (h *NotesHandler) Update(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	ctx := r.Context()
	username, ok := transport.UserFromCtx(ctx)
	if !ok {
		response.HandleError(w, r, domain.ErrUserNotFound)
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		response.HandleError(w, r, domain.ErrInvalidNoteId)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		response.HandleError(w, r, domain.ErrInvalidNoteId)
		return
	}

	var req updateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.HandleError(w, r, domain.ErrInvalidJSON)
		return
	}

	note, err := h.service.Save(r.Context(), username, id, req.Title, req.Content, req.Version)
	if err != nil {
		response.HandleError(w, r, err)
		return
	}

	noteVersionResponse := noteVersionResponse{
		ID:        note.ID,
		UpdatedAt: note.UpdatedAt,
		Version:   note.Version,
	}

	response.JSON(w, http.StatusOK, noteVersionResponse)
}

// POST /notes/{id}
func (h *NotesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	ctx := r.Context()
	username, ok := transport.UserFromCtx(ctx)
	if !ok {
		response.HandleError(w, r, domain.ErrUserNotFound)
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		response.HandleError(w, r, domain.ErrInvalidNoteId)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		response.HandleError(w, r, domain.ErrInvalidNoteId)
		return
	}

	if err := h.service.Delete(r.Context(), username, id); err != nil {
		response.HandleError(w, r, err)
		return
	}

	response.JSON(w, http.StatusOK, nil)
}
