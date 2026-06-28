package config

import "github.com/caarlos0/env"

type Config struct {
	Postgres PostgresConfig
	Redis    RedisConfig
	App      AppConfig
}

type PostgresConfig struct {
	DBUser string `env:"DB_USER,required"`
	DBPass string `env:"DB_PASSWORD,required"`
	DBHost string `env:"DB_HOST,required"`
	DBPort string `env:"DB_PORT,required"`
	DBName string `env:"DB_NAME,required"`
}

type RedisConfig struct {
	RDBAddr        string `env:"RDB_ADDR,required"`
	RDBPort        string `env:"RDB_PORT,required"`
	RDBPass        string `env:"RDB_PASS,required"`
	RDBSession     int    `env:"RDB_SESSIONS_DB,required"`
	RDBRateLimiter int    `env:"RDB_RATE_LIMITER_DB,required"`
}

type AppConfig struct {
	Port string `env:"PORT,required"`
}

func NewLoadConfig() (*Config, error) {
	var cfg Config

	if err := env.Parse(&cfg.Postgres); err != nil {
		return nil, err
	}
	if err := env.Parse(&cfg.App); err != nil {
		return nil, err
	}
	if err := env.Parse(&cfg.Redis); err != nil {
		return nil, err
	}

	return &cfg, nil
}
