package config

import (
	"errors"
	"os"
)

type Config struct {
	DatabaseURL string
}

func Load() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, errors.New("DATABASE_URL not found")
	}

	return &Config{
		DatabaseURL: dbURL,
	}, nil
}
