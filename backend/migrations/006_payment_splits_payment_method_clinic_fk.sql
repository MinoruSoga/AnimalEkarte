ALTER TABLE payment_methods
    ADD CONSTRAINT uq_payment_methods_id_clinic UNIQUE (id, clinic_id);

-- Adding the composite foreign key validates all existing rows and fails if a
-- non-NULL payment_method_id belongs to another clinic. Before applying this
-- migration, use the following query to identify conflicting rows:
--
-- SELECT
--     payment_split.id,
--     payment_split.clinic_id,
--     payment_split.payment_method_id,
--     payment_method.clinic_id AS payment_method_clinic_id
-- FROM payment_splits AS payment_split
-- LEFT JOIN payment_methods AS payment_method
--     ON payment_method.id = payment_split.payment_method_id
-- WHERE payment_split.payment_method_id IS NOT NULL
--   AND (
--       payment_method.id IS NULL
--       OR payment_method.clinic_id IS DISTINCT FROM payment_split.clinic_id
--   );

ALTER TABLE payment_splits
    ADD CONSTRAINT fk_payment_splits_payment_method_clinic
    FOREIGN KEY (payment_method_id, clinic_id)
    REFERENCES payment_methods (id, clinic_id)
    ON DELETE RESTRICT;
