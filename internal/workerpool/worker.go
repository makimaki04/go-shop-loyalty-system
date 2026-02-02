package workerpool

import (
	"context"
	"errors"
	"time"

	"github.com/makimaki04/go-shop-loyalty-system/internal/accrual"
	"github.com/makimaki04/go-shop-loyalty-system/internal/models"
	"github.com/makimaki04/go-shop-loyalty-system/internal/repository"
	"go.uber.org/zap"
)

type Worker struct {
	client      *accrual.Client
	poolChan    <-chan models.PendingOrder
	ordersRepo  repository.Orders
	balanceRepo repository.Balance
	logger      *zap.SugaredLogger
}

func NewWorker(client *accrual.Client, poolChan <-chan models.PendingOrder, ordersRepo repository.Orders, balanceRepo repository.Balance, logger *zap.SugaredLogger) *Worker {
	return &Worker{
		client:      client,
		poolChan:    poolChan,
		ordersRepo:  ordersRepo,
		balanceRepo: balanceRepo,
		logger:      logger,
	}
}

func (w *Worker) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case o, ok := <-w.poolChan:
			if !ok {
				w.logger.Info("pool channel closed, worker exiting")
				return
			}
			w.processOrder(ctx, o)
		}
	}
}

func (w *Worker) processOrder(ctx context.Context, order models.PendingOrder) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	start := time.Now()

	if order.Status == string(repository.OrderStatusProcessed) {
		w.logger.Debugw("order already processed, skipping", "order", order.Number)
		return
	}

	r, err := w.client.GetAccrual(ctx, order.Number)
	if err != nil {
		var aErr accrual.AccrualError
		if errors.As(err, &aErr) && aErr.IsRetryable() {
			if aErr.RetryAfter < time.Second {
				aErr.RetryAfter = time.Second
			}

			w.logger.Warnw("accrual unavailable, retrying later",
				"order", order.Number,
				"status_code", aErr.StatusCode,
				"retry_after", aErr.RetryAfter,
			)

			select {
			case <-time.After(aErr.RetryAfter):
			case <-ctx.Done():
				return
			}

			return
		}
		w.logger.Warnw("failed to get accrual", "order", order, "error", err)
		return
	}

	status := repository.OrderStatus(r.Status)
	if !status.IsValid() {
		w.logger.Warnw("invalid status received from accrual system", "status", r.Status, "order", order)
		return
	}

	if status == repository.OrderStatus(order.Status) {
		w.logger.Infow("nothing has changed", "order", order.Number, "status", order.Status)
		return
	}

	accrual := 0.0
	if r.Accrual != nil {
		accrual = *r.Accrual
	}

	if err := w.ordersRepo.UpdateOrderStatus(ctx, order.Number, status, accrual); err != nil {
		w.logger.Errorw("failed to update order status", "order", order, "error", err)
		return
	}

	if status == repository.OrderStatusProcessed {
		if err := w.balanceRepo.AddAccrual(ctx, order.Number, accrual); err != nil {
			w.logger.Errorw("failed to update user balance", "order", order, "error", err)
			return
		}
	}

	w.logger.Infow("order processed successfully",
		"order", order.Number,
		"status", r.Status,
		"duration", time.Since(start),
	)
}
