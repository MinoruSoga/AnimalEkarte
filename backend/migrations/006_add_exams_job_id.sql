-- 006_add_exams_job_id.sql
--
-- 由来: 旧 008_add_exams_job_id.sql（2026-07-04 に 001_init.sql §7 へ畳み込み済み）。
-- 目的: schema_migrations に薄い/統合前の 001_init.sql のみが記録されている DB へ、
--       畳み込み DDL を additive incremental として適用する upgrade path。
-- 冪等: fresh DB（001 統合スキーマ適用済み）でも IF NOT EXISTS / duplicate_object ガードで通る。
-- 注意: 001_init.sql 本文は変更しない（checksum 維持）。002_checkup_field_clinic_composite_fk.sql
--       （checkup_types↔checkup_type_fields 複合 FK）とは重複しない。

-- Phase 4B.2: exams.job_id nullable FK to lab_import_jobs
--
-- Decision (Phase 4B.1): ADD as uuid NULL with ON DELETE SET NULL so that
-- job deletion does not cascade-delete exam rows (business data must be preserved).
-- Nullable to remain backward-compatible with hand-created exams (NULL = no import job).

DO $$
BEGIN
    ALTER TABLE exams
        ADD COLUMN job_id uuid NULL
        REFERENCES lab_import_jobs(id) ON DELETE SET NULL;
EXCEPTION
    WHEN duplicate_column THEN NULL;
END $$;

-- Index for ListJobReportSummaries: "give me all exams for this job under this clinic"
-- clinic_id + job_id covers the primary query access pattern.
-- Partial index (WHERE job_id IS NOT NULL) keeps it small for hand-created exams.
CREATE INDEX IF NOT EXISTS idx_exams_clinic_job
    ON exams (clinic_id, job_id)
    WHERE job_id IS NOT NULL;

COMMENT ON COLUMN exams.job_id IS 'lab_import_jobs.id FK — NULL for hand-created exams. ON DELETE SET NULL preserves exam rows when job is deleted (Phase 4B.2).';
