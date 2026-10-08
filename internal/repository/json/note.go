package json

import (
	"context"
	"notes-go/internal/model"
	"strconv"
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
	Notes  JSONStructure
	lastID int
	mu     sync.Mutex
}

func NewNoteJSONRepository() *NoteJSONRepository {
	rep := &NoteJSONRepository{
		Notes:  JSONStructure{Users: make(Users)},
		lastID: 0,
	}
	return rep
}

func (r *NoteJSONRepository) Create(_ context.Context, username, title, content string) (*model.Note, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Проверяем, если ти пользователь
	notes := r.Notes.Users[username]
	if notes == nil {
		return nil, model.ErrUserNotFound
	}

	r.lastID++
	id := strconv.Itoa(r.lastID)
	note := &model.Note{
		ID:        id,
		UserID:    username,
		Title:     title,
		Content:   content,
		UpdatedAt: time.Now(),
		Version:   1,
	}

	notes[id] = note

	return note, nil
}

func (r *NoteJSONRepository) Get(_ context.Context, userID, id string) (*model.Note, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Проверяем, есть ли пользователь
	notes, userExists := r.Notes.Users[userID]
	if !userExists {
		return nil, model.ErrUserNotFound
	}

	// Проверяем, есть ли такая заметка у пользователя
	note, noteExists := notes[id]
	if !noteExists {
		return nil, model.ErrNoteNotFound
	}

	return note, nil
}

func (r *NoteJSONRepository) Update(_ context.Context, userID, id, title, content string, version int64) (*model.Note, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Проверяем наличие пользователя
	notes, userExists := r.Notes.Users[userID]
	if !userExists {
		return nil, model.ErrUserNotFound
	}

	// Проверяем наличие заметки
	note, noteExists := notes[id]
	if !noteExists {
		return nil, model.ErrNoteNotFound
	}

	// Проверка оптимистичной блокировки (Optimistic Locking)
	// Если версия в базе не совпадает с тем, что прислал клиент — значит данные устарели
	if note.Version != version {
		return nil, model.ErrVersionMismatch
	}

	// Обновляем поля заметки
	note.Title = title
	note.Content = content
	note.UpdatedAt = time.Now()
	note.Version++ // Инкрементируем версию при каждом успешном обновлении

	return note, nil
}

func (r *NoteJSONRepository) Delete(_ context.Context, userID, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Проверяем наличие пользователя
	notes, userExists := r.Notes.Users[userID]
	if !userExists {
		return model.ErrUserNotFound
	}

	// Проверяем наличие заметки перед удалением
	if _, noteExists := notes[id]; !noteExists {
		return model.ErrNoteNotFound
	}

	// Удаляем заметку из карты пользователя
	delete(notes, id)

	// Опционально: если у пользователя больше нет заметок, очищаем карту Users, чтобы не копился мусор
	if len(notes) == 0 {
		delete(r.Notes.Users, userID)
	}

	return nil
}
