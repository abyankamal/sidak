package repository

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/abyankamal/sidak/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oklog/ulid/v2"
)

func Slugify(s string) string {
	s = strings.ToLower(s)
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	s = reg.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

type KontenRepository struct {
	db *pgxpool.Pool
}

func NewKontenRepository(db *pgxpool.Pool) *KontenRepository {
	return &KontenRepository{db: db}
}

func (r *KontenRepository) ListPublic(ctx context.Context, tipe string, page, limit int) ([]domain.KontenPublikListItem, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	whereClauses := []string{"is_published = TRUE"}
	args := []any{}
	argIdx := 1

	if tipe != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("tipe = $%d", argIdx))
		args = append(args, strings.ToUpper(tipe))
		argIdx++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM konten_publik WHERE %s", whereSQL)
	var totalRecords int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&totalRecords); err != nil {
		return nil, 0, err
	}

	dataQuery := fmt.Sprintf(`
		SELECT id, tipe, judul, slug, ringkasan, thumbnail_file_path, published_at
		FROM konten_publik
		WHERE %s
		ORDER BY published_at DESC NULLS LAST, created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := []domain.KontenPublikListItem{}
	for rows.Next() {
		var item domain.KontenPublikListItem
		if err := rows.Scan(&item.ID, &item.Tipe, &item.Judul, &item.Slug, &item.Ringkasan, &item.ThumbnailURL, &item.PublishedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return list, totalRecords, nil
}

func (r *KontenRepository) GetBySlugPublic(ctx context.Context, slug string) (*domain.KontenPublikDetail, error) {
	query := `
		SELECT id, tipe, judul, slug, ringkasan, isi_konten, thumbnail_file_path, published_at, author_nama
		FROM konten_publik
		WHERE slug = $1 AND is_published = TRUE
	`
	var k domain.KontenPublikDetail
	err := r.db.QueryRow(ctx, query, slug).Scan(
		&k.ID, &k.Tipe, &k.Judul, &k.Slug, &k.Ringkasan, &k.IsiKonten, &k.ThumbnailURL, &k.PublishedAt, &k.AuthorNama,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &k, nil
}

func (r *KontenRepository) ListAdmin(ctx context.Context, tipe string, page, limit int) ([]domain.KontenPublikAdminItem, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	whereClauses := []string{"1=1"}
	args := []any{}
	argIdx := 1

	if tipe != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("tipe = $%d", argIdx))
		args = append(args, strings.ToUpper(tipe))
		argIdx++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM konten_publik WHERE %s", whereSQL)
	var totalRecords int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&totalRecords); err != nil {
		return nil, 0, err
	}

	dataQuery := fmt.Sprintf(`
		SELECT id, tipe, judul, slug, is_published, published_at, author_nama
		FROM konten_publik
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := []domain.KontenPublikAdminItem{}
	for rows.Next() {
		var item domain.KontenPublikAdminItem
		if err := rows.Scan(&item.ID, &item.Tipe, &item.Judul, &item.Slug, &item.IsPublished, &item.PublishedAt, &item.AuthorNama); err != nil {
			return nil, 0, err
		}
		list = append(list, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return list, totalRecords, nil
}

func (r *KontenRepository) Create(ctx context.Context, input domain.KontenPublikInput, authorID, authorNama string) (*domain.KontenPublik, error) {
	id := ulid.Make().String()
	slug := Slugify(input.Judul)

	// Ensure unique slug
	slug = fmt.Sprintf("%s-%s", slug, id[len(id)-6:])

	isPublished := false
	var publishedAt *time.Time
	if input.IsPublished != nil && *input.IsPublished {
		isPublished = true
		now := time.Now()
		publishedAt = &now
	}

	query := `
		INSERT INTO konten_publik (
			id, tipe, judul, slug, ringkasan, isi_konten, thumbnail_file_path, is_published, published_at, author_id, author_nama, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW()
		)
		RETURNING id, tipe, judul, slug, ringkasan, isi_konten, thumbnail_file_path, is_published, published_at, author_id, author_nama, created_at, updated_at
	`
	var k domain.KontenPublik
	err := r.db.QueryRow(ctx, query,
		id, strings.ToUpper(input.Tipe), input.Judul, slug, input.Ringkasan, input.IsiKonten,
		input.ThumbnailFilePath, isPublished, publishedAt, authorID, authorNama,
	).Scan(
		&k.ID, &k.Tipe, &k.Judul, &k.Slug, &k.Ringkasan, &k.IsiKonten, &k.ThumbnailFilePath,
		&k.IsPublished, &k.PublishedAt, &k.AuthorID, &k.AuthorNama, &k.CreatedAt, &k.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &k, nil
}

func (r *KontenRepository) Update(ctx context.Context, id string, input domain.KontenPublikInput) error {
	var isPublishedVal *bool = input.IsPublished
	var setPublishedAtSQL string = ""

	if isPublishedVal != nil && *isPublishedVal {
		setPublishedAtSQL = ", published_at = COALESCE(published_at, NOW())"
	}

	query := fmt.Sprintf(`
		UPDATE konten_publik
		SET tipe = $1,
		    judul = $2,
		    ringkasan = $3,
		    isi_konten = $4,
		    thumbnail_file_path = COALESCE($5, thumbnail_file_path),
		    is_published = COALESCE($6, is_published)%s,
		    updated_at = NOW()
		WHERE id = $7
	`, setPublishedAtSQL)

	_, err := r.db.Exec(ctx, query,
		strings.ToUpper(input.Tipe), input.Judul, input.Ringkasan, input.IsiKonten,
		input.ThumbnailFilePath, input.IsPublished, id,
	)
	return err
}

func (r *KontenRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM konten_publik WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
