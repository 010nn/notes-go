package postgres

import (
	"context"
	"errors"
	"fmt"
	"notes-go/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// GetByUsername выбирает пользователя по его username
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		SELECT username, password_hash
		FROM users
		WHERE username = $1;
	`

	var u domain.User
	err := r.db.QueryRow(ctx, query, username).Scan(
		&u.Username,
		&u.PasswordHash,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Обязательно возвращаем доменную ошибку (не забудь объявить её в domain)
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("db repo user GetByUsername() username: %s err: %w", username, err)
	}

	// Поле ID в базе данных отсутствует согласно твоей миграции,
	// оставляем его нулевым или дефолтным.
	u.ID = 0

	return &u, nil
}

// Create создает нового пользователя в базе данных
func (r *UserRepository) Create(ctx context.Context, username string, passwordHash []byte) (*domain.User, error) {
	query := `
		INSERT INTO users (username, password_hash)
		VALUES ($1, $2);
	`

	_, err := r.db.Exec(ctx, query, username, passwordHash)
	if err != nil {
		// Проверяем специфичную ошибку PostgreSQL на нарушение уникальности (пользователь уже существует)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrUserAlreadyExists // Объяви эту ошибку в домене
		}
		return nil, fmt.Errorf("db repo user Create() username: %s err: %w", username, err)
	}

	return &domain.User{
		ID:           0,
		Username:     username,
		PasswordHash: passwordHash,
	}, nil
}
