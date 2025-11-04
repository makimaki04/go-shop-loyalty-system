package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/makimaki04/go-shop-loyalty-system/internal/models"
	"go.uber.org/zap"
)

const (
	selectBalanceQuery = `
		SELECT current, withdrawn
		FROM balances
		WHERE user_id = $1
	`
	selectBalanceForUpdate = `
		SELECT current
		FROM balances
		WHERE user_id = $1
		FOR UPDATE
	`
	updateBalance = `
		UPDATE balances
		SET current = current - $1, withdrawn = withdrawn + $1
		WHERE user_id = $2
	`
	insertWithdrawal = `
		INSERT INTO withdrawals (user_id, order_number, amount)
		VALUES ($1, $2, $3)
	`
	selectWithdrawals = `
		SELECT order_number, amount, processed_at
		FROM withdrawals
		WHERE user_id = $1
		ORDER BY processed_at DESC
	`
	updateCurrentBalance= `
		UPDATE balances
		SET current = current + $1
		WHERE user_id = (
			SELECT user_id FROM orders 
			WHERE number = $2 AND status = 'PROCESSED'
		)
	`
)

type BalanceRepository struct {
	db     *sql.DB
	logger *zap.SugaredLogger
}

func NewBalanceRepository(db *sql.DB, logger *zap.SugaredLogger) *BalanceRepository {
	return &BalanceRepository{
		db:     db,
		logger: logger,
	}
}

func (r *BalanceRepository) GetBalance(ctx context.Context, userID int64) (models.Balance, error) {
	var balance models.Balance

	r.logger.Infof("getting balance info for user %d", userID)

	err := r.db.QueryRowContext(ctx, selectBalanceQuery, userID).Scan(&balance.Current, &balance.Withdrawn)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Warnw("no balance record found, returning zeros", "user_id", userID)
			return models.Balance{UserID: userID, Current: 0, Withdrawn: 0}, nil
		}

		r.logger.Errorw("failed to get balance from DB", "user_id", userID, "error", err)
		return models.Balance{}, err
	}

	r.logger.Infof("user %d balance successfully fetched from DB", userID)
	return balance, nil
}

func (r *BalanceRepository) WithdrawBonuses(ctx context.Context, withdraw models.Withdraw) (err error) {
	r.logger.Infow("starting withdrawal transaction",
		"user_id", withdraw.UserID,
		"order", withdraw.Order,
		"sum", withdraw.Sum,
	)

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		r.logger.Errorw("failed to start transaction", "error", err)
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
			r.logger.Warnw("transaction rolled back", "user_id", withdraw.UserID, "error", err)
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				r.logger.Errorw("failed to commit transaction", "user_id", withdraw.UserID, "error", commitErr)
				err = fmt.Errorf("commit: %w", commitErr)
			} else {
				r.logger.Debugw("transaction committed successfully", "user_id", withdraw.UserID)
			}
		}
	}()

	var current float64

	err = tx.QueryRowContext(ctx, selectBalanceForUpdate, withdraw.UserID).Scan(&current)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Warnw("no balances found for user", "user_id", withdraw.UserID)
			return fmt.Errorf("no balance found: %w", err)
		}
		r.logger.Errorw("failed to get balance for update", "user_id", withdraw.UserID, "error", err)
		return fmt.Errorf("failed to get balance for update: %w", err)
	}

	if current < withdraw.Sum {
		r.logger.Warnw("insufficient funds", "user_id", withdraw.UserID, "current", current, "sum", withdraw.Sum)
		return fmt.Errorf("insufficient funds: current=%.2f, need=%.2f", current, withdraw.Sum)
	}

	_, err = tx.ExecContext(ctx, updateBalance, withdraw.Sum, withdraw.UserID)
	if err != nil {
		r.logger.Errorw("failed to update balance", "user_id", withdraw.UserID, "error", err)
		return fmt.Errorf("failed to update balance: %w", err)
	}

	_, err = tx.ExecContext(ctx, insertWithdrawal, withdraw.UserID, withdraw.Order, withdraw.Sum)
	if err != nil {
		r.logger.Errorw("failed to insert withdrawal record", "user_id", withdraw.UserID, "order", withdraw.Order, "error", err)
		return fmt.Errorf("failed to insert withdrawal: %w", err)
	}

	r.logger.Infow("withdrawal completed successfully",
		"user_id", withdraw.UserID,
		"order", withdraw.Order,
		"sum", withdraw.Sum,
	)

	return nil
}

func (r *BalanceRepository) GetWithdrawals(ctx context.Context, userID int64) ([]models.Withdrawals, error) {
	r.logger.Infow("fetching withdrawals for user", "user_id", userID)

	rows, err := r.db.QueryContext(ctx, selectWithdrawals, userID)
	if err != nil {
		r.logger.Errorf("query row context error %v", err)
		return nil, fmt.Errorf("failed to get withdrawals for user %d: %w", userID, err)
	}
	defer rows.Close()

	var withdrawals []models.Withdrawals

	for rows.Next() {
		w := models.Withdrawals{
			UserID: userID,
		}
		err := rows.Scan(&w.OrderNumber, &w.Sum, &w.ProcessedAt)
		if err != nil {
			r.logger.Errorw("failed rows scan", "error", err)
			return nil, fmt.Errorf("failed to get withdrawals for user %d: %w", userID, err)
		}

		withdrawals = append(withdrawals, w)
	}

	if err := rows.Err(); err != nil {
		r.logger.Errorw("rows iteration error", "user_id", userID, "error", err)
		return nil, fmt.Errorf("failed to get withdrawals for user %d: %w", userID, err)
	}

	if len(withdrawals) == 0 {
		r.logger.Infow("no withdrawals found", "userID", userID)
		return []models.Withdrawals{}, nil
	}

	r.logger.Infow("withdrawals successfully fetched", "user_id", userID, "count", len(withdrawals))

	return withdrawals, nil
}

func (r *BalanceRepository) AddAccrual(ctx context.Context, order string, accrual float64) error {
	_, err := r.db.ExecContext(ctx, updateCurrentBalance, accrual, order)
	if err != nil {
		r.logger.Warnw("failed to update user balance", "order", order, "error", err)
		return fmt.Errorf("failed to update order %s status: %v", order, err)
	}

	return nil
}