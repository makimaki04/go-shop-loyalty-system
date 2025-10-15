package service

import (
	"github.com/makimaki04/go-shop-loyalty-system/internal/models"
	"github.com/makimaki04/go-shop-loyalty-system/internal/repository"
	"go.uber.org/zap"
)

type Authorization interface {
	CreateUser(login, password string) (int, error)
	GenerateToken(id int) (accessToken string, err error)
	LoginUser(login, password string) (models.User, error)
}

type Service struct {
	Authorization
}

func NewService(repo *repository.Repository, logger *zap.SugaredLogger) *Service {
	return &Service{
		Authorization: NewAuthService(repo.Authorization, logger),
	}
}