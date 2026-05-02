package main

import (
	"database/sql"
	"fmt"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/onbehalfofhim/gofermart/internal/auth"
	"github.com/onbehalfofhim/gofermart/internal/config"
	"github.com/onbehalfofhim/gofermart/internal/handler"
	"github.com/onbehalfofhim/gofermart/internal/repository"
	"github.com/onbehalfofhim/gofermart/internal/repository/postrges"
	"github.com/onbehalfofhim/gofermart/internal/service"
	"github.com/onbehalfofhim/gofermart/migrations"
)

func main() {
	// получение параметров конфигурации приложения
	cfg, err := config.ParseFlags()
	if err != nil {
		// logger.Error("Error in parse flags and variables", "error", error)
	}

	if err := run(cfg); err != nil {
		// logger.Error("Error in run server", "error", err)
		fmt.Printf("Error: %w", err)
	}

}

func run(cfg config.Config) error {
	// передаем в приложение параметры JWT
	jwt := auth.NewJWT(cfg.JWTSecret)

	var userRepo repository.UserRepo

	db, err := sql.Open("pgx", cfg.DatabaseURI)
	if err != nil {
		// logger.Error("Error connect to data base", "error", err)
		return fmt.Errorf("can't connect to DB: %w", err)
	}
	defer db.Close()

	if err := migrations.ApplyMigrations(db, "file://migrations"); err != nil {
		// logger.Error("Error apply migrations", "error", err)
		return fmt.Errorf("can't apply migrations: %w", err)
	}

	userRepo = postrges.NewUserRepository(db)
	userService := service.NewUserService(userRepo, jwt)
	handler := handler.NewHandler(userService)

	return http.ListenAndServe(cfg.RunAddr, handler.Route())
}
