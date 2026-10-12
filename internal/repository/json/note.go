package json

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"notes-go/internal/domain"
	"os"
	"path/filepath"
	"sync"
	"time"
	"uuid"
)

const (
	notesPath = "pgdata/notes.json"
)

type noteJSON struct {
	Title     string    `json:"title"`
	Content   string    `json:"content,omitempty"`
	UpdatedAt time.Time `json:"updatedAt"`
	CreatedAt time.Time `json:"createdAt"`
	Version   int64     `json:"version"`
}

func (note *noteJSON) toDomainNote(username string, id uuid.UUID) *domain.Note {
	return &domain.Note{
		ID:        id,
		Username:  username,
		Title:     note.Title,
		Content:   note.Content,
		UpdatedAt: note.UpdatedAt,
		CreatedAt: note.CreatedAt,
		Version:   note.Version,
	}
}

type NotesList map[uuid.UUID]*noteJSON
type Users map[string]NotesList

type JSONStructure struct {
	Users Users `json:"users"`
}

type NoteJSONRepository struct {
	Notes JSONStructure
	mu    sync.Mutex
}

func NewNoteJSONRepository() (*NoteJSONRepository, error) {
	r := &NoteJSONRepository{
		Notes: JSONStructure{Users: make(Users)},
	}
	if err := r.load(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *NoteJSONRepository) load() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := os.ReadFile(notesPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			r.Notes = JSONStructure{Users: make(Users)}
			return nil
		}
		return err
	}

	if len(data) == 0 {
		r.Notes = JSONStructure{Users: make(Users)}
		return nil
	}

	var structure JSONStructure
	if err := json.Unmarshal(data, &structure); err != nil {
		return fmt.Errorf("NoteJSON load: %w", err)
	}

	if structure.Users == nil {
		structure.Users = make(Users)
	}

	r.Notes = structure

	return nil
}

func (r *NoteJSONRepository) save() error {
	if err := os.MkdirAll(filepath.Dir(notesPath), 0o755); err != nil {
		return fmt.Errorf("NoteJSON save: %w", err)
	}

	data, err := json.MarshalIndent(r.Notes, "", "  ")
	if err != nil {
		return fmt.Errorf("NoteJSON save: %w", err)
	}

	// Атомарная запись: сначала во временный файл, потом переименование
	tmpPath := notesPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return fmt.Errorf("NoteJSON save: %w", err)
	}

	return os.Rename(tmpPath, notesPath)
}

func (r *NoteJSONRepository) List(_ context.Context, username string) ([]*domain.Note, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	notes := r.getNotesByName(username)

	notesArr := make([]*domain.Note, 0, len(*notes))

	for id, note := range *notes {
		notesArr = append(notesArr, note.toDomainNote(username, id))
	}
	return notesArr, nil
}

func (r *NoteJSONRepository) Get(_ context.Context, username string, id uuid.UUID) (*domain.Note, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Проверяем, есть ли пользователь
	notes := *r.getNotesByName(username)

	// Проверяем, есть ли такая заметка у пользователя
	note, noteExists := notes[id]
	if !noteExists {
		return nil, domain.ErrNoteNotFound
	}

	return note.toDomainNote(username, id), nil
}

func (r *NoteJSONRepository) Create(_ context.Context, username string, id uuid.UUID, title, content string, version int64) (*domain.Note, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Проверяем, если ти пользователь
	notes := *r.getNotesByName(username)

	note := &noteJSON{
		Title:     title,
		Content:   content,
		UpdatedAt: time.Now(),
		CreatedAt: time.Now(),
		Version:   version,
	}

	notes[id] = note

	if err := r.save(); err != nil {
		delete(notes, id)
		return nil, fmt.Errorf("save json: %w", err)
	}

	return note.toDomainNote(username, id), nil
}

func (r *NoteJSONRepository) Update(_ context.Context, username string, id uuid.UUID, title, content string, version int64) (*domain.Note, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Проверяем наличие пользователя
	notes := *r.getNotesByName(username)

	// Проверяем наличие заметки
	note, noteExists := notes[id]
	if !noteExists {
		return nil, domain.ErrNoteNotFound
	}

	// сохраняем старое состояние для отката
	oldTitle := note.Title
	oldContent := note.Content
	oldUpdatedAt := note.UpdatedAt
	oldVersion := note.Version

	// Обновляем поля заметки
	note.Title = title
	note.Content = content
	note.UpdatedAt = time.Now()
	note.Version++ // Инкрементируем версию при каждом успешном обновлении

	if err := r.save(); err != nil {
		note.Title = oldTitle
		note.Content = oldContent
		note.UpdatedAt = oldUpdatedAt
		note.Version = oldVersion
		return nil, fmt.Errorf("save json: %w", err)
	}

	return note.toDomainNote(username, id), nil
}

func (r *NoteJSONRepository) Delete(_ context.Context, username string, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Проверяем наличие пользователя
	notes := *r.getNotesByName(username)

	// Проверяем наличие заметки перед удалением
	note, noteExists := notes[id]
	if !noteExists {
		return domain.ErrNoteNotFound
	}

	// Удаляем заметку из карты пользователя
	delete(notes, id)

	// Опционально: если у пользователя больше нет заметок, очищаем карту Users, чтобы не копился мусор
	if len(notes) == 0 {
		delete(r.Notes.Users, username)
	}

	if err := r.save(); err != nil {
		if _, ok := r.Notes.Users[username]; !ok {
			r.Notes.Users[username] = make(NotesList)
		}
		r.Notes.Users[username][id] = note
		return fmt.Errorf("save json: %w", err)
	}

	return nil
}

func (r *NoteJSONRepository) getNotesByName(username string) *NotesList {
	notes, userExists := r.Notes.Users[username]
	if !userExists {
		notes = make(NotesList)
		r.Notes.Users[username] = notes
	}
	return &notes
}
