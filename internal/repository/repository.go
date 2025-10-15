package repository

import (
	"database/sql"

	"github.com/makimaki04/go-shop-loyalty-system/internal/models"
	"go.uber.org/zap"
)

type Authorization interface {
	CreateUser(user models.User) (int, error)
	LoginUser(login string) (models.User, error)
}

type Repository struct {
	Authorization
}

func NewRepository(db *sql.DB, logger *zap.SugaredLogger) *Repository {
	return &Repository{
		Authorization: NewAuthRepository(db, logger),
	}
}