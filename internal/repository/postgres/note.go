package postgres

import (
	"context"
	"errors"
	"fmt"
	"notes-go/internal/domain"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NoteRepository struct {
	db *pgxpool.Pool
}

func NewNoteRepository(db *pgxpool.Pool) *NoteRepository {
	return &NoteRepository{
		db: db,
	}
}

func (r *NoteRepository) List(ctx context.Context, username string) ([]*domain.Note, error) {
	query := `
		SELECT id, username, title, content, updated_at, created_at, version
		FROM notes
		WHERE username = $1
	`

	rows, err := r.db.Query(ctx, query, username)
	if err != nil {
		return nil, fmt.Errorf("db repo note List() username %s: %w", username, err)
	}

	notes, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*domain.Note, error) {
		var n domain.Note
		err := row.Scan(&n.ID, &n.Username, &n.Title, &n.Content, &n.UpdatedAt, &n.CreatedAt, &n.Version)
		return &n, err
	})

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("db repo note List() rows error for username %s: %w", username, err)
	}

	return notes, nil
}

func (r *NoteRepository) Get(ctx context.Context, _ string, id uuid.UUID) (*domain.Note, error) {
	query := `
		SELECT id, username, title, content, updated_at, created_at, version
		FROM notes
		WHERE id = $1
	`

	var n domain.Note
	err := r.db.QueryRow(ctx, query, id).Scan(
		&n.ID, &n.Username, &n.Title, &n.Content, &n.UpdatedAt, &n.CreatedAt, &n.Version,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNoteNotFound
		}
		return nil, fmt.Errorf("db repo note Get() id %s: %w", id, err)
	}

	return &n, nil
}

func (r *NoteRepository) Save(ctx context.Context, note *domain.Note, oldVersion int64) (*domain.Note, error) {
	query := `
		INSERT INTO notes (id, username, title, content, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (id) DO UPDATE
		SET
			title = EXCLUDED.title,
			content = EXCLUDED.content,
			version = EXCLUDED.version,
			updated_at = NOW()
		WHERE notes.version = $6
		RETURNING created_at, updated_at;
	`

	err := r.db.QueryRow(ctx, query,
		note.ID,
		note.Username,
		note.Title,
		note.Content,
		note.Version,
		oldVersion,
	).Scan(
		&note.CreatedAt,
		&note.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrVersionMismatch
		}
		return nil, fmt.Errorf("db repo note Save(): %w", err)
	}

	return note, nil
}

func (r *NoteRepository) Delete(ctx context.Context, username string, id uuid.UUID) error {
	query := `
		DELETE FROM notes
		WHERE id = $1 AND username = $2;
	`

	ct, err := r.db.Exec(ctx, query, id, username)
	if err != nil {
		return fmt.Errorf("db note repo Delete(): %w", err)
	}

	if ct.RowsAffected() == 0 {
		return fmt.Errorf("note not deleted")
	}

	return nil
}
