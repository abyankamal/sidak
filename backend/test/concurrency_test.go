package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/abyankamal/sidak/backend/internal/domain"
	"github.com/oklog/ulid/v2"
)

func TestConcurrencyInFlightLock(t *testing.T) {
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

	// Fire 10 concurrent requests with the SAME Idempotency-Key
	sharedTrxID := ulid.Make().String()
	payload, _ := json.Marshal(domain.SyncCommitRequest{
		TransaksiID: sharedTrxID,
		WargaNIK:    "3205017777770003",
		LayananID:   "SK_DOMISILI",
		DataIsian: json.RawMessage(`{
			"nama_usaha": "Warung Berkah Jaya",
			"jenis_usaha": "Kuliner",
			"alamat_usaha": "Jl. Mawar No. 12 RT 01 RW 02",
			"lama_usaha_tahun": 2
		}`),
	})

	var wg sync.WaitGroup
	var count429 int32
	var count201 int32
	var count409 int32

	concurrentClients := 10
	wg.Add(concurrentClients)

	for i := 0; i < concurrentClients; i++ {
		go func() {
			defer wg.Done()
			req, err := http.NewRequest(http.MethodPost, env.Server.URL+"/api/v1/sync/commit", bytes.NewBuffer(payload))
			if err != nil {
				return
			}
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Idempotency-Key", sharedTrxID)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			switch resp.StatusCode {
			case http.StatusCreated:
				atomic.AddInt32(&count201, 1)
			case http.StatusConflict:
				atomic.AddInt32(&count409, 1)
			case http.StatusTooManyRequests:
				atomic.AddInt32(&count429, 1)
			}
		}()
	}

	wg.Wait()

	t.Logf("Concurrency results: 201 Created=%d, 409 Conflict=%d, 429 TooManyRequests=%d", count201, count409, count429)

	// Exactly 1 request must have succeeded in creating the initial transaction (201)
	if count201 != 1 {
		t.Errorf("Expected exactly 1 request to return 201 Created, got %d", count201)
	}

	// The remaining requests must be either 429 (caught in-flight) or 409 (replayed after commit)
	totalHandled := count201 + count409 + count429
	if totalHandled != int32(concurrentClients) {
		t.Errorf("Expected all %d requests to be handled, got %d", concurrentClients, totalHandled)
	}
}
