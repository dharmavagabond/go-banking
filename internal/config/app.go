package config

import (
	"errors"
	"time"

	"github.com/cristalhq/aconfig"
	"github.com/rs/zerolog/log"
)

type AppConfig = struct {
	Host                 string        `default:"localhost"  env:"HOST"`
	Env                  string        `default:"production" env:"ENV"`
	Secret               string        `                     env:"SECRET"`
	TokenSymmetricKey    string        `                     env:"TOKEN_SYMMETRIC_KEY"`
	HTTPPort             int           `default:"0"          env:"HTTP_PORT"`
	GrpcPort             int           `default:"9090"       env:"GRPC_PORT"`
	AccessTokenDuration  time.Duration `default:"15m"        env:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration time.Duration `default:"24h"        env:"REFRESH_TOKEN_DURATION"`
	IsDev                bool          `default:"false"`
}

var App AppConfig

func init() {
	configOptions := getDefaultConfig()
	configOptions.EnvPrefix = "APP"
	loader := aconfig.LoaderFor(&App, *configOptions)

	if err := loader.Load(); err != nil {
		log.Fatal().Err(err)
	}

	if App.HTTPPort < 0 {
		log.Fatal().Err(errors.New("el puerto no puede ser negativo"))
	}

	App.IsDev = App.Env == "development"
}
