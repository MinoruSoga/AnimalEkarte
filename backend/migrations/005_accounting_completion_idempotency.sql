-- 005_accounting_completion_idempotency.sql
-- BUG-018: atomic accounting completion idempotency keys on billings.
-- 適用: USER が make migrate を手動実行すること（エージェントは auto-apply しない）。
-- CASCADE DELETE 禁止。既存 migration (001-004) は編集しない。
--
-- 方針:
--   1. completion_request_id / completion_request_hash を billings に追加
--   2. clinic-first UNIQUE で同一キーの再利用を拒否（soft-delete 後も再利用不可）
--   3. 部分 UNIQUE にしない（deleted_at を WHERE から外し、論理削除後の key 再利用を塞ぐ）

ALTER TABLE billings
    ADD COLUMN IF NOT EXISTS completion_request_id   UUID,
    ADD COLUMN IF NOT EXISTS completion_request_hash TEXT;

-- clinic-first composite unique（prompt 契約: UNIQUE(clinic_id, completion_request_id)）
-- NULL completion_request_id は PostgreSQL の UNIQUE で複数許可（legacy create 経路）
CREATE UNIQUE INDEX IF NOT EXISTS uq_billings_clinic_completion_request_id
    ON billings (clinic_id, completion_request_id);

COMMENT ON COLUMN billings.completion_request_id IS
    'BUG-018: POST /accountings/complete の Idempotency-Key（UUID）。soft-delete 後も再利用不可。';
COMMENT ON COLUMN billings.completion_request_hash IS
    'BUG-018: complete request の正規化 digest（hex）。同一 key で異 digest は 409。';
