package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/s0meuserhere/antibrutforce/internal/app"
	"github.com/s0meuserhere/antibrutforce/internal/config"
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

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	a, err := app.New(ctx, cfg.App)
	if err != nil {
		return fmt.Errorf("start app: %w", err)
	}

	defer func() {
		err = errors.Join(err, a.Close())
	}()

	<-ctx.Done()

	return nil
}
