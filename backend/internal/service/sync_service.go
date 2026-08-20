package service

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/abyankamal/sidak/backend/config"
	"github.com/abyankamal/sidak/backend/internal/domain"
	"github.com/abyankamal/sidak/backend/internal/repository"
)

var (
	ErrConcurrentProcessing = errors.New("request dengan Idempotency-Key ini sedang diproses sistem")
	ErrInvalidDataIsian      = errors.New("payload data isian tidak sesuai JSON Schema layanan")
	ErrLogicalDuplicate24h   = errors.New("pengajuan layanan yang sama untuk NIK ini masih dalam antrean review 24 jam terakhir")
	ErrDuplicateTransaction  = errors.New("transaksi duplikat (sudah pernah diproses sebelumnya)")
)

type SyncService struct {
	transaksiRepo *repository.TransaksiRepository
	schemaCache   *SchemaCache
	cfg           *config.Config
	inFlightLocks sync.Map // map[string]struct{}
}

func NewSyncService(transaksiRepo *repository.TransaksiRepository, schemaCache *SchemaCache, cfg *config.Config) *SyncService {
	return &SyncService{
		transaksiRepo: transaksiRepo,
		schemaCache:   schemaCache,
		cfg:           cfg,
	}
}

func (s *SyncService) GeneratePresignedUpload(ctx context.Context, req domain.PresignUploadRequest, wargaNIK string) (*domain.PresignUploadResponse, error) {
	// Storage path isolation: lampiran/{warga_nik}/{transaksi_id}_{file_name}
	r2FilePath := fmt.Sprintf("lampiran/%s/%s_%s", wargaNIK, req.TransaksiID, req.FileName)

	// Direct upload target URL (Cloudflare R2)
	uploadURL := fmt.Sprintf("%s/%s", s.cfg.R2PublicURL, r2FilePath)
	if s.cfg.R2AccountID != "" {
		uploadURL = fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s/%s", s.cfg.R2AccountID, s.cfg.R2BucketName, r2FilePath)
	}

	return &domain.PresignUploadResponse{
		UploadURL:  uploadURL,
		FilePathR2: r2FilePath,
	}, nil
}

func (s *SyncService) Commit(ctx context.Context, idempotencyKey string, req domain.SyncCommitRequest) (*domain.StandardMessageResponse, error) {
	if idempotencyKey == "" {
		idempotencyKey = req.TransaksiID
	}

	// 1. In-flight Concurrency Lock
	_, loaded := s.inFlightLocks.LoadOrStore(idempotencyKey, struct{}{})
	if loaded {
		return nil, ErrConcurrentProcessing
	}
	defer s.inFlightLocks.Delete(idempotencyKey)

	// 2. Validate Dynamic JSON Schema
	if err := s.schemaCache.Validate(req.LayananID, req.DataIsian); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidDataIsian, err)
	}

	// 3. Check 24-Hour Logical Duplicate (NIK + LayananID while still 'menunggu_review')
	isDup, err := s.transaksiRepo.CheckLogicalDuplicate24h(ctx, req.WargaNIK, req.LayananID, idempotencyKey)
	if err != nil {
		return nil, err
	}
	if isDup {
		return nil, ErrLogicalDuplicate24h
	}

	// 4. Atomic Insert to Database
	trx := &domain.TransaksiPelayanan{
		ID:        idempotencyKey,
		WargaNIK:  req.WargaNIK,
		LayananID: req.LayananID,
		DataIsian: req.DataIsian,
		Lampiran:  req.Lampiran,
		Status:    "menunggu_review",
	}

	isInserted, err := s.transaksiRepo.CreateAtomic(ctx, trx)
	if err != nil {
		return nil, err
	}

	if !isInserted {
		// Idempotent duplicate replay -> return 409
		return nil, ErrDuplicateTransaction
	}

	return &domain.StandardMessageResponse{
		Message: "Transaksi berhasil disimpan",
	}, nil
}
