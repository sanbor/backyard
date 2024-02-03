package domain

import (
	"time"
)

type Config struct {
	ID              string
	Title           string
	Description     string
	ImageHome       string
	Favicon         string
	Footer          string
	BackyardVersion string
	Active          bool
	AdminUserID     string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
