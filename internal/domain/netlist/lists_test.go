package netlist_test

import (
	"net/netip"
	"testing"

	"github.com/s0meuserhere/antibrutforce/internal/domain/netlist"
	"github.com/s0meuserhere/antibrutforce/internal/domain/subnet"
)

func subnets(t *testing.T, cidrs ...string) []subnet.Subnet {
	t.Helper()

	result := make([]subnet.Subnet, 0, len(cidrs))
	for _, cidr := range cidrs {
		s, err := subnet.New(cidr)
		if err != nil {
			t.Fatalf("subnet.New(%q) error = %v", cidr, err)
		}

		result = append(result, s)
	}

	return result
}

func TestListsAllow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		ip         string
		whitelist  []string
		blacklist  []string
		wantOK     bool
		wantListed bool
	}{
		{name: "empty lists", ip: "192.1.1.10"},
		{
			name: "outside both lists", ip: "192.1.1.200",
			whitelist: []string{"192.1.1.0/25"}, blacklist: []string{"10.0.0.0/8"},
		},
		{
			name: "in whitelist", ip: "192.1.1.10",
			whitelist: []string{"192.1.1.0/25"},
			wantOK:    true, wantListed: true,
		},
		{
			name: "in blacklist", ip: "10.1.2.3",
			blacklist:  []string{"10.0.0.0/8"},
			wantListed: true,
		},
		{
			name: "in second subnet of blacklist", ip: "172.16.0.1",
			blacklist:  []string{"10.0.0.0/8", "172.16.0.0/12"},
			wantListed: true,
		},
		{
			name: "whitelist wins over blacklist", ip: "192.1.1.10",
			whitelist: []string{"192.1.1.0/25"}, blacklist: []string{"192.1.0.0/16"},
			wantOK: true, wantListed: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			lists := netlist.New(subnets(t, tt.whitelist...), subnets(t, tt.blacklist...))

			ok, listed := lists.Allow(netip.MustParseAddr(tt.ip))
			if ok != tt.wantOK || listed != tt.wantListed {
				t.Errorf("Allow(%s) = %v, %v, want %v, %v", tt.ip, ok, listed, tt.wantOK, tt.wantListed)
			}
		})
	}
}
