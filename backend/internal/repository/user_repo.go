package repository

import (
	"context"
	"errors"

	"github.com/abyankamal/sidak/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByNIK(ctx context.Context, nik string) (*domain.User, error) {
	query := `
		SELECT id, nik, nama, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE nik = $1
	`
	var u domain.User
	err := r.db.QueryRow(ctx, query, nik).Scan(
		&u.ID, &u.NIK, &u.Nama, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, nik, nama, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	var u domain.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.NIK, &u.Nama, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}
