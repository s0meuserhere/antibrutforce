//go:build integration

package integration

import (
	"testing"

	antibfv1 "github.com/s0meuserhere/antibrutforce/gen/antibf/v1"
)

// Подсети тестов списков не пересекаются друг с другом и с адресами uniqueIP.

func TestBlacklist(t *testing.T) {
	addToBlacklist(t, "192.0.2.0/24")

	a := newAttempt()
	a.ip = "192.0.2.10"

	// Адрес из blacklist отклоняется на каждой попытке и не тратит бакеты.
	for i := range limitLogin + 1 {
		if check(t, a) {
			t.Fatalf("attempt %d from blacklist is allowed", i+1)
		}
	}

	removeFromBlacklist(t, "192.0.2.0/24")

	if !check(t, a) {
		t.Fatal("attempts from blacklist were counted in buckets")
	}
}

func TestWhitelist(t *testing.T) {
	addToWhitelist(t, "203.0.113.0/24")

	a := newAttempt()
	a.ip = "203.0.113.10"

	// Адрес из whitelist пропускается сверх лимита и не тратит бакеты.
	for i := range limitLogin + 1 {
		if !check(t, a) {
			t.Fatalf("attempt %d from whitelist is denied", i+1)
		}
	}

	removeFromWhitelist(t, "203.0.113.0/24")

	if !check(t, a) {
		t.Fatal("attempts from whitelist were counted in buckets")
	}
}

func TestWhitelistWinsOverBlacklist(t *testing.T) {
	addToBlacklist(t, "198.51.100.0/24")
	addToWhitelist(t, "198.51.100.0/25")

	inBoth := newAttempt()
	inBoth.ip = "198.51.100.10"

	if !check(t, inBoth) {
		t.Fatal("address from both lists is denied")
	}

	onlyBlack := newAttempt()
	onlyBlack.ip = "198.51.100.200"

	if check(t, onlyBlack) {
		t.Fatal("address only from blacklist is allowed")
	}

	removeFromWhitelist(t, "198.51.100.0/25")
	removeFromBlacklist(t, "198.51.100.0/24")
}

// Подсеть сохраняется по адресу сети, поэтому удаляется в нормальном виде.
func TestSubnetIsMasked(t *testing.T) {
	addToBlacklist(t, "100.64.5.9/24")

	a := newAttempt()
	a.ip = "100.64.5.1"

	if check(t, a) {
		t.Fatal("address from 100.64.5.9/24 is allowed")
	}

	removeFromBlacklist(t, "100.64.5.0/24")

	if !check(t, a) {
		t.Fatal("100.64.5.0/24 is not removed")
	}
}

// Повторное добавление и удаление отсутствующей подсети не считаются ошибкой.
func TestListsAreIdempotent(t *testing.T) {
	addToBlacklist(t, "100.65.0.0/16")
	addToBlacklist(t, "100.65.0.0/16")
	removeFromBlacklist(t, "100.65.0.0/16")
	removeFromBlacklist(t, "100.65.0.0/16")

	addToWhitelist(t, "100.66.0.0/16")
	addToWhitelist(t, "100.66.0.0/16")
	removeFromWhitelist(t, "100.66.0.0/16")
	removeFromWhitelist(t, "100.66.0.0/16")
}

func addToBlacklist(t *testing.T, subnet string) {
	t.Helper()

	if _, err := client.AddToBlacklistV1(t.Context(), &antibfv1.AddToBlacklistV1Request{Subnet: subnet}); err != nil {
		t.Fatalf("add %s to blacklist: %v", subnet, err)
	}
}

func removeFromBlacklist(t *testing.T, subnet string) {
	t.Helper()

	req := &antibfv1.RemoveFromBlacklistV1Request{Subnet: subnet}
	if _, err := client.RemoveFromBlacklistV1(t.Context(), req); err != nil {
		t.Fatalf("remove %s from blacklist: %v", subnet, err)
	}
}

func addToWhitelist(t *testing.T, subnet string) {
	t.Helper()

	if _, err := client.AddToWhitelistV1(t.Context(), &antibfv1.AddToWhitelistV1Request{Subnet: subnet}); err != nil {
		t.Fatalf("add %s to whitelist: %v", subnet, err)
	}
}

func removeFromWhitelist(t *testing.T, subnet string) {
	t.Helper()

	req := &antibfv1.RemoveFromWhitelistV1Request{Subnet: subnet}
	if _, err := client.RemoveFromWhitelistV1(t.Context(), req); err != nil {
		t.Fatalf("remove %s from whitelist: %v", subnet, err)
	}
}
