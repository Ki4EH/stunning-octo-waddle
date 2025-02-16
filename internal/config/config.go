package config

import (
	"fmt"
	"github.com/caarlos0/env/v11"
)

type Config struct {
	DatabasePort     string `env:"DATABASE_PORT" envDefault:"5432"`
	DatabaseUser     string `env:"DATABASE_USER,required"`
	DatabasePassword string `env:"DATABASE_PASSWORD,required"`
	DatabaseName     string `env:"DATABASE_NAME,required"`
	DatabaseHost     string `env:"DATABASE_HOST,required"`
	ServerPort       string `env:"SERVER_PORT" envDefault:"8080"`
	Environment      string `env:"ENVIRONMENT"`
}

func LoadConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return cfg, nil
}
