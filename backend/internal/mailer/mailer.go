package mailer

import "context"

// Mailer sends OTP challenges via pluggable delivery backends.
type Mailer interface {
	SendOTP(ctx context.Context, email, code string, expiresInMinutes int) error
}
