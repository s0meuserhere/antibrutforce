package config

import "log/slog"

// Logger - настройки логов.
type Logger struct {
	// Level - нижний уровень записей: debug, info, warn или error.
	Level slog.Level `env:"LOG_LEVEL" env-default:"info"`
}
