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
	UpdatedAt time.Time
	CreatedAt time.Time
	Version   int64
}
