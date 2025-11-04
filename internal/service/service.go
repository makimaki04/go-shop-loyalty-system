package service

import (
	"context"

	"github.com/makimaki04/go-shop-loyalty-system/internal/models"
	"github.com/makimaki04/go-shop-loyalty-system/internal/repository"
	"go.uber.org/zap"
)

type Authorization interface {
	CreateUser(ctx context.Context, login, password string) (int64, error)
	GenerateToken(id int64) (accessToken string, err error)
	LoginUser(ctx context.Context, login, password string) (models.User, error)
}

type Orders interface {
	LoadOrder(ctx context.Context, userID int64, number string) error
	GetOrders(ctx context.Context, userID int64) ([]models.Order, error)
}

type Balance interface {
	GetBalance(ctx context.Context, userID int64) (models.Balance, error)
	WithdrawBonuses(ctx context.Context, withdraw models.Withdraw) error
	GetWithdrawals(ctx context.Context, userID int64) ([]models.Withdrawals, error)
}

type Service struct {
	Authorization
	Orders
	Balance
}

func NewService(repo *repository.Repository, logger *zap.SugaredLogger) *Service {
	return &Service{
		Authorization: NewAuthService(repo.Authorization, logger),
		Orders:        NewOrdersService(repo.Orders, logger),
		Balance:       NewBalanceService(repo.Balance, logger),
	}
}
