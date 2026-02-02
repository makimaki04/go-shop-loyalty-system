package repository

import (
	"context"
	"database/sql"

	"github.com/makimaki04/go-shop-loyalty-system/internal/models"
	"go.uber.org/zap"
)

type Authorization interface {
	CreateUser(ctx context.Context, user models.User) (int64, error)
	LoginUser(ctx context.Context, login string) (models.User, error)
}

type Orders interface {
	LoadOrder(ctx context.Context, userID int64, number string) error
	GetOrders(ctx context.Context, userID int64) ([]models.Order, error)
	GetPendingorders(ctx context.Context) ([]models.PendingOrder, error)
	UpdateOrderStatus(ctx context.Context, order string, status OrderStatus, accrual float64) error
}

type Balance interface {
	GetBalance(ctx context.Context, userID int64) (models.Balance, error)
	WithdrawBonuses(ctx context.Context, withdraw models.Withdraw) error
	GetWithdrawals(ctx context.Context, userID int64) ([]models.Withdrawals, error)
	AddAccrual(ctx context.Context, order string, accrual float64) error
}

type Repository struct {
	Authorization
	Orders
	Balance
}

func NewRepository(db *sql.DB, logger *zap.SugaredLogger) *Repository {
	return &Repository{
		Authorization: NewAuthRepository(db, logger),
		Orders:        NewOrdersRepository(db, logger),
		Balance:       NewBalanceRepository(db, logger),
	}
}
