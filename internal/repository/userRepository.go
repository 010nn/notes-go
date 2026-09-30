package repository

import (
	"errors"
	"math/rand/v2"
	"notes-go/internal/model"
)

type UsersRepository struct {
}

func NewUsersRepository() *UsersRepository {
	usersdb = make(map[int64]model.User)

	return &UsersRepository{}
}

func (r *UsersRepository) GetByUsername(username string) (*model.User, error) {
	for _, u := range usersdb {
		if username == u.Username {
			return &u, nil
		}
	}
	return nil, errors.New("User 404")
}

func (r *UsersRepository) SaveUser(username string, passwordHash []byte) (*model.User, error) {

	userID := rand.Int64N(999999)
	user := model.User{
		ID:           userID,
		Username:     username,
		PasswordHash: passwordHash,
	}
	usersdb[userID] = user
	return &user, nil
}

var usersdb map[int64]model.User
