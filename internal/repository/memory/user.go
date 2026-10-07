package memory

import (
	"notes-go/internal/model"
)

type UserDB map[string]*model.User

type UserMemoryRepository struct {
	usersdb UserDB
}

func NewUserMemoryRepository() *UserMemoryRepository {
	return &UserMemoryRepository{
		usersdb: make(UserDB),
	}
}

func (r *UserMemoryRepository) GetByUsername(username string) (*model.User, error) {
	return r.usersdb[username], nil
}

func (r *UserMemoryRepository) Create(username string, passwordHash []byte) (*model.User, error) {

	user := &model.User{
		Username:     username,
		PasswordHash: passwordHash,
	}
	r.usersdb[username] = user
	return user, nil
}
