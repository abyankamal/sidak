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
	ErrPathTraversal   = errors.New("deteksi manipulasi jalur berkas (path traversal)")
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

	safeWargaNIK := filepath.Base(wargaNIK)
	if safeWargaNIK == "." || safeWargaNIK == "/" || safeWargaNIK == "\\" {
		safeWargaNIK = "umum"
	}

	safeTransaksiID := filepath.Base(transaksiID)
	if safeTransaksiID == "." || safeTransaksiID == "/" || safeTransaksiID == "\\" {
		safeTransaksiID = "draft"
	}

	var relDir string
	var relFilePath string

	switch category {
	case "lampiran":
		if wargaNIK == "" {
			safeWargaNIK = "umum"
		}
		if transaksiID == "" {
			safeTransaksiID = "draft"
		}
		relDir = filepath.Join("lampiran", safeWargaNIK)
		relFilePath = filepath.Join(relDir, fmt.Sprintf("%s_%s", safeTransaksiID, cleanFileName))

	case "cms":
		year := fmt.Sprintf("%d", time.Now().Year())
		relDir = filepath.Join("public", "cms", year)
		relFilePath = filepath.Join(relDir, cleanFileName)

	case "dokumen":
		if transaksiID == "" {
			safeTransaksiID = "draft"
		}
		relDir = "dokumen"
		relFilePath = filepath.Join(relDir, fmt.Sprintf("%s_%s", safeTransaksiID, cleanFileName))

	default:
		return nil, ErrInvalidCategory
	}

	baseAbsPath, err := filepath.Abs(s.cfg.StorageBasePath)
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan path absolut storage: %w", err)
	}

	targetAbsPath := filepath.Join(baseAbsPath, relFilePath)
	targetAbsPath = filepath.Clean(targetAbsPath)

	if !strings.HasPrefix(targetAbsPath, baseAbsPath+string(os.PathSeparator)) && targetAbsPath != baseAbsPath {
		return nil, ErrPathTraversal
	}

	targetDir := filepath.Dir(targetAbsPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return nil, fmt.Errorf("gagal membuat direktori penyimpanan: %w", err)
	}
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
