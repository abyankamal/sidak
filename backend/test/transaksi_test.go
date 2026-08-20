package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/abyankamal/sidak/backend/internal/domain"
	"github.com/oklog/ulid/v2"
)

func TestTransaksiReviewFlow(t *testing.T) {
	env := SetupTestEnv(t)

	// 1. Authenticate as KADER to insert a transaction
	kaderLogin, _ := json.Marshal(domain.LoginRequest{
		Identifier: "3205010303920003",
		Password:   "AdminSidak2026!",
	})
	kaderResp, err := http.Post(env.Server.URL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(kaderLogin))
	if err != nil || kaderResp.StatusCode != http.StatusOK {
		t.Fatalf("Login as kader failed: %v", err)
	}
	var kaderAuth domain.LoginResponse
	_ = json.NewDecoder(kaderResp.Body).Decode(&kaderAuth)
	kaderResp.Body.Close()

	trxID := ulid.Make().String()
	uniqueNIK := fmt.Sprintf("320599%010d", time.Now().UnixNano()%10000000000)
	commitBody, _ := json.Marshal(domain.SyncCommitRequest{
		TransaksiID: trxID,
		WargaNIK:    uniqueNIK,
		LayananID:   "SKTM",
		DataIsian:   json.RawMessage(`{"keperluan": "Bantuan Beras", "penghasilan_bulanan": 600000, "jumlah_tanggungan": 3}`),
		Lampiran:    []string{"uploads/lampiran/" + uniqueNIK + "/" + trxID + "_ktp.jpg"},
	})

	cReq, _ := http.NewRequest(http.MethodPost, env.Server.URL+"/api/v1/sync/commit", bytes.NewBuffer(commitBody))
	cReq.Header.Set("Authorization", "Bearer "+kaderAuth.Token)
	cReq.Header.Set("Idempotency-Key", trxID)
	cReq.Header.Set("Content-Type", "application/json")
	cResp, err := http.DefaultClient.Do(cReq)
	if err != nil || cResp.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create transaction for test: %v, status: %d", err, cResp.StatusCode)
	}
	cResp.Body.Close()

	// 2. Authenticate as SEKLUR
	seklurLogin, _ := json.Marshal(domain.LoginRequest{
		Identifier: "198001012005011001",
		Password:   "AdminSidak2026!",
	})
	seklurResp, err := http.Post(env.Server.URL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(seklurLogin))
	if err != nil || seklurResp.StatusCode != http.StatusOK {
		t.Fatalf("Login as seklur failed: %v", err)
	}
	var seklurAuth domain.LoginResponse
	_ = json.NewDecoder(seklurResp.Body).Decode(&seklurAuth)
	seklurResp.Body.Close()

	// 3. List Transactions as Seklur
	listReq, _ := http.NewRequest(http.MethodGet, env.Server.URL+"/api/v1/transaksi?limit=10", nil)
	listReq.Header.Set("Authorization", "Bearer "+seklurAuth.Token)
	listResp, err := http.DefaultClient.Do(listReq)
	if err != nil || listResp.StatusCode != http.StatusOK {
		t.Fatalf("List transaksi failed: %v, status: %d", err, listResp.StatusCode)
	}
	listResp.Body.Close()

	// 4. Get Detail Transaction as Seklur
	detailReq, _ := http.NewRequest(http.MethodGet, env.Server.URL+"/api/v1/transaksi/"+trxID, nil)
	detailReq.Header.Set("Authorization", "Bearer "+seklurAuth.Token)
	detailResp, err := http.DefaultClient.Do(detailReq)
	if err != nil || detailResp.StatusCode != http.StatusOK {
		t.Fatalf("Get detail transaksi failed: %v, status: %d", err, detailResp.StatusCode)
	}
	var detail domain.TransaksiDetailResponse
	_ = json.NewDecoder(detailResp.Body).Decode(&detail)
	detailResp.Body.Close()

	if detail.ID != trxID {
		t.Errorf("Expected ID %s, got %s", trxID, detail.ID)
	}
	if len(detail.LampiranURLs) != 1 {
		t.Errorf("Expected 1 Lampiran URL, got %d", len(detail.LampiranURLs))
	}

	// 5. Review Transaction (SEKLUR: sudah_di_review) -> 200 OK
	reviewBody, _ := json.Marshal(domain.ReviewTransaksiRequest{
		Status:        "sudah_di_review",
		CatatanReview: stringPtr("Berkas telah sesuai dengan kondisi riil di lapangan"),
	})
	reviewReq, _ := http.NewRequest(http.MethodPatch, env.Server.URL+"/api/v1/transaksi/"+trxID+"/review", bytes.NewBuffer(reviewBody))
	reviewReq.Header.Set("Authorization", "Bearer "+seklurAuth.Token)
	reviewReq.Header.Set("Content-Type", "application/json")
	reviewResp, err := http.DefaultClient.Do(reviewReq)
	if err != nil || reviewResp.StatusCode != http.StatusOK {
		t.Fatalf("Review transaksi failed: %v, status: %d", err, reviewResp.StatusCode)
	}
	reviewResp.Body.Close()

	// 6. Verify Transaction Status after Review
	detailReq2, _ := http.NewRequest(http.MethodGet, env.Server.URL+"/api/v1/transaksi/"+trxID, nil)
	detailReq2.Header.Set("Authorization", "Bearer "+seklurAuth.Token)
	detailResp2, _ := http.DefaultClient.Do(detailReq2)
	var detailUpdated domain.TransaksiDetailResponse
	_ = json.NewDecoder(detailResp2.Body).Decode(&detailUpdated)
	detailResp2.Body.Close()

	if detailUpdated.Status != "sudah_di_review" {
		t.Errorf("Expected status 'sudah_di_review', got '%s'", detailUpdated.Status)
	}

	// 7. Verify KADER cannot review (403 Forbidden)
	kaderReviewReq, _ := http.NewRequest(http.MethodPatch, env.Server.URL+"/api/v1/transaksi/"+trxID+"/review", bytes.NewBuffer(reviewBody))
	kaderReviewReq.Header.Set("Authorization", "Bearer "+kaderAuth.Token)
	kaderReviewReq.Header.Set("Content-Type", "application/json")
	kaderReviewResp, err := http.DefaultClient.Do(kaderReviewReq)
	if err != nil {
		t.Fatalf("Kader review request failed: %v", err)
	}
	if kaderReviewResp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected status 403 for Kader review, got %d", kaderReviewResp.StatusCode)
	}
	kaderReviewResp.Body.Close()
}

func stringPtr(s string) *string {
	return &s
}
