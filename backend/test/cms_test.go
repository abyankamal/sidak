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

	// 1. Unauthenticated Public CMS Access (Profil)
	pResp, err := http.Get(env.Server.URL + "/api/v1/public/profil")
	if err != nil || pResp.StatusCode != http.StatusOK {
		t.Fatalf("Public profil endpoint failed: %v, status: %d", err, pResp.StatusCode)
	}
	var profil domain.ProfilWilayah
	_ = json.NewDecoder(pResp.Body).Decode(&profil)
	pResp.Body.Close()

	if profil.NamaKelurahan != "Sukanegla" {
		t.Errorf("Expected kelurahan Sukanegla, got %s", profil.NamaKelurahan)
	}

	// 2. Unauthenticated Public CMS Access (Menu Hierarchy)
	mResp, err := http.Get(env.Server.URL + "/api/v1/public/menu")
	if err != nil || mResp.StatusCode != http.StatusOK {
		t.Fatalf("Public menu endpoint failed: %v, status: %d", err, mResp.StatusCode)
	}
	var menu []domain.NavigasiMenuPublic
	_ = json.NewDecoder(mResp.Body).Decode(&menu)
	mResp.Body.Close()

	if len(menu) == 0 {
		t.Errorf("Expected non-empty public menu list")
	}

	// 3. Unauthenticated Public CMS Access (Konten List & Detail)
	kListResp, err := http.Get(env.Server.URL + "/api/v1/public/konten")
	if err != nil || kListResp.StatusCode != http.StatusOK {
		t.Fatalf("Public konten endpoint failed: %v, status: %d", err, kListResp.StatusCode)
	}
	var kList domain.KontenPublikListResponse
	_ = json.NewDecoder(kListResp.Body).Decode(&kList)
	kListResp.Body.Close()

	if len(kList.Data) == 0 {
		t.Errorf("Expected non-empty public konten list")
	}

	kDetailResp, err := http.Get(env.Server.URL + "/api/v1/public/konten/penyaluran-bansos-tahap-2")
	if err != nil || kDetailResp.StatusCode != http.StatusOK {
		t.Fatalf("Public konten detail failed: %v, status: %d", err, kDetailResp.StatusCode)
	}
	var kDetail domain.KontenPublikDetail
	_ = json.NewDecoder(kDetailResp.Body).Decode(&kDetail)
	kDetailResp.Body.Close()

	if kDetail.Slug != "penyaluran-bansos-tahap-2" {
		t.Errorf("Expected slug penyaluran-bansos-tahap-2, got %s", kDetail.Slug)
	}

	// 4. Test Protected Admin CMS without Token -> 401 Unauthorized
	admResp, err := http.Post(env.Server.URL+"/api/v1/cms/konten", "application/json", bytes.NewBuffer([]byte(`{}`)))
	if err != nil {
		t.Fatalf("Admin konten unauth failed: %v", err)
	}
	if admResp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized for admin CMS without auth, got %d", admResp.StatusCode)
	}
	admResp.Body.Close()

	// 5. Authenticate as SEKLUR
	seklurLogin, _ := json.Marshal(domain.LoginRequest{
		Identifier: "198001012005011001",
		Password:   "AdminSidak2026!",
	})
	loginResp, err := http.Post(env.Server.URL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(seklurLogin))
	if err != nil || loginResp.StatusCode != http.StatusOK {
		t.Fatalf("Login as seklur failed: %v", err)
	}
	var auth domain.LoginResponse
	_ = json.NewDecoder(loginResp.Body).Decode(&auth)
	loginResp.Body.Close()

	// 6. Create News as Admin -> 201 Created
	isPub := true
	newsInput, _ := json.Marshal(domain.KontenPublikInput{
		Tipe:              "BERITA",
		Judul:             "Gotong Royong Kebersihan Lingkungan",
		Ringkasan:         "Warga bersama perangkat kelurahan mengadakan kerja bakti massal.",
		IsiKonten:         "<p>Kerja bakti serentak diadakan di seluruh wilayah RW...</p>",
		ThumbnailFilePath: stringPtr("uploads/public/cms/2026/kerja_bakti.jpg"),
		IsPublished:       &isPub,
	})
	createNewsReq, _ := http.NewRequest(http.MethodPost, env.Server.URL+"/api/v1/cms/konten", bytes.NewBuffer(newsInput))
	createNewsReq.Header.Set("Authorization", "Bearer "+auth.Token)
	createNewsReq.Header.Set("Content-Type", "application/json")
	createNewsResp, err := http.DefaultClient.Do(createNewsReq)
	if err != nil || createNewsResp.StatusCode != http.StatusCreated {
		t.Fatalf("Create news failed: %v, status: %d", err, createNewsResp.StatusCode)
	}
	createNewsResp.Body.Close()

	// 7. Update Profil as Admin -> 200 OK
	profilInput, _ := json.Marshal(domain.ProfilWilayahInput{
		NamaKelurahan: "Sukanegla Terdepan",
		Kecamatan:     "Garut Kota",
		KabupatenKota: "Kabupaten Garut",
		Visi:          "Terwujudnya Pelayanan Digital Terdepan",
		Misi:          []string{"Misi 1", "Misi 2"},
		AlamatKantor:  "Jl. Baru No. 1",
	})
	updateProfReq, _ := http.NewRequest(http.MethodPut, env.Server.URL+"/api/v1/cms/profil", bytes.NewBuffer(profilInput))
	updateProfReq.Header.Set("Authorization", "Bearer "+auth.Token)
	updateProfReq.Header.Set("Content-Type", "application/json")
	updateProfResp, err := http.DefaultClient.Do(updateProfReq)
	if err != nil || updateProfResp.StatusCode != http.StatusOK {
		t.Fatalf("Update profil failed: %v, status: %d", err, updateProfResp.StatusCode)
	}
	updateProfResp.Body.Close()

	// 8. Create Menu as Admin -> 201 Created
	menuInput, _ := json.Marshal(domain.NavigasiMenuInput{
		Label:  "Galeri Kegiatan",
		URL:    "/galeri",
		Urutan: 6,
	})
	createMenuReq, _ := http.NewRequest(http.MethodPost, env.Server.URL+"/api/v1/cms/menu", bytes.NewBuffer(menuInput))
	createMenuReq.Header.Set("Authorization", "Bearer "+auth.Token)
	createMenuReq.Header.Set("Content-Type", "application/json")
	createMenuResp, err := http.DefaultClient.Do(createMenuReq)
	if err != nil || createMenuResp.StatusCode != http.StatusCreated {
		t.Fatalf("Create menu failed: %v, status: %d", err, createMenuResp.StatusCode)
	}
	createMenuResp.Body.Close()
}
