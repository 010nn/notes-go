package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"notes-go/internal/domain"
	"notes-go/internal/transport"

	"golang.org/x/crypto/bcrypt"
)

// === User Repository ===
type userRepository interface {
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
}

// === Service ===
type UserService struct {
	repo       userRepository
	jwtManager *transport.JWTManager
}

func NewUserService(repo userRepository, jwtManager *transport.JWTManager) *UserService {
	return &UserService{
		repo:       repo,
		jwtManager: jwtManager,
	}
}

func (s *UserService) GetUser(ctx context.Context, username string) (*domain.User, error) {

	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	return user, nil

}

func (s *UserService) Register(ctx context.Context, username, password string) error {

	// Validation test
	if username == "" {
		return domain.ErrUsernameValidationFailed
	}

	if password == "" {
		return domain.ErrPasswordValidationFailed
	}

	// Find user by username
	_, err := s.repo.GetByUsername(ctx, username)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Get password hash
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to generate password hash: %w", err)
	}

	// Generate salt
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	user := &domain.User{
		Username:     username,
		PasswordHash: passwordHash,
		NotesSalt:    salt,
	}

	// Save
	if err = s.repo.Create(ctx, user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (s *UserService) Login(ctx context.Context, username, password string) (*domain.User, string, error) {

	// Validation test
	if username == "" {
		return nil, "", domain.ErrUsernameValidationFailed
	}

	if password == "" {
		return nil, "", domain.ErrPasswordValidationFailed
	}

	// Find user
	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, "", domain.ErrInvalidCredentials
		}
		return nil, "", fmt.Errorf("get user: %w", err)
	}

	// Check password hash
	if err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		return nil, "", domain.ErrInvalidCredentials
	}

	// Generate JWT token
	token, _, err := s.jwtManager.GenerateTokens(user.Username)
	if err != nil {
		log.Printf("error jwt: %v", err)
		return nil, "", fmt.Errorf("failed to generate JWT token: %w", err)
	}

	return user, token, nil
}
