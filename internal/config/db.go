package config

import (
	"log"

	"github.com/cristalhq/aconfig"
)

type DBConfig = struct {
	DSN string `env:"DSN"`
}

var DB DBConfig

func init() {
	configOptions := getDefaultConfig()
	configOptions.EnvPrefix = "DB"
	loader := aconfig.LoaderFor(&DB, *configOptions)

	if err := loader.Load(); err != nil {
		log.Fatal(err)
	}
}
