-- =============================================================================
-- 003_seed_demo.sql
-- デモデータ投入（飼主・ペット一覧ページ対応）
-- 内容: 飼主・ペット・取引記録（カルテ・予約・会計等）
-- 依存: 001_init.sql, 002_seed_master.sql
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 1. owners（飼主: 22件）
-- -----------------------------------------------------------------------------
INSERT INTO owners (id, clinic_id, owner_name, owner_name_kana, birth_date, company, postal_code, address1, address2, phone, company_phone, email, remarks, is_dangerous, discount_rate, membership_type) VALUES
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
    owner_name      = EXCLUDED.owner_name,
    owner_name_kana = EXCLUDED.owner_name_kana,
    updated_at      = now();

SELECT setval(pg_get_serial_sequence('owners', 'id'), (SELECT MAX(id) FROM owners));

-- -----------------------------------------------------------------------------
-- 2. pets（ペット: 28件）
-- -----------------------------------------------------------------------------
INSERT INTO pets (id, clinic_id, owner_id, pet_number, name, pet_name_kana, animal_species_id, gender, status, birth_date, breed, color, weight, weight_unit, insurance_id, last_visit) VALUES
    (1,  3, 1,  '1-1', 'Iris(イリス)', 'イリス', 1, 'male',   'alive', '2015-04-14', 'ゴールデンレトリーバー',     '茶色',           26.5,  'Kg', 1, '2015-08-28'),
    (2,  3, 1,  '1-2', 'Max(マックス)', 'マックス', 1, 'male', 'alive', '2018-06-20', 'ラブラドール',               'ゴールデン',     15.2,  'Kg', NULL, '2024-11-15'),
    (3,  3, 2,  '2-1', 'ミケ',         'ミケ',     2, 'female','alive', '2020-03-10', '三毛猫',                     '三毛',            4200, 'g',  2, '2024-11-18'),
    (4,  3, 3,  '3-1', 'タロウ',       'タロウ',   1, 'male',  'alive', '2019-05-15', '柴犬',                       'レッド',          8.3,  'Kg', NULL, NULL),
    (5,  3, 3,  '3-2', 'ジロウ',       'ジロウ',   1, 'male',  'alive', '2021-08-10', '柴犬',                       'ブラック',        7.1,  'Kg', NULL, NULL),
    (6,  3, 4,  '4-1', 'チョコ',       'チョコ',   1, 'female','alive', '2017-11-20', 'トイプードル',               'チョコ',          3800, 'g',  1, NULL),
    (7,  3, 5,  '5-1', 'レオ',         'レオ',     2, 'male',  'alive', '2016-07-04', 'スコティッシュフォールド',   'グレー',          5.5,  'Kg', NULL, NULL),
    (8,  3, 6,  '6-1', 'ハチ',         'ハチ',     1, 'male',  'alive', '2018-03-25', '秋田犬',                     'ホワイト',       22.0,  'Kg', NULL, NULL),
    (9,  3, 7,  '7-1', 'モモ',         'モモ',     2, 'female','alive', '2022-01-15', 'マンチカン',                 'キャリコ',        3.2,  'Kg', 2, NULL),
    (10, 3, 8,  '8-1', 'ロッキー',     'ロッキー', 1, 'male',  'alive', '2014-09-08', 'ボーダーコリー',             'ブラックホワイト',18.5,  'Kg', NULL, NULL),
    (11, 3, 9,  '9-1', 'ルナ',         'ルナ',     2, 'female','alive', '2021-02-28', 'ペルシャ',                   'シルバー',        4800, 'g',  1, NULL),
    (12, 3, 10, '10-1', 'ケン',        'ケン',     1, 'male',  'alive', '2013-06-18', 'ジャーマンシェパード',       'ブラックタン',   32.0,  'Kg', NULL, NULL),
    (13, 3, 11, '11-1', 'ソラ',        'ソラ',     2, 'male',  'alive', '2023-04-01', 'アメリカンショートヘア',     'タビー',          3.0,  'Kg', NULL, NULL),
    (14, 3, 12, '12-1', 'ゴン',        'ゴン',     1, 'male',  'alive', '2016-12-05', '紀州犬',                     'ホワイト',       19.5,  'Kg', NULL, NULL),
    (15, 3, 13, '13-1', 'シロ',        'シロ',     1, 'male',  'alive', '2020-08-10', 'ミックス犬',                 'ホワイト',        6.2,  'Kg', NULL, NULL),
    (16, 3, 14, '14-1', 'トラ',        'トラ',     2, 'male',  'alive', '2019-10-22', 'トラ猫',                     'トラ',            5.1,  'Kg', NULL, NULL),
    (17, 3, 15, '15-1', 'ベロ',        'ベロ',     1, 'male',  'alive', '2018-05-03', 'ビーグル',                   'トライカラー',   13.2,  'Kg', NULL, NULL),
    (18, 3, 16, '16-1', 'チビ',        'チビ',     2, 'female','alive', '2022-06-20', 'ミックス猫',                 'サビ',            3500, 'g',  NULL, NULL),
    (19, 3, 17, '17-1', 'ポチ',        'ポチ',     1, 'male',  'alive', '2017-02-14', 'ダックスフンド',             'チョコ',          7.8,  'Kg', NULL, NULL),
    (20, 3, 18, '18-1', 'モカ',        'モカ',     2, 'female','alive', '2022-05-10', 'ミックス猫',                 'ホワイト',        4.1,  'Kg', NULL, NULL),
    (21, 3, 18, '18-2', 'クルミ',      'クルミ',   1, 'male',  'alive', '2020-08-20', 'ミックス犬',                 'ベージュ',        8.3,  'Kg', NULL, NULL),
    (22, 3, 19, '19-1', 'ハル',        'ハル',     1, 'male',  'alive', '2019-03-15', 'ミックス犬',                 'ブラック',       12.5,  'Kg', NULL, NULL),
    (23, 3, 19, '19-2', 'ユキ',        'ユキ',     2, 'female','alive', '2021-12-01', 'ミックス猫',                 'ホワイト',        3800, 'g',  NULL, NULL),
    (24, 3, 20, '20-1', 'ピーチ',      'ピーチ',   2, 'female','alive', '2023-01-07', 'ミックス猫',                 'オレンジ',        3.2,  'Kg', NULL, NULL),
    (25, 3, 21, '21-1', 'コタ',        'コタ',     1, 'male',  'alive', '2018-09-23', 'ミックス犬',                 'ブラウン',       22.0,  'Kg', NULL, NULL),
    (26, 3, 21, '21-2', 'アン',        'アン',     2, 'female','alive', '2020-04-11', 'ミックス猫',                 'キャリコ',        4.5,  'Kg', NULL, NULL),
    (27, 3, 22, '22-1', 'ゴマ',        'ゴマ',     2, 'male',  'alive', '2022-11-30', 'ミックス猫',                 'グレー',          5.0,  'Kg', NULL, NULL),
    (28, 3, 22, '22-2', 'マル',        'マル',     1, 'female','alive', '2021-06-18', 'ミックス犬',                 'ゴールデン',      9.7,  'Kg', NULL, NULL)
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();

SELECT setval(pg_get_serial_sequence('pets', 'id'), (SELECT MAX(id) FROM pets));

-- -----------------------------------------------------------------------------
-- 3. reservation_appointments（予約: 10件）
-- service_type: 1=一般診察, 2=予防接種, 3=健康診断
-- -----------------------------------------------------------------------------
INSERT INTO reservation_appointments (id, clinic_id, start_time, end_time, owner_id, pet_id, visit_type, service_type_id, doctor_id, is_designated, status, notes) VALUES
    (1,  3, '2026-03-12 09:00:00+09', '2026-03-12 09:30:00+09', 1,  1,  'revisit', 1, 1, true,  'completed',       '皮膚の経過観察'),
    (2,  3, '2026-03-12 09:30:00+09', '2026-03-12 10:00:00+09', 2,  3,  'revisit', 3, 2, false, 'accounting',      '猫の定期検診'),
    (3,  3, '2026-03-12 10:00:00+09', '2026-03-12 10:30:00+09', 3,  4,  'revisit', 1, 1, true,  'in_consultation',  '足を引きずっている'),
    (4,  3, '2026-03-12 10:30:00+09', '2026-03-12 11:00:00+09', 4,  6,  'first',   2, 2, false, 'checked_in',      'ワクチン接種希望'),
    (5,  3, '2026-03-12 14:00:00+09', '2026-03-12 14:30:00+09', 6,  8,  'revisit', 1, 1, false, 'confirmed',       '食欲低下が続いている'),
    (6,  3, '2026-03-13 09:00:00+09', '2026-03-13 09:30:00+09', 7,  9,  'revisit', 1, 2, true,  'confirmed',       '耳の治療経過確認'),
    (7,  3, '2026-03-13 10:00:00+09', '2026-03-13 10:30:00+09', 8,  10, 'first',   1, 1, false, 'confirmed',       '嘔吐が続いている'),
    (8,  3, '2026-03-14 09:30:00+09', '2026-03-14 10:00:00+09', 9,  11, 'revisit', 1, 2, false, 'confirmed',       'ルナの経過観察'),
    (9,  3, '2026-03-15 11:00:00+09', '2026-03-15 11:30:00+09', 10, 12, 'first',   2, 1, false, 'confirmed',       '初回ワクチン接種'),
    (10, 3, '2026-03-16 14:00:00+09', '2026-03-16 14:30:00+09', 11, 13, 'revisit', 1, 2, true,  'confirmed',       '腎臓値の経過観察')
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();

SELECT setval(pg_get_serial_sequence('reservation_appointments', 'id'), (SELECT MAX(id) FROM reservation_appointments));

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
    (1,  3,  NULL, '狂犬病ワクチン接種',         '狂犬病ワクチン接種。体調良好。', 1),
    (2,  2,  NULL, '定期健診',                   '特に異常なし。次回は6ヶ月後を推奨。', 2),
    (3,  3,  6,    '右足の跛行',                 '右後肢の跛行。レントゲン撮影で膝蓋骨脱臼を確認。手術を検討。', 1),
    (4,  4,  NULL, 'フィラリア予防',             'フィラリア予防薬の処方。', 2),
    (5,  5,  NULL, '5種混合ワクチン接種',        '体調良好。ワクチン接種実施。', 1),
    (6,  6,  NULL, '5種混合ワクチン接種',        '体調良好。年次ワクチン接種。', 2),
    (7,  7,  NULL, 'ノミダニ予防薬',             'ノミダニ予防薬の処方。', 1),
    (8,  8,  3,    '皮膚の痒み',                 '全身の痒みと発赤。アトピー性皮膚炎の疑い。抗ヒスタミン処方。', 2),
    (9,  9,  3,    'トリミング後の皮膚チェック', 'トリミング後に赤みあり。経過観察。', 1),
    (10, 10, 1,    '食欲不振',                   '2日前から食欲が落ちている。腹部エコーで検査予定。', 2),
    (11, 11, NULL, '耳を痒がる',                 '外耳炎の疑い。耳道洗浄と点耳薬処方。', 1),
    (12, 12, NULL, '定期健診・予防接種',         '年次健康診断。血液検査・予防接種実施。', 2),
    (13, 13, 2,    '嘔吐・下痢',                 '昨日から嘔吐3回。下痢あり。食欲なし。', 1),
    (14, 14, NULL, '生化学検査',                 'シニア健診。血液化学検査実施。', 2),
    (15, 15, NULL, 'ジステンパーワクチン接種',   '初回ワクチン接種。体調良好。', 1),
    (16, 16, NULL, '血液検査',                   '年次血液検査。特に異常なし。', 2),
    (17, 17, NULL, '歯石除去',                   '重度の歯石。全麻下での歯石除去処置予定。', 1),
    (18, 18, NULL, '定期検診',                   '特に異常なし。体重管理継続を指導。', 2),
    (19, 19, 6,    '再診（右足跛行）',           '前回の経過観察。改善傾向あり。', 1),
    (20, 20, NULL, '5種混合ワクチン接種',        '体調良好。ワクチン接種実施。', 2)
ON CONFLICT (id) DO UPDATE SET
    medical_record_id = EXCLUDED.medical_record_id,
    updated_at        = now();

SELECT setval(pg_get_serial_sequence('inquiries', 'id'), (SELECT MAX(id) FROM inquiries));

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
INSERT INTO treatments (id, medical_record_id, item_type, consultation_id, procedure_id, medicine_id, inventory_id, selected, status, content, unit_price, quantity, sort_order) VALUES
    (1, 3, 'consultation', 2,    NULL, NULL, NULL, true, 'completed', '再診料',                    800,  1, 1),
    (2, 1, 'medicine',     NULL, NULL, 1,    1,    true, 'completed', 'アモキシシリン 50mg x 7日分', 500,  7, 2),
    (3, 2, 'consultation', 2,    NULL, NULL, NULL, true, 'completed', '再診料',                    800,  1, 1),
    (4, 3, 'consultation', 2,    NULL, NULL, NULL, true, 'completed', '再診料',                    800,  1, 1),
    (5, 3, 'procedure',    NULL, 4,    NULL, NULL, true, 'completed', '耳道洗浄（左耳）',          2500, 1, 2),
    (6, 4, 'consultation', 1,    NULL, NULL, NULL, true, 'completed', '初診料',                    2000, 1, 1),
    (7, 4, 'medicine',     NULL, NULL, 1,    1,    true, 'completed', 'アモキシシリン 50mg x 5日分', 500,  5, 2),
    (8, 5, 'consultation', 2,    NULL, NULL, NULL, true, 'completed', '再診料',                    800,  1, 1)
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();

SELECT setval(pg_get_serial_sequence('treatments', 'id'), (SELECT MAX(id) FROM treatments));

-- -----------------------------------------------------------------------------
-- 8. billings / billing_items / payments
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
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();

SELECT setval(pg_get_serial_sequence('billing_items', 'id'), (SELECT MAX(id) FROM billing_items));

INSERT INTO payments (id, billing_id, subtotal, tax_total, total_amount, insurance_name, insurance_ratio, insurance_amount, discount_amount, billing_amount, received_amount, change_amount, method) VALUES
    (1, 1, 4300, 430, 4730, 'アニコム損保', 0.70, 3311, 0, 1419, 1500, 81, 'cash'),
    (2, 2, 3300, 330, 3630, 'アニコム損保', 0.70, 2541, 0, 1089, 1100, 11, 'credit_card')
ON CONFLICT (id) DO UPDATE SET
    updated_at = now();

SELECT setval(pg_get_serial_sequence('payments', 'id'), (SELECT MAX(id) FROM payments));

-- -----------------------------------------------------------------------------
-- 9. trimming_records（トリミング: 8件）
-- -----------------------------------------------------------------------------
INSERT INTO trimming_records (id, clinic_id, date, pet_id, bw, bw_unit, style_request, staff_id, status, course_id) VALUES
    (1, 3, '2025-10-10', 1,  26.5,  'Kg', 'サマーカット希望',        6,  'completed',   3),
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
-- 10. audit_logs（監査ログ: 8件）
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
