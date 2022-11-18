package config

import (
	"log"

	"github.com/cristalhq/aconfig"
)

type AppConfig = struct {
	Host  string `default:"localhost" env:"HOST"`
	Port  int    `default:"8080" env:"PORT"`
	Env   string `default:"production" env:"ENV"`
	IsDev bool   `default:"false"`
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
