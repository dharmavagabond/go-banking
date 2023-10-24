package config

import (
	"fmt"

	"github.com/cristalhq/aconfig"
	"github.com/rs/zerolog/log"
)

type PostgresConfig = struct {
	User     string `env:"USER"`
	Password string `env:"PASSWORD"`
	DB       string `env:"DB"       default:"simple-bank"`
	Host     string `env:"HOST"     default:"localhost"`
	Port     int    `env:"PORT"     default:"5432"`
	DSN      string `               default:"dsn"`
	SSLMode  string `env:"SSLMODE"  default:"disable"`
}

var Postgres PostgresConfig

func init() {
	configOptions := getDefaultConfig()
	configOptions.EnvPrefix = "POSTGRES"
	loader := aconfig.LoaderFor(&Postgres, *configOptions)

	if err := loader.Load(); err != nil {
		log.Fatal().Err(err)
	}

	Postgres.DSN = fmt.Sprintf(
		"user=%s password=%s host=%s port=%d dbname=%s sslmode=%s",
		Postgres.User,
		Postgres.Password,
		Postgres.Host,
		Postgres.Port,
		Postgres.DB,
		Postgres.SSLMode,
	)
}
