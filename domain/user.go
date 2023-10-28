package domain

import (
	"errors"
	"time"
)

type User struct {
	ID        string
	Username  string
	Email     *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (u User) ValidateEmail() error {
	if u.Email != nil && len(*u.Email) < 3 {
		return errors.New("bad")
	}
	return nil
}
