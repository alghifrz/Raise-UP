package jwtutil_test

import (
	"testing"
	"time"

	db "github.com/diuk/raiseup/db/generated"
	jwtutil "github.com/diuk/raiseup/pkg/jwt"
)

func TestGenerateAndParse(t *testing.T) {
	manager := jwtutil.NewManager("test-secret-at-least-16", time.Hour)

	token, expiresAt, err := manager.Generate("user-1", "admin@example.com", db.UserRoleADMINRW)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if token == "" {
		t.Fatal("Generate() returned empty token")
	}
	if time.Until(expiresAt) <= 0 {
		t.Fatal("expiresAt should be in the future")
	}

	claims, err := manager.Parse(token)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if claims.UserID != "user-1" {
		t.Fatalf("UserID = %q, want user-1", claims.UserID)
	}
	if claims.Email != "admin@example.com" {
		t.Fatalf("Email = %q, want admin@example.com", claims.Email)
	}
	if claims.Role != db.UserRoleADMINRW {
		t.Fatalf("Role = %q, want ADMIN_RW", claims.Role)
	}
	if claims.ExpiresAt == nil || claims.IssuedAt == nil {
		t.Fatal("expected exp and iat claims")
	}
}

func TestParseExpiredToken(t *testing.T) {
	manager := jwtutil.NewManager("test-secret-at-least-16", -time.Minute)

	token, _, err := manager.Generate("user-1", "admin@example.com", db.UserRoleSUPERADMIN)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	_, err = manager.Parse(token)
	if err != jwtutil.ErrExpiredToken {
		t.Fatalf("Parse() error = %v, want ErrExpiredToken", err)
	}
}

func TestParseMalformedToken(t *testing.T) {
	manager := jwtutil.NewManager("test-secret-at-least-16", time.Hour)

	_, err := manager.Parse("not-a-jwt")
	if err != jwtutil.ErrInvalidToken {
		t.Fatalf("Parse() error = %v, want ErrInvalidToken", err)
	}
}

func TestParseWrongSecret(t *testing.T) {
	issuer := jwtutil.NewManager("issuer-secret-12345", time.Hour)
	verifier := jwtutil.NewManager("other-secret-123456", time.Hour)

	token, _, err := issuer.Generate("user-1", "admin@example.com", db.UserRoleADMINRW)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	_, err = verifier.Parse(token)
	if err != jwtutil.ErrInvalidToken {
		t.Fatalf("Parse() error = %v, want ErrInvalidToken", err)
	}
}
