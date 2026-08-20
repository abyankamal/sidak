package repository

import (
	"context"

	"github.com/abyankamal/sidak/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oklog/ulid/v2"
)

type MenuRepository struct {
	db *pgxpool.Pool
}

func NewMenuRepository(db *pgxpool.Pool) *MenuRepository {
	return &MenuRepository{db: db}
}

func (r *MenuRepository) GetPublicHierarchy(ctx context.Context) ([]domain.NavigasiMenuPublic, error) {
	// 1. Fetch active parent items
	parentQuery := `
		SELECT id, label, url, urutan
		FROM navigasi_menu
		WHERE parent_id IS NULL AND is_active = TRUE
		ORDER BY urutan ASC
	`
	rows, err := r.db.Query(ctx, parentQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	parentList := []domain.NavigasiMenuPublic{}
	parentIndices := make(map[string]int)

	for rows.Next() {
		var p domain.NavigasiMenuPublic
		p.Children = []domain.NavigasiMenuPublicChild{}
		if err := rows.Scan(&p.ID, &p.Label, &p.URL, &p.Urutan); err != nil {
			return nil, err
		}
		parentIndices[p.ID] = len(parentList)
		parentList = append(parentList, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// 2. Fetch active children items
	childQuery := `
		SELECT id, parent_id, label, url, urutan
		FROM navigasi_menu
		WHERE parent_id IS NOT NULL AND is_active = TRUE
		ORDER BY urutan ASC
	`
	cRows, err := r.db.Query(ctx, childQuery)
	if err != nil {
		return nil, err
	}
	defer cRows.Close()

	for cRows.Next() {
		var c domain.NavigasiMenuPublicChild
		var parentID string
		if err := cRows.Scan(&c.ID, &parentID, &c.Label, &c.URL, &c.Urutan); err != nil {
			return nil, err
		}
		if idx, ok := parentIndices[parentID]; ok {
			parentList[idx].Children = append(parentList[idx].Children, c)
		}
	}
	if err := cRows.Err(); err != nil {
		return nil, err
	}

	return parentList, nil
}

func (r *MenuRepository) GetAllAdmin(ctx context.Context) ([]domain.NavigasiMenuAdmin, error) {
	query := `
		SELECT id, parent_id, label, url, urutan, is_active
		FROM navigasi_menu
		ORDER BY COALESCE(parent_id, id), urutan ASC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []domain.NavigasiMenuAdmin{}
	for rows.Next() {
		var m domain.NavigasiMenuAdmin
		if err := rows.Scan(&m.ID, &m.ParentID, &m.Label, &m.URL, &m.Urutan, &m.IsActive); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *MenuRepository) Create(ctx context.Context, input domain.NavigasiMenuInput) (*domain.NavigasiMenu, error) {
	id := ulid.Make().String()
	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	query := `
		INSERT INTO navigasi_menu (id, parent_id, label, url, urutan, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, parent_id, label, url, urutan, is_active, created_at, updated_at
	`
	var m domain.NavigasiMenu
	err := r.db.QueryRow(ctx, query, id, input.ParentID, input.Label, input.URL, input.Urutan, isActive).Scan(
		&m.ID, &m.ParentID, &m.Label, &m.URL, &m.Urutan, &m.IsActive, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *MenuRepository) Update(ctx context.Context, id string, input domain.NavigasiMenuInput) error {
	query := `
		UPDATE navigasi_menu
		SET parent_id = $1,
		    label = $2,
		    url = $3,
		    urutan = $4,
		    is_active = COALESCE($5, is_active),
		    updated_at = NOW()
		WHERE id = $6
	`
	_, err := r.db.Exec(ctx, query, input.ParentID, input.Label, input.URL, input.Urutan, input.IsActive, id)
	return err
}

func (r *MenuRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM navigasi_menu WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
