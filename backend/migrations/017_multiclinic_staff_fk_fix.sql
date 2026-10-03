-- Fix composite (staff_col, clinic_id) -> staffs(id, clinic_id) FKs that wrongly
-- require the referenced staff's PRIMARY staffs.clinic_id to equal the row's
-- clinic_id. Membership is assignment-based (staff_clinic_assignments): a staff
-- legitimately assigned to a non-home clinic could not be referenced there —
-- the composite FK rejected the insert with 23503 surfaced to users as
-- 「参照先が存在しません」 (EMR-114 reproduction on appointments.doctor_id).
-- App-layer clinic authorization stays authoritative (same precedent as
-- 002_medical_records_entered_by_staff_fk.sql / 003_appointments_created_by_staff_fk.sql).
-- Each constraint keeps its original ON DELETE semantics; only the incorrect
-- clinic-match requirement is removed.
-- Do not edit 001_init.sql. User must run: make migrate

ALTER TABLE appointments
    DROP CONSTRAINT IF EXISTS fk_appointments_doctor_clinic;
ALTER TABLE appointments
    ADD CONSTRAINT fk_appointments_doctor
        FOREIGN KEY (doctor_id) REFERENCES staffs (id) ON DELETE SET NULL;

ALTER TABLE hospitalizations
    DROP CONSTRAINT IF EXISTS fk_hospitalizations_doctor_clinic;
ALTER TABLE hospitalizations
    ADD CONSTRAINT fk_hospitalizations_doctor
        FOREIGN KEY (doctor_id) REFERENCES staffs (id) ON DELETE SET NULL;

ALTER TABLE medical_records
    DROP CONSTRAINT IF EXISTS fk_medical_records_doctor_clinic;
ALTER TABLE medical_records
    ADD CONSTRAINT fk_medical_records_doctor
        FOREIGN KEY (doctor_id) REFERENCES staffs (id) ON DELETE SET NULL;

ALTER TABLE cash_register_close_adjustments
    DROP CONSTRAINT IF EXISTS fk_cash_register_close_adjustments_actor_clinic;
ALTER TABLE cash_register_close_adjustments
    ADD CONSTRAINT fk_cash_register_close_adjustments_actor
        FOREIGN KEY (actor_id) REFERENCES staffs (id) ON DELETE RESTRICT;

ALTER TABLE medical_record_image_upload_quota
    DROP CONSTRAINT IF EXISTS fk_medical_record_image_upload_quota_staff_clinic;
ALTER TABLE medical_record_image_upload_quota
    ADD CONSTRAINT fk_medical_record_image_upload_quota_staff
        FOREIGN KEY (staff_id) REFERENCES staffs (id) ON DELETE RESTRICT;

ALTER TABLE lab_device_waits
    DROP CONSTRAINT IF EXISTS fk_lab_device_waits_staff_clinic;
ALTER TABLE lab_device_waits
    ADD CONSTRAINT fk_lab_device_waits_staff
        FOREIGN KEY (staff_id) REFERENCES staffs (id) ON DELETE RESTRICT;
