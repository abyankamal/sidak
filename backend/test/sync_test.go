package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"testing"
	"time"

	"github.com/abyankamal/sidak/backend/internal/domain"
	"github.com/oklog/ulid/v2"
)

func TestSyncFlow(t *testing.T) {
	env := SetupTestEnv(t)

	// 1. Authenticate as KADER
	kaderLogin, _ := json.Marshal(domain.LoginRequest{
		Identifier: "3205010303920003",
		Password:   "AdminSidak2026!",
	})
	loginResp, err := http.Post(env.Server.URL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(kaderLogin))
	if err != nil || loginResp.StatusCode != http.StatusOK {
		t.Fatalf("Login as kader failed: %v", err)
	}
	var auth domain.LoginResponse
	_ = json.NewDecoder(loginResp.Body).Decode(&auth)
	loginResp.Body.Close()

	transaksiID := ulid.Make().String()
	uniqueNIK := fmt.Sprintf("320599%010d", time.Now().UnixNano()%10000000000)

	// 2. Test Local Storage Upload (/api/v1/storage/upload)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("category", "lampiran")
	_ = writer.WriteField("transaksi_id", transaksiID)
	part, _ := writer.CreateFormFile("file", "ktp_warga.jpg")
	_, _ = io.WriteString(part, "fake_image_bytes_for_ktp")
	_ = writer.Close()

	uploadReq, _ := http.NewRequest(http.MethodPost, env.Server.URL+"/api/v1/storage/upload", body)
	uploadReq.Header.Set("Authorization", "Bearer "+auth.Token)
	uploadReq.Header.Set("Content-Type", writer.FormDataContentType())

	uploadResp, err := http.DefaultClient.Do(uploadReq)
	if err != nil {
		t.Fatalf("Storage upload request failed: %v", err)
	}
	if uploadResp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status 201 for storage upload, got %d", uploadResp.StatusCode)
	}
	var fileResp domain.FileUploadResponse
	_ = json.NewDecoder(uploadResp.Body).Decode(&fileResp)
	uploadResp.Body.Close()

	if fileResp.FilePath == "" || fileResp.FileURL == "" {
		t.Fatalf("Expected non-empty FilePath and FileURL")
	}

	// 3. Test Commit (/api/v1/sync/commit) -> 201 Created
	commitBody, _ := json.Marshal(domain.SyncCommitRequest{
		TransaksiID: transaksiID,
		WargaNIK:    uniqueNIK,
		LayananID:   "SKTM",
		DataIsian:   json.RawMessage(`{"keperluan": "Beasiswa Kuliah", "penghasilan_bulanan": 800000, "jumlah_tanggungan": 2}`),
		Lampiran:    []string{fileResp.FilePath},
	})

	req, _ := http.NewRequest(http.MethodPost, env.Server.URL+"/api/v1/sync/commit", bytes.NewBuffer(commitBody))
	req.Header.Set("Authorization", "Bearer "+auth.Token)
	req.Header.Set("Idempotency-Key", transaksiID)
	req.Header.Set("Content-Type", "application/json")

	commitResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Commit request failed: %v", err)
	}
	if commitResp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status 201 for first commit, got %d", commitResp.StatusCode)
	}
	commitResp.Body.Close()

	// 4. Test Idempotent Duplicate Replay (same Idempotency-Key) -> 409 Conflict
	reqDup, _ := http.NewRequest(http.MethodPost, env.Server.URL+"/api/v1/sync/commit", bytes.NewBuffer(commitBody))
	reqDup.Header.Set("Authorization", "Bearer "+auth.Token)
	reqDup.Header.Set("Idempotency-Key", transaksiID)
	reqDup.Header.Set("Content-Type", "application/json")

	dupResp, err := http.DefaultClient.Do(reqDup)
	if err != nil {
		t.Fatalf("Duplicate commit request failed: %v", err)
	}
	if dupResp.StatusCode != http.StatusConflict {
		t.Errorf("Expected status 409 for duplicate replay, got %d", dupResp.StatusCode)
	}
	dupResp.Body.Close()

	// 5. Test Logical Duplicate (different ULID, same NIK + LayananID within 24h) -> 400 Bad Request
	newTrxID := ulid.Make().String()
	commitBodyLogicalDup, _ := json.Marshal(domain.SyncCommitRequest{
		TransaksiID: newTrxID,
		WargaNIK:    uniqueNIK,
		LayananID:   "SKTM",
		DataIsian:   json.RawMessage(`{"keperluan": "Keringanan UKT", "penghasilan_bulanan": 850000, "jumlah_tanggungan": 2}`),
	})
	reqLogicalDup, _ := http.NewRequest(http.MethodPost, env.Server.URL+"/api/v1/sync/commit", bytes.NewBuffer(commitBodyLogicalDup))
	reqLogicalDup.Header.Set("Authorization", "Bearer "+auth.Token)
	reqLogicalDup.Header.Set("Idempotency-Key", newTrxID)
	reqLogicalDup.Header.Set("Content-Type", "application/json")

	logDupResp, err := http.DefaultClient.Do(reqLogicalDup)
	if err != nil {
		t.Fatalf("Logical duplicate commit request failed: %v", err)
	}
	if logDupResp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 for 24h logical duplicate, got %d", logDupResp.StatusCode)
	}
	logDupResp.Body.Close()

	// 6. Test JSON Schema Validation Failure (missing required field) -> 400 Bad Request
	schemaFailID := ulid.Make().String()
	otherNIK := fmt.Sprintf("320599%010d", (time.Now().UnixNano()+999)%10000000000)
	commitBodyInvalidSchema, _ := json.Marshal(domain.SyncCommitRequest{
		TransaksiID: schemaFailID,
		WargaNIK:    otherNIK,
		LayananID:   "SKTM",
		DataIsian:   json.RawMessage(`{"penghasilan_bulanan": 800000}`), // missing 'keperluan' and 'jumlah_tanggungan'
	})
	reqSchemaFail, _ := http.NewRequest(http.MethodPost, env.Server.URL+"/api/v1/sync/commit", bytes.NewBuffer(commitBodyInvalidSchema))
	reqSchemaFail.Header.Set("Authorization", "Bearer "+auth.Token)
	reqSchemaFail.Header.Set("Idempotency-Key", schemaFailID)
	reqSchemaFail.Header.Set("Content-Type", "application/json")

	schemaFailResp, err := http.DefaultClient.Do(reqSchemaFail)
	if err != nil {
		t.Fatalf("Invalid schema commit request failed: %v", err)
	}
	if schemaFailResp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid JSON Schema, got %d", schemaFailResp.StatusCode)
	}
	schemaFailResp.Body.Close()
}
