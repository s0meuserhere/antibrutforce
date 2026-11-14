package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/s0meuserhere/antibrutforce/internal/app"
	"github.com/s0meuserhere/antibrutforce/internal/config"
	"github.com/s0meuserhere/antibrutforce/internal/grpcserver"
	"github.com/s0meuserhere/antibrutforce/internal/infrastructure/redisdb"
	"github.com/s0meuserhere/antibrutforce/internal/logger"
)

func main() {
	configPath := flag.String("config", "./configs/antibf.env", "файл с переменными окружения, необязателен")

	flag.Parse()

	if err := run(*configPath); err != nil {
		fmt.Fprintf(os.Stderr, "antibf: %v\n", err)
		os.Exit(1)
	}
}

func run(configPath string) (err error) {
	cfg, err := config.LoadServer(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log := logger.New(cfg.Environment, cfg.Logger.Level)
	redisdb.SetLogger(log)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	a, err := app.New(ctx, cfg.App)
	if err != nil {
		return fmt.Errorf("start app: %w", err)
	}

	defer func() {
		err = errors.Join(err, a.Close())
	}()

	var listenConfig net.ListenConfig

	listener, err := listenConfig.Listen(ctx, "tcp", cfg.GRPC.Addr())
	if err != nil {
		return fmt.Errorf("listen grpc: %w", err)
	}

	server := grpcserver.New(a.Service(), log, cfg.RequestTimeout)

	// После сигнала сервер дожидается текущих запросов, и Serve возвращает nil.
	go func() {
		<-ctx.Done()
		server.GracefulStop()
	}()

	log.Info("antibf started", "env", cfg.Environment, "grpc", cfg.GRPC.Addr())

	if err := server.Serve(listener); err != nil {
		return fmt.Errorf("serve grpc: %w", err)
	}

	return nil
}
