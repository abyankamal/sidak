package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/abyankamal/sidak/backend/internal/domain"
	"github.com/oklog/ulid/v2"
)

func TestPDFFlow(t *testing.T) {
	env := SetupTestEnv(t)

	// 1. Authenticate as KADER
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

	// 2. Submit a transaction
	trxID := ulid.Make().String()
	uniqueNIK := fmt.Sprintf("320599%010d", time.Now().UnixNano()%10000000000)
	commitBody, _ := json.Marshal(domain.SyncCommitRequest{
		TransaksiID: trxID,
		WargaNIK:    uniqueNIK,
		LayananID:   "SKTM",
		DataIsian:   json.RawMessage(`{"keperluan": "Pendaftaran Kartu Indonesia Pintar (KIP)", "penghasilan_bulanan": 750000, "jumlah_tanggungan": 3, "keterangan_pekerjaan": "Buruh Harian Lepas"}`),
	})

	cReq, _ := http.NewRequest(http.MethodPost, env.Server.URL+"/api/v1/sync/commit", bytes.NewBuffer(commitBody))
	cReq.Header.Set("Authorization", "Bearer "+kaderAuth.Token)
	cReq.Header.Set("Idempotency-Key", trxID)
	cReq.Header.Set("Content-Type", "application/json")
	cResp, err := http.DefaultClient.Do(cReq)
	if err != nil || cResp.StatusCode != http.StatusCreated {
		t.Fatalf("Create transaction failed: %v, status: %d", err, cResp.StatusCode)
	}
	cResp.Body.Close()

	// 3. Try to Generate PDF before review -> Must return 400 Bad Request
	genReqUnreviewed, _ := http.NewRequest(http.MethodPost, env.Server.URL+"/api/v1/layanan/"+trxID+"/generate-pdf", nil)
	genReqUnreviewed.Header.Set("Authorization", "Bearer "+kaderAuth.Token)
	genRespUnreviewed, err := http.DefaultClient.Do(genReqUnreviewed)
	if err != nil {
		t.Fatalf("Generate PDF request failed: %v", err)
	}
	if genRespUnreviewed.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request for unreviewed transaction PDF generation, got %d", genRespUnreviewed.StatusCode)
	}
	genRespUnreviewed.Body.Close()

	// 4. Authenticate as SEKLUR & Review Transaction
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

	reviewBody, _ := json.Marshal(domain.ReviewTransaksiRequest{
		Status:        "sudah_di_review",
		CatatanReview: stringPtr("Berkas telah diverifikasi lengkap dan valid"),
	})
	revReq, _ := http.NewRequest(http.MethodPatch, env.Server.URL+"/api/v1/transaksi/"+trxID+"/review", bytes.NewBuffer(reviewBody))
	revReq.Header.Set("Authorization", "Bearer "+seklurAuth.Token)
	revReq.Header.Set("Content-Type", "application/json")
	revResp, err := http.DefaultClient.Do(revReq)
	if err != nil || revResp.StatusCode != http.StatusOK {
		t.Fatalf("Review transaction failed: %v, status: %d", err, revResp.StatusCode)
	}
	revResp.Body.Close()

	// 5. Trigger PDF Generation -> 202 Accepted
	genReq, _ := http.NewRequest(http.MethodPost, env.Server.URL+"/api/v1/layanan/"+trxID+"/generate-pdf", nil)
	genReq.Header.Set("Authorization", "Bearer "+seklurAuth.Token)
	genResp, err := http.DefaultClient.Do(genReq)
	if err != nil || genResp.StatusCode != http.StatusAccepted {
		t.Fatalf("Generate PDF request failed: %v, status: %d", err, genResp.StatusCode)
	}

	var genResult map[string]any
	_ = json.NewDecoder(genResp.Body).Decode(&genResult)
	genResp.Body.Close()

	jobID, ok := genResult["job_id"].(string)
	if !ok || jobID == "" {
		t.Fatalf("Expected non-empty job_id in 202 response")
	}

	// 6. Poll Dokumen Status until READY (timeout 15 seconds)
	var finalStatus domain.DokumenStatusResponse
	deadline := time.Now().Add(15 * time.Second)
	isReady := false

	for time.Now().Before(deadline) {
		statusReq, _ := http.NewRequest(http.MethodGet, env.Server.URL+"/api/v1/dokumen/"+jobID+"/status", nil)
		statusReq.Header.Set("Authorization", "Bearer "+seklurAuth.Token)
		statusResp, err := http.DefaultClient.Do(statusReq)
		if err == nil && statusResp.StatusCode == http.StatusOK {
			_ = json.NewDecoder(statusResp.Body).Decode(&finalStatus)
			statusResp.Body.Close()

			if finalStatus.Status == "READY" {
				isReady = true
				break
			}
			if finalStatus.Status == "FAILED" {
				t.Fatalf("PDF conversion worker failed")
			}
		}
		time.Sleep(500 * time.Millisecond)
	}

	if !isReady {
		t.Fatalf("Timed out waiting for PDF worker to finish (status: %s)", finalStatus.Status)
	}

	if finalStatus.DownloadURL == nil || *finalStatus.DownloadURL == "" {
		t.Fatalf("Expected non-empty download_url when status is READY")
	}

	t.Logf("Generated PDF Download URL: %s", *finalStatus.DownloadURL)

	// 7. Verify downloading the PDF file
	pdfDownloadURL := *finalStatus.DownloadURL
	if strings.HasPrefix(pdfDownloadURL, "http://localhost:8080") {
		// Replace host/port with httptest server URL
		pdfDownloadURL = env.Server.URL + strings.TrimPrefix(pdfDownloadURL, "http://localhost:8080")
	}

	downloadResp, err := http.Get(pdfDownloadURL)
	if err != nil || downloadResp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to download generated PDF: %v, status: %d", err, downloadResp.StatusCode)
	}
	pdfBytes, err := io.ReadAll(downloadResp.Body)
	downloadResp.Body.Close()

	if len(pdfBytes) < 1000 {
		t.Errorf("Expected PDF size > 1000 bytes, got %d", len(pdfBytes))
	}
	if !bytes.HasPrefix(pdfBytes, []byte("%PDF-")) {
		t.Errorf("Expected valid PDF header starting with %%PDF-, got %q", string(pdfBytes[:min(10, len(pdfBytes))]))
	}

	t.Logf("Successfully verified generated PDF (%d bytes) with valid header!", len(pdfBytes))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
