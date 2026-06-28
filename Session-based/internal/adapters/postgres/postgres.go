package postgres

import (
	"context"
	"fmt"
	"sessions-based/config"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	timeout = 5
)

func NewPool(cfg *config.PostgresConfig) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		cfg.DBUser,
		cfg.DBPass,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return pool, nil
}
func ClosePool(pool *pgxpool.Pool) {
	pool.Close()
}
