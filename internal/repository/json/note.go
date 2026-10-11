package json

import (
	"context"
	"encoding/json"
	"errors"
	"notes-go/internal/model"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	notesPath = "pgdata/notes.json"
)

type NotesList map[string]*model.Note
type Users map[string]NotesList

type JSONStructure struct {
	Users Users `json:"users"`
}

type NoteJSONRepository struct {
	Notes JSONStructure
	mu    sync.Mutex
}

func NewNoteJSONRepository() *NoteJSONRepository {
	r := &NoteJSONRepository{
		Notes: JSONStructure{Users: make(Users)},
	}
	r.load()
	return r
}

func (r *NoteJSONRepository) load() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := os.ReadFile(notesPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Файла ещё нет — начинаем с пустой структуры
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
		return err
	}

	if structure.Users == nil {
		structure.Users = make(Users)
	}

	r.Notes = structure

	return nil
}

func (r *NoteJSONRepository) save() error {
	if err := os.MkdirAll(filepath.Dir(notesPath), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(r.Notes, "", "  ")
	if err != nil {
		return err
	}

	// Атомарная запись: сначала во временный файл, потом переименование
	tmpPath := notesPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return err
	}

	return os.Rename(tmpPath, notesPath)
}

func (r *NoteJSONRepository) List(_ context.Context, username string) ([]*model.Note, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	notes := r.getNotesByName(username)

	notesArr := make([]*model.Note, 0, len(*notes))

	for _, v := range *notes {
		notesArr = append(notesArr, v)
	}
	return notesArr, nil
}

func (r *NoteJSONRepository) Get(_ context.Context, username, id string) (*model.Note, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Проверяем, есть ли пользователь
	notes := *r.getNotesByName(username)

	// Проверяем, есть ли такая заметка у пользователя
	note, noteExists := notes[id]
	if !noteExists {
		return nil, model.ErrNoteNotFound
	}

	return note, nil
}

func (r *NoteJSONRepository) Create(_ context.Context, username, id, title, content string, version int64) (*model.Note, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Проверяем, если ти пользователь
	notes := *r.getNotesByName(username)

	note := &model.Note{
		ID:        id,
		UserID:    username,
		Title:     title,
		Content:   content,
		UpdatedAt: time.Now(),
		CreatedAt: time.Now(),
		Version:   version,
	}

	notes[id] = note

	if err := r.save(); err != nil {
		delete(notes, id)
		return nil, err
	}

	return note, nil
}

func (r *NoteJSONRepository) Update(_ context.Context, username, id, title, content string, version int64) (*model.Note, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Проверяем наличие пользователя
	notes := *r.getNotesByName(username)

	// Проверяем наличие заметки
	note, noteExists := notes[id]
	if !noteExists {
		return nil, model.ErrNoteNotFound
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
		return nil, err
	}

	return note, nil
}

func (r *NoteJSONRepository) Delete(_ context.Context, username, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Проверяем наличие пользователя
	notes := *r.getNotesByName(username)

	// Проверяем наличие заметки перед удалением
	note, noteExists := notes[id]
	if !noteExists {
		return model.ErrNoteNotFound
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
		return err
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
