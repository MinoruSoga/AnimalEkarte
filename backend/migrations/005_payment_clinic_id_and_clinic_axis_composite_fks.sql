-- TASK-445: payments.clinic_id + clinic-axis composite FK hardening
-- BUG-454: pets(clinic_id, owner_id) → owners(clinic_id, id)
--
-- Scope:
--   1. payments.clinic_id ADD + backfill from billings + NOT NULL + FK clinics
--   2. composite FK payments(billing_id, clinic_id) → billings(id, clinic_id)
--   3. composite FK payments(payment_method_id, clinic_id) → payment_methods(id, clinic_id)
--   4. BUG-454: pets(clinic_id, owner_id) → owners(clinic_id, id)
--   5. medical_records UNIQUE(id, clinic_id) + clinic-axis FKs to pets/owners
--   6. vaccinations clinic-axis FKs to pets / medical_records
--   7. billings clinic-axis FKs to pets / owners (nullable components: PG skips FK when any is NULL)
--
-- New constraints use RESTRICT (or default). Do not introduce delete cascade here.

-- =============================================================================
-- 1–3. payments.clinic_id + composite FKs
-- =============================================================================

ALTER TABLE payments
    ADD COLUMN clinic_id bigint;

-- Backfill from the owning billing row (1:1 on billing_id).
UPDATE payments AS payment
SET clinic_id = billing.clinic_id
FROM billings AS billing
WHERE billing.id = payment.billing_id
  AND payment.clinic_id IS NULL;

-- Adding NOT NULL / composite FKs validates all existing rows and fails if any
-- payment lacks a matching billing.clinic_id, or if payment_method_id (when set)
-- belongs to another clinic. Before applying this migration, use the following
-- queries to identify conflicting rows:
--
-- SELECT
--     payment.id,
--     payment.billing_id,
--     payment.clinic_id AS payment_clinic_id,
--     billing.clinic_id AS billing_clinic_id
-- FROM payments AS payment
-- LEFT JOIN billings AS billing
--     ON billing.id = payment.billing_id
-- WHERE payment.clinic_id IS NULL
--    OR billing.id IS NULL
--    OR billing.clinic_id IS DISTINCT FROM payment.clinic_id;
--
-- SELECT
--     payment.id,
--     payment.clinic_id,
--     payment.payment_method_id,
--     payment_method.clinic_id AS payment_method_clinic_id
-- FROM payments AS payment
-- LEFT JOIN payment_methods AS payment_method
--     ON payment_method.id = payment.payment_method_id
-- WHERE payment.payment_method_id IS NOT NULL
--   AND (
--       payment_method.id IS NULL
--       OR payment_method.clinic_id IS DISTINCT FROM payment.clinic_id
--   );

ALTER TABLE payments
    ALTER COLUMN clinic_id SET NOT NULL;

ALTER TABLE payments
    ADD CONSTRAINT fk_payments_clinic_id
    FOREIGN KEY (clinic_id)
    REFERENCES clinics (id)
    ON DELETE RESTRICT;

-- Parent UNIQUE already exists: uq_billings_id_clinic UNIQUE (id, clinic_id)
ALTER TABLE payments
    ADD CONSTRAINT fk_payments_billing_clinic
    FOREIGN KEY (billing_id, clinic_id)
    REFERENCES billings (id, clinic_id)
    ON DELETE RESTRICT;

-- Parent UNIQUE already exists: uq_payment_methods_id_clinic UNIQUE (id, clinic_id)
-- payment_method_id is nullable; PG skips the FK check when any component is NULL.
ALTER TABLE payments
    ADD CONSTRAINT fk_payments_payment_method_clinic
    FOREIGN KEY (payment_method_id, clinic_id)
    REFERENCES payment_methods (id, clinic_id)
    ON DELETE RESTRICT;

CREATE INDEX idx_payments_clinic_id ON payments (clinic_id);

-- =============================================================================
-- 4. BUG-454: pets owner must belong to the same clinic
-- =============================================================================

-- Parent UNIQUE already exists: uq_owners_clinic_id_id UNIQUE (clinic_id, id)
-- Before applying, identify cross-clinic pet.owner_id rows:
--
-- SELECT
--     pet.id,
--     pet.clinic_id,
--     pet.owner_id,
--     owner.clinic_id AS owner_clinic_id
-- FROM pets AS pet
-- LEFT JOIN owners AS owner
--     ON owner.id = pet.owner_id
-- WHERE owner.id IS NULL
--    OR owner.clinic_id IS DISTINCT FROM pet.clinic_id;

ALTER TABLE pets
    ADD CONSTRAINT fk_pets_clinic_owner
    FOREIGN KEY (clinic_id, owner_id)
    REFERENCES owners (clinic_id, id)
    ON DELETE RESTRICT;

-- =============================================================================
-- 5. medical_records: parent UNIQUE + clinic-axis FKs to pets/owners
-- =============================================================================

ALTER TABLE medical_records
    ADD CONSTRAINT uq_medical_records_id_clinic UNIQUE (id, clinic_id);

-- Parent UNIQUEs: uq_pets_clinic_id_id (clinic_id, id), uq_owners_clinic_id_id (clinic_id, id)
-- pet_id / owner_id are nullable; PG skips FK when any component is NULL.
-- Before applying, identify conflicting rows:
--
-- SELECT
--     medical_record.id,
--     medical_record.clinic_id,
--     medical_record.pet_id,
--     pet.clinic_id AS pet_clinic_id
-- FROM medical_records AS medical_record
-- LEFT JOIN pets AS pet
--     ON pet.id = medical_record.pet_id
-- WHERE medical_record.pet_id IS NOT NULL
--   AND (
--       pet.id IS NULL
--       OR pet.clinic_id IS DISTINCT FROM medical_record.clinic_id
--   );
--
-- SELECT
--     medical_record.id,
--     medical_record.clinic_id,
--     medical_record.owner_id,
--     owner.clinic_id AS owner_clinic_id
-- FROM medical_records AS medical_record
-- LEFT JOIN owners AS owner
--     ON owner.id = medical_record.owner_id
-- WHERE medical_record.owner_id IS NOT NULL
--   AND (
--       owner.id IS NULL
--       OR owner.clinic_id IS DISTINCT FROM medical_record.clinic_id
--   );

ALTER TABLE medical_records
    ADD CONSTRAINT fk_medical_records_clinic_pet
    FOREIGN KEY (clinic_id, pet_id)
    REFERENCES pets (clinic_id, id)
    ON DELETE RESTRICT;

ALTER TABLE medical_records
    ADD CONSTRAINT fk_medical_records_clinic_owner
    FOREIGN KEY (clinic_id, owner_id)
    REFERENCES owners (clinic_id, id)
    ON DELETE RESTRICT;

-- =============================================================================
-- 6. vaccinations: clinic-axis FKs to pets / medical_records
-- =============================================================================

-- Parent UNIQUEs: uq_pets_clinic_id_id (clinic_id, id),
--                 uq_medical_records_id_clinic (id, clinic_id) added above.
-- pet_id / medical_record_id are nullable; PG skips FK when any component is NULL.
-- Before applying, identify conflicting rows:
--
-- SELECT
--     vaccination.id,
--     vaccination.clinic_id,
--     vaccination.pet_id,
--     pet.clinic_id AS pet_clinic_id
-- FROM vaccinations AS vaccination
-- LEFT JOIN pets AS pet
--     ON pet.id = vaccination.pet_id
-- WHERE vaccination.pet_id IS NOT NULL
--   AND (
--       pet.id IS NULL
--       OR pet.clinic_id IS DISTINCT FROM vaccination.clinic_id
--   );
--
-- SELECT
--     vaccination.id,
--     vaccination.clinic_id,
--     vaccination.medical_record_id,
--     medical_record.clinic_id AS medical_record_clinic_id
-- FROM vaccinations AS vaccination
-- LEFT JOIN medical_records AS medical_record
--     ON medical_record.id = vaccination.medical_record_id
-- WHERE vaccination.medical_record_id IS NOT NULL
--   AND (
--       medical_record.id IS NULL
--       OR medical_record.clinic_id IS DISTINCT FROM vaccination.clinic_id
--   );

ALTER TABLE vaccinations
    ADD CONSTRAINT fk_vaccinations_clinic_pet
    FOREIGN KEY (clinic_id, pet_id)
    REFERENCES pets (clinic_id, id)
    ON DELETE RESTRICT;

ALTER TABLE vaccinations
    ADD CONSTRAINT fk_vaccinations_medical_record_clinic
    FOREIGN KEY (medical_record_id, clinic_id)
    REFERENCES medical_records (id, clinic_id)
    ON DELETE RESTRICT;

-- =============================================================================
-- 7. billings: clinic-axis FKs to pets / owners
-- =============================================================================

-- Parent UNIQUEs: uq_pets_clinic_id_id (clinic_id, id), uq_owners_clinic_id_id (clinic_id, id)
-- pet_id / owner_id are nullable; PG skips FK when any component is NULL.
-- Before applying, identify conflicting rows:
--
-- SELECT
--     billing.id,
--     billing.clinic_id,
--     billing.pet_id,
--     pet.clinic_id AS pet_clinic_id
-- FROM billings AS billing
-- LEFT JOIN pets AS pet
--     ON pet.id = billing.pet_id
-- WHERE billing.pet_id IS NOT NULL
--   AND (
--       pet.id IS NULL
--       OR pet.clinic_id IS DISTINCT FROM billing.clinic_id
--   );
--
-- SELECT
--     billing.id,
--     billing.clinic_id,
--     billing.owner_id,
--     owner.clinic_id AS owner_clinic_id
-- FROM billings AS billing
-- LEFT JOIN owners AS owner
--     ON owner.id = billing.owner_id
-- WHERE billing.owner_id IS NOT NULL
--   AND (
--       owner.id IS NULL
--       OR owner.clinic_id IS DISTINCT FROM billing.clinic_id
--   );

ALTER TABLE billings
    ADD CONSTRAINT fk_billings_clinic_pet
    FOREIGN KEY (clinic_id, pet_id)
    REFERENCES pets (clinic_id, id)
    ON DELETE RESTRICT;

ALTER TABLE billings
    ADD CONSTRAINT fk_billings_clinic_owner
    FOREIGN KEY (clinic_id, owner_id)
    REFERENCES owners (clinic_id, id)
    ON DELETE RESTRICT;
