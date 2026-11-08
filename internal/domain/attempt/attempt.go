package attempt

import (
	"errors"
	"fmt"
	"net/netip"
	"time"

	"github.com/s0meuserhere/antibrutforce/internal/domain/bucket"
	"github.com/s0meuserhere/antibrutforce/internal/domain/netlist"
	"github.com/s0meuserhere/antibrutforce/internal/domain/subnet"
)

var (
	// ErrEmptyLogin означает попытку без логина.
	ErrEmptyLogin = errors.New("login is empty")

	// ErrEmptyPassword означает попытку без пароля.
	ErrEmptyPassword = errors.New("password is empty")
)

// Attempt - попытка авторизации: логин, пароль и IPv4 клиента.
type Attempt struct {
	login    string
	password string
	ip       netip.Addr
}

// New возвращает попытку с непустыми логином и паролем и адресом IPv4.
func New(login, password, ip string) (Attempt, error) {
	if login == "" {
		return Attempt{}, ErrEmptyLogin
	}

	if password == "" {
		return Attempt{}, ErrEmptyPassword
	}

	addr, err := subnet.ParseIP(ip)
	if err != nil {
		return Attempt{}, fmt.Errorf("attempt: %w", err)
	}

	return Attempt{login: login, password: password, ip: addr}, nil
}

// Login возвращает логин.
func (a Attempt) Login() string {
	return a.login
}

// Password возвращает пароль.
func (a Attempt) Password() string {
	return a.password
}

// IP возвращает адрес клиента.
func (a Attempt) IP() netip.Addr {
	return a.ip
}

// Buckets - бакеты, в которых учитывается попытка.
type Buckets struct {
	login    *bucket.Bucket
	password *bucket.Bucket
	ip       *bucket.Bucket
}

// NewBuckets возвращает бакеты логина, пароля и IP попытки.
func NewBuckets(login, password, ip *bucket.Bucket) Buckets {
	return Buckets{login: login, password: password, ip: ip}
}

// AllowByLists - первый шаг проверки: решает попытку по спискам.
// listed = false, если адреса нет в списках и нужен второй шаг.
func (a Attempt) AllowByLists(lists netlist.Lists) (ok, listed bool) {
	return lists.Allow(a.ip)
}

// AllowByBuckets - второй шаг проверки: учитывает попытку в каждом бакете,
// где есть место, и разрешает её, только если место нашлось во всех.
// Бакеты не останавливаются на первом отказе: попытка, отклонённая по логину,
// всё равно считается в бакетах пароля и IP.
func (a Attempt) AllowByBuckets(now time.Time, buckets Buckets) bool {
	ok := true

	for _, b := range []*bucket.Bucket{buckets.login, buckets.password, buckets.ip} {
		if !b.Allow(now) {
			ok = false
		}
	}

	return ok
}
