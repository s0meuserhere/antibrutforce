package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/s0meuserhere/antibrutforce/internal/service/guard"
)

func (c *cli) reset() *cobra.Command {
	var login, ip string

	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Сбросить бакеты логина и IP, нужно хотя бы одно из них",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return c.withService(cmd.Context(), func(ctx context.Context, s *guard.Service) error {
				if err := s.Reset(ctx, login, ip); err != nil {
					return fmt.Errorf("reset: %w", err)
				}

				return nil
			})
		},
	}

	cmd.Flags().StringVar(&login, "login", "", "логин")
	cmd.Flags().StringVar(&ip, "ip", "", "IPv4-адрес")

	return cmd
}
