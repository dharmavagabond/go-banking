package main

import (
	"github.com/dharmavagabond/simple-bank/internal/api"
	"github.com/dharmavagabond/simple-bank/internal/config"
	"github.com/dharmavagabond/simple-bank/internal/db/sqlc"
	"github.com/labstack/echo/v4"
)

func main() {
	logger := echo.New().Logger
	store := db.NewStore()
	server := api.NewServer(store)
	logger.Fatal(server.Start(config.App.Port))
}
