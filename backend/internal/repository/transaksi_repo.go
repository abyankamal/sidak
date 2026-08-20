package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/abyankamal/sidak/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransaksiRepository struct {
	db *pgxpool.Pool
}

func NewTransaksiRepository(db *pgxpool.Pool) *TransaksiRepository {
	return &TransaksiRepository{db: db}
}

func (r *TransaksiRepository) CheckLogicalDuplicate24h(ctx context.Context, wargaNIK, layananID, excludeID string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM transaksi_pelayanan
			WHERE warga_nik = $1 
			  AND layanan_id = $2 
			  AND status = 'menunggu_review'
			  AND id != $3
			  AND created_at >= NOW() - INTERVAL '24 hours'
		)
	`
	var exists bool
	err := r.db.QueryRow(ctx, query, wargaNIK, layananID, excludeID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *TransaksiRepository) CreateAtomic(ctx context.Context, trx *domain.TransaksiPelayanan) (bool, error) {
	query := `
		INSERT INTO transaksi_pelayanan (
			id, warga_nik, layanan_id, data_isian, lampiran, status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, NOW(), NOW()
		)
		ON CONFLICT (id) DO NOTHING
	`
	lampiran := trx.Lampiran
	if lampiran == nil {
		lampiran = []string{}
	}
	status := trx.Status
	if status == "" {
		status = "menunggu_review"
	}

	tag, err := r.db.Exec(ctx, query, trx.ID, trx.WargaNIK, trx.LayananID, trx.DataIsian, lampiran, status)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (r *TransaksiRepository) GetByID(ctx context.Context, id string) (*domain.TransaksiPelayanan, error) {
	query := `
		SELECT id, warga_nik, layanan_id, data_isian, lampiran, status, catatan_review, reviewed_by, reviewed_at, created_at, updated_at
		FROM transaksi_pelayanan
		WHERE id = $1
	`
	var t domain.TransaksiPelayanan
	err := r.db.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.WargaNIK, &t.LayananID, &t.DataIsian, &t.Lampiran, &t.Status,
		&t.CatatanReview, &t.ReviewedBy, &t.ReviewedAt, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (r *TransaksiRepository) List(ctx context.Context, filter domain.TransaksiFilter) ([]domain.TransaksiListItem, int, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 20
	}
	offset := (filter.Page - 1) * filter.Limit

	whereClauses := []string{"1=1"}
	args := []any{}
	argIdx := 1

	if filter.Status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}

	if filter.LayananID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("layanan_id = $%d", argIdx))
		args = append(args, filter.LayananID)
		argIdx++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM transaksi_pelayanan WHERE %s", whereSQL)
	var totalRecords int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&totalRecords); err != nil {
		return nil, 0, err
	}

	dataQuery := fmt.Sprintf(`
		SELECT id, warga_nik, layanan_id, status, created_at
		FROM transaksi_pelayanan
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIdx, argIdx+1)

	args = append(args, filter.Limit, offset)

	rows, err := r.db.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := []domain.TransaksiListItem{}
	for rows.Next() {
		var item domain.TransaksiListItem
		if err := rows.Scan(&item.ID, &item.WargaNIK, &item.LayananID, &item.Status, &item.CreatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return list, totalRecords, nil
}

func (r *TransaksiRepository) UpdateReview(ctx context.Context, id, status, catatanReview string, reviewedBy string) error {
	now := time.Now()
	query := `
		UPDATE transaksi_pelayanan
		SET status = $1, catatan_review = $2, reviewed_by = $3, reviewed_at = $4, updated_at = $4
		WHERE id = $5
	`
	tag, err := r.db.Exec(ctx, query, status, catatanReview, reviewedBy, now, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("transaksi not found")
	}
	return nil
}
