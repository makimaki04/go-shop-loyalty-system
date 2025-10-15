package models

import "time"

type User struct {
	ID           int
	Login        string
	PasswordHash string
	CreatedAt    time.Time
}