-- EMR-308: tentative appointment (次回予定) schema foundation — part 2 of 2.
-- Partial unique index guaranteeing at most one ACTIVE tentative appointment
-- per (clinic_id, pet_id). Rows whose status is not 'tentative', rows with
-- deleted_at set, and rows with a NULL pet_id never conflict.
--
-- Requires 028_appointments_tentative_status.sql to have committed the
-- 'tentative' enum value first: a new enum value cannot be referenced within
-- the transaction that added it, and cmd/migrate runs one transaction per file
-- (007/008 precedent). Do not write BEGIN;/COMMIT; inside this file.
-- Apply is manual (make migrate). Do not edit 001_init.sql.

CREATE UNIQUE INDEX uq_appointments_one_active_tentative
    ON appointments (clinic_id, pet_id)
    WHERE status = 'tentative' AND deleted_at IS NULL;
