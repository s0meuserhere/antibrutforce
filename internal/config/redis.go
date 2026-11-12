package config

import (
	"net"
	"strconv"
)

// Redis - параметры подключения к Redis.
type Redis struct {
	Host     string `env:"REDIS_HOST"     env-default:"localhost"`
	Port     int    `env:"REDIS_PORT"     env-default:"6379"`
	Password string `env:"REDIS_PASSWORD"`
	DB       int    `env:"REDIS_DB"       env-default:"0"`
}

// Addr возвращает адрес Redis в виде host:port.
func (r Redis) Addr() string {
	return net.JoinHostPort(r.Host, strconv.Itoa(r.Port))
}
