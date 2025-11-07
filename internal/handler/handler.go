package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/makimaki04/go-shop-loyalty-system/internal/dto"
	"github.com/makimaki04/go-shop-loyalty-system/internal/middleware"
	"github.com/makimaki04/go-shop-loyalty-system/internal/models"
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

const maxBodySize = 1 << 20

func parseJSONBody[T any](w http.ResponseWriter, r *http.Request) (T, error) {
	var req T

	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		if errors.Is(err, io.EOF) {
			return req, errors.New("empty request body")
		}
		return req, err
	}

	return req, nil
}

func (h *Handler) SignUp(w http.ResponseWriter, r *http.Request) {
	req, err := parseJSONBody[dto.RegisterRequest](w, r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), h.logger)
		return
	}

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

	ctx := r.Context()

	userID, err := h.service.Authorization.CreateUser(ctx, login, password)
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
	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	encodeResponse(w, dto.SignUpResponse{
		ID:       userID,
		JWTToken: token,
	}, h.logger)
}

func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	req, err := parseJSONBody[dto.LoginRequest](w, r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), h.logger)
		return
	}

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

	ctx := r.Context()

	user, err := h.service.Authorization.LoginUser(ctx, login, password)
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
	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	encodeResponse(w, dto.LoginResponse{
		User: dto.UserResponse{
			ID:    user.ID,
			Login: user.Login,
		},
		JWTToken: token,
	}, h.logger)
}

func (h *Handler) PostOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "unauthorized", h.logger)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "couldn't read request body", h.logger)
		return
	}
	number := strings.TrimSpace(string(body))

	ctx := r.Context()

	if err := h.service.Orders.LoadOrder(ctx, userID, number); err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidOrderNumber):
			respondWithError(w, http.StatusUnprocessableEntity, "invalid order number", h.logger)
			return
		case errors.Is(err, repository.ErrOrderExists):
			w.WriteHeader(http.StatusOK)
			return
		case errors.Is(err, repository.ErrOrderOwnedByAnotherUser):
			respondWithError(w, http.StatusConflict, "order owned by another user", h.logger)
			return
		default:
			respondWithError(w, http.StatusInternalServerError, "internal server error", h.logger)
			return
		}
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "unauthorized", h.logger)
		return
	}

	ctx := r.Context()

	orders, err := h.service.Orders.GetOrders(ctx, userID)
	if err != nil {
		if err == repository.ErrNoOrders {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		respondWithError(w, http.StatusInternalServerError, "internal server error", h.logger)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	encodeResponse(w, dto.ConvertOrders(orders), h.logger)
}

func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "unauthorized", h.logger)
		return
	}

	ctx := r.Context()

	balance, err := h.service.Balance.GetBalance(ctx, userID)
	if err != nil {
		h.logger.Errorw("failed to get user balance", "userID", userID, "error", err)
		respondWithError(w, http.StatusInternalServerError, "internal server error", h.logger)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	encodeResponse(w, dto.BalanceResponse{
		Current:   balance.Current,
		Withdrawn: balance.Withdrawn,
	}, h.logger)
}

func (h *Handler) PostWithdraw(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "unauthorized", h.logger)
		return
	}

	req, err := parseJSONBody[dto.WithdrawRequest](w, r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), h.logger)
		return
	}

	if req.Order == "" {
		respondWithError(w, http.StatusUnprocessableEntity, "order number required", h.logger)
		return
	}

	withdraw := models.Withdraw{
		UserID: userID,
		Order:  req.Order,
		Sum:    req.Sum,
	}

	ctx := r.Context()

	err = h.service.Balance.WithdrawBonuses(ctx, withdraw)
	switch {
	case errors.Is(err, service.ErrInvalidOrderNumber):
		respondWithError(w, http.StatusUnprocessableEntity, "invalid order number", h.logger)
		return
	case errors.Is(err, service.ErrInsufficientFunds):
		respondWithError(w, http.StatusPaymentRequired, "insufficient funds to withdraw", h.logger)
		return
	case err != nil:
		h.logger.Errorw("failed to process withdraw", "userID", userID, "error", err)
		respondWithError(w, http.StatusInternalServerError, "internal server error", h.logger)
		return
	}

	h.logger.Infow("withdraw request processed", "userID", userID, "sum", req.Sum)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "unauthorized", h.logger)
		return
	}

	ctx := r.Context()

	withdrawals, err := h.service.Balance.GetWithdrawals(ctx, userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "internal server error", h.logger)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	encodeResponse(w, dto.ConvertWithdrawals(withdrawals), h.logger)
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
