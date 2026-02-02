package config

import (
	"flag"
	"fmt"
	"log"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

type Config struct {
	Address     string `env:"RUN_ADDRESS"`
	DatabaseURI string `env:"DATABASE_URI"`
	AccrualURL  string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	JWTSecret 	string `env:"JWT_SECRET"`
}

func SetConfig() Config {
	var cfg Config

	_ = godotenv.Load(".env")

	if err := env.Parse(&cfg); err != nil {
		log.Fatal(fmt.Errorf(`couldn't parse env: %w`, err))
	}

	flag.StringVar(&cfg.Address, "a", cfg.Address, "Server address")
	flag.StringVar(&cfg.DatabaseURI, "d", cfg.DatabaseURI, "Database address")
	flag.StringVar(&cfg.AccrualURL, "r", cfg.AccrualURL, "Accrual system address")

	flag.Parse()

	return cfg
}

//WIN "postgres://go_shop_user:password@localhost:5432/gophermart?sslmode=disable"
//WSL "postgres://go_shop_user:password@172.31.80.1:5432/gophermart?sslmode=disable"
