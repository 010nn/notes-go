package service

import (
	"context"
	"errors"
	"notes-go/internal/model"
	"notes-go/internal/repository"
)

type NoteService struct {
	repo repository.NoteRepository
}

func NewNoteService(repo repository.NoteRepository) *NoteService {
	return &NoteService{
		repo: repo,
	}
}

func (s *NoteService) List(ctx context.Context, username string) ([]*model.Note, error) {
	return s.repo.List(ctx, username)
}

func (s *NoteService) Get(ctx context.Context, username, id string) (*model.Note, error) {
	return s.repo.Get(ctx, username, id)
}

func (s *NoteService) Update(ctx context.Context, username, id, title, content string, version int64) (*model.Note, error) {
	note, err := s.repo.Get(ctx, username, id)
	if err != nil {
		if errors.Is(err, model.ErrNoteNotFound) {
			return s.repo.Create(ctx, username, id, title, content, 1)
		}
		return nil, err
	}

	// Проверка версии
	if note.Version != version {
		return nil, model.ErrVersionMismatch
	}

	if _, err := s.repo.Update(ctx, username, id, title, content, version); err != nil {
		return nil, err
	}

	return note, nil
}

func (s *NoteService) Delete(ctx context.Context, username, id string) error {
	return s.repo.Delete(ctx, username, id)
}
