-- 005_add_exams_dup_check_index.sql
--
-- 由来: 旧 007_add_exams_dup_check_index.sql（2026-07-04 に 001_init.sql §7 へ畳み込み済み）。
-- 目的: schema_migrations に薄い/統合前の 001_init.sql のみが記録されている DB へ、
--       畳み込み DDL を additive incremental として適用する upgrade path。
-- 冪等: fresh DB（001 統合スキーマ適用済み）でも IF NOT EXISTS / duplicate_object ガードで通る。
-- 注意: 001_init.sql 本文は変更しない（checksum 維持）。002_checkup_field_clinic_composite_fk.sql
--       （checkup_types↔checkup_type_fields 複合 FK）とは重複しない。

-- Phase 2: composite index for LabImportDuplicateCheckerDB hot path.
--
-- IsDuplicate queries exams with:
--   WHERE clinic_id = ? AND exam_type_id = ? AND date = ? AND deleted_at IS NULL
-- Without a composite index PostgreSQL performs a bitmap-AND over individual single-column
-- indexes (idx_exams_clinic_id, idx_exams_exam_type_id), which degrades toward a seq scan
-- on any table with more than a few thousand rows.
--
-- Column order: clinic_id (most selective for multi-tenant), exam_type_id, date (equality).
-- pet_id is handled by Go-side NULL branching and cannot be added to a single composite
-- index without expression tricks; the index narrows to (clinic_id, exam_type_id, date)
-- first and pet_id is applied as a recheck predicate.
--
-- Partial index on deleted_at IS NULL keeps the index smaller (soft-deleted exams are
-- excluded from import duplicate checks by design).
--
-- Note: CONCURRENTLY was removed because the migration runner wraps each file in an
-- explicit transaction (cmd/migrate/main.go:tx.Begin/Exec/Commit), and PostgreSQL
-- does not allow CREATE INDEX CONCURRENTLY inside a transaction block. The exams table
-- exists with migrated data but lab import is not yet live, so a plain CREATE INDEX is
-- safe at migration time.
--
-- Phase 3A decision: no DB unique constraint added.
-- 87 duplicate groups exist in local migrated data on the 4-column key
-- (clinic_id, exam_type_id, date, pet_id). 84/85 non-null groups have distinct
-- medical_record_ids (same pet, different karte visits on the same day) and are
-- legitimate. A DB unique constraint on this key would reject valid historical records.
-- Service-level duplicate prevention is the formal policy until production data is
-- verified and a 5-column partial unique index (adding medical_record_id) can be assessed.
-- See: docs/lab-go/app-integration-boundary.md Phase 3A section.

CREATE INDEX IF NOT EXISTS idx_exams_clinic_exam_type_date
    ON exams (clinic_id, exam_type_id, date)
    WHERE deleted_at IS NULL;

COMMENT ON INDEX idx_exams_clinic_exam_type_date
    IS 'Phase 2/3A: LabImportDuplicateCheckerDB (clinic_id, exam_type_id, date) lookup; no unique constraint — see Phase 3A decision';
