-- 004_add_exam_results_exam_id_index.sql
--
-- 由来: 旧 006_add_exam_results_exam_id_index.sql（2026-07-04 に 001_init.sql §7 へ畳み込み済み）。
-- 目的: schema_migrations に薄い/統合前の 001_init.sql のみが記録されている DB へ、
--       畳み込み DDL を additive incremental として適用する upgrade path。
-- 冪等: fresh DB（001 統合スキーマ適用済み）でも IF NOT EXISTS / duplicate_object ガードで通る。
-- 注意: 001_init.sql 本文は変更しない（checksum 維持）。002_checkup_field_clinic_composite_fk.sql
--       （checkup_types↔checkup_type_fields 複合 FK）とは重複しない。

-- Phase 2: exam_results.exam_id index for lab import batch performance.
--
-- ReplaceItemsByExamID (used by LabImportExaminationService) runs a DELETE + INSERT
-- keyed on exam_id. Without this index each call is a full table scan on exam_results.
-- Phase 1 comment noted this migration must be applied before large-batch lab import runs.
--
-- Note: CONCURRENTLY was removed because the migration runner wraps each file in an
-- explicit transaction (cmd/migrate/main.go:tx.Begin/Exec/Commit), and PostgreSQL
-- does not allow CREATE INDEX CONCURRENTLY inside a transaction block. The tables
-- created in this phase are new (005_add_lab_import_tables.sql) with no concurrent
-- traffic at migration time, so a plain CREATE INDEX is safe.
-- IF NOT EXISTS makes the migration idempotent.
--
-- Phase 3A decision: no DB unique constraint on (clinic_id, exam_type_id, date, pet_id).
-- Local data check showed 87 duplicate groups (95 extra rows) in migrated data; these
-- are legitimate multi-visit records with distinct medical_record_ids, not true import
-- duplicates. Duplicate prevention is enforced at service level via LabImportDuplicateCheckerDB.
-- Production data must be verified before adding any DB unique constraint.
-- See: docs/lab-go/app-integration-boundary.md Phase 3A section.

CREATE INDEX IF NOT EXISTS idx_exam_results_exam_id
    ON exam_results (exam_id);

COMMENT ON INDEX idx_exam_results_exam_id
    IS 'Phase 2: supports ReplaceItemsByExamID DELETE+INSERT in lab import batches';
