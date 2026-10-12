package service

import (
	"context"
	"log"
	"notes-go/internal/domain"
	"notes-go/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

// === User Repository ===
type userRepository interface {
	GetByUsername(username string) (*domain.User, error)
	Create(username string, passwordHash []byte) (*domain.User, error)
}

// === Service ===
type UserService struct {
	repo       userRepository
	jwtManager *jwt.JWTManager
}

func NewUserService(repo userRepository, jwtManager *jwt.JWTManager) *UserService {
	return &UserService{
		repo:       repo,
		jwtManager: jwtManager,
	}
}

func (s *UserService) Register(ctx context.Context, username, password string) error {

	// Validation test
	if username == "" {
		return domain.ErrUsernameValidationFailed
	}

	if password == "" {
		return domain.ErrPasswordValidationFailed
	}

	// Ищем пользователя в БД по username
	if user, _ := s.repo.GetByUsername(username); user != nil {
		return domain.ErrUserAlreadyExists
	}

	// Получение хэша пароля
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.ErrInternal
	}

	// Сохранение в БД
	if _, err = s.repo.Create(username, passwordHash); err != nil {
		return domain.ErrInternal
	}

	return nil
}

func (s *UserService) Login(ctx context.Context, username, password string) (string, error) {

	// Validation test
	if username == "" {
		return "", domain.ErrUsernameValidationFailed
	}

	if password == "" {
		return "", domain.ErrPasswordValidationFailed
	}

	// Find user
	user, err := s.repo.GetByUsername(username)
	if user == nil {
		return "", domain.ErrInvalidCredentials
	}

	// Check password hash
	if err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		return "", domain.ErrInvalidCredentials
	}

	// Generate JWT token
	token, _, err := s.jwtManager.GenerateTokens(user.Username)
	if err != nil {
		log.Printf("error jwt: %v", err)
		return "", domain.ErrInternal
	}

	return token, nil
}
