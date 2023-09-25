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
		dbconfig *pgxpool.Config
		dbpool   *pgxpool.Pool
		err      error
	)

	if dbconfig, err = pgxpool.ParseConfig(config.Postgres.DSN); err != nil {
		log.Fatal(err)
	}

	if dbpool, err = pgxpool.NewWithConfig(context.Background(), dbconfig); err != nil {
		log.Fatal(err)
	}

	testQueries = New(dbpool)

	os.Exit(m.Run())
}
