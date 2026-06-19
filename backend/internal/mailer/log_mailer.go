package mailer

import (
	"context"
	"log/slog"
	"strings"
)

const (
	ModeLog = "log"
)

type LogMailer struct {
	logger *slog.Logger
}

func NewLogMailer(logger *slog.Logger) *LogMailer {
	if logger == nil {
		logger = slog.Default()
	}
	return &LogMailer{logger: logger.With("component", "mailer")}
}

func (m *LogMailer) SendOTP(ctx context.Context, email, code string, expiresInMinutes int) error {
	_ = ctx
	m.logger.Info("otp_delivery_logged",
		"email", normalizeEmail(email),
		"code", code,
		"expires_in_minutes", expiresInMinutes,
	)
	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
