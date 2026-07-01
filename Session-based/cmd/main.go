package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"sessions-based/config"
	"sessions-based/internal/adapters/migrations"
	"sessions-based/internal/adapters/postgres"
	redisdb "sessions-based/internal/adapters/redis"
	"sessions-based/internal/api"
	"sessions-based/internal/api/handlers"
	"syscall"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.NewLoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	SessionRDB, err := redisdb.NewRedis(&cfg.Redis, cfg.Redis.RDBSession)
	if err != nil {
		log.Fatal(err)
	}

	LimitterRDB, err := redisdb.NewRedis(&cfg.Redis, cfg.Redis.RDBRateLimiter)
	if err != nil {
		log.Fatal(err)
	}

	db, err := postgres.NewPool(&cfg.Postgres)
	if err != nil {
		log.Fatal(err)
	}

	migrations.RunMigrations(db)
	if err := migrations.RunMigrations(db); err != nil {
		log.Fatal(err)
	}
	srvs := handlers.NewServices(SessionRDB, LimitterRDB, db)
	h := handlers.NewHandlers(srvs)
	serv := api.NewServer(h.Auth, srvs.Middleware)

	go func() {
		if err := serv.Run(); err != nil {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	postgres.ClosePool(db)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	serv.Shutdown(shutdownCtx)
}
