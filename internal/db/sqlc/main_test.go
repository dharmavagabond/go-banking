package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/dharmavagabond/simple-bank/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

var testQueries *Queries

func TestMain(m *testing.M) {
	var (
		dbpool *pgxpool.Pool
		err    error
	)

	if dbpool, err = pgxpool.Connect(context.Background(), config.Postgres.DSN); err != nil {
		log.Fatal("[Err]: ", err)
	}

	testQueries = New(dbpool)

	os.Exit(m.Run())
}
