-- =============================================================================
-- Animal Ekarte - マスタデータ v2.0
-- PostgreSQL 18
-- 冪等性保証: ON CONFLICT DO NOTHING / DO UPDATE / WHERE NOT EXISTS
-- 内容: システム共通マスタデータ（clinic_id なし・全クリニック共通）
-- 統合: ext-011 (lstep_tag_code_mappings デフォルト)
-- 依存: 001_init.sql
-- 実行順: 001_init.sql → 002_seed_master.sql → 003_seed_demo.sql
-- 注記: ext-019 (permission_group_rules) は 003_seed_demo.sql 末尾に追加
--       (group_id FK が 003 で生成されるため 002 には置けない)
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

-- -----------------------------------------------------------------------------
-- 3. lstep_tag_code_mappings デフォルト（ext-011 統合）
--    専門検診候補タグのコードタイプ × 全クリニック。age_min は暫定値。
-- -----------------------------------------------------------------------------
INSERT INTO lstep_tag_code_mappings (clinic_id, tag_name, code_type, codes, age_min)
SELECT c.id, 'HLTH_専門検診候補', 'specialty_dental', '{}', 5
FROM clinics c
WHERE NOT EXISTS (
    SELECT 1 FROM lstep_tag_code_mappings m
    WHERE m.clinic_id = c.id
      AND m.tag_name  = 'HLTH_専門検診候補'
      AND m.code_type = 'specialty_dental'
      AND m.deleted_at IS NULL
);

INSERT INTO lstep_tag_code_mappings (clinic_id, tag_name, code_type, codes, age_min)
SELECT c.id, 'HLTH_専門検診候補', 'specialty_skin_ear', '{}', 5
FROM clinics c
WHERE NOT EXISTS (
    SELECT 1 FROM lstep_tag_code_mappings m
    WHERE m.clinic_id = c.id
      AND m.tag_name  = 'HLTH_専門検診候補'
      AND m.code_type = 'specialty_skin_ear'
      AND m.deleted_at IS NULL
);

INSERT INTO lstep_tag_code_mappings (clinic_id, tag_name, code_type, codes, age_min)
SELECT c.id, 'HLTH_専門検診候補', 'specialty_ophthalmology', '{}', 7
FROM clinics c
WHERE NOT EXISTS (
    SELECT 1 FROM lstep_tag_code_mappings m
    WHERE m.clinic_id = c.id
      AND m.tag_name  = 'HLTH_専門検診候補'
      AND m.code_type = 'specialty_ophthalmology'
      AND m.deleted_at IS NULL
);

INSERT INTO lstep_tag_code_mappings (clinic_id, tag_name, code_type, codes, age_min)
SELECT c.id, 'HLTH_専門検診候補', 'specialty_kidney', '{}', 7
FROM clinics c
WHERE NOT EXISTS (
    SELECT 1 FROM lstep_tag_code_mappings m
    WHERE m.clinic_id = c.id
      AND m.tag_name  = 'HLTH_専門検診候補'
      AND m.code_type = 'specialty_kidney'
      AND m.deleted_at IS NULL
);
