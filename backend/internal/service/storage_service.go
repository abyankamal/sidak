package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/abyankamal/sidak/backend/config"
	"github.com/abyankamal/sidak/backend/internal/domain"
)

var (
	ErrInvalidCategory = errors.New("kategori upload tidak valid (pilihan: lampiran, cms, dokumen)")
	ErrFileTooLarge    = errors.New("ukuran berkas melebihi batas maksimum 10MB")
)

type StorageService struct {
	cfg *config.Config
}

func NewStorageService(cfg *config.Config) *StorageService {
	return &StorageService{cfg: cfg}
}

func (s *StorageService) SaveUpload(ctx context.Context, header *multipart.FileHeader, file multipart.File, category, transaksiID, wargaNIK string) (*domain.FileUploadResponse, error) {
	// 10MB max limit
	if header.Size > 10*1024*1024 {
		return nil, ErrFileTooLarge
	}

	cleanFileName := filepath.Base(header.Filename)
	cleanFileName = strings.ReplaceAll(cleanFileName, " ", "_")

	var relDir string
	var relFilePath string

	switch category {
	case "lampiran":
		if wargaNIK == "" {
			wargaNIK = "umum"
		}
		if transaksiID == "" {
			transaksiID = "draft"
		}
		relDir = filepath.Join("lampiran", wargaNIK)
		relFilePath = filepath.Join(relDir, fmt.Sprintf("%s_%s", transaksiID, cleanFileName))

	case "cms":
		year := fmt.Sprintf("%d", time.Now().Year())
		relDir = filepath.Join("public", "cms", year)
		relFilePath = filepath.Join(relDir, cleanFileName)

	case "dokumen":
		if transaksiID == "" {
			transaksiID = "draft"
		}
		relDir = "dokumen"
		relFilePath = filepath.Join(relDir, fmt.Sprintf("%s_%s", transaksiID, cleanFileName))

	default:
		return nil, ErrInvalidCategory
	}

	// Target absolute directory
	targetDir := filepath.Join(s.cfg.StorageBasePath, relDir)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return nil, fmt.Errorf("gagal membuat direktori penyimpanan: %w", err)
	}

	// Target absolute file path
	targetAbsPath := filepath.Join(s.cfg.StorageBasePath, relFilePath)
	out, err := os.Create(targetAbsPath)
	if err != nil {
		return nil, fmt.Errorf("gagal menyimpan berkas: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		return nil, fmt.Errorf("gagal menyalin isi berkas: %w", err)
	}

	// Normalized relative file path e.g. "uploads/lampiran/..."
	storedPath := filepath.ToSlash(filepath.Join("uploads", relFilePath))
	fileURL := fmt.Sprintf("%s/%s", strings.TrimRight(s.cfg.StoragePublicURL, "/"), filepath.ToSlash(relFilePath))

	return &domain.FileUploadResponse{
		FilePath: storedPath,
		FileURL:  fileURL,
	}, nil
}
