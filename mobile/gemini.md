# PROTOKOL PENGEMBANGAN: MOBILE (FLUTTER OFFLINE-FIRST)

## 1. Tanggung Jawab Modul
- Aplikasi Mobile Flutter (Android/iOS) untuk Kader/Petugas Lapangan.
- Offline-First Architecture dengan Drift SQLite (`SyncQueue`).
- Karantina & isolasi file di staging area lokal `/offline_staging/`.
- Sinkronisasi 3 Fase via background task (WorkManager):
  1. Minta Presigned PUT URL R2.
  2. Upload langsung berkas ke R2.
  3. Commit transaksi ke `/api/v1/sync/commit` beserta Idempotency-Key.

## 2. Aturan & Pagar Pembatas
1. **SSOT Kontrak:** Semua payload dan request/response wajib tunduk pada `contracts/openapi.yaml`.
2. **Offline-First:** Form input wajib bisa disimpan offline dan otomatis disinkronkan saat koneksi tersedia.
3. **Pembersihan File:** Berkas di staging lokal harus dibersihkan setelah sinkronisasi berhasil.
