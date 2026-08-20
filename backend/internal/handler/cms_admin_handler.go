package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/abyankamal/sidak/backend/internal/domain"
	"github.com/abyankamal/sidak/backend/internal/middleware"
	"github.com/abyankamal/sidak/backend/internal/service"
	"github.com/go-chi/chi/v5"
)

type CMSAdminHandler struct {
	cmsService *service.CMSService
}

func NewCMSAdminHandler(cmsService *service.CMSService) *CMSAdminHandler {
	return &CMSAdminHandler{cmsService: cmsService}
}

// -----------------------------------------------------------------------------
// Profil
// -----------------------------------------------------------------------------

func (h *CMSAdminHandler) UpdateProfil(w http.ResponseWriter, r *http.Request) {
	var input domain.ProfilWilayahInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		Error(w, http.StatusBadRequest, "Payload update profil wilayah tidak valid")
		return
	}

	if input.NamaKelurahan == "" || input.Kecamatan == "" || input.KabupatenKota == "" || input.Visi == "" || input.AlamatKantor == "" {
		Error(w, http.StatusBadRequest, "Field profil wajib (nama_kelurahan, kecamatan, kabupaten_kota, visi, alamat_kantor) harus diisi")
		return
	}

	resp, err := h.cmsService.UpdateProfil(r.Context(), input)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Gagal memperbarui profil wilayah")
		return
	}

	JSON(w, http.StatusOK, resp)
}

// -----------------------------------------------------------------------------
// Menu Navigasi
// -----------------------------------------------------------------------------

func (h *CMSAdminHandler) ListMenu(w http.ResponseWriter, r *http.Request) {
	resp, err := h.cmsService.GetAllMenuAdmin(r.Context())
	if err != nil {
		Error(w, http.StatusInternalServerError, "Gagal mengambil daftar menu navigasi admin")
		return
	}
	JSON(w, http.StatusOK, resp)
}

func (h *CMSAdminHandler) CreateMenu(w http.ResponseWriter, r *http.Request) {
	var input domain.NavigasiMenuInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		Error(w, http.StatusBadRequest, "Payload create menu tidak valid")
		return
	}

	if input.Label == "" || input.URL == "" {
		Error(w, http.StatusBadRequest, "Label dan URL menu wajib diisi")
		return
	}

	resp, err := h.cmsService.CreateMenu(r.Context(), input)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Gagal menambahkan item menu navigasi")
		return
	}

	JSON(w, http.StatusCreated, resp)
}

func (h *CMSAdminHandler) UpdateMenu(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		Error(w, http.StatusBadRequest, "ID menu wajib disertakan")
		return
	}

	var input domain.NavigasiMenuInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		Error(w, http.StatusBadRequest, "Payload update menu tidak valid")
		return
	}

	resp, err := h.cmsService.UpdateMenu(r.Context(), id, input)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Gagal memperbarui item menu navigasi")
		return
	}

	JSON(w, http.StatusOK, resp)
}

func (h *CMSAdminHandler) DeleteMenu(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		Error(w, http.StatusBadRequest, "ID menu wajib disertakan")
		return
	}

	resp, err := h.cmsService.DeleteMenu(r.Context(), id)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Gagal menghapus item menu navigasi")
		return
	}

	JSON(w, http.StatusOK, resp)
}

// -----------------------------------------------------------------------------
// Konten Publik (Berita & Pengumuman)
// -----------------------------------------------------------------------------

func (h *CMSAdminHandler) ListKonten(w http.ResponseWriter, r *http.Request) {
	tipe := r.URL.Query().Get("tipe")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	resp, err := h.cmsService.GetAdminKontenList(r.Context(), tipe, page, limit)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Gagal mengambil daftar konten admin")
		return
	}
	JSON(w, http.StatusOK, resp)
}

func (h *CMSAdminHandler) CreateKonten(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	if claims == nil {
		Error(w, http.StatusUnauthorized, "Sesi tidak sah")
		return
	}

	var input domain.KontenPublikInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		Error(w, http.StatusBadRequest, "Payload create konten tidak valid")
		return
	}

	if input.Tipe == "" || input.Judul == "" || input.Ringkasan == "" || input.IsiKonten == "" {
		Error(w, http.StatusBadRequest, "Tipe, Judul, Ringkasan, dan Isi Konten wajib diisi")
		return
	}

	resp, err := h.cmsService.CreateKonten(r.Context(), input, claims.UserID, claims.Nama)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Gagal membuat konten publik")
		return
	}

	JSON(w, http.StatusCreated, resp)
}

func (h *CMSAdminHandler) UpdateKonten(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		Error(w, http.StatusBadRequest, "ID konten wajib disertakan")
		return
	}

	var input domain.KontenPublikInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		Error(w, http.StatusBadRequest, "Payload update konten tidak valid")
		return
	}

	resp, err := h.cmsService.UpdateKonten(r.Context(), id, input)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Gagal memperbarui konten publik")
		return
	}

	JSON(w, http.StatusOK, resp)
}

func (h *CMSAdminHandler) DeleteKonten(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		Error(w, http.StatusBadRequest, "ID konten wajib disertakan")
		return
	}

	resp, err := h.cmsService.DeleteKonten(r.Context(), id)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Gagal menghapus konten publik")
		return
	}

	JSON(w, http.StatusOK, resp)
}

func (h *CMSAdminHandler) PresignedURL(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FileName    string `json:"file_name"`
		ContentType string `json:"content_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Payload presigned URL tidak valid")
		return
	}

	if req.FileName == "" || req.ContentType == "" {
		Error(w, http.StatusBadRequest, "file_name dan content_type wajib diisi")
		return
	}

	resp, err := h.cmsService.GenerateMediaPresignedURL(r.Context(), req.FileName, req.ContentType)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Gagal membuat presigned upload URL media")
		return
	}

	JSON(w, http.StatusOK, resp)
}
