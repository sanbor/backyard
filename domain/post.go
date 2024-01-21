package domain

import (
	"time"
)

type Post struct {
	ID      string
	Title   string
	Content string
	Draft   bool
	Access
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Access struct {
	UserID   string
	Relation string
}
