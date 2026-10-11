package repository

import (
	"context"
	"notes-go/internal/model"
)

type UserRepository interface {
	GetByUsername(username string) (*model.User, error)
	Create(username string, passwordHash []byte) (*model.User, error)
}

type NoteRepository interface {
	List(ctx context.Context, userID string) ([]*model.Note, error)
	Get(ctx context.Context, userID, id string) (*model.Note, error)
	Create(ctx context.Context, userID, id, title, content string, version int64) (*model.Note, error)
	Update(ctx context.Context, userID, id, title, content string, version int64) (*model.Note, error)
	Delete(ctx context.Context, userID, id string) error
}
