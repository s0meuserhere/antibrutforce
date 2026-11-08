package netlist

import (
	"errors"
	"fmt"

	"github.com/s0meuserhere/antibrutforce/internal/domain/subnet"
)

// ErrUnknownKind означает неизвестный вид списка.
var ErrUnknownKind = errors.New("unknown list kind")

// Entry - подсеть в списке определённого вида.
type Entry struct {
	kind   Kind
	subnet subnet.Subnet
}

// NewEntry возвращает запись списка. Вид списка должен быть известным,
// подсеть - IPv4 в записи CIDR. Адрес хоста в подсети обнуляется:
// 192.1.1.10/25 становится 192.1.1.0/25.
func NewEntry(kind Kind, cidr string) (Entry, error) {
	if kind != Whitelist && kind != Blacklist {
		return Entry{}, fmt.Errorf("%w: %q", ErrUnknownKind, kind)
	}

	s, err := subnet.New(cidr)
	if err != nil {
		return Entry{}, fmt.Errorf("entry: %w", err)
	}

	return Entry{kind: kind, subnet: s}, nil
}

// Kind возвращает вид списка.
func (e Entry) Kind() Kind {
	return e.kind
}

// Subnet возвращает подсеть.
func (e Entry) Subnet() subnet.Subnet {
	return e.subnet
}
