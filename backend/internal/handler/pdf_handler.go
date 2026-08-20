package handler

import (
	"errors"
	"net/http"

	"github.com/abyankamal/sidak/backend/internal/service"
	"github.com/go-chi/chi/v5"
)

type PDFHandler struct {
	pdfService *service.PDFService
}

func NewPDFHandler(pdfService *service.PDFService) *PDFHandler {
	return &PDFHandler{pdfService: pdfService}
}

func (h *PDFHandler) GeneratePDF(w http.ResponseWriter, r *http.Request) {
	transaksiID := chi.URLParam(r, "id")
	if transaksiID == "" {
		Error(w, http.StatusBadRequest, "ID permohonan layanan wajib disertakan")
		return
	}

	doc, err := h.pdfService.EnqueueJob(r.Context(), transaksiID)
	if err != nil {
		if errors.Is(err, service.ErrTransaksiNotFound) {
			Error(w, http.StatusNotFound, "Transaksi permohonan tidak ditemukan")
			return
		}
		if errors.Is(err, service.ErrTransaksiNotReviewed) {
			Error(w, http.StatusBadRequest, err.Error())
			return
		}
		Error(w, http.StatusInternalServerError, "Gagal mendaftarkan antrean pembuatan PDF")
		return
	}

	JSON(w, http.StatusAccepted, map[string]any{
		"job_id": doc.ID,
		"status": doc.Status,
	})
}

func (h *PDFHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		Error(w, http.StatusBadRequest, "ID dokumen atau transaksi wajib disertakan")
		return
	}

	statusResp, err := h.pdfService.GetStatus(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrDokumenNotFound) {
			Error(w, http.StatusNotFound, "Data dokumen tidak ditemukan")
			return
		}
		Error(w, http.StatusInternalServerError, "Gagal mengambil status dokumen PDF")
		return
	}

	JSON(w, http.StatusOK, statusResp)
}
