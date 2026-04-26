package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL       string
	RedisAddr         string
	RedisPassword     string
	RateLimitRequests int
	RateLimitPeriod   int
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

	rateLimitRequestsStr := os.Getenv("RATE_LIMIT_REQUESTS")
	if rateLimitRequestsStr == "" {
		rateLimitRequestsStr = "100"
	}
	rateLimitRequests, err := strconv.Atoi(rateLimitRequestsStr)
	if err != nil {
		return nil, fmt.Errorf("RATE_LIMIT_REQUESTS не число: %w", err)
	}

	rateLimitPeriodStr := os.Getenv("RATE_LIMIT_PERIOD")
	if rateLimitPeriodStr == "" {
		rateLimitPeriodStr = "60"
	}
	rateLimitPeriod, err := strconv.Atoi(rateLimitPeriodStr)
	if err != nil {
		return nil, fmt.Errorf("RATE_LIMIT_PERIOD не число: %w", err)
	}

	return &Config{
		DatabaseURL:       dbURL,
		RedisAddr:         redisAddr,
		RedisPassword:     redisPassword,
		RateLimitRequests: rateLimitRequests,
		RateLimitPeriod:   rateLimitPeriod,
	}, nil
}
