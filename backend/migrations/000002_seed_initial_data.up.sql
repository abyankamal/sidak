-- =============================================================================
-- MIGRATION 000002: SEED INITIAL DATA
-- =============================================================================

-- 1. Seed Users (Default password: AdminSidak2026!)
-- Hash bcrypt cost 10: $2a$10$T81xQvL80mRjW2P2Z8t3a.3Rz9d5gqM8uKj2F7eR1H9vV4W6s8T7e
INSERT INTO users (id, nik, nama, email, password_hash, role, created_at, updated_at)
VALUES 
    (
        '01ARZ3NDEKTSV4RRFFQ69G5001', 
        '3205010101800001', 
        'Drs. H. Mulyadi (Seklur)', 
        'seklur@sukanegla.desa.id', 
        '$2a$10$T81xQvL80mRjW2P2Z8t3a.3Rz9d5gqM8uKj2F7eR1H9vV4W6s8T7e', 
        'SEKLUR', 
        NOW(), 
        NOW()
    ),
    (
        '01ARZ3NDEKTSV4RRFFQ69G5002', 
        '3205010202850002', 
        'Siti Nurhaliza, S.AP (Kasi Pelayanan)', 
        'kasi@sukanegla.desa.id', 
        '$2a$10$T81xQvL80mRjW2P2Z8t3a.3Rz9d5gqM8uKj2F7eR1H9vV4W6s8T7e', 
        'KASI', 
        NOW(), 
        NOW()
    ),
    (
        '01ARZ3NDEKTSV4RRFFQ69G5003', 
        '3205010303920003', 
        'Asep Sunandar (Kader RW 01)', 
        'kader01@sukanegla.desa.id', 
        '$2a$10$T81xQvL80mRjW2P2Z8t3a.3Rz9d5gqM8uKj2F7eR1H9vV4W6s8T7e', 
        'KADER', 
        NOW(), 
        NOW()
    )
ON CONFLICT (id) DO NOTHING;

-- 2. Seed Profil Wilayah (Singleton ID = 1)
INSERT INTO profil_wilayah (
    id, nama_kelurahan, kecamatan, kabupaten_kota, visi, misi, sejarah, alamat_kantor, kontak_telepon, kontak_email, struktur_organisasi_r2_key, updated_at
)
VALUES (
    1,
    'Sukanegla',
    'Garut Kota',
    'Kabupaten Garut',
    'Terwujudnya Pelayanan Kelurahan Sukanegla yang Bersih, Prima, Responsif, dan Terpercaya Berbasis Digital.',
    '[
        "Meningkatkan kualitas pelayanan administrasi terpadu berbasis sistem informasi terintegrasi.",
        "Mewujudkan tata kelola kewilayahan yang transparan, partisipatif, dan akuntabel.",
        "Mengoptimalkan peran kader dan lembaga kemasyarakatan dalam pelayanan warga tingkat RT/RW.",
        "Memelihara ketenteraman, ketertiban, dan keharmonisan sosial masyarakat kewilayahan."
    ]'::jsonb,
    'Kelurahan Sukanegla merupakan salah satu kelurahan di wilayah Kecamatan Garut Kota yang memiliki sejarah panjang dalam pelayanan masyarakat bergotong royong dan kini bertransformasi menjadi kelurahan digital modern terdepan di Kabupaten Garut.',
    'Jl. Sukanegla Raya No. 45, RT 02 / RW 03, Kelurahan Sukanegla, Garut Kota, Jawa Barat 44118',
    '0262-234567',
    'pelayanan@sukanegla.desa.id',
    'public/cms/2026/struktur_organisasi_sukanegla.png',
    NOW()
)
ON CONFLICT (id) DO NOTHING;

-- 3. Seed Template Form Layanan
INSERT INTO template_form (layanan_id, nama_layanan, deskripsi, skema_json, is_active, created_at, updated_at)
VALUES 
    (
        'SKTM',
        'Surat Keterangan Tidak Mampu',
        'Pelayanan penerbitan SKTM untuk keperluan beasiswa, keringanan biaya pendidikan, jaminan kesehatan, atau bantuan sosial.',
        '{
            "$schema": "http://json-schema.org/draft-07/schema#",
            "type": "object",
            "required": ["keperluan", "penghasilan_bulanan", "jumlah_tanggungan"],
            "properties": {
                "keperluan": {
                    "type": "string",
                    "minLength": 3,
                    "description": "Tujuan/alasan pengajuan SKTM"
                },
                "penghasilan_bulanan": {
                    "type": "integer",
                    "minimum": 0,
                    "description": "Nominal estimasi penghasilan per bulan (Rupiah)"
                },
                "jumlah_tanggungan": {
                    "type": "integer",
                    "minimum": 0,
                    "maximum": 20,
                    "description": "Jumlah anggota keluarga yang menjadi tanggungan"
                },
                "keterangan_pekerjaan": {
                    "type": "string",
                    "description": "Jenis pekerjaan utama pemohon/kepala keluarga"
                }
            }
        }'::jsonb,
        TRUE,
        NOW(),
        NOW()
    ),
    (
        'SK_DOMISILI',
        'Surat Keterangan Domisili Usaha',
        'Pelayanan penerbitan surat keterangan domisili bagi usaha mikro, kecil, dan menengah (UMKM) warga.',
        '{
            "$schema": "http://json-schema.org/draft-07/schema#",
            "type": "object",
            "required": ["nama_usaha", "jenis_usaha", "alamat_usaha", "lama_usaha_tahun"],
            "properties": {
                "nama_usaha": {
                    "type": "string",
                    "minLength": 2,
                    "description": "Nama resmi tempat/kegiatan usaha"
                },
                "jenis_usaha": {
                    "type": "string",
                    "description": "Bidang usaha (cth: Kuliner, Jasa, Perdagangan)"
                },
                "alamat_usaha": {
                    "type": "string",
                    "minLength": 5,
                    "description": "Alamat lokasi operasional usaha di kelurahan"
                },
                "lama_usaha_tahun": {
                    "type": "integer",
                    "minimum": 0,
                    "description": "Lama usaha berjalan (dalam tahun)"
                }
            }
        }'::jsonb,
        TRUE,
        NOW(),
        NOW()
    ),
    (
        'SK_BELUM_MENIKAH',
        'Surat Keterangan Belum Menikah',
        'Pelayanan keterangan status lajang/belum pernah menikah untuk keperluan administrasi pekerjaan atau pernikahan.',
        '{
            "$schema": "http://json-schema.org/draft-07/schema#",
            "type": "object",
            "required": ["keperluan"],
            "properties": {
                "keperluan": {
                    "type": "string",
                    "minLength": 3,
                    "description": "Keperluan pengajuan surat keterangan"
                },
                "catatan_tambahan": {
                    "type": "string",
                    "description": "Informasi pendukung opsional"
                }
            }
        }'::jsonb,
        TRUE,
        NOW(),
        NOW()
    )
ON CONFLICT (layanan_id) DO NOTHING;

-- 4. Seed Navigasi Menu Publik (Hierarki Maksimal 2 Level)
-- Header Items
INSERT INTO navigasi_menu (id, parent_id, label, url, urutan, is_active, created_at, updated_at)
VALUES 
    ('01ARZMENU00000000000000001', NULL, 'Beranda', '/', 1, TRUE, NOW(), NOW()),
    ('01ARZMENU00000000000000002', NULL, 'Profil Wilayah', '/profil', 2, TRUE, NOW(), NOW()),
    ('01ARZMENU00000000000000003', NULL, 'Layanan Publik', '/layanan', 3, TRUE, NOW(), NOW()),
    ('01ARZMENU00000000000000004', NULL, 'Berita & Pengumuman', '/berita', 4, TRUE, NOW(), NOW()),
    ('01ARZMENU00000000000000005', NULL, 'Kontak', '/kontak', 5, TRUE, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Dropdown Sub-menu untuk 'Profil Wilayah' (parent_id: 01ARZMENU00000000000000002)
INSERT INTO navigasi_menu (id, parent_id, label, url, urutan, is_active, created_at, updated_at)
VALUES 
    ('01ARZSUB000000000000000001', '01ARZMENU00000000000000002', 'Visi & Misi', '/profil#visi-misi', 1, TRUE, NOW(), NOW()),
    ('01ARZSUB000000000000000002', '01ARZMENU00000000000000002', 'Struktur Organisasi', '/profil#struktur', 2, TRUE, NOW(), NOW()),
    ('01ARZSUB000000000000000003', '01ARZMENU00000000000000002', 'Sejarah Kelurahan', '/profil#sejarah', 3, TRUE, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- 5. Seed Konten Publik Awal (Berita & Pengumuman)
INSERT INTO konten_publik (
    id, tipe, judul, slug, ringkasan, isi_konten, thumbnail_r2_key, is_published, published_at, author_id, author_nama, created_at, updated_at
)
VALUES 
    (
        '01ARZNEWS00000000000000001',
        'BERITA',
        'Penyaluran Bantuan Sosial Pangan Tahap II di Kelurahan Sukanegla',
        'penyaluran-bansos-tahap-2',
        'Pemerintah Kelurahan Sukanegla bersama perwakilan Dinas Sosial melaksanakan penyaluran bansos pangan kepada 250 KPM bertempat di Aula Kelurahan.',
        '<p>Pemerintah Kelurahan Sukanegla, Kecamatan Garut Kota, telah sukses menyalurkan Bantuan Sosial Pangan Tahap II kepada 250 Keluarga Penerima Manfaat (KPM) pada hari ini bertempat di Aula Kantor Kelurahan Sukanegla.</p><p>Lurah Sukanegla menyampaikan apresiasi kepada seluruh kader RW dan RT yang telah aktif memvalidasi data penerima melalui aplikasi SIDAK sehingga penyaluran berjalan tepat sasaran, tertib, dan tanpa antrean panjang.</p>',
        'public/cms/2026/berita_bansos_sukanegla.jpg',
        TRUE,
        NOW() - INTERVAL '2 days',
        '01ARZ3NDEKTSV4RRFFQ69G5001',
        'Drs. H. Mulyadi (Seklur)',
        NOW(),
        NOW()
    ),
    (
        '01ARZNEWS00000000000000002',
        'PENGUMUMAN',
        'Jadwal Pelayanan Keliling Perekaman e-KTP dan IKD RW 01 - RW 05',
        'jadwal-pelayanan-keliling-ikd',
        'Pelayanan jemput bola administrasi kependudukan (e-KTP dan Aktivasi IKD) akan diadakan bergilir mulai Senin mendatang.',
        '<p>Diberitahukan kepada seluruh warga Kelurahan Sukanegla bahwa Tim Pelayanan Administrasi Kependudukan Keliling akan hadir di lingkungan RW masing-masing mulai pukul 08.30 - 14.00 WIB.</p><p>Warga yang belum memiliki e-KTP atau ingin mengaktivasi Identitas Kependudukan Digital (IKD) dapat langsung mendatangi pos pelayanan RW terdekat dengan membawa fotokopi Kartu Keluarga (KK).</p>',
        'public/cms/2026/pengumuman_layanan_keliling.jpg',
        TRUE,
        NOW() - INTERVAL '1 day',
        '01ARZ3NDEKTSV4RRFFQ69G5002',
        'Siti Nurhaliza, S.AP (Kasi Pelayanan)',
        NOW(),
        NOW()
    )
ON CONFLICT (id) DO NOTHING;
