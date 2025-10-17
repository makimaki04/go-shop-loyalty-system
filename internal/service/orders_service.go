package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/makimaki04/go-shop-loyalty-system/internal/models"
	"github.com/makimaki04/go-shop-loyalty-system/internal/repository"
	"go.uber.org/zap"
)

type OrdersService struct {
	repo   repository.Orders
	logger *zap.SugaredLogger
}

func NewOrdersService(repo repository.Orders, logger *zap.SugaredLogger) *OrdersService {
	return &OrdersService{
		repo: repo,
		logger: logger,
	}
}

var ErrInvalidOrderNumber = errors.New("invalid order number")

func (s *OrdersService) LoadOrder(ctx context.Context, userID int64, number string) error {
	if !checkLuhn(number) {
		return ErrInvalidOrderNumber
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return s.repo.LoadOrder(ctx, userID, number)
}


func checkLuhn(numb string) bool {
	numb = strings.TrimSpace(numb)
	if numb == "" {
		return false
	}

	for _, r := range numb {
		if r < '0' || r > '9' {
			return false
		}
	}

	sum := 0
	parity := len(numb) % 2

	for i, r := range numb {
		d := int(r - '0')

		if i % 2 == parity {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}

		sum += d
	}

	return sum % 10 == 0
}

func (s *OrdersService) GetOrders(ctx context.Context, userID int64) ([]models.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()

	return s.repo.GetOrders(ctx, userID)
}