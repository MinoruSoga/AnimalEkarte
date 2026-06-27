-- 008_add_exams_job_id.sql
-- Phase 4B.2: exams.job_id nullable FK to lab_import_jobs
--
-- Decision (Phase 4B.1): ADD as uuid NULL with ON DELETE SET NULL so that
-- job deletion does not cascade-delete exam rows (business data must be preserved).
-- Nullable to remain backward-compatible with hand-created exams (NULL = no import job).

ALTER TABLE exams
    ADD COLUMN job_id uuid NULL
    REFERENCES lab_import_jobs(id) ON DELETE SET NULL;

-- Index for ListJobReportSummaries: "give me all exams for this job under this clinic"
-- clinic_id + job_id covers the primary query access pattern.
-- Partial index (WHERE job_id IS NOT NULL) keeps it small for hand-created exams.
CREATE INDEX idx_exams_clinic_job
    ON exams (clinic_id, job_id)
    WHERE job_id IS NOT NULL;

COMMENT ON COLUMN exams.job_id IS 'lab_import_jobs.id FK — NULL for hand-created exams. ON DELETE SET NULL preserves exam rows when job is deleted (Phase 4B.2).';
