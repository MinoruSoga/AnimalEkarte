-- created_by is the authenticated recording actor, not the assigned doctor.
-- A staff member may register appointments at an assigned clinic other than
-- their home clinic; system administrators may operate without an assignment.
-- The reservation writer validates and locks the active actor's authorization
-- in the insert transaction. Keep historical attribution and prevent deletion.
ALTER TABLE appointments
    DROP CONSTRAINT IF EXISTS fk_appointments_created_by_clinic,
    DROP CONSTRAINT IF EXISTS appointments_created_by_fkey;

ALTER TABLE appointments
    ADD CONSTRAINT fk_appointments_created_by
        FOREIGN KEY (created_by) REFERENCES staffs (id) ON DELETE RESTRICT;
