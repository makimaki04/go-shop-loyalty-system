package middleware

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/makimaki04/go-shop-loyalty-system/internal/service"
	"go.uber.org/zap"
)

type contextKey string

const userIDKey contextKey = "userID"

func WithAuth(logger *zap.SugaredLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		authFn := func (w http.ResponseWriter, r *http.Request)  {

			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				logger.Info("Got invalid token")
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return 
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == "" {
				logger.Info("Got empty token")
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return 
			}
		
			claims := &service.Claims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, 
			func(t *jwt.Token) (interface{}, error) {
				return []byte(os.Getenv("JWT_SECRET")), nil
			})
			if err != nil {
				var verr *jwt.ValidationError
				if errors.As(verr, &verr) {
					switch {
					case errors.Is(err, jwt.ErrTokenExpired):
						logger.Infow("JWT Expired", "err", err)
						http.Error(w, "token expired", http.StatusUnauthorized)
						return 
					case errors.Is(err, jwt.ErrTokenNotValidYet):
						logger.Infow("JWT not valid yet", "err", err)
						http.Error(w, "token not valid yet", http.StatusUnauthorized)
						return
					case errors.Is(err, jwt.ErrTokenMalformed):
						logger.Infow("JWT malformed", "err", err)
						http.Error(w, "invalid token", http.StatusUnauthorized)
						return
					case errors.Is(err, jwt.ErrTokenSignatureInvalid):
						logger.Infow("JWT bad signature", "err", err)
						http.Error(w, "invalid token", http.StatusUnauthorized)
						return
					default:
						logger.Infow("JWT validation error", "err", err)
						http.Error(w, "invalid token", http.StatusUnauthorized)
						return
					}
				}

				logger.Infow("JWT parse error", "err", err)
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			//По идее err уже not nill и данный блок не должен срабатывать
			if !token.Valid {
				logger.Info("JWT not valid (no explicit reason)")
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)		
		}

		return http.HandlerFunc(authFn)
	}
}

func GetUserID(r *http.Request) (int, bool) {
	id, ok := r.Context().Value(userIDKey).(int)
	return id, ok
}