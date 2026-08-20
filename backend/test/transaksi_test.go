package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/abyankamal/sidak/backend/internal/domain"
	"github.com/oklog/ulid/v2"
)

func TestTransaksiReviewFlow(t *testing.T) {
	env := SetupTestEnv(t)

	// 1. Create a transaction as KADER
	kaderLogin, _ := json.Marshal(domain.LoginRequest{
		NIK:      "3205010303920003",
		Password: "AdminSidak2026!",
	})
	kResp, _ := http.Post(env.Server.URL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(kaderLogin))
	var kLoginResp domain.LoginResponse
	_ = json.NewDecoder(kResp.Body).Decode(&kLoginResp)
	kResp.Body.Close()

	trxID := ulid.Make().String()
	trxPayload, _ := json.Marshal(domain.SyncCommitRequest{
		TransaksiID: trxID,
		WargaNIK:    "3205016666660004",
		LayananID:   "SK_BELUM_MENIKAH",
		DataIsian: json.RawMessage(`{
			"keperluan": "Persyaratan Melamar Pekerjaan BUMN"
		}`),
		Lampiran: []string{"lampiran/3205016666660004/" + trxID + "_kk.jpg"},
	})

	cReq, _ := http.NewRequest(http.MethodPost, env.Server.URL+"/api/v1/sync/commit", bytes.NewBuffer(trxPayload))
	cReq.Header.Set("Authorization", "Bearer "+kLoginResp.Token)
	cReq.Header.Set("Content-Type", "application/json")
	cReq.Header.Set("Idempotency-Key", trxID)
	cResp, err := http.DefaultClient.Do(cReq)
	if err != nil || cResp.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create test transaction: %v (Status: %d)", err, cResp.StatusCode)
	}
	cResp.Body.Close()

	// 2. KASI Login
	kasiLogin, _ := json.Marshal(domain.LoginRequest{
		NIK:      "3205010202850002",
		Password: "AdminSidak2026!",
	})
	kasiResp, _ := http.Post(env.Server.URL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(kasiLogin))
	var kasiLoginResp domain.LoginResponse
	_ = json.NewDecoder(kasiResp.Body).Decode(&kasiLoginResp)
	kasiResp.Body.Close()

	// 3. KASI lists transactions
	listReq, _ := http.NewRequest(http.MethodGet, env.Server.URL+"/api/v1/transaksi?limit=10", nil)
	listReq.Header.Set("Authorization", "Bearer "+kasiLoginResp.Token)
	listResp, err := http.DefaultClient.Do(listReq)
	if err != nil || listResp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to list transactions: %v", err)
	}
	var listData domain.TransaksiListResponse
	_ = json.NewDecoder(listResp.Body).Decode(&listData)
	listResp.Body.Close()

	if len(listData.Data) == 0 {
		t.Errorf("Expected at least 1 transaction in list")
	}

	// 4. KASI gets detail
	detailReq, _ := http.NewRequest(http.MethodGet, env.Server.URL+"/api/v1/transaksi/"+trxID, nil)
	detailReq.Header.Set("Authorization", "Bearer "+kasiLoginResp.Token)
	detailResp, err := http.DefaultClient.Do(detailReq)
	if err != nil || detailResp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to get detail transaction: %v", err)
	}
	var detailData domain.TransaksiDetailResponse
	_ = json.NewDecoder(detailResp.Body).Decode(&detailData)
	detailResp.Body.Close()

	if detailData.Status != "menunggu_review" {
		t.Errorf("Expected initial status 'menunggu_review', got %s", detailData.Status)
	}
	if len(detailData.LampiranPresignedURLs) != 1 {
		t.Errorf("Expected 1 lampiran url, got %d", len(detailData.LampiranPresignedURLs))
	}

	// 5. KASI reviews transaction -> 'sudah_di_review'
	catatan := "Data telah diverifikasi valid dan sesuai dokumen warga"
	reviewPayload, _ := json.Marshal(domain.ReviewTransaksiRequest{
		Status:        "sudah_di_review",
		CatatanReview: &catatan,
	})
	reviewReq, _ := http.NewRequest(http.MethodPatch, env.Server.URL+"/api/v1/transaksi/"+trxID+"/review", bytes.NewBuffer(reviewPayload))
	reviewReq.Header.Set("Authorization", "Bearer "+kasiLoginResp.Token)
	reviewReq.Header.Set("Content-Type", "application/json")
	reviewResp, err := http.DefaultClient.Do(reviewReq)
	if err != nil || reviewResp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to review transaction: %v", err)
	}
	reviewResp.Body.Close()

	// 6. Verify status updated
	detailReq2, _ := http.NewRequest(http.MethodGet, env.Server.URL+"/api/v1/transaksi/"+trxID, nil)
	detailReq2.Header.Set("Authorization", "Bearer "+kasiLoginResp.Token)
	detailResp2, _ := http.DefaultClient.Do(detailReq2)
	var detailData2 domain.TransaksiDetailResponse
	_ = json.NewDecoder(detailResp2.Body).Decode(&detailData2)
	detailResp2.Body.Close()

	if detailData2.Status != "sudah_di_review" {
		t.Errorf("Expected updated status 'sudah_di_review', got %s", detailData2.Status)
	}
	if detailData2.ReviewedBy == nil || *detailData2.ReviewedBy != kasiLoginResp.User.ID {
		t.Errorf("Expected reviewed_by to match KASI user ID")
	}

	// 7. Test KADER attempting review -> 403 Forbidden
	kaderReviewReq, _ := http.NewRequest(http.MethodPatch, env.Server.URL+"/api/v1/transaksi/"+trxID+"/review", bytes.NewBuffer(reviewPayload))
	kaderReviewReq.Header.Set("Authorization", "Bearer "+kLoginResp.Token)
	kaderReviewReq.Header.Set("Content-Type", "application/json")
	kaderReviewResp, err := http.DefaultClient.Do(kaderReviewReq)
	if err != nil {
		t.Fatalf("Kader review request failed: %v", err)
	}
	if kaderReviewResp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden for KADER attempting review, got %d", kaderReviewResp.StatusCode)
	}
	kaderReviewResp.Body.Close()
}
