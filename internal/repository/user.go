package repository

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

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		SELECT username, password_hash, notes_salt
		FROM users
		WHERE username = $1;
	`

	var u domain.User
	err := r.db.QueryRow(ctx, query, username).Scan(
		&u.Username,
		&u.PasswordHash,
		&u.NotesSalt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("db repo user GetByUsername() username: %s err: %w", username, err)
	}

	return &u, nil
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (username, password_hash, notes_salt)
		VALUES ($1, $2, $3);
	`

	_, err := r.db.Exec(ctx, query, user.Username, user.PasswordHash, user.NotesSalt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrUserAlreadyExists
		}
		return fmt.Errorf("db repo user Create() username: %s err: %w", user.Username, err)
	}

	return nil
}
