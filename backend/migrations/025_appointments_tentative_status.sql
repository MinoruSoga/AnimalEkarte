-- EMR-308: tentative appointment (次回予定) schema foundation — part 1 of 2.
-- Adds 'tentative' to reservation_status and the two nullable provenance columns
-- used by tentative appointments (date known; time, slot and doctor undecided;
-- no capacity consumption; staff-visible only).
--
-- PostgreSQL constraint: an enum value added by ALTER TYPE ... ADD VALUE cannot
-- be referenced inside the same transaction, and cmd/migrate wraps each file in
-- one transaction. Following the 007/008 precedent, the value lands here while
-- the value-referencing DDL ships in the next file:
--   026_appointments_one_active_tentative_index.sql creates
--     UNIQUE INDEX uq_appointments_one_active_tentative
--     ON appointments (clinic_id, pet_id)
--     WHERE status = 'tentative' AND deleted_at IS NULL
--   so at most one ACTIVE tentative exists per (clinic_id, pet_id).
-- The columns added below do not reference the new enum value and are safe in
-- this transaction. Do not write BEGIN;/COMMIT; inside this file.
-- Apply is manual (make migrate). Do not edit 001_init.sql.

ALTER TYPE reservation_status ADD VALUE IF NOT EXISTS 'tentative';

ALTER TABLE appointments
    ADD COLUMN intended_date date NULL,
    ADD COLUMN source_medical_record_id bigint NULL REFERENCES medical_records(id) ON DELETE SET NULL;

COMMENT ON COLUMN appointments.intended_date IS
    'EMR-308 tentative appointment: intended visit date (date known; time, slot and doctor undecided).';
COMMENT ON COLUMN appointments.source_medical_record_id IS
    'EMR-308 tentative appointment: medical record whose next-visit intent created this row. NULL when created elsewhere.';
