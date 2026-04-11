-- =============================================================================
-- Animal Ekarte - デモデータ v1.0
-- PostgreSQL 18
-- 冪等性保証: ON CONFLICT DO NOTHING / DO UPDATE
-- 内容: クリニック設定・スタッフ・飼主・ペット・診療記録・会計等（デモ/テスト用）
-- 依存: 001_init.sql → 002_seed_master.sql
-- 実行順: 001_init.sql → 002_seed_master.sql → 003_seed_demo.sql
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 1. clinics（クリニック: 3件）
-- -----------------------------------------------------------------------------
INSERT INTO clinics (id, company_id, name) VALUES
    (3, 1, '八王子院'),
    (4, 1, '城東医院'),
    (5, 1, '敷島医院')
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('clinics', 'id'), (SELECT MAX(id) FROM clinics));

-- -----------------------------------------------------------------------------
-- 4. occupations（職種: 各医院4件 × 3医院 = 12件）
-- -----------------------------------------------------------------------------
INSERT INTO occupations (id, clinic_id, name, is_active, sort_order) VALUES
    -- 八王子院 (clinic_id=3)
    (1,  3, '獣医師',   true, 1),
    (2,  3, '看護師',   true, 2),
    (3,  3, 'トリマー', true, 3),
    (4,  3, '受付',     true, 4),
    -- 城東医院 (clinic_id=4)
    (5,  4, '獣医師',   true, 1),
    (6,  4, '看護師',   true, 2),
    (7,  4, 'トリマー', true, 3),
    (8,  4, '受付',     true, 4),
    -- 敷島医院 (clinic_id=5)
    (9,  5, '獣医師',   true, 1),
    (10, 5, '看護師',   true, 2),
    (11, 5, 'トリマー', true, 3),
    (12, 5, '受付',     true, 4)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('occupations', 'id'), (SELECT MAX(id) FROM occupations));

-- -----------------------------------------------------------------------------
-- 5. staffs（スタッフ: 32件）
-- 八王子院: ID 1-15（人間11名 + リソース4件）
-- 城東医院: ID 16-25（人間7名 + リソース3件）
-- 敷島医院: ID 26-32（人間5名 + リソース2件）
-- account_id は accounts INSERT 後に UPDATE で設定される
-- clinic_id は staff_clinic_assignments で管理される
-- -----------------------------------------------------------------------------
INSERT INTO staffs (id, account_id, name, is_active, license_number, occupation_id, sort_order, staff_type, reservation_visible) VALUES
    -- 八王子院 人間スタッフ (clinic_id=3)
    (1,  NULL, '林 文明',              true, 'V-10001', 1, 1,  'doctor',   true),
    (2,  NULL, '山﨑 晶子',           true, 'V-10002', 1, 2,  'doctor',   true),
    (3,  NULL, '三井 隆之',           true, 'V-10003', 1, 3,  'doctor',   true),
    (4,  NULL, 'ノア',                 true, 'V-10004', 1, 4,  'doctor',   true),
    (5,  NULL, '加藤 茉里',           true, '',        2, 5,  'nurse',    false),
    (6,  NULL, '金谷 亜美',           true, '',        2, 6,  'nurse',    false),
    (7,  NULL, '稲村 一真',           true, '',        2, 7,  'nurse',    false),
    (8,  NULL, '安田 希恵',           true, '',        2, 8,  'nurse',    false),
    (9,  NULL, '倉田 春香',           true, '',        2, 9,  'nurse',    false),
    (10, NULL, '梶原 梨夢',           true, '',        2, 10, 'nurse',    false),
    (11, NULL, '髙木 賀央里',         true, '',        2, 11, 'nurse',    false),
    -- 八王子院 リソース（予約枠管理用）
    (12, NULL, 'お手入れ・オゾン療法', true, '',        3, 12, 'resource', true),
    (13, NULL, '健診・ワクチン・狂犬病', true, '',     3, 13, 'resource', true),
    (14, NULL, 'ドッグラン(アジリティ解放)', true, '', 4, 14, 'resource', true),
    (15, NULL, 'クイックシャンプー',   true, '',        3, 15, 'resource', true)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('staffs', 'id'), (SELECT MAX(id) FROM staffs));

-- -----------------------------------------------------------------------------
-- 6. accounts（認証用アカウント: 13件）
-- password_hash: bcrypt("password", cost=10)
-- account→staff mapping:
--   1(admin@noavet.jp)→staff 4(三井 隆之 / 院長代理), 2(hayashi)→staff 1(林 文明)
--   3(yamazaki)→staff 2(山﨑 晶子), 4(mitsui)→staff 3(三井 隆之)
--   5(admin@example.com)→staff 8(安田), 6(vet)→staff 9(倉田)
--   7(nurse)→staff 10(梶原), 8(reception)→staff 11(髙木)
--   9(system)→NULL, 10(kimura)→staff 16(木村 健太)
--   11(sasaki)→staff 17(佐々木), 12(fujiwara)→staff 26(藤原)
--   13(matsumoto)→staff 27(松本)
--   14(trimmer@example.com)→新staff(さくら/八王子院デモ)
--   15(joto-vet@example.com)→新staff(城東獣医デモ)
--   16(shiki-vet@example.com)→新staff(敷島獣医デモ)
-- -----------------------------------------------------------------------------
INSERT INTO accounts (id, email, password_hash, is_active, is_system_admin) VALUES
    -- システム管理者（全院アクセス）
    (1, 'admin@noavet.jp',           '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6', true, true),
    -- 八王子院スタッフ（実名メール）
    (2, 'hayashi@noah-vet.co.jp',    '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6', true, true),  -- システム管理者（林 文明）
    (3, 'yamazaki@noah-vet.co.jp',   '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6', true, false),
    (4, 'mitsui@noah-vet.co.jp',     '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6', true, false),
    -- デモアカウント（八王子院・frontend mock-data.ts 対応）
    (5, 'admin@example.com',         '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6', true, false),
    (6, 'vet@example.com',           '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6', true, false),
    (7, 'nurse@example.com',         '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6', true, false),
    (8, 'reception@example.com',     '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6', true, false),
    (9, 'system@example.com',        '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6', true, false),
    -- 城東医院スタッフ
    (10, 'kimura@noah-vet.co.jp',    '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6', true, false),
    (11, 'sasaki@noah-vet.co.jp',    '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6', true, false),
    -- 敷島医院スタッフ
    (12, 'fujiwara@noah-vet.co.jp',  '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6', true, false),
    (13, 'matsumoto@noah-vet.co.jp', '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6', true, false),
    -- デモアカウント（frontend LoginForm の DEMO_ACCOUNTS に対応）
    (14, 'trimmer@example.com',      '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6', true, false),
    (15, 'joto-vet@example.com',     '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6', true, false),
    (16, 'shiki-vet@example.com',    '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6', true, false)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('accounts', 'id'), (SELECT MAX(id) FROM accounts));

-- 城東医院 (clinic_id=4) スタッフ
INSERT INTO staffs (id, account_id, name, is_active, license_number, occupation_id, sort_order, staff_type, reservation_visible) VALUES
    (16, NULL, '木村 健太',       true, 'V-40001', 5, 1,  'doctor',   true),
    (17, NULL, '佐々木 美香',     true, 'V-40002', 5, 2,  'doctor',   true),
    (18, NULL, '高橋 翔太',       true, 'V-40003', 5, 3,  'doctor',   true),
    (19, NULL, '田村 由紀',       true, '',        6, 4,  'nurse',    false),
    (20, NULL, '中村 大輔',       true, '',        6, 5,  'nurse',    false),
    (21, NULL, '小林 麻衣',       true, '',        6, 6,  'nurse',    false),
    (22, NULL, '井上 拓也',       true, '',        6, 7,  'nurse',    false),
    -- 城東医院 リソース
    (23, NULL, '健診・ワクチン・狂犬病', true, '', 7, 8,  'resource', true),
    (24, NULL, 'トリミング',       true, '',        7, 9,  'resource', true),
    (25, NULL, 'クイックシャンプー', true, '',      7, 10, 'resource', true)
ON CONFLICT DO NOTHING;

-- 敷島医院 (clinic_id=5) スタッフ
INSERT INTO staffs (id, account_id, name, is_active, license_number, occupation_id, sort_order, staff_type, reservation_visible) VALUES
    (26, NULL, '藤原 誠一',       true, 'V-50001', 9, 1,  'doctor',   true),
    (27, NULL, '松本 さやか',     true, 'V-50002', 9, 2,  'doctor',   true),
    (28, NULL, '石田 和也',       true, 'V-50003', 9, 3,  'doctor',   true),
    (29, NULL, '岡本 菜々子',     true, '',        10, 4, 'nurse',    false),
    (30, NULL, '西村 健二',       true, '',        10, 5, 'nurse',    false),
    -- 敷島医院 リソース
    (31, NULL, '健診・ワクチン・狂犬病', true, '', 11, 6, 'resource', true),
    (32, NULL, 'トリミング',       true, '',        11, 7, 'resource', true)
ON CONFLICT DO NOTHING;

-- デモアカウント用スタッフ（frontend LoginForm の DEMO_ACCOUNTS に対応）
-- occupation_id: 3=トリマー(八王子), 5=獣医師(城東), 9=獣医師(敷島)
INSERT INTO staffs (id, account_id, name, is_active, occupation_id, staff_type) VALUES
    (33, 14, 'さくら（デモ）',    true, 3, 'doctor'),
    (34, 15, '城東 獣医（デモ）', true, 5, 'doctor'),
    (35, 16, '敷島 獣医（デモ）', true, 9, 'doctor')
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('staffs', 'id'), (SELECT MAX(id) FROM staffs));

-- account_id → staff_id マッピング
UPDATE staffs SET account_id = 1  WHERE id = 4;  -- admin@noavet.jp → 三井 隆之 (院長代理)
UPDATE staffs SET account_id = 2  WHERE id = 1;  -- hayashi@noah-vet.co.jp → 林 文明
UPDATE staffs SET account_id = 3  WHERE id = 2;  -- yamazaki@noah-vet.co.jp → 山﨑 晶子
UPDATE staffs SET account_id = 4  WHERE id = 3;  -- mitsui@noah-vet.co.jp → 三井 隆之
UPDATE staffs SET account_id = 5  WHERE id = 8;  -- admin@example.com → 安田 希恵 (デモ用)
UPDATE staffs SET account_id = 6  WHERE id = 9;  -- vet@example.com → 倉田 春香 (デモ用)
UPDATE staffs SET account_id = 7  WHERE id = 10; -- nurse@example.com → 梶原 梨夢 (デモ用)
UPDATE staffs SET account_id = 8  WHERE id = 11; -- reception@example.com → 髙木 賀央里 (デモ用)
UPDATE staffs SET account_id = 10 WHERE id = 16; -- kimura@noah-vet.co.jp → 木村 健太
UPDATE staffs SET account_id = 11 WHERE id = 17; -- sasaki@noah-vet.co.jp → 佐々木 美香
UPDATE staffs SET account_id = 12 WHERE id = 26; -- fujiwara@noah-vet.co.jp → 藤原 誠一
UPDATE staffs SET account_id = 13 WHERE id = 27; -- matsumoto@noah-vet.co.jp → 松本 さやか
-- デモアカウント（inline account_id: staffs INSERT で設定済みだが念のため）
UPDATE staffs SET account_id = 14 WHERE id = 33; -- trimmer@example.com → さくら（デモ）
UPDATE staffs SET account_id = 15 WHERE id = 34; -- joto-vet@example.com → 城東 獣医（デモ）
UPDATE staffs SET account_id = 16 WHERE id = 35; -- shiki-vet@example.com → 敷島 獣医（デモ）

-- -----------------------------------------------------------------------------
-- 7. staff_clinic_assignments（スタッフ・クリニック割当: 37件）
-- 八王子院: staff 1-15、三井 隆之(4)は城東・敷島にもサブ割当
-- 城東医院: staff 16-25
-- 敷島医院: staff 26-32
-- -----------------------------------------------------------------------------
INSERT INTO staff_clinic_assignments (staff_id, clinic_id, is_main) VALUES
    -- 八王子院 (clinic_id=3)
    (1,  3, true),   -- 林 文明 (hayashi@noah-vet.co.jp)
    (2,  3, true),   -- 山﨑 晶子 (yamazaki@noah-vet.co.jp)
    (3,  3, true),   -- 三井 隆之 (mitsui@noah-vet.co.jp)
    (4,  3, true),   -- ノア (admin@noavet.jp)
    (5,  3, true),   -- 加藤 茉里
    (6,  3, true),   -- 金谷 亜美
    (7,  3, true),   -- 稲村 一真
    (8,  3, true),   -- 安田 希恵 (admin@example.com デモ)
    (9,  3, true),   -- 倉田 春香 (vet@example.com デモ)
    (10, 3, true),   -- 梶原 梨夢 (nurse@example.com デモ)
    (11, 3, true),   -- 髙木 賀央里 (reception@example.com デモ)
    (12, 3, true),   -- お手入れ・オゾン療法 (resource)
    (13, 3, true),   -- 健診・ワクチン・狂犬病 (resource)
    (14, 3, true),   -- ドッグラン(アジリティ解放) (resource)
    (15, 3, true),   -- クイックシャンプー (resource)
    -- システム管理者 admin@noavet.jp は全院アクセス (staff_id=4 = ノア)
    (4,  4, false),  -- 三井 隆之 — 城東医院（管理目的）
    (4,  5, false),  -- 三井 隆之 — 敷島医院（管理目的）
    -- 城東医院 (clinic_id=4)
    (16, 4, true),   -- 木村 健太 (kimura@noah-vet.co.jp)
    (17, 4, true),   -- 佐々木 美香 (sasaki@noah-vet.co.jp)
    (18, 4, true),   -- 高橋 翔太
    (19, 4, true),   -- 田村 由紀
    (20, 4, true),   -- 中村 大輔
    (21, 4, true),   -- 小林 麻衣
    (22, 4, true),   -- 井上 拓也
    (23, 4, true),   -- 健診・ワクチン・狂犬病 (resource)
    (24, 4, true),   -- トリミング (resource)
    (25, 4, true),   -- クイックシャンプー (resource)
    -- 敷島医院 (clinic_id=5)
    (26, 5, true),   -- 藤原 誠一 (fujiwara@noah-vet.co.jp)
    (27, 5, true),   -- 松本 さやか (matsumoto@noah-vet.co.jp)
    (28, 5, true),   -- 石田 和也
    (29, 5, true),   -- 岡本 菜々子
    (30, 5, true),   -- 西村 健二
    (31, 5, true),   -- 健診・ワクチン・狂犬病 (resource)
    (32, 5, true),   -- トリミング (resource)
    -- デモアカウント用割当
    (33, 3, true),   -- さくら（デモ）→ 八王子院
    (34, 4, true),   -- 城東 獣医（デモ）→ 城東医院
    (35, 5, true)    -- 敷島 獣医（デモ）→ 敷島医院
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 7b. permission_groups（権限グループ: 各医院2グループ × 3医院 = 6グループ）
-- -----------------------------------------------------------------------------
INSERT INTO permission_groups (id, clinic_id, name, description, color, is_active, sort_order) VALUES
    -- 八王子院 (clinic_id=3)
    (1, 3, '執行', '全リソースフルアクセス', '#6366F1', true, 1),
    (2, 3, '一般', '基本業務操作（医療・予約・トリミング等の作成・編集）', '#10B981', true, 2),
    -- 城東医院 (clinic_id=4)
    (3, 4, '執行', '全リソースフルアクセス', '#6366F1', true, 1),
    (4, 4, '一般', '基本業務操作（医療・予約・トリミング等の作成・編集）', '#10B981', true, 2),
    -- 敷島医院 (clinic_id=5)
    (5, 5, '執行', '全リソースフルアクセス', '#6366F1', true, 1),
    (6, 5, '一般', '基本業務操作（医療・予約・トリミング等の作成・編集）', '#10B981', true, 2)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('permission_groups', 'id'), (SELECT MAX(id) FROM permission_groups));

-- -----------------------------------------------------------------------------
-- 7c. permission_group_rules（権限グループルール: 23リソース × 2グループ = 46件）
-- V=View, C=Create, E=Edit, D=Delete
-- -----------------------------------------------------------------------------
INSERT INTO permission_group_rules (group_id, resource, can_view, can_create, can_edit, can_delete) VALUES
    -- 執行（group_id=1）: 全リソース全権限
    (1, 'reception',              true, true,  true,  true),
    (1, 'owners',                 true, true,  true,  true),
    (1, 'reservations',           true, true,  true,  true),
    (1, 'medical-records',        true, true,  true,  true),
    (1, 'hospitalization',        true, true,  true,  true),
    (1, 'trimming',               true, true,  true,  true),
    (1, 'examinations',           true, true,  true,  true),
    (1, 'accounting',             true, true,  true,  true),
    (1, 'vaccinations',           true, true,  true,  true),
    (1, 'checkups',               true, true,  true,  true),
    (1, 'inventory',              true, true,  true,  true),
    (1, 'estimates',              true, true,  true,  true),
    (1, 'shifts',                 true, true,  true,  true),
    (1, 'hospital-settings',      true, true,  true,  true),
    (1, 'master-animal-species',  true, true,  true,  true),
    (1, 'master-medical',         true, true,  true,  true),
    (1, 'master-reservation-category',    true, true,  true,  true),
    (1, 'master-hospitalization', true, true,  true,  true),
    (1, 'master-trimming',        true, true,  true,  true),
    (1, 'master-permission',      true, true,  true,  true),
    (1, 'master-staff',           true, true,  true,  true),
    (1, 'master-insurance',       true, true,  true,  true),
    (1, 'master-merchandise',     true, true,  true,  true),
    -- 一般（group_id=2）: 基本業務（マスタは閲覧のみ）
    (2, 'reception',              true, false, false, false),
    (2, 'owners',                 true, true,  true,  false),
    (2, 'reservations',           true, true,  true,  false),
    (2, 'medical-records',        true, true,  true,  false),
    (2, 'hospitalization',        true, true,  true,  false),
    (2, 'trimming',               true, true,  true,  false),
    (2, 'examinations',           true, true,  true,  false),
    (2, 'accounting',             true, false, false, false),
    (2, 'vaccinations',           true, true,  true,  false),
    (2, 'checkups',               true, false, false, false),
    (2, 'inventory',              true, false, false, false),
    (2, 'estimates',              true, false, false, false),
    (2, 'shifts',                 true, true,  true,  false),
    (2, 'hospital-settings',      true, false, false, false),
    (2, 'master-animal-species',  true, false, false, false),
    (2, 'master-medical',         true, false, false, false),
    (2, 'master-reservation-category',    true, false, false, false),
    (2, 'master-hospitalization', true, false, false, false),
    (2, 'master-trimming',        true, false, false, false),
    (2, 'master-permission',      false, false, false, false),
    (2, 'master-staff',           true, false, false, false),
    (2, 'master-insurance',       true, false, false, false),
    (2, 'master-merchandise',     true, false, false, false),
    -- 城東医院 執行（group_id=3）: 全リソース全権限
    (3, 'reception',              true, true,  true,  true),
    (3, 'owners',                 true, true,  true,  true),
    (3, 'reservations',           true, true,  true,  true),
    (3, 'medical-records',        true, true,  true,  true),
    (3, 'hospitalization',        true, true,  true,  true),
    (3, 'trimming',               true, true,  true,  true),
    (3, 'examinations',           true, true,  true,  true),
    (3, 'accounting',             true, true,  true,  true),
    (3, 'vaccinations',           true, true,  true,  true),
    (3, 'checkups',               true, true,  true,  true),
    (3, 'inventory',              true, true,  true,  true),
    (3, 'estimates',              true, true,  true,  true),
    (3, 'shifts',                 true, true,  true,  true),
    (3, 'hospital-settings',      true, true,  true,  true),
    (3, 'master-animal-species',  true, true,  true,  true),
    (3, 'master-medical',         true, true,  true,  true),
    (3, 'master-reservation-category',    true, true,  true,  true),
    (3, 'master-hospitalization', true, true,  true,  true),
    (3, 'master-trimming',        true, true,  true,  true),
    (3, 'master-permission',      true, true,  true,  true),
    (3, 'master-staff',           true, true,  true,  true),
    (3, 'master-insurance',       true, true,  true,  true),
    (3, 'master-merchandise',     true, true,  true,  true),
    -- 城東医院 一般（group_id=4）
    (4, 'reception',              true, false, false, false),
    (4, 'owners',                 true, true,  true,  false),
    (4, 'reservations',           true, true,  true,  false),
    (4, 'medical-records',        true, true,  true,  false),
    (4, 'hospitalization',        true, true,  true,  false),
    (4, 'trimming',               true, true,  true,  false),
    (4, 'examinations',           true, true,  true,  false),
    (4, 'accounting',             true, false, false, false),
    (4, 'vaccinations',           true, true,  true,  false),
    (4, 'checkups',               true, false, false, false),
    (4, 'inventory',              true, false, false, false),
    (4, 'estimates',              true, false, false, false),
    (4, 'shifts',                 true, true,  true,  false),
    (4, 'hospital-settings',      true, false, false, false),
    (4, 'master-animal-species',  true, false, false, false),
    (4, 'master-medical',         true, false, false, false),
    (4, 'master-reservation-category',    true, false, false, false),
    (4, 'master-hospitalization', true, false, false, false),
    (4, 'master-trimming',        true, false, false, false),
    (4, 'master-permission',      false, false, false, false),
    (4, 'master-staff',           true, false, false, false),
    (4, 'master-insurance',       true, false, false, false),
    (4, 'master-merchandise',     true, false, false, false),
    -- 敷島医院 執行（group_id=5）: 全リソース全権限
    (5, 'reception',              true, true,  true,  true),
    (5, 'owners',                 true, true,  true,  true),
    (5, 'reservations',           true, true,  true,  true),
    (5, 'medical-records',        true, true,  true,  true),
    (5, 'hospitalization',        true, true,  true,  true),
    (5, 'trimming',               true, true,  true,  true),
    (5, 'examinations',           true, true,  true,  true),
    (5, 'accounting',             true, true,  true,  true),
    (5, 'vaccinations',           true, true,  true,  true),
    (5, 'checkups',               true, true,  true,  true),
    (5, 'inventory',              true, true,  true,  true),
    (5, 'estimates',              true, true,  true,  true),
    (5, 'shifts',                 true, true,  true,  true),
    (5, 'hospital-settings',      true, true,  true,  true),
    (5, 'master-animal-species',  true, true,  true,  true),
    (5, 'master-medical',         true, true,  true,  true),
    (5, 'master-reservation-category',    true, true,  true,  true),
    (5, 'master-hospitalization', true, true,  true,  true),
    (5, 'master-trimming',        true, true,  true,  true),
    (5, 'master-permission',      true, true,  true,  true),
    (5, 'master-staff',           true, true,  true,  true),
    (5, 'master-insurance',       true, true,  true,  true),
    (5, 'master-merchandise',     true, true,  true,  true),
    -- 敷島医院 一般（group_id=6）
    (6, 'reception',              true, false, false, false),
    (6, 'owners',                 true, true,  true,  false),
    (6, 'reservations',           true, true,  true,  false),
    (6, 'medical-records',        true, true,  true,  false),
    (6, 'hospitalization',        true, true,  true,  false),
    (6, 'trimming',               true, true,  true,  false),
    (6, 'examinations',           true, true,  true,  false),
    (6, 'accounting',             true, false, false, false),
    (6, 'vaccinations',           true, true,  true,  false),
    (6, 'checkups',               true, false, false, false),
    (6, 'inventory',              true, false, false, false),
    (6, 'estimates',              true, false, false, false),
    (6, 'shifts',                 true, true,  true,  false),
    (6, 'hospital-settings',      true, false, false, false),
    (6, 'master-animal-species',  true, false, false, false),
    (6, 'master-medical',         true, false, false, false),
    (6, 'master-reservation-category',    true, false, false, false),
    (6, 'master-hospitalization', true, false, false, false),
    (6, 'master-trimming',        true, false, false, false),
    (6, 'master-permission',      false, false, false, false),
    (6, 'master-staff',           true, false, false, false),
    (6, 'master-insurance',       true, false, false, false),
    (6, 'master-merchandise',     true, false, false, false)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 7d. staff_permission_groups（スタッフ→権限グループ割当: 全15名）
-- 執行(1): 管理系スタッフ + デモ管理者アカウント
-- 一般(2): 一般業務スタッフ
-- -----------------------------------------------------------------------------
INSERT INTO staff_permission_groups (staff_id, group_id) VALUES
    -- 八王子院 執行グループ (group_id=1)
    (1,  1),  -- 林 文明     (hayashi@noah-vet.co.jp / システム管理者)
    (3,  1),  -- 三井 隆之   (mitsui@noah-vet.co.jp / admin@noavet.jp マッピング)
    (4,  1),  -- ノア        (admin@noavet.jp / 執行権限保持)
    -- 八王子院 一般グループ (group_id=2)
    (2,  2),  -- 山﨑 晶子   (yamazaki@noah-vet.co.jp)
    (5,  2),  -- 加藤 茉里
    (6,  2),  -- 金谷 亜美
    (7,  2),  -- 稲村 一真
    (8,  2),  -- 安田 希恵   (admin@example.com デモ)
    (9,  2),  -- 倉田 春香   (vet@example.com デモ)
    (10, 2),  -- 梶原 梨夢   (nurse@example.com デモ)
    (11, 2),  -- 髙木 賀央里 (reception@example.com デモ)
    (12, 2),  -- お手入れ・オゾン療法 (resource)
    (13, 2),  -- 健診・ワクチン・狂犬病 (resource)
    (14, 2),  -- ドッグラン (resource)
    (15, 2),  -- クイックシャンプー (resource)
    -- 城東医院 執行グループ (group_id=3)
    (16, 3),  -- 木村 健太   (kimura@noah-vet.co.jp)
    (17, 3),  -- 佐々木 美香 (sasaki@noah-vet.co.jp)
    -- 城東医院 一般グループ (group_id=4)
    (18, 4),  -- 高橋 翔太
    (19, 4),  -- 田村 由紀
    (20, 4),  -- 中村 大輔
    (21, 4),  -- 小林 麻衣
    (22, 4),  -- 井上 拓也
    (23, 4),  -- 健診・ワクチン・狂犬病 (resource)
    (24, 4),  -- トリミング (resource)
    (25, 4),  -- クイックシャンプー (resource)
    -- 敷島医院 執行グループ (group_id=5)
    (26, 5),  -- 藤原 誠一   (fujiwara@noah-vet.co.jp)
    (27, 5),  -- 松本 さやか (matsumoto@noah-vet.co.jp)
    -- 敷島医院 一般グループ (group_id=6)
    (28, 6),  -- 石田 和也
    (29, 6),  -- 岡本 菜々子
    (30, 6),  -- 西村 健二
    (31, 6),  -- 健診・ワクチン・狂犬病 (resource)
    (32, 6),  -- トリミング (resource)
    -- デモアカウント用権限グループ割当
    (33, 2),  -- さくら（デモ）→ 八王子院 一般
    (34, 3),  -- 城東 獣医（デモ）→ 城東医院 執行
    (35, 5)   -- 敷島 獣医（デモ）→ 敷島医院 執行
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 8a. reservation_category_groups（予約区分グループ）
-- 八王子院 (clinic_id=3): 7グループ (ID 1-7)
-- 城東医院 (clinic_id=4): 7グループ (ID 8-14)
-- 敷島医院 (clinic_id=5): 6グループ (ID 15-20)
-- -----------------------------------------------------------------------------

-- 八王子院 (clinic_id=3)
INSERT INTO reservation_category_groups (id, clinic_id, name, color, sort_order, is_active) VALUES
    (1,  3, '診療系',           '#3B82F6', 1, true),
    (2,  3, '予防・ワクチン',   '#10B981', 2, true),
    (3,  3, '健康診断・検査',   '#8B5CF6', 3, true),
    (4,  3, 'トリミング・美容', '#F59E0B', 4, true),
    (5,  3, '手術・処置',       '#EF4444', 5, true),
    (6,  3, 'ホテル・入院',     '#06B6D4', 6, true),
    (7,  3, '管理・内部',       '#9CA3AF', 7, true)
ON CONFLICT (id) DO UPDATE
    SET name=EXCLUDED.name, color=EXCLUDED.color, sort_order=EXCLUDED.sort_order, is_active=EXCLUDED.is_active;

-- 城東医院 (clinic_id=4)
INSERT INTO reservation_category_groups (id, clinic_id, name, color, sort_order, is_active) VALUES
    (8,  4, '診療系',           '#3B82F6', 1, true),
    (9,  4, '予防・ワクチン',   '#10B981', 2, true),
    (10, 4, '健康診断・検査',   '#8B5CF6', 3, true),
    (11, 4, 'トリミング・美容', '#F59E0B', 4, true),
    (12, 4, '手術・処置',       '#EF4444', 5, true),
    (13, 4, 'ホテル・入院',     '#06B6D4', 6, true),
    (14, 4, '管理・内部',       '#9CA3AF', 7, true)
ON CONFLICT (id) DO UPDATE
    SET name=EXCLUDED.name, color=EXCLUDED.color, sort_order=EXCLUDED.sort_order, is_active=EXCLUDED.is_active;

-- 敷島医院 (clinic_id=5)
INSERT INTO reservation_category_groups (id, clinic_id, name, color, sort_order, is_active) VALUES
    (15, 5, '診療系',           '#3B82F6', 1, true),
    (16, 5, '予防・ワクチン',   '#10B981', 2, true),
    (17, 5, '健康診断・検査',   '#8B5CF6', 3, true),
    (18, 5, 'トリミング・美容', '#F59E0B', 4, true),
    (19, 5, '手術・処置',       '#EF4444', 5, true),
    (20, 5, '管理・内部',       '#9CA3AF', 6, true)
ON CONFLICT (id) DO UPDATE
    SET name=EXCLUDED.name, color=EXCLUDED.color, sort_order=EXCLUDED.sort_order, is_active=EXCLUDED.is_active;

SELECT setval(pg_get_serial_sequence('reservation_category_groups','id'), (SELECT MAX(id) FROM reservation_category_groups));

-- -----------------------------------------------------------------------------
-- 8b. reservation_categories（予約区分）
-- 八王子院 (clinic_id=3): 25件 (ID 1-25)
-- 城東医院 (clinic_id=4): 19件 (ID 26-44)
-- 敷島医院 (clinic_id=5): 14件 (ID 45-58)
-- -----------------------------------------------------------------------------

-- 八王子院 (clinic_id=3) 公開コース (is_internal=false, reservation_visible=true)
INSERT INTO reservation_categories (id, clinic_id, name, short_name, is_active, description, color, sort_order, duration_minutes, reservation_visible, reservation_comment, is_internal) VALUES
    (1,  3, '一般診察',               '診察',     true, '内科・外科・皮膚科などの一般的な診察',         '#3B82F6', 1,  15, true,  '', false),
    (2,  3, '一般診察(再診)',          '再診',     true, '継続通院の一般診察',                           '#3B82F6', 2,  15, true,  '', false),
    (3,  3, 'ワクチン接種',            'ワクチン', true, '各種ワクチン接種（予防接種）',                 '#10B981', 3,  15, true,  '', false),
    (4,  3, 'お手入れ',               'お手入れ', true, '爪切り・耳掃除・肛門腺絞りなど',               '#F59E0B', 4,  15, true,  '', false),
    (5,  3, '狂犬病',                 '狂犬病',   true, '狂犬病予防法に基づくワクチン接種',             '#10B981', 5,  15, true,  '', false),
    (6,  3, 'フィラリア予防',          'フィラリア', true, 'フィラリア予防薬投与・処方',                '#10B981', 6,  15, true,  '', false),
    (7,  3, '健康診断',               '健診',     true, '定期健康診断・フィラリア検査など',             '#8B5CF6', 7,  15, true,  '', false),
    (8,  3, '健康診断結果報告',        '結果報告', true, '健康診断結果の説明・報告',                     '#8B5CF6', 8,  15, true,  '', false),
    (9,  3, 'トリミングコース',        'トリミング', true, 'カット・シャンプー・ブロー・爪切り・耳掃除', '#F59E0B', 9,  15, true,  '', false),
    (10, 3, 'トリミング部分カットコース', '部分カット', true, '部分的なカット・トリミング',              '#F59E0B', 10, 15, true,  '', false),
    (11, 3, 'トリミングシャンプーコース', 'シャンプー', true, 'シャンプー・ブロー・ブラッシング',        '#F59E0B', 11, 15, true,  '', false),
    (12, 3, 'クイックシャンプー',      'Qシャンプー', true, '短時間シャンプー',                        '#F59E0B', 12, 15, true,  '', false),
    (13, 3, '室内ドッグラン',          'ドッグラン', true, '室内ドッグラン利用（60分）',                '#6B7280', 13, 60, true,  '', false)
ON CONFLICT DO NOTHING;

-- 八王子院 (clinic_id=3) スタッフ専用コース (is_internal=true, reservation_visible=false)
INSERT INTO reservation_categories (id, clinic_id, name, short_name, is_active, description, color, sort_order, duration_minutes, reservation_visible, reservation_comment, is_internal) VALUES
    (14, 3, '手術60',                 '手術60',   true, '手術枠（60分）',                               '#EF4444', 14, 60, false, '', true),
    (15, 3, 'ホテルお迎え',           'お迎え',   true, 'ホテルお迎え対応',                             '#6B7280', 15, 15, false, '', true),
    (16, 3, 'ホテル預かり',           '預かり',   true, 'ペットホテル預かり',                           '#6B7280', 16, 15, false, '', true),
    (17, 3, '面会',                   '面会',     true, '入院動物の面会対応',                           '#6B7280', 17, 15, false, '', true),
    (18, 3, '☎︎',                   '☎︎',      true, '電話対応枠',                                   '#6B7280', 18, 15, false, '', true),
    (19, 3, '手術30',                 '手術30',   true, '手術枠（30分）',                               '#EF4444', 19, 30, false, '', true),
    (20, 3, '休憩枠',                 '休憩',     true, '休憩・ブロック枠',                             '#6B7280', 20, 60, false, '', true),
    (21, 3, '×',                     '×',       true, '予約不可（15分）',                             '#6B7280', 21, 15, false, '', true),
    (22, 3, '電話予約60',             '電話予約', true, '電話予約枠（60分）',                           '#6B7280', 22, 60, false, '', true),
    (23, 3, '予約不可60',             '不可60',   true, '予約不可ブロック（60分）',                     '#6B7280', 23, 60, false, '', true),
    (24, 3, 'エコー枠',               'エコー',   true, '超音波検査専用枠',                             '#8B5CF6', 24, 30, false, '', true),
    (25, 3, '予約不可30',             '不可30',   true, '予約不可ブロック（30分）',                     '#6B7280', 25, 30, false, '', true)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('reservation_categories', 'id'), (SELECT MAX(id) FROM reservation_categories));

-- 八王子院 (clinic_id=3) グループ紐付け
UPDATE reservation_categories SET group_id=1 WHERE clinic_id=3 AND id IN (1,2);
UPDATE reservation_categories SET group_id=2 WHERE clinic_id=3 AND id IN (3,5,6);
UPDATE reservation_categories SET group_id=3 WHERE clinic_id=3 AND id IN (7,8,24);
UPDATE reservation_categories SET group_id=4 WHERE clinic_id=3 AND id IN (4,9,10,11,12);
UPDATE reservation_categories SET group_id=5 WHERE clinic_id=3 AND id IN (14,19);
UPDATE reservation_categories SET group_id=6 WHERE clinic_id=3 AND id IN (13,15,16,17);
UPDATE reservation_categories SET group_id=7, is_internal=true WHERE clinic_id=3 AND id IN (18,20,21,22,23,25);

-- -----------------------------------------------------------------------------
-- 9. cages（ケージ: 8件）
-- -----------------------------------------------------------------------------
INSERT INTO cages (id, clinic_id, name, price, is_active, description, cage_type, cage_size, sort_order) VALUES
    (1, 3, 'ICUケージA',     8000, true, '酸素吸入可・重症患者用',  'icu',     'medium', 1),
    (2, 3, 'ICUケージB',     8000, true, '酸素吸入可・重症患者用',  'icu',     'medium', 2),
    (3, 3, '犬用ケージ（小）', 3000, true, '小型犬・ホテル利用可',    'dog',     'small',  3),
    (4, 3, '犬用ケージ（中）', 3500, true, '中型犬・一般入院用',      'dog',     'medium', 4),
    (5, 3, '犬用ケージ（大）', 4000, true, '大型犬・術後管理用',      'dog',     'large',  5),
    (6, 3, '猫用ケージ（小）', 3000, true, '猫専用・ストレス軽減設計', 'cat',     'small',  6),
    (7, 3, '猫用ケージ（中）', 3000, true, '猫専用・ストレス軽減設計', 'cat',     'medium', 7),
    (8, 3, '汎用ケージA',     2500, true, '小動物・鳥類等対応',      'general', 'small',  8)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('cages', 'id'), (SELECT MAX(id) FROM cages));

-- -----------------------------------------------------------------------------
-- 10. insurances（保険: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO insurances (id, clinic_id, name, is_active, description, coverage_rate, contact_phone, sort_order) VALUES
    (1, 3, 'アニコム損保',         true, 'ペット保険大手・どうぶつ健保シリーズ', 70, '0120-025-034',  1),
    (2, 3, 'アイペット損保',       true, 'うちの子シリーズ',                     70, '0120-956-099',  2),
    (3, 3, 'ペット&ファミリー',     true, 'げんきナンバーワンシリーズ',           80, '0120-81-8505',  3),
    (4, 3, '楽天ペット保険',       true, '楽天が提供するペット保険',             70, '0120-600-810',  4),
    (5, 3, 'その他（自費）',       true, '保険未加入・全額自費',                100, '',              5)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('insurances', 'id'), (SELECT MAX(id) FROM insurances));

-- -----------------------------------------------------------------------------
-- 11. exam_types（検査種別: 5件）+ exam_type_fields（検査項目定義）
-- -----------------------------------------------------------------------------
INSERT INTO exam_types (id, clinic_id, name, price, is_active, description, sort_order) VALUES
    (1, 3, '血液検査（CBC）',     3000, true, '全血球計算（Complete Blood Count）',         1),
    (2, 3, '血液化学検査',         5000, true, '肝機能・腎機能・血糖値など生化学的検査',     2),
    (3, 3, '尿検査',               1500, true, '尿試験紙・尿沈渣検査',                       3),
    (4, 3, 'レントゲン検査',       3000, true, 'X線撮影（胸部・腹部・四肢）',                4),
    (5, 3, '超音波検査（エコー）', 5000, true, '腹部エコー・心臓エコー',                     5)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('exam_types', 'id'), (SELECT MAX(id) FROM exam_types));

-- exam_type_fields: 血液検査（CBC）
INSERT INTO exam_type_fields (id, exam_type_id, name, inspection_value, normal_value, sort_order) VALUES
    (1, 1, 'WBC（白血球数）',      '', '6.0-17.0 x10^3/uL', 1),
    (2, 1, 'RBC（赤血球数）',      '', '5.5-8.5 x10^6/uL',  2),
    (3, 1, 'HCT（ヘマトクリット）', '', '37-55%',            3),
    (4, 1, 'PLT（血小板数）',      '', '175-500 x10^3/uL',  4)
ON CONFLICT DO NOTHING;

-- exam_type_fields: 血液化学検査
INSERT INTO exam_type_fields (id, exam_type_id, name, inspection_value, normal_value, sort_order) VALUES
    (5, 2, 'ALT（GPT）',        '', '10-125 U/L',    1),
    (6, 2, 'BUN（尿素窒素）',   '', '7-27 mg/dL',    2),
    (7, 2, 'CRE（クレアチニン）', '', '0.5-1.8 mg/dL', 3),
    (8, 2, 'GLU（血糖値）',     '', '74-143 mg/dL',   4)
ON CONFLICT DO NOTHING;

-- exam_type_fields: 尿検査
INSERT INTO exam_type_fields (id, exam_type_id, name, inspection_value, normal_value, sort_order) VALUES
    (9,  3, '尿比重',   '', '1.015-1.045', 1),
    (10, 3, '尿pH',     '', '5.5-7.5',     2),
    (11, 3, '尿タンパク', '', '陰性',       3),
    (12, 3, '尿潜血',   '', '陰性',        4)
ON CONFLICT DO NOTHING;

-- exam_type_fields: レントゲン検査
INSERT INTO exam_type_fields (id, exam_type_id, name, inspection_value, normal_value, sort_order) VALUES
    (13, 4, '胸部正面', '', '異常なし', 1),
    (14, 4, '腹部正面', '', '異常なし', 2),
    (15, 4, '四肢',     '', '異常なし', 3)
ON CONFLICT DO NOTHING;

-- exam_type_fields: 超音波検査
INSERT INTO exam_type_fields (id, exam_type_id, name, inspection_value, normal_value, sort_order) VALUES
    (16, 5, '腹部エコー', '', '異常なし', 1),
    (17, 5, '心臓エコー', '', '異常なし', 2)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('exam_type_fields', 'id'), (SELECT MAX(id) FROM exam_type_fields));

-- -----------------------------------------------------------------------------
-- 12. vaccines（ワクチン: 10件）
-- -----------------------------------------------------------------------------
INSERT INTO vaccines (id, clinic_id, name, price, is_active, description, species, interval, sort_order) VALUES
    (1,  3, '混合ワクチン5種（犬）',      4500, true, 'ジステンパー・パルボ・アデノ1型・アデノ2型・パラインフルエンザ',         'dog', '1年',   1),
    (2,  3, '混合ワクチン6種（犬）',      5500, true, '5種＋コロナウイルス',                                                    'dog', '1年',   2),
    (3,  3, '混合ワクチン8種（犬）',      6500, true, '5種＋レプトスピラ3種',                                                   'dog', '1年',   3),
    (4,  3, '混合ワクチン10種（犬）',     8000, true, '5種＋レプトスピラ5種',                                                   'dog', '1年',   4),
    (5,  3, '混合ワクチン3種（猫）',      4000, true, '猫ウイルス性鼻気管炎・カリシウイルス・汎白血球減少症',                     'cat', '1年',   5),
    (6,  3, '混合ワクチン5種（猫）',      5500, true, '3種＋猫白血病・猫クラミジア',                                             'cat', '1年',   6),
    (7,  3, '狂犬病ワクチン',             3000, true, '狂犬病予防法に基づく接種',                                               'dog', '1年',   7),
    (8,  3, 'フィラリア予防薬（小型犬）',  900, true, '体重10kg以下犬用フィラリア予防',                                          'dog', '1ヶ月', 8),
    (9,  3, 'フィラリア予防薬（中型犬）', 1100, true, '体重11-25kg犬用フィラリア予防',                                           'dog', '1ヶ月', 9),
    (10, 3, 'フィラリア予防薬（大型犬）', 1500, true, '体重26kg以上犬用フィラリア予防',                                          'dog', '1ヶ月', 10)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('vaccines', 'id'), (SELECT MAX(id) FROM vaccines));

-- -----------------------------------------------------------------------------
-- 13. medicines（薬剤カテゴリ: 9件 + 薬剤: 15件）
-- カテゴリレコードは id 1001〜1009、price=NULL、parent_id=NULL
-- 薬剤レコードは parent_id でカテゴリを参照
-- -----------------------------------------------------------------------------

-- カテゴリレコード（id 1001〜1009）
INSERT INTO medicines (id, clinic_id, name, price, is_active, description, sort_order) VALUES
    (1001, 3, '抗生剤',       NULL, true, '抗生物質カテゴリ',   1),
    (1002, 3, 'ステロイド',   NULL, true, 'ステロイド剤カテゴリ', 2),
    (1003, 3, '利尿剤',       NULL, true, '利尿剤カテゴリ',     3),
    (1004, 3, '消炎剤',       NULL, true, '消炎鎮痛剤カテゴリ', 4),
    (1005, 3, '神経系薬',     NULL, true, '神経系薬カテゴリ',   5),
    (1006, 3, '制吐剤',       NULL, true, '制吐剤カテゴリ',     6),
    (1007, 3, '消化器用薬',   NULL, true, '消化器用薬カテゴリ', 7),
    (1008, 3, '駆虫剤',       NULL, true, '駆虫剤カテゴリ',     8),
    (1009, 3, '輸液',         NULL, true, '輸液カテゴリ',       9)
ON CONFLICT DO NOTHING;

-- 薬剤レコード（parent_id でカテゴリ参照）
INSERT INTO medicines (id, clinic_id, name, price, is_active, description, dosage_form, medicine_unit, default_quantity, sort_order, parent_id) VALUES
    (1,  3, 'アモキシシリン 50mg',         500,  true, '広域スペクトラム抗生物質',               'tablet',    'per_tablet', 1,   1, 1001),
    (2,  3, 'メトロニダゾール 250mg',       600,  true, '嫌気性菌・原虫感染症治療薬',             'tablet',    'per_tablet', 1,   2, 1001),
    (3,  3, 'プレドニゾロン 5mg',           400,  true, 'ステロイド系抗炎症・免疫抑制剤',         'tablet',    'per_tablet', 1,   3, 1002),
    (4,  3, 'フロセミド注射液 20mg/2ml',    800,  true, '利尿剤（心臓・腎臓病の浮腫治療）',       'injection', 'per_ml',     2,   4, 1003),
    (5,  3, 'メロキシカム経口液',           700,  true, 'NSAIDs・痛み・炎症の緩和',               'liquid',    'per_ml',     1,   5, 1004),
    (6,  3, 'ガバペンチン 100mg',           550,  true, '神経因性疼痛・てんかん補助療法',         'tablet',    'per_tablet', 1,   6, 1005),
    (7,  3, 'マロピタント 16mg',            800,  true, '制吐剤（乗り物酔い・嘔吐治療）',         'tablet',    'per_tablet', 1,   7, 1006),
    (8,  3, 'ラクツロース液',               500,  true, '便秘・肝性脳症の治療',                   'liquid',    'per_ml',     5,   8, 1007),
    (9,  3, 'ノミ・ダニ駆除薬（犬用）',     2500, true, '外部寄生虫予防・駆除（スポットオン）',   'topical',   'per_dose',   1,   9, 1008),
    (10, 3, 'ノミ・ダニ駆除薬（猫用）',     2500, true, '外部寄生虫予防・駆除（スポットオン）',   'topical',   'per_dose',   1,  10, 1008),
    (11, 3, '抗生剤点眼薬',                 600,  true, '眼科用抗菌点眼剤',                       'liquid',    'per_ml',     1,  11, 1001),
    (12, 3, 'デキサメタゾン注射液',         700,  true, '強力ステロイド・アレルギー緊急治療',     'injection', 'per_ml',     1,  12, 1002),
    (13, 3, '生理食塩水 500ml',             400,  true, '点滴・洗浄用生理食塩水',                 'liquid',    'per_ml',     500, 13, 1009),
    (14, 3, 'セファレキシン 250mg',         450,  true, '第1世代セフェム系抗生物質',              'tablet',    'per_tablet', 1,  14, 1001),
    (15, 3, 'オメプラゾール 10mg',          350,  true, 'プロトンポンプ阻害薬（胃酸抑制）',       'tablet',    'per_tablet', 1,  15, 1007)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('medicines', 'id'), (SELECT MAX(id) FROM medicines));

-- -----------------------------------------------------------------------------
-- 14. consultations（診察項目: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO consultations (id, clinic_id, name, price, is_active, description, time_condition, duration, sort_order) VALUES
    (1, 3, '初診料',       2000, true, '初めての受診または6ヶ月以上受診がない場合', 'first_visit',  30, 1),
    (2, 3, '再診料',        800, true, '継続通院の診察料',                         'revisit',      15, 2),
    (3, 3, '往診料',       5000, true, '自宅への往診料（基本料金）',               'anytime',      60, 3),
    (4, 3, '時間外診療料', 3000, true, '診療時間外・休日の緊急診察',               'after_hours',  30, 4),
    (5, 3, '電話相談料',    500, true, '電話による診察相談',                       'anytime',      15, 5)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('consultations', 'id'), (SELECT MAX(id) FROM consultations));

-- -----------------------------------------------------------------------------
-- 15. procedures（処置項目: 10件）
-- -----------------------------------------------------------------------------
INSERT INTO procedures (id, clinic_id, name, price, is_active, description, duration, anesthesia, sort_order) VALUES
    (1,  3, '去勢手術（犬）',   25000, true, '雄犬の去勢手術',                   60,  'general', 1),
    (2,  3, '避妊手術（猫）',   25000, true, '雌猫の避妊手術',                   90,  'general', 2),
    (3,  3, '歯石除去',         15000, true, '全身麻酔下での歯石除去・歯周治療', 45,  'general', 3),
    (4,  3, '耳洗浄',            2500, true, '外耳炎治療・耳道内の洗浄処置',     15,  'none',    4),
    (5,  3, '爪切り',             500, true, '爪のカット・やすりがけ',           10,  'none',    5),
    (6,  3, '皮膚縫合',          5000, true, '裂傷・切傷の縫合処置',             30,  'local',   6),
    (7,  3, '骨折整復',         80000, true, '骨折の外科的整復・固定',          120,  'general', 7),
    (8,  3, '腫瘍摘出',         20000, true, '皮膚腫瘍の外科的摘出',             60,  'local',   8),
    (9,  3, '胃洗浄',           10000, true, '異物誤飲時の胃洗浄処置',           30,  'general', 9),
    (10, 3, '点滴処置',          3000, true, '静脈内点滴（1時間）',              60,  'none',   10)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('procedures', 'id'), (SELECT MAX(id) FROM procedures));

-- -----------------------------------------------------------------------------
-- 16. hospitalization_plans（入院プラン: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO hospitalization_plans (id, clinic_id, name, price, is_active, description, body_size, billing_unit, sort_order) VALUES
    (1, 3, '一般入院（小型）', 3000, true, '体重10kg以下の入院管理料（1日）',  'small',  'per_day',   1),
    (2, 3, '一般入院（中型）', 3500, true, '体重10-25kgの入院管理料（1日）',   'medium', 'per_day',   2),
    (3, 3, '一般入院（大型）', 4500, true, '体重25kg以上の入院管理料（1日）',  'large',  'per_day',   3),
    (4, 3, 'ICU入院',          8000, true, '集中治療室管理料（1日）',          'small',  'per_day',   4),
    (5, 3, 'ホテル（小型）',   2500, true, '体重10kg以下のペットホテル（1泊）', 'small',  'per_night', 5)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('hospitalization_plans', 'id'), (SELECT MAX(id) FROM hospitalization_plans));

-- -----------------------------------------------------------------------------
-- 17. trimming_courses（トリミングコース: 5件）
-- ※ duration は integer (分)
-- -----------------------------------------------------------------------------
INSERT INTO trimming_courses (id, clinic_id, name, price, is_active, description, target_size, duration, sort_order) VALUES
    (1, 3, 'シャンプー&ブロー（小型）', 4000,  true, 'シャンプー・ブロー・ブラッシング',            'small',  60,  1),
    (2, 3, 'シャンプー&ブロー（中型）', 5500,  true, 'シャンプー・ブロー・ブラッシング',            'medium', 90,  2),
    (3, 3, 'フルコース（小型）',        7000,  true, 'カット・シャンプー・ブロー・爪切り・耳掃除', 'small',  120, 3),
    (4, 3, 'フルコース（中型）',        9000,  true, 'カット・シャンプー・ブロー・爪切り・耳掃除', 'medium', 150, 4),
    (5, 3, 'フルコース（大型）',        12000, true, 'カット・シャンプー・ブロー・爪切り・耳掃除', 'large',  180, 5)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('trimming_courses', 'id'), (SELECT MAX(id) FROM trimming_courses));

-- -----------------------------------------------------------------------------
-- 18. trimming_options（トリミングオプション: 5件）
-- ※ is_combinable は boolean, duration は integer（分単位）
-- -----------------------------------------------------------------------------
INSERT INTO trimming_options (id, clinic_id, name, price, is_active, description, duration, is_combinable, sort_order) VALUES
    (1, 3, '爪切り',     300, true, '爪のカット・やすりがけ',       10, true, 1),
    (2, 3, '耳掃除',     500, true, '外耳道の洗浄・清掃',           10, true, 2),
    (3, 3, '歯磨き',     500, true, '歯ブラシによるデンタルケア',   15, true, 3),
    (4, 3, '肛門腺絞り', 300, true, '肛門嚢の分泌液除去',            5, true, 4),
    (5, 3, 'リボン装着', 200, true, '仕上げのアクセサリー装着',      5, true, 5)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('trimming_options', 'id'), (SELECT MAX(id) FROM trimming_options));

-- -----------------------------------------------------------------------------
-- 19. diagnosis_types（診断カテゴリ: 8件）
-- -----------------------------------------------------------------------------
INSERT INTO diagnosis_types (id, clinic_id, name, is_active, description, sort_order) VALUES
    (1, 3, '消化器系',       true, '胃腸・肝臓・膵臓などの消化器系疾患',   1),
    (2, 3, '呼吸器系',       true, '肺・気管・鼻腔などの呼吸器系疾患',     2),
    (3, 3, '皮膚・被毛',     true, 'アレルギー・感染症などの皮膚疾患',     3),
    (4, 3, '泌尿器系',       true, '腎臓・膀胱・尿道などの泌尿器系疾患',   4),
    (5, 3, '神経系',         true, '脳・脊髄・末梢神経などの神経系疾患',   5),
    (6, 3, '感染症・寄生虫', true, '細菌・ウイルス・寄生虫感染症',         6),
    (7, 3, '腫瘍',           true, '良性・悪性腫瘍（がん）',               7),
    (8, 3, '外傷・骨格',     true, '骨折・咬傷・関節疾患など',             8)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('diagnosis_types', 'id'), (SELECT MAX(id) FROM diagnosis_types));

-- -----------------------------------------------------------------------------
-- 20. diagnosis_names（診断名: 各カテゴリ2-3件、計20件）
-- -----------------------------------------------------------------------------
INSERT INTO diagnosis_names (id, clinic_id, name, is_active, description, diagnosis_category_id, sort_order) VALUES
    -- 消化器系
    (1,  3, '胃腸炎',             true, '胃・腸の炎症（嘔吐・下痢）',         1, 1),
    (2,  3, '膵炎',               true, '膵臓の炎症',                         1, 2),
    (3,  3, '肝疾患',             true, '肝炎・肝不全・脂肪肝など',           1, 3),
    -- 呼吸器系
    (4,  3, '気管支炎',           true, '気管支の炎症',                       2, 1),
    (5,  3, '肺炎',               true, '肺の感染性・非感染性炎症',           2, 2),
    -- 皮膚・被毛
    (6,  3, 'アトピー性皮膚炎',   true, 'アレルゲンによるアレルギー性皮膚炎', 3, 1),
    (7,  3, '膿皮症',             true, '細菌性の皮膚感染症',                 3, 2),
    (8,  3, '真菌症',             true, '皮膚糸状菌による感染症',             3, 3),
    -- 泌尿器系
    (9,  3, '膀胱炎',             true, '細菌性・特発性膀胱炎',               4, 1),
    (10, 3, '腎不全',             true, '急性・慢性腎不全',                   4, 2),
    (11, 3, '尿路結石',           true, '腎結石・膀胱結石・尿道結石',         4, 3),
    -- 神経系
    (12, 3, 'てんかん',           true, '反復性の痙攣発作',                   5, 1),
    (13, 3, '椎間板ヘルニア',     true, '頸椎・腰椎の椎間板突出',             5, 2),
    -- 感染症・寄生虫
    (14, 3, 'パルボウイルス感染症', true, '犬パルボウイルスによる感染症',       6, 1),
    (15, 3, 'フィラリア症',       true, '犬糸状虫による心肺疾患',             6, 2),
    (16, 3, '猫風邪（FVR）',      true, '猫ウイルス性鼻気管炎',               6, 3),
    -- 腫瘍
    (17, 3, '肥満細胞腫',         true, '皮膚または内臓の肥満細胞腫瘍',       7, 1),
    (18, 3, 'リンパ腫',           true, '悪性リンパ腫',                       7, 2),
    -- 外傷・骨格
    (19, 3, '骨折',               true, '各部位の骨折',                       8, 1),
    (20, 3, '咬傷',               true, '他動物による咬傷・咬傷感染',         8, 2),
    -- 皮膚・被毛（追加）
    (41, 3, '外耳炎',             true, '外耳道の炎症・細菌/真菌/寄生虫感染', 3, 4),
    (42, 3, '膝蓋骨脱臼',         true, '膝蓋骨の内方/外方脱臼',             8, 3)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('diagnosis_names', 'id'), (SELECT MAX(id) FROM diagnosis_names));

-- -----------------------------------------------------------------------------
-- 21. checkup_types（健診種別: 4件）
-- -----------------------------------------------------------------------------
INSERT INTO checkup_types (id, clinic_id, name, price, is_active, description, interval, target_age, sort_order) VALUES
    (1, 3, '一般健診',       5000,  true, '身体検査・体重測定・問診',                     '1年',   '全年齢', 1),
    (2, 3, '老齢検診',       15000, true, '身体検査＋血液検査＋レントゲン＋超音波',         '6ヶ月', '7歳以上', 2),
    (3, 3, 'フィラリア検査', 2500,  true, 'フィラリア抗原検査（予防シーズン前）',           '1年',   '成犬',   3),
    (4, 3, '歯科検診',       3000,  true, '歯周病チェック・歯石付着度の確認',             '1年',   '成犬',   4)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('checkup_types', 'id'), (SELECT MAX(id) FROM checkup_types));

-- -----------------------------------------------------------------------------
-- 22. chief_complaint_types（主訴区分: 6件）
-- -----------------------------------------------------------------------------
INSERT INTO chief_complaint_types (id, clinic_id, name, is_active, sort_order) VALUES
    (1, 3, '食欲不振',       true, 1),
    (2, 3, '嘔吐・下痢',     true, 2),
    (3, 3, '皮膚・被毛異常', true, 3),
    (4, 3, '呼吸困難',       true, 4),
    (5, 3, '排尿・排泄異常', true, 5),
    (6, 3, '外傷・骨折',     true, 6)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('chief_complaint_types', 'id'), (SELECT MAX(id) FROM chief_complaint_types));

-- -----------------------------------------------------------------------------
-- 23. inquiry_templates（問診定型文: 10件）
-- ※ category は text 型: chief_complaint / history / current_medications / notes
-- -----------------------------------------------------------------------------
INSERT INTO inquiry_templates (id, clinic_id, category, title, content, is_active, sort_order) VALUES
    (1,  3, 'chief_complaint',    '食欲不振（急性）',           'いつ頃から食欲が落ちましたか？完全に食べないのか、減っているだけか確認してください。', true, 1),
    (2,  3, 'chief_complaint',    '嘔吐（回数・内容物）',       '嘔吐の回数、内容物（食物・胆汁・血液など）、嘔吐のタイミング（食後すぐ・空腹時）を確認してください。', true, 2),
    (3,  3, 'chief_complaint',    '下痢（性状・頻度）',         '便の性状（軟便・水様便・血便・粘液便）、排便頻度、いつから続いているか確認してください。', true, 3),
    (4,  3, 'chief_complaint',    '皮膚の痒み・発赤',           '痒がる部位、発症時期、季節性の有無、ノミ・ダニ予防の状況を確認してください。', true, 4),
    (5,  3, 'chief_complaint',    '排尿異常',                   '排尿回数の変化、尿の色・量、排尿時の痛みの有無、血尿の有無を確認してください。', true, 5),
    (6,  3, 'history',            '既往歴確認（手術歴）',       '過去の手術歴（去勢・避妊含む）、入院歴、重大な疾患の既往を確認してください。', true, 6),
    (7,  3, 'history',            '予防接種歴確認',             '最終ワクチン接種日、狂犬病予防接種の有無、フィラリア予防の状況を確認してください。', true, 7),
    (8,  3, 'current_medications', '現在の投薬状況',            '現在服用中の薬剤名、用量、投与期間、処方元の病院を確認してください。', true, 8),
    (9,  3, 'current_medications', 'サプリメント・フード',      '現在与えているサプリメント、療法食、おやつの種類を確認してください。', true, 9),
    (10, 3, 'notes',              '生活環境確認',               '室内飼い/外飼い、同居動物の有無、散歩の頻度・時間を確認してください。', true, 10)
ON CONFLICT DO NOTHING;

-- 城東医院 (clinic_id=4)
INSERT INTO inquiry_templates (id, clinic_id, category, title, content, is_active, sort_order) VALUES
    (11, 4, 'chief_complaint',    '食欲不振',                   'いつ頃から食欲が低下しましたか？完全絶食か減少かを確認してください。', true, 1),
    (12, 4, 'chief_complaint',    '嘔吐',                       '嘔吐の回数・内容物・タイミングを確認してください。', true, 2),
    (13, 4, 'chief_complaint',    '下痢',                       '便の性状・排便頻度・持続期間を確認してください。', true, 3),
    (14, 4, 'history',            '既往歴・手術歴',             '過去の手術歴・入院歴・重大な疾患の既往を確認してください。', true, 4),
    (15, 4, 'history',            '予防接種歴',                 '最終ワクチン接種日・狂犬病予防・フィラリア予防の状況を確認してください。', true, 5),
    (16, 4, 'current_medications', '投薬状況',                  '現在服用中の薬剤名・用量・処方元を確認してください。', true, 6),
    (17, 4, 'notes',              '生活環境',                   '室内飼い/外飼い・同居動物・散歩頻度を確認してください。', true, 7),
-- 敷島医院 (clinic_id=5)
    (18, 5, 'chief_complaint',    '食欲低下',                   'いつから食べなくなったか、食事内容の変化はあるか確認してください。', true, 1),
    (19, 5, 'chief_complaint',    '嘔吐・吐出',                 '嘔吐の頻度・内容物・食事との関連を確認してください。', true, 2),
    (20, 5, 'chief_complaint',    '消化器症状（下痢）',         '便の状態・頻度・血便の有無を確認してください。', true, 3),
    (21, 5, 'history',            '既往歴確認',                 '手術歴・入院歴・アレルギー歴を確認してください。', true, 4),
    (22, 5, 'history',            'ワクチン・予防歴',           'ワクチン接種歴・フィラリア予防の有無を確認してください。', true, 5),
    (23, 5, 'current_medications', '現在の治療状況',            '他院での治療中の薬・サプリメントを確認してください。', true, 6),
    (24, 5, 'notes',              '飼育環境',                   '飼育場所・同居動物・食事内容を確認してください。', true, 7)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('inquiry_templates', 'id'), (SELECT MAX(id) FROM inquiry_templates));

-- -----------------------------------------------------------------------------
-- 24. inventory_items（在庫管理: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO inventory_items (id, clinic_id, name, category, quantity, unit, min_stock_level, location, supplier, status) VALUES
    (1, 3, 'フロントライン プラス（犬用）',     'medicine',   50, '本',    10, '薬品棚A',   'メリアルジャパン',         'sufficient'),
    (2, 3, 'ネクスガード チュアブル',            'medicine',   30, '錠',    10, '薬品棚A',   'ベーリンガーインゲルハイム', 'sufficient'),
    (3, 3, 'ヒルズ i/d（犬用・消化器ケア）',     'food',       20, '袋',     5, '食品棚B',   'ヒルズ・コルゲート',       'sufficient'),
    (4, 3, 'ロイヤルカナン 消化器サポート',      'food',       15, '袋',     5, '食品棚B',   'ロイヤルカナン',           'sufficient'),
    (5, 3, '包帯・ガーゼセット',                 'consumable', 100, 'セット', 20, '消耗品棚C', '白十字',               'sufficient')
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('inventory_items', 'id'), (SELECT MAX(id) FROM inventory_items));

-- -----------------------------------------------------------------------------
-- 25. merchandise_items（物販・フード・その他: 7件）
-- -----------------------------------------------------------------------------
INSERT INTO merchandise_items (id, clinic_id, name, category, unit_price, tax_rate, sort_order) VALUES
    (1, 3, 'ロイヤルカナン 消化器サポート 1kg', 'food', 2800, 0.10, 1),
    (2, 3, 'ヒルズ k/d 2kg', 'food', 3500, 0.10, 2),
    (3, 3, 'ペット用歯ブラシセット', 'goods', 1200, 0.10, 3),
    (4, 3, 'エリザベスカラー（S）', 'goods', 800, 0.10, 4),
    (5, 3, 'ノミ・ダニ予防首輪', 'goods', 1500, 0.10, 5),
    (6, 3, '文書料', 'other', 3000, 0.10, 6),
    (7, 3, '時間外診療費', 'other', 5000, 0.10, 7)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('merchandise_items', 'id'), (SELECT MAX(id) FROM merchandise_items));

-- =============================================================================
-- マスタ設定完了
-- =============================================================================


-- =============================================================================
-- デモデータ投入（飼主・ペット一覧ページ対応）
-- 内容: 飼主・ペット・取引記録（カルテ・予約・会計・入院・在庫・監査ログ等）
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 1. owners（飼主: 22件）
-- -----------------------------------------------------------------------------
INSERT INTO owners (id, clinic_id, name, name_kana, birth_date, company, postal_code, address1, address2, phone, company_phone, email, remarks, is_dangerous, discount_rate, membership_type) VALUES
    (1,  3, '林 文明', 'ハヤシ フミアキ', '1980-05-15', 'サンプル株式会社', '150-0001', '東京都渋谷区神宮前1-2-3', '', '090-1111-2222', '03-1234-5678', 'hayashi@example.com', '定期検診を希望', false, 10, 'member'),
    (2,  3, '田中 花子', 'タナカ ハナコ', '1985-03-20', '', '160-0022', '東京都新宿区新宿1-1-1', '', '080-3333-4444', '', 'tanaka@example.com', '', false, 0, 'non_member'),
    (3,  3, '鈴木 一郎', 'スズキ イチロウ', '1978-11-03', '', '170-0001', '東京都豊島区西巣鴨1-3-5', '', '070-5555-6666', '', 'suzuki@example.com', '', false, 0, 'member'),
    (4,  3, '田中 美咲', 'タナカ ミサキ', '1990-07-22', '', '153-0044', '東京都目黒区大橋2-4-6', '', '090-9999-8888', '', 'misaki.tanaka@example.com', '', false, 0, 'non_member'),
    (5,  3, '佐藤 花子', 'サトウ ハナコ', '1975-02-14', '', '140-0001', '東京都品川区北品川3-5-7', '', '080-2222-3333', '', 'hanako.sato@example.com', '', false, 5, 'member'),
    (6,  3, '伊藤 次郎', 'イトウ ジロウ', '1983-09-30', '', '166-0013', '東京都杉並区堀ノ内1-7-9', '', '090-1234-5678', '', 'jiro.ito@example.com', '', false, 0, 'non_member'),
    (7,  3, '小林 さくら', 'コバヤシ サクラ', '1992-04-05', '', '176-0012', '東京都練馬区豊玉北4-2-8', '', '080-9876-5432', '', 'sakura.kobayashi@example.com', '', false, 0, 'member'),
    (8,  3, '中村 勇気', 'ナカムラ ユウキ', '1987-12-18', '', '174-0041', '東京都板橋区舟渡2-6-10', '', '090-1122-3344', '', 'yuuki.nakamura@example.com', '', false, 0, 'non_member'),
    (9,  3, '加藤 恵', 'カトウ メグミ', '1995-06-25', '', '134-0083', '東京都江戸川区中葛西5-3-2', '', '080-5566-7788', '', 'megumi.kato@example.com', '', false, 10, 'member'),
    (10, 3, '山田 太郎', 'ヤマダ タロウ', '1970-01-10', '', '144-0051', '東京都大田区西蒲田6-8-4', '', '090-2233-4455', '', 'taro.yamada@example.com', '', false, 0, 'non_member'),
    (11, 3, '高橋 由美', 'タカハシ ユミ', '1988-08-15', '', '110-0005', '東京都台東区上野5-1-3', '', '080-6677-8899', '', 'yumi.takahashi@example.com', '', false, 0, 'member'),
    (12, 3, '松本 隆', 'マツモト タカシ', '1965-03-28', '', '125-0061', '東京都葛飾区亀有3-9-7', '', '090-3344-5566', '', 'takashi.matsumoto@example.com', '', false, 0, 'non_member'),
    (13, 3, '吉田 誠', 'ヨシダ マコト', '1982-11-05', '', '123-0845', '東京都足立区西新井7-4-6', '', '080-7788-9900', '', 'makoto.yoshida@example.com', '', false, 0, 'non_member'),
    (14, 3, '井上 京子', 'イノウエ キョウコ', '1973-05-19', '', '189-0023', '東京都東村山市美住町1-5-2', '', '090-4455-6677', '', 'kyoko.inoue@example.com', '', false, 5, 'member'),
    (15, 3, '木村 拓也', 'キムラ タクヤ', '1991-07-14', '', '179-0081', '東京都練馬区北町3-6-9', '', '080-8899-0011', '', 'takuya.kimura@example.com', '', false, 0, 'non_member'),
    (16, 3, '佐々木 亮', 'ササキ リョウ', '1986-02-23', '', '207-0013', '東京都東大和市清水2-4-8', '', '090-5566-7788', '', 'ryo.sasaki@example.com', '', false, 0, 'non_member'),
    (17, 3, '山本 健太', 'ヤマモト ケンタ', '1998-09-12', '', '206-0802', '東京都稲城市東長沼2-8-3', '', '090-1234-9876', '', 'kenta.yamamoto@example.com', '', false, 0, 'non_member'),
    (18, 3, '青木 麻衣', 'アオキ マイ', '1993-03-10', '', '150-0002', '東京都渋谷区渋谷2-1-1', '', '090-1111-1111', '', 'mai.aoki@example.com', '', false, 0, 'non_member'),
    (19, 3, '橋本 俊介', 'ハシモト シュンスケ', '1980-07-25', '', '130-0001', '東京都墨田区吾妻橋1-3-5', '', '080-2222-2222', '', 'shunsuke.h@example.com', '', false, 0, 'member'),
    (20, 3, '福田 裕子', 'フクダ ユウコ', '1977-11-14', '', '145-0062', '東京都大田区北千束2-5-8', '', '090-3333-3333', '', 'yuko.fukuda@example.com', '', false, 5, 'member'),
    (21, 3, '石川 大輔', 'イシカワ ダイスケ', '1989-04-02', '', '167-0041', '東京都杉並区善福寺3-2-6', '', '080-4444-4444', '', 'daisuke.ishikawa@example.com', '', false, 0, 'non_member'),
    (22, 3, '村田 奈々', 'ムラタ ナナ', '1996-09-19', '', '182-0021', '東京都調布市調布ヶ丘1-4-7', '', '090-5555-5555', '', 'nana.murata@example.com', '', false, 0, 'non_member')
ON CONFLICT (id) DO UPDATE SET
    name      = EXCLUDED.name,
    name_kana = EXCLUDED.name_kana,
    updated_at      = now();

SELECT setval(pg_get_serial_sequence('owners', 'id'), (SELECT MAX(id) FROM owners));

-- -----------------------------------------------------------------------------
-- 2. pets（ペット: 28件）
-- -----------------------------------------------------------------------------
INSERT INTO pets (id, clinic_id, owner_id, pet_number, name, name_kana, animal_species_id, gender, status, birth_date, breed, color, weight, insurance_id, last_visit) VALUES
    (1,  3, 1,  '1-1', 'Iris(イリス)', 'イリス', 1, 'male',   'alive', '2015-04-14', 'ゴールデンレトリーバー',     '茶色',           26.5,  1, '2015-08-28'),
    (2,  3, 1,  '1-2', 'Max(マックス)', 'マックス', 1, 'male', 'alive', '2018-06-20', 'ラブラドール',               'ゴールデン',     15.2,  NULL, '2024-11-15'),
    (3,  3, 2,  '2-1', 'ミケ',         'ミケ',     2, 'female','alive', '2020-03-10', '三毛猫',                     '三毛',            4.20, 2, '2024-11-18'),
    (4,  3, 3,  '3-1', 'タロウ',       'タロウ',   1, 'male',  'alive', '2019-05-15', '柴犬',                       'レッド',          8.3,  NULL, NULL),
    (5,  3, 3,  '3-2', 'ジロウ',       'ジロウ',   1, 'male',  'alive', '2021-08-10', '柴犬',                       'ブラック',        7.1,  NULL, NULL),
    (6,  3, 4,  '4-1', 'チョコ',       'チョコ',   1, 'female','alive', '2017-11-20', 'トイプードル',               'チョコ',          3.80, 1, NULL),
    (7,  3, 5,  '5-1', 'レオ',         'レオ',     2, 'male',  'alive', '2016-07-04', 'スコティッシュフォールド',   'グレー',          5.5,  NULL, NULL),
    (8,  3, 6,  '6-1', 'ハチ',         'ハチ',     1, 'male',  'alive', '2018-03-25', '秋田犬',                     'ホワイト',       22.0,  NULL, NULL),
    (9,  3, 7,  '7-1', 'モモ',         'モモ',     2, 'female','alive', '2022-01-15', 'マンチカン',                 'キャリコ',        3.2,  2, NULL),
    (10, 3, 8,  '8-1', 'ロッキー',     'ロッキー', 1, 'male',  'alive', '2014-09-08', 'ボーダーコリー',             'ブラックホワイト',18.5,  NULL, NULL),
    (11, 3, 9,  '9-1', 'ルナ',         'ルナ',     2, 'female','alive', '2021-02-28', 'ペルシャ',                   'シルバー',        4.80, 1, NULL),
    (12, 3, 10, '10-1', 'ケン',        'ケン',     1, 'male',  'alive', '2013-06-18', 'ジャーマンシェパード',       'ブラックタン',   32.0,  NULL, NULL),
    (13, 3, 11, '11-1', 'ソラ',        'ソラ',     2, 'male',  'alive', '2023-04-01', 'アメリカンショートヘア',     'タビー',          3.0,  NULL, NULL),
    (14, 3, 12, '12-1', 'ゴン',        'ゴン',     1, 'male',  'alive', '2016-12-05', '紀州犬',                     'ホワイト',       19.5,  NULL, NULL),
    (15, 3, 13, '13-1', 'シロ',        'シロ',     1, 'male',  'alive', '2020-08-10', 'ミックス犬',                 'ホワイト',        6.2,  NULL, NULL),
    (16, 3, 14, '14-1', 'トラ',        'トラ',     2, 'male',  'alive', '2019-10-22', 'トラ猫',                     'トラ',            5.1,  NULL, NULL),
    (17, 3, 15, '15-1', 'ベロ',        'ベロ',     1, 'male',  'alive', '2018-05-03', 'ビーグル',                   'トライカラー',   13.2,  NULL, NULL),
    (18, 3, 16, '16-1', 'チビ',        'チビ',     2, 'female','alive', '2022-06-20', 'ミックス猫',                 'サビ',            3.50, NULL, NULL),
    (19, 3, 17, '17-1', 'ポチ',        'ポチ',     1, 'male',  'alive', '2017-02-14', 'ダックスフンド',             'チョコ',          7.8,  NULL, NULL),
    (20, 3, 18, '18-1', 'モカ',        'モカ',     2, 'female','alive', '2022-05-10', 'ミックス猫',                 'ホワイト',        4.1,  NULL, NULL),
    (21, 3, 18, '18-2', 'クルミ',      'クルミ',   1, 'male',  'alive', '2020-08-20', 'ミックス犬',                 'ベージュ',        8.3,  NULL, NULL),
    (22, 3, 19, '19-1', 'ハル',        'ハル',     1, 'male',  'alive', '2019-03-15', 'ミックス犬',                 'ブラック',       12.5,  NULL, NULL),
    (23, 3, 19, '19-2', 'ユキ',        'ユキ',     2, 'female','alive', '2021-12-01', 'ミックス猫',                 'ホワイト',        3.80, NULL, NULL),
    (24, 3, 20, '20-1', 'ピーチ',      'ピーチ',   2, 'female','alive', '2023-01-07', 'ミックス猫',                 'オレンジ',        3.2,  NULL, NULL),
    (25, 3, 21, '21-1', 'コタ',        'コタ',     1, 'male',  'alive', '2018-09-23', 'ミックス犬',                 'ブラウン',       22.0,  NULL, NULL),
    (26, 3, 21, '21-2', 'アン',        'アン',     2, 'female','alive', '2020-04-11', 'ミックス猫',                 'キャリコ',        4.5,  NULL, NULL),
    (27, 3, 22, '22-1', 'ゴマ',        'ゴマ',     2, 'male',  'alive', '2022-11-30', 'ミックス猫',                 'グレー',          5.0,  NULL, NULL),
    (28, 3, 22, '22-2', 'マル',        'マル',     1, 'female','alive', '2021-06-18', 'ミックス犬',                 'ゴールデン',      9.7,  NULL, NULL)
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();

SELECT setval(pg_get_serial_sequence('pets', 'id'), (SELECT MAX(id) FROM pets));

-- -----------------------------------------------------------------------------
-- 3. appointments（予約: 10件）
-- -----------------------------------------------------------------------------
INSERT INTO appointments (id, clinic_id, start_time, end_time, owner_id, pet_id, visit_type, reservation_category_id, doctor_id, is_designated, status, notes) VALUES
    (1,  3, '2026-03-12 09:00:00+09', '2026-03-12 09:15:00+09', 1,  1,  'revisit', 1, 1, true,  'completed',       '皮膚の経過観察'),
    (2,  3, '2026-03-12 09:15:00+09', '2026-03-12 09:30:00+09', 2,  3,  'revisit', 7, 2, false, 'accounting',      '猫の定期健診'),
    (3,  3, '2026-03-12 10:00:00+09', '2026-03-12 10:15:00+09', 3,  4,  'revisit', 1, 1, true,  'in_consultation', '足を引きずっている'),
    (4,  3, '2026-03-12 10:15:00+09', '2026-03-12 10:30:00+09', 4,  6,  'first',   3, 2, false, 'checked_in',      'ワクチン接種希望'),
    (5,  3, '2026-03-12 14:00:00+09', '2026-03-12 14:15:00+09', 6,  8,  'revisit', 1, 1, false, 'confirmed',       '食欲低下が続いている'),
    (6,  3, '2026-03-13 09:00:00+09', '2026-03-13 09:15:00+09', 7,  9,  'revisit', 1, 2, true,  'confirmed',       '耳の治療経過確認'),
    (7,  3, '2026-03-13 10:00:00+09', '2026-03-13 10:15:00+09', 8,  10, 'first',   1, 1, false, 'confirmed',       '嘔吐が続いている'),
    (8,  3, '2026-03-14 09:30:00+09', '2026-03-14 09:45:00+09', 9,  11, 'revisit', 1, 2, false, 'confirmed',       'ルナの経過観察'),
    (9,  3, '2026-03-15 11:00:00+09', '2026-03-15 11:15:00+09', 10, 12, 'first',   3, 1, false, 'confirmed',       '初回ワクチン接種'),
    (10, 3, '2026-03-16 14:00:00+09', '2026-03-16 14:15:00+09', 11, 13, 'revisit', 1, 2, true,  'confirmed',       '腎臓値の経過観察')
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();

SELECT setval(pg_get_serial_sequence('appointments', 'id'), (SELECT MAX(id) FROM appointments));

-- -----------------------------------------------------------------------------
-- 4. medical_records（カルテ: 20件）
-- -----------------------------------------------------------------------------
INSERT INTO medical_records (id, clinic_id, record_no, date, owner_id, pet_id, doctor_id, status) VALUES
    (1,  3, 'R-2025-001', '2025-10-10', 1,  1,  1, 'finalized'),
    (2,  3, 'R-2025-002', '2025-12-15', 1,  1,  1, 'finalized'),
    (3,  3, 'R-2026-001', '2026-01-20', 1,  1,  2, 'finalized'),
    (4,  3, 'R-2025-003', '2025-11-05', 1,  2,  2, 'finalized'),
    (5,  3, 'R-2025-004', '2025-09-15', 2,  3,  2, 'finalized'),
    (6,  3, 'R-2026-002', '2026-01-06', 2,  3,  1, 'finalized'),
    (7,  3, 'R-2025-005', '2025-08-22', 3,  4,  2, 'finalized'),
    (8,  3, 'R-2025-006', '2025-10-18', 4,  6,  1, 'finalized'),
    (9,  3, 'R-2025-007', '2025-07-30', 5,  7,  2, 'finalized'),
    (10, 3, 'R-2026-003', '2026-01-15', 6,  8,  1, 'draft'),
    (11, 3, 'R-2025-008', '2025-12-01', 7,  9,  2, 'finalized'),
    (12, 3, 'R-2025-009', '2025-11-20', 8,  10, 1, 'finalized'),
    (13, 3, 'R-2026-004', '2026-02-10', 9,  11, 2, 'draft'),
    (14, 3, 'R-2025-010', '2025-06-15', 10, 12, 1, 'finalized'),
    (15, 3, 'R-2026-005', '2026-01-06', 11, 13, 2, 'finalized'),
    (16, 3, 'R-2025-011', '2025-09-08', 12, 14, 1, 'finalized'),
    (17, 3, 'R-2026-006', '2026-02-28', 13, 15, 2, 'draft'),
    (18, 3, 'R-2025-012', '2025-08-20', 14, 16, 1, 'finalized'),
    (19, 3, 'R-2026-007', '2026-01-03', 15, 17, 2, 'finalized'),
    (20, 3, 'R-2026-008', '2026-01-06', 16, 18, 1, 'finalized')
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();

SELECT setval(pg_get_serial_sequence('medical_records', 'id'), (SELECT MAX(id) FROM medical_records));

-- -----------------------------------------------------------------------------
-- 5. inquiries（問診: 20件）
-- -----------------------------------------------------------------------------
INSERT INTO inquiries (id, medical_record_id, chief_complaint_category_id, chief_complaint, notes, staff_id) VALUES
    (1,  1,  NULL, '狂犬病ワクチン接種',         '体調良好。', 1),
    (2,  2,  NULL, '定期健診',                   '特に異常なし。', 2),
    (3,  3,  6,    '右足の跛行',                 '膝蓋骨脱臼を確認。', 1),
    (4,  4,  NULL, 'フィラリア予防',             '予防薬処方。', 2),
    (5,  5,  NULL, '5種混合ワクチン接種',        '体調良好。', 1),
    (6,  6,  NULL, '5種混合ワクチン接種',        '体調良好。', 2),
    (7,  7,  NULL, 'ノミダニ予防薬',             '予防薬処方。', 1),
    (8,  8,  3,    '皮膚の痒み',                 'アトピー性皮膚炎疑い。', 2),
    (9,  9,  3,    'トリミング後の皮膚チェック', '軽度の赤みあり。', 1),
    (10, 10, 1,    '食欲不振',                   '2日前から食欲減退。', 2),
    (11, 11, NULL, '耳を痒がる',                 '外耳炎疑い。', 1),
    (12, 12, NULL, '定期健診・予防接種',         '年次健診。', 2),
    (13, 13, 2,    '嘔吐・下痢',                 '昨日から嘔吐3回。', 1),
    (14, 14, NULL, '生化学検査',                 'シニア健診。', 2),
    (15, 15, NULL, '3種混合ワクチン接種（猫）',  '初回ワクチン。', 1),
    (16, 16, NULL, '血液検査',                   '異常なし。', 2),
    (17, 17, NULL, '歯石除去',                   '重度の歯石付着。', 1),
    (18, 18, NULL, '定期検診',                   '体重管理継続。', 2),
    (19, 19, 6,    '再診（右足跛行）',           '改善傾向。', 1),
    (20, 20, NULL, '5種混合ワクチン接種',        '体調良好。', 2)
ON CONFLICT (id) DO UPDATE SET
    medical_record_id = EXCLUDED.medical_record_id,
    updated_at        = now();

SELECT setval(pg_get_serial_sequence('inquiries', 'id'), (SELECT MAX(id) FROM inquiries));

-- -----------------------------------------------------------------------------
-- 5b. clinical_plans（診察/治療プラン: 20件）
-- -----------------------------------------------------------------------------
INSERT INTO clinical_plans (id, medical_record_id, physical_exam, diagnosis_category_id, diagnosis_name_id, diagnosis_details, treatment_policy) VALUES
    (1,  1, '体温38.5℃。心肺音正常。', NULL, NULL, '健康状態良好。ワクチン接種可。', '5種混合ワクチン接種実施。'),
    (2,  2, '体重増加あり。他異常なし。', NULL, NULL, '維持状態良好。', '定期検診継続。'),
    (3,  3, '右後肢跛行。パテラG2。', 8, 42, '膝蓋骨脱臼。', '消炎剤処方。体重管理指導。'),
    (4,  4, '異常なし。', NULL, NULL, '予防シーズン開始。', 'フィラリア予防薬処方。'),
    (5,  5, '良好。', NULL, NULL, '年次予防。', 'ワクチン接種。'),
    (6,  6, '良好。', NULL, NULL, '年次予防。', 'ワクチン接種。'),
    (7,  7, '良好。', NULL, NULL, '外部寄生虫予防。', 'スポットオン投与。'),
    (8,  8, '全身に発赤。搔痒感強。', 3, 6, 'アトピー性皮膚炎。', '抗ヒスタミン薬処方。薬用シャンプー推奨。'),
    (9,  9, '皮膚の一部に発赤。', 3, 7, '膿皮症初期。', '洗浄と消毒。'),
    (10, 10, '腹部軽度緊張。', 1, 1, '急性胃腸炎疑い。', '絶食・皮下補液実施。'),
    (11, 11, '耳道内に分泌物。', 3, 41, '外耳炎。', '耳道洗浄・点耳薬処方。'),
    (12, 12, 'シニア期に入る。', NULL, NULL, '健康診断実施。', '結果待ち。'),
    (13, 13, '脱水傾向あり。', 1, 1, '急性胃腸炎。', '対症療法と食事療法。'),
    (14, 14, '良好。', NULL, NULL, '経過観察。', '維持。'),
    (15, 15, '良好。', NULL, NULL, '幼若期検診。', '成長記録。'),
    (16, 16, '良好。', NULL, NULL, 'スクリーニング。', '異常なし。'),
    (17, 17, '重度の歯石。', NULL, NULL, '歯周病。', '抜歯を含めた歯科処置を計画。'),
    (18, 18, '良好。', NULL, NULL, '肥満気味。', 'ダイエットフード提案。'),
    (19, 19, '跛行消失。', 8, 42, '回復期。', '運動制限解除。'),
    (20, 20, '良好。', NULL, NULL, '年次予防。', 'ワクチン接種。')
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();

SELECT setval(pg_get_serial_sequence('clinical_plans', 'id'), (SELECT MAX(id) FROM clinical_plans));

-- -----------------------------------------------------------------------------
-- 6. vital_records（バイタル: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO vital_records (id, pet_id, medical_record_id, recorded_at, staff_id, temperature, heart_rate, respiration_rate, weight, weight_unit, notes) VALUES
    (1, 1,  3, '2026-01-20 09:15:00+09', 1, 38.5, 80,  20, 26.5, 'Kg', '皮膚の搔痒感あり。体重良好。'),
    (2, 1,  2, '2025-12-15 10:00:00+09', 2, 38.8, 82,  22, 26.0, 'Kg', '体重前回比-500g'),
    (3, 1,  3, '2026-01-20 09:30:00+09', 1, 38.3, 78,  20, 26.5, 'Kg', '定期検診。皮膚搔痒感 軽快傾向。'),
    (4, 2,  4, '2025-11-05 11:00:00+09', 1, 39.1, 95,  24, 15.2, 'Kg', '軽度脱水。CRT 2秒。'),
    (5, 3,  5, '2025-09-15 14:30:00+09', 2, 38.2, 160, 30, 4200, 'g',  '粘膜色やや蒼白。食欲低下継続。')
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();

SELECT setval(pg_get_serial_sequence('vital_records', 'id'), (SELECT MAX(id) FROM vital_records));

-- -----------------------------------------------------------------------------
-- 7. treatments（治療明細: 8件）
-- -----------------------------------------------------------------------------
INSERT INTO treatments (id, medical_record_id, item_type, consultation_id, procedure_id, medicine_id, inventory_id, is_selected, status, content, unit_price, quantity, sort_order) VALUES
    (1, 3, 'consultation', 2,    NULL, NULL, NULL, true, 'completed', '再診料',                    800,  1, 1),
    (2, 1, 'medicine',     NULL, NULL, 1,    NULL, true, 'completed', 'アモキシシリン 50mg x 7日分', 500,  7, 2),
    (3, 2, 'consultation', 2,    NULL, NULL, NULL, true, 'completed', '再診料',                    800,  1, 1),
    (4, 3, 'consultation', 2,    NULL, NULL, NULL, true, 'completed', '再診料',                    800,  1, 1),
    (5, 3, 'procedure',    NULL, 4,    NULL, NULL, true, 'completed', '耳道洗浄（左耳）',          2500, 1, 2),
    (6, 4, 'consultation', 1,    NULL, NULL, NULL, true, 'completed', '初診料',                    2000, 1, 1),
    (7, 4, 'medicine',     NULL, NULL, 1,    NULL, true, 'completed', 'アモキシシリン 50mg x 5日分', 500,  5, 2),
    (8, 5, 'consultation', 2,    NULL, NULL, NULL, true, 'completed', '再診料',                    800,  1, 1)
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();

SELECT setval(pg_get_serial_sequence('treatments', 'id'), (SELECT MAX(id) FROM treatments));

-- -----------------------------------------------------------------------------
-- 8. trimming_records（トリミング: 8件）
-- -----------------------------------------------------------------------------
INSERT INTO trimming_records (id, clinic_id, date, pet_id, body_weight, bw_unit, style_request, staff_id, status, course_id) VALUES
    (1, 3, '2025-10-10', 1,  26.5,  'Kg', 'サマーカット希望',        6,  'completed',   5),
    (2, 3, '2025-10-15', 2,  15.2,  'Kg', 'ふんわりカット',          12, 'reserved',    4),
    (3, 3, '2025-10-12', 3,  4.2,   'Kg', '毛玉カット',              6,  'in_progress', 1),
    (4, 3, '2026-01-06', 6,  3800,  'g',  'シャンプーコース',        6,  'completed',   1),
    (5, 3, '2026-01-06', 17, 12.0,  'Kg', '全体カット',              12, 'completed',   4),
    (6, 3, '2026-01-06', 10, 8.0,   'Kg', '爪切り・ブラッシング',   12, 'reserved',    2),
    (7, 3, '2026-01-06', 15, 5.0,   'Kg', 'シャンプー',              6,  'completed',   1),
    (8, 3, '2026-01-06', 6,  3800,  'g',  'トリミング',              6,  'reserved',    3)
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();

SELECT setval(pg_get_serial_sequence('trimming_records', 'id'), (SELECT MAX(id) FROM trimming_records));

-- -----------------------------------------------------------------------------
-- 9. hospitalizations（入院: 7件）
-- -----------------------------------------------------------------------------
INSERT INTO hospitalizations (id, clinic_id, owner_id, pet_id, hospitalization_type, start_date, end_date, status, cage_id, doctor_id, memo, owner_request, staff_notes) VALUES
    (1, 3, 3, 5,  'hospitalization', '2026-03-10', '2026-03-14', 'admitted',   5,    1, '急性胃腸炎による脱水治療。点滴管理中。',  '食事のアレルギーに注意してほしい（鶏肉不可）', '3/10入院開始。静脈点滴開始。3/11嘔吐1回。3/12状態改善傾向。'),
    (2, 3, 6, 8,  'hospitalization', '2026-02-25', '2026-02-28', 'discharged', 4,    1, '外耳炎重症化に伴う入院治療。',             '怖がりなので優しく接してほしい',               '耳道洗浄を毎日実施。2/28退院時、症状改善。点耳薬処方。'),
    (3, 3, 17, 19, 'hospitalization', '2026-02-10', '2026-02-20', 'discharged', NULL, 1, '骨折治療による入院。手術後経過観察。', '', '2/10手術実施。2/15抜糸。2/20退院。'),
    (4, 3, 4,  6,  'hotel',           '2026-03-15', '2026-03-18', 'reserved',   NULL, 1, '旅行中のホテル預かり。', 'フードはロイヤルカナンのみ', ''),
    (5, 3, 1,  1,  'hospitalization', '2026-03-20', '2026-03-25', 'reserved',   NULL, 1, '膝蓋骨脱臼手術予定。術前検査済み。', '怖がりなので静かな環境を希望', ''),
    (6, 3, 9,  11, 'hospitalization', '2026-03-05', '2026-03-12', 'admitted',   1,    2, '慢性腎臓病の集中治療。点滴管理中。', 'ペルシャ猫のため温度管理に注意', '3/5入院。毎日皮下補液実施。3/12現在状態安定。'),
    (7, 3, 3,  4,  'hospitalization', '2026-01-03', '2026-01-06', 'discharged', NULL, 1, '急性胃腸炎による脱水治療。', 'チキンアレルギーあり', '1/3入院。点滴開始。1/6状態改善し退院。')
ON CONFLICT (id) DO UPDATE SET
    updated_at            = now();

SELECT setval(pg_get_serial_sequence('hospitalizations', 'id'), (SELECT MAX(id) FROM hospitalizations));

-- -----------------------------------------------------------------------------
-- 10. care_plan_items（ケアプラン: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO care_plan_items (id, hospitalization_id, type, name, description, timing, status, notes, medicine_id, procedure_id, hospitalization_plan_id, unit_price, category, sort_order) VALUES
    (1, 1, 'food',        '療法食（消化器ケア）', '1日3回、少量ずつ与える', ARRAY['morning','noon','night']::plan_timing[], 'active', '鶏肉不可。', NULL, NULL, NULL, 0, '食事', 1),
    (2, 1, 'medicine',    'アモキシシリン',       '1回1錠、朝夕食後',       ARRAY['morning','night']::plan_timing[],       'active', '抗生剤。', 1,    NULL, NULL, 500, '投薬', 2),
    (3, 1, 'instruction', 'バイタルチェック',     '1日3回測定',             ARRAY['morning','noon','night']::plan_timing[], 'active', '異常値報告。', NULL, NULL, NULL, 0, '観察', 3),
    (4, 2, 'treatment',   '耳道洗浄',             '1日1回、朝に実施',       ARRAY['morning']::plan_timing[],               'completed', '左耳。', NULL, 4,    NULL, 2500, '処置', 1),
    (5, 2, 'item',        '入院管理料',           '小型犬1日分',            ARRAY['morning']::plan_timing[],               'completed', '', NULL, NULL, 1,    3000, '入院', 2)
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();

SELECT setval(pg_get_serial_sequence('care_plan_items', 'id'), (SELECT MAX(id) FROM care_plan_items));

-- -----------------------------------------------------------------------------
-- 11. inventory_items（在庫管理: 9件追加）
-- -----------------------------------------------------------------------------
INSERT INTO inventory_items (id, clinic_id, name, category, quantity, unit, min_stock_level, location, supplier, status) VALUES
    (6,  3, '5種混合ワクチン',               'medicine',   25,  'バイアル', 15, '冷蔵庫 1',    '共立製薬',                'sufficient'),
    (7,  3, '留置針 22G',                    'consumable',  0,   '本',       50, '処置室 棚D',  'テルモ',                  'out_of_stock'),
    (8,  3, 'シリンジ 5mL',                  'consumable', 300,  '本',      100, '処置室 棚D',  'テルモ',                  'sufficient'),
    (9,  3, 'メトクロプラミド注 10mg',        'medicine',    8,   'アンプル', 10, '薬品棚 A-3', '日本全薬工業',            'low'),
    (10, 3, '療法食 消化器サポート（猫用）',  'food',       10,   '袋',        5, 'フード棚 C-1','ヒルズ',                 'sufficient'),
    (11, 3, 'エリザベスカラー（S）',          'other',      15,   '個',        5, '倉庫 A',     'ペットメディカルサプライ', 'sufficient'),
    (12, 3, 'ガーゼ 滅菌 7.5cm',            'consumable',  45,   '枚',       50, '処置室 棚E',  '白十字',                  'low'),
    (13, 3, 'フィラリア予防薬（S）',          'medicine',   60,   '錠',       30, '薬品棚 B-1', 'メリアル・ジャパン',       'sufficient'),
    (14, 3, 'ノミダニ駆除薬 スポット',        'medicine',   40,   'ピペット',  20, '薬品棚 B-2', 'エランコジャパン',         'sufficient')
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('inventory_items', 'id'), (SELECT MAX(id) FROM inventory_items));

-- -----------------------------------------------------------------------------
-- 12. billings / billing_items / payments
-- -----------------------------------------------------------------------------
INSERT INTO billings (id, clinic_id, medical_record_id, hospitalization_id, owner_id, pet_id, subtotal, tax_total, total_amount, has_insurance, status, scheduled_date, completed_at, memo) VALUES
    (1, 3, 1,    NULL, 1,  1,  4300, 430, 4730, true, 'completed', '2026-02-15', '2026-02-15 10:30:00+09', 'アニコム保険適用'),
    (2, 3, 3,    NULL, 1,  1,  3300, 330, 3630, true, 'completed', '2026-02-28', '2026-02-28 11:00:00+09', 'アニコム保険適用（Iris 耳炎治療）'),
    (3, 3, 6,    NULL, 2,  3,  800,  80,  880,  true, 'waiting',   '2026-03-12', NULL,                     'アニコム保険適用。会計待ち。')
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();

SELECT setval(pg_get_serial_sequence('billings', 'id'), (SELECT MAX(id) FROM billings));

INSERT INTO billing_items (id, billing_id, category, name, unit_price, quantity, tax_rate, is_insurance_applicable, source, sort_order) VALUES
    (1, 1, 'other',    '再診料',                    800,  1, 0.10, true, 'medical_record', 1),
    (2, 1, 'medicine', 'アモキシシリン 50mg x 7日分', 500,  7, 0.10, true, 'medical_record', 2),
    (3, 2, 'other',    '再診料',                    800,  1, 0.10, true, 'medical_record', 1),
    (4, 2, 'procedure','耳道洗浄',                  2500, 1, 0.10, true, 'medical_record', 2),
    (5, 3, 'other',    '再診料',                    800,  1, 0.10, true, 'medical_record', 1)
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('billing_items', 'id'), (SELECT MAX(id) FROM billing_items));

INSERT INTO payments (id, billing_id, subtotal, tax_total, total_amount, insurance_name, insurance_ratio, insurance_amount, discount_amount, billing_amount, received_amount, change_amount, method) VALUES
    (1, 1, 4300, 430, 4730, 'アニコム損保', 0.70, 3311, 0, 1419, 1500, 81, 'cash'),
    (2, 2, 3300, 330, 3630, 'アニコム損保', 0.70, 2541, 0, 1089, 1100, 11, 'credit_card')
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('payments', 'id'), (SELECT MAX(id) FROM payments));

-- -----------------------------------------------------------------------------
-- 13. billing_refunds（返金デモデータ）
-- -----------------------------------------------------------------------------
INSERT INTO billing_refunds (id, clinic_id, billing_id, amount, reason, refunded_at) VALUES
    (1, 3, 1, 919,  '処置内容の変更に伴う部分返金',   '2026-02-16 10:00:00+09'),
    (2, 3, 1, 500,  '薬剤変更による差額返金',         '2026-02-20 14:30:00+09'),
    (3, 3, 2, 500,  '診察キャンセル分の返金',          '2026-03-01 09:00:00+09')
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('billing_refunds', 'id'), (SELECT MAX(id) FROM billing_refunds));

-- =============================================================================
-- 城東医院 (clinic_id=4) マスタデータ
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 城東医院 reservation_categories（サービス種別: 19件、ID 26-44）
-- -----------------------------------------------------------------------------

-- 城東医院 公開コース
INSERT INTO reservation_categories (id, clinic_id, name, short_name, is_active, description, color, sort_order, duration_minutes, reservation_visible, reservation_comment, is_internal) VALUES
    (26, 4, '一般診察',               '診察',     true, '内科・外科・皮膚科などの一般的な診察',         '#3B82F6', 1,  15, true,  '', false),
    (27, 4, '一般診察(再診)',          '再診',     true, '継続通院の一般診察',                           '#3B82F6', 2,  15, true,  '', false),
    (28, 4, 'ワクチン接種',            'ワクチン', true, '各種ワクチン接種（予防接種）',                 '#10B981', 3,  15, true,  '', false),
    (29, 4, '狂犬病',                 '狂犬病',   true, '狂犬病予防法に基づくワクチン接種',             '#10B981', 4,  15, true,  '', false),
    (30, 4, 'フィラリア予防',          'フィラリア', true, 'フィラリア予防薬投与・処方',                '#10B981', 5,  15, true,  '', false),
    (31, 4, '健康診断',               '健診',     true, '定期健康診断・フィラリア検査など',             '#8B5CF6', 6,  15, true,  '', false),
    (32, 4, '健康診断結果報告',        '結果報告', true, '健康診断結果の説明・報告',                     '#8B5CF6', 7,  15, true,  '', false),
    (33, 4, 'トリミングコース',        'トリミング', true, 'カット・シャンプー・ブロー・爪切り・耳掃除', '#F59E0B', 8,  15, true,  '', false),
    (34, 4, 'トリミングシャンプーコース', 'シャンプー', true, 'シャンプー・ブロー・ブラッシング',        '#F59E0B', 9,  15, true,  '', false),
    (35, 4, 'クイックシャンプー',      'Qシャンプー', true, '短時間シャンプー',                        '#F59E0B', 10, 15, true,  '', false)
ON CONFLICT DO NOTHING;

-- 城東医院 スタッフ専用コース
INSERT INTO reservation_categories (id, clinic_id, name, short_name, is_active, description, color, sort_order, duration_minutes, reservation_visible, reservation_comment, is_internal) VALUES
    (36, 4, '手術60',                 '手術60',   true, '手術枠（60分）',                               '#EF4444', 11, 60, false, '', true),
    (37, 4, '手術30',                 '手術30',   true, '手術枠（30分）',                               '#EF4444', 12, 30, false, '', true),
    (38, 4, 'ホテルお迎え',           'お迎え',   true, 'ホテルお迎え対応',                             '#6B7280', 13, 15, false, '', true),
    (39, 4, 'ホテル預かり',           '預かり',   true, 'ペットホテル預かり',                           '#6B7280', 14, 15, false, '', true),
    (40, 4, '休憩枠',                 '休憩',     true, '休憩・ブロック枠',                             '#6B7280', 15, 60, false, '', true),
    (41, 4, '×',                     '×',       true, '予約不可（15分）',                             '#6B7280', 16, 15, false, '', true),
    (42, 4, '予約不可60',             '不可60',   true, '予約不可ブロック（60分）',                     '#6B7280', 17, 60, false, '', true),
    (43, 4, '予約不可30',             '不可30',   true, '予約不可ブロック（30分）',                     '#6B7280', 18, 30, false, '', true),
    (44, 4, 'エコー枠',               'エコー',   true, '超音波検査専用枠',                             '#8B5CF6', 19, 30, false, '', true)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('reservation_categories', 'id'), (SELECT MAX(id) FROM reservation_categories));

-- 城東医院 (clinic_id=4) グループ紐付け
UPDATE reservation_categories SET group_id=8  WHERE clinic_id=4 AND id IN (26,27);
UPDATE reservation_categories SET group_id=9  WHERE clinic_id=4 AND id IN (28,29,30);
UPDATE reservation_categories SET group_id=10 WHERE clinic_id=4 AND id IN (31,32,44);
UPDATE reservation_categories SET group_id=11 WHERE clinic_id=4 AND id IN (33,34,35);
UPDATE reservation_categories SET group_id=12 WHERE clinic_id=4 AND id IN (36,37);
UPDATE reservation_categories SET group_id=13 WHERE clinic_id=4 AND id IN (38,39);
UPDATE reservation_categories SET group_id=14, is_internal=true WHERE clinic_id=4 AND id IN (40,41,42,43);

-- -----------------------------------------------------------------------------
-- 城東医院 cages（ケージ: 4件）
-- -----------------------------------------------------------------------------
INSERT INTO cages (id, clinic_id, name, price, is_active, description, cage_type, cage_size, sort_order) VALUES
    (9,  4, 'ICUケージ',       7500, true, '酸素吸入可・重症患者用',    'icu',     'medium', 1),
    (10, 4, '犬用ケージ（小）', 2800, true, '小型犬・ホテル利用可',      'dog',     'small',  2),
    (11, 4, '犬用ケージ（中）', 3200, true, '中型犬・一般入院用',        'dog',     'medium', 3),
    (12, 4, '猫用ケージ',       2800, true, '猫専用・ストレス軽減設計',  'cat',     'small',  4)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('cages', 'id'), (SELECT MAX(id) FROM cages));

-- -----------------------------------------------------------------------------
-- 城東医院 insurances（保険: 3件）
-- -----------------------------------------------------------------------------
INSERT INTO insurances (id, clinic_id, name, is_active, description, coverage_rate, contact_phone, sort_order) VALUES
    (6, 4, 'アニコム損保',   true, 'ペット保険大手・どうぶつ健保シリーズ', 70, '0120-025-034', 1),
    (7, 4, 'アイペット損保', true, 'うちの子シリーズ',                     70, '0120-956-099', 2),
    (8, 4, 'その他（自費）', true, '保険未加入・全額自費',                100, '',             3)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('insurances', 'id'), (SELECT MAX(id) FROM insurances));

-- -----------------------------------------------------------------------------
-- 城東医院 exam_types（検査種別: 3件）
-- -----------------------------------------------------------------------------
INSERT INTO exam_types (id, clinic_id, name, price, is_active, description, sort_order) VALUES
    (6, 4, '血液検査（CBC）', 3000, true, '全血球計算（Complete Blood Count）',     1),
    (7, 4, '血液化学検査',     5000, true, '肝機能・腎機能・血糖値など生化学的検査', 2),
    (8, 4, 'レントゲン検査',   3200, true, 'X線撮影（胸部・腹部・四肢）',            3)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('exam_types', 'id'), (SELECT MAX(id) FROM exam_types));

-- exam_type_fields for clinic 4
INSERT INTO exam_type_fields (id, exam_type_id, name, inspection_value, normal_value, sort_order) VALUES
    (18, 6, 'WBC（白血球数）', '', '6.0-17.0 x10^3/uL', 1),
    (19, 6, 'RBC（赤血球数）', '', '5.5-8.5 x10^6/uL',  2),
    (20, 6, 'HCT（ヘマトクリット）', '', '37-55%',       3),
    (21, 7, 'ALT（GPT）',      '', '10-125 U/L',         1),
    (22, 7, 'BUN（尿素窒素）', '', '7-27 mg/dL',         2),
    (23, 7, 'CRE（クレアチニン）', '', '0.5-1.8 mg/dL',  3),
    (24, 8, '胸部正面',        '', '異常なし',            1),
    (25, 8, '腹部正面',        '', '異常なし',            2)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('exam_type_fields', 'id'), (SELECT MAX(id) FROM exam_type_fields));

-- -----------------------------------------------------------------------------
-- 城東医院 vaccines（ワクチン: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO vaccines (id, clinic_id, name, price, is_active, description, species, interval, sort_order) VALUES
    (11, 4, '混合ワクチン5種（犬）',  4800, true, 'ジステンパー・パルボ・アデノ1型・アデノ2型・パラインフルエンザ', 'dog', '1年',   1),
    (12, 4, '混合ワクチン8種（犬）',  6800, true, '5種＋レプトスピラ3種',                                          'dog', '1年',   2),
    (13, 4, '混合ワクチン3種（猫）',  4200, true, '猫ウイルス性鼻気管炎・カリシウイルス・汎白血球減少症',           'cat', '1年',   3),
    (14, 4, '狂犬病ワクチン',         3000, true, '狂犬病予防法に基づく接種',                                      'dog', '1年',   4),
    (15, 4, 'フィラリア予防薬（小型犬）', 950, true, '体重10kg以下犬用フィラリア予防',                             'dog', '1ヶ月', 5)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('vaccines', 'id'), (SELECT MAX(id) FROM vaccines));

-- -----------------------------------------------------------------------------
-- 城東医院 medicines（薬剤: カテゴリ4件 + 薬剤10件）
-- -----------------------------------------------------------------------------
INSERT INTO medicines (id, clinic_id, name, price, is_active, description, sort_order) VALUES
    (2001, 4, '抗生剤',     NULL, true, '抗生物質カテゴリ',   1),
    (2002, 4, 'ステロイド', NULL, true, 'ステロイド剤カテゴリ', 2),
    (2003, 4, '消炎剤',     NULL, true, '消炎鎮痛剤カテゴリ', 3),
    (2004, 4, '駆虫剤',     NULL, true, '駆虫剤カテゴリ',     4),
    (2005, 4, '輸液',       NULL, true, '輸液カテゴリ',       5)
ON CONFLICT DO NOTHING;

INSERT INTO medicines (id, clinic_id, name, price, is_active, description, dosage_form, medicine_unit, default_quantity, sort_order, parent_id) VALUES
    (101, 4, 'アモキシシリン 50mg',        520,  true, '広域スペクトラム抗生物質',             'tablet',    'per_tablet', 1,    1, 2001),
    (102, 4, 'セファレキシン 250mg',        470,  true, '第1世代セフェム系抗生物質',           'tablet',    'per_tablet', 1,    2, 2001),
    (103, 4, 'メトロニダゾール 250mg',      620,  true, '嫌気性菌・原虫感染症治療薬',          'tablet',    'per_tablet', 1,    3, 2001),
    (104, 4, 'プレドニゾロン 5mg',          420,  true, 'ステロイド系抗炎症・免疫抑制剤',      'tablet',    'per_tablet', 1,    4, 2002),
    (105, 4, 'デキサメタゾン注射液',        720,  true, '強力ステロイド・アレルギー緊急治療',  'injection', 'per_ml',     1,    5, 2002),
    (106, 4, 'メロキシカム経口液',          730,  true, 'NSAIDs・痛み・炎症の緩和',            'liquid',    'per_ml',     1,    6, 2003),
    (107, 4, 'カルプロフェン 25mg',         680,  true, 'NSAIDs・術後疼痛管理',               'tablet',    'per_tablet', 1,    7, 2003),
    (108, 4, 'ノミ・ダニ駆除薬（犬用）',    2600, true, '外部寄生虫予防・駆除（スポットオン）', 'topical',   'per_dose',   1,    8, 2004),
    (109, 4, 'ノミ・ダニ駆除薬（猫用）',    2600, true, '外部寄生虫予防・駆除（スポットオン）', 'topical',   'per_dose',   1,    9, 2004),
    (110, 4, '生理食塩水 500ml',            420,  true, '点滴・洗浄用生理食塩水',              'liquid',    'per_ml',     500,  10, 2005)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('medicines', 'id'), (SELECT MAX(id) FROM medicines));

-- -----------------------------------------------------------------------------
-- 城東医院 consultations（診察項目: 3件）
-- -----------------------------------------------------------------------------
INSERT INTO consultations (id, clinic_id, name, price, is_active, description, time_condition, duration, sort_order) VALUES
    (6, 4, '初診料',       2200, true, '初めての受診または6ヶ月以上受診がない場合', 'first_visit', 30, 1),
    (7, 4, '再診料',        900, true, '継続通院の診察料',                         'revisit',     15, 2),
    (8, 4, '時間外診療料', 3200, true, '診療時間外・休日の緊急診察',               'after_hours', 30, 3)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('consultations', 'id'), (SELECT MAX(id) FROM consultations));

-- -----------------------------------------------------------------------------
-- 城東医院 procedures（処置項目: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO procedures (id, clinic_id, name, price, is_active, description, duration, anesthesia, sort_order) VALUES
    (11, 4, '去勢手術（犬）', 26000, true, '雄犬の去勢手術',                  60, 'general', 1),
    (12, 4, '避妊手術（猫）', 26000, true, '雌猫の避妊手術',                  90, 'general', 2),
    (13, 4, '耳洗浄',          2800, true, '外耳炎治療・耳道内の洗浄処置',    15, 'none',    3),
    (14, 4, '爪切り',           500, true, '爪のカット・やすりがけ',           10, 'none',    4),
    (15, 4, '点滴処置',        3200, true, '静脈内点滴（1時間）',              60, 'none',    5)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('procedures', 'id'), (SELECT MAX(id) FROM procedures));

-- -----------------------------------------------------------------------------
-- 城東医院 hospitalization_plans（入院プラン: 3件）
-- -----------------------------------------------------------------------------
INSERT INTO hospitalization_plans (id, clinic_id, name, price, is_active, description, body_size, billing_unit, sort_order) VALUES
    (6, 4, '一般入院（小型）', 3200, true, '体重10kg以下の入院管理料（1日）', 'small',  'per_day',   1),
    (7, 4, '一般入院（中型）', 3700, true, '体重10-25kgの入院管理料（1日）',  'medium', 'per_day',   2),
    (8, 4, 'ICU入院',          8500, true, '集中治療室管理料（1日）',         'small',  'per_day',   3)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('hospitalization_plans', 'id'), (SELECT MAX(id) FROM hospitalization_plans));

-- -----------------------------------------------------------------------------
-- 城東医院 trimming_courses（トリミングコース: 3件）
-- -----------------------------------------------------------------------------
INSERT INTO trimming_courses (id, clinic_id, name, price, is_active, description, target_size, duration, sort_order) VALUES
    (6, 4, 'シャンプー&ブロー（小型）', 4200, true, 'シャンプー・ブロー・ブラッシング',            'small',  60,  1),
    (7, 4, 'フルコース（小型）',        7200, true, 'カット・シャンプー・ブロー・爪切り・耳掃除', 'small',  120, 2),
    (8, 4, 'フルコース（中型）',        9500, true, 'カット・シャンプー・ブロー・爪切り・耳掃除', 'medium', 150, 3)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('trimming_courses', 'id'), (SELECT MAX(id) FROM trimming_courses));

-- -----------------------------------------------------------------------------
-- 城東医院 trimming_options（トリミングオプション: 3件）
-- -----------------------------------------------------------------------------
INSERT INTO trimming_options (id, clinic_id, name, price, is_active, description, duration, is_combinable, sort_order) VALUES
    (6, 4, '爪切り',     320, true, '爪のカット・やすりがけ',    10, true, 1),
    (7, 4, '耳掃除',     520, true, '外耳道の洗浄・清掃',        10, true, 2),
    (8, 4, '肛門腺絞り', 320, true, '肛門嚢の分泌液除去',         5, true, 3)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('trimming_options', 'id'), (SELECT MAX(id) FROM trimming_options));

-- -----------------------------------------------------------------------------
-- 城東医院 diagnosis_types（診断カテゴリ: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO diagnosis_types (id, clinic_id, name, is_active, description, sort_order) VALUES
    (9,  4, '消化器系',   true, '胃腸・肝臓・膵臓などの消化器系疾患', 1),
    (10, 4, '呼吸器系',   true, '肺・気管・鼻腔などの呼吸器系疾患',   2),
    (11, 4, '皮膚・被毛', true, 'アレルギー・感染症などの皮膚疾患',   3),
    (12, 4, '泌尿器系',   true, '腎臓・膀胱・尿道などの泌尿器系疾患', 4),
    (13, 4, '感染症',     true, '細菌・ウイルス感染症',               5)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('diagnosis_types', 'id'), (SELECT MAX(id) FROM diagnosis_types));

-- -----------------------------------------------------------------------------
-- 城東医院 diagnosis_names（診断名: 10件）
-- -----------------------------------------------------------------------------
INSERT INTO diagnosis_names (id, clinic_id, name, is_active, description, diagnosis_category_id, sort_order) VALUES
    (21, 4, '胃腸炎',             true, '胃・腸の炎症（嘔吐・下痢）',         9,  1),
    (22, 4, '膵炎',               true, '膵臓の炎症',                         9,  2),
    (23, 4, '気管支炎',           true, '気管支の炎症',                       10, 1),
    (24, 4, '肺炎',               true, '肺の感染性・非感染性炎症',           10, 2),
    (25, 4, 'アトピー性皮膚炎',   true, 'アレルゲンによるアレルギー性皮膚炎', 11, 1),
    (26, 4, '膿皮症',             true, '細菌性の皮膚感染症',                 11, 2),
    (27, 4, '膀胱炎',             true, '細菌性・特発性膀胱炎',               12, 1),
    (28, 4, '腎不全',             true, '急性・慢性腎不全',                   12, 2),
    (29, 4, 'パルボウイルス感染症', true, '犬パルボウイルスによる感染症',       13, 1),
    (30, 4, '猫風邪（FVR）',      true, '猫ウイルス性鼻気管炎',               13, 2)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('diagnosis_names', 'id'), (SELECT MAX(id) FROM diagnosis_names));

-- -----------------------------------------------------------------------------
-- 城東医院 checkup_types（健診種別: 3件）
-- -----------------------------------------------------------------------------
INSERT INTO checkup_types (id, clinic_id, name, price, is_active, description, interval, target_age, sort_order) VALUES
    (5, 4, '一般健診',       5200,  true, '身体検査・体重測定・問診',                     '1年',   '全年齢',  1),
    (6, 4, '老齢検診',       16000, true, '身体検査＋血液検査＋レントゲン＋超音波',         '6ヶ月', '7歳以上', 2),
    (7, 4, 'フィラリア検査', 2600,  true, 'フィラリア抗原検査（予防シーズン前）',           '1年',   '成犬',    3)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('checkup_types', 'id'), (SELECT MAX(id) FROM checkup_types));

-- -----------------------------------------------------------------------------
-- 城東医院 chief_complaint_types（主訴区分: 4件）
-- -----------------------------------------------------------------------------
INSERT INTO chief_complaint_types (id, clinic_id, name, is_active, sort_order) VALUES
    (7,  4, '食欲不振',       true, 1),
    (8,  4, '嘔吐・下痢',     true, 2),
    (9,  4, '皮膚・被毛異常', true, 3),
    (10, 4, '排尿・排泄異常', true, 4)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('chief_complaint_types', 'id'), (SELECT MAX(id) FROM chief_complaint_types));

-- -----------------------------------------------------------------------------
-- 城東医院 merchandise_items（物販: 4件）
-- -----------------------------------------------------------------------------
INSERT INTO merchandise_items (id, clinic_id, name, category, unit_price, tax_rate, sort_order) VALUES
    (8,  4, 'ロイヤルカナン 消化器サポート 1kg', 'food',  2900, 0.10, 1),
    (9,  4, 'ヒルズ k/d 2kg',                   'food',  3600, 0.10, 2),
    (10, 4, 'ペット用歯ブラシセット',            'goods', 1250, 0.10, 3),
    (11, 4, '文書料',                            'other', 3000, 0.10, 4)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('merchandise_items', 'id'), (SELECT MAX(id) FROM merchandise_items));

-- =============================================================================
-- 城東医院 (clinic_id=4) デモデータ（飼主・ペット）
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 城東医院 owners（飼主: 8件、ID 23〜30）
-- -----------------------------------------------------------------------------
INSERT INTO owners (id, clinic_id, name, name_kana, birth_date, company, postal_code, address1, address2, phone, company_phone, email, remarks, is_dangerous, discount_rate, membership_type) VALUES
    (23, 4, '大野 健司',   'オオノ ケンジ',   '1979-06-10', '',           '136-0071', '東京都江東区亀戸3-5-8',         '', '090-6601-2233', '', 'kenji.ono@example.com',      '定期通院',   false, 10, 'member'),
    (24, 4, '松田 香織',   'マツダ カオリ',   '1988-02-14', '',           '135-0044', '東京都江東区越中島2-1-4',       '', '080-7702-3344', '', 'kaori.matsuda@example.com',  '',           false, 0,  'non_member'),
    (25, 4, '渡辺 直樹',   'ワタナベ ナオキ', '1972-10-28', '渡辺商事',   '132-0025', '東京都江戸川区松江3-7-2',       '', '090-8803-4455', '', 'naoki.watanabe@example.com', '',           false, 5,  'member'),
    (26, 4, '中島 奈緒',   'ナカジマ ナオ',   '1994-04-17', '',           '133-0065', '東京都江戸川区南篠崎町1-6-3',   '', '080-9904-5566', '', 'nao.nakajima@example.com',   '',           false, 0,  'non_member'),
    (27, 4, '斎藤 浩二',   'サイトウ コウジ', '1967-12-05', '',           '131-0031', '東京都墨田区墨田1-4-9',         '', '090-1105-6677', '', 'koji.saito@example.com',     '',           false, 0,  'non_member'),
    (28, 4, '坂本 真由美', 'サカモト マユミ', '1991-08-22', '',           '130-0022', '東京都墨田区横川4-2-7',         '', '080-2206-7788', '', 'mayumi.sakamoto@example.com','猫アレルギーに注意', false, 0,  'member'),
    (29, 4, '岡田 俊雄',   'オカダ トシオ',   '1983-03-30', 'オカダ工業', '132-0034', '東京都江戸川区小松川1-8-5',     '', '090-3307-8899', '', 'toshio.okada@example.com',   '',           false, 0,  'non_member'),
    (30, 4, '藤田 彩',     'フジタ アヤ',     '1997-11-11', '',           '136-0076', '東京都江東区南砂3-9-1',         '', '080-4408-9900', '', 'aya.fujita@example.com',     '',           false, 0,  'non_member')
ON CONFLICT (id) DO UPDATE SET
    name      = EXCLUDED.name,
    name_kana = EXCLUDED.name_kana,
    updated_at      = now();

SELECT setval(pg_get_serial_sequence('owners', 'id'), (SELECT MAX(id) FROM owners));

-- -----------------------------------------------------------------------------
-- 城東医院 pets（ペット: 10件、ID 29〜38）
-- -----------------------------------------------------------------------------
INSERT INTO pets (id, clinic_id, owner_id, pet_number, name, name_kana, animal_species_id, gender, status, birth_date, breed, color, weight, insurance_id, last_visit) VALUES
    (29, 4, 23, '23-1', 'クロ',   'クロ',   1, 'male',   'alive', '2019-03-20', 'ラブラドール',             'ブラック',   28.0,  6,    '2025-11-10'),
    (30, 4, 23, '23-2', 'シナモン', 'シナモン', 2, 'female','alive', '2021-07-15', 'アビシニアン',           'レッド',      4.1,  NULL, NULL),
    (31, 4, 24, '24-1', 'ポポ',   'ポポ',   2, 'female', 'alive', '2020-05-10', 'ロシアンブルー',           'グレー',      3.8,  7,    '2025-10-20'),
    (32, 4, 25, '25-1', 'ダン',   'ダン',   1, 'male',   'alive', '2017-09-05', 'ウェルシュコーギー',       'セーブルホワイト', 13.2, NULL, NULL),
    (33, 4, 26, '26-1', 'キナ',   'キナ',   2, 'male',   'alive', '2022-02-28', 'ミックス猫',               'オレンジ',    4.5,  NULL, NULL),
    (34, 4, 27, '27-1', 'バロン', 'バロン', 1, 'male',   'alive', '2015-11-18', 'シェパード',               'ブラックタン',30.5, NULL, '2025-12-01'),
    (35, 4, 28, '28-1', 'ユズ',   'ユズ',   2, 'female', 'alive', '2023-03-03', 'ミックス猫',               'キャリコ',    3.0,  NULL, NULL),
    (36, 4, 28, '28-2', 'レン',   'レン',   1, 'male',   'alive', '2021-06-14', 'ビーグル',                 'トライカラー',12.8, NULL, NULL),
    (37, 4, 29, '29-1', 'ナナ',   'ナナ',   2, 'female', 'alive', '2018-08-07', 'メインクーン',             'ブラウンタビー', 6.2, 6,   '2026-01-15'),
    (38, 4, 30, '30-1', 'コウ',   'コウ',   1, 'male',   'alive', '2020-12-25', 'トイプードル',             'アプリコット', 3.5, NULL, NULL)
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();

SELECT setval(pg_get_serial_sequence('pets', 'id'), (SELECT MAX(id) FROM pets));

-- =============================================================================
-- 敷島医院 (clinic_id=5) マスタデータ
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 敷島医院 reservation_categories（サービス種別: 14件、ID 45-58）
-- -----------------------------------------------------------------------------

-- 敷島医院 公開コース
INSERT INTO reservation_categories (id, clinic_id, name, short_name, is_active, description, color, sort_order, duration_minutes, reservation_visible, reservation_comment, is_internal) VALUES
    (45, 5, '一般診察',               '診察',     true, '内科・外科・皮膚科などの一般的な診察',         '#3B82F6', 1,  15, true,  '', false),
    (46, 5, '一般診察(再診)',          '再診',     true, '継続通院の一般診察',                           '#3B82F6', 2,  15, true,  '', false),
    (47, 5, 'ワクチン接種',            'ワクチン', true, '各種ワクチン接種（予防接種）',                 '#10B981', 3,  15, true,  '', false),
    (48, 5, '狂犬病',                 '狂犬病',   true, '狂犬病予防法に基づくワクチン接種',             '#10B981', 4,  15, true,  '', false),
    (49, 5, 'フィラリア予防',          'フィラリア', true, 'フィラリア予防薬投与・処方',                '#10B981', 5,  15, true,  '', false),
    (50, 5, '健康診断',               '健診',     true, '定期健康診断・フィラリア検査など',             '#8B5CF6', 6,  15, true,  '', false),
    (51, 5, '健康診断結果報告',        '結果報告', true, '健康診断結果の説明・報告',                     '#8B5CF6', 7,  15, true,  '', false),
    (52, 5, 'トリミングコース',        'トリミング', true, 'カット・シャンプー・ブロー・爪切り・耳掃除', '#F59E0B', 8,  15, true,  '', false)
ON CONFLICT DO NOTHING;

-- 敷島医院 スタッフ専用コース
INSERT INTO reservation_categories (id, clinic_id, name, short_name, is_active, description, color, sort_order, duration_minutes, reservation_visible, reservation_comment, is_internal) VALUES
    (53, 5, '手術60',                 '手術60',   true, '手術枠（60分）',                               '#EF4444', 9,  60, false, '', true),
    (54, 5, '手術30',                 '手術30',   true, '手術枠（30分）',                               '#EF4444', 10, 30, false, '', true),
    (55, 5, '休憩枠',                 '休憩',     true, '休憩・ブロック枠',                             '#6B7280', 11, 60, false, '', true),
    (56, 5, '×',                     '×',       true, '予約不可（15分）',                             '#6B7280', 12, 15, false, '', true),
    (57, 5, '予約不可60',             '不可60',   true, '予約不可ブロック（60分）',                     '#6B7280', 13, 60, false, '', true),
    (58, 5, '予約不可30',             '不可30',   true, '予約不可ブロック（30分）',                     '#6B7280', 14, 30, false, '', true)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('reservation_categories', 'id'), (SELECT MAX(id) FROM reservation_categories));

-- 敷島医院 (clinic_id=5) グループ紐付け
UPDATE reservation_categories SET group_id=15 WHERE clinic_id=5 AND id IN (45,46);
UPDATE reservation_categories SET group_id=16 WHERE clinic_id=5 AND id IN (47,48,49);
UPDATE reservation_categories SET group_id=17 WHERE clinic_id=5 AND id IN (50,51);
UPDATE reservation_categories SET group_id=18 WHERE clinic_id=5 AND id IN (52);
UPDATE reservation_categories SET group_id=19 WHERE clinic_id=5 AND id IN (53,54);
UPDATE reservation_categories SET group_id=20, is_internal=true WHERE clinic_id=5 AND id IN (55,56,57,58);

-- -----------------------------------------------------------------------------
-- 敷島医院 cages（ケージ: 4件）
-- -----------------------------------------------------------------------------
INSERT INTO cages (id, clinic_id, name, price, is_active, description, cage_type, cage_size, sort_order) VALUES
    (13, 5, 'ICUケージ',       8000, true, '酸素吸入可・重症患者用',     'icu',     'medium', 1),
    (14, 5, '犬用ケージ（小）', 2900, true, '小型犬・ホテル利用可',       'dog',     'small',  2),
    (15, 5, '犬用ケージ（大）', 4200, true, '大型犬・術後管理用',         'dog',     'large',  3),
    (16, 5, '猫用ケージ',       3100, true, '猫専用・ストレス軽減設計',   'cat',     'medium', 4)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('cages', 'id'), (SELECT MAX(id) FROM cages));

-- -----------------------------------------------------------------------------
-- 敷島医院 insurances（保険: 3件）
-- -----------------------------------------------------------------------------
INSERT INTO insurances (id, clinic_id, name, is_active, description, coverage_rate, contact_phone, sort_order) VALUES
    (9,  5, 'アニコム損保',         true, 'ペット保険大手・どうぶつ健保シリーズ', 70, '0120-025-034', 1),
    (10, 5, 'ペット&ファミリー',     true, 'げんきナンバーワンシリーズ',           80, '0120-81-8505', 2),
    (11, 5, 'その他（自費）',       true, '保険未加入・全額自費',                100, '',             3)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('insurances', 'id'), (SELECT MAX(id) FROM insurances));

-- -----------------------------------------------------------------------------
-- 敷島医院 exam_types（検査種別: 3件）
-- -----------------------------------------------------------------------------
INSERT INTO exam_types (id, clinic_id, name, price, is_active, description, sort_order) VALUES
    (9,  5, '血液検査（CBC）', 2800, true, '全血球計算（Complete Blood Count）', 1),
    (10, 5, '尿検査',           1600, true, '尿試験紙・尿沈渣検査',               2),
    (11, 5, '超音波検査（エコー）', 5200, true, '腹部エコー・心臓エコー',         3)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('exam_types', 'id'), (SELECT MAX(id) FROM exam_types));

-- exam_type_fields for clinic 5
INSERT INTO exam_type_fields (id, exam_type_id, name, inspection_value, normal_value, sort_order) VALUES
    (26, 9,  'WBC（白血球数）',        '', '6.0-17.0 x10^3/uL', 1),
    (27, 9,  'RBC（赤血球数）',        '', '5.5-8.5 x10^6/uL',  2),
    (28, 9,  'PLT（血小板数）',        '', '175-500 x10^3/uL',  3),
    (29, 10, '尿比重',                 '', '1.015-1.045',        1),
    (30, 10, '尿pH',                   '', '5.5-7.5',            2),
    (31, 10, '尿タンパク',             '', '陰性',               3),
    (32, 11, '腹部エコー',             '', '異常なし',           1),
    (33, 11, '心臓エコー',             '', '異常なし',           2)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('exam_type_fields', 'id'), (SELECT MAX(id) FROM exam_type_fields));

-- -----------------------------------------------------------------------------
-- 敷島医院 vaccines（ワクチン: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO vaccines (id, clinic_id, name, price, is_active, description, species, interval, sort_order) VALUES
    (16, 5, '混合ワクチン5種（犬）',    4600, true, 'ジステンパー・パルボ・アデノ1型・アデノ2型・パラインフルエンザ', 'dog', '1年',   1),
    (17, 5, '混合ワクチン10種（犬）',   8200, true, '5種＋レプトスピラ5種',                                          'dog', '1年',   2),
    (18, 5, '混合ワクチン3種（猫）',    4100, true, '猫ウイルス性鼻気管炎・カリシウイルス・汎白血球減少症',           'cat', '1年',   3),
    (19, 5, '混合ワクチン5種（猫）',    5600, true, '3種＋猫白血病・猫クラミジア',                                   'cat', '1年',   4),
    (20, 5, '狂犬病ワクチン',           3000, true, '狂犬病予防法に基づく接種',                                      'dog', '1年',   5)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('vaccines', 'id'), (SELECT MAX(id) FROM vaccines));

-- -----------------------------------------------------------------------------
-- 敷島医院 medicines（薬剤: カテゴリ4件 + 薬剤10件）
-- -----------------------------------------------------------------------------
INSERT INTO medicines (id, clinic_id, name, price, is_active, description, sort_order) VALUES
    (3001, 5, '抗生剤',     NULL, true, '抗生物質カテゴリ',   1),
    (3002, 5, 'ステロイド', NULL, true, 'ステロイド剤カテゴリ', 2),
    (3003, 5, '制吐剤',     NULL, true, '制吐剤カテゴリ',     3),
    (3004, 5, '消化器用薬', NULL, true, '消化器用薬カテゴリ', 4)
ON CONFLICT DO NOTHING;

INSERT INTO medicines (id, clinic_id, name, price, is_active, description, dosage_form, medicine_unit, default_quantity, sort_order, parent_id) VALUES
    (201, 5, 'アモキシシリン 50mg',         510,  true, '広域スペクトラム抗生物質',              'tablet',    'per_tablet', 1,    1, 3001),
    (202, 5, 'エンロフロキサシン 15mg',      580,  true, 'フルオロキノロン系抗生物質',            'tablet',    'per_tablet', 1,    2, 3001),
    (203, 5, 'メトロニダゾール 250mg',       610,  true, '嫌気性菌・原虫感染症治療薬',            'tablet',    'per_tablet', 1,    3, 3001),
    (204, 5, 'プレドニゾロン 5mg',           410,  true, 'ステロイド系抗炎症・免疫抑制剤',        'tablet',    'per_tablet', 1,    4, 3002),
    (205, 5, 'デキサメタゾン注射液',         710,  true, '強力ステロイド・アレルギー緊急治療',    'injection', 'per_ml',     1,    5, 3002),
    (206, 5, 'マロピタント 16mg',            820,  true, '制吐剤（乗り物酔い・嘔吐治療）',        'tablet',    'per_tablet', 1,    6, 3003),
    (207, 5, 'メトクロプラミド注',           480,  true, '制吐・消化管運動促進',                  'injection', 'per_ml',     1,    7, 3003),
    (208, 5, 'ラクツロース液',               510,  true, '便秘・肝性脳症の治療',                  'liquid',    'per_ml',     5,    8, 3004),
    (209, 5, 'オメプラゾール 10mg',          360,  true, 'プロトンポンプ阻害薬（胃酸抑制）',      'tablet',    'per_tablet', 1,    9, 3004),
    (210, 5, '生理食塩水 500ml',             410,  true, '点滴・洗浄用生理食塩水',                'liquid',    'per_ml',     500,  10, NULL)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('medicines', 'id'), (SELECT MAX(id) FROM medicines));

-- -----------------------------------------------------------------------------
-- 敷島医院 consultations（診察項目: 3件）
-- -----------------------------------------------------------------------------
INSERT INTO consultations (id, clinic_id, name, price, is_active, description, time_condition, duration, sort_order) VALUES
    (9,  5, '初診料',       1900, true, '初めての受診または6ヶ月以上受診がない場合', 'first_visit', 30, 1),
    (10, 5, '再診料',        800, true, '継続通院の診察料',                         'revisit',     15, 2),
    (11, 5, '往診料',       5500, true, '自宅への往診料（基本料金）',               'anytime',     60, 3)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('consultations', 'id'), (SELECT MAX(id) FROM consultations));

-- -----------------------------------------------------------------------------
-- 敷島医院 procedures（処置項目: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO procedures (id, clinic_id, name, price, is_active, description, duration, anesthesia, sort_order) VALUES
    (16, 5, '去勢手術（犬）', 24000, true, '雄犬の去勢手術',                  60, 'general', 1),
    (17, 5, '避妊手術（猫）', 24000, true, '雌猫の避妊手術',                  90, 'general', 2),
    (18, 5, '歯石除去',       15000, true, '全身麻酔下での歯石除去・歯周治療', 45, 'general', 3),
    (19, 5, '爪切り',            500, true, '爪のカット・やすりがけ',           10, 'none',    4),
    (20, 5, '腫瘍摘出',        22000, true, '皮膚腫瘍の外科的摘出',             60, 'local',   5)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('procedures', 'id'), (SELECT MAX(id) FROM procedures));

-- -----------------------------------------------------------------------------
-- 敷島医院 hospitalization_plans（入院プラン: 3件）
-- -----------------------------------------------------------------------------
INSERT INTO hospitalization_plans (id, clinic_id, name, price, is_active, description, body_size, billing_unit, sort_order) VALUES
    (9,  5, '一般入院（小型）', 3100, true, '体重10kg以下の入院管理料（1日）',  'small',  'per_day',   1),
    (10, 5, '一般入院（大型）', 4800, true, '体重25kg以上の入院管理料（1日）',  'large',  'per_day',   2),
    (11, 5, 'ホテル（小型）',   2600, true, '体重10kg以下のペットホテル（1泊）', 'small',  'per_night', 3)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('hospitalization_plans', 'id'), (SELECT MAX(id) FROM hospitalization_plans));

-- -----------------------------------------------------------------------------
-- 敷島医院 trimming_courses（トリミングコース: 3件）
-- -----------------------------------------------------------------------------
INSERT INTO trimming_courses (id, clinic_id, name, price, is_active, description, target_size, duration, sort_order) VALUES
    (9,  5, 'シャンプー&ブロー（小型）', 3900, true, 'シャンプー・ブロー・ブラッシング',            'small',  60,  1),
    (10, 5, 'シャンプー&ブロー（中型）', 5400, true, 'シャンプー・ブロー・ブラッシング',            'medium', 90,  2),
    (11, 5, 'フルコース（小型）',        6800, true, 'カット・シャンプー・ブロー・爪切り・耳掃除', 'small',  120, 3)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('trimming_courses', 'id'), (SELECT MAX(id) FROM trimming_courses));

-- -----------------------------------------------------------------------------
-- 敷島医院 trimming_options（トリミングオプション: 3件）
-- -----------------------------------------------------------------------------
INSERT INTO trimming_options (id, clinic_id, name, price, is_active, description, duration, is_combinable, sort_order) VALUES
    (9,  5, '爪切り',   300, true, '爪のカット・やすりがけ',    10, true, 1),
    (10, 5, '耳掃除',   500, true, '外耳道の洗浄・清掃',        10, true, 2),
    (11, 5, '歯磨き',   500, true, '歯ブラシによるデンタルケア', 15, true, 3)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('trimming_options', 'id'), (SELECT MAX(id) FROM trimming_options));

-- -----------------------------------------------------------------------------
-- 敷島医院 diagnosis_types（診断カテゴリ: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO diagnosis_types (id, clinic_id, name, is_active, description, sort_order) VALUES
    (14, 5, '消化器系',       true, '胃腸・肝臓・膵臓などの消化器系疾患', 1),
    (15, 5, '皮膚・被毛',     true, 'アレルギー・感染症などの皮膚疾患',   2),
    (16, 5, '泌尿器系',       true, '腎臓・膀胱・尿道などの泌尿器系疾患', 3),
    (17, 5, '腫瘍',           true, '良性・悪性腫瘍（がん）',             4),
    (18, 5, '外傷・骨格',     true, '骨折・咬傷・関節疾患など',           5)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('diagnosis_types', 'id'), (SELECT MAX(id) FROM diagnosis_types));

-- -----------------------------------------------------------------------------
-- 敷島医院 diagnosis_names（診断名: 10件）
-- -----------------------------------------------------------------------------
INSERT INTO diagnosis_names (id, clinic_id, name, is_active, description, diagnosis_category_id, sort_order) VALUES
    (31, 5, '胃腸炎',             true, '胃・腸の炎症（嘔吐・下痢）',         14, 1),
    (32, 5, '肝疾患',             true, '肝炎・肝不全・脂肪肝など',           14, 2),
    (33, 5, 'アトピー性皮膚炎',   true, 'アレルゲンによるアレルギー性皮膚炎', 15, 1),
    (34, 5, '膿皮症',             true, '細菌性の皮膚感染症',                 15, 2),
    (35, 5, '真菌症',             true, '皮膚糸状菌による感染症',             15, 3),
    (36, 5, '膀胱炎',             true, '細菌性・特発性膀胱炎',               16, 1),
    (37, 5, '尿路結石',           true, '腎結石・膀胱結石・尿道結石',         16, 2),
    (38, 5, '肥満細胞腫',         true, '皮膚または内臓の肥満細胞腫瘍',       17, 1),
    (39, 5, '骨折',               true, '各部位の骨折',                       18, 1),
    (40, 5, '咬傷',               true, '他動物による咬傷・咬傷感染',         18, 2)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('diagnosis_names', 'id'), (SELECT MAX(id) FROM diagnosis_names));

-- -----------------------------------------------------------------------------
-- 敷島医院 checkup_types（健診種別: 3件）
-- -----------------------------------------------------------------------------
INSERT INTO checkup_types (id, clinic_id, name, price, is_active, description, interval, target_age, sort_order) VALUES
    (8,  5, '一般健診',       4800,  true, '身体検査・体重測定・問診',                   '1年',   '全年齢',  1),
    (9,  5, '老齢検診',       14500, true, '身体検査＋血液検査＋レントゲン＋超音波',     '6ヶ月', '7歳以上', 2),
    (10, 5, '歯科検診',        3100, true, '歯周病チェック・歯石付着度の確認',           '1年',   '成犬',    3)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('checkup_types', 'id'), (SELECT MAX(id) FROM checkup_types));

-- -----------------------------------------------------------------------------
-- 敷島医院 chief_complaint_types（主訴区分: 4件）
-- -----------------------------------------------------------------------------
INSERT INTO chief_complaint_types (id, clinic_id, name, is_active, sort_order) VALUES
    (11, 5, '食欲不振',     true, 1),
    (12, 5, '嘔吐・下痢',   true, 2),
    (13, 5, '呼吸困難',     true, 3),
    (14, 5, '外傷・骨折',   true, 4)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('chief_complaint_types', 'id'), (SELECT MAX(id) FROM chief_complaint_types));

-- -----------------------------------------------------------------------------
-- 敷島医院 merchandise_items（物販: 4件）
-- -----------------------------------------------------------------------------
INSERT INTO merchandise_items (id, clinic_id, name, category, unit_price, tax_rate, sort_order) VALUES
    (12, 5, 'ロイヤルカナン 消化器サポート 1kg', 'food',  2750, 0.10, 1),
    (13, 5, 'ノミ・ダニ予防首輪',               'goods', 1600, 0.10, 2),
    (14, 5, 'エリザベスカラー（M）',             'goods', 1000, 0.10, 3),
    (15, 5, '文書料',                            'other', 3000, 0.10, 4)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('merchandise_items', 'id'), (SELECT MAX(id) FROM merchandise_items));

-- =============================================================================
-- 敷島医院 (clinic_id=5) デモデータ（飼主・ペット）
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 敷島医院 owners（飼主: 8件、ID 31〜38）
-- -----------------------------------------------------------------------------
INSERT INTO owners (id, clinic_id, name, name_kana, birth_date, company, postal_code, address1, address2, phone, company_phone, email, remarks, is_dangerous, discount_rate, membership_type) VALUES
    (31, 5, '村上 俊平',   'ムラカミ シュンペイ', '1976-04-12', '',           '400-0031', '山梨県甲府市丸の内2-3-4',     '', '090-5501-1122', '', 'shunpei.murakami@example.com', '',           false, 10, 'member'),
    (32, 5, '長谷川 恵子', 'ハセガワ ケイコ',     '1989-09-07', '',           '400-0032', '山梨県甲府市中央3-5-6',       '', '080-6602-2233', '', 'keiko.hasegawa@example.com',   '',           false, 0,  'non_member'),
    (33, 5, '野口 正樹',   'ノグチ マサキ',       '1971-01-25', '野口設計',   '400-0035', '山梨県甲府市丸の内3-7-8',     '', '090-7703-3344', '', 'masaki.noguchi@example.com',   '',           false, 5,  'member'),
    (34, 5, '石田 沙織',   'イシダ サオリ',       '1993-07-18', '',           '400-0801', '山梨県甲府市横根町5-2-1',     '', '080-8804-4455', '', 'saori.ishida@example.com',     '猫2匹飼い', false, 0,  'non_member'),
    (35, 5, '前田 修',     'マエダ オサム',       '1968-11-02', '',           '400-0031', '山梨県甲府市丸の内1-9-3',     '', '090-9905-5566', '', 'osamu.maeda@example.com',      '',           false, 0,  'non_member'),
    (36, 5, '菊池 里奈',   'キクチ リナ',         '1995-06-14', '',           '400-0032', '山梨県甲府市中央5-4-7',       '', '080-1106-6677', '', 'rina.kikuchi@example.com',     '',           false, 0,  'member'),
    (37, 5, '清水 和彦',   'シミズ カズヒコ',     '1980-03-28', '清水工務店', '400-0034', '山梨県甲府市北口1-6-2',       '', '090-2207-7788', '', 'kazuhiko.shimizu@example.com', '',           false, 0,  'non_member'),
    (38, 5, '岩崎 美穂',   'イワサキ ミホ',       '1998-12-05', '',           '400-0803', '山梨県甲府市横根町2-8-9',     '', '080-3308-8899', '', 'miho.iwasaki@example.com',     '',           false, 0,  'non_member')
ON CONFLICT (id) DO UPDATE SET
    name      = EXCLUDED.name,
    name_kana = EXCLUDED.name_kana,
    updated_at      = now();

SELECT setval(pg_get_serial_sequence('owners', 'id'), (SELECT MAX(id) FROM owners));

-- -----------------------------------------------------------------------------
-- 敷島医院 pets（ペット: 10件、ID 39〜48）
-- -----------------------------------------------------------------------------
INSERT INTO pets (id, clinic_id, owner_id, pet_number, name, name_kana, animal_species_id, gender, status, birth_date, breed, color, weight, insurance_id, last_visit) VALUES
    (39, 5, 31, '31-1', 'フク',   'フク',   1, 'male',   'alive', '2018-05-05', 'シベリアンハスキー',       'グレーホワイト', 24.0, 9,    '2025-10-05'),
    (40, 5, 32, '32-1', 'アズキ', 'アズキ', 2, 'female', 'alive', '2021-11-20', 'ノルウェージャンフォレストキャット', 'ブラウンタビー', 4.8, NULL, '2025-12-18'),
    (41, 5, 33, '33-1', 'カイ',   'カイ',   1, 'male',   'alive', '2016-07-10', 'ゴールデンレトリーバー',   'ゴールデン',     31.2, NULL, '2026-01-08'),
    (42, 5, 34, '34-1', 'キキ',   'キキ',   2, 'female', 'alive', '2022-04-25', 'ミックス猫',               'トラ',           3.5,  NULL, NULL),
    (43, 5, 34, '34-2', 'ニコ',   'ニコ',   2, 'female', 'alive', '2020-09-12', 'ミックス猫',               'サビ',           4.2,  NULL, NULL),
    (44, 5, 35, '35-1', 'ジャック', 'ジャック', 1, 'male', 'alive', '2014-02-28', 'ジャックラッセルテリア', 'ホワイトタン',   6.8,  NULL, '2025-09-20'),
    (45, 5, 36, '36-1', 'ミル',   'ミル',   2, 'female', 'alive', '2023-01-15', 'ミックス猫',               'ホワイト',       3.2,  NULL, NULL),
    (46, 5, 37, '37-1', 'ガイア', 'ガイア', 1, 'male',   'alive', '2017-08-30', 'アラスカンマラミュート',   'グレー',         36.5, NULL, '2025-11-25'),
    (47, 5, 38, '38-1', 'ハナ',   'ハナ',   2, 'female', 'alive', '2021-03-08', 'アメリカンショートヘア',   'シルバータビー', 4.0,  9,    '2026-02-10'),
    (48, 5, 38, '38-2', 'ソウ',   'ソウ',   1, 'male',   'alive', '2019-10-01', 'ミックス犬',               'ベージュ',       9.2,  NULL, NULL)
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();

SELECT setval(pg_get_serial_sequence('pets', 'id'), (SELECT MAX(id) FROM pets));

-- -----------------------------------------------------------------------------
-- ワクチン接種記録（八王子院: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO vaccinations (id, clinic_id, medical_record_id, pet_id, vaccine_id, doctor_id, date, lot1, next_schedule_type, next_date, remarks) VALUES
    (1, 3, 1,  1,  1,  1, '2025-10-10', 'LOT-2025-A001', '1year',  '2026-10-10', '5種混合ワクチン接種。体調良好。'),
    (2, 3, 5,  3,  5,  1, '2025-09-15', 'LOT-2025-C001', '1year',  '2026-09-15', '3種混合ワクチン接種。'),
    (3, 3, 6,  3,  5,  2, '2026-01-06', 'LOT-2026-C001', '1year',  '2027-01-06', '3種混合ワクチン追加接種。'),
    (4, 3, 15, 13, 5,  1, '2026-01-06', 'LOT-2026-C003', '4weeks', '2026-02-03', '初回3種混合（猫）。4週後に2回目接種予定。'),
    (5, 3, 20, 18, 5,  1, '2026-01-06', 'LOT-2026-C002', '1year',  '2027-01-06', '3種混合ワクチン接種。体調良好。')
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('vaccinations', 'id'), (SELECT MAX(id) FROM vaccinations));

-- -----------------------------------------------------------------------------
-- 検査記録（八王子院: 3件）
-- -----------------------------------------------------------------------------
INSERT INTO exams (id, clinic_id, medical_record_id, exam_type_id, doctor_id, date, result_summary, status) VALUES
    (1, 3, 2,  1, 1, '2025-12-15', 'CBC全項目正常範囲内。', 'completed'),
    (2, 3, 14, 2, 2, '2025-06-15', 'ALT軽度上昇（145 U/L）。他正常。', 'completed'),
    (3, 3, 13, 1, 1, '2026-02-10', 'WBC上昇（19.2）。脱水を反映。', 'result_entered')
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('exams', 'id'), (SELECT MAX(id) FROM exams));

INSERT INTO exam_results (id, exam_id, exam_type_item_id, inspection_value, status) VALUES
    -- Exam 1 (CBC for MR2)
    (34, 1, 1, '9.8',  'normal'),
    (35, 1, 2, '7.2',  'normal'),
    (36, 1, 3, '45',   'normal'),
    (37, 1, 4, '320',  'normal'),
    -- Exam 2 (Chemistry for MR14)
    (38, 2, 5, '145',  'high'),
    (39, 2, 6, '18',   'normal'),
    (40, 2, 7, '1.2',  'normal'),
    (41, 2, 8, '98',   'normal'),
    -- Exam 3 (CBC for MR13)
    (42, 3, 1, '19.2', 'high'),
    (43, 3, 2, '8.1',  'normal'),
    (44, 3, 3, '52',   'normal'),
    (45, 3, 4, '280',  'normal')
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('exam_results', 'id'), (SELECT MAX(id) FROM exam_results));

-- -----------------------------------------------------------------------------
-- 城東医院 inventory_items（在庫管理: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO inventory_items (id, clinic_id, name, category, quantity, unit, min_stock_level, location, supplier, status) VALUES
    (15, 4, '5種混合ワクチン',               'medicine',   20,  'バイアル', 10, '冷蔵庫',     '共立製薬',               'sufficient'),
    (16, 4, 'アモキシシリン 50mg',           'medicine',   100, '錠',       30, '薬品棚A',    '日本全薬工業',           'sufficient'),
    (17, 4, 'ロイヤルカナン 消化器サポート', 'food',       12,  '袋',        5, 'フード棚',   'ロイヤルカナン',         'sufficient'),
    (18, 4, '包帯・ガーゼセット',            'consumable', 40,  'セット',   20, '処置室',     '白十字',                 'sufficient'),
    (19, 4, 'ノミダニ駆除薬 スポット（犬用）','medicine',  25,  'ピペット',  10, '薬品棚B',   'エランコジャパン',       'sufficient')
ON CONFLICT (id) DO NOTHING;

-- -----------------------------------------------------------------------------
-- 敷島医院 inventory_items（在庫管理: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO inventory_items (id, clinic_id, name, category, quantity, unit, min_stock_level, location, supplier, status) VALUES
    (20, 5, '5種混合ワクチン（犬）',         'medicine',   15,  'バイアル', 10, '冷蔵庫',     '共立製薬',               'sufficient'),
    (21, 5, 'アモキシシリン 50mg',           'medicine',   80,  '錠',       30, '薬品棚A',    '日本全薬工業',           'sufficient'),
    (22, 5, 'ヒルズ k/d',                   'food',        8,   '袋',        5, 'フード棚',   'ヒルズ・コルゲート',     'sufficient'),
    (23, 5, 'シリンジ 5mL',                 'consumable', 200,  '本',       50, '処置室',     'テルモ',                 'sufficient'),
    (24, 5, '生理食塩水 500ml',             'medicine',    30,  '本',       10, '薬品棚B',    '大塚製薬工場',           'sufficient')
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('inventory_items', 'id'), (SELECT MAX(id) FROM inventory_items));

-- -----------------------------------------------------------------------------
-- 14. audit_logs（監査ログ: 8件）
-- -----------------------------------------------------------------------------
INSERT INTO audit_logs (clinic_id, actor_id, actor_type, action, resource, resource_id, old_value, new_value, ip_address, user_agent) VALUES
    (3, 10, 'staff', 'permission_rules.update', 'permission_groups', 1, '{"can_delete": false}', '{"can_delete": true}', '192.168.1.1', 'Mozilla/5.0...'),
    (3, 10, 'staff', 'auth.login.success', 'user_accounts', 10, NULL, NULL, '192.168.1.1', 'Mozilla/5.0...'),
    (3, 11, 'staff', 'owner.update', 'owners', 1, '{"phone": "old"}', '{"phone": "090-1234-5678"}', '192.168.1.2', 'Mozilla/5.0...'),
    (3, 10, 'staff', 'pet.create', 'pets', 28, NULL, '{"name": "マル"}', '192.168.1.1', 'Mozilla/5.0...'),
    (3, 10, 'staff', 'medical_record.finalize', 'medical_records', 1, '{"status": "draft"}', '{"status": "finalized"}', '192.168.1.1', 'Mozilla/5.0...'),
    (3, 11, 'staff', 'auth.login.failure', 'user_accounts', NULL, '{"email": "exec@example.com"}', NULL, '192.168.1.5', 'PostmanRuntime/7.26.8'),
    (3, 10, 'staff', 'inventory.decrease', 'inventory_items', 1, '{"quantity": 50}', '{"quantity": 43}', '192.168.1.10', 'Mozilla/5.0...'),
    (3, 10, 'staff', 'treatment.create', 'treatments', 2, NULL, '{"content": "アモキシシリン"}', '192.168.1.10', 'Mozilla/5.0...')
ON CONFLICT DO NOTHING;

-- =============================================================================
-- LINE予約システム シードデータ
-- =============================================================================

-- A. reservation_categories の予約用カラムはすべて INSERT 時に設定済み（更新不要）
-- B. staffs の予約用カラムはすべて INSERT 時に設定済み（更新不要）

-- C. line_reservation_settings 初期データ
INSERT INTO line_reservation_settings (
    clinic_id, status, business_hours, break_hours, daily_limit,
    booking_window_max_days, booking_window_min_days, calendar_months,
    phone_number, time_slot_mode, time_slot_interval_minutes,
    no_staff_mode, show_no_staff_option, national_holiday_closed, additional_fields
)
SELECT
    c.id, 'stopped',
    '{"start":"0900","end":"1900"}',
    '[{"start":"1200","end":"1300"}]',
    1, 30, 2, 2, '0552334126', 'minimize_gaps', 15,
    'first_available', true, false,
    '[
        {"key":"phone","label":"電話番号","required":true,"placeholder":"例) 090-1234-5678"},
        {"key":"owner_name","label":"飼い主名","required":true,"placeholder":""},
        {"key":"pet_info","label":"ペットの名前と種類","required":true,"placeholder":"例) ポチ（柴犬）"},
        {"key":"symptoms","label":"診察内容","required":false,"placeholder":""}
    ]'
FROM clinics c
ON CONFLICT (clinic_id) DO NOTHING;

-- C-2. テスト環境用 LINE クレデンシャル（docs/LINE-SETUP.md 参照）
-- clinic_id=3: テスト-八王子（@642hdxoh）
-- clinic_id=4: テスト-城東（@151lnsqa）
UPDATE line_reservation_settings SET
    status              = 'running',
    header_text         = 'ノア動物病院 八王子',
    line_channel_id     = '2009755544',
    line_channel_secret = '5344ef84eb7072b5894f7e087db28827',
    liff_id             = '2009755581-w5NOA3EW',
    line_access_token   = 'pwMi3yP6jhRa0xbmnR0IPEcE5l+OIp21a7ia3hmoiuFSCvqkR5Tmmfm6fLoSTB1Bt7uQjAe9NN7fZ+LBDtNKLGnrqBrjDmhTnws9PVxQKLyinomNzUAb61KADX7NJmFBfEsLQQ9VmlU+tMJcWh+zswdB04t89/1O/w1cDnyilFU='
WHERE clinic_id = 3;

UPDATE line_reservation_settings SET
    status              = 'running',
    header_text         = 'ノア動物病院 城東',
    line_channel_id     = '2009755545',
    line_channel_secret = '25e4661a8de553953a4b34c6ac7a91cb',
    liff_id             = '2009755586-nvKfG3Cp',
    line_access_token   = 'CUAtYMry8doD9ALCF/6Y0JocVqRrxC8IbOCMyRyxwDw5EJhyujJ4lQ8mVGrt7WawTi+voAxZ79mKAg+4qlsUPBVU6VMZdk7wEA42NZXQAl8gBr2tSYmZpbRzAiLfuGhxuba+koBHVk8yTuaCCjLBzAdB04t89/1O/w1cDnyilFU='
WHERE clinic_id = 4;

-- D. staff_reservation_exclusions 初期データ
-- 獣医スタッフはトリミング系コースを非対応とする（同一クリニック内のみ）
INSERT INTO staff_reservation_exclusions (staff_id, reservation_category_id)
SELECT s.id, st.id
FROM staffs s
JOIN staff_clinic_assignments sca ON sca.staff_id = s.id
JOIN reservation_categories st ON sca.clinic_id = st.clinic_id
WHERE s.staff_type = 'doctor'
  AND st.name IN (
      'トリミングコース', 'トリミング部分カットコース', 'トリミングシャンプーコース',
      'クイックシャンプー', 'お手入れ', '室内ドッグラン'
  )
ON CONFLICT (staff_id, reservation_category_id) DO NOTHING;

-- E. line_customers テスト用 LINE 顧客データ
INSERT INTO line_customers (clinic_id, line_user_id, display_name, real_name, additional_fields, owner_id)
VALUES
    (3, 'U_test_hachioji_001', 'テスト 太郎', '執行 太郎', '{"phone":"090-1234-5678","owner_name":"執行 太郎","pet_name":"ポチ","pet_type":"柴犬"}', 1),
    (3, 'U_test_hachioji_002', 'テスト 花子', '一般 花子', '{"phone":"080-9876-5432","owner_name":"一般 花子","pet_name":"ミケ","pet_type":"三毛猫"}', 2),
    (4, 'U_test_joto_001', 'テスト 次郎', '城東テスト', '{"phone":"070-1111-2222","owner_name":"城東テスト","pet_name":"チョコ","pet_type":"トイプードル"}', NULL)
ON CONFLICT DO NOTHING;

-- F. 城東医院 (clinic_id=4) テスト予約 3件
-- ※ owners/pets が先に挿入されている必要がある
INSERT INTO appointments (id, clinic_id, start_time, end_time, owner_id, pet_id, visit_type, reservation_category_id, doctor_id, is_designated, status, notes) VALUES
    (11, 4, '2026-03-12 10:00:00+09', '2026-03-12 10:15:00+09', 23, 29, 'revisit', 26, 16, true,  'confirmed', 'クロの定期診察'),
    (12, 4, '2026-03-13 14:00:00+09', '2026-03-13 14:15:00+09', 24, 31, 'first',   28, 16, false, 'confirmed', 'ポポのワクチン接種'),
    (13, 4, '2026-03-14 11:00:00+09', '2026-03-14 11:15:00+09', 25, 32, 'revisit', 31, 16, false, 'confirmed', 'ダンの健康診断')
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();

SELECT setval(pg_get_serial_sequence('appointments', 'id'), (SELECT MAX(id) FROM appointments));
