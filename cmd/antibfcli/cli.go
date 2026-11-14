package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/s0meuserhere/antibrutforce/internal/app"
	"github.com/s0meuserhere/antibrutforce/internal/config"
	"github.com/s0meuserhere/antibrutforce/internal/domain/netlist"
	"github.com/s0meuserhere/antibrutforce/internal/service/guard"
)

// cli собирает тот же сервис, что и сервер, и вызывает его методы напрямую, без gRPC.
type cli struct {
	configPath string
}

func newCLI() *cli {
	return &cli{}
}

func (c *cli) root() *cobra.Command {
	root := &cobra.Command{
		Use:           "antibfcli",
		Short:         "Ручное администрирование анти-брутфорса",
		SilenceUsage:  true,
		SilenceErrors: true,

		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	}

	root.PersistentFlags().StringVarP(&c.configPath, "config", "c", "./configs/antibf.env",
		"файл с переменными окружения, необязателен")

	root.AddCommand(
		c.check(),
		c.reset(),
		c.list("blacklist", netlist.Blacklist),
		c.list("whitelist", netlist.Whitelist),
	)

	return root
}

// withService собирает сервис, вызывает fn со сроком REQUEST_TIMEOUT и закрывает подключения.
func (c *cli) withService(ctx context.Context, fn func(ctx context.Context, s *guard.Service) error) (err error) {
	cfg, err := config.LoadApp(c.configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	a, err := app.New(ctx, cfg)
	if err != nil {
		return fmt.Errorf("start app: %w", err)
	}

	defer func() {
		err = errors.Join(err, a.Close())
	}()

	ctx, cancel := context.WithTimeout(ctx, cfg.RequestTimeout)
	defer cancel()

	return fn(ctx, a.Service())
}
