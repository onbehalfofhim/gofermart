package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockJWTValidator struct {
	userId string
	err    error
}

func (m *mockJWTValidator) ValidateToken(token string) (string, error) {
	if m.err != nil {
		return "", m.err
	}

	return m.userId, nil
}

func TestAuth(t *testing.T) {
	tests := []struct {
		name               string
		authHeader         string
		validator          JWTValidator
		expectedStatusCode int
		expectedBody       string
		expectNextCalled   bool
		expectedUserId     string
	}{
		{
			name:               "missing authorization header",
			authHeader:         "",
			validator:          &mockJWTValidator{},
			expectedStatusCode: http.StatusUnauthorized,
			expectedBody:       "authorization required\n",
			expectNextCalled:   false,
		},
		{
			name:               "invalid authorization prefix",
			authHeader:         "Token abc123",
			validator:          &mockJWTValidator{},
			expectedStatusCode: http.StatusUnauthorized,
			expectedBody:       "invalid authorization header\n",
			expectNextCalled:   false,
		},
		{
			name:               "empty token",
			authHeader:         "Bearer ",
			validator:          &mockJWTValidator{},
			expectedStatusCode: http.StatusUnauthorized,
			expectedBody:       "empty token\n",
			expectNextCalled:   false,
		},
		{
			name:       "invalid token",
			authHeader: "Bearer invalid-token",
			validator: &mockJWTValidator{
				err: errors.New("invalid token"),
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedBody:       "invalid token\n",
			expectNextCalled:   false,
		},
		{
			name:       "valid token",
			authHeader: "Bearer valid-token",
			validator: &mockJWTValidator{
				userId: "user-123",
			},
			expectedStatusCode: http.StatusOK,
			expectNextCalled:   true,
			expectedUserId:     "user-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false

			var receivedUserId string

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				userId, ok := GetUserId(r.Context())

				assert.True(t, ok)

				receivedUserId = userId

				w.WriteHeader(http.StatusOK)
			})

			handler := Auth(tt.validator)(nextHandler)

			req := httptest.NewRequest(http.MethodGet, "/", nil)

			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatusCode, rec.Code)

			assert.Equal(t, tt.expectNextCalled, nextCalled)

			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, rec.Body.String())
			}

			if tt.expectedUserId != "" {
				assert.Equal(t, tt.expectedUserId, receivedUserId)
			}
		})
	}
}

func TestGetUserId(t *testing.T) {
	t.Run("user id exists", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), UserIdKey, "user-123")

		userId, ok := GetUserId(ctx)

		assert.True(t, ok)
		assert.Equal(t, "user-123", userId)
	})

	t.Run("user id missing", func(t *testing.T) {
		userId, ok := GetUserId(context.Background())

		assert.False(t, ok)
		assert.Empty(t, userId)
	})
}
