package otp

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
	"time"

	"realtime-session-coordination/backend/internal/auth"
	"realtime-session-coordination/backend/internal/mailer"
	"realtime-session-coordination/backend/internal/user"
)

var (
	ErrVerificationFailed     = errors.New("verification failed")
	ErrOTPExpired             = errors.New("otp challenge has expired")
	ErrOTPAlreadyUsed         = errors.New("otp challenge has already been used")
	ErrMaxAttemptsReached     = errors.New("maximum verification attempts reached")
	ErrResendCooldown         = errors.New("please wait before requesting a new code")
	ErrEmailAlreadyRegistered = errors.New("an account with this email already exists")
)

// RequestResult is returned by RequestOTP.
type RequestResult struct {
	ChallengeID      string
	ExpiresInMinutes int
}

// VerifyResult is returned by VerifyOTP on success.
type VerifyResult struct {
	User  *user.User
	Token string
}

// ServiceConfig configures OTP service behaviour.
type ServiceConfig struct {
	ExpiryMinutes  int
	MaxAttempts    int
	ResendCooldown time.Duration
}

// Service orchestrates OTP request and verification for signup and signin.
type Service struct {
	challenges  Store
	users       user.Store
	authService *auth.Service
	mailer      mailer.Mailer
	logger      *slog.Logger
	cfg         ServiceConfig
}

func NewService(
	challenges Store,
	users user.Store,
	authService *auth.Service,
	m mailer.Mailer,
	logger *slog.Logger,
	cfg ServiceConfig,
) *Service {
	if cfg.ExpiryMinutes <= 0 {
		cfg.ExpiryMinutes = 10
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 5
	}
	if cfg.ResendCooldown <= 0 {
		cfg.ResendCooldown = 30 * time.Second
	}
	if logger == nil {
		logger = slog.Default()
	}

	return &Service{
		challenges:  challenges,
		users:       users,
		authService: authService,
		mailer:      m,
		logger:      logger.With("component", "otp_service"),
		cfg:         cfg,
	}
}

// RequestOTP creates an OTP challenge for the given email and intent.
// For signup: any email is accepted (user may or may not exist; we let verify enforce).
// For signin: also always returns success to avoid user enumeration.
func (s *Service) RequestOTP(ctx context.Context, email, intent string) (RequestResult, error) {
	normalizedEmail := NormalizeEmail(email)
	if normalizedEmail == "" || !strings.Contains(normalizedEmail, "@") {
		return RequestResult{}, fmt.Errorf("invalid email address")
	}

	normalizedIntent, err := NormalizeIntent(intent)
	if err != nil {
		return RequestResult{}, err
	}

	// Enforce resend cooldown using the latest challenge for this email+intent.
	if latest, lookupErr := s.challenges.GetLatestByEmailIntent(normalizedEmail, normalizedIntent); lookupErr == nil {
		cooldownUntil := latest.CreatedAt.Add(s.cfg.ResendCooldown)
		if time.Now().UTC().Before(cooldownUntil) {
			return RequestResult{}, ErrResendCooldown
		}
	}

	code, err := generateCode()
	if err != nil {
		return RequestResult{}, fmt.Errorf("failed to generate otp code: %w", err)
	}

	codeHash := hashCode(code)

	now := time.Now().UTC()
	challenge := &Challenge{
		ID:          newChallengeID(),
		Email:       normalizedEmail,
		Intent:      normalizedIntent,
		CodeHash:    codeHash,
		MaxAttempts: s.cfg.MaxAttempts,
		CreatedAt:   now,
		UpdatedAt:   now,
		ExpiresAt:   now.Add(time.Duration(s.cfg.ExpiryMinutes) * time.Minute),
	}

	created, err := s.challenges.Create(challenge)
	if err != nil {
		return RequestResult{}, fmt.Errorf("failed to store otp challenge: %w", err)
	}

	if err := s.mailer.SendOTP(ctx, normalizedEmail, code, s.cfg.ExpiryMinutes); err != nil {
		// Non-fatal: log the failure but return the challenge ID so the caller can retry delivery.
		s.logger.Error("otp_mailer_send_failed", "error", err, "email", normalizedEmail, "intent", normalizedIntent)
	}

	s.logger.Info("otp_requested", "email", normalizedEmail, "intent", normalizedIntent, "challenge_id", created.ID)

	return RequestResult{
		ChallengeID:      created.ID,
		ExpiresInMinutes: s.cfg.ExpiryMinutes,
	}, nil
}

// VerifyOTP validates an OTP code for a given challenge and issues a JWT on success.
// Signup: creates a new normal user. Signin: looks up existing user by email.
// Returns ErrVerificationFailed generically when the code is wrong or the user is not found
// for signin, to avoid leaking account existence.
func (s *Service) VerifyOTP(ctx context.Context, email, intent, challengeID, code string) (VerifyResult, error) {
	_ = ctx

	normalizedEmail := NormalizeEmail(email)
	normalizedIntent, err := NormalizeIntent(intent)
	if err != nil {
		return VerifyResult{}, ErrVerificationFailed
	}

	challenge, err := s.challenges.GetByID(challengeID)
	if err != nil {
		return VerifyResult{}, ErrVerificationFailed
	}

	// Validate ownership: email and intent must match the stored challenge.
	if challenge.Email != normalizedEmail || challenge.Intent != normalizedIntent {
		return VerifyResult{}, ErrVerificationFailed
	}

	if challenge.UsedAt != nil {
		return VerifyResult{}, ErrOTPAlreadyUsed
	}

	if time.Now().UTC().After(challenge.ExpiresAt) {
		return VerifyResult{}, ErrOTPExpired
	}

	if challenge.AttemptCount >= challenge.MaxAttempts {
		return VerifyResult{}, ErrMaxAttemptsReached
	}

	// Increment attempts before checking the code to prevent timing-based enumeration.
	updated, err := s.challenges.IncrementAttempts(challengeID, time.Now().UTC())
	if err != nil {
		return VerifyResult{}, fmt.Errorf("failed to record attempt: %w", err)
	}

	if hashCode(strings.TrimSpace(code)) != challenge.CodeHash {
		if updated.AttemptCount >= updated.MaxAttempts {
			s.logger.Warn("otp_max_attempts_reached", "challenge_id", challengeID, "email", normalizedEmail)
			return VerifyResult{}, ErrMaxAttemptsReached
		}

		return VerifyResult{}, ErrVerificationFailed
	}

	// Code is valid — mark challenge as used before any user mutations.
	if _, err := s.challenges.MarkVerifiedAndUsed(challengeID, time.Now().UTC()); err != nil {
		return VerifyResult{}, fmt.Errorf("failed to mark challenge as used: %w", err)
	}

	var targetUser *user.User

	switch normalizedIntent {
	case IntentSignup:
		targetUser, err = s.handleSignupVerify(normalizedEmail)
		if err != nil {
			return VerifyResult{}, err
		}

	case IntentSignin:
		targetUser, err = s.handleSigninVerify(normalizedEmail)
		if err != nil {
			return VerifyResult{}, err
		}
	}

	token, err := s.authService.IssueTokenForUser(targetUser)
	if err != nil {
		return VerifyResult{}, fmt.Errorf("failed to issue token: %w", err)
	}

	s.logger.Info("otp_verified", "email", normalizedEmail, "intent", normalizedIntent, "user_id", targetUser.ID)

	return VerifyResult{User: targetUser, Token: token}, nil
}

func (s *Service) handleSignupVerify(email string) (*user.User, error) {
	existing, err := s.users.GetByEmail(email)
	if err == nil && existing != nil {
		return nil, ErrEmailAlreadyRegistered
	}

	now := time.Now().UTC()
	newUser := &user.User{
		ID:              newUserID(),
		Email:           &email,
		EmailVerifiedAt: &now,
		Type:            user.TypeNormal,
		Role:            user.RoleUser,
		CreatedAt:       now,
		UpdatedAt:       now,
		IsVisible:       true,
		IsActive:        true,
	}

	created, err := s.users.Create(newUser)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return created, nil
}

func (s *Service) handleSigninVerify(email string) (*user.User, error) {
	u, err := s.users.GetByEmail(email)
	if err != nil {
		// Return generic error to avoid leaking whether the email is registered.
		return nil, ErrVerificationFailed
	}

	if !u.IsActive || u.DeletedAt != nil {
		return nil, ErrVerificationFailed
	}

	return u, nil
}

func generateCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%06d", n.Int64()), nil
}

func hashCode(code string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(code)))
	return hex.EncodeToString(sum[:])
}

func newChallengeID() string {
	b := make([]byte, 10)
	_, _ = rand.Read(b)
	return fmt.Sprintf("chall_%s", hex.EncodeToString(b))
}

func newUserID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("user_%s", hex.EncodeToString(b))
}
