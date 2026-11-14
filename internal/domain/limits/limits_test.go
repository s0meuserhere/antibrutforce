package limits_test

import (
	"errors"
	"testing"

	"github.com/s0meuserhere/antibrutforce/internal/domain/limits"
)

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		login    int
		password int
		ip       int
		wantErr  error
	}{
		{name: "defaults", login: 10, password: 100, ip: 1000},
		{name: "lower bound", login: 1, password: 1, ip: 1},
		{name: "upper bound", login: limits.Max, password: limits.Max, ip: limits.Max},
		{name: "zero login", login: 0, password: 100, ip: 1000, wantErr: limits.ErrInvalidLimit},
		{name: "negative password", login: 10, password: -1, ip: 1000, wantErr: limits.ErrInvalidLimit},
		{name: "ip above max", login: 10, password: 100, ip: limits.Max + 1, wantErr: limits.ErrInvalidLimit},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := limits.New(tt.login, tt.password, tt.ip)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("New() error = %v, want %v", err, tt.wantErr)
			}

			if tt.wantErr != nil {
				return
			}

			if got.PerLogin() != tt.login || got.PerPassword() != tt.password || got.PerIP() != tt.ip {
				t.Errorf("New() = %d/%d/%d, want %d/%d/%d",
					got.PerLogin(), got.PerPassword(), got.PerIP(), tt.login, tt.password, tt.ip)
			}
		})
	}
}
