package config

import (
	"errors"
	"fmt"
	"io/fs"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

// ErrInvalid - значение настройки задано, но не подходит.
var ErrInvalid = errors.New("invalid config")

// App - настройки, общие для сервера и CLI: оба собирают один и тот же сервис.
type App struct {
	Postgres Postgres
	Redis    Redis
	Limits   Limits

	// NetlistCacheTTL ограничивает, насколько кэш списков может отстать от PostgreSQL,
	// если запись в базу и чтение промаха пересеклись.
	NetlistCacheTTL time.Duration `env:"NETLIST_CACHE_TTL" env-default:"1m"`

	// RequestTimeout - срок одного запроса к сервису. Без срока повтор
	// транзакции бакетов при конфликте не остановится.
	RequestTimeout time.Duration `env:"REQUEST_TIMEOUT" env-default:"1s"`
}

// Server - настройки сервера.
type Server struct {
	App

	// Environment - окружение запуска. В local логи пишутся текстом, в остальных - JSON.
	Environment string `env:"ENVIRONMENT" env-default:"local"`

	GRPC   GRPC
	Logger Logger
}

// LoadServer читает настройки сервера. Файл path необязателен,
// переменные окружения важнее значений из него.
func LoadServer(path string) (Server, error) {
	var cfg Server
	if err := load(path, &cfg); err != nil {
		return Server{}, err
	}

	if err := cfg.validate(); err != nil {
		return Server{}, err
	}

	return cfg, nil
}

// LoadApp читает настройки, общие для сервера и CLI. Файл path необязателен,
// переменные окружения важнее значений из него.
func LoadApp(path string) (App, error) {
	var cfg App
	if err := load(path, &cfg); err != nil {
		return App{}, err
	}

	if err := cfg.validate(); err != nil {
		return App{}, err
	}

	return cfg, nil
}

// load подгружает файл в окружение, не затирая уже заданные переменные,
// и читает cfg из окружения.
func load(path string, cfg any) error {
	if err := godotenv.Load(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("read config file %q: %w", path, err)
	}

	if err := cleanenv.ReadEnv(cfg); err != nil {
		return fmt.Errorf("read env: %w", err)
	}

	return nil
}

func (a *App) validate() error {
	if a.NetlistCacheTTL <= 0 {
		return fmt.Errorf("%w: NETLIST_CACHE_TTL must be positive, got %s", ErrInvalid, a.NetlistCacheTTL)
	}

	if a.RequestTimeout <= 0 {
		return fmt.Errorf("%w: REQUEST_TIMEOUT must be positive, got %s", ErrInvalid, a.RequestTimeout)
	}

	return nil
}
