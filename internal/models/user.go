package models

// import "time"

import "github.com/google/uuid"

// модель пользователя приложения
type User struct {
	ID           uuid.UUID `json:"id"`
	Login        string    `json:"login"`
	PasswordHash string    `json:"-"`
	// CreatedAt    time.Time `json:"created_at"`
	// UpdatedAt    time.Time `json:"updated_at"`
}

// модель запроса аутентификации
type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// модель ответа на аутентификацию пользователя
type AuthResponse struct {
	Token string `json:"token"`
}
