package otp

import "time"

type Store interface {
	Create(challenge *Challenge) (*Challenge, error)
	GetByID(id string) (*Challenge, error)
	GetLatestByEmailIntent(email, intent string) (*Challenge, error)
	IncrementAttempts(id string, updatedAt time.Time) (*Challenge, error)
	IncrementResendCount(id string, updatedAt time.Time) (*Challenge, error)
	MarkVerifiedAndUsed(id string, verifiedAt time.Time) (*Challenge, error)
	DeleteExpired(now time.Time) (int, error)
}
