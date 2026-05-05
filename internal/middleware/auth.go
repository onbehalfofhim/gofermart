package middleware

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const UserIDKey contextKey = "UserId"

type JWTValidator interface {
	ValidateToken(tokenString string) (string, error)
}

// middleware для валидации токена в запросе
func Auth(jwtValidator JWTValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "authorization required", http.StatusUnauthorized)
				return
			}

			const prefix = "Bearer "
			if !strings.HasPrefix(authHeader, prefix) {
				http.Error(w, "invalid authorization header", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, prefix)
			if tokenString == "" {
				http.Error(w, "empty token", http.StatusUnauthorized)
				return
			}

			userID, err := jwtValidator.ValidateToken(tokenString)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// получение идентификатора пользователя из токена
func GetUserID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(UserIDKey).(string)
	return id, ok

}
