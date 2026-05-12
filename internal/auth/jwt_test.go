package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestJWT_GenerateToken(t *testing.T) {
	j := NewJWT("test-secret")

	userId := "user-123"

	token, err := j.GenerateToken(userId)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestJWT_ValidateToken(t *testing.T) {
	j := NewJWT("test-secret")

	validUserId := "user-123"

	validToken, err := j.GenerateToken(validUserId)
	assert.NoError(t, err)

	// expired token
	expiredClaims := jwt.MapClaims{
		"UserId":    validUserId,
		"ExpiresAt": time.Now().Add(-1 * time.Hour).Unix(),
		"IssuedAt":  time.Now().Add(-2 * time.Hour).Unix(),
	}

	expiredJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)

	expiredToken, err := expiredJWT.SignedString([]byte("test-secret"))
	assert.NoError(t, err)

	// token without user id
	noUserClaims := jwt.MapClaims{
		"ExpiresAt": time.Now().Add(1 * time.Hour).Unix(),
		"IssuedAt":  time.Now().Unix(),
	}

	noUserJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, noUserClaims)

	noUserToken, err := noUserJWT.SignedString([]byte("test-secret"))
	assert.NoError(t, err)

	tests := []struct {
		name        string
		token       string
		expectedId  string
		expectedErr error
	}{
		{
			name:        "valid token",
			token:       validToken,
			expectedId:  validUserId,
			expectedErr: nil,
		},
		{
			name:        "invalid token",
			token:       "invalid-token",
			expectedId:  "",
			expectedErr: errors.New("invalid token"),
		},
		{
			name:        "expired token",
			token:       expiredToken,
			expectedId:  "",
			expectedErr: errors.New("token expired"),
		},
		{
			name:        "token without user id",
			token:       noUserToken,
			expectedId:  "",
			expectedErr: errors.New("invalid user id"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			userId, err := j.ValidateToken(tt.token)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.EqualError(
					t,
					err,
					tt.expectedErr.Error(),
				)

				assert.Empty(t, userId)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedId, userId)
		})
	}
}
