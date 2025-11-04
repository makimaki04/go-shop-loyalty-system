package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/makimaki04/go-shop-loyalty-system/internal/dto"
	"go.uber.org/zap"
)

type Client struct {
	baseURL string
	logger *zap.SugaredLogger
	client  *resty.Client
}

func NewClient(baseURL string, logger *zap.SugaredLogger) *Client {
	client := resty.New().
		SetTimeout(5*time.Second)

	return &Client{
		baseURL: baseURL,
		logger: logger,
		client: client,
	}
}

type AccrualError struct {
	StatusCode int
	RetryAfter time.Duration
	Msg string
}

func (e AccrualError) IsRetryable() bool {
    return e.StatusCode == http.StatusTooManyRequests || 
           e.StatusCode == http.StatusInternalServerError
}

func(e AccrualError) Error() string {
	return e.Msg
}

func (c *Client) GetAccrual(ctx context.Context, orderNumber string) (dto.AccrualResponse, error) {
	var response dto.AccrualResponse

	r, err := c.client.R().
		SetContext(ctx).
		Get(fmt.Sprintf("%s/api/orders/%s", c.baseURL, orderNumber))
	if err != nil {
		c.logger.Warnw("failed to get order accrual status", "order", orderNumber, "error", err)
		return dto.AccrualResponse{}, fmt.Errorf("failed to get order %s accrual: %v", orderNumber, err)
	}

	switch r.StatusCode() {
	case http.StatusNoContent:
		return dto.AccrualResponse{}, fmt.Errorf("order %s not registered in accrual system", orderNumber)
	case http.StatusTooManyRequests:
		var duration time.Duration

		retry := r.Header().Get("Retry-After")
		if retry == "" {
			duration = 10*time.Second
		} else {
			c.logger.Debugw("retry header received", "retry_after", retry)
			sec, err := strconv.Atoi(retry)
			if err != nil {
				duration = 10*time.Second
			} else {
				duration = time.Duration(sec)*time.Second
			}
		}
		return dto.AccrualResponse{}, AccrualError{
			StatusCode: r.StatusCode(),
			RetryAfter: duration,
			Msg: fmt.Sprintf("accrual returned %d (%s)", r.StatusCode(), http.StatusText(r.StatusCode())),
		}
	case http.StatusInternalServerError:
		return dto.AccrualResponse{}, AccrualError{
			StatusCode: r.StatusCode(),
			RetryAfter: 10*time.Second,
			Msg: fmt.Sprintf("accrual returned %d (%s)", r.StatusCode(), http.StatusText(r.StatusCode())),
		}
	case http.StatusOK:
		//continue
	default:
		return dto.AccrualResponse{}, fmt.Errorf("unexpected status code: %d", r.StatusCode())
	}
	
	if err := json.Unmarshal(r.Body(), &response); err != nil {
		c.logger.Warnw("couldn't deserialize response body", "error", err)
		return dto.AccrualResponse{}, fmt.Errorf("couldn't deserialize response body: %v", err)
	}

	c.logger.Infow("accrual response received",
		"order", orderNumber,
		"status", response.Status,
		"accrual", response.Accrual,
	)

	return response, nil
}