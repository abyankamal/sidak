package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DokumenRepository struct {
	db *pgxpool.Pool
}

func NewDokumenRepository(db *pgxpool.Pool) *DokumenRepository {
	return &DokumenRepository{db: db}
}

type DokumenOutput struct {
	ID          string     `json:"id"`
	TransaksiID string     `json:"transaksi_id"`
	Status      string     `json:"status"` // PROCESSING, READY, FAILED
	FilePath    *string    `json:"file_path,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (r *DokumenRepository) UpsertProcessing(ctx context.Context, id, transaksiID string) (*DokumenOutput, error) {
	query := `
		INSERT INTO dokumen_output (id, transaksi_id, status, file_path, created_at, updated_at)
		VALUES ($1, $2, 'PROCESSING', NULL, NOW(), NOW())
		ON CONFLICT (id) DO UPDATE
		SET status = 'PROCESSING', file_path = NULL, updated_at = NOW()
		RETURNING id, transaksi_id, status, file_path, created_at, updated_at
	`
	var d DokumenOutput
	err := r.db.QueryRow(ctx, query, id, transaksiID).Scan(
		&d.ID, &d.TransaksiID, &d.Status, &d.FilePath, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *DokumenRepository) GetByID(ctx context.Context, id string) (*DokumenOutput, error) {
	query := `
		SELECT id, transaksi_id, status, file_path, created_at, updated_at
		FROM dokumen_output
		WHERE id = $1
	`
	var d DokumenOutput
	err := r.db.QueryRow(ctx, query, id).Scan(
		&d.ID, &d.TransaksiID, &d.Status, &d.FilePath, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

func (r *DokumenRepository) GetByTransaksiID(ctx context.Context, transaksiID string) (*DokumenOutput, error) {
	query := `
		SELECT id, transaksi_id, status, file_path, created_at, updated_at
		FROM dokumen_output
		WHERE transaksi_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	var d DokumenOutput
	err := r.db.QueryRow(ctx, query, transaksiID).Scan(
		&d.ID, &d.TransaksiID, &d.Status, &d.FilePath, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

func (r *DokumenRepository) UpdateStatus(ctx context.Context, id, status string, filePath *string) error {
	query := `
		UPDATE dokumen_output
		SET status = $1, file_path = $2, updated_at = NOW()
		WHERE id = $3
	`
	_, err := r.db.Exec(ctx, query, status, filePath, id)
	return err
}
