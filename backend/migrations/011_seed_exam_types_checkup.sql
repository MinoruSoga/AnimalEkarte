-- #160: 健康診断カテゴリを exam_types/exam_type_fields マスタに投入
-- 要望: 健診を既存 検査(Examination) 機能で管理できるよう、
--       親「健康診断」+ 子カテゴリ + 検査項目テンプレートを clinic 1 に追加する。
--
-- 設計判断: 健診(Checkup)を新規スキーマで拡張せず、既存 exam_types 階層に統合する（案A採用）。
--   parent_id でカテゴリ階層化、exam_type_fields で検査項目テンプレートを定義。
--   ExamItemsTable.tsx が既存の階層描画・異常判定に対応しているためコード変更不要。
--
-- ID 採番:
--   exam_types     : 12000-12003 (既存上限 11008 と衝突なし)
--   exam_type_fields: 45-58      (既存上限 44 と衝突なし)

-- -----------------------------------------------------------------------------
-- exam_types: 健康診断 カテゴリ（clinic 1）
-- -----------------------------------------------------------------------------
INSERT INTO exam_types (id, clinic_id, name, price, is_active, description, parent_id, sort_order, is_non_insurance) VALUES
    -- 親カテゴリ
    (12000, 1, '健康診断', 0, true, '健康診断カテゴリ', NULL, 20, false),
    -- 子カテゴリ
    (12001, 1, '一般健診',   0, true, '基本バイタル・全身状態の確認',         12000, 1, false),
    (12002, 1, '歯科健診',   0, true, '口腔・歯科状態の確認',                 12000, 2, false),
    (12003, 1, '老齢健診',   0, true, '7歳以上の高齢動物向け総合健康チェック', 12000, 3, false)
ON CONFLICT (id) DO UPDATE
    SET clinic_id        = EXCLUDED.clinic_id,
        name             = EXCLUDED.name,
        price            = EXCLUDED.price,
        is_active        = EXCLUDED.is_active,
        description      = EXCLUDED.description,
        parent_id        = EXCLUDED.parent_id,
        sort_order       = EXCLUDED.sort_order,
        is_non_insurance = EXCLUDED.is_non_insurance,
        updated_at       = now();

SELECT setval(pg_get_serial_sequence('exam_types', 'id'), (SELECT MAX(id) FROM exam_types));

-- -----------------------------------------------------------------------------
-- exam_type_fields: 各健診カテゴリの検査項目テンプレート
-- -----------------------------------------------------------------------------

-- 一般健診（12001）: 基本バイタル項目
INSERT INTO exam_type_fields (id, exam_type_id, name, inspection_value, normal_value, unit, sort_order) VALUES
    (45, 12001, '体重',         '', '',          'kg',    1),
    (46, 12001, '体温',         '', '38.0-39.2', '℃',    2),
    (47, 12001, '心拍数',       '', '60-180',    'bpm',   3),
    (48, 12001, '呼吸数',       '', '10-30',     '回/分', 4),
    (49, 12001, '口腔粘膜色',   '', 'ピンク',    '',      5),
    (50, 12001, '体型スコア',   '', '3/5',       '',      6)

ON CONFLICT (id) DO UPDATE SET
    exam_type_id     = EXCLUDED.exam_type_id,
    name             = EXCLUDED.name,
    inspection_value = EXCLUDED.inspection_value,
    normal_value     = EXCLUDED.normal_value,
    unit             = EXCLUDED.unit,
    sort_order       = EXCLUDED.sort_order;

-- 歯科健診（12002）: 口腔・歯科項目
INSERT INTO exam_type_fields (id, exam_type_id, name, inspection_value, normal_value, unit, sort_order) VALUES
    (51, 12002, '歯石付着度',   '', 'グレード0', '',  1),
    (52, 12002, '歯肉炎程度',   '', 'グレード0', '',  2),
    (53, 12002, '口臭',         '', '陰性',      '',  3),
    (54, 12002, '欠損歯数',     '', '0',         '本', 4)

ON CONFLICT (id) DO UPDATE SET
    exam_type_id     = EXCLUDED.exam_type_id,
    name             = EXCLUDED.name,
    inspection_value = EXCLUDED.inspection_value,
    normal_value     = EXCLUDED.normal_value,
    unit             = EXCLUDED.unit,
    sort_order       = EXCLUDED.sort_order;

-- 老齢健診（12003）: 高齢動物向け検査項目
INSERT INTO exam_type_fields (id, exam_type_id, name, inspection_value, normal_value, unit, sort_order) VALUES
    (55, 12003, '関節可動域',   '', '正常',  '',  1),
    (56, 12003, '筋肉量評価',   '', '正常',  '',  2),
    (57, 12003, '認知機能評価', '', '正常',  '',  3),
    (58, 12003, '視力評価',     '', '異常なし', '', 4)

ON CONFLICT (id) DO UPDATE SET
    exam_type_id     = EXCLUDED.exam_type_id,
    name             = EXCLUDED.name,
    inspection_value = EXCLUDED.inspection_value,
    normal_value     = EXCLUDED.normal_value,
    unit             = EXCLUDED.unit,
    sort_order       = EXCLUDED.sort_order;

SELECT setval(pg_get_serial_sequence('exam_type_fields', 'id'), (SELECT MAX(id) FROM exam_type_fields));
