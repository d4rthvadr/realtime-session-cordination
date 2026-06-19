package mailer

import (
	"fmt"
	"log/slog"
	"strings"
)

func New(mode string, logger *slog.Logger) (Mailer, error) {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "", ModeLog:
		return NewLogMailer(logger), nil
	default:
		return nil, fmt.Errorf("MAILER_MODE must be 'log'")
	}
}
