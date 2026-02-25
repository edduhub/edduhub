package jwt

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestGenerateAndVerifyToken(t *testing.T) {
	manager, err := NewJWTManager("test-secret-key-that-is-at-least-32-chars", time.Hour)
	if err != nil {
		t.Fatalf("NewJWTManager returned error: %v", err)
	}

	token, err := manager.Generate(7, "kratos-123", "student@example.com", "student", "1", "Jane", "Doe")
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	claims, err := manager.Verify(token)
	if err != nil {
		t.Fatalf("Verify returned error: %v", err)
	}

	if claims.UserID != 7 {
		t.Fatalf("expected user id 7, got %d", claims.UserID)
	}
	if claims.KratosID != "kratos-123" {
		t.Fatalf("expected kratos id kratos-123, got %s", claims.KratosID)
	}
	if claims.Email != "student@example.com" {
		t.Fatalf("expected email student@example.com, got %s", claims.Email)
	}
	if claims.Role != "student" {
		t.Fatalf("expected role student, got %s", claims.Role)
	}
}

func TestVerifyExpiredToken(t *testing.T) {
	manager, err := NewJWTManager("test-secret-key-that-is-at-least-32-chars", -1*time.Minute)
	if err != nil {
		t.Fatalf("NewJWTManager returned error: %v", err)
	}
	token, err := manager.Generate(7, "kratos-123", "student@example.com", "student", "1", "Jane", "Doe")
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	_, err = manager.Verify(token)
	if err == nil {
		t.Fatalf("expected expired token verification to fail")
	}

	if !errors.Is(err, ErrExpiredToken) && !strings.Contains(strings.ToLower(err.Error()), "expired") {
		t.Fatalf("expected expired token error, got %v", err)
	}
}

func TestVerifyInvalidSignature(t *testing.T) {
	issuer, err := NewJWTManager("this-is-a-very-long-secret-key-for-issuer", time.Hour)
	if err != nil {
		t.Fatalf("NewJWTManager returned error: %v", err)
	}
	validator, err := NewJWTManager("this-is-a-very-long-secret-key-for-validator", time.Hour)
	if err != nil {
		t.Fatalf("NewJWTManager returned error: %v", err)
	}

	token, err := issuer.Generate(7, "kratos-123", "student@example.com", "student", "1", "Jane", "Doe")
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	_, err = validator.Verify(token)
	if err == nil {
		t.Fatalf("expected invalid signature verification to fail")
	}
}

func TestVerifyMalformedToken(t *testing.T) {
	manager, err := NewJWTManager("test-secret-key-that-is-at-least-32-chars", time.Hour)
	if err != nil {
		t.Fatalf("NewJWTManager returned error: %v", err)
	}

	_, err = manager.Verify("not-a-jwt-token")
	if err == nil {
		t.Fatalf("expected malformed token verification to fail")
	}
}
