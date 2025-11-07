package workerpool

import (
	"context"
	"time"

	"github.com/makimaki04/go-shop-loyalty-system/internal/models"
	"github.com/makimaki04/go-shop-loyalty-system/internal/repository"
	"go.uber.org/zap"
)

type Scheduler struct {
	repo     repository.Orders
	poolCh   chan<- models.PendingOrder
	interval time.Duration
	logger   *zap.SugaredLogger
}

func NewScheduler(repo repository.Orders, poolCh chan<- models.PendingOrder, interval time.Duration, logger *zap.SugaredLogger) *Scheduler {
	return &Scheduler{
		repo:     repo,
		poolCh:   poolCh,
		interval: interval,
		logger:   logger,
	}
}

func (s *Scheduler) LoadOrders(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer func() {
		ticker.Stop()
		close(s.poolCh)
		s.logger.Info("scheduler stopped, channel closed")
	}()

	s.logger.Info("scheduler started, polling orders every", s.interval)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			dbCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			orders, err := s.repo.GetPendingorders(dbCtx)
			cancel()

			if err != nil {
				s.logger.Warnw("failed to get pending orders", "error", err)
				continue
			}

			for _, o := range orders {
				select {
				case s.poolCh <- o:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}
