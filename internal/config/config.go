package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// структура для хранения аргументов командной строки
type Config struct {
	RunAddr     string `env:"RUN_ADDRESS" env-default:":8081"`
	DatabaseURI string `env:"DATABASE_URI" env-default:"postgresql://user:password@localhost:5437/loyalty?sslmode=disable"`
	AccuralAddr string `env:"ACCRUAL_SYSTEM_ADDRESS" env-default:"http://localhost:8080"`
	JWTSecret   string `env:"JWT_SECRET"`
}

// обработка аргументов командной строки
// и сохранение их значений в структуре
func ParseFlags() (Config, error) {
	var cfg Config

	// 1. .env → в окружение
	_ = godotenv.Load()

	// 2. env → в структуру
	if err := env.Parse(&cfg); err != nil {
		return cfg, fmt.Errorf("can't parse env: %w", err)
	}

	// 3. flags → поверх всего
	flag.StringVar(&cfg.RunAddr, "a", cfg.RunAddr, "address and port")
	flag.StringVar(&cfg.AccuralAddr, "r", cfg.AccuralAddr, "accrual addr")
	flag.StringVar(&cfg.DatabaseURI, "d", cfg.DatabaseURI, "db uri")
	flag.StringVar(&cfg.JWTSecret, "k", cfg.JWTSecret, "jwt secret")

	flag.Parse()

	// 4. валидация
	if cfg.JWTSecret == "" {
		return cfg, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}
