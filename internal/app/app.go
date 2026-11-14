package app

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/s0meuserhere/antibrutforce/internal/config"
	"github.com/s0meuserhere/antibrutforce/internal/domain/limits"
	"github.com/s0meuserhere/antibrutforce/internal/infrastructure/postgres"
	"github.com/s0meuserhere/antibrutforce/internal/infrastructure/redisdb"
	bucketrepo "github.com/s0meuserhere/antibrutforce/internal/repository/bucket"
	netlistrepo "github.com/s0meuserhere/antibrutforce/internal/repository/netlist"
	"github.com/s0meuserhere/antibrutforce/internal/service/guard"
)

// App - сервис анти-брутфорса вместе с подключениями к хранилищам.
// Его одинаково собирают сервер и CLI.
type App struct {
	service  *guard.Service
	postgres *pgxpool.Pool
	redis    *redis.Client
}

// New подключается к PostgreSQL и Redis и собирает сервис.
func New(ctx context.Context, cfg config.App) (*App, error) {
	l, err := limits.New(cfg.Limits.Login, cfg.Limits.Password, cfg.Limits.IP)
	if err != nil {
		return nil, fmt.Errorf("limits: %w", err)
	}

	pool, err := postgres.Connect(ctx, cfg.Postgres.DSN())
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	client, err := redisdb.Connect(ctx, cfg.Redis.Addr(), cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		pool.Close()

		return nil, fmt.Errorf("connect redis: %w", err)
	}

	lists := netlistrepo.New(pool, client, cfg.NetlistCacheTTL)
	buckets := bucketrepo.New(client)

	return &App{
		service:  guard.New(lists, buckets, l, time.Now),
		postgres: pool,
		redis:    client,
	}, nil
}

// Service возвращает собранный сервис.
func (a *App) Service() *guard.Service {
	return a.service
}

// Close закрывает подключения к хранилищам.
func (a *App) Close() error {
	a.postgres.Close()

	if err := a.redis.Close(); err != nil {
		return fmt.Errorf("close redis: %w", err)
	}

	return nil
}
