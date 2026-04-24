package store

import (
	"context"
	"fmt"
	"service/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)

	if err != nil {
		return nil, fmt.Errorf("некорректный DatabaseURL: %s", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("Не удалось создать пул: %s", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("Не удалось подключиться к БД: %s", err)
	}

	return pool, nil
}

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
	CREATE TABLE IF NOT EXISTS notes (
		id SERIAL PRIMARY KEY,
		title TEXT NOT NULL,
		content TEXT not null default '',
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	`

	_, err := pool.Exec(ctx, query)

	if err != nil {
		return fmt.Errorf("миграция не удалась: %w", err)
	}

	return nil
}
