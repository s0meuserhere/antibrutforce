package subnet_test

import (
	"errors"
	"net/netip"
	"testing"

	"github.com/s0meuserhere/antibrutforce/internal/domain/subnet"
)

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cidr    string
		want    string
		wantErr error
	}{
		{name: "host bits are cleared", cidr: "192.1.1.10/25", want: "192.1.1.0/25"},
		{name: "network address", cidr: "192.1.1.0/25", want: "192.1.1.0/25"},
		{name: "single host", cidr: "10.0.0.1/32", want: "10.0.0.1/32"},
		{name: "whole space", cidr: "1.2.3.4/0", want: "0.0.0.0/0"},
		{name: "empty", cidr: "", wantErr: subnet.ErrInvalidSubnet},
		{name: "no mask", cidr: "192.1.1.0", wantErr: subnet.ErrInvalidSubnet},
		{name: "mask too long", cidr: "192.1.1.0/33", wantErr: subnet.ErrInvalidSubnet},
		{name: "octet out of range", cidr: "300.1.1.0/24", wantErr: subnet.ErrInvalidSubnet},
		{name: "IPv6", cidr: "2001:db8::/32", wantErr: subnet.ErrInvalidSubnet},
		{name: "IPv4-mapped IPv6", cidr: "::ffff:192.1.1.0/120", wantErr: subnet.ErrInvalidSubnet},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := subnet.New(tt.cidr)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("New(%q) error = %v, want %v", tt.cidr, err, tt.wantErr)
			}

			if tt.wantErr != nil {
				return
			}

			if got.String() != tt.want {
				t.Errorf("New(%q) = %s, want %s", tt.cidr, got, tt.want)
			}
		})
	}
}

func TestContains(t *testing.T) {
	t.Parallel()

	s, err := subnet.New("192.1.1.0/25")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	tests := []struct {
		ip   string
		want bool
	}{
		{ip: "192.1.1.0", want: true},
		{ip: "192.1.1.127", want: true},
		{ip: "192.1.1.128", want: false},
		{ip: "192.1.2.1", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			t.Parallel()

			if got := s.Contains(netip.MustParseAddr(tt.ip)); got != tt.want {
				t.Errorf("Contains(%s) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

func TestRestore(t *testing.T) {
	t.Parallel()

	prefix := netip.MustParsePrefix("10.0.0.0/8")
	if got := subnet.Restore(prefix).Prefix(); got != prefix {
		t.Errorf("Restore().Prefix() = %s, want %s", got, prefix)
	}
}

func TestParseIP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ip      string
		wantErr error
	}{
		{name: "IPv4", ip: "192.1.1.10"},
		{name: "empty", ip: "", wantErr: subnet.ErrInvalidIP},
		{name: "with mask", ip: "192.1.1.10/25", wantErr: subnet.ErrInvalidIP},
		{name: "IPv6", ip: "2001:db8::1", wantErr: subnet.ErrInvalidIP},
		{name: "IPv4-mapped IPv6", ip: "::ffff:192.1.1.10", wantErr: subnet.ErrInvalidIP},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := subnet.ParseIP(tt.ip)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ParseIP(%q) error = %v, want %v", tt.ip, err, tt.wantErr)
			}

			if tt.wantErr == nil && got.String() != tt.ip {
				t.Errorf("ParseIP(%q) = %s", tt.ip, got)
			}
		})
	}
}
