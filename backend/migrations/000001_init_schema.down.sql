-- =============================================================================
-- MIGRATION 000001 DOWN: DROP SCHEMA
-- =============================================================================

DROP TABLE IF EXISTS dokumen_output CASCADE;
DROP TABLE IF EXISTS konten_publik CASCADE;
DROP TABLE IF EXISTS navigasi_menu CASCADE;
DROP TABLE IF EXISTS profil_wilayah CASCADE;
DROP TABLE IF EXISTS transaksi_pelayanan CASCADE;
DROP TABLE IF EXISTS template_form CASCADE;
DROP TABLE IF EXISTS users CASCADE;
