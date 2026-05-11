-- Q21 dormant prevention 閾値の clinic_settings 外部化 (SPEC-004 / PO-QA 2026-05-08)
-- clinic 単位で 4 段階閾値を調整可能にする。デフォルトは SPEC-004 整合値。
-- 既存行には PostgreSQL ALTER TABLE ADD COLUMN の仕様により DEFAULT 値が即時適用される。

ALTER TABLE clinic_settings
    ADD COLUMN dormant_prevention_180_days INTEGER NOT NULL DEFAULT 180,
    ADD COLUMN dormant_prevention_210_days INTEGER NOT NULL DEFAULT 210,
    ADD COLUMN dormant_prevention_240_days INTEGER NOT NULL DEFAULT 240,
    ADD COLUMN dormant_prevention_365_days INTEGER NOT NULL DEFAULT 365;

COMMENT ON COLUMN clinic_settings.dormant_prevention_180_days IS 'dormant_prevention_1st 配信トリガー閾値日数 (Q21、デフォルト 180)';
COMMENT ON COLUMN clinic_settings.dormant_prevention_210_days IS 'dormant_prevention_2nd 配信トリガー閾値日数 (Q21、デフォルト 210)';
COMMENT ON COLUMN clinic_settings.dormant_prevention_240_days IS 'dormant_prevention_3rd 配信トリガー閾値日数 (Q21、デフォルト 240)';
COMMENT ON COLUMN clinic_settings.dormant_prevention_365_days IS 'dormant_prevention_4th 配信トリガー閾値日数 (Q21、デフォルト 365)';

-- down migration (rollback):
-- ALTER TABLE clinic_settings
--     DROP COLUMN dormant_prevention_180_days,
--     DROP COLUMN dormant_prevention_210_days,
--     DROP COLUMN dormant_prevention_240_days,
--     DROP COLUMN dormant_prevention_365_days;
