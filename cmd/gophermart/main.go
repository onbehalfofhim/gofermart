package main

import (
	"net/http"

	"github.com/onbehalfofhim/gofermart/internal/handler"
	"github.com/onbehalfofhim/gofermart/internal/repository/inmemory"
	"github.com/onbehalfofhim/gofermart/internal/service"
)

func main() {
	userRepo := inmemory.NewUsersMemStorage()

	userService := service.NewUserService(userRepo)
	handler := handler.NewHandler(userService)

	http.ListenAndServe("localhost:8080", handler.Route())
}
