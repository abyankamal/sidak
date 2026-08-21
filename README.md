# SIDAK (Sistem Data Kewilayahan)

SIDAK adalah platform digital terintegrasi untuk pelayanan administrasi kelurahan/desa dan portal informasi publik berbasis monorepo. Sistem ini dirancang untuk bekerja secara efisien, tangguh, dan optimal pada infrastruktur server dengan sumber daya terbatas (VPS 2 vCPU, 8 GB RAM), serta mendukung pengumpulan data warga di lapangan secara *offline-first*.

---

## 🏛️ Arsitektur & Filosofi Sistem

1. **Contract-Driven First:** Spesifikasi API OpenAPI v3 (`contracts/openapi.yaml`) adalah *Single Source of Truth (SSOT)* mutlak bagi backend, web portal, dan aplikasi mobile.
2. **Offline-First Resilience:** Aplikasi mobile (Flutter) menggunakan SQLite lokal sebagai *staging area* dan *job queue* agar kader dapat melayani warga tanpa jaringan internet.
3. **Local Volume Storage:** Penyimpanan berkas lampiran warga (KTP/KK) dan media publik (CMS) dikelola langsung di volume disk server lokal (`/uploads/`) yang terisolasi dan aman.
4. **Asynchronous Heavy I/O:** Pembuatan PDF surat dinas diproses secara asinkron menggunakan Gotenberg Chromium Engine dengan *worker pool* tunggal (1 concurrent worker) untuk menjaga kestabilan memori server.
5. **Hybrid Single-Instance Web:** Portal profil publik (*Company Profile*) dan Dasbor Verifikasi Admin disatukan dalam satu instance Next.js App Router (`app/(public)` & `app/(admin)`) untuk efisiensi RAM.

---

## 📁 Struktur Monorepo

```text
sidak/
├── contracts/
│   └── openapi.yaml           # [SSOT] Kontrak API OpenAPI v3 (Pelayanan, Auth, CMS)
├── backend/                   # [MODUL] Golang API & Database Migrations (Go 1.22+)
│   ├── cmd/api/               # Entrypoint server API
│   ├── config/                # Environment configuration loader
│   ├── internal/
│   │   ├── domain/            # Domain models & DTOs
│   │   ├── handler/           # HTTP handlers & Chi router
│   │   ├── middleware/        # JWT & Cookie Auth, Role guards
│   │   ├── repository/        # PostgreSQL pgx queries
│   │   └── service/           # Logic bisnis, JSON Schema cache, & PDF engine
│   ├── migrations/            # Skrip migrasi DDL & Seed data PostgreSQL
│   └── test/                  # Test suite integrasi & unit testing (-race)
├── web/                       # [MODUL] Next.js App Router (Public Portal + Admin)
│   └── src/app/
│       ├── (public)/          # Halaman publik (Berita, Profil, Pelayanan Warga)
│       └── (admin)/           # Dasbor verifikasi permohonan & CMS manager
├── mobile/                    # [MODUL] Flutter Mobile (Offline-First untuk Kader)
├── deploy/                    # [INFRA] Docker Compose (PostgreSQL, Gotenberg) & Caddy
├── Makefile                   # [DEV] Perintah otomasi development & testing
└── README.md                  # Dokumentasi utama proyek
```

---

## 👥 Peran Pengguna & Autentikasi (RBAC)

Autentikasi mendukung login fleksibel menggunakan **NIK (16 Digit)** atau **NIP (18 Digit)** dan kata sandi tanpa ketergantungan email.

| Role | Identitas | Hak Akses Utama |
| :--- | :--- | :--- |
| **LURAH** | NIP (18 Digit) | Meninjau/menyetujui permohonan warga, cetak PDF, kelola CMS profil & berita. |
| **SEKLUR** | NIP (18 Digit) | Verifikasi permohonan surat, cetak PDF, kelola menu navigasi & CMS. |
| **KASI** | NIP (18 Digit) | Verifikasi berkas permohonan pelayanan teknis kewilayahan & cetak PDF. |
| **KADER** | NIK (16 Digit) | Penginputan form & sinkronisasi data warga via aplikasi mobile di lapangan. |

### Akun Bawaan (Default Seed):
- **Lurah:** NIP `197503151998031001` (Password: `AdminSidak2026!`)
- **Seklur:** NIP `198001012005011001` (Password: `AdminSidak2026!`)
- **Kasi Pelayanan:** NIP `198502022010012002` (Password: `AdminSidak2026!`)
- **Kader RW 01:** NIK `3205010303920003` (Password: `AdminSidak2026!`)

---

## 🛠️ Prasyarat & Instalasi Lokal

### 1. Kebutuhan Sistem
- **Docker & Docker Compose** (atau **Colima** untuk macOS ringan)
- **Go 1.22+**
- **Node.js 20+** (untuk modul Web Next.js)
- **Flutter 3.x** (untuk modul Mobile)

### 2. Menjalankan Infrastruktur (PostgreSQL & Gotenberg)

Bagi pengguna macOS dengan Colima:
```bash
# Nyalakan runtime Colima
make colima-start
```

Salin berkas konfigurasi lingkungan:
```bash
cp .env.example .env
```

Jalankan container database dan PDF engine via Docker Compose:
```bash
make infra-up
```

### 3. Migrasi & Seed Database
Jalankan migrasi tabel dan data awal:
```bash
make migrate-up
```

### 4. Menjalankan Backend API
```bash
cd backend
go run ./cmd/api
```
Server API backend akan berjalan di `http://localhost:8080`.

---

## 🧪 Menjalankan Pengujian (Testing)

Jalankan seluruh test suite backend dengan detektor race condition:
```bash
make test-backend
```
Atau langsung di dalam direktori `backend/`:
```bash
cd backend && go test -v -count=1 -race ./...
```

---

## 📑 Rute API Utama (`/api/v1`)

- **Autentikasi:**
  - `POST /api/v1/auth/login` - Login via NIK/NIP + Password (JWT Bearer & Cookie).
  - `POST /api/v1/auth/logout` - Logout dan pembersihan sesi cookie.
  - `GET /api/v1/auth/me` - Informasi profil pengguna aktif.
- **Penyimpanan Berkas (Local Storage):**
  - `POST /api/v1/storage/upload` - Unggah berkas multipart (lampiran, CMS, dokumen).
- **Pelayanan & Sinkronisasi:**
  - `GET /api/v1/template-form` - Mengambil daftar form & JSON Schema draft-07.
  - `POST /api/v1/sync/commit` - Commit transaksi permohonan (dengan `Idempotency-Key`).
  - `GET /api/v1/transaksi` - Daftar antrean permohonan warga.
  - `GET /api/v1/transaksi/{id}` - Detail data isian dan preview berkas lampiran.
  - `PATCH /api/v1/transaksi/{id}/review` - Verifikasi & catatan review oleh Lurah/Seklur/Kasi.
- **Dokumen PDF:**
  - `POST /api/v1/layanan/{id}/generate-pdf` - Antrean asinkron render PDF Gotenberg (202 Accepted).
  - `GET /api/v1/dokumen/{id}/status` - Polling status render & link unduh PDF.
- **CMS Publik (Unauthenticated):**
  - `GET /api/v1/public/profil` - Profil kelurahan (visi, misi, kontak).
  - `GET /api/v1/public/menu` - Struktur navigasi menu (hierarki 2 level).
  - `GET /api/v1/public/konten` - Daftar berita dan pengumuman terbit.
  - `GET /api/v1/public/konten/{slug}` - Detail isi berita/pengumuman.
