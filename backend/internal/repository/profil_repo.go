package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/abyankamal/sidak/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProfilRepository struct {
	db *pgxpool.Pool
}

func NewProfilRepository(db *pgxpool.Pool) *ProfilRepository {
	return &ProfilRepository{db: db}
}

func (r *ProfilRepository) Get(ctx context.Context) (*domain.ProfilWilayah, error) {
	query := `
		SELECT nama_kelurahan, kecamatan, kabupaten_kota, visi, misi, sejarah, alamat_kantor, kontak_telepon, kontak_email, struktur_organisasi_file_path
		FROM profil_wilayah
		WHERE id = 1
	`
	var p domain.ProfilWilayah
	var misiRaw []byte
	err := r.db.QueryRow(ctx, query).Scan(
		&p.NamaKelurahan, &p.Kecamatan, &p.KabupatenKota, &p.Visi, &misiRaw,
		&p.Sejarah, &p.AlamatKantor, &p.KontakTelepon, &p.KontakEmail, &p.StrukturOrganisasiURL,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if len(misiRaw) > 0 {
		_ = json.Unmarshal(misiRaw, &p.Misi)
	}
	if p.Misi == nil {
		p.Misi = []string{}
	}

	return &p, nil
}

func (r *ProfilRepository) Update(ctx context.Context, input domain.ProfilWilayahInput) error {
	misiJSON, err := json.Marshal(input.Misi)
	if err != nil {
		return err
	}

	query := `
		UPDATE profil_wilayah
		SET nama_kelurahan = $1,
		    kecamatan = $2,
		    kabupaten_kota = $3,
		    visi = $4,
		    misi = $5,
		    sejarah = $6,
		    alamat_kantor = $7,
		    kontak_telepon = $8,
		    kontak_email = $9,
		    struktur_organisasi_file_path = $10,
		    updated_at = NOW()
		WHERE id = 1
	`
	_, err = r.db.Exec(ctx, query,
		input.NamaKelurahan, input.Kecamatan, input.KabupatenKota, input.Visi,
		misiJSON, input.Sejarah, input.AlamatKantor, input.KontakTelepon,
		input.KontakEmail, input.StrukturOrganisasiFilePath,
	)
	return err
}
