package config

import (
	"net"
	"net/url"
	"strconv"
)

// Postgres - параметры подключения к PostgreSQL.
type Postgres struct {
	Host     string `env:"DB_HOST"     env-default:"localhost"`
	Port     int    `env:"DB_PORT"     env-default:"5432"`
	User     string `env:"DB_USER"     env-required:"true"`
	Password string `env:"DB_PASSWORD" env-required:"true"`
	Name     string `env:"DB_NAME"     env-required:"true"`
	SSLMode  string `env:"DB_SSL"      env-default:"disable"`
}

// DSN возвращает строку подключения, в которой логин и пароль экранированы.
func (p Postgres) DSN() string {
	dsn := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(p.User, p.Password),
		Host:     net.JoinHostPort(p.Host, strconv.Itoa(p.Port)),
		Path:     p.Name,
		RawQuery: url.Values{"sslmode": {p.SSLMode}}.Encode(),
	}

	return dsn.String()
}
