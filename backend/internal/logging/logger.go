package logging

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

func Default() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
}

func New(levelRaw, formatRaw string) (*slog.Logger, error) {
	level, err := parseLevel(levelRaw)
	if err != nil {
		return nil, err
	}

	format := strings.ToLower(strings.TrimSpace(formatRaw))
	if format == "" {
		format = "json"
	}

	handlerOpts := &slog.HandlerOptions{Level: level}

	switch format {
	case "json":
		return slog.New(slog.NewJSONHandler(os.Stdout, handlerOpts)), nil
	case "text":
		return slog.New(slog.NewTextHandler(os.Stdout, handlerOpts)), nil
	default:
		return nil, fmt.Errorf("LOG_FORMAT must be 'json' or 'text'")
	}
}

func parseLevel(levelRaw string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(levelRaw)) {
	case "", "info":
		return slog.LevelInfo, nil
	case "debug":
		return slog.LevelDebug, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("LOG_LEVEL must be one of: debug, info, warn, error")
	}
}
