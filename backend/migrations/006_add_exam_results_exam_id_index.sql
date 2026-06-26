-- 006_add_exam_results_exam_id_index.sql
-- Phase 2: exam_results.exam_id index for lab import batch performance.
--
-- ReplaceItemsByExamID (used by LabImportExaminationService) runs a DELETE + INSERT
-- keyed on exam_id. Without this index each call is a full table scan on exam_results.
-- Phase 1 comment noted this migration must be applied before large-batch lab import runs.
--
-- Uses CONCURRENTLY so the build does not lock exam_results against concurrent writes.
-- CONCURRENTLY cannot run inside an explicit transaction block; if the migration runner
-- wraps scripts in a transaction, add a `-- no-transaction` pragma or split this file.
-- IF NOT EXISTS makes the migration idempotent.
--
-- No unique constraint is added: duplicate prevention for
-- (clinic_id, exam_type_id, date, pet_id) is enforced at service level.
-- A DB-level unique constraint requires confirming zero violations in production data;
-- that check is deferred to Phase 3.

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_exam_results_exam_id
    ON exam_results (exam_id);

COMMENT ON INDEX idx_exam_results_exam_id
    IS 'Phase 2: supports ReplaceItemsByExamID DELETE+INSERT in lab import batches';
