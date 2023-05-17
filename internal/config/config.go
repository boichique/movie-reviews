package config

import (
	"fmt"

	"github.com/caarlos0/env/v8"
)

type Config struct {
	DBUrl string `env:"DB_URL"`
	Port  int    `env:"PORT" envDefault:"8080"`
}

func NewConfig() (*Config, error) {
	var c Config
	if err := env.Parse(&c); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &c, nil
}
