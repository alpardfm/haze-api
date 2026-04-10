package auth

import "time"

type Admin struct {
	ID           int64
	Name         string
	Email        string
	Phone        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
