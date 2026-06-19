package otp

import (
	"errors"
	"strings"
	"time"
)

const (
	IntentSignup = "signup"
	IntentSignin = "signin"
)

var (
	ErrNotFound      = errors.New("otp challenge not found")
	ErrInvalidIntent = errors.New("invalid otp intent")
)

type Challenge struct {
	ID           string
	Email        string
	Intent       string
	CodeHash     string
	AttemptCount int
	MaxAttempts  int
	ResendCount  int
	CreatedAt    time.Time
	UpdatedAt    time.Time
	ExpiresAt    time.Time
	UsedAt       *time.Time
	VerifiedAt   *time.Time
}

type Snapshot struct {
	ID           string     `json:"id"`
	Email        string     `json:"email"`
	Intent       string     `json:"intent"`
	AttemptCount int        `json:"attemptCount"`
	MaxAttempts  int        `json:"maxAttempts"`
	ResendCount  int        `json:"resendCount"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	ExpiresAt    time.Time  `json:"expiresAt"`
	UsedAt       *time.Time `json:"usedAt,omitempty"`
	VerifiedAt   *time.Time `json:"verifiedAt,omitempty"`
}

func ToSnapshot(challenge *Challenge) Snapshot {
	if challenge == nil {
		return Snapshot{}
	}

	return Snapshot{
		ID:           challenge.ID,
		Email:        challenge.Email,
		Intent:       challenge.Intent,
		AttemptCount: challenge.AttemptCount,
		MaxAttempts:  challenge.MaxAttempts,
		ResendCount:  challenge.ResendCount,
		CreatedAt:    challenge.CreatedAt,
		UpdatedAt:    challenge.UpdatedAt,
		ExpiresAt:    challenge.ExpiresAt,
		UsedAt:       cloneTimePtr(challenge.UsedAt),
		VerifiedAt:   cloneTimePtr(challenge.VerifiedAt),
	}
}

func NormalizeIntent(intent string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(intent))
	switch normalized {
	case IntentSignup, IntentSignin:
		return normalized, nil
	default:
		return "", ErrInvalidIntent
	}
}

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func cloneChallenge(challenge *Challenge) *Challenge {
	if challenge == nil {
		return nil
	}

	return &Challenge{
		ID:           challenge.ID,
		Email:        challenge.Email,
		Intent:       challenge.Intent,
		CodeHash:     challenge.CodeHash,
		AttemptCount: challenge.AttemptCount,
		MaxAttempts:  challenge.MaxAttempts,
		ResendCount:  challenge.ResendCount,
		CreatedAt:    challenge.CreatedAt,
		UpdatedAt:    challenge.UpdatedAt,
		ExpiresAt:    challenge.ExpiresAt,
		UsedAt:       cloneTimePtr(challenge.UsedAt),
		VerifiedAt:   cloneTimePtr(challenge.VerifiedAt),
	}
}

func cloneTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}

	copy := *value
	return &copy
}
