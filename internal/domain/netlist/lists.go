package netlist

import (
	"net/netip"

	"github.com/s0meuserhere/antibrutforce/internal/domain/subnet"
)

// Lists - whitelist и blacklist вместе: по ним попытка решается без бакетов.
type Lists struct {
	whitelist []subnet.Subnet
	blacklist []subnet.Subnet
}

// New возвращает пару списков.
func New(whitelist, blacklist []subnet.Subnet) Lists {
	return Lists{whitelist: whitelist, blacklist: blacklist}
}

// Allow решает попытку по спискам. Адрес из whitelist разрешается,
// из blacklist отклоняется, whitelist важнее: адрес из обоих списков разрешается.
// listed = false, если адреса нет в списках и попытку решают бакеты.
func (l Lists) Allow(ip netip.Addr) (ok, listed bool) {
	if contains(l.whitelist, ip) {
		return true, true
	}

	if contains(l.blacklist, ip) {
		return false, true
	}

	return false, false
}

func contains(subnets []subnet.Subnet, ip netip.Addr) bool {
	for _, s := range subnets {
		if s.Contains(ip) {
			return true
		}
	}

	return false
}
