package reset

import (
	"errors"
	"fmt"
	"net/netip"

	"github.com/s0meuserhere/antibrutforce/internal/domain/subnet"
)

// ErrEmpty означает сброс без логина и без IP.
var ErrEmpty = errors.New("login or ip is required")

// Target - что сбросить: бакет логина, бакет IP или оба.
type Target struct {
	login string
	ip    netip.Addr
}

// New возвращает цель сброса. Нужен хотя бы логин или IP,
// заданный IP должен быть адресом IPv4.
func New(login, ip string) (Target, error) {
	if login == "" && ip == "" {
		return Target{}, ErrEmpty
	}

	if ip == "" {
		return Target{login: login}, nil
	}

	addr, err := subnet.ParseIP(ip)
	if err != nil {
		return Target{}, fmt.Errorf("reset: %w", err)
	}

	return Target{login: login, ip: addr}, nil
}

// Login возвращает логин или пустую строку, если логин сбрасывать не нужно.
func (t Target) Login() string {
	return t.login
}

// IP возвращает адрес или нулевой адрес, если IP сбрасывать не нужно.
func (t Target) IP() netip.Addr {
	return t.ip
}
