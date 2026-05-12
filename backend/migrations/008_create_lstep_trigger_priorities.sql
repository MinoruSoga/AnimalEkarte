-- Q23 配信衝突優先順位制御 (PO-QA 2026-05-08 B採用 10段階)
-- +migrate Up
CREATE TABLE lstep_trigger_priorities (
    id           BIGSERIAL PRIMARY KEY,
    clinic_id    BIGINT NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    trigger_type VARCHAR(64) NOT NULL,
    priority     INTEGER NOT NULL CHECK (priority >= 1),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (clinic_id, trigger_type)
);

CREATE INDEX idx_lstep_trigger_priorities_clinic ON lstep_trigger_priorities(clinic_id);

COMMENT ON TABLE lstep_trigger_priorities IS 'Q23 配信トリガー優先順位 (clinic単位カスタマイズ可)';
COMMENT ON COLUMN lstep_trigger_priorities.priority IS '小さいほど優先 (1=最優先)。同日複数トリガー発火時、MIN(priority) のみ実配信';

-- +migrate Down
DROP TABLE IF EXISTS lstep_trigger_priorities;
