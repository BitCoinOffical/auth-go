package redisdb

import (
	"context"
	"fmt"
	"sessions-based/config"

	"github.com/redis/go-redis/v9"
)

func NewRedis(cfg *config.RedisConfig, dbNum int) (*redis.Client, error) {
	addr := fmt.Sprintf("%s:%s", cfg.RDBAddr, cfg.RDBPort)
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.RDBPass,
		DB:       dbNum,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return client, nil
}
