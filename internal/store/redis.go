package store

import (
	"context"
	"fmt"
	"service/internal/config"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

func NewRedisClient(ctx context.Context, cfg *config.Config) (*redis.Client, error) {
	rbd := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0,
	})

	if err := redisotel.InstrumentTracing(rbd); err != nil {
		return nil, fmt.Errorf("редиска не трейсится: %w", err)
	}

	if err := rbd.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("редиска не подключилась %w", err)
	}

	return rbd, nil
}
