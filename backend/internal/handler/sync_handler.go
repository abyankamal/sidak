package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/abyankamal/sidak/backend/internal/domain"
	"github.com/abyankamal/sidak/backend/internal/middleware"
	"github.com/abyankamal/sidak/backend/internal/service"
)

type SyncHandler struct {
	syncService *service.SyncService
}

func NewSyncHandler(syncService *service.SyncService) *SyncHandler {
	return &SyncHandler{syncService: syncService}
}

func (h *SyncHandler) RequestPresignedURL(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	if claims == nil {
		Error(w, http.StatusUnauthorized, "Sesi tidak sah")
		return
	}

	var req domain.PresignUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Payload request presigned URL tidak valid")
		return
	}

	if req.TransaksiID == "" || req.FileName == "" || req.ContentType == "" {
		Error(w, http.StatusBadRequest, "transaksi_id, file_name, dan content_type wajib diisi")
		return
	}

	resp, err := h.syncService.GeneratePresignedUpload(r.Context(), req, claims.NIK)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Gagal membuat presigned upload URL")
		return
	}

	JSON(w, http.StatusOK, resp)
}

func (h *SyncHandler) Commit(w http.ResponseWriter, r *http.Request) {
	idempotencyKey := r.Header.Get("Idempotency-Key")
	if idempotencyKey == "" || len(idempotencyKey) != 26 {
		Error(w, http.StatusBadRequest, "Header Idempotency-Key wajib disertakan dalam format ULID (26 karakter)")
		return
	}

	var req domain.SyncCommitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Payload request commit tidak valid")
		return
	}

	if req.TransaksiID == "" || req.WargaNIK == "" || req.LayananID == "" || len(req.DataIsian) == 0 {
		Error(w, http.StatusBadRequest, "transaksi_id, warga_nik, layanan_id, dan data_isian wajib diisi")
		return
	}

	resp, err := h.syncService.Commit(r.Context(), idempotencyKey, req)
	if err != nil {
		if errors.Is(err, service.ErrConcurrentProcessing) {
			Error(w, http.StatusTooManyRequests, "Request dengan Idempotency-Key ini sedang diproses sistem")
			return
		}
		if errors.Is(err, service.ErrDuplicateTransaction) {
			// Idempotent duplicate replay -> 409
			Message(w, http.StatusConflict, "Transaksi duplikat (sudah pernah diproses sebelumnya)")
			return
		}
		if errors.Is(err, service.ErrInvalidDataIsian) || errors.Is(err, service.ErrLogicalDuplicate24h) {
			Error(w, http.StatusBadRequest, err.Error())
			return
		}

		Error(w, http.StatusInternalServerError, "Gagal memproses commit transaksi")
		return
	}

	JSON(w, http.StatusCreated, resp)
}
