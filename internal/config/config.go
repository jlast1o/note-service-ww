package config

import (
	"errors"
	"os"
)

type Config struct {
	DatabaseURL   string
	RedisAddr     string
	RedisPassword string
}

func Load() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, errors.New("DATABASE_URL not found")
	}

	redisAddr := os.Getenv("REDIS_ADDR")

	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")

	return &Config{
		DatabaseURL:   dbURL,
		RedisAddr:     redisAddr,
		RedisPassword: redisPassword,
	}, nil
}
