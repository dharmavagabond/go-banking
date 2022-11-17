package main

import (
	"context"
	"os"

	"github.com/dharmavagabond/simple-bank/api"
	"github.com/dharmavagabond/simple-bank/db/sqlc"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/echo/v4"
)

func main() {
	var (
		dbDsn         string
		err           error
		isOk          bool
		dbpool        *pgxpool.Pool
		serverAddress = "0.0.0.0:8080"
	)

	logger := echo.New().Logger

	if dbDsn, isOk = os.LookupEnv("POSTGRES_GOAPP_DSN"); !isOk {
		logger.Fatal("[Err]: `POSTGRES_GOAPP_DSN` no está ajustada.")
	}

	if dbpool, err = pgxpool.Connect(context.Background(), dbDsn); err != nil {
		logger.Fatal("[Err]: ", err)
	}

	store := db.NewStore(dbpool)
	server := api.NewServer(store)
	logger.Fatal(server.Start(serverAddress))
}
