package repository

import (
	"notes-go/internal/model"
)

type UserDB map[string]*model.User

type UsersRepository struct {
	usersdb UserDB
}

func NewUsersRepository() *UsersRepository {
	return &UsersRepository{
		usersdb: make(UserDB),
	}
}

func (r *UsersRepository) GetByUsername(username string) (*model.User, error) {
	return r.usersdb[username], nil
}

func (r *UsersRepository) SaveUser(username string, passwordHash []byte) (*model.User, error) {

	user := &model.User{
		Username:     username,
		PasswordHash: passwordHash,
	}
	r.usersdb[username] = user
	return user, nil
}
