-- 007_add_lstep_sync_settings.sql
-- Lステップ同期設定テーブル: クリニックごとの同期有効フラグと有効化日時

CREATE TABLE IF NOT EXISTS lstep_settings (
    id               bigserial    PRIMARY KEY,
    clinic_id        bigint       NOT NULL UNIQUE REFERENCES clinics(id),
    is_sync_enabled  boolean      NOT NULL DEFAULT false,
    sync_enabled_at  timestamptz  NULL,
    created_at       timestamptz  NOT NULL DEFAULT now(),
    updated_at       timestamptz  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_lstep_settings_clinic_id ON lstep_settings(clinic_id);
