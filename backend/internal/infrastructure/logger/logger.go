// Package logger provides a single, structured slog.Logger for the whole
// process. JSON in production (machine-parseable for Loki), human-friendly
// text in development. No third-party logging dependency required.
package logger

import (
	"log/slog"
	"os"
	"strings"
)

// New builds the application logger. `env` switches the handler format and
// `level` ("debug"|"info"|"warn"|"error") sets the minimum severity.
func New(env, level string) *slog.Logger {
	opts := &slog.HandlerOptions{Level: parseLevel(level)}

	var handler slog.Handler
	if strings.EqualFold(env, "development") {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler).With(slog.String("service", "iot-api"))
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
