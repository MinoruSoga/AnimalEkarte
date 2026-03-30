-- =============================================================================
-- 003_seed_demo.sql
-- デモデータ投入（飼主・ペット一覧ページ対応）
-- 対象クリニック: 渋谷院 (id=1)
-- ui-sample/src/features/pets/api/mockData.ts と完全一致
-- 冪等性: ON CONFLICT (id) DO UPDATE SET で再実行可能
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 1. owners（飼主: 17件）
-- ownerId 30042〜30058 → 整数 1〜17
-- -----------------------------------------------------------------------------
INSERT INTO owners (id, clinic_id, owner_name, owner_name_kana, birth_date, company, postal_code, address1, address2, phone, company_phone, email, remarks, is_dangerous, discount_rate, membership_type) VALUES
    -- 30042: 林 文明（詳細データあり）
    (1, 3, '林 文明', 'ハヤシ フミアキ', '1980-05-15', 'サンプル株式会社',
     '150-0001', '東京都渋谷区神宮前1-2-3', '', '090-1111-2222', '03-1234-5678',
     'hayashi@example.com', '定期検診を希望', false, 10, 'member'),
    -- 30043: 田中 花子（詳細データあり）
    (2, 3, '田中 花子', 'タナカ ハナコ', '1985-03-20', '',
     '160-0022', '東京都新宿区新宿1-1-1', '', '080-3333-4444', '',
     'tanaka@example.com', '', false, 0, 'non_member'),
    -- 30044: 鈴木 一郎
    (3, 3, '鈴木 一郎', 'スズキ イチロウ', '1978-11-03', '',
     '170-0001', '東京都豊島区西巣鴨1-3-5', '', '070-5555-6666', '',
     'suzuki@example.com', '', false, 0, 'member'),
    -- 30045: 田中 美咲
    (4, 3, '田中 美咲', 'タナカ ミサキ', '1990-07-22', '',
     '153-0044', '東京都目黒区大橋2-4-6', '', '090-9999-8888', '',
     'misaki.tanaka@example.com', '', false, 0, 'non_member'),
    -- 30046: 佐藤 花子
    (5, 3, '佐藤 花子', 'サトウ ハナコ', '1975-02-14', '',
     '140-0001', '東京都品川区北品川3-5-7', '', '080-2222-3333', '',
     'hanako.sato@example.com', '', false, 5, 'member'),
    -- 30047: 伊藤 次郎
    (6, 3, '伊藤 次郎', 'イトウ ジロウ', '1983-09-30', '',
     '166-0013', '東京都杉並区堀ノ内1-7-9', '', '090-1234-5678', '',
     'jiro.ito@example.com', '', false, 0, 'non_member'),
    -- 30048: 小林 さくら
    (7, 3, '小林 さくら', 'コバヤシ サクラ', '1992-04-05', '',
     '176-0012', '東京都練馬区豊玉北4-2-8', '', '080-9876-5432', '',
     'sakura.kobayashi@example.com', '', false, 0, 'member'),
    -- 30049: 中村 勇気
    (8, 3, '中村 勇気', 'ナカムラ ユウキ', '1987-12-18', '',
     '174-0041', '東京都板橋区舟渡2-6-10', '', '090-1122-3344', '',
     'yuuki.nakamura@example.com', '', false, 0, 'non_member'),
    -- 30050: 加藤 恵
    (9, 3, '加藤 恵', 'カトウ メグミ', '1995-06-25', '',
     '134-0083', '東京都江戸川区中葛西5-3-2', '', '080-5566-7788', '',
     'megumi.kato@example.com', '', false, 10, 'member'),
    -- 30051: 山田 太郎
    (10, 3, '山田 太郎', 'ヤマダ タロウ', '1970-01-10', '',
     '144-0051', '東京都大田区西蒲田6-8-4', '', '090-2233-4455', '',
     'taro.yamada@example.com', '', false, 0, 'non_member'),
    -- 30052: 高橋 由美
    (11, 3, '高橋 由美', 'タカハシ ユミ', '1988-08-15', '',
     '110-0005', '東京都台東区上野5-1-3', '', '080-6677-8899', '',
     'yumi.takahashi@example.com', '', false, 0, 'member'),
    -- 30053: 松本 隆
    (12, 3, '松本 隆', 'マツモト タカシ', '1965-03-28', '',
     '125-0061', '東京都葛飾区亀有3-9-7', '', '090-3344-5566', '',
     'takashi.matsumoto@example.com', '', false, 0, 'non_member'),
    -- 30054: 吉田 誠
    (13, 3, '吉田 誠', 'ヨシダ マコト', '1982-11-05', '',
     '123-0845', '東京都足立区西新井7-4-6', '', '080-7788-9900', '',
     'makoto.yoshida@example.com', '', false, 0, 'non_member'),
    -- 30055: 井上 京子
    (14, 3, '井上 京子', 'イノウエ キョウコ', '1973-05-19', '',
     '189-0023', '東京都東村山市美住町1-5-2', '', '090-4455-6677', '',
     'kyoko.inoue@example.com', '', false, 5, 'member'),
    -- 30056: 木村 拓也
    (15, 3, '木村 拓也', 'キムラ タクヤ', '1991-07-14', '',
     '179-0081', '東京都練馬区北町3-6-9', '', '080-8899-0011', '',
     'takuya.kimura@example.com', '', false, 0, 'non_member'),
    -- 30057: 佐々木 亮
    (16, 3, '佐々木 亮', 'ササキ リョウ', '1986-02-23', '',
     '207-0013', '東京都東大和市清水2-4-8', '', '090-5566-7788', '',
     'ryo.sasaki@example.com', '', false, 0, 'non_member'),
    -- 30058: 山本 健太
    (17, 3, '山本 健太', 'ヤマモト ケンタ', '1998-09-12', '',
     '206-0802', '東京都稲城市東長沼2-8-3', '', '090-1234-9876', '',
     'kenta.yamamoto@example.com', '', false, 0, 'non_member'),
    -- ページネーションテスト用追加オーナー (id 18-22)
    (18, 3, '青木 麻衣', 'アオキ マイ', '1993-03-10', '',
     '150-0002', '東京都渋谷区渋谷2-1-1', '', '090-1111-1111', '',
     'mai.aoki@example.com', '', false, 0, 'non_member'),
    (19, 3, '橋本 俊介', 'ハシモト シュンスケ', '1980-07-25', '',
     '130-0001', '東京都墨田区吾妻橋1-3-5', '', '080-2222-2222', '',
     'shunsuke.h@example.com', '', false, 0, 'member'),
    (20, 3, '福田 裕子', 'フクダ ユウコ', '1977-11-14', '',
     '145-0062', '東京都大田区北千束2-5-8', '', '090-3333-3333', '',
     'yuko.fukuda@example.com', '', false, 5, 'member'),
    (21, 3, '石川 大輔', 'イシカワ ダイスケ', '1989-04-02', '',
     '167-0041', '東京都杉並区善福寺3-2-6', '', '080-4444-4444', '',
     'daisuke.ishikawa@example.com', '', false, 0, 'non_member'),
    (22, 3, '村田 奈々', 'ムラタ ナナ', '1996-09-19', '',
     '182-0021', '東京都調布市調布ヶ丘1-4-7', '', '090-5555-5555', '',
     'nana.murata@example.com', '', false, 0, 'non_member')
ON CONFLICT (id) DO UPDATE SET
    phone           = EXCLUDED.phone,
    owner_name      = EXCLUDED.owner_name,
    owner_name_kana = EXCLUDED.owner_name_kana,
    email           = EXCLUDED.email,
    membership_type = EXCLUDED.membership_type,
    discount_rate   = EXCLUDED.discount_rate,
    remarks         = EXCLUDED.remarks,
    updated_at      = now();

SELECT setval(pg_get_serial_sequence('owners', 'id'), (SELECT MAX(id) FROM owners));

-- -----------------------------------------------------------------------------
-- 2. pets（ペット: 19件）
-- animal_species: 犬=1, 猫=2
-- insurance: 1=ペット保険A, 2=ペット保険B
-- -----------------------------------------------------------------------------
INSERT INTO pets (id, clinic_id, owner_id, pet_number, name, pet_name_kana, animal_species_id, gender, status, birth_date, breed, color, weight, neutered_date, danger_level, insurance_id, remarks, environment, last_visit) VALUES
    -- id:1  林 文明 → Iris(イリス) 犬 (owner_id=1, 1頭目)
    (1,  3, 1,  '1-1', 'Iris(イリス)', 'イリス', 1, 'male',   'alive', '2015-04-14', 'ゴールデンレトリーバー',     '茶色',           26.5, '2016-03-10', 'low',    1,    '定期検診中',     '室内(散歩する)', '2015-08-28'),
    -- id:2  林 文明 → Max(マックス) 犬 (owner_id=1, 2頭目)
    (2,  3, 1,  '1-2', 'Max(マックス)', 'マックス', 1, 'male', 'alive', '2018-06-20', 'ラブラドール',               'ゴールデン',     15.2, NULL,         'low',    NULL, '',               '室内(散歩する)', '2024-11-15'),
    -- id:3  田中 花子 → ミケ 猫 (owner_id=2, 1頭目)
    (3,  3, 2,  '2-1', 'ミケ',         'ミケ',     2, 'female','alive', '2020-03-10', '三毛猫',                     '三毛',            4.2, '2021-05-20', 'low',    2,    '',               '室内',           '2024-11-18'),
    -- id:4  鈴木 一郎 → タロウ 犬 (owner_id=3, 1頭目)
    (4,  3, 3,  '3-1', 'タロウ',       'タロウ',   1, 'male',  'alive', '2019-05-15', '柴犬',                       'レッド',          8.3, '2020-06-01', 'low',    NULL, '',               '',               NULL),
    -- id:5  鈴木 一郎 → ジロウ 犬 (owner_id=3, 2頭目)
    (5,  3, 3,  '3-2', 'ジロウ',       'ジロウ',   1, 'male',  'alive', '2021-08-10', '柴犬',                       'ブラック',        7.1, NULL,         'low',    NULL, '',               '',               NULL),
    -- id:6  田中 美咲 → チョコ 犬 (owner_id=4, 1頭目)
    (6,  3, 4,  '4-1', 'チョコ',       'チョコ',   1, 'female','alive', '2017-11-20', 'トイプードル',               'チョコ',          3.8, '2018-12-15', 'low',    1,    '',               '',               NULL),
    -- id:7  佐藤 花子 → レオ 猫 (owner_id=5, 1頭目)
    (7,  3, 5,  '5-1', 'レオ',         'レオ',     2, 'male',  'alive', '2016-07-04', 'スコティッシュフォールド',   'グレー',          5.5, '2017-08-20', 'low',    NULL, '',               '',               NULL),
    -- id:8  伊藤 次郎 → ハチ 犬 (owner_id=6, 1頭目)
    (8,  3, 6,  '6-1', 'ハチ',         'ハチ',     1, 'male',  'alive', '2018-03-25', '秋田犬',                     'ホワイト',       22.0, '2019-04-10', 'medium', NULL, '噛み付く場合あり', '',             NULL),
    -- id:9  小林 さくら → モモ 猫 (owner_id=7, 1頭目)
    (9,  3, 7,  '7-1', 'モモ',         'モモ',     2, 'female','alive', '2022-01-15', 'マンチカン',                 'キャリコ',        3.2, NULL,         'low',    2,    '',               '',               NULL),
    -- id:10 中村 勇気 → ロッキー 犬 (owner_id=8, 1頭目)
    (10, 3, 8,  '8-1', 'ロッキー',     'ロッキー', 1, 'male',  'alive', '2014-09-08', 'ボーダーコリー',             'ブラックホワイト',18.5, '2015-10-01', 'low',    NULL, '',               '',               NULL),
    -- id:11 加藤 恵 → ルナ 猫 (owner_id=9, 1頭目)
    (11, 3, 9,  '9-1', 'ルナ',         'ルナ',     2, 'female','alive', '2021-02-28', 'ペルシャ',                   'シルバー',        4.8, NULL,         'low',    1,    '',               '',               NULL),
    -- id:12 山田 太郎 → ケン 犬 (owner_id=10, 1頭目)
    (12, 3, 10, '10-1', 'ケン',        'ケン',     1, 'male',  'alive', '2013-06-18', 'ジャーマンシェパード',       'ブラックタン',   32.0, '2014-07-05', 'low',    NULL, '',               '',               NULL),
    -- id:13 高橋 由美 → ソラ 猫 (owner_id=11, 1頭目)
    (13, 3, 11, '11-1', 'ソラ',        'ソラ',     2, 'male',  'alive', '2023-04-01', 'アメリカンショートヘア',     'タビー',          3.0, NULL,         'low',    NULL, '',               '',               NULL),
    -- id:14 松本 隆 → ゴン 犬 (owner_id=12, 1頭目)
    (14, 3, 12, '12-1', 'ゴン',        'ゴン',     1, 'male',  'alive', '2016-12-05', '紀州犬',                     'ホワイト',       19.5, '2017-12-20', 'low',    NULL, '',               '',               NULL),
    -- id:15 吉田 誠 → シロ 犬 (owner_id=13, 1頭目)
    (15, 3, 13, '13-1', 'シロ',        'シロ',     1, 'male',  'alive', '2020-08-10', 'ミックス犬',                 'ホワイト',        6.2, '2021-09-01', 'low',    NULL, '',               '',               NULL),
    -- id:16 井上 京子 → トラ 猫 (owner_id=14, 1頭目)
    (16, 3, 14, '14-1', 'トラ',        'トラ',     2, 'male',  'alive', '2019-10-22', 'トラ猫',                     'トラ',            5.1, NULL,         'low',    NULL, '',               '',               NULL),
    -- id:17 木村 拓也 → ベロ 犬 (owner_id=15, 1頭目)
    (17, 3, 15, '15-1', 'ベロ',        'ベロ',     1, 'male',  'alive', '2018-05-03', 'ビーグル',                   'トライカラー',   13.2, '2019-06-15', 'low',    NULL, '',               '',               NULL),
    -- id:18 佐々木 亮 → チビ 猫 (owner_id=16, 1頭目)
    (18, 3, 16, '16-1', 'チビ',        'チビ',     2, 'female','alive', '2022-06-20', 'ミックス猫',                 'サビ',            3.5, NULL,         'low',    NULL, '',               '',               NULL),
    -- id:19 山本 健太 → ポチ 犬 (owner_id=17, 1頭目)
    (19, 3, 17, '17-1', 'ポチ',        'ポチ',     1, 'male',  'alive', '2017-02-14', 'ダックスフンド',             'チョコ',          7.8, '2018-03-01', 'low',    NULL, '',               '',               NULL),
    -- id:20-28 ページネーションテスト用追加データ
    (20, 3, 18, '18-1', 'モカ',        'モカ',     2, 'female','alive', '2022-05-10', 'ミックス猫',    'ホワイト',        4.1, NULL, 'low', NULL, '', '', NULL),
    (21, 3, 18, '18-2', 'クルミ',      'クルミ',   1, 'male',  'alive', '2020-08-20', 'ミックス犬',    'ベージュ',        8.3, NULL, 'low', NULL, '', '', NULL),
    (22, 3, 19, '19-1', 'ハル',        'ハル',     1, 'male',  'alive', '2019-03-15', 'ミックス犬',    'ブラック',       12.5, NULL, 'low', NULL, '', '', NULL),
    (23, 3, 19, '19-2', 'ユキ',        'ユキ',     2, 'female','alive', '2021-12-01', 'ミックス猫',    'ホワイト',        3.8, NULL, 'low', NULL, '', '', NULL),
    (24, 3, 20, '20-1', 'ピーチ',      'ピーチ',   2, 'female','alive', '2023-01-07', 'ミックス猫',    'オレンジ',        3.2, NULL, 'low', NULL, '', '', NULL),
    (25, 3, 21, '21-1', 'コタ',        'コタ',     1, 'male',  'alive', '2018-09-23', 'ミックス犬',    'ブラウン',       22.0, NULL, 'low', NULL, '', '', NULL),
    (26, 3, 21, '21-2', 'アン',        'アン',     2, 'female','alive', '2020-04-11', 'ミックス猫',    'キャリコ',        4.5, NULL, 'low', NULL, '', '', NULL),
    (27, 3, 22, '22-1', 'ゴマ',        'ゴマ',     2, 'male',  'alive', '2022-11-30', 'ミックス猫',    'グレー',          5.0, NULL, 'low', NULL, '', '', NULL),
    (28, 3, 22, '22-2', 'マル',        'マル',     1, 'female','alive', '2021-06-18', 'ミックス犬',    'ゴールデン',      9.7, NULL, 'low', NULL, '', '', NULL)
ON CONFLICT (id) DO UPDATE SET
    owner_id           = EXCLUDED.owner_id,
    pet_number         = EXCLUDED.pet_number,
    name               = EXCLUDED.name,
    animal_species_id  = EXCLUDED.animal_species_id,
    gender             = EXCLUDED.gender,
    birth_date         = EXCLUDED.birth_date,
    breed              = EXCLUDED.breed,
    color              = EXCLUDED.color,
    weight             = EXCLUDED.weight,
    environment        = EXCLUDED.environment,
    last_visit         = EXCLUDED.last_visit,
    updated_at         = now();

SELECT setval(pg_get_serial_sequence('pets', 'id'), (SELECT MAX(id) FROM pets));

-- -----------------------------------------------------------------------------
-- 3. medical_records（カルテ: 20件）
-- doctor_id: 山田 太郎=1, 高橋 健一=2
-- -----------------------------------------------------------------------------
INSERT INTO medical_records (id, clinic_id, record_no, date, owner_id, pet_id, doctor_id, status) VALUES
    -- Iris(イリス) / 林 文明
    (1,  3, 'R-2025-001', '2025-10-10', 1,  1,  1, 'finalized'),
    (2,  3, 'R-2025-002', '2025-12-15', 1,  1,  1, 'finalized'),
    (3,  3, 'R-2026-001', '2026-01-20', 1,  1,  2, 'finalized'),
    -- Max(マックス) / 林 文明
    (4,  3, 'R-2025-003', '2025-11-05', 1,  2,  2, 'finalized'),
    -- ミケ / 田中 花子
    (5,  3, 'R-2025-004', '2025-09-15', 2,  3,  2, 'finalized'),
    (6,  3, 'R-2026-002', '2026-01-06', 2,  3,  1, 'finalized'),
    -- タロウ / 鈴木 一郎
    (7,  3, 'R-2025-005', '2025-08-22', 3,  4,  2, 'finalized'),
    -- チョコ / 田中 美咲
    (8,  3, 'R-2025-006', '2025-10-18', 4,  6,  1, 'finalized'),
    -- レオ / 佐藤 花子
    (9,  3, 'R-2025-007', '2025-07-30', 5,  7,  2, 'finalized'),
    -- ハチ / 伊藤 次郎
    (10, 3, 'R-2026-003', '2026-01-15', 6,  8,  1, 'draft'),
    -- モモ / 小林 さくら
    (11, 3, 'R-2025-008', '2025-12-01', 7,  9,  2, 'finalized'),
    -- ロッキー / 中村 勇気
    (12, 3, 'R-2025-009', '2025-11-20', 8,  10, 1, 'finalized'),
    -- ルナ / 加藤 恵
    (13, 3, 'R-2026-004', '2026-02-10', 9,  11, 2, 'draft'),
    -- ケン / 山田 太郎
    (14, 3, 'R-2025-010', '2025-06-15', 10, 12, 1, 'finalized'),
    -- ソラ / 高橋 由美
    (15, 3, 'R-2026-005', '2026-01-06', 11, 13, 2, 'finalized'),
    -- ゴン / 松本 隆
    (16, 3, 'R-2025-011', '2025-09-08', 12, 14, 1, 'finalized'),
    -- シロ / 吉田 誠
    (17, 3, 'R-2026-006', '2026-02-28', 13, 15, 2, 'draft'),
    -- トラ / 井上 京子
    (18, 3, 'R-2025-012', '2025-08-20', 14, 16, 1, 'finalized'),
    -- ベロ / 木村 拓也
    (19, 3, 'R-2026-007', '2026-01-03', 15, 17, 2, 'finalized'),
    -- チビ / 佐々木 亮
    (20, 3, 'R-2026-008', '2026-01-06', 16, 18, 1, 'finalized')
ON CONFLICT (id) DO UPDATE SET
    clinic_id  = EXCLUDED.clinic_id,
    record_no  = EXCLUDED.record_no,
    date       = EXCLUDED.date,
    owner_id   = EXCLUDED.owner_id,
    pet_id     = EXCLUDED.pet_id,
    doctor_id  = EXCLUDED.doctor_id,
    status     = EXCLUDED.status,
    updated_at = now();

SELECT setval(pg_get_serial_sequence('medical_records', 'id'), (SELECT MAX(id) FROM medical_records));

-- -----------------------------------------------------------------------------
-- 4. inquiries（問診: 20件）
-- chief_complaint_category_id: 食欲不振=1, 嘔吐・下痢=2, 皮膚・被毛=3, 外傷・骨折=6
-- -----------------------------------------------------------------------------
INSERT INTO inquiries (id, medical_record_id, chief_complaint_category_id, chief_complaint, notes) VALUES
    (1,  3,  NULL, '狂犬病ワクチン接種',         '狂犬病ワクチン接種。体調良好。'),
    (2,  2,  NULL, '定期健診',                   '特に異常なし。次回は6ヶ月後を推奨。'),
    (3,  3,  6,    '右足の跛行',                 '右後肢の跛行。レントゲン撮影で膝蓋骨脱臼を確認。手術を検討。'),
    (4,  4,  NULL, 'フィラリア予防',             'フィラリア予防薬の処方。'),
    (5,  5,  NULL, '5種混合ワクチン接種',        '体調良好。ワクチン接種実施。'),
    (6,  6,  NULL, '5種混合ワクチン接種',        '体調良好。年次ワクチン接種。'),
    (7,  7,  NULL, 'ノミダニ予防薬',             'ノミダニ予防薬の処方。'),
    (8,  8,  3,    '皮膚の痒み',                 '全身の痒みと発赤。アトピー性皮膚炎の疑い。抗ヒスタミン処方。'),
    (9,  9,  3,    'トリミング後の皮膚チェック', 'トリミング後に赤みあり。経過観察。'),
    (10, 10, 1,    '食欲不振',                   '2日前から食欲が落ちている。腹部エコーで検査予定。'),
    (11, 11, NULL, '耳を痒がる',                 '外耳炎の疑い。耳道洗浄と点耳薬処方。'),
    (12, 12, NULL, '定期健診・予防接種',         '年次健康診断。血液検査・予防接種実施。'),
    (13, 13, 2,    '嘔吐・下痢',                 '昨日から嘔吐3回。下痢あり。食欲なし。'),
    (14, 14, NULL, '生化学検査',                 'シニア健診。血液化学検査実施。'),
    (15, 15, NULL, 'ジステンパーワクチン接種',   '初回ワクチン接種。体調良好。'),
    (16, 16, NULL, '血液検査',                   '年次血液検査。特に異常なし。'),
    (17, 17, NULL, '歯石除去',                   '重度の歯石。全麻下での歯石除去処置予定。'),
    (18, 18, NULL, '定期検診',                   '特に異常なし。体重管理継続を指導。'),
    (19, 19, 6,    '再診（右足跛行）',           '前回の経過観察。改善傾向あり。'),
    (20, 20, NULL, '5種混合ワクチン接種',        '体調良好。ワクチン接種実施。')
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('inquiries', 'id'), (SELECT MAX(id) FROM inquiries));

-- -----------------------------------------------------------------------------
-- 5. reservation_appointments（予約: 10件 — 002の誤データを正しい owner/pet で上書き）
-- visit_type: 'revisit' | 'first'
-- status: 'completed','accounting','in_consultation','checked_in','confirmed','pending'
-- service_type: 1=一般診察, 2=予防接種, 8=定期健診
-- doctor: 1=山田太郎, 2=高橋健一
-- -----------------------------------------------------------------------------
INSERT INTO reservation_appointments (id, clinic_id, start_time, end_time, owner_id, pet_id, visit_type, service_type_id, doctor_id, is_designated, status, notes) VALUES
    (1,  3, '2026-03-12 09:00:00+09', '2026-03-12 09:30:00+09', 1,  1,  'revisit', 1, 1, true,  'completed',       '皮膚の経過観察'),
    (2,  3, '2026-03-12 09:30:00+09', '2026-03-12 10:00:00+09', 2,  3,  'revisit', 1, 2, false, 'accounting',      '猫の定期健診'),
    (3,  3, '2026-03-12 10:00:00+09', '2026-03-12 10:30:00+09', 3,  4,  'revisit', 1, 1, true,  'in_consultation',  '足を引きずっている'),
    (4,  3, '2026-03-12 10:30:00+09', '2026-03-12 11:00:00+09', 4,  6,  'first',   2, 2, false, 'checked_in',      'ワクチン接種希望'),
    (5,  3, '2026-03-12 14:00:00+09', '2026-03-12 14:30:00+09', 6,  8,  'revisit', 1, 1, false, 'confirmed',       '食欲低下が続いている'),
    (6,  3, '2026-03-13 09:00:00+09', '2026-03-13 09:30:00+09', 7,  9,  'revisit', 1, 2, true,  'confirmed',       '耳の治療経過確認'),
    (7,  3, '2026-03-13 10:00:00+09', '2026-03-13 10:30:00+09', 8,  10, 'first',   1, 1, false, 'confirmed',       '嘔吐が続いている'),
    (8,  3, '2026-03-14 09:30:00+09', '2026-03-14 10:00:00+09', 9,  11, 'revisit', 1, 2, false, 'confirmed',       'ルナの経過観察'),
    (9,  3, '2026-03-15 11:00:00+09', '2026-03-15 11:30:00+09', 10, 12, 'first',   2, 1, false, 'confirmed',       '初回ワクチン接種'),
    (10, 3, '2026-03-16 14:00:00+09', '2026-03-16 14:30:00+09', 11, 13, 'revisit', 1, 2, true,  'confirmed',       '腎臓値の経過観察')
ON CONFLICT (id) DO UPDATE SET
    owner_id        = EXCLUDED.owner_id,
    pet_id          = EXCLUDED.pet_id,
    service_type_id = EXCLUDED.service_type_id,
    visit_type      = EXCLUDED.visit_type,
    status          = EXCLUDED.status,
    notes           = EXCLUDED.notes,
    updated_at      = now();

SELECT setval(pg_get_serial_sequence('reservation_appointments', 'id'), (SELECT MAX(id) FROM reservation_appointments));

-- -----------------------------------------------------------------------------
-- 6. trimming_records（トリミング記録: 8件 NEW）
-- status: 'completed' | 'reserved' | 'in_progress'
-- course_id: 1=シャンプーコース, 2=爪切り・ブラッシング, 3=サマーカット, 4=全体カット
-- staff: 6=鈴木一郎(トリマー), 12=高橋さくら(トリマー)
-- -----------------------------------------------------------------------------
INSERT INTO trimming_records (id, clinic_id, date, pet_id, bw, style_request, staff_id, status, course_id) VALUES
    (1, 3, '2025-10-10', 1,  '26.5', 'サマーカット希望',        6,  'completed',   3),
    (2, 3, '2025-10-15', 2,  '15.2', 'ふんわりカット',          12, 'reserved',    4),
    (3, 3, '2025-10-12', 3,  '4.2',  '毛玉カット',              6,  'in_progress', 1),
    (4, 3, '2026-01-06', 6,  '3.8',  'シャンプーコース',        6,  'completed',   1),
    (5, 3, '2026-01-06', 17, '12.0', '全体カット',              12, 'completed',   4),
    (6, 3, '2026-01-06', 10, '8.0',  '爪切り・ブラッシング',   12, 'reserved',    2),
    (7, 3, '2026-01-06', 15, '5.0',  'シャンプー',              6,  'completed',   1),
    (8, 3, '2026-01-06', 6,  '3.8',  'トリミング',              6,  'reserved',    3)
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('trimming_records', 'id'), (SELECT MAX(id) FROM trimming_records));

-- -----------------------------------------------------------------------------
-- 7. vaccinations（id 1-7, 全件 DO UPDATE）
-- 002 の id=1-3 を正しい値で上書き。id=4-7 は 003 追加分。
-- id=2: mr=8(チョコ/田中美咲), id=3: mr=16(ゴン/松本隆) + vaccine_id=1(混合5種犬) に修正
-- vaccine_id: 1=混合5種犬, 5=3種猫, 7=狂犬病, 8=フィラリア予防薬(小型犬)
-- -----------------------------------------------------------------------------
INSERT INTO vaccinations (id, clinic_id, medical_record_id, pet_id, vaccine_id, date, next_date, next_schedule_type, doctor_id, remarks) VALUES
    (1, 3, 1,  1,  1, '2026-02-15', '2027-02-15', '1year', 1, '接種後30分経過観察。異常なし。'),
    (2, 3, 8,  6,  7, '2026-03-05', '2027-03-05', '1year', 1, '狂犬病予防法に基づく接種。済票発行。'),
    (3, 3, 16, 14, 1, '2026-03-10', '2027-03-10', '1year', 2, '接種後の副反応なし。'),
    (4, 3, 7,  4,  8, '2025-08-22', '2025-09-22', 'other', 2, 'フィラリア予防薬投与。体重9kg未満のため小型犬用。'),
    (5, 3, 20, 18, 5, '2026-01-06', '2027-01-06', '1year', 1, '体調良好。ワクチン接種実施。'),
    (6, 3, 6,  3,  5, '2026-01-06', '2027-01-06', '1year', 1, '年次ワクチン接種。'),
    (7, 3, 15, 13, 5, '2026-01-06', '2027-01-06', '1year', 2, '体調良好。初回3種混合ワクチン。')
ON CONFLICT (id) DO UPDATE SET
    medical_record_id  = EXCLUDED.medical_record_id,
    pet_id             = EXCLUDED.pet_id,
    vaccine_id         = EXCLUDED.vaccine_id,
    date               = EXCLUDED.date,
    next_date          = EXCLUDED.next_date,
    next_schedule_type = EXCLUDED.next_schedule_type,
    doctor_id          = EXCLUDED.doctor_id,
    remarks            = EXCLUDED.remarks,
    updated_at         = now();

SELECT setval(pg_get_serial_sequence('vaccinations', 'id'), (SELECT MAX(id) FROM vaccinations));

-- -----------------------------------------------------------------------------
-- 8. hospitalizations（id 1-7, 全件 DO UPDATE）
-- 002 の id=1,2 を上書き: id=2 の pet_id を 10(ロッキー/中村勇気) → 8(ハチ/伊藤次郎) に修正
-- type: 'hospitalization' | 'hotel'
-- status: 'admitted' | 'discharged' | 'reserved'
-- -----------------------------------------------------------------------------
INSERT INTO hospitalizations (id, clinic_id, owner_id, pet_id, hospitalization_type, start_date, end_date, status, cage_id, doctor_id, memo, owner_request, staff_notes) VALUES
    (1, 3, 3, 5,  'hospitalization', '2026-03-10', '2026-03-14', 'admitted',   5,    1, '急性胃腸炎による脱水治療。点滴管理中。',  '食事のアレルギーに注意してほしい（鶏肉不可）', '3/10入院開始。静脈点滴開始。3/11嘔吐1回。3/12状態改善傾向。'),
    (2, 3, 6, 8,  'hospitalization', '2026-02-25', '2026-02-28', 'discharged', 4,    1, '外耳炎重症化に伴う入院治療。',             '怖がりなので優しく接してほしい',               '耳道洗浄を毎日実施。2/28退院時、症状改善。点耳薬処方。'),
    (3, 3, 17, 19, 'hospitalization', '2026-02-10', '2026-02-20', 'discharged', NULL, 1, '骨折治療による入院。手術後経過観察。', '', '2/10手術実施。2/15抜糸。2/20退院。'),
    (4, 3, 4,  6,  'hotel',           '2026-03-15', '2026-03-18', 'reserved',   NULL, 1, '旅行中のホテル預かり。', 'フードはロイヤルカナンのみ', ''),
    (5, 3, 1,  1,  'hospitalization', '2026-03-20', '2026-03-25', 'reserved',   NULL, 1, '膝蓋骨脱臼手術予定。術前検査済み。', '怖がりなので静かな環境を希望', ''),
    (6, 3, 9,  11, 'hospitalization', '2026-03-05', '2026-03-12', 'admitted',   6,    2, '慢性腎臓病の集中治療。点滴管理中。', 'ペルシャ猫のため温度管理に注意', '3/5入院。毎日皮下補液実施。3/12現在状態安定。'),
    (7, 3, 3,  4,  'hospitalization', '2026-01-03', '2026-01-06', 'discharged', NULL, 1, '急性胃腸炎による脱水治療。', 'チキンアレルギーあり', '1/3入院。点滴開始。1/6状態改善し退院。')
ON CONFLICT (id) DO UPDATE SET
    owner_id              = EXCLUDED.owner_id,
    pet_id                = EXCLUDED.pet_id,
    hospitalization_type  = EXCLUDED.hospitalization_type,
    start_date            = EXCLUDED.start_date,
    end_date              = EXCLUDED.end_date,
    status                = EXCLUDED.status,
    cage_id               = EXCLUDED.cage_id,
    doctor_id             = EXCLUDED.doctor_id,
    memo                  = EXCLUDED.memo,
    owner_request         = EXCLUDED.owner_request,
    staff_notes           = EXCLUDED.staff_notes,
    updated_at            = now();

SELECT setval(pg_get_serial_sequence('hospitalizations', 'id'), (SELECT MAX(id) FROM hospitalizations));

-- -----------------------------------------------------------------------------
-- 9. inventory_items（id 6-14, 9件追加）
-- 002 の 1-5 件に加え、mockData の inv-006〜inv-014 を追加。
-- category: 'medicine' | 'consumable' | 'food' | 'other'
-- status: 'sufficient' | 'low' | 'out_of_stock'
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
-- 10. billings / billing_items / payments（002 の不整合を修正）
-- billings id=2: owner_id 6→1(林文明), pet_id 10→1(Iris) ← medical_record=3(Iris/林文明) と一致
-- billings id=3: medical_record_id 2→6(ミケ/田中花子 finalized) ← owner=2/pet=3 と一致
-- payments id=1: billing_id 3→1(billing 1 total=4730 と一致)
-- -----------------------------------------------------------------------------
INSERT INTO billings (id, clinic_id, medical_record_id, hospitalization_id, owner_id, pet_id, subtotal, tax_total, total_amount, has_insurance, status, scheduled_date, completed_at, memo) VALUES
    (2, 3, 3,    NULL, 1,  1,  3300, 330, 3630, true, 'completed', '2026-02-28', '2026-02-28 11:00:00+09', 'アニコム保険適用（Iris 耳炎治療）'),
    (3, 3, 6,    NULL, 2,  3,  800,  80,  880,  true, 'waiting',   '2026-03-12', NULL,                     'アニコム保険適用。会計待ち。')
ON CONFLICT (id) DO UPDATE SET
    medical_record_id = EXCLUDED.medical_record_id,
    owner_id          = EXCLUDED.owner_id,
    pet_id            = EXCLUDED.pet_id,
    memo              = EXCLUDED.memo,
    updated_at        = now();

INSERT INTO payments (id, billing_id, subtotal, tax_total, total_amount, insurance_name, insurance_ratio, insurance_amount, discount_amount, billing_amount, received_amount, change_amount, method) VALUES
    (1, 1, 4300, 430, 4730, 'アニコム損保', 0.70, 3311, 0, 1419, 1500, 81, 'cash')
ON CONFLICT (id) DO UPDATE SET
    billing_id      = EXCLUDED.billing_id,
    updated_at      = now();

-- -----------------------------------------------------------------------------
-- 11. vital_records（002 の体重・記録日時を修正）
-- 003 の medical_records 修正後の正しいペット体重に合わせる
-- id=1-3: Iris(ゴールデンレトリーバー 26.5kg), id=4: Max(ラブラドール 15.2kg)
-- id=5: ミケ(3.8kg) は変更なし
-- pet_id: mr2→pet1(Iris), mr2→pet1(Iris), mr3→pet10, mr4→pet5, mr5→pet14
-- -----------------------------------------------------------------------------
INSERT INTO vital_records (id, pet_id, medical_record_id, recorded_at, staff_id, temperature, heart_rate, respiration_rate, weight, notes) VALUES
    (1, 1,  3, '2026-01-20 09:15:00+09', 1, 38.5, 80,  20, 26.5, '皮膚の搔痒感あり。体重良好。'),
    (2, 1,  2, '2025-12-15 10:00:00+09', 2, 38.8, 82,  22, 26.0, '体重前回比-500g'),
    (3, 1,  3, '2026-01-20 09:30:00+09', 1, 38.3, 78,  20, 26.5, '定期検診。皮膚搔痒感 軽快傾向。'),
    (4, 2,  4, '2025-11-05 11:00:00+09', 1, 39.1, 95,  24, 15.2, '軽度脱水。CRT 2秒。'),
    (5, 3,  5, '2025-09-15 14:30:00+09', 2, 38.2, 160, 30,  4.2, '粘膜色やや蒼白。食欲低下継続。')
ON CONFLICT (id) DO UPDATE SET
    pet_id            = EXCLUDED.pet_id,
    medical_record_id = EXCLUDED.medical_record_id,
    recorded_at       = EXCLUDED.recorded_at,
    temperature       = EXCLUDED.temperature,
    heart_rate        = EXCLUDED.heart_rate,
    respiration_rate  = EXCLUDED.respiration_rate,
    weight            = EXCLUDED.weight,
    notes             = EXCLUDED.notes;

SELECT setval(pg_get_serial_sequence('vital_records', 'id'), (SELECT MAX(id) FROM vital_records));

-- -----------------------------------------------------------------------------
-- 12. billing_refunds（返金デモデータ）
-- billing id=1 (林文明/Iris, total=4730): 2件の部分返金（合計 ¥1,419 → 残額 ¥0）
-- billing id=2 (林文明/Iris, total=3630): 1件の部分返金（¥500 → 残額 ¥3,130）
-- -----------------------------------------------------------------------------
INSERT INTO billing_refunds (id, clinic_id, billing_id, amount, reason, refunded_at) VALUES
    (1, 3, 1, 919,  '処置内容の変更に伴う部分返金',   '2026-02-16 10:00:00+09'),
    (2, 3, 1, 500,  '薬剤変更による差額返金',         '2026-02-20 14:30:00+09'),
    (3, 3, 2, 500,  '診察キャンセル分の返金',          '2026-03-01 09:00:00+09')
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('billing_refunds', 'id'), (SELECT MAX(id) FROM billing_refunds));
