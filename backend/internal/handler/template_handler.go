package handler

import (
	"net/http"

	"github.com/abyankamal/sidak/backend/internal/repository"
)

type TemplateHandler struct {
	templateRepo *repository.TemplateRepository
}

func NewTemplateHandler(templateRepo *repository.TemplateRepository) *TemplateHandler {
	return &TemplateHandler{templateRepo: templateRepo}
}

func (h *TemplateHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := h.templateRepo.GetActiveTemplates(r.Context())
	if err != nil {
		Error(w, http.StatusInternalServerError, "Gagal mengambil daftar template formulir")
		return
	}

	JSON(w, http.StatusOK, templates)
}
