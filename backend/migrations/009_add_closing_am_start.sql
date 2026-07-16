-- 009_add_closing_am_start.sql
--
-- 由来: 旧 011_add_closing_am_start.sql（2026-07-04 に 001_init.sql §7 へ畳み込み済み）。
-- 目的: schema_migrations に薄い/統合前の 001_init.sql のみが記録されている DB へ、
--       畳み込み DDL を additive incremental として適用する upgrade path。
-- 冪等: fresh DB（001 統合スキーマ適用済み）でも IF NOT EXISTS / duplicate_object ガードで通る。
-- 注意: 001_init.sql 本文は変更しない（checksum 維持）。002_checkup_field_clinic_composite_fk.sql
--       （checkup_types↔checkup_type_fields 複合 FK）とは重複しない。

-- #215: AM 開始時刻（既定 09:00）を clinic_settings に追加する（additive）。
-- 締めレンジは AM=[am_start, boundary) / PM=[boundary, pm_end) / EMG=[pm_end, 翌日 am_start) になり、
-- 深夜 0:00〜am_start の緊急会計は前日の EMG に帰属する。
-- 既存の締め記録（cash_register_closes のスナップショット）は再計算しない（過去データ非破壊）。
ALTER TABLE clinic_settings
    ADD COLUMN IF NOT EXISTS closing_am_start time NOT NULL DEFAULT '09:00';
