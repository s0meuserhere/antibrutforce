package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/s0meuserhere/antibrutforce/internal/service/guard"
)

func (c *cli) check() *cobra.Command {
	var login, password, ip string

	cmd := &cobra.Command{
		Use:   "check",
		Short: "Проверить попытку авторизации так же, как сервер",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return c.withService(cmd.Context(), func(ctx context.Context, s *guard.Service) error {
				ok, err := s.Check(ctx, login, password, ip)
				if err != nil {
					return fmt.Errorf("check: %w", err)
				}

				cmd.Printf("ok=%t\n", ok)

				return nil
			})
		},
	}

	cmd.Flags().StringVar(&login, "login", "", "логин")
	cmd.Flags().StringVar(&password, "password", "", "пароль")
	cmd.Flags().StringVar(&ip, "ip", "", "IPv4-адрес")

	return cmd
}
