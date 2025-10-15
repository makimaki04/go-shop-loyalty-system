package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/makimaki04/go-shop-loyalty-system/internal/models"
	"go.uber.org/zap"
)

const (
	insertUserQuery = `
		INSERT INTO users (login, password_hash)
		VALUES ($1, $2)
		RETURNING id
	`
	selectUserQuery = `
		SELECT id, login, password_hash
		FROM users
		WHERE login = $1
	`
)

var (
	ErrUserExists = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
)

type AuthRepository struct {
	db *sql.DB
	logger *zap.SugaredLogger
}

func NewAuthRepository(db *sql.DB, logger *zap.SugaredLogger) *AuthRepository {
	return &AuthRepository{
		db: db,
		logger: logger,
	}
}

func (r *AuthRepository) CreateUser(user models.User) (int, error) {
	var id int
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := r.db.QueryRowContext(ctx, insertUserQuery, user.Login, user.PasswordHash).Scan(&id) 
	if err != nil {
		var pqErr *pgconn.PgError
		if errors.As(err, &pqErr) && pqErr.Code == pgerrcode.UniqueViolation {
				return 0, ErrUserExists
		}
		return 0, fmt.Errorf("failed to set user %q: %w", user.Login, err)
	}

	return id, nil
}

func (r *AuthRepository) LoginUser(login string) (models.User, error) {
	var user models.User 
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err :=r.db.QueryRowContext(ctx, selectUserQuery, login).Scan(&user.ID, &user.Login, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, fmt.Errorf("failed to get user %q: %w", login, err)
	}

	return user, nil
}