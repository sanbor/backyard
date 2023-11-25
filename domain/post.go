package domain

import (
	"time"
)

type Post struct {
	ID        string
	Title     string
	Content   string
	Author    string
	Draft     bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
