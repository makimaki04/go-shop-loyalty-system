package dto

import (
	"time"

	"github.com/makimaki04/go-shop-loyalty-system/internal/models"
)

type BalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

type WithdrawalsResponse struct {
	Order        string  `json:"order"`
	Sum          float64 `json:"sum"`
	ProcessedAt time.Time `json:"processed_at" time_format:"2006-01-02T15:04:05Z07:00"`
}

func ConvertWithdrawals(withdrawals []models.Withdrawals) []WithdrawalsResponse {
	response := make([]WithdrawalsResponse, 0, len(withdrawals))

	for _, w := range withdrawals {
		response = append(response, WithdrawalsResponse{
			Order: w.OrderNumber,
			Sum: w.Sum,
			ProcessedAt: w.ProcessedAt,
		})
	}

	return response
}