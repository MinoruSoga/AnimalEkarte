-- 011_add_appointment_checked_in_at.sql
--
-- 由来: 旧 005_add_appointment_checked_in_at.sql（2026-07-04 に 001_init.sql §7 へ畳み込み済み）。
-- 目的: schema_migrations に薄い/統合前の 001_init.sql のみが記録されている DB へ、
--       畳み込み DDL を additive incremental として適用する upgrade path。
-- 冪等: fresh DB（001 統合スキーマ適用済み）でも IF NOT EXISTS / duplicate_object ガードで通る。
-- 注意: 001_init.sql 本文は変更しない（checksum 維持）。002_checkup_field_clinic_composite_fk.sql
--       （checkup_types↔checkup_type_fields 複合 FK）とは重複しない。
-- 注: 旧番号 005 は lab_import 再採番と衝突した時期の別名。論理由来は checked_in_at 単体ファイル。

-- 受付ヘッダー テレメトリ表示（change-ui.md Phase 2）: 待ち時間算出のための受付済み遷移時刻を記録する。
-- appointments.updated_at は autoUpdateTime のため予約編集全般でリセットされ待ち時間算出に流用できない。
-- そのため checked_in ステータスへ遷移した時点の時刻専用カラムを新設する。

ALTER TABLE appointments
    ADD COLUMN IF NOT EXISTS checked_in_at timestamptz;
