package main

import (
	"fmt"
	"net/http"

	"github.com/onbehalfofhim/gofermart/internal/auth"
	"github.com/onbehalfofhim/gofermart/internal/config"
	"github.com/onbehalfofhim/gofermart/internal/handler"
	"github.com/onbehalfofhim/gofermart/internal/repository/inmemory"
	"github.com/onbehalfofhim/gofermart/internal/service"
)

func main() {
	// получение параметров конфигурации приложения
	cfg, err := config.ParseFlags()
	if err != nil {
		// logger.Error("Error in parse flags and variables", "error", error)
	}

	fmt.Printf("%s", cfg.RunAddr)
	fmt.Printf("%s", cfg.JWTSecret)

	// передаем в приложение параметры JWT
	jwt := auth.NewJWT(cfg.JWTSecret)

	userRepo := inmemory.NewUsersMemStorage()
	userService := service.NewUserService(userRepo, jwt)
	handler := handler.NewHandler(userService)

	http.ListenAndServe(cfg.RunAddr, handler.Route())
}
