package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/abyankamal/sidak/backend/internal/domain"
	"github.com/oklog/ulid/v2"
)

func TestSyncFlow(t *testing.T) {
	env := SetupTestEnv(t)

	// Login as KADER to obtain JWT
	loginBody, _ := json.Marshal(domain.LoginRequest{
		NIK:      "3205010303920003",
		Password: "AdminSidak2026!",
	})
	lResp, err := http.Post(env.Server.URL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(loginBody))
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	var loginResp domain.LoginResponse
	_ = json.NewDecoder(lResp.Body).Decode(&loginResp)
	lResp.Body.Close()
	token := loginResp.Token

	// 1. Test Request Presigned URL
	trxID := ulid.Make().String()
	presignBody, _ := json.Marshal(domain.PresignUploadRequest{
		TransaksiID: trxID,
		FileName:    "ktp_warga.jpg",
		ContentType: "image/jpeg",
	})
	pReq, _ := http.NewRequest(http.MethodPost, env.Server.URL+"/api/v1/sync/presigned-url", bytes.NewBuffer(presignBody))
	pReq.Header.Set("Authorization", "Bearer "+token)
	pReq.Header.Set("Content-Type", "application/json")
	pResp, err := http.DefaultClient.Do(pReq)
	if err != nil {
		t.Fatalf("Presigned URL request failed: %v", err)
	}
	if pResp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for presigned url, got %d", pResp.StatusCode)
	}
	var presignResp domain.PresignUploadResponse
	_ = json.NewDecoder(pResp.Body).Decode(&presignResp)
	pResp.Body.Close()

	expectedPath := "lampiran/3205010303920003/" + trxID + "_ktp_warga.jpg"
	if presignResp.FilePathR2 != expectedPath {
		t.Errorf("Expected file_path_r2 '%s', got '%s'", expectedPath, presignResp.FilePathR2)
	}

	// 2. Test Commit New Valid Transaction -> 201 Created
	wargaNIK := "3205019999990001"
	validCommitBody, _ := json.Marshal(domain.SyncCommitRequest{
		TransaksiID: trxID,
		WargaNIK:    wargaNIK,
		LayananID:   "SKTM",
		DataIsian: json.RawMessage(`{
			"keperluan": "Pendaftaran Beasiswa Mahasiswa",
			"penghasilan_bulanan": 750000,
			"jumlah_tanggungan": 3,
			"keterangan_pekerjaan": "Buruh Harian Lepas"
		}`),
		Lampiran: []string{presignResp.FilePathR2},
	})

	cReq, _ := http.NewRequest(http.MethodPost, env.Server.URL+"/api/v1/sync/commit", bytes.NewBuffer(validCommitBody))
	cReq.Header.Set("Authorization", "Bearer "+token)
	cReq.Header.Set("Content-Type", "application/json")
	cReq.Header.Set("Idempotency-Key", trxID)
	cResp, err := http.DefaultClient.Do(cReq)
	if err != nil {
		t.Fatalf("Commit request failed: %v", err)
	}
	if cResp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status 201 Created for new commit, got %d", cResp.StatusCode)
	}
	cResp.Body.Close()

	// 3. Test Commit Replay with Same Idempotency-Key -> 409 Conflict
	cReq2, _ := http.NewRequest(http.MethodPost, env.Server.URL+"/api/v1/sync/commit", bytes.NewBuffer(validCommitBody))
	cReq2.Header.Set("Authorization", "Bearer "+token)
	cReq2.Header.Set("Content-Type", "application/json")
	cReq2.Header.Set("Idempotency-Key", trxID)
	cResp2, err := http.DefaultClient.Do(cReq2)
	if err != nil {
		t.Fatalf("Replay commit request failed: %v", err)
	}
	if cResp2.StatusCode != http.StatusConflict {
		t.Errorf("Expected status 409 Conflict for idempotent replay, got %d", cResp2.StatusCode)
	}
	cResp2.Body.Close()

	// 4. Test 24-Hour Logical Duplicate Prevention (Same Warga NIK + Layanan ID while still 'menunggu_review') -> 400 Bad Request
	trxID2 := ulid.Make().String()
	dupLogicalBody, _ := json.Marshal(domain.SyncCommitRequest{
		TransaksiID: trxID2,
		WargaNIK:    wargaNIK, // same NIK
		LayananID:   "SKTM",   // same layanan
		DataIsian: json.RawMessage(`{
			"keperluan": "Pendaftaran Beasiswa Lain",
			"penghasilan_bulanan": 800000,
			"jumlah_tanggungan": 2
		}`),
	})

	cReq3, _ := http.NewRequest(http.MethodPost, env.Server.URL+"/api/v1/sync/commit", bytes.NewBuffer(dupLogicalBody))
	cReq3.Header.Set("Authorization", "Bearer "+token)
	cReq3.Header.Set("Content-Type", "application/json")
	cReq3.Header.Set("Idempotency-Key", trxID2)
	cResp3, err := http.DefaultClient.Do(cReq3)
	if err != nil {
		t.Fatalf("Duplicate logical commit request failed: %v", err)
	}
	if cResp3.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 Bad Request for 24h logical duplicate, got %d", cResp3.StatusCode)
	}
	cResp3.Body.Close()

	// 5. Test Schema Validation Failure (missing required field 'jumlah_tanggungan' in SKTM) -> 400 Bad Request
	trxID3 := ulid.Make().String()
	invalidSchemaBody, _ := json.Marshal(domain.SyncCommitRequest{
		TransaksiID: trxID3,
		WargaNIK:    "3205018888880002",
		LayananID:   "SKTM",
		DataIsian: json.RawMessage(`{
			"keperluan": "Pendaftaran Sekolah"
			// missing required fields penghasilan_bulanan & jumlah_tanggungan
		}`),
	})

	cReq4, _ := http.NewRequest(http.MethodPost, env.Server.URL+"/api/v1/sync/commit", bytes.NewBuffer(invalidSchemaBody))
	cReq4.Header.Set("Authorization", "Bearer "+token)
	cReq4.Header.Set("Content-Type", "application/json")
	cReq4.Header.Set("Idempotency-Key", trxID3)
	cResp4, err := http.DefaultClient.Do(cReq4)
	if err != nil {
		t.Fatalf("Invalid schema commit request failed: %v", err)
	}
	if cResp4.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 Bad Request for schema violation, got %d", cResp4.StatusCode)
	}
	cResp4.Body.Close()
}
