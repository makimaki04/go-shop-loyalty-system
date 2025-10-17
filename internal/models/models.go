package models

import "time"

type User struct {
	ID           int64
	Login        string
	PasswordHash string
	CreatedAt    time.Time
}

type Order struct {
	Number string
	Status string
	Accrual float64
	UploadedAt time.Time
}