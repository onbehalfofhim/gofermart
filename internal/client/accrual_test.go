package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAccrualClient_GetOrder(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		responseBody  string
		retryAfter    string
		expectedResp  *AccrualResponse
		expectedErr   string
		expectedRetry time.Duration
	}{
		{
			name:       "200 OK",
			statusCode: http.StatusOK,
			responseBody: `{
				"order":"12345678903",
				"status":"PROCESSED",
				"accrual":500
			}`,
			expectedResp: &AccrualResponse{
				Order:   "12345678903",
				Status:  "PROCESSED",
				Accrual: 500,
			},
		},
		{
			name:         "204 No Content",
			statusCode:   http.StatusNoContent,
			expectedResp: nil,
		},
		{
			name:          "429 Too Many Requests",
			statusCode:    http.StatusTooManyRequests,
			retryAfter:    "120",
			expectedErr:   "retry after",
			expectedRetry: 120 * time.Second,
		},
		{
			name:        "500 Internal Server Error",
			statusCode:  http.StatusInternalServerError,
			expectedErr: "Internal system error",
		},
		{
			name:        "unexpected status",
			statusCode:  http.StatusBadRequest,
			expectedErr: "unexpected status code: 400",
		},
		{
			name:         "invalid json",
			statusCode:   http.StatusOK,
			responseBody: `invalid-json`,
			expectedErr:  "invalid character",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, http.MethodGet, r.Method)

					if tt.retryAfter != "" {
						w.Header().Set("Retry-After", tt.retryAfter)
					}

					w.WriteHeader(tt.statusCode)
					_, _ = w.Write([]byte(tt.responseBody))
				}),
			)

			defer server.Close()

			c := NewAccrualClient(server.URL)

			resp, err := c.GetOrder(context.Background(), "12345678903")

			if tt.expectedErr != "" {
				assert.Error(t, err)

				// RetryAfterError
				var retryErr *RetryAfterError

				if errors.As(err, &retryErr) {
					assert.Equal(t, tt.expectedRetry, retryErr.RetryAfter)
				} else {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}

				assert.Nil(t, resp)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResp, resp)
		})
	}
}
