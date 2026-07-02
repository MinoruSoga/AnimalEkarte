-- #215: AM 開始時刻（既定 09:00）を clinic_settings に追加する（additive）。
-- 締めレンジは AM=[am_start, boundary) / PM=[boundary, pm_end) / EMG=[pm_end, 翌日 am_start) になり、
-- 深夜 0:00〜am_start の緊急会計は前日の EMG に帰属する。
-- 既存の締め記録（cash_register_closes のスナップショット）は再計算しない（過去データ非破壊）。
ALTER TABLE clinic_settings
    ADD COLUMN IF NOT EXISTS closing_am_start time NOT NULL DEFAULT '09:00';
