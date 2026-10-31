package domain

import (
	"time"
	"uuid"
)

type Note struct {
	ID        uuid.UUID
	Username  string
	Title     string
	Content   string
	Position  int64
	Version   int64
	CreatedAt time.Time
	UpdatedAt time.Time
}
