package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/makimaki04/go-shop-loyalty-system/internal/dto"
	"github.com/makimaki04/go-shop-loyalty-system/internal/repository"
	"github.com/makimaki04/go-shop-loyalty-system/internal/service"
	"go.uber.org/zap"
)

type Handler struct {
	service *service.Service
	logger  *zap.SugaredLogger
}

func NewHandler(service *service.Service, logger *zap.SugaredLogger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

type SignUpResponse struct {
	ID       int    `json:"id"`
	JWTtoken string `json:"jwt-token"`
}

func (h *Handler) SignUp(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest

	r.Body = http.MaxBytesReader(w, r.Body, 1 << 20)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "couldn't read request body", h.logger)
		return
	}
	defer r.Body.Close()

	login := strings.TrimSpace(req.Login)
	password := strings.TrimSpace(req.Password)
	if login == "" || password == "" {
		respondWithError(w, http.StatusBadRequest, "login and password are required", h.logger)
		return
	}
	if len(password) < 8 {
		respondWithError(w, http.StatusBadRequest, "password must be at least 8 characters", h.logger)
		return
	}

	userID, err := h.service.Authorization.CreateUser(login, password)
	if err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			respondWithError(w, http.StatusConflict, repository.ErrUserExists.Error(), h.logger)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "internal server error", h.logger)
		return
	}

	token, err := h.service.Authorization.GenerateToken(userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "something went wrong", h.logger)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Authorization", "Bearer " + token)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	encodeResponse(w, SignUpResponse{
		ID:       userID,
		JWTtoken: token,
	}, h.logger)
}

type loginResponse struct {
	User     userResponse `json:"user"`
	JWTtoken string       `json:"jwt-token"`
}

type userResponse struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
}

func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest

	r.Body = http.MaxBytesReader(w, r.Body, 1 << 20)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "couldn't read request body", h.logger)
		return
	}
	defer r.Body.Close()

	login := strings.TrimSpace(req.Login)
	password := strings.TrimSpace(req.Password)
	if login == "" || password == "" {
		respondWithError(w, http.StatusBadRequest, "login and password are required", h.logger)
		return
	}
	if len(password) < 8 {
		respondWithError(w, http.StatusBadRequest, "password must be at least 8 characters", h.logger)
		return
	}

	user, err := h.service.Authorization.LoginUser(login, password)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) || errors.Is(err, service.ErrInvalidCredentials) {
			respondWithError(w, http.StatusUnauthorized, "wrong login or password", h.logger)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "internal server error", h.logger)
		return
	}

	token, err := h.service.Authorization.GenerateToken(user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "something went wrong", h.logger)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Authorization", "Bearer " + token)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	encodeResponse(w, loginResponse{
		User: userResponse{
			ID:    user.ID,
			Login: user.Login,
		},
		JWTtoken: token,
	}, h.logger)
}

func respondWithError(w http.ResponseWriter, code int, message string, logger *zap.SugaredLogger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": message}); err != nil {
		logger.Errorw("failed to encode response", "err", err)
	}
}

func encodeResponse[T any](w http.ResponseWriter, resp T, logger *zap.SugaredLogger) {
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Errorw("failed to encode response", "err", err)
	}
}
