package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/onbehalfofhim/gofermart/internal/auth"
	"github.com/onbehalfofhim/gofermart/internal/client"
	"github.com/onbehalfofhim/gofermart/internal/config"
	"github.com/onbehalfofhim/gofermart/internal/handler"
	"github.com/onbehalfofhim/gofermart/internal/logger"
	"github.com/onbehalfofhim/gofermart/internal/repository/postrges"
	"github.com/onbehalfofhim/gofermart/internal/service"
	"github.com/onbehalfofhim/gofermart/migrations"
)

func main() {
	// root ctx приложения
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// получение параметров конфигурации приложения
	cfg, err := config.ParseFlags()
	logger := logger.NewLogger()

	if err != nil {
		logger.Error("Error in parse flags and variables", "error", err)
	}

	if err := run(ctx, cfg, logger); err != nil {
		logger.Error("Error in run server", "error", err)
	}
}

func run(ctx context.Context, cfg config.Config, logger *logger.Logger) error {
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

	// repositories
	userRepo := postrges.NewUsersRepository(db)
	orderRepo := postrges.NewOrdersRepository(db)
	balanceRepo := postrges.NewBalanceRepository(db)

	// services
	userService := service.NewUserService(userRepo)
	orderService := service.NewOrderService(orderRepo)
	balanceService := service.NewBalanceService(balanceRepo)

	// external client
	accrualClient := client.NewAccrualClient(cfg.AccuralAddr)

	// processor
	accrualProcessor := service.NewAccrualProcessor(orderService, balanceService, accrualClient, logger)

	// стартуем processor
	go accrualProcessor.Start(ctx)

	handler := handler.NewHandler(userService, orderService, balanceService, logger, jwt)
	server := &http.Server{
		Addr:    cfg.RunAddr,
		Handler: handler.Route(logger, jwt),
	}

	// server errors
	serverErr := make(chan error, 1)
	go func() {
		logger.Info("starting server", "addr", cfg.RunAddr)
		serverErr <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		err := server.Shutdown(shutdownCtx)
		if err != nil {
			return fmt.Errorf("can't shutdown server: %w", err)
		}

		logger.Info("server stopped")
		return nil

	case err := <-serverErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return fmt.Errorf("server error: %w", err)
	}
}
