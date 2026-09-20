package jwtutil

import (
	"errors"
	"fmt"
	"time"

	db "github.com/diuk/raiseup/db/generated"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

// Claims are the JWT claims used by RAISE UP authentication.
type Claims struct {
	UserID string      `json:"user_id"`
	Email  string      `json:"email"`
	Role   db.UserRole `json:"role"`
	jwt.RegisteredClaims
}

// Manager signs and verifies access tokens.
type Manager struct {
	secret    []byte
	expiresIn time.Duration
}

// NewManager creates a JWT manager.
func NewManager(secret string, expiresIn time.Duration) *Manager {
	return &Manager{
		secret:    []byte(secret),
		expiresIn: expiresIn,
	}
}

// ExpiresIn returns the configured access-token lifetime.
func (m *Manager) ExpiresIn() time.Duration {
	return m.expiresIn
}

// Generate creates a signed access token for the given identity.
func (m *Manager) Generate(userID, email string, role db.UserRole) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(m.expiresIn)

	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}

	return signed, expiresAt, nil
}

// Parse verifies a token and returns its claims.
func (m *Manager) Parse(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("%w: unexpected signing method", ErrInvalidToken)
		}
		return m.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	if claims.UserID == "" || claims.Email == "" || claims.Role == "" {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
