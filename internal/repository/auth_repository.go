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

func (r *AuthRepository) CreateUser(ctx context.Context, user models.User) (int64, error) {
	var id int64

	r.logger.Infof("creating user %s", user.Login)

	err := r.db.QueryRowContext(ctx, insertUserQuery, user.Login, user.PasswordHash).Scan(&id) 
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				r.logger.Warnw("user already exists", "login", user.Login, "error", err)
				return 0, ErrUserExists
		}
		r.logger.Errorw("failed to create user", "login", user.Login, "error", err)
		return 0, fmt.Errorf("failed to create user %q: %w", user.Login, err)
	}

	r.logger.Infow("user created", "login", user.Login, "userID", id)
	return id, nil
}

func (r *AuthRepository) LoginUser(ctx context.Context, login string) (models.User, error) {
	var user models.User

	r.logger.Infof("user %s trying to login", login)

	err := r.db.QueryRowContext(ctx, selectUserQuery, login).Scan(&user.ID, &user.Login, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Warnw("user not found", "login", login, "error", err)
			return models.User{}, ErrUserNotFound
		}
		r.logger.Errorw("failed to get user", "login", login, "error", err)
		return models.User{}, fmt.Errorf("failed to get user %q: %w", login, err)
	}

	r.logger.Infow("user logged in", "login", user.Login, "userID", user.ID)
	return user, nil
}