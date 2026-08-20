-- =============================================================================
-- MIGRATION 000001: INITIAL SCHEMA FOR SIDAK
-- =============================================================================

-- 1. Tabel Pengguna (RBAC Sederhana: SEKDES, KASI, KADER)
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(26) PRIMARY KEY, -- ULID
    nik VARCHAR(16) UNIQUE NOT NULL,
    nama VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('SEKLUR', 'KASI', 'KADER')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. Tabel Template Form (JSON Schema Layanan Publik)
CREATE TABLE IF NOT EXISTS template_form (
    layanan_id VARCHAR(50) PRIMARY KEY,
    nama_layanan VARCHAR(255) NOT NULL,
    deskripsi TEXT,
    skema_json JSONB NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 3. Tabel Transaksi Pelayanan (Offline-First Sinkronisasi & Verifikasi)
CREATE TABLE IF NOT EXISTS transaksi_pelayanan (
    id VARCHAR(26) PRIMARY KEY, -- ULID dari klien (Idempotency Key)
    warga_nik VARCHAR(16) NOT NULL,
    layanan_id VARCHAR(50) NOT NULL REFERENCES template_form(layanan_id) ON UPDATE CASCADE,
    data_isian JSONB NOT NULL,
    lampiran TEXT[] DEFAULT '{}',
    status VARCHAR(30) NOT NULL DEFAULT 'menunggu_review' CHECK (status IN ('menunggu_review', 'sudah_di_review', 'butuh_revisi')),
    catatan_review TEXT,
    reviewed_by VARCHAR(26) REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indeks GIN untuk query dinamis JSONB data_isian
CREATE INDEX IF NOT EXISTS idx_transaksi_data_isian ON transaksi_pelayanan USING gin (data_isian);

-- Indeks untuk pencegahan duplikasi logis 24 jam & filter antrean pelayanan
CREATE INDEX IF NOT EXISTS idx_transaksi_dedup_24h ON transaksi_pelayanan (warga_nik, layanan_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_transaksi_status_created ON transaksi_pelayanan (status, created_at DESC);

-- 4. Tabel Profil Wilayah (Singleton Locked ID = 1)
CREATE TABLE IF NOT EXISTS profil_wilayah (
    id INT PRIMARY KEY CHECK (id = 1),
    nama_kelurahan VARCHAR(255) NOT NULL,
    kecamatan VARCHAR(255) NOT NULL,
    kabupaten_kota VARCHAR(255) NOT NULL,
    visi TEXT NOT NULL,
    misi JSONB NOT NULL DEFAULT '[]'::jsonb,
    sejarah TEXT,
    alamat_kantor TEXT NOT NULL,
    kontak_telepon VARCHAR(50),
    kontak_email VARCHAR(255),
    struktur_organisasi_r2_key VARCHAR(500),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 5. Tabel Navigasi Menu (Hierarki Maksimal 2 Level)
CREATE TABLE IF NOT EXISTS navigasi_menu (
    id VARCHAR(26) PRIMARY KEY, -- ULID
    parent_id VARCHAR(26) REFERENCES navigasi_menu(id) ON DELETE CASCADE,
    label VARCHAR(100) NOT NULL,
    url VARCHAR(255) NOT NULL,
    urutan INT NOT NULL DEFAULT 1,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_navigasi_menu_parent_urutan ON navigasi_menu (parent_id, urutan);

-- 6. Tabel Konten Publik (Berita & Pengumuman CMS)
CREATE TABLE IF NOT EXISTS konten_publik (
    id VARCHAR(26) PRIMARY KEY, -- ULID
    tipe VARCHAR(20) NOT NULL CHECK (tipe IN ('BERITA', 'PENGUMUMAN')),
    judul VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    ringkasan TEXT NOT NULL,
    isi_konten TEXT NOT NULL,
    thumbnail_r2_key VARCHAR(500),
    is_published BOOLEAN NOT NULL DEFAULT FALSE,
    published_at TIMESTAMPTZ,
    author_id VARCHAR(26) REFERENCES users(id) ON DELETE SET NULL,
    author_nama VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_konten_publik_slug ON konten_publik (slug);
CREATE INDEX IF NOT EXISTS idx_konten_publik_list ON konten_publik (tipe, is_published, published_at DESC);

-- 7. Tabel Status Dokumen Output PDF (Asinkron Gotenberg Engine)
CREATE TABLE IF NOT EXISTS dokumen_output (
    id VARCHAR(26) PRIMARY KEY, -- ULID
    transaksi_id VARCHAR(26) NOT NULL REFERENCES transaksi_pelayanan(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'PROCESSING' CHECK (status IN ('PROCESSING', 'READY', 'FAILED')),
    file_path_r2 VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_dokumen_output_transaksi ON dokumen_output (transaksi_id);
