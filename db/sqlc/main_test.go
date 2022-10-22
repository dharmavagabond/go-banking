package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v4"
)

var testQueries *Queries

func TestMain(m *testing.M) {
	var (
		conn  *pgx.Conn
		err   error
		dbDsn string
		isOk  bool
	)

	if dbDsn, isOk = os.LookupEnv("POSTGRES_DSN"); !isOk {
		log.Fatal("[Err]: `POSTGRES_DSN` no está ajustada.")
	}

	if conn, err = pgx.Connect(context.Background(), dbDsn); err != nil {
		log.Fatal("[Err]: ", err)
	}

	testQueries = New(conn)

	os.Exit(m.Run())
}
