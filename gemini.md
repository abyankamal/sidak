# MASTER GEMINI PROTOCOL: SIDAK (SISTEM DATA KEWILAYAHAN) - MONOREPO

## 1. Identitas & Filosofi Sistem
SIDAK (Sistem Data Kewilayahan) adalah sistem digital terpadu administrasi kelurahan dan portal informasi publik berbasis monorepo yang dirancang untuk efisiensi tinggi pada infrastruktur terbatas.

### Prinsip Fondasi:
1. **Contract-Driven First:** Spesifikasi API di `contracts/openapi.yaml` adalah *Single Source of Truth (SSOT)* mutlak. Kode klien dan server tidak boleh menyimpang dari kontrak ini.
2. **Offline-First Resilience:** Aplikasi mobile dirancang tahan terhadap ketiadaan jaringan dengan memperlakukan SQLite lokal murni sebagai *Staging Area* & *Job Queue*.
3. **Zero-Proxy Static Storage:** Backend Golang tidak boleh menjadi perantara pengunduhan/pengunggahan berkas KTP/KK/PDF/Thumbnail. Semua lalu lintas berkas statis dialihkan langsung ke Cloudflare R2 via *Presigned URL*.
4. **Asynchronous Heavy I/O:** Pembuatan PDF via Gotenberg bersifat asinkron dan dibatasi bebannya agar tidak membebani komputasi server.
5. **Hybrid Web Architecture (Single Instance):** Portal publik (*company profile*) dan web admin disatukan dalam satu *instance* Next.js menggunakan *Route Groups* (`app/(public)` dan `app/(admin)`) untuk menghemat konsumsi RAM server.

---

## 2. Batasan Lingkungan & Infrastruktur Target
Semua keputusan penulisan kode backend, frontend, dan konfigurasi deployment harus mematuhi batasan perangkat keras berikut:
- **Server VPS:** Monolith node dengan spesifikasi **2 vCPU, 8 GB RAM, Ubuntu 24.04 LTS**.
- **Gotenberg Limit:** Maksimal alokasi 0.75 vCPU, 1 concurrent worker di level Golang.
- **Database:** PostgreSQL 16 dengan skema hibrida (Relasional + JSONB terindeks GIN).
- **Reverse Proxy:** Caddy v2 (Otomatisasi HTTPS + Subdomain routing).
- **Storage:** Cloudflare R2 (Protokol AWS S3 Compatible).
- **Caching Strategy:** Incremental Static Regeneration (ISR `revalidate = 300`) untuk halaman publik guna mengisolasi beban lalu lintas warga dari PostgreSQL.

---

## 3. Peta Navigasi Direktori Monorepo

Setiap modul bersifat otonom dan memiliki panduan spesifiknya masing-masing:

```text
sidak-system/
├── contracts/
│   └── openapi.yaml           # [SSOT] Kontrak API OpenAPI v3 (Pelayanan & CMS)
├── backend/                   # [MODUL] Golang API & Database Migrations
│   └── gemini.md              # Aturan ketat Go, JSON Schema, & Query PostgreSQL
├── web/                       # [MODUL] Next.js App Router (Public Portal + Admin)
│   ├── src/app/
│   │   ├── (public)/          # Halaman Publik (Landing Page, Profil, Berita, Navigasi)
│   │   └── (admin)/           # Dasbor Verifikasi, Pelayanan, & CMS Manager
│   └── gemini.md              # Aturan Server Components, ISR, Cookie Auth, & Polling
├── mobile/                    # [MODUL] Mobile Flutter Offline-First
│   └── gemini.md              # Aturan Drift SQLite, WorkManager, & Karantina File
├── deploy/                    # [INFRA] Docker Compose, Caddyfile, Cron Backup
├── Makefile                   # [DEV] Perintah terpusat local development
└── gemini.md                  # [ROOT] File master ini (Arsitektur Global & Alur MVP)
```

### Aturan Pergerakan Agen AI:
- Saat diminta mengubah atau menulis kode di folder `mobile/`, AI **wajib** membaca `mobile/gemini.md`.
- Saat diminta mengubah atau menulis kode di folder `backend/`, AI **wajib** membaca `backend/gemini.md`.
- Saat diminta mengubah atau menulis kode di folder `web/`, AI **wajib** membaca `web/gemini.md`.
- DILARANG mengimpor (*cross-import*) kode langsung antar-folder (`mobile` tidak boleh mengimpor dari `backend` atau `web`, dan sebaliknya).

---

## 4. Pagar Pembatas Global (Universal Architectural Rules)

Setiap agen AI yang memproses tugas di modul mana pun WAJIB mematuhi aturan universal berikut:

### A. Integritas Data, Validasi & Routing
1. **Aturan SSOT Kontrak:** Semua nama atribut data (JSON key) wajib mengambil referensi dari `contracts/openapi.yaml`. DILARANG membuat nama variabel atau *field* baru yang tidak tercatat di file kontrak.
2. **API Versioning Wajib:** Seluruh rute API wajib menggunakan awalan versi eksplisit (contoh: `/api/v1/sync/commit`, `/api/v1/public/profil`, `/api/v1/cms/menu`).
3. **Validasi Dua Tingkat Transaksi:**
   - *Tingkat 1 (Format Jaringan):* OpenAPI memvalidasi kolom statis/amplop.
   - *Tingkat 2 (Logika Bisnis):* Golang memvalidasi `data_isian` dinamis menggunakan skema yang dimuat dari database.
4. **Penyimpanan Skema Dinamis:** Aturan JSON Schema untuk setiap jenis formulir WAJIB disimpan di tabel `template_form` di PostgreSQL, kemudian dimuat (*cached*) ke dalam memori RAM Golang (`sync.Map`) saat startup sistem.
5. **Pencegahan Duplikasi Logis:** Backend Golang wajib menolak pengajuan jika `warga_nik` yang sama mengajukan `layanan_id` yang sama dalam kurun waktu 24 jam terakhir dan status transaksi sebelumnya masih `menunggu_review`.

### B. Desain Modul CMS & Company Profile
1. **Singleton Profil Wilayah:** Data profil kelurahan (visi, misi, kontak) disimpan di tabel `profil_wilayah` yang dikunci hanya memiliki 1 baris (`id = 1`). Operasi diizinkan hanya `UPDATE` atau `GET`.
2. **Navigasi Terkontrol (Max 2 Tingkat):** Menu publik dikelola di tabel `navigasi_menu` dengan relasi *self-referencing* `parent_id` yang dibatasi maksimal 2 level hierarki (*header* & *dropdown child*). Dilarang membuat pohon navigasi tak terbatas.
3. **Optimasi Konten Berita:** Berita dan pengumuman disimpan di tabel `konten_publik` dengan `slug` terindeks unik untuk *routing* SEO di Next.js `app/(public)/berita/[slug]`.
4. **On-Demand Cache Invalidation:** Setiap pembaruan data menu, profil, atau berita di dasbor admin wajib memicu `revalidatePath()` di Next.js agar konten publik terbarui tanpa mematikan *cache* ISR.

### C. Otentikasi, RBAC Sederhana & Status Workflow
1. **Simplified RBAC:** Peran pengguna (`role`) wajib diintegrasikan langsung ke dalam tabel utama `users` (`SEKLUR`, `KASI`, `KADER`). DILARANG membuat tabel terpisah seperti `jabatan` atau hierarki relasional berlebih.
2. **Workflow Review State Machine:** Status alur dokumen pelayanan menggunakan alur *review* catatan:
   - `menunggu_review` (Status awal setelah data masuk/sinkron)
   - `sudah_di_review` (Disetujui/Diverifikasi oleh Kasi/Sekdes dengan catatan)
   - `butuh_revisi` (Memerlukan perbaikan berkas dari warga/kader)
   - DILARANG menggunakan mekanisme penolakan mutlak (*hard rejection*) tanpa riwayat catatan review.

### D. Keamanan & Akses Data Sensitif (PII) vs Publik
1. **Pemisahan Akses R2:**
   - *Private Storage (Strict 5 Min Presigned URL):* Lampiran KTP, KK, berkas rahasia warga, dan PDF cetak surat.
   - *Public/CDN Storage:* Thumbnail berita publik dan bagan struktur organisasi kelurahan.
2. **Isolasi Path Penyimpanan:** Format path berkas di R2:
   - Dokumen Warga: `lampiran/{warga_nik}/{transaksi_id}_{file_name}`
   - Media Publik: `public/cms/{tahun}/{slug}_{file_name}`

### E. Efisiensi Komputasi & I/O
1. **Bypass VPS File Traffic:** Klien (Next.js & Flutter) wajib mengunggah dan mengunduh berkas langsung dari Cloudflare R2. Backend Golang DILARANG bertindak sebagai *file proxy* atau penyimpan berkas sementara di disk VPS.
2. **Gotenberg Rate-Limiting:** Operasi render PDF via Gotenberg wajib dibatasi maksimal **1 concurrent worker** via Go channel buffer.
3. **Idempotency Mandatory:** Semua mutasi data dari Flutter wajib menyertakan HTTP header `Idempotency-Key` (ULID) dan ditangani secara atomik via `ON CONFLICT (id) DO NOTHING`.

---

## 5. Roadmap Eksekusi MVP & Gerbang Validasi (Validation Gates)

Pengembangan sistem WAJIB dikerjakan secara linier sesuai fase berikut. Agen AI DILARANG menghasilkan kode untuk Fase berikutnya sebelum kriteria selesai (*exit gate*) pada fase berjalan terpenuhi.

```text
┌────────────────────────┐     ┌────────────────────────┐     ┌────────────────────────┐
│        FASE 1          │ ──► │        FASE 2          │ ──► │        FASE 3          │
│ Kontrak OpenAPI & DB   │     │ Backend Core, CMS API  │     │ Integrasi R2           │
│ (Pelayanan + Profil)   │     │      & Unit Test       │     │     & Gotenberg        │
└────────────────────────┘     └────────────────────────┘     └────────────────────────┘
                                                                           │
                               ┌────────────────────────┐                  ▼
                               │        FASE 5          │     ┌────────────────────────┐
                               │     Mobile Client      │ ◄── │        FASE 4          │
                               │   (Flutter Offline)    │     │  Web (Public + Admin)  │
                               └────────────────────────┘     └────────────────────────┘
```

### Fase 1: Kontrak API & Skema Database
- **Fokus:** Menulis `contracts/openapi.yaml` dan skrip SQL migrasi (`users`, `template_form`, `transaksi_pelayanan`, `profil_wilayah`, `navigasi_menu`, `konten_publik`).
- **Exit Gate:** 
  - File `contracts/openapi.yaml` valid tanpa syntax error untuk rute transaksi dan CMS.
  - Skrip migrasi berhasil dieksekusi oleh `golang-migrate` ke PostgreSQL lokal.

### Fase 2: Backend Core, CMS API & Unit Testing
- **Fokus:** Routing Golang, database connection pool (`pgxpool`), in-memory concurrency lock (`sync.Map`), in-memory cache JSON Schema, deduplikasi logis 24 jam, CRUD CMS publik, dan middleware auth JWT/Cookie.
- **Exit Gate:** 
  - Unit test `go test -race ./...` lolos 100%.
  - Endpoint `/api/v1/sync/commit` mengembalikan 201 untuk data baru, 400 untuk schema tidak valid/duplikasi logis, 409 untuk retry data duplikat, dan 429 untuk request bersamaan.
  - Endpoint publik `/api/v1/public/*` dapat diakses tanpa autentikasi.

### Fase 3: Integrasi Eksternal (Cloudflare R2 & Gotenberg)
- **Fokus:** Presigned URL generation dan implementasi worker pool PDF asinkron dengan Go buffered channel.
- **Exit Gate:** 
  - File berhasil diunggah langsung ke R2 via Presigned PUT.
  - Endpoint `/api/v1/layanan/{id}/generate-pdf` mengembalikan HTTP 202, dan PDF berhasil dirender lalu diunggah ke R2 tanpa lonjakan memori VPS.

### Fase 4: Web Portal Publik & Dashboard Admin (Next.js)
- **Fokus:** 
  - `app/(public)`: Landing page, profil, arsip/detail berita dengan ISR (`revalidate = 300`), dan navigasi menu dinamis.
  - `app/(admin)`: Middleware proteksi cookie, verifikasi berkas JSONB & preview lampiran R2, polling status PDF, serta panel manajemen CMS menu & konten.
- **Exit Gate:** 
  - Halaman publik ter-render cepat berbasis cache statis.
  - Admin/Kasi dapat memverifikasi permohonan dan mengelola menu/konten publik dengan `revalidatePath` otomatis.

### Fase 5: Mobile Client Offline-First (Flutter)
- **Fokus:** Tabel antrean Drift SQLite (`SyncQueue`), isolasi berkas di folder `/offline_staging/`, WorkManager background isolate, dan Riverpod reactive stream.
- **Exit Gate:** 
  - Pengisian form saat mode pesawat (offline) tersimpan aman di SQLite & staging folder.
  - Saat jaringan pulih, WorkManager sukses mengeksekusi 3 fase sync, berkas lokal dibersihkan, dan UI otomatis ter-update.

---

## 6. Protokol Operasional Agen AI

Saat menerima prompt pengerjaan:
1. **Identifikasi Fase:** Pastikan tugas yang diminta berada pada fase yang sedang aktif.
2. **Baca Konteks Khusus:** Buka file `gemini.md` di folder modul terkait (`backend/`, `web/`, atau `mobile/`).
3. **Verifikasi Kontrak:** Pastikan struktur *request body*, *response*, dan *headers* sesuai dengan `contracts/openapi.yaml` (menggunakan rute `/api/v1/*`).
4. **Tolak Scope Creep:** Tolak penambahan fitur baru di luar spesifikasi 5 fase di atas.
