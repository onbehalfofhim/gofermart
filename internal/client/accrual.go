package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// структура клиента для работы с внешней системой
type AccrualClient struct {
	baseURL string
	client  *http.Client
}

// конструктор для создания клиента
func NewAccrualClient(addr string) *AccrualClient {
	return &AccrualClient{
		baseURL: addr,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// модель ответа от внешней системы
type AccrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

// взаимодействие с системой расчёта начислений баллов лояльности
func (c *AccrualClient) GetOrder(ctx context.Context, orderNumber string) (*AccrualResponse, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.baseURL, orderNumber)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("can't create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("can't send a get-request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var result AccrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}

		return &result, nil

	case http.StatusNoContent:
		return nil, nil

	case http.StatusTooManyRequests:
		retryAfter := 60 * time.Second
		value := resp.Header.Get("Retry-After")

		if value != "" {
			seconds, err := strconv.Atoi(value)
			if err == nil {
				retryAfter = time.Duration(seconds) * time.Second
			}
		}

		return nil, &RetryAfterError{
			RetryAfter: retryAfter,
		}

	case http.StatusInternalServerError:
		return nil, fmt.Errorf("Internal system error")

	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
