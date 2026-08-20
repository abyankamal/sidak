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
	ErrKontenNotFound = errors.New("berita atau pengumuman tidak ditemukan")
)

type CMSService struct {
	profilRepo *repository.ProfilRepository
	menuRepo   *repository.MenuRepository
	kontenRepo *repository.KontenRepository
	cfg        *config.Config
}

func NewCMSService(
	profilRepo *repository.ProfilRepository,
	menuRepo *repository.MenuRepository,
	kontenRepo *repository.KontenRepository,
	cfg *config.Config,
) *CMSService {
	return &CMSService{
		profilRepo: profilRepo,
		menuRepo:   menuRepo,
		kontenRepo: kontenRepo,
		cfg:        cfg,
	}
}

// -----------------------------------------------------------------------------
// CMS PUBLIC (Unauthenticated)
// -----------------------------------------------------------------------------

func (s *CMSService) GetPublicProfil(ctx context.Context) (*domain.ProfilWilayah, error) {
	p, err := s.profilRepo.Get(ctx)
	if err != nil {
		return nil, err
	}
	if p != nil && p.StrukturOrganisasiURL != nil && *p.StrukturOrganisasiURL != "" {
		cleanPath := strings.TrimPrefix(*p.StrukturOrganisasiURL, "uploads/")
		url := fmt.Sprintf("%s/%s", strings.TrimRight(s.cfg.StoragePublicURL, "/"), cleanPath)
		p.StrukturOrganisasiURL = &url
	}
	return p, nil
}

func (s *CMSService) GetPublicMenu(ctx context.Context) ([]domain.NavigasiMenuPublic, error) {
	return s.menuRepo.GetPublicHierarchy(ctx)
}

func (s *CMSService) GetPublicKontenList(ctx context.Context, tipe string, page, limit int) (*domain.KontenPublikListResponse, error) {
	data, total, err := s.kontenRepo.ListPublic(ctx, tipe, page, limit)
	if err != nil {
		return nil, err
	}

	for i := range data {
		if data[i].ThumbnailURL != nil && *data[i].ThumbnailURL != "" {
			cleanPath := strings.TrimPrefix(*data[i].ThumbnailURL, "uploads/")
			url := fmt.Sprintf("%s/%s", strings.TrimRight(s.cfg.StoragePublicURL, "/"), cleanPath)
			data[i].ThumbnailURL = &url
		}
	}

	totalPages := 1
	if limit > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	return &domain.KontenPublikListResponse{
		Data: data,
		Meta: domain.PaginationMeta{
			CurrentPage:  page,
			TotalPages:   totalPages,
			TotalRecords: total,
		},
	}, nil
}

func (s *CMSService) GetPublicKontenDetail(ctx context.Context, slug string) (*domain.KontenPublikDetail, error) {
	k, err := s.kontenRepo.GetBySlugPublic(ctx, slug)
	if err != nil {
		return nil, err
	}
	if k == nil {
		return nil, ErrKontenNotFound
	}

	if k.ThumbnailURL != nil && *k.ThumbnailURL != "" {
		cleanPath := strings.TrimPrefix(*k.ThumbnailURL, "uploads/")
		url := fmt.Sprintf("%s/%s", strings.TrimRight(s.cfg.StoragePublicURL, "/"), cleanPath)
		k.ThumbnailURL = &url
	}

	return k, nil
}

// -----------------------------------------------------------------------------
// CMS ADMIN (LURAH / SEKLUR / KASI)
// -----------------------------------------------------------------------------

func (s *CMSService) UpdateProfil(ctx context.Context, input domain.ProfilWilayahInput) (*domain.StandardMessageResponse, error) {
	if err := s.profilRepo.Update(ctx, input); err != nil {
		return nil, err
	}
	return &domain.StandardMessageResponse{
		Message: "Profil wilayah berhasil diperbarui",
	}, nil
}

func (s *CMSService) GetAllMenuAdmin(ctx context.Context) ([]domain.NavigasiMenuAdmin, error) {
	return s.menuRepo.GetAllAdmin(ctx)
}

func (s *CMSService) CreateMenu(ctx context.Context, input domain.NavigasiMenuInput) (*domain.StandardMessageResponse, error) {
	if _, err := s.menuRepo.Create(ctx, input); err != nil {
		return nil, err
	}
	return &domain.StandardMessageResponse{
		Message: "Item navigasi menu berhasil ditambahkan",
	}, nil
}

func (s *CMSService) UpdateMenu(ctx context.Context, id string, input domain.NavigasiMenuInput) (*domain.StandardMessageResponse, error) {
	if err := s.menuRepo.Update(ctx, id, input); err != nil {
		return nil, err
	}
	return &domain.StandardMessageResponse{
		Message: "Item navigasi menu berhasil diperbarui",
	}, nil
}

func (s *CMSService) DeleteMenu(ctx context.Context, id string) (*domain.StandardMessageResponse, error) {
	if err := s.menuRepo.Delete(ctx, id); err != nil {
		return nil, err
	}
	return &domain.StandardMessageResponse{
		Message: "Item navigasi menu berhasil dihapus",
	}, nil
}

func (s *CMSService) GetAdminKontenList(ctx context.Context, tipe string, page, limit int) (*domain.KontenPublikAdminListResponse, error) {
	data, total, err := s.kontenRepo.ListAdmin(ctx, tipe, page, limit)
	if err != nil {
		return nil, err
	}

	totalPages := 1
	if limit > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	return &domain.KontenPublikAdminListResponse{
		Data: data,
		Meta: domain.PaginationMeta{
			CurrentPage:  page,
			TotalPages:   totalPages,
			TotalRecords: total,
		},
	}, nil
}

func (s *CMSService) CreateKonten(ctx context.Context, input domain.KontenPublikInput, authorID, authorNama string) (*domain.StandardMessageResponse, error) {
	if _, err := s.kontenRepo.Create(ctx, input, authorID, authorNama); err != nil {
		return nil, err
	}
	return &domain.StandardMessageResponse{
		Message: "Konten publik berhasil dibuat",
	}, nil
}

func (s *CMSService) UpdateKonten(ctx context.Context, id string, input domain.KontenPublikInput) (*domain.StandardMessageResponse, error) {
	if err := s.kontenRepo.Update(ctx, id, input); err != nil {
		return nil, err
	}
	return &domain.StandardMessageResponse{
		Message: "Konten publik berhasil diperbarui",
	}, nil
}

func (s *CMSService) DeleteKonten(ctx context.Context, id string) (*domain.StandardMessageResponse, error) {
	if err := s.kontenRepo.Delete(ctx, id); err != nil {
		return nil, err
	}
	return &domain.StandardMessageResponse{
		Message: "Konten publik berhasil dihapus",
	}, nil
}
