-- =============================================================================
-- MIGRATION 000002 DOWN: CLEAN SEED DATA
-- =============================================================================

DELETE FROM konten_publik WHERE id IN ('01ARZNEWS00000000000000001', '01ARZNEWS00000000000000002');
DELETE FROM navigasi_menu WHERE id LIKE '01ARZMENU%' OR id LIKE '01ARZSUB%';
DELETE FROM template_form WHERE layanan_id IN ('SKTM', 'SK_DOMISILI', 'SK_BELUM_MENIKAH');
DELETE FROM profil_wilayah WHERE id = 1;
DELETE FROM users WHERE id IN ('01ARZ3NDEKTSV4RRFFQ69G5001', '01ARZ3NDEKTSV4RRFFQ69G5002', '01ARZ3NDEKTSV4RRFFQ69G5003');
