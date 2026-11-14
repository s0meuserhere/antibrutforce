package config

import (
	"net"
	"strconv"
)

// GRPC - адрес, на котором сервер принимает запросы.
type GRPC struct {
	Host string `env:"GRPC_HOST" env-default:"0.0.0.0"`
	Port int    `env:"GRPC_PORT" env-default:"50051"`
}

// Addr возвращает адрес в виде host:port.
func (g GRPC) Addr() string {
	return net.JoinHostPort(g.Host, strconv.Itoa(g.Port))
}
