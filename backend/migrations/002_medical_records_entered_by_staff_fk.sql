-- Replace composite entered_by×clinic FK with entered_by→staffs(id).
-- Assigned creators may record on a non-home clinic; clinic authz is enforced in app code.
-- Do not edit 001_init.sql. User must run: make migrate

ALTER TABLE medical_records
    DROP CONSTRAINT IF EXISTS fk_medical_records_entered_by_clinic,
    DROP CONSTRAINT IF EXISTS medical_records_entered_by_fkey;

ALTER TABLE medical_records
    ADD CONSTRAINT fk_medical_records_entered_by
        FOREIGN KEY (entered_by)
        REFERENCES staffs (id)
        ON DELETE RESTRICT;
