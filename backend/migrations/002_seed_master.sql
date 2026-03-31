-- =============================================================================
-- Animal Ekarte - 統合シード v20.0
-- PostgreSQL 18
-- 冪等性保証: ON CONFLICT DO NOTHING
-- 内容: マスタデータ + デモアカウント
-- 依存: 001_init.sql
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 1. company（本部情報: 1件）
-- -----------------------------------------------------------------------------
INSERT INTO company (name) VALUES
    ('ノア動物病院')
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 2. clinics（クリニック: 3件）
-- -----------------------------------------------------------------------------
INSERT INTO clinics (id, company_id, name) VALUES
    (3, 1, '八王子院'),
    (4, 1, '城東医院'),
    (5, 1, '敷島医院')
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('clinics', 'id'), (SELECT MAX(id) FROM clinics));

-- -----------------------------------------------------------------------------
-- 3. animal_species（ペット種類: 6件、システム共通・clinic_idなし）
-- -----------------------------------------------------------------------------
INSERT INTO animal_species (id, name, is_active, sort_order) VALUES
    (1, '犬',         true, 1),
    (2, '猫',         true, 2),
    (3, '鳥',         true, 3),
    (4, 'うさぎ',     true, 4),
    (5, 'ハムスター', true, 5),
    (6, 'その他',     true, 6)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('animal_species', 'id'), (SELECT MAX(id) FROM animal_species));

-- -----------------------------------------------------------------------------
-- 4. job_titles（職種: 5件）
-- -----------------------------------------------------------------------------
INSERT INTO job_titles (id, clinic_id, name, is_active, sort_order) VALUES
    (1, 3, '獣医師',     true, 1),
    (2, 3, '動物看護師', true, 2),
    (3, 3, 'トリマー',   true, 3),
    (4, 3, '受付',       true, 4),
    (5, 3, '管理者',     true, 5)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('job_titles', 'id'), (SELECT MAX(id) FROM job_titles));

-- -----------------------------------------------------------------------------
-- 5. staffs（スタッフ: 12件）
-- -----------------------------------------------------------------------------
INSERT INTO staffs (id, clinic_id, name, is_active, staff_role, license_number, job_title_id, sort_order) VALUES
    (1,  3, '山田 太郎',   true, 'veterinarian', 'V-10001', 1, 1),
    (2,  3, '高橋 健一',   true, 'veterinarian', 'V-10002', 1, 2),
    (3,  3, '渡辺 博',     true, 'manager',      '',        5, 3),
    (4,  3, '佐藤 花子',   true, 'nurse',        '',        2, 4),
    (5,  3, '伊藤 さくら', true, 'nurse',        '',        2, 5),
    (6,  3, '木村 健太',   true, 'trimmer',      '',        3, 6),
    (7,  3, '田中 美咲',   true, 'reception',    '',        4, 7),
    -- デモアカウント用スタッフ（八王子院）
    (8,  3, '田中 太郎',   true, 'veterinarian', 'V-20001', 1, 1),
    (9,  3, '山田 花子',   true, 'veterinarian', 'V-20002', 1, 2),
    (10, 3, '佐藤 美咲',   true, 'nurse',        '',        2, 3),
    (11, 3, '鈴木 一郎',   true, 'reception',    '',        4, 4),
    (12, 3, '高橋 さくら', true, 'trimmer',      '',        3, 5),
    -- 管理者・執行グループ デモアカウント用スタッフ
    (13, 3, '渡辺 院長',   true, 'manager',      '',        5, 6),
    (14, 3, '小林 部長',   true, 'manager',      '',        5, 7)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('staffs', 'id'), (SELECT MAX(id) FROM staffs));

-- -----------------------------------------------------------------------------
-- 6. user_accounts（ユーザーアカウント: 9件）
-- password_hash: bcrypt("password", cost=10)
-- -----------------------------------------------------------------------------
INSERT INTO user_accounts (id, email, display_name, display_name_kana, user_type, job_title_id, status, staff_id, password_hash) VALUES
    -- 渋谷院・新宿院スタッフ
    (1, 'admin@noavet.jp',   'システム管理者', 'システムカンリシャ',   'system_admin', 5, 'active', 3,    '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6'),
    (2, 'clinic1@noavet.jp', '渋谷院管理者',   'シブヤインカンリシャ', 'clinic_admin', 5, 'active', NULL, '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6'),
    (3, 'yamada@noavet.jp',  '山田 太郎',      'ヤマダ タロウ',        'staff',        1, 'active', 1,    '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6'),
    -- デモアカウント（八王子院・frontend mock-data.ts 対応）
    (4, 'admin@example.com',     '田中 太郎',  'タナカ タロウ',    'clinic_admin', 1, 'active', 8,    '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6'),
    (5, 'vet@example.com',       '山田 花子',  'ヤマダ ハナコ',    'staff',        1, 'active', 9,    '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6'),
    (6, 'nurse@example.com',     '佐藤 美咲',  'サトウ ミサキ',    'staff',        2, 'active', 10,   '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6'),
    (7, 'reception@example.com', '鈴木 一郎',  'スズキ イチロウ',  'staff',        4, 'active', 11,   '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6'),
    (8, 'trimmer@example.com',   '高橋 さくら','タカハシ サクラ',  'staff',        3, 'active', 12,   '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6'),
    (9,  'system@example.com',   '本部 管理者', 'ホンブ カンリシャ','system_admin', 5, 'active', NULL, '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6'),
    -- 管理者・執行グループ デモアカウント
    (10, 'manager@example.com',  '渡辺 院長',  'ワタナベ インチョウ','staff',        5, 'active', 13, '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6'),
    (11, 'exec@example.com',     '小林 部長',  'コバヤシ ブチョウ', 'staff',        5, 'active', 14, '$2a$10$jr4KmlfkPGeu2FXPA0jPtOLbCpdHAf3PUGMkI2ZVtWb6pKNYjWyQ6')
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('user_accounts', 'id'), (SELECT MAX(id) FROM user_accounts));

-- -----------------------------------------------------------------------------
-- 7. user_clinic_memberships（ユーザー所属クリニック: 8件）
-- -----------------------------------------------------------------------------
INSERT INTO user_clinic_memberships (id, user_id, clinic_id, is_main) VALUES
    -- デモアカウント（system=本部管理者: 全3院、他: 八王子院のみ）
    (5,  4, 3, true),
    (6,  5, 3, true),
    (7,  6, 3, true),
    (8,  7, 3, true),
    (9,  8, 3, true),
    (10, 9, 3, true),
    (11, 9, 4, false),
    (12, 9, 5, false),
    -- 管理者・執行グループ ユーザー（八王子院）
    (13, 10, 3, true),
    (14, 11, 3, true)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('user_clinic_memberships', 'id'), (SELECT MAX(id) FROM user_clinic_memberships));

-- -----------------------------------------------------------------------------
-- 7b. permission_groups（権限グループ）& user_permission_groups（割当）
-- -----------------------------------------------------------------------------
-- ノア動物病院 (company_id=1): 管理者・執行・一般 の3グループ
INSERT INTO permission_groups (id, company_id, name, description, color) VALUES
    (1, 1, '管理者', '全機能フルアクセス・権限設定管理', '#EF4444'),
    (2, 1, '執行',   '業務全般閲覧・権限設定変更',       '#6366F1'),
    (3, 1, '一般',   '基本的な業務操作',                 '#10B981')
ON CONFLICT DO NOTHING;

-- グループルール（管理者: 全リソースフルアクセス）
INSERT INTO permission_group_rules (group_id, resource, can_view, can_create, can_edit, can_delete) VALUES
    (1, 'dashboard',        true, false, false, false),
    (1, 'owners',           true, true,  true,  true),
    (1, 'reservations',     true, true,  true,  true),
    (1, 'medical-records',  true, true,  true,  true),
    (1, 'hospitalization',  true, true,  true,  true),
    (1, 'trimming',         true, true,  true,  true),
    (1, 'examinations',     true, true,  true,  true),
    (1, 'accounting',       true, true,  true,  true),
    (1, 'vaccinations',     true, true,  true,  true),
    (1, 'checkups',         true, true,  true,  true),
    (1, 'inventory',        true, true,  true,  true),
    (1, 'estimates',        true, true,  true,  true),
    (1, 'shifts',           true, true,  true,  true),
    (1, 'master',           true, true,  true,  true),
    (1, 'hospital-settings',true, true,  true,  true)
ON CONFLICT DO NOTHING;

-- グループルール（執行: 業務全般閲覧＋権限設定変更）
INSERT INTO permission_group_rules (group_id, resource, can_view, can_create, can_edit, can_delete) VALUES
    (2, 'dashboard',        true, false, false, false),
    (2, 'owners',           true, true,  true,  false),
    (2, 'reservations',     true, true,  true,  false),
    (2, 'medical-records',  true, false, false, false),
    (2, 'hospitalization',  true, true,  true,  false),
    (2, 'trimming',         true, false, false, false),
    (2, 'examinations',     true, false, false, false),
    (2, 'accounting',       true, true,  true,  false),
    (2, 'vaccinations',     true, false, false, false),
    (2, 'checkups',         true, false, false, false),
    (2, 'inventory',        true, true,  true,  false),
    (2, 'estimates',        true, true,  true,  false),
    (2, 'shifts',           true, true,  true,  false),
    (2, 'master',           true, true,  true,  false),
    (2, 'hospital-settings',true, false, false, false)
ON CONFLICT DO NOTHING;

-- グループルール（一般: 基本的な業務操作）
INSERT INTO permission_group_rules (group_id, resource, can_view, can_create, can_edit, can_delete) VALUES
    (3, 'dashboard',        true, false, false, false),
    (3, 'owners',           true, true,  true,  false),
    (3, 'reservations',     true, true,  true,  false),
    (3, 'medical-records',  true, true,  true,  false),
    (3, 'hospitalization',  true, true,  true,  false),
    (3, 'trimming',         true, true,  true,  false),
    (3, 'examinations',     true, true,  true,  false),
    (3, 'accounting',       true, false, false, false),
    (3, 'vaccinations',     true, true,  true,  false),
    (3, 'checkups',         true, false, false, false),
    (3, 'inventory',        true, false, false, false),
    (3, 'estimates',        true, false, false, false),
    (3, 'shifts',           true, true,  true,  false),
    (3, 'master',           true, false, false, false),
    (3, 'hospital-settings',true, false, false, false)
ON CONFLICT DO NOTHING;

-- ユーザーへのグループ割当
-- manager@example.com (user_id=10) → 管理者
INSERT INTO user_permission_groups (user_id, group_id) VALUES (10, 1) ON CONFLICT DO NOTHING;
-- exec@example.com (user_id=11) → 執行
INSERT INTO user_permission_groups (user_id, group_id) VALUES (11, 2) ON CONFLICT DO NOTHING;
-- vet@example.com (user_id=5) → 一般
INSERT INTO user_permission_groups (user_id, group_id) VALUES (5, 3) ON CONFLICT DO NOTHING;
-- nurse@example.com (user_id=6) → 一般
INSERT INTO user_permission_groups (user_id, group_id) VALUES (6, 3) ON CONFLICT DO NOTHING;
-- reception@example.com (user_id=7) → 一般
INSERT INTO user_permission_groups (user_id, group_id) VALUES (7, 3) ON CONFLICT DO NOTHING;
-- trimmer@example.com (user_id=8) → 一般
INSERT INTO user_permission_groups (user_id, group_id) VALUES (8, 3) ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('permission_groups', 'id'), (SELECT MAX(id) FROM permission_groups));

-- -----------------------------------------------------------------------------
-- 8. service_types（サービス種別: 7件）
-- 「再診」は visit_type（予約区分: 初診/再診）で管理するため service_types には含めない
-- -----------------------------------------------------------------------------
INSERT INTO service_types (id, clinic_id, name, is_active, description, color, sort_order) VALUES
    (1, 3, '一般診療',     true, '内科・外科・皮膚科などの一般的な診療', '#3B82F6', 1),
    (2, 3, 'ワクチン接種', true, '各種ワクチン接種（予防接種）',         '#10B981', 2),
    (3, 3, '健康診断',     true, '定期健康診断・フィラリア検査など',     '#8B5CF6', 3),
    (4, 3, '手術・処置',   true, '去勢・避妊・その他外科手術',           '#EF4444', 4),
    (5, 3, 'トリミング',   true, 'グルーミング・爪切り・耳掃除など',     '#F59E0B', 5),
    (6, 3, '入院',         true, '入院・ホテル管理',                     '#6B7280', 6),
    (7, 3, '検査',         true, '血液検査・尿検査・画像診断など',       '#EC4899', 7)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('service_types', 'id'), (SELECT MAX(id) FROM service_types));

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
-- 11. exam_types（検査種別: 5件）+ exam_type_items（検査項目）
-- -----------------------------------------------------------------------------
INSERT INTO exam_types (id, clinic_id, name, price, is_active, description, sort_order) VALUES
    (1, 3, '血液検査（CBC）',     3000, true, '全血球計算（Complete Blood Count）',         1),
    (2, 3, '血液化学検査',         5000, true, '肝機能・腎機能・血糖値など生化学的検査',     2),
    (3, 3, '尿検査',               1500, true, '尿試験紙・尿沈渣検査',                       3),
    (4, 3, 'レントゲン検査',       3000, true, 'X線撮影（胸部・腹部・四肢）',                4),
    (5, 3, '超音波検査（エコー）', 5000, true, '腹部エコー・心臓エコー',                     5)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('exam_types', 'id'), (SELECT MAX(id) FROM exam_types));

-- exam_type_items: 血液検査（CBC）
INSERT INTO exam_type_items (id, exam_type_id, name, inspection_value, normal_value, sort_order) VALUES
    (1, 1, 'WBC（白血球数）',      '', '6.0-17.0 x10^3/uL', 1),
    (2, 1, 'RBC（赤血球数）',      '', '5.5-8.5 x10^6/uL',  2),
    (3, 1, 'HCT（ヘマトクリット）', '', '37-55%',            3),
    (4, 1, 'PLT（血小板数）',      '', '175-500 x10^3/uL',  4)
ON CONFLICT DO NOTHING;

-- exam_type_items: 血液化学検査
INSERT INTO exam_type_items (id, exam_type_id, name, inspection_value, normal_value, sort_order) VALUES
    (5, 2, 'ALT（GPT）',        '', '10-125 U/L',    1),
    (6, 2, 'BUN（尿素窒素）',   '', '7-27 mg/dL',    2),
    (7, 2, 'CRE（クレアチニン）', '', '0.5-1.8 mg/dL', 3),
    (8, 2, 'GLU（血糖値）',     '', '74-143 mg/dL',   4)
ON CONFLICT DO NOTHING;

-- exam_type_items: 尿検査
INSERT INTO exam_type_items (id, exam_type_id, name, inspection_value, normal_value, sort_order) VALUES
    (9,  3, '尿比重',   '', '1.015-1.045', 1),
    (10, 3, '尿pH',     '', '5.5-7.5',     2),
    (11, 3, '尿タンパク', '', '陰性',       3),
    (12, 3, '尿潜血',   '', '陰性',        4)
ON CONFLICT DO NOTHING;

-- exam_type_items: レントゲン検査
INSERT INTO exam_type_items (id, exam_type_id, name, inspection_value, normal_value, sort_order) VALUES
    (13, 4, '胸部正面', '', '異常なし', 1),
    (14, 4, '腹部正面', '', '異常なし', 2),
    (15, 4, '四肢',     '', '異常なし', 3)
ON CONFLICT DO NOTHING;

-- exam_type_items: 超音波検査
INSERT INTO exam_type_items (id, exam_type_id, name, inspection_value, normal_value, sort_order) VALUES
    (16, 5, '腹部エコー', '', '異常なし', 1),
    (17, 5, '心臓エコー', '', '異常なし', 2)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('exam_type_items', 'id'), (SELECT MAX(id) FROM exam_type_items));

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
-- ※ combinable は boolean, duration は integer（分単位）
-- -----------------------------------------------------------------------------
INSERT INTO trimming_options (id, clinic_id, name, price, is_active, description, duration, combinable, sort_order) VALUES
    (1, 3, '爪切り',     300, true, '爪のカット・やすりがけ',       10, true, 1),
    (2, 3, '耳掃除',     500, true, '外耳道の洗浄・清掃',           10, true, 2),
    (3, 3, '歯磨き',     500, true, '歯ブラシによるデンタルケア',   15, true, 3),
    (4, 3, '肛門腺絞り', 300, true, '肛門嚢の分泌液除去',            5, true, 4),
    (5, 3, 'リボン装着', 200, true, '仕上げのアクセサリー装着',      5, true, 5)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('trimming_options', 'id'), (SELECT MAX(id) FROM trimming_options));

-- -----------------------------------------------------------------------------
-- 19. diagnosis_categories（診断カテゴリ: 8件）
-- -----------------------------------------------------------------------------
INSERT INTO diagnosis_categories (id, clinic_id, name, is_active, description, sort_order) VALUES
    (1, 3, '消化器系',       true, '胃腸・肝臓・膵臓などの消化器系疾患',   1),
    (2, 3, '呼吸器系',       true, '肺・気管・鼻腔などの呼吸器系疾患',     2),
    (3, 3, '皮膚・被毛',     true, 'アレルギー・感染症などの皮膚疾患',     3),
    (4, 3, '泌尿器系',       true, '腎臓・膀胱・尿道などの泌尿器系疾患',   4),
    (5, 3, '神経系',         true, '脳・脊髄・末梢神経などの神経系疾患',   5),
    (6, 3, '感染症・寄生虫', true, '細菌・ウイルス・寄生虫感染症',         6),
    (7, 3, '腫瘍',           true, '良性・悪性腫瘍（がん）',               7),
    (8, 3, '外傷・骨格',     true, '骨折・咬傷・関節疾患など',             8)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('diagnosis_categories', 'id'), (SELECT MAX(id) FROM diagnosis_categories));

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
    (20, 3, '咬傷',               true, '他動物による咬傷・咬傷感染',         8, 2)
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
-- 22. chief_complaint_categories（主訴区分: 6件）
-- -----------------------------------------------------------------------------
INSERT INTO chief_complaint_categories (id, clinic_id, name, is_active, sort_order) VALUES
    (1, 3, '食欲不振',       true, 1),
    (2, 3, '嘔吐・下痢',     true, 2),
    (3, 3, '皮膚・被毛異常', true, 3),
    (4, 3, '呼吸困難',       true, 4),
    (5, 3, '排尿・排泄異常', true, 5),
    (6, 3, '外傷・骨折',     true, 6)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('chief_complaint_categories', 'id'), (SELECT MAX(id) FROM chief_complaint_categories));

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
