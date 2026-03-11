-- =============================================================================
-- Animal Ekarte - マスタデータシード
-- PostgreSQL 18
-- 冪等性保証: ON CONFLICT DO NOTHING
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 1. staffs（スタッフ）
-- -----------------------------------------------------------------------------
INSERT INTO staffs (id, code, name, status, staff_role, license_number, sort_order) VALUES
    (gen_random_uuid(), 'S-001', '山田 太郎',   'active', 'veterinarian', 'V-001', 1),
    (gen_random_uuid(), 'S-002', '高橋 健一',   'active', 'veterinarian', 'V-002', 2),
    (gen_random_uuid(), 'S-003', '渡辺 博',     'active', 'manager',      '',      3),
    (gen_random_uuid(), 'S-004', '佐藤 花子',   'active', 'nurse',        '',      4),
    (gen_random_uuid(), 'S-005', '伊藤 さくら', 'active', 'nurse',        '',      5),
    (gen_random_uuid(), 'S-006', '鈴木 一郎',   'active', 'trimmer',      '',      6),
    (gen_random_uuid(), 'S-007', '田中 美咲',   'active', 'reception',    '',      7)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 2. examination_types（検査種別）
-- -----------------------------------------------------------------------------
INSERT INTO examination_types (id, code, name, price, status, description, sort_order) VALUES
    (gen_random_uuid(), 'ET-001', '血液検査（CBC）',   3000.00, 'active', '全血球計算（Complete Blood Count）',     1),
    (gen_random_uuid(), 'ET-002', '血液生化学検査',    5000.00, 'active', '臓器機能・代謝状態を評価する血液検査',     2),
    (gen_random_uuid(), 'ET-003', '尿検査',            1500.00, 'active', '腎機能・膀胱・尿路系の評価',              3),
    (gen_random_uuid(), 'ET-004', 'X線検査（胸部）',   3500.00, 'active', '胸部レントゲン撮影',                      4),
    (gen_random_uuid(), 'ET-005', 'X線検査（腹部）',   3500.00, 'active', '腹部レントゲン撮影',                      5),
    (gen_random_uuid(), 'ET-006', 'エコー検査',        5000.00, 'active', '超音波による腹部・心臓の画像診断',         6),
    (gen_random_uuid(), 'ET-007', '心電図検査',        3000.00, 'active', '心臓の電気的活動を記録する検査',           7),
    (gen_random_uuid(), 'ET-008', '眼科検査',          2000.00, 'active', '眼圧・眼底・スリットランプ検査',           8),
    (gen_random_uuid(), 'ET-009', '皮膚検査',          2000.00, 'active', '皮膚掻爬・毛検査・真菌培養',              9),
    (gen_random_uuid(), 'ET-010', '糞便検査',          1000.00, 'active', '寄生虫卵・細菌の検査',                   10)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 3. examination_type_items（検査項目定義）
--    examination_type_id はサブクエリで参照
-- -----------------------------------------------------------------------------

-- 血液検査（CBC）の項目
INSERT INTO examination_type_items (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '白血球数（WBC）',  '5000〜15000 /μL',  1 FROM examination_types et WHERE et.code = 'ET-001'
ON CONFLICT DO NOTHING;

INSERT INTO examination_type_items (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '赤血球数（RBC）',  '550〜850 万/μL',    2 FROM examination_types et WHERE et.code = 'ET-001'
ON CONFLICT DO NOTHING;

INSERT INTO examination_type_items (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, 'ヘモグロビン（Hb）', '12〜18 g/dL',     3 FROM examination_types et WHERE et.code = 'ET-001'
ON CONFLICT DO NOTHING;

INSERT INTO examination_type_items (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '血小板数（PLT）',  '20〜50 万/μL',      4 FROM examination_types et WHERE et.code = 'ET-001'
ON CONFLICT DO NOTHING;

-- 血液生化学検査の項目
INSERT INTO examination_type_items (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, 'ALT（GPT）',       '10〜100 U/L',        1 FROM examination_types et WHERE et.code = 'ET-002'
ON CONFLICT DO NOTHING;

INSERT INTO examination_type_items (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, 'AST（GOT）',       '10〜60 U/L',         2 FROM examination_types et WHERE et.code = 'ET-002'
ON CONFLICT DO NOTHING;

INSERT INTO examination_type_items (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, 'BUN（尿素窒素）',  '8〜30 mg/dL',        3 FROM examination_types et WHERE et.code = 'ET-002'
ON CONFLICT DO NOTHING;

INSERT INTO examination_type_items (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, 'クレアチニン',     '0.4〜1.4 mg/dL',     4 FROM examination_types et WHERE et.code = 'ET-002'
ON CONFLICT DO NOTHING;

INSERT INTO examination_type_items (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '総タンパク（TP）', '5.5〜7.5 g/dL',      5 FROM examination_types et WHERE et.code = 'ET-002'
ON CONFLICT DO NOTHING;

-- 尿検査の項目
INSERT INTO examination_type_items (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, 'pH',               '5.5〜7.5',            1 FROM examination_types et WHERE et.code = 'ET-003'
ON CONFLICT DO NOTHING;

INSERT INTO examination_type_items (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '比重',             '1.015〜1.045',        2 FROM examination_types et WHERE et.code = 'ET-003'
ON CONFLICT DO NOTHING;

INSERT INTO examination_type_items (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, 'タンパク',         '陰性（−）',           3 FROM examination_types et WHERE et.code = 'ET-003'
ON CONFLICT DO NOTHING;

INSERT INTO examination_type_items (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '糖',               '陰性（−）',           4 FROM examination_types et WHERE et.code = 'ET-003'
ON CONFLICT DO NOTHING;

-- X線検査（胸部）の項目
INSERT INTO examination_type_items (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '心臓サイズ',       '正常範囲内',          1 FROM examination_types et WHERE et.code = 'ET-004'
ON CONFLICT DO NOTHING;

INSERT INTO examination_type_items (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '肺野',             '透過性良好・異常陰影なし', 2 FROM examination_types et WHERE et.code = 'ET-004'
ON CONFLICT DO NOTHING;

INSERT INTO examination_type_items (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '縦隔',             '正常',                3 FROM examination_types et WHERE et.code = 'ET-004'
ON CONFLICT DO NOTHING;

-- X線検査（腹部）の項目
INSERT INTO examination_type_items (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '消化管ガス',       '生理的範囲内',        1 FROM examination_types et WHERE et.code = 'ET-005'
ON CONFLICT DO NOTHING;

INSERT INTO examination_type_items (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '肝臓・脾臓',       '腫大なし',            2 FROM examination_types et WHERE et.code = 'ET-005'
ON CONFLICT DO NOTHING;

INSERT INTO examination_type_items (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '結石・異物',       '認めない',            3 FROM examination_types et WHERE et.code = 'ET-005'
ON CONFLICT DO NOTHING;

-- エコー検査の項目
INSERT INTO examination_type_items (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '肝臓エコー輝度',  '均一・正常',          1 FROM examination_types et WHERE et.code = 'ET-006'
ON CONFLICT DO NOTHING;

INSERT INTO examination_type_items (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '腎臓',             '大きさ・形態正常',    2 FROM examination_types et WHERE et.code = 'ET-006'
ON CONFLICT DO NOTHING;

INSERT INTO examination_type_items (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '膀胱',             '壁肥厚なし・結石なし', 3 FROM examination_types et WHERE et.code = 'ET-006'
ON CONFLICT DO NOTHING;

-- 心電図検査の項目
INSERT INTO examination_type_items (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '心拍数',           '犬60〜160 bpm / 猫120〜240 bpm', 1 FROM examination_types et WHERE et.code = 'ET-007'
ON CONFLICT DO NOTHING;

INSERT INTO examination_type_items (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, 'リズム',           '正常洞調律',          2 FROM examination_types et WHERE et.code = 'ET-007'
ON CONFLICT DO NOTHING;

INSERT INTO examination_type_items (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, 'P波・QRS波',       '形態正常',            3 FROM examination_types et WHERE et.code = 'ET-007'
ON CONFLICT DO NOTHING;

-- 眼科検査の項目
INSERT INTO examination_type_items (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '眼圧',             '犬10〜25 mmHg / 猫10〜25 mmHg', 1 FROM examination_types et WHERE et.code = 'ET-008'
ON CONFLICT DO NOTHING;

INSERT INTO examination_type_items (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, 'シルマーティア試験', '15 mm/min 以上',    2 FROM examination_types et WHERE et.code = 'ET-008'
ON CONFLICT DO NOTHING;

INSERT INTO examination_type_items (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '角膜フルオレセイン染色', '染色なし',       3 FROM examination_types et WHERE et.code = 'ET-008'
ON CONFLICT DO NOTHING;

-- 皮膚検査の項目
INSERT INTO examination_type_items (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '皮膚掻爬検査',     '寄生虫・真菌なし',    1 FROM examination_types et WHERE et.code = 'ET-009'
ON CONFLICT DO NOTHING;

INSERT INTO examination_type_items (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '毛検査',           '異常なし',            2 FROM examination_types et WHERE et.code = 'ET-009'
ON CONFLICT DO NOTHING;

INSERT INTO examination_type_items (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, 'ウッド灯検査',     '蛍光なし',            3 FROM examination_types et WHERE et.code = 'ET-009'
ON CONFLICT DO NOTHING;

-- 糞便検査の項目
INSERT INTO examination_type_items (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '寄生虫卵',         '検出なし',            1 FROM examination_types et WHERE et.code = 'ET-010'
ON CONFLICT DO NOTHING;

INSERT INTO examination_type_items (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '細菌（グラム染色）', '正常腸内細菌叢',     2 FROM examination_types et WHERE et.code = 'ET-010'
ON CONFLICT DO NOTHING;

INSERT INTO examination_type_items (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, 'ジアルジア・コクシジウム', '陰性',         3 FROM examination_types et WHERE et.code = 'ET-010'
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 4. vaccines（ワクチン）
-- -----------------------------------------------------------------------------
INSERT INTO vaccines (id, code, name, price, status, species, interval, description, sort_order) VALUES
    (gen_random_uuid(), 'VAC-001', '混合ワクチン5種（犬）',   5000.00, 'active', 'dog',  '1year',   'ジステンパー・パルボ・肝炎・レプト・パラインフルエンザ5種混合', 1),
    (gen_random_uuid(), 'VAC-002', '混合ワクチン8種（犬）',   6000.00, 'active', 'dog',  '1year',   '5種に加えレプトスピラ3型を含む8種混合',                         2),
    (gen_random_uuid(), 'VAC-003', '狂犬病ワクチン',          3000.00, 'active', 'dog',  '1year',   '狂犬病予防法に基づく義務接種',                                  3),
    (gen_random_uuid(), 'VAC-004', '混合ワクチン3種（猫）',   4000.00, 'active', 'cat',  '1year',   '猫ヘルペス・カリシ・パルボウイルス3種混合（コアワクチン）',      4),
    (gen_random_uuid(), 'VAC-005', '混合ワクチン5種（猫）',   5000.00, 'active', 'cat',  '1year',   '3種に加え猫白血病・クラミジアを含む5種混合',                     5),
    (gen_random_uuid(), 'VAC-006', 'ノミ・マダニ予防薬',      1500.00, 'active', 'both', '',        '経口または外用による月1回投与の予防薬',                         6),
    (gen_random_uuid(), 'VAC-007', 'フィラリア予防薬',         800.00, 'active', 'dog',  '1year',   '毎月1回投与・蚊のいる季節（5〜12月）に実施',                    7),
    (gen_random_uuid(), 'VAC-008', '猫白血病ワクチン',        4000.00, 'active', 'cat',  '1year',   '猫白血病ウイルス感染予防（外出猫・多頭飼育猫に推奨）',           8)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 5. medicines（薬剤）
-- -----------------------------------------------------------------------------
INSERT INTO medicines (id, code, name, price, status, dosage_form, medicine_unit, description, default_quantity, sort_order) VALUES
    (gen_random_uuid(), 'MED-001', 'アモキシシリン',             50.00, 'active', 'tablet',    'per_tablet', '広域スペクトル抗生物質・皮膚・呼吸器・消化器感染に使用',    1, 1),
    (gen_random_uuid(), 'MED-002', 'エンロフロキサシン',         80.00, 'active', 'tablet',    'per_tablet', 'フルオロキノロン系抗菌薬・尿路・皮膚・消化器感染に使用',    1, 2),
    (gen_random_uuid(), 'MED-003', 'プレドニゾロン',             30.00, 'active', 'tablet',    'per_tablet', 'ステロイド系抗炎症薬・アレルギー・免疫疾患に使用',          1, 3),
    (gen_random_uuid(), 'MED-004', 'フロセミド',                 20.00, 'active', 'tablet',    'per_tablet', 'ループ利尿薬・心疾患・腎疾患による浮腫に使用',              1, 4),
    (gen_random_uuid(), 'MED-005', 'メトロニダゾール',           40.00, 'active', 'tablet',    'per_tablet', '消化器感染・ジアルジア・嫌気性菌感染に使用',                1, 5),
    (gen_random_uuid(), 'MED-006', 'ビタミンB12注射',           500.00, 'active', 'injection', 'per_dose',   '貧血・神経疾患・食欲不振時の補助療法',                     1, 6),
    (gen_random_uuid(), 'MED-007', '生理食塩水',                  5.00, 'active', 'liquid',    'per_ml',     '点滴・洗浄・希釈に使用する等張電解質液',                   50, 7),
    (gen_random_uuid(), 'MED-008', 'クロルヘキシジン',           10.00, 'active', 'liquid',    'per_ml',     '創傷・皮膚・耳道の消毒・洗浄に使用',                       10, 8),
    (gen_random_uuid(), 'MED-009', 'ネクスガード',             1500.00, 'active', 'tablet',    'per_tablet', 'アフォキソラネル製剤・ノミ・マダニ駆除（月1回経口投与）',   1, 9),
    (gen_random_uuid(), 'MED-010', 'モキシデクチン',            800.00, 'active', 'liquid',    'per_dose',   'フィラリア・回虫・鉤虫予防に使用',                          1, 10),
    (gen_random_uuid(), 'MED-011', 'フィプロスポット',          1200.00, 'active', 'topical',   'per_dose',   'フィプロニル製剤・ノミ・マダニ駆除用スポットオン',          1, 11),
    (gen_random_uuid(), 'MED-012', '酢酸メチルプレドニゾロン',  800.00, 'active', 'injection', 'per_dose',   '長時間作用型注射ステロイド・アレルギー・関節炎に使用',      1, 12),
    (gen_random_uuid(), 'MED-013', '硫酸アトロピン点眼液',      200.00, 'active', 'liquid',    'per_ml',     '散瞳・眼底検査・虹彩毛様体炎治療に使用',                    1, 13),
    (gen_random_uuid(), 'MED-014', 'テトラサイクリン眼軟膏',    600.00, 'active', 'topical',   'per_dose',   '細菌性結膜炎・角膜炎の治療に使用',                          1, 14),
    (gen_random_uuid(), 'MED-015', 'カルプロフェン',            100.00, 'active', 'tablet',    'per_tablet', 'NSAIDs系鎮痛消炎薬・術後疼痛・骨関節炎に使用',             1, 15)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 6. insurances（保険）
-- -----------------------------------------------------------------------------
INSERT INTO insurances (id, code, name, status, coverage_rate, contact_phone, description, sort_order) VALUES
    (gen_random_uuid(), 'INS-001', 'アニコム損害保険',           'active', '70', '0120-051-140', 'どうぶつ健保ふぁみりぃ・ぷち等を取り扱い', 1),
    (gen_random_uuid(), 'INS-002', 'アイペット損害保険',         'active', '70', '0120-917-800', 'うちの子・うちの子Light等を取り扱い',       2),
    (gen_random_uuid(), 'INS-003', 'ペット&ファミリー損害保険', 'active', '50', '0120-81-8320', 'げんきナンバーわんスリム等を取り扱い',     3),
    (gen_random_uuid(), 'INS-004', 'PS保険',                    'active', '50', '0120-099-909', 'ペットメディカルサポートが提供する保険',    4),
    (gen_random_uuid(), 'INS-005', '楽天ペット保険',             'active', '70', '0800-600-0204', '楽天グループのペット保険サービス',         5)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 7. cages（ケージ）
-- -----------------------------------------------------------------------------
INSERT INTO cages (id, code, name, status, cage_type, cage_size, description, sort_order) VALUES
    (gen_random_uuid(), 'CGE-001', 'ICUケージ大',   'active', 'icu',     'large',  '酸素濃度・温度管理機能付き集中治療ケージ（大型犬用）', 1),
    (gen_random_uuid(), 'CGE-002', 'ICUケージ中',   'active', 'icu',     'medium', '酸素濃度・温度管理機能付き集中治療ケージ（中型犬用）', 2),
    (gen_random_uuid(), 'CGE-003', '犬用ケージ大1', 'active', 'dog',     'large',  '大型犬用入院ケージ（第1号）',                         3),
    (gen_random_uuid(), 'CGE-004', '犬用ケージ大2', 'active', 'dog',     'large',  '大型犬用入院ケージ（第2号）',                         4),
    (gen_random_uuid(), 'CGE-005', '犬用ケージ中1', 'active', 'dog',     'medium', '中型犬用入院ケージ（第1号）',                         5),
    (gen_random_uuid(), 'CGE-006', '犬用ケージ中2', 'active', 'dog',     'medium', '中型犬用入院ケージ（第2号）',                         6),
    (gen_random_uuid(), 'CGE-007', '犬用ケージ小1', 'active', 'dog',     'small',  '小型犬用入院ケージ（第1号）',                         7),
    (gen_random_uuid(), 'CGE-008', '猫用ケージ1',   'active', 'cat',     'small',  '猫専用入院ケージ（第1号）',                           8),
    (gen_random_uuid(), 'CGE-009', '猫用ケージ2',   'active', 'cat',     'small',  '猫専用入院ケージ（第2号）',                           9),
    (gen_random_uuid(), 'CGE-010', '猫用ケージ3',   'active', 'cat',     'small',  '猫専用入院ケージ（第3号）',                          10),
    (gen_random_uuid(), 'CGE-011', '一般ケージ大',  'active', 'general', 'large',  '多目的入院ケージ・大サイズ',                          11),
    (gen_random_uuid(), 'CGE-012', '一般ケージ中',  'active', 'general', 'medium', '多目的入院ケージ・中サイズ',                          12)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 8. service_types（サービス種別・予約区分）
-- -----------------------------------------------------------------------------
INSERT INTO service_types (id, code, name, status, color, description, sort_order) VALUES
    (gen_random_uuid(), 'SVC-001', '一般診察',     'active', '#3B82F6', '通常の外来診察',                     1),
    (gen_random_uuid(), 'SVC-002', 'ワクチン接種', 'active', '#10B981', '各種予防接種',                       2),
    (gen_random_uuid(), 'SVC-003', '手術',         'active', '#EF4444', '外科的手術・処置',                   3),
    (gen_random_uuid(), 'SVC-004', '健康診断',     'active', '#F59E0B', '定期健診・シニア検診',               4),
    (gen_random_uuid(), 'SVC-005', '緊急診察',     'active', '#DC2626', '時間外・緊急対応',                   5),
    (gen_random_uuid(), 'SVC-006', 'トリミング予約', 'active', '#8B5CF6', 'グルーミング・トリミング予約',     6)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 9. consultations（診察項目）
-- -----------------------------------------------------------------------------
INSERT INTO consultations (id, code, name, price, status, time_condition, description, sort_order) VALUES
    (gen_random_uuid(), 'CON-001', '初診料',         3000.00, 'active', '初診',   '初めて来院または前回受診から6ヶ月以上経過した場合',  1),
    (gen_random_uuid(), 'CON-002', '再診料',         1500.00, 'active', '再診',   '前回受診から6ヶ月以内の再来院',                      2),
    (gen_random_uuid(), 'CON-003', '時間外診察料',   5000.00, 'active', '時間外', '診療時間外（18時以降・休診日）の診察',               3),
    (gen_random_uuid(), 'CON-004', '救急診察料',     8000.00, 'active', '救急',   '緊急処置・救急対応が必要な診察',                     4),
    (gen_random_uuid(), 'CON-005', '往診料',         5000.00, 'active', '',       '自宅・施設への往診',                                 5),
    (gen_random_uuid(), 'CON-006', '電話相談料',      500.00, 'active', '',       '電話による症状相談・診察アドバイス',                  6),
    (gen_random_uuid(), 'CON-007', '皮膚科診察',     2000.00, 'active', '',       '皮膚疾患専門診察（皮膚検査含む）',                   7),
    (gen_random_uuid(), 'CON-008', '眼科診察',       2000.00, 'active', '',       '眼科専門診察（眼圧測定含む）',                       8)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 10. procedures（処置項目）
-- -----------------------------------------------------------------------------
INSERT INTO procedures (id, code, name, price, status, anesthesia, description, sort_order) VALUES
    (gen_random_uuid(), 'PRO-001', '注射（皮下）',      500.00, 'active', 'none',    '皮下組織への薬剤投与',                         1),
    (gen_random_uuid(), 'PRO-002', '注射（静脈内）',    800.00, 'active', 'none',    '静脈内への薬剤・輸液投与',                     2),
    (gen_random_uuid(), 'PRO-003', '点滴（1時間）',   3000.00, 'active', 'none',    '持続点滴（1時間あたり）',                      3),
    (gen_random_uuid(), 'PRO-004', '採血',            1000.00, 'active', 'none',    '検査用の採血処置',                             4),
    (gen_random_uuid(), 'PRO-005', '外科縫合（小）',  5000.00, 'active', 'local',   '局所麻酔下での小創傷縫合',                     5),
    (gen_random_uuid(), 'PRO-006', '外科縫合（中）', 10000.00, 'active', 'general', '全身麻酔下での中程度創傷縫合',                 6),
    (gen_random_uuid(), 'PRO-007', '去勢手術（犬）', 30000.00, 'active', 'general', '雄犬の精巣摘出術（全身麻酔・術後管理含む）',  7),
    (gen_random_uuid(), 'PRO-008', '避妊手術（犬）', 50000.00, 'active', 'general', '雌犬の子宮卵巣摘出術（全身麻酔・術後管理含む）', 8),
    (gen_random_uuid(), 'PRO-009', '去勢手術（猫）', 20000.00, 'active', 'general', '雄猫の精巣摘出術（全身麻酔・術後管理含む）',  9),
    (gen_random_uuid(), 'PRO-010', '避妊手術（猫）', 35000.00, 'active', 'general', '雌猫の子宮卵巣摘出術（全身麻酔・術後管理含む）', 10),
    (gen_random_uuid(), 'PRO-011', 'カテーテル挿入',  2000.00, 'active', 'none',    '尿路カテーテルの挿入・固定',                  11),
    (gen_random_uuid(), 'PRO-012', '浣腸',           1500.00, 'active', 'none',    '便秘・消化管閉塞時の浣腸処置',                12)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 11. hospitalization_plans（入院プラン）
-- -----------------------------------------------------------------------------
INSERT INTO hospitalization_plans (id, code, name, price, status, body_size, billing_unit, description, sort_order) VALUES
    (gen_random_uuid(), 'HOS-001', '入院（小型犬・猫）', 5000.00, 'active', 'small',  'per_day',   '10kg以下の小型犬・猫の入院（24時間ケア込み）',  1),
    (gen_random_uuid(), 'HOS-002', '入院（中型犬）',     7000.00, 'active', 'medium', 'per_day',   '10〜25kgの中型犬の入院（24時間ケア込み）',      2),
    (gen_random_uuid(), 'HOS-003', '入院（大型犬）',     9000.00, 'active', 'large',  'per_day',   '25kg超の大型犬の入院（24時間ケア込み）',         3),
    (gen_random_uuid(), 'HOS-004', 'ホテル（小型犬・猫）', 3000.00, 'active', 'small', 'per_night', '10kg以下の小型犬・猫のペットホテル（1泊）',    4),
    (gen_random_uuid(), 'HOS-005', 'ホテル（中型犬）',   4000.00, 'active', 'medium', 'per_night', '10〜25kgの中型犬のペットホテル（1泊）',         5),
    (gen_random_uuid(), 'HOS-006', 'ホテル（大型犬）',   5000.00, 'active', 'large',  'per_night', '25kg超の大型犬のペットホテル（1泊）',            6)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 12. trimming_courses（トリミングコース）
-- -----------------------------------------------------------------------------
INSERT INTO trimming_courses (id, code, name, price, status, target_size, duration, description, sort_order) VALUES
    (gen_random_uuid(), 'TRC-001', '猫コース（フルグルーミング）', 6000.00, 'active', 'cat',    '90分', 'シャンプー・ドライ・爪切り・耳掃除・ブラッシング込み', 1),
    (gen_random_uuid(), 'TRC-002', '小型犬コース',                 4000.00, 'active', 'small',  '60分', '10kg以下の小型犬向けフルトリミング',                  2),
    (gen_random_uuid(), 'TRC-003', '中型犬コース',                 6000.00, 'active', 'medium', '90分', '10〜25kgの中型犬向けフルトリミング',                  3),
    (gen_random_uuid(), 'TRC-004', '大型犬コース',                 9000.00, 'active', 'large',  '120分','25kg超の大型犬向けフルトリミング',                    4),
    (gen_random_uuid(), 'TRC-005', 'シャンプー＆ドライのみ',       2500.00, 'active', 'small',  '30分', 'シャンプーとドライのみのシンプルコース（小型犬・猫）', 5)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 13. trimming_options（トリミングオプション）
-- -----------------------------------------------------------------------------
INSERT INTO trimming_options (id, code, name, price, status, combinable, duration, description, sort_order) VALUES
    (gen_random_uuid(), 'TRO-001', '爪切り',         500.00, 'active', 'yes', '10分', '前後肢の爪のカット',                              1),
    (gen_random_uuid(), 'TRO-002', '耳掃除',         500.00, 'active', 'yes', '10分', '耳道の汚れ除去・外耳炎予防',                      2),
    (gen_random_uuid(), 'TRO-003', '肛門腺絞り',     500.00, 'active', 'yes', '5分',  '肛門腺（臭腺）の分泌物排出',                      3),
    (gen_random_uuid(), 'TRO-004', '歯磨き',         800.00, 'active', 'yes', '10分', '歯垢除去・口腔内ケア',                            4),
    (gen_random_uuid(), 'TRO-005', '薬浴シャンプー', 2000.00, 'active', 'yes', '15分', '皮膚疾患・アレルギー用薬用シャンプーで洗浄',     5),
    (gen_random_uuid(), 'TRO-006', 'ノミ取りシャンプー', 1500.00, 'active', 'yes', '15分', 'ノミ駆除成分配合シャンプーによる洗浄',       6)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 14. diagnosis_categories（診断カテゴリ）
-- -----------------------------------------------------------------------------
INSERT INTO diagnosis_categories (id, code, name, status, description, sort_order) VALUES
    (gen_random_uuid(), 'DC-001', '内科疾患',     'active', '消化器・泌尿器・内分泌等の内科的疾患',   1),
    (gen_random_uuid(), 'DC-002', '外科疾患',     'active', '骨折・外傷・腫瘍等の外科的疾患',         2),
    (gen_random_uuid(), 'DC-003', '皮膚科疾患',   'active', 'アレルギー・感染・寄生虫等の皮膚疾患',  3),
    (gen_random_uuid(), 'DC-004', '眼科疾患',     'active', '結膜炎・白内障・緑内障等の眼科疾患',    4),
    (gen_random_uuid(), 'DC-005', '耳科疾患',     'active', '外耳炎・中耳炎等の耳科疾患',            5),
    (gen_random_uuid(), 'DC-006', '循環器疾患',   'active', '心臓弁膜症・不整脈等の循環器疾患',      6),
    (gen_random_uuid(), 'DC-007', '感染症',       'active', 'ウイルス・細菌・寄生虫による感染症',    7),
    (gen_random_uuid(), 'DC-008', 'その他',       'active', '健康診断異常・予防処置・その他',        8)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 15. diagnosis_names（診断名）
--     diagnosis_category_id はサブクエリで参照
-- -----------------------------------------------------------------------------

-- 内科疾患
INSERT INTO diagnosis_names (id, code, name, status, diagnosis_category_id, sort_order)
SELECT gen_random_uuid(), 'DN-001', '胃腸炎',     'active', dc.id, 1 FROM diagnosis_categories dc WHERE dc.code = 'DC-001'
ON CONFLICT DO NOTHING;

INSERT INTO diagnosis_names (id, code, name, status, diagnosis_category_id, sort_order)
SELECT gen_random_uuid(), 'DN-002', '膵炎',       'active', dc.id, 2 FROM diagnosis_categories dc WHERE dc.code = 'DC-001'
ON CONFLICT DO NOTHING;

INSERT INTO diagnosis_names (id, code, name, status, diagnosis_category_id, sort_order)
SELECT gen_random_uuid(), 'DN-003', '慢性腎臓病', 'active', dc.id, 3 FROM diagnosis_categories dc WHERE dc.code = 'DC-001'
ON CONFLICT DO NOTHING;

INSERT INTO diagnosis_names (id, code, name, status, diagnosis_category_id, sort_order)
SELECT gen_random_uuid(), 'DN-004', '肝臓病',     'active', dc.id, 4 FROM diagnosis_categories dc WHERE dc.code = 'DC-001'
ON CONFLICT DO NOTHING;

-- 外科疾患
INSERT INTO diagnosis_names (id, code, name, status, diagnosis_category_id, sort_order)
SELECT gen_random_uuid(), 'DN-005', '骨折',         'active', dc.id, 1 FROM diagnosis_categories dc WHERE dc.code = 'DC-002'
ON CONFLICT DO NOTHING;

INSERT INTO diagnosis_names (id, code, name, status, diagnosis_category_id, sort_order)
SELECT gen_random_uuid(), 'DN-006', '脱臼',         'active', dc.id, 2 FROM diagnosis_categories dc WHERE dc.code = 'DC-002'
ON CONFLICT DO NOTHING;

INSERT INTO diagnosis_names (id, code, name, status, diagnosis_category_id, sort_order)
SELECT gen_random_uuid(), 'DN-007', '創傷・裂傷',   'active', dc.id, 3 FROM diagnosis_categories dc WHERE dc.code = 'DC-002'
ON CONFLICT DO NOTHING;

INSERT INTO diagnosis_names (id, code, name, status, diagnosis_category_id, sort_order)
SELECT gen_random_uuid(), 'DN-008', '腫瘍（外科）', 'active', dc.id, 4 FROM diagnosis_categories dc WHERE dc.code = 'DC-002'
ON CONFLICT DO NOTHING;

-- 皮膚科疾患
INSERT INTO diagnosis_names (id, code, name, status, diagnosis_category_id, sort_order)
SELECT gen_random_uuid(), 'DN-009', 'アレルギー性皮膚炎', 'active', dc.id, 1 FROM diagnosis_categories dc WHERE dc.code = 'DC-003'
ON CONFLICT DO NOTHING;

INSERT INTO diagnosis_names (id, code, name, status, diagnosis_category_id, sort_order)
SELECT gen_random_uuid(), 'DN-010', '疥癬（ヒゼンダニ）', 'active', dc.id, 2 FROM diagnosis_categories dc WHERE dc.code = 'DC-003'
ON CONFLICT DO NOTHING;

INSERT INTO diagnosis_names (id, code, name, status, diagnosis_category_id, sort_order)
SELECT gen_random_uuid(), 'DN-011', '真菌感染（皮膚糸状菌）', 'active', dc.id, 3 FROM diagnosis_categories dc WHERE dc.code = 'DC-003'
ON CONFLICT DO NOTHING;

INSERT INTO diagnosis_names (id, code, name, status, diagnosis_category_id, sort_order)
SELECT gen_random_uuid(), 'DN-012', '膿皮症（細菌感染）', 'active', dc.id, 4 FROM diagnosis_categories dc WHERE dc.code = 'DC-003'
ON CONFLICT DO NOTHING;

-- 眼科疾患
INSERT INTO diagnosis_names (id, code, name, status, diagnosis_category_id, sort_order)
SELECT gen_random_uuid(), 'DN-013', '結膜炎', 'active', dc.id, 1 FROM diagnosis_categories dc WHERE dc.code = 'DC-004'
ON CONFLICT DO NOTHING;

INSERT INTO diagnosis_names (id, code, name, status, diagnosis_category_id, sort_order)
SELECT gen_random_uuid(), 'DN-014', '白内障', 'active', dc.id, 2 FROM diagnosis_categories dc WHERE dc.code = 'DC-004'
ON CONFLICT DO NOTHING;

INSERT INTO diagnosis_names (id, code, name, status, diagnosis_category_id, sort_order)
SELECT gen_random_uuid(), 'DN-015', '緑内障', 'active', dc.id, 3 FROM diagnosis_categories dc WHERE dc.code = 'DC-004'
ON CONFLICT DO NOTHING;

-- 耳科疾患
INSERT INTO diagnosis_names (id, code, name, status, diagnosis_category_id, sort_order)
SELECT gen_random_uuid(), 'DN-016', '外耳炎', 'active', dc.id, 1 FROM diagnosis_categories dc WHERE dc.code = 'DC-005'
ON CONFLICT DO NOTHING;

INSERT INTO diagnosis_names (id, code, name, status, diagnosis_category_id, sort_order)
SELECT gen_random_uuid(), 'DN-017', '中耳炎', 'active', dc.id, 2 FROM diagnosis_categories dc WHERE dc.code = 'DC-005'
ON CONFLICT DO NOTHING;

-- 循環器疾患
INSERT INTO diagnosis_names (id, code, name, status, diagnosis_category_id, sort_order)
SELECT gen_random_uuid(), 'DN-018', '心臓弁膜症（僧帽弁閉鎖不全）', 'active', dc.id, 1 FROM diagnosis_categories dc WHERE dc.code = 'DC-006'
ON CONFLICT DO NOTHING;

INSERT INTO diagnosis_names (id, code, name, status, diagnosis_category_id, sort_order)
SELECT gen_random_uuid(), 'DN-019', '不整脈', 'active', dc.id, 2 FROM diagnosis_categories dc WHERE dc.code = 'DC-006'
ON CONFLICT DO NOTHING;

-- 感染症
INSERT INTO diagnosis_names (id, code, name, status, diagnosis_category_id, sort_order)
SELECT gen_random_uuid(), 'DN-020', 'パルボウイルス感染症', 'active', dc.id, 1 FROM diagnosis_categories dc WHERE dc.code = 'DC-007'
ON CONFLICT DO NOTHING;

INSERT INTO diagnosis_names (id, code, name, status, diagnosis_category_id, sort_order)
SELECT gen_random_uuid(), 'DN-021', '猫白血病ウイルス感染症（FeLV）', 'active', dc.id, 2 FROM diagnosis_categories dc WHERE dc.code = 'DC-007'
ON CONFLICT DO NOTHING;

INSERT INTO diagnosis_names (id, code, name, status, diagnosis_category_id, sort_order)
SELECT gen_random_uuid(), 'DN-022', '猫免疫不全ウイルス感染症（FIV）', 'active', dc.id, 3 FROM diagnosis_categories dc WHERE dc.code = 'DC-007'
ON CONFLICT DO NOTHING;

-- その他
INSERT INTO diagnosis_names (id, code, name, status, diagnosis_category_id, sort_order)
SELECT gen_random_uuid(), 'DN-023', '健康診断異常', 'active', dc.id, 1 FROM diagnosis_categories dc WHERE dc.code = 'DC-008'
ON CONFLICT DO NOTHING;

INSERT INTO diagnosis_names (id, code, name, status, diagnosis_category_id, sort_order)
SELECT gen_random_uuid(), 'DN-024', '予防処置', 'active', dc.id, 2 FROM diagnosis_categories dc WHERE dc.code = 'DC-008'
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 16. checkup_types（健診種別）
-- -----------------------------------------------------------------------------
INSERT INTO checkup_types (id, code, name, price, status, interval, target_age, description, sort_order) VALUES
    (gen_random_uuid(), 'CKP-001', '年1回健康診断',              5000.00, 'active', '1year',   '',      '身体検査・体重測定・便検査・聴診を含む総合健診',                1),
    (gen_random_uuid(), 'CKP-002', 'シニア健康診断（7歳以上）', 8000.00, 'active', '6months', '7歳以上', '血液検査・X線・エコーを含む高齢動物向け精密健診（年2回推奨）', 2),
    (gen_random_uuid(), 'CKP-003', '子犬・子猫健康診断',        3000.00, 'active', '',        '1歳未満', '初回ワクチン前後の身体検査・寄生虫検査',                        3),
    (gen_random_uuid(), 'CKP-004', '歯科検診',                  2000.00, 'active', '1year',   '',      '口腔内チェック・歯垢・歯石の評価',                              4),
    (gen_random_uuid(), 'CKP-005', '心臓検診',                  5000.00, 'active', '',        '',      '聴診・心電図・胸部X線による心臓専門検診',                       5)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 17. clinic_info（病院情報）
-- -----------------------------------------------------------------------------
INSERT INTO clinic_info (id, name, branch_name, postal_code, address, phone_number, fax_number, registration_number, director_name, email, website)
VALUES (
    gen_random_uuid(),
    'アニマルクリニック田中',
    '',
    '150-0001',
    '東京都渋谷区神宮前1-1-1',
    '03-1234-5678',
    '03-1234-5679',
    'REG-123456',
    '田中 博',
    'info@animal-clinic-tanaka.jp',
    'https://animal-clinic-tanaka.jp'
)
ON CONFLICT DO NOTHING;

-- =============================================================================
-- シード完了
-- =============================================================================
