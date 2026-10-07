package repository

import "notes-go/internal/model"

type UserRepository interface {
	GetByUsername(username string) (*model.User, error)
	Create(username string, passwordHash []byte) (*model.User, error)
}
