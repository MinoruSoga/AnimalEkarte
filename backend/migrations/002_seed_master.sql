-- =============================================================================
-- Animal Ekarte - マスタデータ v1.0
-- PostgreSQL 18
-- 冪等性保証: ON CONFLICT DO NOTHING / DO UPDATE
-- 内容: システム共通マスタデータ（clinic_id なし・全クリニック共通）
-- 依存: 001_init.sql
-- 実行順: 001_init.sql → 002_seed_master.sql → 003_seed_demo.sql
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 1. companies（本部情報: 1件）
-- -----------------------------------------------------------------------------
INSERT INTO companies (id, name) VALUES
    (1, 'ノア動物病院')
ON CONFLICT (id) DO UPDATE
    SET name = EXCLUDED.name,
        updated_at = now();

SELECT setval(pg_get_serial_sequence('companies', 'id'), (SELECT MAX(id) FROM companies));

-- -----------------------------------------------------------------------------
-- 2. animal_species（ペット種類: 6件、システム共通・clinic_idなし）
-- -----------------------------------------------------------------------------
INSERT INTO animal_species (id, name, is_active, sort_order) VALUES
    (1, '犬',         true, 1),
    (2, '猫',         true, 2),
    (3, '鳥',         true, 3),
    (4, 'うさぎ',     true, 4),
    (5, 'ハムスター', true, 5),
    (6, 'その他',     true, 6)
ON CONFLICT (id) DO UPDATE
    SET name = EXCLUDED.name,
        is_active = EXCLUDED.is_active,
        sort_order = EXCLUDED.sort_order,
        updated_at = now();

SELECT setval(pg_get_serial_sequence('animal_species', 'id'), (SELECT MAX(id) FROM animal_species));
