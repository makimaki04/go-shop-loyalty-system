package dto

import (
	"time"

	"github.com/makimaki04/go-shop-loyalty-system/internal/models"
)

type OrderResponse struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual"`
	UploadedAt time.Time `json:"uploaded_at" time_format:"2006-01-02T15:04:05Z07:00"`
}

func ConvertOrders(orders []models.Order) []OrderResponse {
	response := make([]OrderResponse, 0, len(orders))

	for _, o := range orders {
		response = append(response, OrderResponse{
			Number:     o.Number,
			Status:     o.Status,
			Accrual:    o.Accrual,
			UploadedAt: o.UploadedAt,
		})
	}
	
	return response
}