package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/abyankamal/sidak/backend/internal/domain"
)

func TestCMSFlow(t *testing.T) {
	env := SetupTestEnv(t)

	// 1. CMS Public - Unauthenticated GET /public/profil -> 200 OK
	pResp, err := http.Get(env.Server.URL + "/api/v1/public/profil")
	if err != nil || pResp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to fetch public profil: %v (Status: %d)", err, pResp.StatusCode)
	}
	var profil domain.ProfilWilayah
	_ = json.NewDecoder(pResp.Body).Decode(&profil)
	pResp.Body.Close()

	if profil.NamaKelurahan != "Sukanegla" {
		t.Errorf("Expected nama kelurahan 'Sukanegla', got '%s'", profil.NamaKelurahan)
	}

	// 2. CMS Public - Unauthenticated GET /public/menu -> 200 OK
	mResp, err := http.Get(env.Server.URL + "/api/v1/public/menu")
	if err != nil || mResp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to fetch public menu: %v", err)
	}
	var menuList []domain.NavigasiMenuPublic
	_ = json.NewDecoder(mResp.Body).Decode(&menuList)
	mResp.Body.Close()

	if len(menuList) == 0 {
		t.Errorf("Expected at least 1 public menu")
	}

	// 3. CMS Public - Unauthenticated GET /public/konten -> 200 OK
	kResp, err := http.Get(env.Server.URL + "/api/v1/public/konten")
	if err != nil || kResp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to fetch public konten: %v", err)
	}
	var kontenList domain.KontenPublikListResponse
	_ = json.NewDecoder(kResp.Body).Decode(&kontenList)
	kResp.Body.Close()

	if len(kontenList.Data) == 0 {
		t.Errorf("Expected at least 1 public article/announcement")
	}

	// 4. CMS Public - GET /public/konten/penyaluran-bansos-tahap-2 -> 200 OK
	kdResp, err := http.Get(env.Server.URL + "/api/v1/public/konten/penyaluran-bansos-tahap-2")
	if err != nil || kdResp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to fetch public konten detail: %v", err)
	}
	var detail domain.KontenPublikDetail
	_ = json.NewDecoder(kdResp.Body).Decode(&detail)
	kdResp.Body.Close()

	if detail.Slug != "penyaluran-bansos-tahap-2" {
		t.Errorf("Expected slug 'penyaluran-bansos-tahap-2', got '%s'", detail.Slug)
	}

	// 5. CMS Admin - Unauthenticated POST /cms/konten -> 401 Unauthorized
	unauthReq, _ := http.NewRequest(http.MethodPost, env.Server.URL+"/api/v1/cms/konten", bytes.NewBuffer([]byte("{}")))
	unauthResp, err := http.DefaultClient.Do(unauthReq)
	if err != nil || unauthResp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized for unauthenticated admin access, got %d", unauthResp.StatusCode)
	}
	unauthResp.Body.Close()

	// 6. Login as SEKLUR to perform Admin CMS operations
	seklurLogin, _ := json.Marshal(domain.LoginRequest{
		NIK:      "3205010101800001",
		Password: "AdminSidak2026!",
	})
	sResp, _ := http.Post(env.Server.URL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(seklurLogin))
	var sLoginResp domain.LoginResponse
	_ = json.NewDecoder(sResp.Body).Decode(&sLoginResp)
	sResp.Body.Close()

	// 7. SEKLUR creates a new news article via POST /cms/konten -> 201 Created
	isPub := true
	createKontenBody, _ := json.Marshal(domain.KontenPublikInput{
		Tipe:        "BERITA",
		Judul:       "Pembangunan Taman Baca Masyarakat RW 03 Sukanegla",
		Ringkasan:   "Fasilitas ruang literasi terbuka warga RW 03 telah rampung dibangun.",
		IsiKonten:   "<p>Pembangunan fasilitas ruang literasi publik yang ramah anak dan keluarga telah selesai 100%.</p>",
		IsPublished: &isPub,
	})

	ckReq, _ := http.NewRequest(http.MethodPost, env.Server.URL+"/api/v1/cms/konten", bytes.NewBuffer(createKontenBody))
	ckReq.Header.Set("Authorization", "Bearer "+sLoginResp.Token)
	ckReq.Header.Set("Content-Type", "application/json")
	ckResp, err := http.DefaultClient.Do(ckReq)
	if err != nil || ckResp.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create news article: %v (Status: %d)", err, ckResp.StatusCode)
	}
	ckResp.Body.Close()

	// 8. SEKLUR updates Profil Wilayah via PUT /cms/profil -> 200 OK
	updateProfilBody, _ := json.Marshal(domain.ProfilWilayahInput{
		NamaKelurahan: "Sukanegla",
		Kecamatan:     "Garut Kota",
		KabupatenKota: "Kabupaten Garut",
		Visi:          "Terwujudnya Pelayanan Kelurahan Sukanegla yang Unggul dan Modern.",
		Misi:          []string{"Misi 1: Digitalisasi Layanan", "Misi 2: Partisipasi Warga"},
		AlamatKantor:  "Jl. Sukanegla Raya No. 45",
	})
	upReq, _ := http.NewRequest(http.MethodPut, env.Server.URL+"/api/v1/cms/profil", bytes.NewBuffer(updateProfilBody))
	upReq.Header.Set("Authorization", "Bearer "+sLoginResp.Token)
	upReq.Header.Set("Content-Type", "application/json")
	upResp, err := http.DefaultClient.Do(upReq)
	if err != nil || upResp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to update profil: %v (Status: %d)", err, upResp.StatusCode)
	}
	upResp.Body.Close()

	// 9. SEKLUR creates new Menu Item via POST /cms/menu -> 201 Created
	createMenuBody, _ := json.Marshal(domain.NavigasiMenuInput{
		Label:  "Galeri Kegiatan",
		URL:    "/galeri",
		Urutan: 6,
	})
	cmReq, _ := http.NewRequest(http.MethodPost, env.Server.URL+"/api/v1/cms/menu", bytes.NewBuffer(createMenuBody))
	cmReq.Header.Set("Authorization", "Bearer "+sLoginResp.Token)
	cmReq.Header.Set("Content-Type", "application/json")
	cmResp, err := http.DefaultClient.Do(cmReq)
	if err != nil || cmResp.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create menu item: %v (Status: %d)", err, cmResp.StatusCode)
	}
	cmResp.Body.Close()
}
