package service

import (
	"context"
	"fmt"
	"log"
	"notes-go/internal/domain"
	"notes-go/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

// === User Repository ===
type userRepository interface {
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	Create(ctx context.Context, username string, passwordHash []byte) (*domain.User, error)
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
	if user, _ := s.repo.GetByUsername(ctx, username); user != nil {
		return domain.ErrUserAlreadyExists
	}

	// Получение хэша пароля
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to generate password hash: %w", err)
	}

	// Сохранение в БД
	if _, err = s.repo.Create(ctx, username, passwordHash); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
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
	user, err := s.repo.GetByUsername(ctx, username)
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
		return "", fmt.Errorf("failed to generate JWT token: %w", err)
	}

	return token, nil
}
