-- 008_add_lstep_sync_error_counters.sql
-- Lステップ連携エラーカウンター: 飼い主ごとの API 失敗回数を追跡する

CREATE TABLE IF NOT EXISTS lstep_sync_error_counters (
    id            bigserial   PRIMARY KEY,
    clinic_id     bigint      NOT NULL REFERENCES clinics(id),
    owner_id      bigint      NOT NULL REFERENCES owners(id),
    failure_count int         NOT NULL DEFAULT 0,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    UNIQUE (clinic_id, owner_id)
);

CREATE INDEX IF NOT EXISTS idx_lstep_sync_error_counters_clinic_owner
    ON lstep_sync_error_counters(clinic_id, owner_id);
