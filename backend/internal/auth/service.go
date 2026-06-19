package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"realtime-session-coordination/backend/internal/user"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
)

type Claims struct {
	UserType string `json:"user_type"`
	IsActive bool   `json:"is_active"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type Service struct {
	users  user.Store
	secret []byte
	issuer string
	expiry time.Duration
}

func NewService(users user.Store, secret string, expiry time.Duration, issuer string) (*Service, error) {
	if users == nil {
		return nil, fmt.Errorf("users store is required")
	}
	if strings.TrimSpace(secret) == "" {
		return nil, fmt.Errorf("jwt secret is required")
	}
	if expiry <= 0 {
		return nil, fmt.Errorf("jwt expiry must be > 0")
	}
	if issuer == "" {
		issuer = "realtime-session-coordination"
	}

	return &Service{
		users:  users,
		secret: []byte(secret),
		issuer: issuer,
		expiry: expiry,
	}, nil
}

func (s *Service) CreateGuest() (*user.User, string, error) {
	now := time.Now().UTC()

	u := &user.User{
		ID:        newUserID(),
		Type:      user.TypeGuest,
		CreatedAt: now,
		UpdatedAt: now,
		IsVisible: true,
		IsActive:  true,
	}

	created, err := s.users.Create(u)
	if err != nil {
		return nil, "", err
	}

	token, err := s.issueToken(created)
	if err != nil {
		return nil, "", err
	}

	return created, token, nil
}

func (s *Service) ValidateToken(rawToken string) (*Claims, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return nil, ErrUnauthorized
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrUnauthorized
	}

	if claims.Subject == "" || claims.UserType == "" || !claims.IsActive || claims.Role == "" {
		return nil, ErrUnauthorized
	}

	return claims, nil
}

// IssueTokenForUser mints a signed JWT for an existing user.
// Used by the OTP service to issue tokens after email verification.
func (s *Service) IssueTokenForUser(u *user.User) (string, error) {
	return s.issueToken(u)
}

func (s *Service) issueToken(u *user.User) (string, error) {
	now := time.Now().UTC()

	claims := Claims{
		UserType: u.Type,
		IsActive: u.IsActive,
		Role:     u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   u.ID,
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.expiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.secret)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func newUserID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("user_%s", hex.EncodeToString(b))
}
