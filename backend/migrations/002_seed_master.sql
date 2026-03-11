-- =============================================================================
-- Animal Ekarte - マスタデータシード
-- PostgreSQL 18
-- 冪等性保証: ON CONFLICT (code) DO NOTHING
-- 対象DBスキーマ: 現行DBに合わせた実テーブル構造に準拠
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 1. staff_members（スタッフ）
--    実テーブル: staff_members
--    カラム: id, code, name, name_kana, role, license_number, email, phone,
--            is_active, sort_order
-- -----------------------------------------------------------------------------
INSERT INTO staff_members (id, code, name, name_kana, role, license_number, email, phone, is_active, sort_order) VALUES
    (gen_random_uuid(), 'S-001', '山田 太郎',   'ヤマダ タロウ',   'veterinarian', 'V-001', 'yamada@animal-clinic-tanaka.jp',   '090-1111-0001', true, 1),
    (gen_random_uuid(), 'S-002', '高橋 健一',   'タカハシ ケンイチ', 'veterinarian', 'V-002', 'takahashi@animal-clinic-tanaka.jp', '090-1111-0002', true, 2),
    (gen_random_uuid(), 'S-003', '渡辺 博',     'ワタナベ ヒロシ', 'manager',      '',      'watanabe@animal-clinic-tanaka.jp', '090-1111-0003', true, 3),
    (gen_random_uuid(), 'S-004', '佐藤 花子',   'サトウ ハナコ',   'nurse',        '',      'sato@animal-clinic-tanaka.jp',     '090-1111-0004', true, 4),
    (gen_random_uuid(), 'S-005', '伊藤 さくら', 'イトウ サクラ',   'nurse',        '',      'ito@animal-clinic-tanaka.jp',      '090-1111-0005', true, 5),
    (gen_random_uuid(), 'S-006', '鈴木 一郎',   'スズキ イチロウ', 'trimmer',      '',      'suzuki@animal-clinic-tanaka.jp',   '090-1111-0006', true, 6),
    (gen_random_uuid(), 'S-007', '田中 美咲',   'タナカ ミサキ',   'reception',    '',      'tanaka@animal-clinic-tanaka.jp',   '090-1111-0007', true, 7)
ON CONFLICT (code) DO NOTHING;

-- -----------------------------------------------------------------------------
-- 2. examination_types（検査種別）
--    実テーブル: examination_types
--    カラム: id, code, name, color, price, parent_id, description, is_active, sort_order
-- -----------------------------------------------------------------------------
INSERT INTO examination_types (id, code, name, price, description, is_active, sort_order) VALUES
    (gen_random_uuid(), 'ET-001', '血液検査（CBC）',     3000.00, '全血球計算（Complete Blood Count）',                  true,  1),
    (gen_random_uuid(), 'ET-002', '血液生化学検査',      5000.00, '臓器機能・代謝状態を評価する血液検査',                true,  2),
    (gen_random_uuid(), 'ET-003', '尿検査',              1500.00, '腎機能・膀胱・尿路系の評価',                          true,  3),
    (gen_random_uuid(), 'ET-004', 'X線検査（胸部）',     3500.00, '胸部レントゲン撮影',                                  true,  4),
    (gen_random_uuid(), 'ET-005', 'X線検査（腹部）',     3500.00, '腹部レントゲン撮影',                                  true,  5),
    (gen_random_uuid(), 'ET-006', 'エコー検査',          5000.00, '超音波による腹部・心臓の画像診断',                    true,  6),
    (gen_random_uuid(), 'ET-007', '心電図検査',          3000.00, '心臓の電気的活動を記録する検査',                      true,  7),
    (gen_random_uuid(), 'ET-008', '眼科検査',            2000.00, '眼圧・眼底・スリットランプ検査',                      true,  8),
    (gen_random_uuid(), 'ET-009', '皮膚検査',            2000.00, '皮膚掻爬・毛検査・真菌培養',                          true,  9),
    (gen_random_uuid(), 'ET-010', '糞便検査',            1000.00, '寄生虫卵・細菌の検査',                                true, 10)
ON CONFLICT (code) DO NOTHING;

-- -----------------------------------------------------------------------------
-- 3. examination_type_inspections（検査項目定義）
--    実テーブル: examination_type_inspections
--    カラム: id, examination_type_id, name, inspection_value, normal_value, sort_order
--    ※ テーブルにUNIQUE制約がないため、既存データとの重複チェックをWHEREで行う
-- -----------------------------------------------------------------------------

-- 血液検査（CBC）の項目
INSERT INTO examination_type_inspections (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '白血球数（WBC）',     '5000〜15000 /μL',      1
FROM examination_types et WHERE et.code = 'ET-001'
  AND NOT EXISTS (SELECT 1 FROM examination_type_inspections ei WHERE ei.examination_type_id = et.id AND ei.name = '白血球数（WBC）');

INSERT INTO examination_type_inspections (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '赤血球数（RBC）',     '550〜850 万/μL',         2
FROM examination_types et WHERE et.code = 'ET-001'
  AND NOT EXISTS (SELECT 1 FROM examination_type_inspections ei WHERE ei.examination_type_id = et.id AND ei.name = '赤血球数（RBC）');

INSERT INTO examination_type_inspections (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, 'ヘモグロビン（Hb）',  '12〜18 g/dL',            3
FROM examination_types et WHERE et.code = 'ET-001'
  AND NOT EXISTS (SELECT 1 FROM examination_type_inspections ei WHERE ei.examination_type_id = et.id AND ei.name = 'ヘモグロビン（Hb）');

INSERT INTO examination_type_inspections (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '血小板数（PLT）',     '20〜50 万/μL',           4
FROM examination_types et WHERE et.code = 'ET-001'
  AND NOT EXISTS (SELECT 1 FROM examination_type_inspections ei WHERE ei.examination_type_id = et.id AND ei.name = '血小板数（PLT）');

-- 血液生化学検査の項目
INSERT INTO examination_type_inspections (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, 'ALT（GPT）',          '10〜100 U/L',             1
FROM examination_types et WHERE et.code = 'ET-002'
  AND NOT EXISTS (SELECT 1 FROM examination_type_inspections ei WHERE ei.examination_type_id = et.id AND ei.name = 'ALT（GPT）');

INSERT INTO examination_type_inspections (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, 'AST（GOT）',          '10〜60 U/L',              2
FROM examination_types et WHERE et.code = 'ET-002'
  AND NOT EXISTS (SELECT 1 FROM examination_type_inspections ei WHERE ei.examination_type_id = et.id AND ei.name = 'AST（GOT）');

INSERT INTO examination_type_inspections (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, 'BUN（尿素窒素）',     '8〜30 mg/dL',             3
FROM examination_types et WHERE et.code = 'ET-002'
  AND NOT EXISTS (SELECT 1 FROM examination_type_inspections ei WHERE ei.examination_type_id = et.id AND ei.name = 'BUN（尿素窒素）');

INSERT INTO examination_type_inspections (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, 'クレアチニン',        '0.4〜1.4 mg/dL',          4
FROM examination_types et WHERE et.code = 'ET-002'
  AND NOT EXISTS (SELECT 1 FROM examination_type_inspections ei WHERE ei.examination_type_id = et.id AND ei.name = 'クレアチニン');

INSERT INTO examination_type_inspections (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '総タンパク（TP）',    '5.5〜7.5 g/dL',           5
FROM examination_types et WHERE et.code = 'ET-002'
  AND NOT EXISTS (SELECT 1 FROM examination_type_inspections ei WHERE ei.examination_type_id = et.id AND ei.name = '総タンパク（TP）');

-- 尿検査の項目
INSERT INTO examination_type_inspections (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, 'pH',                  '5.5〜7.5',                1
FROM examination_types et WHERE et.code = 'ET-003'
  AND NOT EXISTS (SELECT 1 FROM examination_type_inspections ei WHERE ei.examination_type_id = et.id AND ei.name = 'pH');

INSERT INTO examination_type_inspections (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '比重',                '1.015〜1.045',            2
FROM examination_types et WHERE et.code = 'ET-003'
  AND NOT EXISTS (SELECT 1 FROM examination_type_inspections ei WHERE ei.examination_type_id = et.id AND ei.name = '比重');

INSERT INTO examination_type_inspections (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, 'タンパク',            '陰性（−）',               3
FROM examination_types et WHERE et.code = 'ET-003'
  AND NOT EXISTS (SELECT 1 FROM examination_type_inspections ei WHERE ei.examination_type_id = et.id AND ei.name = 'タンパク');

INSERT INTO examination_type_inspections (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '糖',                  '陰性（−）',               4
FROM examination_types et WHERE et.code = 'ET-003'
  AND NOT EXISTS (SELECT 1 FROM examination_type_inspections ei WHERE ei.examination_type_id = et.id AND ei.name = '糖');

-- X線検査（胸部）の項目
INSERT INTO examination_type_inspections (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '心臓サイズ',          '正常範囲内',              1
FROM examination_types et WHERE et.code = 'ET-004'
  AND NOT EXISTS (SELECT 1 FROM examination_type_inspections ei WHERE ei.examination_type_id = et.id AND ei.name = '心臓サイズ');

INSERT INTO examination_type_inspections (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '肺野',                '透過性良好・異常陰影なし', 2
FROM examination_types et WHERE et.code = 'ET-004'
  AND NOT EXISTS (SELECT 1 FROM examination_type_inspections ei WHERE ei.examination_type_id = et.id AND ei.name = '肺野');

INSERT INTO examination_type_inspections (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '縦隔',                '正常',                    3
FROM examination_types et WHERE et.code = 'ET-004'
  AND NOT EXISTS (SELECT 1 FROM examination_type_inspections ei WHERE ei.examination_type_id = et.id AND ei.name = '縦隔');

-- X線検査（腹部）の項目
INSERT INTO examination_type_inspections (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '消化管ガス',          '生理的範囲内',            1
FROM examination_types et WHERE et.code = 'ET-005'
  AND NOT EXISTS (SELECT 1 FROM examination_type_inspections ei WHERE ei.examination_type_id = et.id AND ei.name = '消化管ガス');

INSERT INTO examination_type_inspections (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '肝臓・脾臓',          '腫大なし',                2
FROM examination_types et WHERE et.code = 'ET-005'
  AND NOT EXISTS (SELECT 1 FROM examination_type_inspections ei WHERE ei.examination_type_id = et.id AND ei.name = '肝臓・脾臓');

INSERT INTO examination_type_inspections (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '結石・異物',          '認めない',                3
FROM examination_types et WHERE et.code = 'ET-005'
  AND NOT EXISTS (SELECT 1 FROM examination_type_inspections ei WHERE ei.examination_type_id = et.id AND ei.name = '結石・異物');

-- エコー検査の項目
INSERT INTO examination_type_inspections (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '肝臓エコー輝度',      '均一・正常',              1
FROM examination_types et WHERE et.code = 'ET-006'
  AND NOT EXISTS (SELECT 1 FROM examination_type_inspections ei WHERE ei.examination_type_id = et.id AND ei.name = '肝臓エコー輝度');

INSERT INTO examination_type_inspections (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '腎臓',                '大きさ・形態正常',        2
FROM examination_types et WHERE et.code = 'ET-006'
  AND NOT EXISTS (SELECT 1 FROM examination_type_inspections ei WHERE ei.examination_type_id = et.id AND ei.name = '腎臓');

INSERT INTO examination_type_inspections (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '膀胱',                '壁肥厚なし・結石なし',    3
FROM examination_types et WHERE et.code = 'ET-006'
  AND NOT EXISTS (SELECT 1 FROM examination_type_inspections ei WHERE ei.examination_type_id = et.id AND ei.name = '膀胱');

-- 心電図検査の項目
INSERT INTO examination_type_inspections (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '心拍数',              '犬60〜160 bpm / 猫120〜240 bpm', 1
FROM examination_types et WHERE et.code = 'ET-007'
  AND NOT EXISTS (SELECT 1 FROM examination_type_inspections ei WHERE ei.examination_type_id = et.id AND ei.name = '心拍数');

INSERT INTO examination_type_inspections (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, 'リズム',              '正常洞調律',              2
FROM examination_types et WHERE et.code = 'ET-007'
  AND NOT EXISTS (SELECT 1 FROM examination_type_inspections ei WHERE ei.examination_type_id = et.id AND ei.name = 'リズム');

INSERT INTO examination_type_inspections (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, 'P波・QRS波',          '形態正常',                3
FROM examination_types et WHERE et.code = 'ET-007'
  AND NOT EXISTS (SELECT 1 FROM examination_type_inspections ei WHERE ei.examination_type_id = et.id AND ei.name = 'P波・QRS波');

-- 眼科検査の項目
INSERT INTO examination_type_inspections (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '眼圧',                '10〜25 mmHg',             1
FROM examination_types et WHERE et.code = 'ET-008'
  AND NOT EXISTS (SELECT 1 FROM examination_type_inspections ei WHERE ei.examination_type_id = et.id AND ei.name = '眼圧');

INSERT INTO examination_type_inspections (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, 'シルマーティア試験',  '15 mm/min 以上',          2
FROM examination_types et WHERE et.code = 'ET-008'
  AND NOT EXISTS (SELECT 1 FROM examination_type_inspections ei WHERE ei.examination_type_id = et.id AND ei.name = 'シルマーティア試験');

INSERT INTO examination_type_inspections (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '角膜フルオレセイン染色', '染色なし',              3
FROM examination_types et WHERE et.code = 'ET-008'
  AND NOT EXISTS (SELECT 1 FROM examination_type_inspections ei WHERE ei.examination_type_id = et.id AND ei.name = '角膜フルオレセイン染色');

-- 皮膚検査の項目
INSERT INTO examination_type_inspections (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '皮膚掻爬検査',        '寄生虫・真菌なし',        1
FROM examination_types et WHERE et.code = 'ET-009'
  AND NOT EXISTS (SELECT 1 FROM examination_type_inspections ei WHERE ei.examination_type_id = et.id AND ei.name = '皮膚掻爬検査');

INSERT INTO examination_type_inspections (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '毛検査',              '異常なし',                2
FROM examination_types et WHERE et.code = 'ET-009'
  AND NOT EXISTS (SELECT 1 FROM examination_type_inspections ei WHERE ei.examination_type_id = et.id AND ei.name = '毛検査');

INSERT INTO examination_type_inspections (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, 'ウッド灯検査',        '蛍光なし',                3
FROM examination_types et WHERE et.code = 'ET-009'
  AND NOT EXISTS (SELECT 1 FROM examination_type_inspections ei WHERE ei.examination_type_id = et.id AND ei.name = 'ウッド灯検査');

-- 糞便検査の項目
INSERT INTO examination_type_inspections (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '寄生虫卵',            '検出なし',                1
FROM examination_types et WHERE et.code = 'ET-010'
  AND NOT EXISTS (SELECT 1 FROM examination_type_inspections ei WHERE ei.examination_type_id = et.id AND ei.name = '寄生虫卵');

INSERT INTO examination_type_inspections (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, '細菌（グラム染色）',  '正常腸内細菌叢',          2
FROM examination_types et WHERE et.code = 'ET-010'
  AND NOT EXISTS (SELECT 1 FROM examination_type_inspections ei WHERE ei.examination_type_id = et.id AND ei.name = '細菌（グラム染色）');

INSERT INTO examination_type_inspections (id, examination_type_id, name, normal_value, sort_order)
SELECT gen_random_uuid(), et.id, 'ジアルジア・コクシジウム', '陰性',               3
FROM examination_types et WHERE et.code = 'ET-010'
  AND NOT EXISTS (SELECT 1 FROM examination_type_inspections ei WHERE ei.examination_type_id = et.id AND ei.name = 'ジアルジア・コクシジウム');

-- -----------------------------------------------------------------------------
-- 4. master_items - vaccines（ワクチン）
--    実テーブル: master_items (category='vaccine')
--    カラム: id, code, name, category, price, status, description, species, interval, sort_order
--    ※ codeにUNIQUE制約なし -> NOT EXISTSで冪等性を保証
-- -----------------------------------------------------------------------------
INSERT INTO master_items (id, code, name, category, price, status, description, species, interval, sort_order)
SELECT gen_random_uuid(), 'VAC-007', '混合ワクチン8種（犬）', 'vaccine', 6000.00, 'active', '5種にレプトスピラ3型を加えた8種混合ワクチン', 'dog', '1year', 7
WHERE NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'VAC-007');

INSERT INTO master_items (id, code, name, category, price, status, description, species, interval, sort_order)
SELECT gen_random_uuid(), 'VAC-008', '混合ワクチン5種（猫）', 'vaccine', 5000.00, 'active', '3種に猫白血病・クラミジアを含む5種混合（拡張型）', 'cat', '1year', 8
WHERE NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'VAC-008');

INSERT INTO master_items (id, code, name, category, price, status, description, species, interval, sort_order)
SELECT gen_random_uuid(), 'VAC-009', 'ノミ・マダニ予防薬', 'vaccine', 1500.00, 'active', '経口または外用による月1回投与の予防薬', 'both', '', 9
WHERE NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'VAC-009');

INSERT INTO master_items (id, code, name, category, price, status, description, species, interval, sort_order)
SELECT gen_random_uuid(), 'VAC-010', 'フィラリア予防薬', 'vaccine', 800.00, 'active', '毎月1回投与・蚊のいる季節（5〜12月）に実施', 'dog', '1year', 10
WHERE NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'VAC-010');

-- -----------------------------------------------------------------------------
-- 5. master_items - serviceType（サービス種別・予約区分）
--    category='serviceType'
-- -----------------------------------------------------------------------------
INSERT INTO master_items (id, code, name, category, status, description, sort_order)
SELECT gen_random_uuid(), 'SRVTYPE-006', '健康診断', 'serviceType', 'active', '定期健診・シニア検診', 6
WHERE NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'SRVTYPE-006');

INSERT INTO master_items (id, code, name, category, status, description, sort_order)
SELECT gen_random_uuid(), 'SRVTYPE-007', '緊急診察', 'serviceType', 'active', '時間外・緊急対応', 7
WHERE NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'SRVTYPE-007');

-- -----------------------------------------------------------------------------
-- 6. master_items - consultation（診察項目）
--    category='consultation', time_condition カラム使用
-- -----------------------------------------------------------------------------
INSERT INTO master_items (id, code, name, category, price, status, time_condition, description, sort_order)
SELECT gen_random_uuid(), 'CONS-005', '救急診察料', 'consultation', 8000.00, 'active', '救急', '緊急処置・救急対応が必要な診察', 5
WHERE NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'CONS-005');

INSERT INTO master_items (id, code, name, category, price, status, time_condition, description, sort_order)
SELECT gen_random_uuid(), 'CONS-006', '電話相談料', 'consultation', 500.00, 'active', '', '電話による症状相談・診察アドバイス', 6
WHERE NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'CONS-006');

INSERT INTO master_items (id, code, name, category, price, status, time_condition, description, sort_order)
SELECT gen_random_uuid(), 'CONS-007', '皮膚科診察', 'consultation', 2000.00, 'active', '', '皮膚疾患専門診察（皮膚検査含む）', 7
WHERE NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'CONS-007');

INSERT INTO master_items (id, code, name, category, price, status, time_condition, description, sort_order)
SELECT gen_random_uuid(), 'CONS-008', '眼科診察', 'consultation', 2000.00, 'active', '', '眼科専門診察（眼圧測定含む）', 8
WHERE NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'CONS-008');

-- -----------------------------------------------------------------------------
-- 7. master_items - procedure（処置項目）
--    category='procedure', anesthesia カラム使用
-- -----------------------------------------------------------------------------
INSERT INTO master_items (id, code, name, category, price, status, anesthesia, description, sort_order)
SELECT gen_random_uuid(), 'PROC-009', '点滴（1時間）', 'procedure', 3000.00, 'active', 'none', '持続点滴（1時間あたり）', 9
WHERE NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'PROC-009');

INSERT INTO master_items (id, code, name, category, price, status, anesthesia, description, sort_order)
SELECT gen_random_uuid(), 'PROC-010', '外科縫合（小）', 'procedure', 5000.00, 'active', 'local', '局所麻酔下での小創傷縫合', 10
WHERE NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'PROC-010');

INSERT INTO master_items (id, code, name, category, price, status, anesthesia, description, sort_order)
SELECT gen_random_uuid(), 'PROC-011', '外科縫合（中）', 'procedure', 10000.00, 'active', 'general', '全身麻酔下での中程度創傷縫合', 11
WHERE NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'PROC-011');

INSERT INTO master_items (id, code, name, category, price, status, anesthesia, description, sort_order)
SELECT gen_random_uuid(), 'PROC-012', '去勢手術（犬）', 'procedure', 30000.00, 'active', 'general', '雄犬の精巣摘出術（全身麻酔・術後管理含む）', 12
WHERE NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'PROC-012');

INSERT INTO master_items (id, code, name, category, price, status, anesthesia, description, sort_order)
SELECT gen_random_uuid(), 'PROC-013', '避妊手術（犬）', 'procedure', 50000.00, 'active', 'general', '雌犬の子宮卵巣摘出術（全身麻酔・術後管理含む）', 13
WHERE NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'PROC-013');

INSERT INTO master_items (id, code, name, category, price, status, anesthesia, description, sort_order)
SELECT gen_random_uuid(), 'PROC-014', '去勢手術（猫）', 'procedure', 20000.00, 'active', 'general', '雄猫の精巣摘出術（全身麻酔・術後管理含む）', 14
WHERE NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'PROC-014');

INSERT INTO master_items (id, code, name, category, price, status, anesthesia, description, sort_order)
SELECT gen_random_uuid(), 'PROC-015', '避妊手術（猫）', 'procedure', 35000.00, 'active', 'general', '雌猫の子宮卵巣摘出術（全身麻酔・術後管理含む）', 15
WHERE NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'PROC-015');

INSERT INTO master_items (id, code, name, category, price, status, anesthesia, description, sort_order)
SELECT gen_random_uuid(), 'PROC-016', '浣腸', 'procedure', 1500.00, 'active', 'none', '便秘・消化管閉塞時の浣腸処置', 16
WHERE NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'PROC-016');

-- -----------------------------------------------------------------------------
-- 8. master_items - hospitalization（入院プラン）
--    category='hospitalization'
-- -----------------------------------------------------------------------------
INSERT INTO master_items (id, code, name, category, price, status, description, sort_order)
SELECT gen_random_uuid(), 'HOSP-004', '入院（中型犬）', 'hospitalization', 7000.00, 'active', '10〜25kgの中型犬の入院（24時間ケア込み）', 4
WHERE NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'HOSP-004');

INSERT INTO master_items (id, code, name, category, price, status, description, sort_order)
SELECT gen_random_uuid(), 'HOSP-005', '入院（大型犬）', 'hospitalization', 9000.00, 'active', '25kg超の大型犬の入院（24時間ケア込み）', 5
WHERE NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'HOSP-005');

INSERT INTO master_items (id, code, name, category, price, status, description, sort_order)
SELECT gen_random_uuid(), 'HOSP-006', 'ホテル（小型犬・猫）', 'hospitalization', 3000.00, 'active', '10kg以下の小型犬・猫のペットホテル（1泊）', 6
WHERE NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'HOSP-006');

INSERT INTO master_items (id, code, name, category, price, status, description, sort_order)
SELECT gen_random_uuid(), 'HOSP-007', 'ホテル（中型犬）', 'hospitalization', 4000.00, 'active', '10〜25kgの中型犬のペットホテル（1泊）', 7
WHERE NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'HOSP-007');

INSERT INTO master_items (id, code, name, category, price, status, description, sort_order)
SELECT gen_random_uuid(), 'HOSP-008', 'ホテル（大型犬）', 'hospitalization', 5000.00, 'active', '25kg超の大型犬のペットホテル（1泊）', 8
WHERE NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'HOSP-008');

-- -----------------------------------------------------------------------------
-- 9. master_items - diagnosis_category（診断カテゴリ）
--    category='diagnosis_category'
-- -----------------------------------------------------------------------------
INSERT INTO master_items (id, code, name, category, status, description, sort_order)
SELECT gen_random_uuid(), 'DIAGCAT-005', '内分泌・代謝疾患', 'diagnosis_category', 'active', '糖尿病・甲状腺・副腎疾患等の内分泌代謝疾患', 5
WHERE NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'DIAGCAT-005');

INSERT INTO master_items (id, code, name, category, status, description, sort_order)
SELECT gen_random_uuid(), 'DIAGCAT-006', '泌尿器科', 'diagnosis_category', 'active', '膀胱炎・尿路結石・腎疾患等の泌尿器系疾患', 6
WHERE NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'DIAGCAT-006');

INSERT INTO master_items (id, code, name, category, status, description, sort_order)
SELECT gen_random_uuid(), 'DIAGCAT-007', '神経科', 'diagnosis_category', 'active', 'てんかん・脊髄疾患等の神経系疾患', 7
WHERE NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'DIAGCAT-007');

INSERT INTO master_items (id, code, name, category, status, description, sort_order)
SELECT gen_random_uuid(), 'DIAGCAT-008', '眼科', 'diagnosis_category', 'active', '結膜炎・白内障・緑内障等の眼科疾患', 8
WHERE NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'DIAGCAT-008');

INSERT INTO master_items (id, code, name, category, status, description, sort_order)
SELECT gen_random_uuid(), 'DIAGCAT-009', '耳科', 'diagnosis_category', 'active', '外耳炎・中耳炎等の耳科疾患', 9
WHERE NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'DIAGCAT-009');

INSERT INTO master_items (id, code, name, category, status, description, sort_order)
SELECT gen_random_uuid(), 'DIAGCAT-010', '感染症', 'diagnosis_category', 'active', 'ウイルス・細菌・寄生虫による感染症', 10
WHERE NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'DIAGCAT-010');

-- -----------------------------------------------------------------------------
-- 10. master_items - diagnosis_name（診断名）
--     category='diagnosis_name', parent_id は diagnosis_category のIDを参照
-- -----------------------------------------------------------------------------

-- 内分泌・代謝疾患 の診断名
INSERT INTO master_items (id, code, name, category, status, parent_id, sort_order)
SELECT gen_random_uuid(), 'DIAG-016', '糖尿病', 'diagnosis_name', 'active', mi.id, 1
FROM master_items mi WHERE mi.code = 'DIAGCAT-005'
  AND NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'DIAG-016');

INSERT INTO master_items (id, code, name, category, status, parent_id, sort_order)
SELECT gen_random_uuid(), 'DIAG-017', '甲状腺機能低下症', 'diagnosis_name', 'active', mi.id, 2
FROM master_items mi WHERE mi.code = 'DIAGCAT-005'
  AND NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'DIAG-017');

INSERT INTO master_items (id, code, name, category, status, parent_id, sort_order)
SELECT gen_random_uuid(), 'DIAG-018', 'クッシング症候群（副腎皮質機能亢進症）', 'diagnosis_name', 'active', mi.id, 3
FROM master_items mi WHERE mi.code = 'DIAGCAT-005'
  AND NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'DIAG-018');

-- 泌尿器科 の診断名
INSERT INTO master_items (id, code, name, category, status, parent_id, sort_order)
SELECT gen_random_uuid(), 'DIAG-019', '膀胱炎', 'diagnosis_name', 'active', mi.id, 1
FROM master_items mi WHERE mi.code = 'DIAGCAT-006'
  AND NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'DIAG-019');

INSERT INTO master_items (id, code, name, category, status, parent_id, sort_order)
SELECT gen_random_uuid(), 'DIAG-020', '尿路結石', 'diagnosis_name', 'active', mi.id, 2
FROM master_items mi WHERE mi.code = 'DIAGCAT-006'
  AND NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'DIAG-020');

INSERT INTO master_items (id, code, name, category, status, parent_id, sort_order)
SELECT gen_random_uuid(), 'DIAG-021', '慢性腎臓病', 'diagnosis_name', 'active', mi.id, 3
FROM master_items mi WHERE mi.code = 'DIAGCAT-006'
  AND NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'DIAG-021');

-- 神経科 の診断名
INSERT INTO master_items (id, code, name, category, status, parent_id, sort_order)
SELECT gen_random_uuid(), 'DIAG-022', 'てんかん', 'diagnosis_name', 'active', mi.id, 1
FROM master_items mi WHERE mi.code = 'DIAGCAT-007'
  AND NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'DIAG-022');

INSERT INTO master_items (id, code, name, category, status, parent_id, sort_order)
SELECT gen_random_uuid(), 'DIAG-023', '椎間板ヘルニア', 'diagnosis_name', 'active', mi.id, 2
FROM master_items mi WHERE mi.code = 'DIAGCAT-007'
  AND NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'DIAG-023');

-- 眼科 の診断名
INSERT INTO master_items (id, code, name, category, status, parent_id, sort_order)
SELECT gen_random_uuid(), 'DIAG-024', '結膜炎', 'diagnosis_name', 'active', mi.id, 1
FROM master_items mi WHERE mi.code = 'DIAGCAT-008'
  AND NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'DIAG-024');

INSERT INTO master_items (id, code, name, category, status, parent_id, sort_order)
SELECT gen_random_uuid(), 'DIAG-025', '白内障', 'diagnosis_name', 'active', mi.id, 2
FROM master_items mi WHERE mi.code = 'DIAGCAT-008'
  AND NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'DIAG-025');

INSERT INTO master_items (id, code, name, category, status, parent_id, sort_order)
SELECT gen_random_uuid(), 'DIAG-026', '緑内障', 'diagnosis_name', 'active', mi.id, 3
FROM master_items mi WHERE mi.code = 'DIAGCAT-008'
  AND NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'DIAG-026');

-- 耳科 の診断名
INSERT INTO master_items (id, code, name, category, status, parent_id, sort_order)
SELECT gen_random_uuid(), 'DIAG-027', '外耳炎', 'diagnosis_name', 'active', mi.id, 1
FROM master_items mi WHERE mi.code = 'DIAGCAT-009'
  AND NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'DIAG-027');

INSERT INTO master_items (id, code, name, category, status, parent_id, sort_order)
SELECT gen_random_uuid(), 'DIAG-028', '中耳炎', 'diagnosis_name', 'active', mi.id, 2
FROM master_items mi WHERE mi.code = 'DIAGCAT-009'
  AND NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'DIAG-028');

-- 感染症 の診断名
INSERT INTO master_items (id, code, name, category, status, parent_id, sort_order)
SELECT gen_random_uuid(), 'DIAG-029', 'パルボウイルス感染症', 'diagnosis_name', 'active', mi.id, 1
FROM master_items mi WHERE mi.code = 'DIAGCAT-010'
  AND NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'DIAG-029');

INSERT INTO master_items (id, code, name, category, status, parent_id, sort_order)
SELECT gen_random_uuid(), 'DIAG-030', '猫白血病ウイルス感染症（FeLV）', 'diagnosis_name', 'active', mi.id, 2
FROM master_items mi WHERE mi.code = 'DIAGCAT-010'
  AND NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'DIAG-030');

INSERT INTO master_items (id, code, name, category, status, parent_id, sort_order)
SELECT gen_random_uuid(), 'DIAG-031', '猫免疫不全ウイルス感染症（FIV）', 'diagnosis_name', 'active', mi.id, 3
FROM master_items mi WHERE mi.code = 'DIAGCAT-010'
  AND NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'DIAG-031');

-- -----------------------------------------------------------------------------
-- 11. master_items - checkup（健診種別）
--     category='checkup', target_age カラム使用
-- -----------------------------------------------------------------------------
INSERT INTO master_items (id, code, name, category, price, status, interval, target_age, description, sort_order)
SELECT gen_random_uuid(), 'CHECKUP-004', '歯科検診', 'checkup', 2000.00, 'active', '1year', '', '口腔内チェック・歯垢・歯石の評価', 4
WHERE NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'CHECKUP-004');

INSERT INTO master_items (id, code, name, category, price, status, interval, target_age, description, sort_order)
SELECT gen_random_uuid(), 'CHECKUP-005', '心臓検診', 'checkup', 5000.00, 'active', '', '', '聴診・心電図・胸部X線による心臓専門検診', 5
WHERE NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'CHECKUP-005');

INSERT INTO master_items (id, code, name, category, price, status, interval, target_age, description, sort_order)
SELECT gen_random_uuid(), 'CHECKUP-006', '子犬・子猫健康診断', 'checkup', 3000.00, 'active', '', '1歳未満', '初回ワクチン前後の身体検査・寄生虫検査', 6
WHERE NOT EXISTS (SELECT 1 FROM master_items WHERE code = 'CHECKUP-006');

-- -----------------------------------------------------------------------------
-- 12. medicines（薬剤）
--    実テーブル: medicines
--    カラム: id, code, name, dosage_form, medicine_unit, price, default_quantity,
--            inventory_id, description, is_active, sort_order
-- -----------------------------------------------------------------------------
INSERT INTO medicines (id, code, name, dosage_form, medicine_unit, price, default_quantity, description, is_active, sort_order) VALUES
    (gen_random_uuid(), 'MED-009', 'エンロフロキサシン錠',           'tablet',    'per_tablet',    80.00, 1, 'フルオロキノロン系抗菌薬・尿路・皮膚・消化器感染に使用',     true,  9),
    (gen_random_uuid(), 'MED-010', 'プレドニゾロン錠',               'tablet',    'per_tablet',    30.00, 1, 'ステロイド系抗炎症薬・アレルギー・免疫疾患に使用',           true, 10),
    (gen_random_uuid(), 'MED-011', 'フロセミド錠',                   'tablet',    'per_tablet',    20.00, 1, 'ループ利尿薬・心疾患・腎疾患による浮腫に使用',               true, 11),
    (gen_random_uuid(), 'MED-012', 'ビタミンB12注射液',              'injection', 'per_dose',     500.00, 1, '貧血・神経疾患・食欲不振時の補助療法',                       true, 12),
    (gen_random_uuid(), 'MED-013', 'クロルヘキシジン消毒液',         'liquid',    'per_ml',        10.00, 10, '創傷・皮膚・耳道の消毒・洗浄に使用',                         true, 13),
    (gen_random_uuid(), 'MED-014', 'ネクスガード（ノミ・マダニ）',   'tablet',    'per_tablet', 1500.00, 1, 'アフォキソラネル製剤・ノミ・マダニ駆除（月1回経口投与）',    true, 14),
    (gen_random_uuid(), 'MED-015', 'モキシデクチン',                 'liquid',    'per_dose',     800.00, 1, 'フィラリア・回虫・鉤虫予防に使用',                           true, 15),
    (gen_random_uuid(), 'MED-016', 'フィプロスポット（ノミ予防）',   'topical',   'per_dose',    1200.00, 1, 'フィプロニル製剤・ノミ・マダニ駆除用スポットオン',           true, 16),
    (gen_random_uuid(), 'MED-017', '酢酸メチルプレドニゾロン注射',  'injection', 'per_dose',     800.00, 1, '長時間作用型注射ステロイド・アレルギー・関節炎に使用',       true, 17),
    (gen_random_uuid(), 'MED-018', '硫酸アトロピン点眼液',           'liquid',    'per_ml',       200.00, 1, '散瞳・眼底検査・虹彩毛様体炎治療に使用',                     true, 18),
    (gen_random_uuid(), 'MED-019', 'テトラサイクリン眼軟膏',         'topical',   'per_dose',     600.00, 1, '細菌性結膜炎・角膜炎の治療に使用',                           true, 19),
    (gen_random_uuid(), 'MED-020', 'カルプロフェン（NSAIDs）',       'tablet',    'per_tablet',   100.00, 1, 'NSAIDs系鎮痛消炎薬・術後疼痛・骨関節炎に使用',              true, 20)
ON CONFLICT (code) DO NOTHING;

-- -----------------------------------------------------------------------------
-- 13. insurance_companies（保険会社）
--    実テーブル: insurance_companies
--    カラム: id, code, name, coverage_rate, contact_phone, description, is_active, sort_order
-- -----------------------------------------------------------------------------
INSERT INTO insurance_companies (id, code, name, coverage_rate, contact_phone, description, is_active, sort_order) VALUES
    (gen_random_uuid(), 'INS-004', 'ペット&ファミリー損害保険', '50', '0120-81-8320',  'げんきナンバーわんスリム等を取り扱い',  true, 4),
    (gen_random_uuid(), 'INS-005', '楽天ペット保険',            '70', '0800-600-0204', '楽天グループのペット保険サービス',      true, 5)
ON CONFLICT (code) DO NOTHING;

-- -----------------------------------------------------------------------------
-- 14. cages（ケージ）
--    実テーブル: cages
--    カラム: id, code, name, cage_type, cage_size, body_size, billing_unit,
--            price, is_active, sort_order
-- -----------------------------------------------------------------------------
INSERT INTO cages (id, code, name, cage_type, cage_size, price, is_active, sort_order) VALUES
    (gen_random_uuid(), 'CAGE-007', 'ICUケージ大',   'icu',     'large',  0.00, true,  7),
    (gen_random_uuid(), 'CAGE-008', 'ICUケージ中',   'icu',     'medium', 0.00, true,  8),
    (gen_random_uuid(), 'CAGE-009', '犬用ケージ大1', 'dog',     'large',  0.00, true,  9),
    (gen_random_uuid(), 'CAGE-010', '犬用ケージ大2', 'dog',     'large',  0.00, true, 10),
    (gen_random_uuid(), 'CAGE-011', '犬用ケージ中1', 'dog',     'medium', 0.00, true, 11),
    (gen_random_uuid(), 'CAGE-012', '犬用ケージ中2', 'dog',     'medium', 0.00, true, 12),
    (gen_random_uuid(), 'CAGE-013', '犬用ケージ小1', 'dog',     'small',  0.00, true, 13),
    (gen_random_uuid(), 'CAGE-014', '猫用ケージ1',   'cat',     'small',  0.00, true, 14),
    (gen_random_uuid(), 'CAGE-015', '猫用ケージ2',   'cat',     'small',  0.00, true, 15),
    (gen_random_uuid(), 'CAGE-016', '猫用ケージ3',   'cat',     'small',  0.00, true, 16),
    (gen_random_uuid(), 'CAGE-017', '一般ケージ大',  'general', 'large',  0.00, true, 17),
    (gen_random_uuid(), 'CAGE-018', '一般ケージ中',  'general', 'medium', 0.00, true, 18)
ON CONFLICT (code) DO NOTHING;

-- -----------------------------------------------------------------------------
-- 15. trimming_courses（トリミングコース）
--    実テーブル: trimming_courses
--    カラム: id, code, name, target_size, duration, price, parent_id, description, is_active, sort_order
-- -----------------------------------------------------------------------------
INSERT INTO trimming_courses (id, code, name, target_size, duration, price, description, is_active, sort_order) VALUES
    (gen_random_uuid(), 'TRIM-006', '猫フルコース（スタンダード）', 'cat',    '90分',  6500.00, 'シャンプー・ドライ・爪切り・耳掃除・ブラッシング込み', true, 6),
    (gen_random_uuid(), 'TRIM-007', '小型犬プレミアムコース',       'small',  '75分',  5000.00, 'フルトリミング＋歯磨き・肛門腺絞り付き',              true, 7),
    (gen_random_uuid(), 'TRIM-008', '中型犬プレミアムコース',       'medium', '105分', 7500.00, 'フルトリミング＋歯磨き・肛門腺絞り付き',              true, 8),
    (gen_random_uuid(), 'TRIM-009', '大型犬プレミアムコース',       'large',  '150分', 11000.00, 'フルトリミング＋歯磨き・肛門腺絞り付き',             true, 9),
    (gen_random_uuid(), 'TRIM-010', 'シャンプー＆ブラッシング',     'small',  '45分',  3000.00, 'シャンプー・ドライ・ブラッシングのみ（中型犬以上は別途加算）', true, 10)
ON CONFLICT (code) DO NOTHING;

-- -----------------------------------------------------------------------------
-- 16. trimming_options（トリミングオプション）
--    実テーブル: trimming_options
--    カラム: id, code, name, target_size, combinable, price, parent_id, description, is_active, sort_order
-- -----------------------------------------------------------------------------
INSERT INTO trimming_options (id, code, name, combinable, price, description, is_active, sort_order) VALUES
    (gen_random_uuid(), 'TRIMOPT-006', '薬浴シャンプー',   'yes', 2000.00, '皮膚疾患・アレルギー用薬用シャンプーで洗浄',        true, 6),
    (gen_random_uuid(), 'TRIMOPT-007', 'ノミ取りシャンプー', 'yes', 1500.00, 'ノミ駆除成分配合シャンプーによる洗浄',            true, 7),
    (gen_random_uuid(), 'TRIMOPT-008', '毛並み整えブラッシング', 'yes', 500.00, '丁寧なブラッシングと毛並み整え',                true, 8)
ON CONFLICT (code) DO NOTHING;

-- -----------------------------------------------------------------------------
-- 17. clinic_info（病院情報）
--    実テーブル: clinic_info（1レコードのみ想定）
--    カラム: id, name, branch_name, postal_code, address, phone_number,
--            fax_number, registration_number, director_name, email, website
-- -----------------------------------------------------------------------------
INSERT INTO clinic_info (id, name, branch_name, postal_code, address, phone_number, fax_number, registration_number, director_name, email, website)
SELECT
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
WHERE NOT EXISTS (SELECT 1 FROM clinic_info);

-- =============================================================================
-- シード完了
-- =============================================================================
