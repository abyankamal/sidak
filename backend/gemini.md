# PROTOKOL PENGEMBANGAN: BACKEND (GOLANG & POSTGRESQL)

## 1. Tanggung Jawab Modul
- REST API Server dengan Golang (Go 1.22+).
- Migrasi database PostgreSQL 16 (`golang-migrate`).
- Validasi transaksi 2 tingkat (OpenAPI envelope + dynamic JSON Schema).
- In-memory cache skema (`sync.Map`) & concurrency lock per `Idempotency-Key` (ULID).
- Penyimpanan berkas lokal server (`/uploads/`) terisolasi untuk lampiran privat warga dan media publik CMS.
- Gotenberg worker pool (maks. 1 concurrent) untuk render PDF.

## 2. Aturan & Pagar Pembatas
1. **SSOT Kontrak:** Semua rute dan struktur response/request WAJIB mengikuti `contracts/openapi.yaml`.
2. **Autentikasi NIK / NIP:** Login menggunakan `identifier` (NIK 16 digit untuk Kader atau NIP 18 digit untuk PNS: Lurah, Seklur, Kasi) + password.
3. **Penyimpanan Lokal Terisolasi:** Berkas warga disimpan di `uploads/lampiran/{warga_nik}/{transaksi_id}_{file_name}` dan media publik di `uploads/public/cms/{tahun}/{slug}_{file_name}`.
4. **Idempotency:** Wajib menangani HTTP header `Idempotency-Key` (ULID) secara atomik via `ON CONFLICT (id) DO NOTHING`.
5. **Deduplikasi 24 Jam:** Tolak jika NIK yang sama mengajukan layanan_id yang sama dalam 24 jam terakhir saat status `menunggu_review`.
