package repository

import (
	"context"
	"errors"

	"github.com/abyankamal/sidak/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TemplateRepository struct {
	db *pgxpool.Pool
}

func NewTemplateRepository(db *pgxpool.Pool) *TemplateRepository {
	return &TemplateRepository{db: db}
}

func (r *TemplateRepository) GetActiveTemplates(ctx context.Context) ([]domain.TemplateForm, error) {
	query := `
		SELECT layanan_id, nama_layanan, deskripsi, skema_json, is_active, created_at, updated_at
		FROM template_form
		WHERE is_active = TRUE
		ORDER BY layanan_id ASC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.TemplateForm
	for rows.Next() {
		var t domain.TemplateForm
		if err := rows.Scan(&t.LayananID, &t.NamaLayanan, &t.Deskripsi, &t.SkemaJSON, &t.IsActive, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *TemplateRepository) GetByID(ctx context.Context, layananID string) (*domain.TemplateForm, error) {
	query := `
		SELECT layanan_id, nama_layanan, deskripsi, skema_json, is_active, created_at, updated_at
		FROM template_form
		WHERE layanan_id = $1
	`
	var t domain.TemplateForm
	err := r.db.QueryRow(ctx, query, layananID).Scan(
		&t.LayananID, &t.NamaLayanan, &t.Deskripsi, &t.SkemaJSON, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}
