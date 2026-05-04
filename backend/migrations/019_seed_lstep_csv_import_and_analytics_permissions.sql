-- =============================================================================
-- FEAT-385: Lステップ CSV インポート・分析 permission_group_rules シード
-- 対象リソース: lstep-csv-import, lstep-analytics
-- 既存 6 グループ (group_id=1〜6) に追加
-- ON CONFLICT DO NOTHING で再投入耐性あり
-- =============================================================================

INSERT INTO permission_group_rules (group_id, resource, can_view, can_create, can_edit, can_delete) VALUES
    -- 八王子病院 執行（group_id=1）
    (1, 'lstep-csv-import', true, true,  true,  true),
    (1, 'lstep-analytics',  true, false, false, false),
    -- 八王子病院 一般（group_id=2）
    (2, 'lstep-csv-import', true, false, false, false),
    (2, 'lstep-analytics',  true, false, false, false),
    -- 城東センター病院 執行（group_id=3）
    (3, 'lstep-csv-import', true, true,  true,  true),
    (3, 'lstep-analytics',  true, false, false, false),
    -- 城東センター病院 一般（group_id=4）
    (4, 'lstep-csv-import', true, false, false, false),
    (4, 'lstep-analytics',  true, false, false, false),
    -- 敷島医院 執行（group_id=5）
    (5, 'lstep-csv-import', true, true,  true,  true),
    (5, 'lstep-analytics',  true, false, false, false),
    -- 敷島医院 一般（group_id=6）
    (6, 'lstep-csv-import', true, false, false, false),
    (6, 'lstep-analytics',  true, false, false, false)
ON CONFLICT DO NOTHING;
