package auth

import (
	// "errors"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// структура для управления JWT-токенами
type JWT struct {
	secret []byte
}

// конструктор для структуры упрвления токенами
func NewJWT(secret string) *JWT {
	return &JWT{
		secret: []byte(secret),
	}
}

// генерация JWT-токена для пользователя
func (j *JWT) GenerateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"UserId":    userID,
		"ExpiresAt": time.Now().Add(24 * time.Hour).Unix(),
		"IssuedAt":  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

// ValidateToken валидирует JWT токен и возвращает claims
func (j *JWT) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return j.secret, nil
	})
	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid claims")
	}

	// проверка exp
	exp, ok := claims["ExpiresAt"].(float64)
	if !ok || int64(exp) < time.Now().Unix() {
		return "", errors.New("token expired")
	}

	userID, ok := claims["UserId"].(string)
	if !ok {
		return "", errors.New("invalid user id")
	}

	return userID, nil
}
