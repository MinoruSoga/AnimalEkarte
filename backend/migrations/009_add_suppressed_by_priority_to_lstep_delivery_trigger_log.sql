-- Q23 抑制トリガー記録 (priority重複排除でスキップされた配信をaudit用に保持)
-- +migrate Up
ALTER TABLE lstep_delivery_trigger_log
    ADD COLUMN suppressed_by_priority BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN suppression_reason VARCHAR(255) NULL;

CREATE INDEX idx_lstep_delivery_trigger_log_suppressed
    ON lstep_delivery_trigger_log(clinic_id, suppressed_by_priority)
    WHERE suppressed_by_priority = TRUE;

COMMENT ON COLUMN lstep_delivery_trigger_log.suppressed_by_priority IS 'Q23 優先順位により抑制されたか (FALSE=実配信 / TRUE=ログのみ)';
COMMENT ON COLUMN lstep_delivery_trigger_log.suppression_reason IS '抑制理由 (例: "owner_id=42 already triggered by dormant_365d at 2026-05-11")';

-- +migrate Down
DROP INDEX IF EXISTS idx_lstep_delivery_trigger_log_suppressed;
ALTER TABLE lstep_delivery_trigger_log
    DROP COLUMN IF EXISTS suppression_reason,
    DROP COLUMN IF EXISTS suppressed_by_priority;
