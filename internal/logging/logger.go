package logging

import (
	"fmt"
	"io"
	"log/slog"
)

func New(level string, output io.Writer) (*slog.Logger, error) {
	slogLevel, err := parseLevel(level)
	if err != nil {
		return nil, err
	}

	handler := slog.NewTextHandler(output, &slog.HandlerOptions{Level: slogLevel})
	return slog.New(handler), nil
}

func parseLevel(level string) (slog.Level, error) {
	switch level {
	case "", "info":
		return slog.LevelInfo, nil
	case "debug":
		return slog.LevelDebug, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unsupported log level: %s", level)
	}
}
