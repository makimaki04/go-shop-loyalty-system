package workerpool

import (
	"context"
	"sync"
	"time"

	"github.com/makimaki04/go-shop-loyalty-system/internal/accrual"
	"github.com/makimaki04/go-shop-loyalty-system/internal/models"
	"github.com/makimaki04/go-shop-loyalty-system/internal/repository"
	"go.uber.org/zap"
)

type WorkerPool struct {
	client           *accrual.Client
	ordersRepo       repository.Orders
	balanceRepo      repository.Balance
	ch               chan models.PendingOrder
	shedulerInterval time.Duration
	wg               sync.WaitGroup
	logger           *zap.SugaredLogger
}

func NewWorkerPool(client *accrual.Client, ordersRepo repository.Orders, balanceRepo repository.Balance, shedulerInterval time.Duration, logger *zap.SugaredLogger) *WorkerPool {
	ch := make(chan (models.PendingOrder), 100)

	return &WorkerPool{
		client:           client,
		ordersRepo:       ordersRepo,
		balanceRepo:      balanceRepo,
		ch:               ch,
		shedulerInterval: shedulerInterval,
		logger:           logger,
	}
}

func (p *WorkerPool) Start(ctx context.Context, count int) {
	p.logger.Infof("starting worker pool with %d workers", count)
	defer p.logger.Info("worker pool stopped")

	sheduler := NewScheduler(p.ordersRepo, p.ch, p.shedulerInterval, p.logger)

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		sheduler.LoadOrders(ctx)
	}()

	for i := 0; i < int(count); i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			worker := NewWorker(p.client, p.ch, p.ordersRepo, p.balanceRepo, p.logger)
			worker.Run(ctx)
		}()
	}
}

func (p *WorkerPool) Stop(cancel context.CancelFunc) {
	cancel()
	p.wg.Wait()
	p.logger.Info("worker pool stopped gracefully")
}
