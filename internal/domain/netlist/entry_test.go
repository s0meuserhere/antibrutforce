package netlist_test

import (
	"errors"
	"testing"

	"github.com/s0meuserhere/antibrutforce/internal/domain/netlist"
	"github.com/s0meuserhere/antibrutforce/internal/domain/subnet"
)

func TestNewEntry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		kind       netlist.Kind
		cidr       string
		wantSubnet string
		wantErr    error
	}{
		{name: "whitelist", kind: netlist.Whitelist, cidr: "10.0.0.0/8", wantSubnet: "10.0.0.0/8"},
		{name: "blacklist host bits", kind: netlist.Blacklist, cidr: "192.1.1.10/25", wantSubnet: "192.1.1.0/25"},
		{name: "unknown kind", kind: netlist.Kind("grey"), cidr: "10.0.0.0/8", wantErr: netlist.ErrUnknownKind},
		{name: "bad subnet", kind: netlist.Whitelist, cidr: "10.0.0.0", wantErr: subnet.ErrInvalidSubnet},
		{name: "IPv6", kind: netlist.Blacklist, cidr: "2001:db8::/32", wantErr: subnet.ErrInvalidSubnet},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := netlist.NewEntry(tt.kind, tt.cidr)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("NewEntry() error = %v, want %v", err, tt.wantErr)
			}

			if tt.wantErr != nil {
				return
			}

			if got.Kind() != tt.kind || got.Subnet().String() != tt.wantSubnet {
				t.Errorf("NewEntry() = %s %s, want %s %s", got.Kind(), got.Subnet(), tt.kind, tt.wantSubnet)
			}
		})
	}
}
