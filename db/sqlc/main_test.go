package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	dbpool      *pgxpool.Pool
	testQueries *Queries
)

func TestMain(m *testing.M) {
	var (
		dbDsn string
		err   error
		isOk  bool
	)

	if dbDsn, isOk = os.LookupEnv("POSTGRES_DSN"); !isOk {
		log.Fatal("[Err]: `POSTGRES_DSN` no está ajustada.")
	}

	if dbpool, err = pgxpool.Connect(context.Background(), dbDsn); err != nil {
		log.Fatal("[Err]: ", err)
	}

	testQueries = New(dbpool)

	os.Exit(m.Run())
}
