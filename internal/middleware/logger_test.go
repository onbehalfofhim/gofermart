package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/onbehalfofhim/gofermart/internal/logger"
)

func TestRequestLogger(t *testing.T) {
	var buf bytes.Buffer

	handler := slog.NewJSONHandler(
		&buf,
		&slog.HandlerOptions{
			Level: slog.LevelInfo,
		},
	)

	log := logger.NewWithSlog(slog.New(handler))

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)

		_, _ = w.Write([]byte("response body"))
	})

	middlewareHandler := RequestLogger(log)(nextHandler)

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", nil)

	rec := httptest.NewRecorder()

	middlewareHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	logOutput := buf.String()

	assert.Contains(t, logOutput, `"msg":"request"`)
	assert.Contains(t, logOutput, `"uri":"/api/user/orders"`)
	assert.Contains(t, logOutput, `"method":"POST"`)
	assert.Contains(t, logOutput, `"status":201`)
	assert.Contains(t, logOutput, `"size":13`)
	assert.Contains(t, logOutput, `"duration"`)
}
