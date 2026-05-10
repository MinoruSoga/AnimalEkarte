ALTER TABLE clinic_settings
    ADD COLUMN IF NOT EXISTS cpm_version VARCHAR(8) NOT NULL DEFAULT 'v1'
        CHECK (cpm_version IN ('v1', 'v2'));

COMMENT ON COLUMN clinic_settings.cpm_version IS 'CPM 判定方式 (v1: 既存 6-stage / v2: Q19 来院回数 5-stage, 2026-05-08 確定)';
