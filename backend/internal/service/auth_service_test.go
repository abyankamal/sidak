package service

import (
	"testing"
	"time"

	"github.com/abyankamal/sidak/backend/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthServiceUnit(t *testing.T) {
	jwtSecret := "test_super_secret_jwt_key_sidak_2026_test"
	authService := NewAuthService(nil, jwtSecret)

	// 1. Test Password Hashing
	password := "AdminSidak2026!"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		t.Fatalf("Failed to generate password hash: %v", err)
	}

	if err := bcrypt.CompareHashAndPassword(hash, []byte(password)); err != nil {
		t.Errorf("Password compare failed: %v", err)
	}

	if err := bcrypt.CompareHashAndPassword(hash, []byte("WrongPass")); err == nil {
		t.Errorf("Expected compare error for wrong password")
	}

	// 2. Test JWT Token Generation & Parsing
	testUser := &domain.User{
		ID:        "01ARZ3NDEKTSV4RRFFQ69G5001",
		NIK:       "3205010101800001",
		Nama:      "Drs. H. Mulyadi (Seklur)",
		Role:      "SEKLUR",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tokenStr, err := authService.GenerateToken(testUser)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}
	if tokenStr == "" {
		t.Fatalf("Expected non-empty token string")
	}

	claims, err := authService.ParseToken(tokenStr)
	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}

	if claims.UserID != testUser.ID {
		t.Errorf("Expected UserID %s, got %s", testUser.ID, claims.UserID)
	}
	if claims.Role != "SEKLUR" {
		t.Errorf("Expected Role SEKLUR, got %s", claims.Role)
	}

	// 3. Test Invalid Token
	_, err = authService.ParseToken("invalid.token.string")
	if err == nil {
		t.Errorf("Expected error for invalid token string")
	}
}
