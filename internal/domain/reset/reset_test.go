package reset_test

import (
	"errors"
	"net/netip"
	"testing"

	"github.com/s0meuserhere/antibrutforce/internal/domain/reset"
	"github.com/s0meuserhere/antibrutforce/internal/domain/subnet"
)

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		login   string
		ip      string
		wantIP  netip.Addr
		wantErr error
	}{
		{name: "login and ip", login: "user", ip: "192.1.1.10", wantIP: netip.MustParseAddr("192.1.1.10")},
		{name: "login only", login: "user"},
		{name: "ip only", ip: "192.1.1.10", wantIP: netip.MustParseAddr("192.1.1.10")},
		{name: "nothing", wantErr: reset.ErrEmpty},
		{name: "bad ip", login: "user", ip: "192.1.1", wantErr: subnet.ErrInvalidIP},
		{name: "IPv6", ip: "2001:db8::1", wantErr: subnet.ErrInvalidIP},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := reset.New(tt.login, tt.ip)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("New() error = %v, want %v", err, tt.wantErr)
			}

			if tt.wantErr != nil {
				return
			}

			if got.Login() != tt.login || got.IP() != tt.wantIP {
				t.Errorf("New() = %q/%v, want %q/%v", got.Login(), got.IP(), tt.login, tt.wantIP)
			}
		})
	}
}
