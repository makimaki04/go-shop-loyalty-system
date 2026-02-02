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

type Balance struct {
	UserID int64
	Current float64
	Withdrawn float64
}

type Withdraw struct {
	UserID int64
	Order string
	Sum float64
}

type Withdrawals struct {
	UserID int64
	OrderNumber string
	Sum float64
	ProcessedAt time.Time
}

type PendingOrder struct {
	Number string
	Status string
}