package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/abyankamal/sidak/backend/internal/service"
	"github.com/go-chi/chi/v5"
)

type CMSPublicHandler struct {
	cmsService *service.CMSService
}

func NewCMSPublicHandler(cmsService *service.CMSService) *CMSPublicHandler {
	return &CMSPublicHandler{cmsService: cmsService}
}

func (h *CMSPublicHandler) GetProfil(w http.ResponseWriter, r *http.Request) {
	resp, err := h.cmsService.GetPublicProfil(r.Context())
	if err != nil {
		Error(w, http.StatusInternalServerError, "Gagal mengambil profil wilayah")
		return
	}
	if resp == nil {
		Error(w, http.StatusNotFound, "Data profil wilayah belum tersedia")
		return
	}
	JSON(w, http.StatusOK, resp)
}

func (h *CMSPublicHandler) GetMenu(w http.ResponseWriter, r *http.Request) {
	resp, err := h.cmsService.GetPublicMenu(r.Context())
	if err != nil {
		Error(w, http.StatusInternalServerError, "Gagal mengambil struktur menu navigasi")
		return
	}
	JSON(w, http.StatusOK, resp)
}

func (h *CMSPublicHandler) GetKontenList(w http.ResponseWriter, r *http.Request) {
	tipe := r.URL.Query().Get("tipe")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	resp, err := h.cmsService.GetPublicKontenList(r.Context(), tipe, page, limit)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Gagal mengambil daftar berita/pengumuman")
		return
	}
	JSON(w, http.StatusOK, resp)
}

func (h *CMSPublicHandler) GetKontenDetail(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		Error(w, http.StatusBadRequest, "Slug konten wajib disertakan")
		return
	}

	resp, err := h.cmsService.GetPublicKontenDetail(r.Context(), slug)
	if err != nil {
		if errors.Is(err, service.ErrKontenNotFound) {
			Error(w, http.StatusNotFound, "Berita atau pengumuman tidak ditemukan")
			return
		}
		Error(w, http.StatusInternalServerError, "Gagal mengambil detail berita/pengumuman")
		return
	}
	JSON(w, http.StatusOK, resp)
}
