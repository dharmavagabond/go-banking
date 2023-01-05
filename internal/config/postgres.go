package config

import (
	"fmt"
	"log"

	"github.com/cristalhq/aconfig"
)

type PostgresConfig = struct {
	User     string `env:"USER"`
	Password string `env:"PASSWORD"`
	Db       string `default:"simple-bank" env:"DB"`
	Host     string `default:"localhost" env:"HOST"`
	Port     int    `default:"5432" env:"PORT"`
	DSN      string `default:"dsn"`
	SSLMode  string `default:"disable" env:"SSLMODE"`
}

var Postgres PostgresConfig

func init() {
	configOptions := getDefaultConfig()
	configOptions.EnvPrefix = "POSTGRES"
	loader := aconfig.LoaderFor(&Postgres, *configOptions)

	if err := loader.Load(); err != nil {
		log.Fatal(err)
	}

	Postgres.DSN = fmt.Sprintf(
		"user=%s password=%s host=%s port=%d dbname=%s sslmode=%s",
		Postgres.User,
		Postgres.Password,
		Postgres.Host,
		Postgres.Port,
		Postgres.Db,
		Postgres.SSLMode,
	)
}
