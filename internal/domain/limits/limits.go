package limits

import (
	"errors"
	"fmt"
)

// Max - наибольшее число попыток в минуту для одного лимита.
const Max = 1_000_000

// ErrInvalidLimit означает лимит вне диапазона от 1 до Max.
var ErrInvalidLimit = errors.New("limit is out of range")

// Limits хранит допустимое число попыток в минуту на один логин, пароль и IP.
type Limits struct {
	perLogin    int
	perPassword int
	perIP       int
}

// New возвращает лимиты, если каждый лежит в диапазоне от 1 до Max.
func New(perLogin, perPassword, perIP int) (Limits, error) {
	checks := []struct {
		name  string
		value int
	}{
		{name: "login", value: perLogin},
		{name: "password", value: perPassword},
		{name: "ip", value: perIP},
	}
	for _, check := range checks {
		if check.value < 1 || check.value > Max {
			return Limits{}, fmt.Errorf("%s limit %d: %w", check.name, check.value, ErrInvalidLimit)
		}
	}

	return Limits{perLogin: perLogin, perPassword: perPassword, perIP: perIP}, nil
}

// PerLogin возвращает лимит попыток в минуту на один логин.
func (l Limits) PerLogin() int {
	return l.perLogin
}

// PerPassword возвращает лимит попыток в минуту на один пароль.
func (l Limits) PerPassword() int {
	return l.perPassword
}

// PerIP возвращает лимит попыток в минуту на один IP.
func (l Limits) PerIP() int {
	return l.perIP
}
