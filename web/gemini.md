# PROTOKOL PENGEMBANGAN: WEB (NEXT.JS HYBRID SINGLE INSTANCE)

## 1. Tanggung Jawab Modul
- Next.js App Router (Single Instance).
- `app/(public)`: Company profile, profil kelurahan, berita publik dengan ISR (`revalidate = 300`), navigasi dinamis.
- `app/(admin)`: Dasbor verifikasi berkas, review permohonan, preview lampiran R2, polling status PDF, dan CMS manager (Menu & Berita).

## 2. Aturan & Pagar Pembatas
1. **SSOT Kontrak:** Semua konsumsi API mengacu ke `contracts/openapi.yaml`.
2. **Cookie Auth:** Autentikasi sesi admin menggunakan HTTP-only cookie.
3. **On-Demand Revalidation:** Mutasi CMS pada admin wajib memicu `revalidatePath()` untuk menjaga kesegaran cache ISR publik.
4. **Direct R2:** File media dan lampiran diakses langsung via Presigned URL / CDN Cloudflare R2.
