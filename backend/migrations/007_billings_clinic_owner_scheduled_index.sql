-- EMR-201: GET /api/v1/accountings?owner_id= のレイテンシ改善。
-- 一覧クエリは常に billings.clinic_id (認可済みスコープ) と deleted_at IS NULL を
-- 固定述語に持ち、owner_id 等値 + scheduled_date/created_at DESC ソートを行う。
-- 既存 idx_billings_owner_id は clinic スコープ・soft-delete・ソート列を含まない
-- 単一列インデックスのため、clinic 先頭の複合部分インデックスを追加する。
-- (照会パターン: WHERE clinic_id IN (...) AND owner_id = ? AND deleted_at IS NULL
--  ORDER BY scheduled_date DESC, created_at DESC LIMIT n)
CREATE INDEX IF NOT EXISTS idx_billings_clinic_owner_scheduled
    ON billings (clinic_id, owner_id, scheduled_date DESC, created_at DESC)
    WHERE deleted_at IS NULL;
