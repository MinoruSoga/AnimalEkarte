-- =============================================================================
-- Animal Ekarte - マスタデータ v2.1
-- PostgreSQL 18
-- 冪等性保証: ON CONFLICT DO NOTHING / DO UPDATE / WHERE NOT EXISTS
-- 内容: システム共通マスタデータ（clinic_id なし・全クリニック共通）
-- 統合: ext-011 (lstep_tag_code_mappings デフォルト)
--       mig-010 (lstep_auto_managed_prefixes / lstep_condition_tag_mappings / lstep_send_purpose_tag_prefixes seed)
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
-- 3. lstep_auto_managed_prefixes（自動管理タグプレフィックス, mig-010 統合）
-- B / C1 / C2 / C3 カテゴリの seed
-- -----------------------------------------------------------------------------
INSERT INTO lstep_auto_managed_prefixes (prefix, category, description) VALUES
    -- B カテゴリ（予約・キャンセル・未来院）
    ('reserved_',      'B',  '予約タグ'),
    ('canceled_visit', 'B',  'キャンセルタグ'),
    ('no_show_',       'B',  '未来院タグ'),
    -- C1 カテゴリ（ペット基本情報）
    ('breed_',         'C1', '犬種・猫種タグ'),
    ('sex_',           'C1', '性別タグ'),
    ('pet_birthday_',  'C1', '誕生日タグ'),
    ('birth_year_',    'C1', '生年タグ'),
    ('spay_neutered',  'C1', '避妊去勢済みタグ'),
    ('intact',         'C1', '未避妊去勢タグ'),
    -- C2 カテゴリ（診療・ケアスケジュール）
    ('vaccine_',       'C2', 'ワクチンタグ'),
    ('checkup_',       'C2', '健診タグ'),
    ('next_checkup_',  'C2', '次回健診タグ'),
    ('refill_due_',    'C2', '処方箋補充期限タグ'),
    ('next_visit_',    'C2', '次回来院タグ'),
    ('chronic_',       'C2', '慢性疾患タグ'),
    -- C3 カテゴリ（証明書・術後・退院・死亡）
    ('cert_sent_',         'C3', '証明書送付タグ'),
    ('post_surgery_',      'C3', '術後フォロータグ'),
    ('post_discharge_',    'C3', '退院フォロータグ'),
    ('pet_deceased_',      'C3', 'ペット死亡タグ')
ON CONFLICT (prefix) DO UPDATE
    SET category = EXCLUDED.category,
        description = EXCLUDED.description,
        updated_at = now();

-- -----------------------------------------------------------------------------
-- 4. lstep_condition_tag_mappings（慢性疾患コード→タグ名マッピング, mig-010 統合）
-- -----------------------------------------------------------------------------
INSERT INTO lstep_condition_tag_mappings (condition_code, tag_name, description) VALUES
    ('ckd',      'chronic_ckd',      '慢性腎臓病'),
    ('heart',    'chronic_heart',    '心臓病'),
    ('skin',     'chronic_skin',     '皮膚疾患'),
    ('diabetes', 'chronic_diabetes', '糖尿病'),
    ('liver',    'chronic_liver',    '肝疾患'),
    ('thyroid',  'chronic_thyroid',  '甲状腺疾患'),
    ('other',    'chronic_other',    'その他慢性疾患')
ON CONFLICT (condition_code) DO UPDATE
    SET tag_name = EXCLUDED.tag_name,
        description = EXCLUDED.description,
        updated_at = now();

-- -----------------------------------------------------------------------------
-- 5. lstep_send_purpose_tag_prefixes（LINE送信目的→タグプレフィックスマッピング, mig-010 統合）
-- -----------------------------------------------------------------------------
INSERT INTO lstep_send_purpose_tag_prefixes (purpose, tag_prefix, description) VALUES
    ('vaccine_cert',       'cert_sent_vaccine_',        'ワクチン証明書送付'),
    ('inspection_result',  'cert_sent_inspection_',     '検査結果送付'),
    ('post_surgery',       'post_surgery_followup_',    '術後フォローアップ'),
    ('post_discharge',     'post_discharge_followup_',  '退院フォローアップ')
ON CONFLICT (purpose) DO UPDATE
    SET tag_prefix = EXCLUDED.tag_prefix,
        description = EXCLUDED.description,
        updated_at = now();
