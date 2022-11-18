package config

import "github.com/cristalhq/aconfig"

func getDefaultConfig() *aconfig.Config {
	return &aconfig.Config{
		AllFieldRequired: true,
		SkipFiles:        true,
		SkipFlags:        true,
	}
}
