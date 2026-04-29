package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
)

// структура для хранения аргументов командной строки
type Config struct {
	RunAddr     string `env:"RUN_ADDRESS"`
	DatabaseURI string `env:"DATABASE_URI"`
	AccuralAddr string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	JWTSecret   string `env:"JWT_SECRET"`
}

// обработка аргументов командной строки
// и сохранение их значений в структуре
func ParseFlags() (Config, error) {
	var cfg Config

	flag.StringVar(&cfg.RunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&cfg.AccuralAddr, "r", "localhost:8080", "accrual system address")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "address to connect DataBase")
	flag.StringVar(&cfg.JWTSecret, "k", "", "JWT secret key")

	// парсинг переданные серверу аргументы командной строки в зарегистрированные переменные
	flag.Parse()

	if cfg.JWTSecret == "" {
		return cfg, fmt.Errorf("JWT_SECRET is required")
	}

	// парсим переменные окружения
	err := env.Parse(&cfg)
	if err != nil {
		return cfg, fmt.Errorf("can't parse environment variables: %w", err)
	}

	return cfg, nil
}
