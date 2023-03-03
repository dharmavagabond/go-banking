package main

import (
	"github.com/dharmavagabond/simple-bank/internal/api"
	db "github.com/dharmavagabond/simple-bank/internal/db/sqlc"
	"github.com/dharmavagabond/simple-bank/internal/http/grpc"
	"github.com/labstack/echo/v4"
)

func main() {
	logger := echo.New().Logger
	store := db.NewStore()
	logger.Fatal(rungRPCServer(store))
}

func rungRPCServer(store db.Store) error {
	var (
		server *grpc.Server
		err    error
	)

	if server, err = grpc.NewServer(store); err != nil {
		return err
	}

	return server.Start()
}

func runHttpServer(store db.Store) error {
	var (
		server *api.Server
		err    error
	)

	if server, err = api.NewServer(store); err != nil {
		return err
	}

	return server.Start()
}
