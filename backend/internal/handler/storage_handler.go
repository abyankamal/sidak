package handler

import (
	"net/http"

	"github.com/abyankamal/sidak/backend/internal/middleware"
	"github.com/abyankamal/sidak/backend/internal/service"
)

type StorageHandler struct {
	storageService *service.StorageService
}

func NewStorageHandler(storageService *service.StorageService) *StorageHandler {
	return &StorageHandler{storageService: storageService}
}

func (h *StorageHandler) Upload(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	if claims == nil {
		Error(w, http.StatusUnauthorized, "Sesi tidak sah")
		return
	}

	// Limit request size to 10MB
	if err := r.ParseMultipartForm(10 * 1024 * 1024); err != nil {
		Error(w, http.StatusBadRequest, "Ukuran berkas melebihi batas atau format multipart tidak valid")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		Error(w, http.StatusBadRequest, "Form file 'file' wajib disertakan")
		return
	}
	defer file.Close()

	category := r.FormValue("category")
	if category == "" {
		category = "lampiran"
	}
	transaksiID := r.FormValue("transaksi_id")

	userIdentifier := "umum"
	if claims.NIK != nil && *claims.NIK != "" {
		userIdentifier = *claims.NIK
	} else if claims.NIP != nil && *claims.NIP != "" {
		userIdentifier = *claims.NIP
	}

	resp, err := h.storageService.SaveUpload(r.Context(), header, file, category, transaksiID, userIdentifier)
	if err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusCreated, resp)
}
