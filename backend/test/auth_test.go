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
		Identifier: "198001012005011001",
		Password:   "WrongPassword!",
	})
	resp, err := http.Post(env.Server.URL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(invalidBody))
	if err != nil {
		t.Fatalf("Login request failed: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for wrong password, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 2. Test Login with NIP 18 digit (LURAH) -> 200 + Token + Cookie
	lurahLogin, _ := json.Marshal(domain.LoginRequest{
		Identifier: "197503151998031001", // NIP Lurah
		Password:   "AdminSidak2026!",
	})
	lResp, err := http.Post(env.Server.URL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(lurahLogin))
	if err != nil || lResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200 for Lurah login, got %d", lResp.StatusCode)
	}
	var lurahResp domain.LoginResponse
	_ = json.NewDecoder(lResp.Body).Decode(&lurahResp)
	lResp.Body.Close()

	if lurahResp.User.Role != "LURAH" {
		t.Errorf("Expected role LURAH, got %s", lurahResp.User.Role)
	}
	if lurahResp.User.NIP == nil || *lurahResp.User.NIP != "197503151998031001" {
		t.Errorf("Expected NIP 197503151998031001")
	}

	// 3. Test Login with NIP 18 digit (SEKLUR) -> 200
	seklurLogin, _ := json.Marshal(domain.LoginRequest{
		Identifier: "198001012005011001", // NIP Seklur
		Password:   "AdminSidak2026!",
	})
	resp, err = http.Post(env.Server.URL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(seklurLogin))
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200 for Seklur login, got %d", resp.StatusCode)
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

	// 4. Test Login with NIK 16 digit (KADER) -> 200
	kaderLogin, _ := json.Marshal(domain.LoginRequest{
		Identifier: "3205010303920003", // NIK Kader
		Password:   "AdminSidak2026!",
	})
	kResp, err := http.Post(env.Server.URL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(kaderLogin))
	if err != nil || kResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200 for Kader login, got %d", kResp.StatusCode)
	}
	var kaderResp domain.LoginResponse
	_ = json.NewDecoder(kResp.Body).Decode(&kaderResp)
	kResp.Body.Close()

	if kaderResp.User.Role != "KADER" {
		t.Errorf("Expected role KADER, got %s", kaderResp.User.Role)
	}

	// 5. Test /auth/me with Bearer token
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

	if meUser.NIP == nil || *meUser.NIP != "198001012005011001" {
		t.Errorf("Expected NIP 198001012005011001")
	}

	// 6. Test /auth/logout
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
