package memory

import "notes-go/internal/domain"

type UserDB map[string]*domain.User

type UserMemoryRepository struct {
	usersdb UserDB
}

func NewUserMemoryRepository() *UserMemoryRepository {
	return &UserMemoryRepository{
		usersdb: make(UserDB),
	}
}

func (r *UserMemoryRepository) GetByUsername(username string) (*domain.User, error) {
	return r.usersdb[username], nil
}

func (r *UserMemoryRepository) Create(username string, passwordHash []byte) (*domain.User, error) {

	user := &domain.User{
		Username:     username,
		PasswordHash: passwordHash,
	}
	r.usersdb[username] = user
	return user, nil
}
