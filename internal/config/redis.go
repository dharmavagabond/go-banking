package config

import (
	"github.com/cristalhq/aconfig"
	"github.com/rs/zerolog/log"
)

type RedisConfig = struct {
	Host string `env:"HOST"`
	Port string `env:"PORT"`
}

var Redis RedisConfig

func init() {
	configOptions := getDefaultConfig()
	configOptions.EnvPrefix = "REDIS"
	loader := aconfig.LoaderFor(&Redis, *configOptions)

	if err := loader.Load(); err != nil {
		log.Fatal().Err(err)
	}
}
