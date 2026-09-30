package service

import (
	"errors"
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

func (s *UserService) Register(username string, password string) error {

	// Ищем пользователя в БД по username
	if user, _ := s.repo.GetByUsername(username); user != nil {
		return errors.New("Username already exists")
	}

	// Получение хэша пароля
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("Password hash")
	}

	// Сохранение в БД
	if _, err = s.repo.SaveUser(username, passwordHash); err != nil {
		return errors.New("Saving user")
	}

	return nil
}

func (s *UserService) Login(username string, password string) (*model.User, error) {

	// Ищем пользователя в БД по username
	user, err := s.repo.GetByUsername(username)
	if err != nil {
		return nil, errors.New("User not found")
	}

	// Проверяем пароль (сравниваем хэш из БД с введенным паролем)
	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("Invalid username or password")
	}

	return user, nil
}
