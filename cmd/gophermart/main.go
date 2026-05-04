package main

import (
	"database/sql"
	"fmt"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/onbehalfofhim/gofermart/internal/auth"
	"github.com/onbehalfofhim/gofermart/internal/config"
	"github.com/onbehalfofhim/gofermart/internal/handler"
	"github.com/onbehalfofhim/gofermart/internal/logger"
	"github.com/onbehalfofhim/gofermart/internal/repository/postrges"
	"github.com/onbehalfofhim/gofermart/internal/service"
	"github.com/onbehalfofhim/gofermart/migrations"
)

func main() {
	// получение параметров конфигурации приложения
	cfg, err := config.ParseFlags()
	logger := logger.NewLogger()

	if err != nil {
		logger.Error("Error in parse flags and variables", "error", err)
	}

	logger.Info("server run")

	if err := run(cfg, logger); err != nil {
		logger.Error("Error in run server", "error", err)
	}
}

func run(cfg config.Config, logger *logger.Logger) error {
	// передаем в приложение параметры JWT
	jwt := auth.NewJWT(cfg.JWTSecret)

	// подключение к БД
	db, err := sql.Open("pgx", cfg.DatabaseURI)
	if err != nil {
		logger.Error("Error connect to data base", "error", err)
		return fmt.Errorf("can't connect to DB: %w", err)
	}
	defer db.Close()

	// применение миграций
	if err := migrations.ApplyMigrations(db, "file://migrations"); err != nil {
		logger.Error("Error apply migrations", "error", err)
		return fmt.Errorf("can't apply migrations: %w", err)
	}

	userRepo := postrges.NewUsersRepository(db)
	orderRepo := postrges.NewOrdersRepository(db)
	balanceRepo := postrges.NewBalanceRepository(db)

	userService := service.NewUserService(userRepo)
	orderService := service.NewOrderService(orderRepo)
	balanceService := service.NewBalanceService(balanceRepo)

	handler := handler.NewHandler(userService, orderService, balanceService, logger, jwt)

	return http.ListenAndServe(cfg.RunAddr, handler.Route(logger, jwt))
}
