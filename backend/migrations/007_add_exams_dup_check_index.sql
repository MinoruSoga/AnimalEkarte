-- 007_add_exams_dup_check_index.sql
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
-- Uses CONCURRENTLY so the build does not lock exams against concurrent writes.
-- CONCURRENTLY cannot run inside an explicit transaction block; if the migration runner
-- wraps scripts in a transaction, add a `-- no-transaction` pragma or split this file.

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_exams_clinic_exam_type_date
    ON exams (clinic_id, exam_type_id, date)
    WHERE deleted_at IS NULL;

COMMENT ON INDEX idx_exams_clinic_exam_type_date
    IS 'Phase 2: LabImportDuplicateCheckerDB (clinic_id, exam_type_id, date) lookup';
