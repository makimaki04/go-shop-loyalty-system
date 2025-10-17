package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/makimaki04/go-shop-loyalty-system/internal/models"
	"go.uber.org/zap"
)

const (
	insertOrderQuery = `
		INSERT INTO orders (user_id, number, status, accrual)
		VALUES ($1, $2, $3, $4)
	`
	checkUserQuery = `
		SELECT user_id
		FROM orders
		WHERE number = $1
	`
	getUserOrdersQuery = `
		SELECT number, status, accrual, uploaded_at
		FROM orders as o
		WHERE user_id = $1
		ORDER BY uploaded_at DESC
	`
)

type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

type OrdersRepository struct {
	db     *sql.DB
	logger *zap.SugaredLogger
}

func NewOrdersRepository(db *sql.DB, logger *zap.SugaredLogger) *OrdersRepository {
	return &OrdersRepository{
		db:     db,
		logger: logger,
	}
}

var (
	ErrOrderExists             = errors.New("order already uploaded")
	ErrOrderOwnedByAnotherUser = errors.New("order owned by another user")
)

func (r *OrdersRepository) LoadOrder(ctx context.Context, userID int64, number string) error {
	r.logger.Infof("user %d trying to upload order %s", userID, number)

	_, err := r.db.ExecContext(ctx, insertOrderQuery, userID, number, OrderStatusNew, 0)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			var existingUser int64

			checkErr := r.db.QueryRowContext(ctx, checkUserQuery, number).Scan(&existingUser)
			if checkErr != nil {
				r.logger.Errorf("user %d failed to check existing order %s", userID, number)
				return fmt.Errorf("failed to check existing order: %w", checkErr)
			}

			if existingUser == userID {
				r.logger.Errorf("user %d already uploaded order %s", userID, number)
				return ErrOrderExists
			}

			r.logger.Errorf("user %d trying to upload order %s owned by another user", userID, number)
			return ErrOrderOwnedByAnotherUser
		}
		r.logger.Errorf("user %d failed to set order %s", userID, number)
		return fmt.Errorf("failed to set order %s: %w", number, err)
	}

	return nil
}

var (
	ErrNoOrders = errors.New("user has no orders")
)

func (r *OrdersRepository) GetOrders(ctx context.Context, userID int64) ([]models.Order, error) {
	r.logger.Infof("user %v trying to get his orders", userID)

	rows, err := r.db.QueryContext(ctx, getUserOrdersQuery, userID)
	if err != nil {
		r.logger.Errorf("query row context error %v", err)
		return nil, fmt.Errorf("failed to get orders for user %d: %w", userID, err)
	}
	defer rows.Close()

	var orders []models.Order

	for rows.Next() {
		var o models.Order
		err := rows.Scan(&o.Number, &o.Status, &o.Accrual, &o.UploadedAt)
		if err != nil {
			r.logger.Errorf("rows scan error %v", err)
			return nil, fmt.Errorf("failed to get orders for user %d: %w", userID, err)
		}

		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		r.logger.Errorf("rows.Err() returned the error %w", err)
		return nil, fmt.Errorf("failed to get orders for user %d: %w", userID, err)
	}

	if len(orders) == 0 {
		r.logger.Errorw("no orders found", "userID", userID)
		return nil, ErrNoOrders
	}

	return orders, nil
}
