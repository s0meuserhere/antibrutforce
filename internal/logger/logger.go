package logger

import (
	"log/slog"
	"os"
)

// Local - окружение разработчика: логи пишутся текстом, чтобы их было удобно читать глазами.
const Local = "local"

// New возвращает логгер, который пишет в stdout записи не ниже level.
// В окружении Local записи идут текстом, в остальных - JSON для сборщика логов.
func New(env string, level slog.Level) *slog.Logger {
	opts := &slog.HandlerOptions{Level: level, AddSource: true}

	if env == Local {
		return slog.New(slog.NewTextHandler(os.Stdout, opts))
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, opts))
}
