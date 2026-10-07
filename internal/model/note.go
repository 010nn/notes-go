package model

import "time"

type Note struct {
	ID        string    `json:"id"`
	UserID    string    `json:"-"`
	Title     string    `json:"title"`
	Content   string    `json:"content,omitempty"`
	UpdatedAt time.Time `json:"updatedAt"`
	Version   int64     `json:"version"`
}

type NoteMeta struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	UpdatedAt time.Time `json:"updatedAt"`
}
