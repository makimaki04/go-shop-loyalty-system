package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/makimaki04/go-shop-loyalty-system/internal/models"
	"github.com/makimaki04/go-shop-loyalty-system/internal/repository"
	"go.uber.org/zap"
)

type BalanceService struct {
	balanceRepo repository.Balance
	logger      *zap.SugaredLogger
}

func NewBalanceService(balanceRepo repository.Balance, logger *zap.SugaredLogger) *BalanceService {
	return &BalanceService{
		balanceRepo: balanceRepo,
		logger:      logger,
	}
}

func (s *BalanceService) GetBalance(ctx context.Context, userID int64) (models.Balance, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return s.balanceRepo.GetBalance(ctx, userID)
}

var (
	ErrInsufficientFunds = errors.New("insufficient funds to withdraw")
)

func (s *BalanceService) WithdrawBonuses(ctx context.Context, withdraw models.Withdraw) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if withdraw.Sum <= 0 {
		return fmt.Errorf("invalid withdraw sum: %.2f", withdraw.Sum)
	}

	ok := CheckLuhn(withdraw.Order)
	if !ok {
		return ErrInvalidOrderNumber
	}

	balance, err := s.balanceRepo.GetBalance(ctx, withdraw.UserID)
	if err != nil {
		s.logger.Warnw("BalanceService: failed to get user balance", "userID", withdraw.UserID, "error", err)
		return fmt.Errorf("failed to get balance: %w", err)
	}

	if balance.Current < withdraw.Sum {
		return ErrInsufficientFunds
	}

	s.logger.Infow("withdrawal processed successfully",
		"user_id", withdraw.UserID,
		"sum", withdraw.Sum,
		"order", withdraw.Order,
	)

	return s.balanceRepo.WithdrawBonuses(ctx, withdraw)
}

func (s *BalanceService) GetWithdrawals(ctx context.Context, userID int64) ([]models.Withdrawals, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return s.balanceRepo.GetWithdrawals(ctx, userID)
}
