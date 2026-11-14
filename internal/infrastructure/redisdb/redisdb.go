package redisdb

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

// slogLogger передаёт внутренние сообщения go-redis в slog с уровнем warn.
type slogLogger struct {
	log *slog.Logger
}

// SetLogger направляет внутренние сообщения go-redis в log. Логгер go-redis
// общий для всех клиентов процесса, поэтому вызывается один раз при старте.
func SetLogger(log *slog.Logger) {
	redis.SetLogger(slogLogger{log: log})
}

func (l slogLogger) Printf(ctx context.Context, format string, args ...any) {
	l.log.WarnContext(ctx, fmt.Sprintf(format, args...), "component", "redis")
}

// Connect создаёт клиент Redis по адресу host:port и проверяет, что сервер отвечает.
func Connect(ctx context.Context, addr, password string, db int) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()

		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return client, nil
}
