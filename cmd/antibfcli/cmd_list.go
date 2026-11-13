package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/s0meuserhere/antibrutforce/internal/domain/netlist"
	"github.com/s0meuserhere/antibrutforce/internal/service/guard"
)

// list возвращает команду с подкомандами add и remove для списка kind.
func (c *cli) list(name string, kind netlist.Kind) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: "Управлять подсетями в " + name,
	}

	add := &cobra.Command{
		Use:   "add <subnet>",
		Short: "Добавить подсеть, например 192.1.1.0/25",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.withService(cmd.Context(), func(ctx context.Context, s *guard.Service) error {
				if err := s.Add(ctx, kind, args[0]); err != nil {
					return fmt.Errorf("add to %s: %w", name, err)
				}

				return nil
			})
		},
	}

	remove := &cobra.Command{
		Use:   "remove <subnet>",
		Short: "Удалить подсеть",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.withService(cmd.Context(), func(ctx context.Context, s *guard.Service) error {
				if err := s.Remove(ctx, kind, args[0]); err != nil {
					return fmt.Errorf("remove from %s: %w", name, err)
				}

				return nil
			})
		},
	}

	cmd.AddCommand(add, remove)

	return cmd
}
