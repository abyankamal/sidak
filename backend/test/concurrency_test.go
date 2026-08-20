package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/abyankamal/sidak/backend/internal/domain"
	"github.com/oklog/ulid/v2"
)

func TestConcurrencyInFlightLock(t *testing.T) {
	env := SetupTestEnv(t)

	// 1. Authenticate as KADER
	kaderLogin, _ := json.Marshal(domain.LoginRequest{
		Identifier: "3205010303920003",
		Password:   "AdminSidak2026!",
	})
	loginResp, err := http.Post(env.Server.URL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(kaderLogin))
	if err != nil || loginResp.StatusCode != http.StatusOK {
		t.Fatalf("Login failed: %v", err)
	}
	var auth domain.LoginResponse
	_ = json.NewDecoder(loginResp.Body).Decode(&auth)
	loginResp.Body.Close()

	// 2. Launch 10 simultaneous commit requests with identical ULID
	identicalULID := ulid.Make().String()
	uniqueNIK := fmt.Sprintf("320599%010d", time.Now().UnixNano()%10000000000)
	commitBody, _ := json.Marshal(domain.SyncCommitRequest{
		TransaksiID: identicalULID,
		WargaNIK:    uniqueNIK,
		LayananID:   "SKTM",
		DataIsian:   json.RawMessage(`{"keperluan": "Bansos Sembako", "penghasilan_bulanan": 450000, "jumlah_tanggungan": 4}`),
	})

	const concurrentWorkers = 10
	var wg sync.WaitGroup
	statusCodeCount := make(map[int]int)
	var mu sync.Mutex

	for i := 0; i < concurrentWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			req, err := http.NewRequest(http.MethodPost, env.Server.URL+"/api/v1/sync/commit", bytes.NewBuffer(commitBody))
			if err != nil {
				return
			}
			req.Header.Set("Authorization", "Bearer "+auth.Token)
			req.Header.Set("Idempotency-Key", identicalULID)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			mu.Lock()
			statusCodeCount[resp.StatusCode]++
			mu.Unlock()
		}()
	}

	wg.Wait()

	t.Logf("Concurrency results: 201 Created=%d, 409 Conflict=%d, 429 TooManyRequests=%d",
		statusCodeCount[http.StatusCreated],
		statusCodeCount[http.StatusConflict],
		statusCodeCount[http.StatusTooManyRequests],
	)

	// Exactly 1 request should be Created (201)
	if statusCodeCount[http.StatusCreated] != 1 {
		t.Errorf("Expected exactly 1 request to return 201 Created, got %d", statusCodeCount[http.StatusCreated])
	}

	// Concurrent in-flight collisions must return 429 or 409
	if statusCodeCount[http.StatusTooManyRequests]+statusCodeCount[http.StatusConflict] != concurrentWorkers-1 {
		t.Errorf("Expected remaining %d requests to be 429 or 409, got breakdown: %+v",
			concurrentWorkers-1, statusCodeCount)
	}
}
