package service

import (
	"context"
	"errors"
	"fmt"
	"notes-go/internal/domain"
	"uuid"
)

// === Repository Interface ===
type noteRepository interface {
	List(ctx context.Context, username string) ([]*domain.Note, error)
	Get(ctx context.Context, username string, id uuid.UUID) (*domain.Note, error)
	Create(ctx context.Context, username string, id uuid.UUID, title, content string, version int64) (*domain.Note, error)
	Update(ctx context.Context, username string, id uuid.UUID, title, content string, version int64) (*domain.Note, error)
	Delete(ctx context.Context, username string, id uuid.UUID) error
}

// === Service ===
type NoteService struct {
	repo noteRepository
}

func NewNoteService(repo noteRepository) *NoteService {
	return &NoteService{
		repo: repo,
	}
}

func (s *NoteService) List(ctx context.Context, username string) ([]*domain.Note, error) {
	return s.repo.List(ctx, username)
}

func (s *NoteService) Get(ctx context.Context, username string, id uuid.UUID) (*domain.Note, error) {
	return s.repo.Get(ctx, username, id)
}

func (s *NoteService) Update(ctx context.Context, username string, id uuid.UUID, title, content string, version int64) (*domain.Note, error) {
	note, err := s.repo.Get(ctx, username, id)
	if err != nil {
		if errors.Is(err, domain.ErrNoteNotFound) {
			return s.repo.Create(ctx, username, id, title, content, 1)
		}
		return nil, fmt.Errorf("service get note: %w", err)
	}

	// Проверка версии
	if note.Version != version {
		return nil, domain.ErrVersionMismatch
	}

	updatedNote, err := s.repo.Update(ctx, username, id, title, content, version)
	if err != nil {
		return nil, fmt.Errorf("service update note: %w", err)
	}

	return updatedNote, nil
}

func (s *NoteService) Delete(ctx context.Context, username string, id uuid.UUID) error {
	return s.repo.Delete(ctx, username, id)
}
