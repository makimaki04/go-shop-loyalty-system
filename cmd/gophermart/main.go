package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/makimaki04/go-shop-loyalty-system/internal/config"
	"github.com/makimaki04/go-shop-loyalty-system/internal/handler"
	"github.com/makimaki04/go-shop-loyalty-system/internal/logger"
	"github.com/makimaki04/go-shop-loyalty-system/internal/middleware"
	"github.com/makimaki04/go-shop-loyalty-system/internal/migrations"
	"github.com/makimaki04/go-shop-loyalty-system/internal/repository"
	"github.com/makimaki04/go-shop-loyalty-system/internal/service"
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
			r.Route("/orders", func(r chi.Router) {

			})
			r.Route("/balance", func(r chi.Router) {

				r.Route("/withdraw", func(r chi.Router) {

				})
			})
			r.Route("/withdraw", func(r chi.Router) {

			})
		})
	})

	if err := http.ListenAndServe(cfg.Address, r); err != nil {
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
