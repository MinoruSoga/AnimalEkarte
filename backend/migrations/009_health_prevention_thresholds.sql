-- P9: 健診・予防タグ判定の日数閾値を clinic 単位設定化
-- health_prevention_lookback_days: 健診・予防履歴の参照期間（日数）, default 365
-- vaccine_deadline_days: ワクチン期限間近とみなす残日数, default 60
ALTER TABLE clinic_settings
    ADD COLUMN IF NOT EXISTS health_prevention_lookback_days INT NOT NULL DEFAULT 365,
    ADD COLUMN IF NOT EXISTS vaccine_deadline_days INT NOT NULL DEFAULT 60;
