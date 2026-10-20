package postgres

import (
	"context"
	"errors"
	"fmt"
	"notes-go/internal/domain"

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

func (r *NoteRepository) Create(ctx context.Context, note *domain.Note) (int64, error) {
	query := `
		INSERT INTO notes (user_id, title, content, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id;
	`

	var id int64
	err := r.db.QueryRow(ctx, query, note.Username, note.Title, note.Content, note.CreatedAt, note.UpdatedAt).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert note: %w", err)
	}

	return id, nil
}

func (r *NoteRepository) GetByID(ctx context.Context, id int64) (*domain.Note, error) {
	query := `
		SELECT id, user_id, title, content, created_at, updated_at
		FROM notes
		WHERE id = $1;
	`

	var note domain.Note
	err := r.db.QueryRow(ctx, query, id).Scan(
		&note.ID,
		&note.Username,
		&note.Title,
		&note.Content,
		&note.CreatedAt,
		&note.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNoteNotFound
		}
		return nil, fmt.Errorf("select note by id: %w", err)
	}

	return &note, nil
}
