package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/makimaki04/go-shop-loyalty-system/internal/accrual"
	"github.com/makimaki04/go-shop-loyalty-system/internal/config"
	"github.com/makimaki04/go-shop-loyalty-system/internal/handler"
	"github.com/makimaki04/go-shop-loyalty-system/internal/logger"
	"github.com/makimaki04/go-shop-loyalty-system/internal/middleware"
	"github.com/makimaki04/go-shop-loyalty-system/internal/migrations"
	"github.com/makimaki04/go-shop-loyalty-system/internal/repository"
	"github.com/makimaki04/go-shop-loyalty-system/internal/service"
	"github.com/makimaki04/go-shop-loyalty-system/internal/worker_pool"
	"go.uber.org/zap"
)

func main() {
	cfg := config.SetConfig()
	logger, sugar, err := logger.NewLogger("./configs/logger.json")
	if err != nil {
		panic(fmt.Errorf("logger failed to initialize"))
	}
	defer logger.Sync()

	db, err := initDB(cfg.DatabaseURI, sugar)
	if err != nil {
		panic(fmt.Errorf("something went wrong during db initialization"))
	}
	repo := repository.NewRepository(db, sugar)
	service := service.NewService(repo, sugar)
	handler := handler.NewHandler(service, sugar)

	r := chi.NewRouter()
	r.Use(chiMiddleware.RequestID)
	r.Use(middleware.WithLogging(sugar))

	r.Route("/", func(r chi.Router) {
		r.Route("/api/user", func(r chi.Router) {
			r.Route("/register", func(r chi.Router) {
				r.Post("/", handler.SignUp)
			})
			r.Route("/login", func(r chi.Router) {
				r.Post("/", handler.LoginUser)
			})

			r.Group(func(r chi.Router) {
				r.Use(middleware.WithAuth(sugar))
				r.Route("/orders", func(r chi.Router) {
					r.Post("/", handler.PostOrder)
					r.Get("/", handler.GetOrders)
				})
				r.Route("/balance", func(r chi.Router) {
					r.Get("/", handler.GetBalance)
					r.Route("/withdraw", func(r chi.Router) {
						r.Post("/", handler.PostWithdraw)
					})
				})
				r.Route("/withdrawals", func(r chi.Router) {
					r.Get("/", handler.GetWithdrawals)
				})
			})
		})
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	accrualClient := accrual.NewClient(cfg.AccrualURL, sugar)
	pool := workerpool.NewWorkerPool(accrualClient, repo.Orders, repo.Balance, 5*time.Second, sugar)
	go pool.Start(ctx, 4)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		Addr:    cfg.Address,
		Handler: r,
	}

	go func() {
		<-sigCh
		sugar.Info("Received shutdown signal")

		// сначала останавливаем воркеров
		pool.Stop(cancel)

		// потом HTTP сервер
		ctxShutdown, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if err := srv.Shutdown(ctxShutdown); err != nil {
			sugar.Errorw("failed to gracefully shutdown http server", "error", err)
		} else {
			sugar.Info("HTTP server stopped gracefully")
		}
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		sugar.Fatalw("Server failed to start", "address", cfg.Address, "error", err)
	}
}

func initDB(dsn string, logger *zap.SugaredLogger) (*sql.DB, error) {
	logger.Infof("Using DSN: %s", dsn)
	if err := migrations.RunMigration(dsn); err != nil {
		logger.Fatal("Error when starting migrations: %v", zap.Error(err))
		return nil, err
	}
	logger.Info("Migration successfully started")

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.Fatal("Database connection error:" + err.Error())
		return nil, err
	}

	return db, nil
}
