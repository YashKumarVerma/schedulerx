package utils

import (
	"context"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	CacheURLScheme  string `env:"CACHE_URL_SCHEME" envDefault:"redis"`
	CacheClusterURL string `env:"CACHE_CLUSTER_URL" envDefault:"localhost"`
	CachePassword   string `env:"CACHE_PASSWORD" envDefault:""`
	CacheUsername   string `env:"CACHE_USERNAME" envDefault:""`
	CacheTLSDomain  string `env:"CACHE_TLS_DOMAIN" envDefault:""`
	PodID           string `env:"POD_ID" envDefault:""`
}

var appConfig *Config

// GetConfig creates a new Config struct.
func GetConfig(ctx context.Context) *Config {
	if appConfig != nil {
		return appConfig
	} else {
		err := godotenv.Load(".env")
		if err != nil {
			GetAppLogger(ctx).Warnf("Unable to load .env file. Continuing without loading it...")
		}
		appConfig = &Config{}
		if err = env.Parse(appConfig); err != nil {
			panic(err)
		}
		return appConfig
	}
}
