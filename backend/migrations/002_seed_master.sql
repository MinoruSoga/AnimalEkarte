-- =============================================================================
-- Animal Ekarte - マスタデータシード (新スキーマ対応版)
-- PostgreSQL 18
-- 冪等性保証: ON CONFLICT DO NOTHING
-- 対象スキーマ: 001_init.sql (45テーブル・専用マスタテーブル版)
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
-- 2. service_types（サービス種別）
-- -----------------------------------------------------------------------------
INSERT INTO service_types (id, code, name, status, description, color, sort_order) VALUES
    (gen_random_uuid(), 'SV-001', '一般診療',     'active', '内科・外科・皮膚科などの一般的な診療',   '#3B82F6', 1),
    (gen_random_uuid(), 'SV-002', 'ワクチン接種', 'active', '各種ワクチン接種（予防接種）',           '#10B981', 2),
    (gen_random_uuid(), 'SV-003', '健康診断',     'active', '定期健康診断・フィラリア検査など',       '#8B5CF6', 3),
    (gen_random_uuid(), 'SV-004', '手術・処置',   'active', '去勢・避妊・その他外科手術',             '#EF4444', 4),
    (gen_random_uuid(), 'SV-005', 'トリミング',   'active', 'グルーミング・爪切り・耳掃除など',       '#F59E0B', 5),
    (gen_random_uuid(), 'SV-006', '入院',         'active', '入院・ホテル管理',                       '#6B7280', 6),
    (gen_random_uuid(), 'SV-007', '検査',         'active', '血液検査・尿検査・画像診断など',         '#EC4899', 7),
    (gen_random_uuid(), 'SV-008', '再診',         'active', '前回診察の経過確認・投薬管理',           '#06B6D4', 8)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 3. consultations（診察項目）
-- -----------------------------------------------------------------------------
INSERT INTO consultations (id, code, name, price, status, description, sort_order) VALUES
    (gen_random_uuid(), 'C-001', '初診料',         2000.00, 'active', '初めての受診または6ヶ月以上受診がない場合', 1),
    (gen_random_uuid(), 'C-002', '再診料',          800.00, 'active', '継続通院の診察料',                         2),
    (gen_random_uuid(), 'C-003', '往診料',         5000.00, 'active', '自宅への往診料（基本料金）',               3),
    (gen_random_uuid(), 'C-004', '時間外診療料',   3000.00, 'active', '診療時間外・休日の緊急診察',               4),
    (gen_random_uuid(), 'C-005', '電話相談料',      500.00, 'active', '電話による診察相談',                       5)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 4. procedures（処置項目）
-- -----------------------------------------------------------------------------
INSERT INTO procedures (id, code, name, price, status, description, anesthesia, sort_order) VALUES
    (gen_random_uuid(), 'P-001', '去勢手術（犬）',         25000.00, 'active', '雄犬の去勢手術',                         'general', 1),
    (gen_random_uuid(), 'P-002', '去勢手術（猫）',         20000.00, 'active', '雄猫の去勢手術',                         'general', 2),
    (gen_random_uuid(), 'P-003', '避妊手術（犬）',         35000.00, 'active', '雌犬の避妊手術',                         'general', 3),
    (gen_random_uuid(), 'P-004', '避妊手術（猫）',         25000.00, 'active', '雌猫の避妊手術',                         'general', 4),
    (gen_random_uuid(), 'P-005', '歯石除去',               15000.00, 'active', '全身麻酔下での歯石除去・歯周治療',       'general', 5),
    (gen_random_uuid(), 'P-006', '腫瘍摘出（皮膚）',       20000.00, 'active', '皮膚腫瘍の外科的摘出',                   'local',   6),
    (gen_random_uuid(), 'P-007', '骨折手術',               80000.00, 'active', '骨折の外科的整復・固定',                 'general', 7),
    (gen_random_uuid(), 'P-008', '傷口縫合',                5000.00, 'active', '裂傷・切傷の縫合処置',                   'local',   8),
    (gen_random_uuid(), 'P-009', '点滴（静脈内）',          3000.00, 'active', '静脈内点滴（1時間）',                    'none',    9),
    (gen_random_uuid(), 'P-010', '注射（皮下・筋肉）',      1000.00, 'active', '皮下注射・筋肉注射',                     'none',   10),
    (gen_random_uuid(), 'P-011', '耳道洗浄',                2500.00, 'active', '外耳炎治療・耳道内の洗浄処置',           'none',   11),
    (gen_random_uuid(), 'P-012', '膀胱洗浄',                3000.00, 'active', '膀胱内の洗浄処置',                       'none',   12),
    (gen_random_uuid(), 'P-013', '肛門嚢絞り',               500.00, 'active', '肛門嚢の分泌液除去',                     'none',   13),
    (gen_random_uuid(), 'P-014', '爪切り（医療）',           500.00, 'active', '医療目的の爪切り・巻き爪処置',           'none',   14),
    (gen_random_uuid(), 'P-015', 'カテーテル挿入',          2000.00, 'active', '尿路閉塞・採尿のカテーテル挿入',         'none',   15)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 5. vaccines（ワクチン）
-- -----------------------------------------------------------------------------
INSERT INTO vaccines (id, code, name, price, status, description, species, interval, sort_order) VALUES
    (gen_random_uuid(), 'V-001', '混合ワクチン5種（犬）',   4500.00, 'active', 'ジステンパー・パルボ・アデノ1・アデノ2・パラインフルエンザ', 'dog', '1年', 1),
    (gen_random_uuid(), 'V-002', '混合ワクチン8種（犬）',   6500.00, 'active', '5種＋レプトスピラ3種',                                          'dog', '1年', 2),
    (gen_random_uuid(), 'V-003', '混合ワクチン10種（犬）',  8000.00, 'active', '5種＋レプトスピラ5種',                                          'dog', '1年', 3),
    (gen_random_uuid(), 'V-004', '狂犬病ワクチン（犬）',    3000.00, 'active', '狂犬病予防法に基づく接種',                                      'dog', '1年', 4),
    (gen_random_uuid(), 'V-005', '混合ワクチン3種（猫）',   4000.00, 'active', '猫ウイルス性鼻気管炎・カリシウイルス・汎白血球減少症',          'cat', '1年', 5),
    (gen_random_uuid(), 'V-006', '混合ワクチン5種（猫）',   5500.00, 'active', '3種＋猫白血病・猫クラミジア',                                   'cat', '1年', 6),
    (gen_random_uuid(), 'V-007', '猫白血病ワクチン',        4000.00, 'active', '猫白血病ウイルス感染症予防',                                    'cat', '1年', 7),
    (gen_random_uuid(), 'V-008', '猫エイズワクチン',        5000.00, 'active', '猫免疫不全ウイルス感染症予防',                                  'cat', '1年', 8)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 6. medicines（薬剤）
-- -----------------------------------------------------------------------------
INSERT INTO medicines (id, code, name, price, status, description, dosage_form, medicine_unit, default_quantity, sort_order) VALUES
    (gen_random_uuid(), 'M-001', 'アモキシシリン 50mg',       500.00, 'active', '広域スペクトラム抗生物質',                 'tablet',    'per_tablet', 1,   1),
    (gen_random_uuid(), 'M-002', 'メトロニダゾール 250mg',    600.00, 'active', '嫌気性菌・原虫感染症治療薬',               'tablet',    'per_tablet', 1,   2),
    (gen_random_uuid(), 'M-003', 'プレドニゾロン 5mg',        400.00, 'active', 'ステロイド系抗炎症・免疫抑制剤',           'tablet',    'per_tablet', 1,   3),
    (gen_random_uuid(), 'M-004', 'フロセミド注射液 20mg/2ml', 800.00, 'active', '利尿剤（心臓・腎臓病の浮腫治療）',        'injection', 'per_ml',     2,   4),
    (gen_random_uuid(), 'M-005', 'メロキシカム経口液',        700.00, 'active', 'NSAIDs・痛み・炎症の緩和',                 'liquid',    'per_ml',     1,   5),
    (gen_random_uuid(), 'M-006', 'ガバペンチン 100mg',        550.00, 'active', '神経因性疼痛・てんかん補助療法',           'tablet',    'per_tablet', 1,   6),
    (gen_random_uuid(), 'M-007', 'フィラリア予防薬（小型犬）',900.00, 'active', '体重10kg以下犬用フィラリア予防',           'tablet',    'per_tablet', 1,   7),
    (gen_random_uuid(), 'M-008', 'フィラリア予防薬（中型犬）',1100.00,'active', '体重11-25kg犬用フィラリア予防',            'tablet',    'per_tablet', 1,   8),
    (gen_random_uuid(), 'M-009', 'ノミ・ダニ駆除薬（犬）',   2500.00, 'active', '外部寄生虫予防・駆除（スポットオン）',    'topical',   'per_dose',   1,   9),
    (gen_random_uuid(), 'M-010', 'ノミ・ダニ駆除薬（猫）',   2500.00, 'active', '外部寄生虫予防・駆除（スポットオン）',    'topical',   'per_dose',   1,  10),
    (gen_random_uuid(), 'M-011', 'マロピタント 16mg',          800.00, 'active', '制吐剤（乗り物酔い・嘔吐治療）',           'tablet',    'per_tablet', 1,  11),
    (gen_random_uuid(), 'M-012', 'ラクツロース液',             500.00, 'active', '便秘・肝性脳症の治療',                     'liquid',    'per_ml',     5,  12),
    (gen_random_uuid(), 'M-013', '抗生剤点眼薬',               600.00, 'active', '眼科用抗菌点眼剤',                         'liquid',    'per_ml',     1,  13),
    (gen_random_uuid(), 'M-014', 'デキサメタゾン注射液',       700.00, 'active', '強力ステロイド・アレルギー緊急治療',       'injection', 'per_ml',     1,  14),
    (gen_random_uuid(), 'M-015', '生理食塩水 500ml',           400.00, 'active', '点滴・洗浄用生理食塩水',                   'liquid',    'per_ml',   500,  15)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 7. insurances（ペット保険）
-- -----------------------------------------------------------------------------
INSERT INTO insurances (id, code, name, status, description, coverage_rate, contact_phone, sort_order) VALUES
    (gen_random_uuid(), 'I-001', 'アニコム損保',      'active', 'ペット保険大手・どうぶつ健保シリーズ',  '70', '0120-025-034',  1),
    (gen_random_uuid(), 'I-002', 'アイペット損保',    'active', 'うちの子シリーズ',                      '70', '0120-956-099',  2),
    (gen_random_uuid(), 'I-003', 'ペット&ファミリー', 'active', 'げんきナンバーワンシリーズ',             '70', '0120-81-8505',  3),
    (gen_random_uuid(), 'I-004', 'ドコモペット保険',  'active', 'ドコモが提供するペット保険',             '70', '0120-001-731',  4),
    (gen_random_uuid(), 'I-005', 'SBIいきいき少短',   'active', 'SBIグループのペット保険',               '80', '0800-888-0819', 5),
    (gen_random_uuid(), 'I-006', 'PS保険',            'active', 'プリズムコール・70%補償プラン',          '70', '0120-099-317',  6),
    (gen_random_uuid(), 'I-007', 'FPC',               'active', 'ファミリー少額短期保険',                '50', '0120-210-616',  7),
    (gen_random_uuid(), 'I-008', '自費（保険なし）',  'active', '保険未加入・全額自費',                   '50', '',              8)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 8. cages（ケージ）
-- -----------------------------------------------------------------------------
INSERT INTO cages (id, code, name, price, status, description, cage_type, cage_size, sort_order) VALUES
    (gen_random_uuid(), 'CG-001', 'ICUケージA',      8000.00, 'active', '酸素吸入可・重症患者用',    'icu',     'medium', 1),
    (gen_random_uuid(), 'CG-002', 'ICUケージB',      8000.00, 'active', '酸素吸入可・重症患者用',    'icu',     'medium', 2),
    (gen_random_uuid(), 'CG-003', '犬用大型ケージA', 4000.00, 'active', '大型犬・術後管理用',         'dog',     'large',  3),
    (gen_random_uuid(), 'CG-004', '犬用大型ケージB', 4000.00, 'active', '大型犬・術後管理用',         'dog',     'large',  4),
    (gen_random_uuid(), 'CG-005', '犬用中型ケージA', 3500.00, 'active', '中型犬・一般入院用',         'dog',     'medium', 5),
    (gen_random_uuid(), 'CG-006', '犬用中型ケージB', 3500.00, 'active', '中型犬・一般入院用',         'dog',     'medium', 6),
    (gen_random_uuid(), 'CG-007', '犬用小型ケージA', 3000.00, 'active', '小型犬・ホテル利用可',       'dog',     'small',  7),
    (gen_random_uuid(), 'CG-008', '犬用小型ケージB', 3000.00, 'active', '小型犬・ホテル利用可',       'dog',     'small',  8),
    (gen_random_uuid(), 'CG-009', '猫用ケージA',     3000.00, 'active', '猫専用・ストレス軽減設計',   'cat',     'medium', 9),
    (gen_random_uuid(), 'CG-010', '猫用ケージB',     3000.00, 'active', '猫専用・ストレス軽減設計',   'cat',     'medium', 10),
    (gen_random_uuid(), 'CG-011', '汎用ケージA',     2500.00, 'active', '小動物・鳥類等対応',         'general', 'small',  11),
    (gen_random_uuid(), 'CG-012', '汎用ケージB',     2500.00, 'active', '小動物・鳥類等対応',         'general', 'small',  12)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 9. hospitalization_plans（入院プラン）
-- -----------------------------------------------------------------------------
INSERT INTO hospitalization_plans (id, code, name, price, status, description, body_size, billing_unit, sort_order) VALUES
    (gen_random_uuid(), 'HP-001', '入院プラン（小型・1日）',   3000.00, 'active', '体重10kg以下の入院管理料（1日）',    'small',  'per_day',   1),
    (gen_random_uuid(), 'HP-002', '入院プラン（中型・1日）',   3500.00, 'active', '体重10-25kgの入院管理料（1日）',    'medium', 'per_day',   2),
    (gen_random_uuid(), 'HP-003', '入院プラン（大型・1日）',   4500.00, 'active', '体重25kg以上の入院管理料（1日）',   'large',  'per_day',   3),
    (gen_random_uuid(), 'HP-004', 'ICU管理料（1日）',          8000.00, 'active', '集中治療室管理料（1日）',            'small',  'per_day',   4),
    (gen_random_uuid(), 'HP-005', 'ホテルプラン（小型・1泊）', 2500.00, 'active', '体重10kg以下のペットホテル（1泊）', 'small',  'per_night', 5),
    (gen_random_uuid(), 'HP-006', 'ホテルプラン（中型・1泊）', 3000.00, 'active', '体重10-25kgのペットホテル（1泊）',  'medium', 'per_night', 6),
    (gen_random_uuid(), 'HP-007', 'ホテルプラン（大型・1泊）', 3500.00, 'active', '体重25kg以上のペットホテル（1泊）', 'large',  'per_night', 7)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 10. trimming_courses（トリミングコース）
-- -----------------------------------------------------------------------------
INSERT INTO trimming_courses (id, code, name, price, status, description, target_size, duration, sort_order) VALUES
    (gen_random_uuid(), 'TC-001', 'シャンプーコース（小型）',   4000.00, 'active', 'シャンプー・ブロー・ブラッシング',            'small',  '60分',  1),
    (gen_random_uuid(), 'TC-002', 'シャンプーコース（中型）',   5500.00, 'active', 'シャンプー・ブロー・ブラッシング',            'medium', '90分',  2),
    (gen_random_uuid(), 'TC-003', 'シャンプーコース（大型）',   7000.00, 'active', 'シャンプー・ブロー・ブラッシング',            'large',  '120分', 3),
    (gen_random_uuid(), 'TC-004', 'フルトリミング（小型）',     7000.00, 'active', 'カット・シャンプー・ブロー・爪切り・耳掃除', 'small',  '120分', 4),
    (gen_random_uuid(), 'TC-005', 'フルトリミング（中型）',     9000.00, 'active', 'カット・シャンプー・ブロー・爪切り・耳掃除', 'medium', '150分', 5),
    (gen_random_uuid(), 'TC-006', 'フルトリミング（大型）',    12000.00, 'active', 'カット・シャンプー・ブロー・爪切り・耳掃除', 'large',  '180分', 6),
    (gen_random_uuid(), 'TC-007', 'シャンプーコース（猫）',     5000.00, 'active', '猫専用シャンプー・ブロー・ブラッシング',      'cat',    '90分',  7),
    (gen_random_uuid(), 'TC-008', 'フルトリミング（猫）',       8000.00, 'active', '猫専用カット・シャンプー・ブロー・爪切り',   'cat',    '120分', 8)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 11. trimming_options（トリミングオプション）
-- -----------------------------------------------------------------------------
INSERT INTO trimming_options (id, code, name, price, status, description, combinable, sort_order) VALUES
    (gen_random_uuid(), 'TO-001', '爪切り',                  500.00, 'active', '爪のカット・やすりがけ',            'yes', 1),
    (gen_random_uuid(), 'TO-002', '耳掃除',                  500.00, 'active', '外耳道の洗浄・清掃',                'yes', 2),
    (gen_random_uuid(), 'TO-003', '肛門嚢絞り',              500.00, 'active', '肛門嚢の分泌液除去',                'yes', 3),
    (gen_random_uuid(), 'TO-004', '歯磨き',                  800.00, 'active', '歯ブラシによるデンタルケア',        'yes', 4),
    (gen_random_uuid(), 'TO-005', '足裏バリカン',            500.00, 'active', '足裏の毛のバリカン処理',            'yes', 5),
    (gen_random_uuid(), 'TO-006', 'リボン・バンダナ',        300.00, 'active', '仕上げのアクセサリー装着',          'yes', 6),
    (gen_random_uuid(), 'TO-007', 'ノミ・ダニ駆除スプレー', 1000.00, 'active', '外部寄生虫予防スプレー処理',        'yes', 7),
    (gen_random_uuid(), 'TO-008', 'フレッシュグルーミング',  500.00, 'active', '仕上げの整毛・コーム通し',          'yes', 8)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 12. examination_types（検査種別）
-- -----------------------------------------------------------------------------
INSERT INTO examination_types (id, code, name, price, status, description, sort_order) VALUES
    (gen_random_uuid(), 'ET-001', '血液検査（CBC）',            3000.00, 'active', '全血球計算（Complete Blood Count）',                   1),
    (gen_random_uuid(), 'ET-002', '血液化学検査（BIO）',        5000.00, 'active', '肝機能・腎機能・血糖値など生化学的検査',               2),
    (gen_random_uuid(), 'ET-003', '血液検査（CBC+BIO）',        7000.00, 'active', '全血球計算＋生化学検査のセット',                       3),
    (gen_random_uuid(), 'ET-004', '尿検査',                     1500.00, 'active', '尿試験紙・尿沈渣検査',                                 4),
    (gen_random_uuid(), 'ET-005', '便検査',                     1500.00, 'active', '糞便検査・腸内寄生虫検査',                             5),
    (gen_random_uuid(), 'ET-006', 'レントゲン検査（胸部）',     3000.00, 'active', '胸部X線撮影（2方向）',                                6),
    (gen_random_uuid(), 'ET-007', 'レントゲン検査（腹部）',     3000.00, 'active', '腹部X線撮影（2方向）',                                7),
    (gen_random_uuid(), 'ET-008', 'レントゲン検査（四肢）',     3000.00, 'active', '四肢X線撮影（2方向）',                                8),
    (gen_random_uuid(), 'ET-009', '超音波検査（腹部）',         5000.00, 'active', '腹部エコー検査',                                       9),
    (gen_random_uuid(), 'ET-010', '超音波検査（心臓）',         6000.00, 'active', '心臓エコー・心臓超音波検査',                          10),
    (gen_random_uuid(), 'ET-011', 'フィラリア抗原検査',         2000.00, 'active', 'フィラリア感染の抗原検査',                            11),
    (gen_random_uuid(), 'ET-012', 'ノミ・ダニ検査',             1000.00, 'active', '外部寄生虫の検査',                                    12),
    (gen_random_uuid(), 'ET-013', '皮膚掻爬検査',               2000.00, 'active', '皮膚病変の顕微鏡検査',                                13),
    (gen_random_uuid(), 'ET-014', '培養・感受性検査',           5000.00, 'active', '細菌培養と抗生物質感受性試験',                        14),
    (gen_random_uuid(), 'ET-015', '甲状腺機能検査',             4000.00, 'active', 'T4・TSH測定（甲状腺疾患スクリーニング）',             15)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 13. diagnosis_categories（診断カテゴリ）
-- -----------------------------------------------------------------------------
INSERT INTO diagnosis_categories (id, code, name, status, description, sort_order) VALUES
    (gen_random_uuid(), 'DC-001', '消化器疾患',    'active', '胃腸・肝臓・膵臓などの消化器系疾患',   1),
    (gen_random_uuid(), 'DC-002', '呼吸器疾患',    'active', '肺・気管・鼻腔などの呼吸器系疾患',     2),
    (gen_random_uuid(), 'DC-003', '循環器疾患',    'active', '心臓・血管などの循環器系疾患',          3),
    (gen_random_uuid(), 'DC-004', '泌尿器疾患',    'active', '腎臓・膀胱・尿道などの泌尿器系疾患',   4),
    (gen_random_uuid(), 'DC-005', '皮膚疾患',      'active', 'アレルギー・感染症などの皮膚疾患',     5),
    (gen_random_uuid(), 'DC-006', '整形外科疾患',  'active', '骨・関節・筋肉などの整形外科疾患',     6),
    (gen_random_uuid(), 'DC-007', '神経疾患',      'active', '脳・脊髄・末梢神経などの神経系疾患',   7),
    (gen_random_uuid(), 'DC-008', '眼科疾患',      'active', '目・眼瞼などの眼科疾患',               8),
    (gen_random_uuid(), 'DC-009', '耳科疾患',      'active', '外耳・中耳などの耳の疾患',             9),
    (gen_random_uuid(), 'DC-010', '内分泌疾患',    'active', '糖尿病・甲状腺・副腎などの内分泌疾患',10),
    (gen_random_uuid(), 'DC-011', '腫瘍性疾患',    'active', '良性・悪性腫瘍（がん）',              11),
    (gen_random_uuid(), 'DC-012', '感染症',        'active', '細菌・ウイルス・寄生虫感染症',        12),
    (gen_random_uuid(), 'DC-013', '歯科・口腔疾患','active', '歯周病・口内炎などの口腔内疾患',      13),
    (gen_random_uuid(), 'DC-014', '生殖器疾患',    'active', '子宮・卵巣・精巣などの生殖器疾患',    14),
    (gen_random_uuid(), 'DC-015', '外傷・中毒',    'active', '骨折・咬傷・中毒症状など',            15)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 14. diagnosis_names（診断名）- diagnosis_categoriesに依存
-- -----------------------------------------------------------------------------
INSERT INTO diagnosis_names (id, code, name, status, description, diagnosis_category_id, sort_order)
SELECT
    gen_random_uuid(), dn.code, dn.name, 'active', dn.description, dc.id, dn.sort_order
FROM (VALUES
    ('DN-001', '嘔吐症',                 '原因を問わない嘔吐症状',               'DC-001', 1),
    ('DN-002', '下痢症',                 '軟便・水様便・血便など',               'DC-001', 2),
    ('DN-003', '便秘',                   '排便困難・排便回数の減少',             'DC-001', 3),
    ('DN-004', '胃腸炎',                 '胃・腸の炎症',                         'DC-001', 4),
    ('DN-005', '肝臓疾患',               '肝炎・肝不全・脂肪肝など',             'DC-001', 5),
    ('DN-006', '膵炎',                   '膵臓の炎症',                           'DC-001', 6),
    ('DN-010', '気管支炎',               '気管支の炎症',                         'DC-002', 1),
    ('DN-011', '肺炎',                   '肺の感染性・非感染性炎症',             'DC-002', 2),
    ('DN-012', '鼻炎',                   '鼻腔の炎症・鼻水',                     'DC-002', 3),
    ('DN-020', '心臓病（僧帽弁閉鎖不全）','犬に多い僧帽弁の変性疾患',           'DC-003', 1),
    ('DN-021', '心筋症',                 '心筋の疾患（拡張型・肥大型）',         'DC-003', 2),
    ('DN-030', '膀胱炎',                 '細菌性・特発性膀胱炎',                 'DC-004', 1),
    ('DN-031', '腎不全',                 '急性・慢性腎不全',                     'DC-004', 2),
    ('DN-032', '尿路結石',               '腎結石・膀胱結石・尿道結石',           'DC-004', 3),
    ('DN-040', 'アトピー性皮膚炎',       'アレルゲンによるアレルギー性皮膚炎',  'DC-005', 1),
    ('DN-041', '細菌性皮膚炎',           '膿皮症・毛包炎など',                   'DC-005', 2),
    ('DN-042', 'ノミアレルギー性皮膚炎', 'ノミ刺咬アレルギー',                  'DC-005', 3),
    ('DN-050', '椎間板ヘルニア',         '頸椎・腰椎の椎間板突出',               'DC-006', 1),
    ('DN-051', '骨折',                   '各部位の骨折',                         'DC-006', 2),
    ('DN-052', '股関節形成不全',         '犬の股関節発育不全',                   'DC-006', 3),
    ('DN-060', '白内障',                 '水晶体の混濁',                         'DC-008', 1),
    ('DN-061', '結膜炎',                 '結膜の炎症・充血・分泌物',             'DC-008', 2),
    ('DN-062', '緑内障',                 '眼圧上昇による視神経障害',             'DC-008', 3),
    ('DN-070', '外耳炎',                 '外耳道の炎症',                         'DC-009', 1),
    ('DN-071', '中耳炎',                 '中耳腔の感染・炎症',                   'DC-009', 2),
    ('DN-080', '糖尿病',                 'インスリン不足による高血糖',           'DC-010', 1),
    ('DN-081', '甲状腺機能低下症',       '甲状腺ホルモン分泌低下（犬）',         'DC-010', 2),
    ('DN-082', '甲状腺機能亢進症',       '甲状腺ホルモン過剰分泌（猫）',         'DC-010', 3),
    ('DN-083', '副腎皮質機能亢進症',     'クッシング症候群',                     'DC-010', 4),
    ('DN-090', '肥満細胞腫',             '皮膚または内臓の肥満細胞腫瘍',         'DC-011', 1),
    ('DN-091', 'リンパ腫',               '悪性リンパ腫',                         'DC-011', 2),
    ('DN-092', '乳腺腫瘍',               '乳腺の良性・悪性腫瘍',                 'DC-011', 3),
    ('DN-100', 'パルボウイルス感染症',   '犬パルボウイルスによる感染症',         'DC-012', 1),
    ('DN-101', '猫汎白血球減少症',       '猫パルボウイルスによる感染症',         'DC-012', 2),
    ('DN-102', 'フィラリア症',           '犬糸状虫による心肺疾患',               'DC-012', 3),
    ('DN-110', '歯周病',                 '歯肉炎・歯周炎・歯槽骨融解',           'DC-013', 1),
    ('DN-111', '口内炎',                 '口腔粘膜の炎症',                       'DC-013', 2),
    ('DN-120', '子宮蓄膿症',             '細菌による子宮内膿汁蓄積',             'DC-014', 1),
    ('DN-121', '前立腺肥大',             '雄犬の前立腺肥大症',                   'DC-014', 2),
    ('DN-130', '咬傷',                   '他動物による咬傷・咬傷感染',           'DC-015', 1),
    ('DN-131', '中毒症状',               '食物・薬物・植物中毒',                 'DC-015', 2),
    ('DN-132', '熱中症',                 '高温環境による体温上昇・脱水',         'DC-015', 3)
) AS dn(code, name, description, dc_code, sort_order)
JOIN diagnosis_categories dc ON dc.code = dn.dc_code
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 15. checkup_types（健診種別）
-- -----------------------------------------------------------------------------
INSERT INTO checkup_types (id, code, name, price, status, description, interval, target_age, sort_order) VALUES
    (gen_random_uuid(), 'CK-001', '一般健康診断（基本）',        5000.00, 'active', '身体検査・体重測定・問診',                              '6ヶ月', '全年齢',  1),
    (gen_random_uuid(), 'CK-002', '一般健康診断（血液検査付）', 10000.00, 'active', '身体検査＋血液検査（CBC+BIO）',                         '6ヶ月', '全年齢',  2),
    (gen_random_uuid(), 'CK-003', 'シニア健康診断（7歳以上）',  15000.00, 'active', '身体検査＋血液検査＋レントゲン＋超音波',                 '6ヶ月', '7歳以上', 3),
    (gen_random_uuid(), 'CK-004', 'フィラリア検査',              2500.00, 'active', 'フィラリア抗原検査（予防シーズン前）',                  '1年',   '全年齢',  4),
    (gen_random_uuid(), 'CK-005', '術前健康診断',                8000.00, 'active', '手術前の全身状態確認（血液・レントゲン・心電図）',       '随時',  '手術前',  5),
    (gen_random_uuid(), 'CK-006', '妊婦健診',                    5000.00, 'active', '妊娠中の定期健康確認',                                  '随時',  '妊娠中',  6)
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 16. clinic_info（クリニック基本情報）- 1件のみ
-- -----------------------------------------------------------------------------
INSERT INTO clinic_info (id, name, branch_name, postal_code, address, phone_number, fax_number, registration_number, director_name, email, website)
SELECT
    gen_random_uuid(),
    '田中動物病院',
    '',
    '150-0001',
    '東京都渋谷区神宮前1-1-1',
    '03-1234-5678',
    '03-1234-5679',
    'VH-001',
    '田中 太郎',
    'info@animal-clinic-tanaka.jp',
    'https://animal-clinic-tanaka.jp'
WHERE NOT EXISTS (SELECT 1 FROM clinic_info);

-- -----------------------------------------------------------------------------
-- 17. clinics（クリニック一覧）
-- -----------------------------------------------------------------------------
INSERT INTO clinics (id, name, branch_name, postal_code, address, phone_number, fax_number, registration_number, director_name, email, website, is_active)
SELECT
    gen_random_uuid(),
    '田中動物病院',
    '本院',
    '150-0001',
    '東京都渋谷区神宮前1-1-1',
    '03-1234-5678',
    '03-1234-5679',
    'VH-001',
    '田中 太郎',
    'info@animal-clinic-tanaka.jp',
    'https://animal-clinic-tanaka.jp',
    true
WHERE NOT EXISTS (SELECT 1 FROM clinics WHERE registration_number = 'VH-001');
