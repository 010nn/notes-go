package service

import (
	"notes-go/internal/model"
	"notes-go/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo *repository.UsersRepository
}

func NewUserService(repo *repository.UsersRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) Register(auth model.AuthRequest) error {

	// Validation test
	if auth.Username == "" {
		return model.ErrUsernameValidationFailed
	}

	if auth.Password == "" {
		return model.ErrPasswordValidationFailed
	}

	// Ищем пользователя в БД по username
	if user, _ := s.repo.GetByUsername(auth.Username); user != nil {
		return model.ErrUserAlreadyExists
	}

	// Получение хэша пароля
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(auth.Password), bcrypt.DefaultCost)
	if err != nil {
		return model.ErrInternal
	}

	// Сохранение в БД
	if _, err = s.repo.SaveUser(auth.Username, passwordHash); err != nil {
		return model.ErrInternal
	}

	return nil
}

func (s *UserService) Login(auth model.AuthRequest) (*model.User, error) {

	// Validation test
	if auth.Username == "" {
		return nil, model.ErrUsernameValidationFailed
	}

	if auth.Password == "" {
		return nil, model.ErrPasswordValidationFailed
	}

	// Find user
	user, err := s.repo.GetByUsername(auth.Username)
	if user == nil {
		return nil, model.ErrInvalidCredentials
	}

	// Check password hash
	if err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(auth.Password)); err != nil {
		return nil, model.ErrInvalidCredentials
	}

	return user, nil
}
