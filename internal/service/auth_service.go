package service

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/makimaki04/go-shop-loyalty-system/internal/models"
	"github.com/makimaki04/go-shop-loyalty-system/internal/repository"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo   repository.Authorization
	logger *zap.SugaredLogger
}

func NewAuthService(repo repository.Authorization, logger *zap.SugaredLogger) *AuthService {
	return &AuthService{
		repo:   repo,
		logger: logger,
	}
}

func (s *AuthService) CreateUser(ctx context.Context, login, password string) (int64, error) {
	passwordHash, err := generatePasswordHash(password)
	if err != nil {
		return 0, err
	}
	user := models.User{Login: login, PasswordHash: passwordHash}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return s.repo.CreateUser(ctx, user)
}

func generatePasswordHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

var ErrInvalidCredentials = errors.New("invalid login or password")

func (s *AuthService) LoginUser(ctx context.Context, login, password string) (models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	userDB, err := s.repo.LoginUser(ctx, login)
	if err != nil {
		return models.User{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userDB.PasswordHash), []byte(password)); err != nil {
		return models.User{}, ErrInvalidCredentials
	}

	return models.User{
		ID: userDB.ID,
		Login: userDB.Login,
	}, nil
}

type Claims struct {
	jwt.RegisteredClaims
	UserID int64 `json:"id"`
}

func (s *AuthService) GenerateToken(id int64) (accessToken string, err error) {
	now := time.Now()
	expirationTime := now.Add(15 * time.Minute)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt: jwt.NewNumericDate(now),
			Issuer: "gophermart",
		},
		UserID: id,
	})

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", errors.New("JWT_SECRET is not set")
	}

	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		return "",  err
	}

	return tokenStr, nil
}