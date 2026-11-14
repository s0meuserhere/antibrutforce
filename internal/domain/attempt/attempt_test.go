package attempt_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/s0meuserhere/antibrutforce/internal/domain/attempt"
	"github.com/s0meuserhere/antibrutforce/internal/domain/bucket"
	"github.com/s0meuserhere/antibrutforce/internal/domain/netlist"
	"github.com/s0meuserhere/antibrutforce/internal/domain/subnet"
)

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		login    string
		password string
		ip       string
		wantErr  error
	}{
		{name: "valid", login: "user", password: "secret", ip: "192.1.1.10"},
		{name: "empty login", login: "", password: "secret", ip: "192.1.1.10", wantErr: attempt.ErrEmptyLogin},
		{name: "empty password", login: "user", password: "", ip: "192.1.1.10", wantErr: attempt.ErrEmptyPassword},
		{name: "empty ip", login: "user", password: "secret", ip: "", wantErr: subnet.ErrInvalidIP},
		{name: "IPv6", login: "user", password: "secret", ip: "2001:db8::1", wantErr: subnet.ErrInvalidIP},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := attempt.New(tt.login, tt.password, tt.ip)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("New() error = %v, want %v", err, tt.wantErr)
			}

			if tt.wantErr != nil {
				return
			}

			if got.Login() != tt.login || got.Password() != tt.password || got.IP().String() != tt.ip {
				t.Errorf("New() = %s/%s/%s, want %s/%s/%s",
					got.Login(), got.Password(), got.IP(), tt.login, tt.password, tt.ip)
			}
		})
	}
}

func TestNewHidesPassword(t *testing.T) {
	t.Parallel()

	const password = "top-secret-password"

	_, err := attempt.New("user", password, "not-an-ip")
	if err == nil {
		t.Fatal("New() error = nil, want error")
	}

	if got := err.Error(); strings.Contains(got, password) {
		t.Errorf("New() error %q contains password", got)
	}
}

func mustAttempt(t *testing.T, ip string) attempt.Attempt {
	t.Helper()

	a, err := attempt.New("user", "secret", ip)
	if err != nil {
		t.Fatalf("attempt.New() error = %v", err)
	}

	return a
}

func mustSubnets(t *testing.T, cidrs ...string) []subnet.Subnet {
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

func TestAllowByLists(t *testing.T) {
	t.Parallel()

	lists := netlist.New(mustSubnets(t, "192.1.1.0/25"), mustSubnets(t, "10.0.0.0/8"))

	tests := []struct {
		name       string
		ip         string
		wantOK     bool
		wantListed bool
	}{
		{name: "whitelist", ip: "192.1.1.10", wantOK: true, wantListed: true},
		{name: "blacklist", ip: "10.1.2.3", wantListed: true},
		{name: "not listed", ip: "172.16.0.1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ok, listed := mustAttempt(t, tt.ip).AllowByLists(lists)
			if ok != tt.wantOK || listed != tt.wantListed {
				t.Errorf("AllowByLists() = %v, %v, want %v, %v", ok, listed, tt.wantOK, tt.wantListed)
			}
		})
	}
}

func TestAllowByBuckets(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		levels     [3]int
		want       bool
		wantLevels [3]int
	}{
		{name: "room in every bucket", levels: [3]int{0, 0, 0}, want: true, wantLevels: [3]int{1, 1, 1}},
		{name: "login bucket full", levels: [3]int{10, 5, 5}, want: false, wantLevels: [3]int{10, 6, 6}},
		{name: "ip bucket full", levels: [3]int{5, 5, 1000}, want: false, wantLevels: [3]int{6, 6, 1000}},
		{name: "every bucket full", levels: [3]int{10, 100, 1000}, want: false, wantLevels: [3]int{10, 100, 1000}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			login := bucket.Restore(10, tt.levels[0], now)
			password := bucket.Restore(100, tt.levels[1], now)
			ip := bucket.Restore(1000, tt.levels[2], now)

			buckets := attempt.NewBuckets(login, password, ip)
			if got := mustAttempt(t, "192.1.1.10").AllowByBuckets(now, buckets); got != tt.want {
				t.Errorf("AllowByBuckets() = %v, want %v", got, tt.want)
			}

			got := [3]int{login.Level(), password.Level(), ip.Level()}
			if got != tt.wantLevels {
				t.Errorf("levels = %v, want %v", got, tt.wantLevels)
			}
		})
	}
}
