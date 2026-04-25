-- LSTEP-BE-018: 既存飼い主データ一括同期の進捗管理テーブル
CREATE TABLE lstep_migration_progress (
    id            BIGSERIAL PRIMARY KEY,
    clinic_id     BIGINT      NOT NULL REFERENCES clinics(id),
    owner_id      BIGINT      NOT NULL REFERENCES owners(id),
    status        VARCHAR(20) NOT NULL DEFAULT 'pending',  -- pending | success | partial | failed | skipped
    tags_added    INT         NOT NULL DEFAULT 0,
    tags_failed   INT         NOT NULL DEFAULT 0,
    error_message TEXT,
    started_at    TIMESTAMPTZ,
    completed_at  TIMESTAMPTZ,
    UNIQUE (clinic_id, owner_id)
);

CREATE INDEX idx_lstep_migration_progress_clinic_id
    ON lstep_migration_progress (clinic_id);
