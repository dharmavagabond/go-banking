package config

import (
	"log"
	"time"

	"github.com/cristalhq/aconfig"
)

type AppConfig = struct {
	Host                 string        `default:"localhost" env:"HOST"`
	HttpPort             int           `default:"8080" env:"HTTP_PORT"`
	GrpcPort             int           `default:"9090" env:"GRPC_PORT"`
	Env                  string        `default:"production" env:"ENV"`
	IsDev                bool          `default:"false"`
	Secret               string        `env:"SECRET"`
	TokenSymmetricKey    string        `env:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration  time.Duration `default:"15m" env:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration time.Duration `default:"24h" env:"REFRESH_TOKEN_DURATION"`
}

var App AppConfig

func init() {
	configOptions := getDefaultConfig()
	configOptions.EnvPrefix = "APP"
	loader := aconfig.LoaderFor(&App, *configOptions)

	if err := loader.Load(); err != nil {
		log.Fatal(err)
	}

	App.IsDev = App.Env == "development"
}
