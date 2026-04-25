-- 005_lstep_clinic_integrations.sql
-- Lステップ/LINE連携設定保存テーブル

CREATE TABLE clinic_integrations (
    id          BIGSERIAL PRIMARY KEY,
    clinic_id   BIGINT NOT NULL REFERENCES clinics(id),
    service     VARCHAR(50)  NOT NULL,
    key_name    VARCHAR(100) NOT NULL,
    key_value   TEXT         NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (clinic_id, service, key_name)
);

CREATE INDEX idx_clinic_integrations_clinic_service
    ON clinic_integrations (clinic_id, service);
