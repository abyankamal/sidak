package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/abyankamal/sidak/backend/config"
	"github.com/abyankamal/sidak/backend/internal/domain"
	"github.com/abyankamal/sidak/backend/internal/repository"
)

var (
	ErrTransaksiNotFound = errors.New("transaksi tidak ditemukan")
	ErrInvalidReviewRole = errors.New("hanya LURAH, SEKLUR, atau KASI yang berwenang melakukan review")
)

type TransaksiService struct {
	transaksiRepo *repository.TransaksiRepository
	cfg           *config.Config
}

func NewTransaksiService(transaksiRepo *repository.TransaksiRepository, cfg *config.Config) *TransaksiService {
	return &TransaksiService{
		transaksiRepo: transaksiRepo,
		cfg:           cfg,
	}
}

func (s *TransaksiService) List(ctx context.Context, filter domain.TransaksiFilter) (*domain.TransaksiListResponse, error) {
	data, total, err := s.transaksiRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	totalPages := 1
	if filter.Limit > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(filter.Limit)))
	}

	return &domain.TransaksiListResponse{
		Data: data,
		Meta: domain.PaginationMeta{
			CurrentPage:  filter.Page,
			TotalPages:   totalPages,
			TotalRecords: total,
		},
	}, nil
}

func (s *TransaksiService) GetDetail(ctx context.Context, id string) (*domain.TransaksiDetailResponse, error) {
	t, err := s.transaksiRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, ErrTransaksiNotFound
	}

	// Generate download URLs for attachments
	lampiranURLs := make([]string, len(t.Lampiran))
	for i, lampiranKey := range t.Lampiran {
		cleanKey := strings.TrimPrefix(lampiranKey, "uploads/")
		lampiranURLs[i] = fmt.Sprintf("%s/%s", strings.TrimRight(s.cfg.StoragePublicURL, "/"), cleanKey)
	}

	return &domain.TransaksiDetailResponse{
		ID:            t.ID,
		WargaNIK:      t.WargaNIK,
		LayananID:     t.LayananID,
		DataIsian:     t.DataIsian,
		LampiranURLs:  lampiranURLs,
		Status:        t.Status,
		CatatanReview: t.CatatanReview,
		ReviewedBy:    t.ReviewedBy,
		ReviewedAt:    t.ReviewedAt,
		CreatedAt:     t.CreatedAt,
	}, nil
}

func (s *TransaksiService) Review(ctx context.Context, id string, req domain.ReviewTransaksiRequest, reviewedBy, userRole string) (*domain.StandardMessageResponse, error) {
	role := strings.ToUpper(userRole)
	if role != "LURAH" && role != "SEKLUR" && role != "KASI" {
		return nil, ErrInvalidReviewRole
	}

	if req.Status != "sudah_di_review" && req.Status != "butuh_revisi" {
		return nil, errors.New("status review harus 'sudah_di_review' atau 'butuh_revisi'")
	}

	catatan := ""
	if req.CatatanReview != nil {
		catatan = *req.CatatanReview
	}

	if err := s.transaksiRepo.UpdateReview(ctx, id, req.Status, catatan, reviewedBy); err != nil {
		return nil, err
	}

	return &domain.StandardMessageResponse{
		Message: "Status verifikasi transaksi berhasil diperbarui",
	}, nil
}
