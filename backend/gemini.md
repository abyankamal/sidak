# PROTOKOL PENGEMBANGAN: BACKEND (GOLANG & POSTGRESQL)

## 1. Tanggung Jawab Modul
- REST API Server dengan Golang (Go 1.22+).
- Migrasi database PostgreSQL 16 (`golang-migrate`).
- Validasi transaksi 2 tingkat (OpenAPI envelope + dynamic JSON Schema).
- In-memory cache skema (`sync.Map`) & concurrency lock.
- Integrasi Cloudflare R2 Presigned URLs & Gotenberg worker pool (maks. 1 concurrent).

## 2. Aturan & Pagar Pembatas
1. **SSOT Kontrak:** Semua rute dan struktur response/request WAJIB mengikuti `contracts/openapi.yaml`.
2. **Zero-Proxy:** Backend HANYA menghasilkan Presigned URL untuk R2, tidak menjadi perantara transmisi file statis.
3. **Idempotency:** Wajib menangani HTTP header `Idempotency-Key` (ULID) secara atomik.
4. **Deduplikasi 24 Jam:** Tolak jika NIK yang sama mengajukan layanan_id yang sama dalam 24 jam terakhir saat status `menunggu_review`.
