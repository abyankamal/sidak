package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/abyankamal/sidak/backend/internal/domain"
)

func TestAuthFlow(t *testing.T) {
	env := SetupTestEnv(t)

	// 1. Test Login with invalid credentials -> 401
	invalidBody, _ := json.Marshal(domain.LoginRequest{
		NIK:      "3205010101800001",
		Password: "WrongPassword!",
	})
	resp, err := http.Post(env.Server.URL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(invalidBody))
	if err != nil {
		t.Fatalf("Login request failed: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for wrong password, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 2. Test Login with valid SEKDES/SEKLUR -> 200 + Token + Cookie
	validBody, _ := json.Marshal(domain.LoginRequest{
		NIK:      "3205010101800001",
		Password: "AdminSidak2026!",
	})
	resp, err = http.Post(env.Server.URL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(validBody))
	if err != nil {
		t.Fatalf("Login request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200 for valid login, got %d", resp.StatusCode)
	}

	var loginResp domain.LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		t.Fatalf("Failed to decode login response: %v", err)
	}
	resp.Body.Close()

	if loginResp.Token == "" {
		t.Errorf("Expected non-empty JWT token")
	}
	if loginResp.User.Role != "SEKLUR" {
		t.Errorf("Expected role SEKLUR, got %s", loginResp.User.Role)
	}

	// 3. Test /auth/me with Bearer token
	req, _ := http.NewRequest(http.MethodGet, env.Server.URL+"/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+loginResp.Token)
	meResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Get me request failed: %v", err)
	}
	if meResp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for /auth/me, got %d", meResp.StatusCode)
	}
	var meUser domain.UserResponse
	_ = json.NewDecoder(meResp.Body).Decode(&meUser)
	meResp.Body.Close()

	if meUser.NIK != "3205010101800001" {
		t.Errorf("Expected NIK 3205010101800001, got %s", meUser.NIK)
	}

	// 4. Test /auth/logout
	logoutReq, _ := http.NewRequest(http.MethodPost, env.Server.URL+"/api/v1/auth/logout", nil)
	logoutReq.Header.Set("Authorization", "Bearer "+loginResp.Token)
	logoutResp, err := http.DefaultClient.Do(logoutReq)
	if err != nil {
		t.Fatalf("Logout request failed: %v", err)
	}
	if logoutResp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for logout, got %d", logoutResp.StatusCode)
	}
	logoutResp.Body.Close()
}
