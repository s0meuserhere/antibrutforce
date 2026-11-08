package subnet

import (
	"errors"
	"fmt"
	"net/netip"
)

var (
	// ErrInvalidSubnet означает строку, которая не является подсетью IPv4 в записи CIDR.
	ErrInvalidSubnet = errors.New("invalid IPv4 subnet")

	// ErrInvalidIP означает строку, которая не является адресом IPv4.
	ErrInvalidIP = errors.New("invalid IPv4 address")
)

// Subnet - подсеть IPv4, приведённая к адресу сети.
type Subnet struct {
	prefix netip.Prefix
}

// New разбирает подсеть IPv4 в записи CIDR и приводит её к адресу сети:
// 192.1.1.10/25 становится 192.1.1.0/25.
func New(cidr string) (Subnet, error) {
	prefix, err := netip.ParsePrefix(cidr)
	if err != nil {
		return Subnet{}, fmt.Errorf("%w: %w", ErrInvalidSubnet, err)
	}

	if !prefix.Addr().Is4() {
		return Subnet{}, fmt.Errorf("%w: %s is not IPv4", ErrInvalidSubnet, cidr)
	}

	return Subnet{prefix: prefix.Masked()}, nil
}

// Restore собирает подсеть из сохранённого префикса без проверок New.
func Restore(prefix netip.Prefix) Subnet {
	return Subnet{prefix: prefix}
}

// Prefix возвращает подсеть как netip.Prefix.
func (s Subnet) Prefix() netip.Prefix {
	return s.prefix
}

// String возвращает подсеть в записи CIDR.
func (s Subnet) String() string {
	return s.prefix.String()
}

// Contains сообщает, входит ли адрес в подсеть.
func (s Subnet) Contains(ip netip.Addr) bool {
	return s.prefix.Contains(ip)
}

// ParseIP разбирает адрес IPv4.
func ParseIP(s string) (netip.Addr, error) {
	ip, err := netip.ParseAddr(s)
	if err != nil {
		return netip.Addr{}, fmt.Errorf("%w: %w", ErrInvalidIP, err)
	}

	if !ip.Is4() {
		return netip.Addr{}, fmt.Errorf("%w: %s is not IPv4", ErrInvalidIP, s)
	}

	return ip, nil
}
