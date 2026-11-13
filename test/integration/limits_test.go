//go:build integration

package integration

import (
	"testing"

	antibfv1 "github.com/s0meuserhere/antibrutforce/gen/antibf/v1"
)

// Попытки в каждом подтесте отличаются всем, кроме проверяемого поля.
func TestLimits(t *testing.T) {
	t.Run("login", func(t *testing.T) {
		login := unique("login")

		exhaust(t, limitLogin, func() attempt {
			a := newAttempt()
			a.login = login

			return a
		})
	})

	t.Run("password", func(t *testing.T) {
		password := unique("password")

		exhaust(t, limitPassword, func() attempt {
			a := newAttempt()
			a.password = password

			return a
		})
	})

	t.Run("ip", func(t *testing.T) {
		ip := uniqueIP()

		exhaust(t, limitIP, func() attempt {
			a := newAttempt()
			a.ip = ip

			return a
		})
	})
}

func TestReset(t *testing.T) {
	t.Run("login", func(t *testing.T) {
		login := unique("login")
		next := func() attempt {
			a := newAttempt()
			a.login = login

			return a
		}

		exhaust(t, limitLogin, next)
		reset(t, &antibfv1.ResetV1Request{Login: login})

		if !check(t, next()) {
			t.Fatal("login is denied after reset")
		}
	})

	t.Run("ip", func(t *testing.T) {
		ip := uniqueIP()
		next := func() attempt {
			a := newAttempt()
			a.ip = ip

			return a
		}

		exhaust(t, limitIP, next)
		reset(t, &antibfv1.ResetV1Request{Ip: ip})

		if !check(t, next()) {
			t.Fatal("ip is denied after reset")
		}
	})

	t.Run("password bucket stays", func(t *testing.T) {
		password := unique("password")
		next := func() attempt {
			a := newAttempt()
			a.password = password

			return a
		}

		exhaust(t, limitPassword, next)

		a := next()
		reset(t, &antibfv1.ResetV1Request{Login: a.login, Ip: a.ip})

		if check(t, a) {
			t.Fatal("password is allowed after reset of login and ip")
		}
	})
}

func reset(t *testing.T, req *antibfv1.ResetV1Request) {
	t.Helper()

	if _, err := client.ResetV1(t.Context(), req); err != nil {
		t.Fatalf("reset: %v", err)
	}
}
