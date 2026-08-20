package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/abyankamal/sidak/backend/internal/domain"
	"github.com/abyankamal/sidak/backend/internal/middleware"
	"github.com/abyankamal/sidak/backend/internal/service"
	"github.com/go-chi/chi/v5"
)

type TransaksiHandler struct {
	transaksiService *service.TransaksiService
}

func NewTransaksiHandler(transaksiService *service.TransaksiService) *TransaksiHandler {
	return &TransaksiHandler{transaksiService: transaksiService}
}

func (h *TransaksiHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	filter := domain.TransaksiFilter{
		Status:    r.URL.Query().Get("status"),
		LayananID: r.URL.Query().Get("layanan_id"),
		Page:      page,
		Limit:     limit,
	}

	resp, err := h.transaksiService.List(r.Context(), filter)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Gagal mengambil daftar antrean transaksi")
		return
	}

	JSON(w, http.StatusOK, resp)
}

func (h *TransaksiHandler) GetDetail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		Error(w, http.StatusBadRequest, "ID transaksi wajib diisi")
		return
	}

	resp, err := h.transaksiService.GetDetail(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrTransaksiNotFound) {
			Error(w, http.StatusNotFound, "Transaksi tidak ditemukan")
			return
		}
		Error(w, http.StatusInternalServerError, "Gagal mengambil detail transaksi")
		return
	}

	JSON(w, http.StatusOK, resp)
}

func (h *TransaksiHandler) Review(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		Error(w, http.StatusBadRequest, "ID transaksi wajib diisi")
		return
	}

	claims := middleware.GetUserClaims(r)
	if claims == nil {
		Error(w, http.StatusUnauthorized, "Sesi tidak sah")
		return
	}

	var req domain.ReviewTransaksiRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Payload review tidak valid")
		return
	}

	if req.Status == "" {
		Error(w, http.StatusBadRequest, "Status review wajib diisi")
		return
	}

	resp, err := h.transaksiService.Review(r.Context(), id, req, claims.UserID, claims.Role)
	if err != nil {
		if errors.Is(err, service.ErrInvalidReviewRole) {
			Error(w, http.StatusForbidden, err.Error())
			return
		}
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusOK, resp)
}
