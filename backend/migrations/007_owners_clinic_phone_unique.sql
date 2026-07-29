-- POC-06 / U-X05-OWNER-PHONE: feed-owner phone uniqueness at the DB boundary.
--
-- Mirrors uk_owners_clinic_email (partial unique on clinic_id + contact field)
-- so concurrent Create/Update cannot insert two active owners with the same
-- non-empty phone inside one clinic. Application ensureOwnerPhoneUnique remains
-- for friendly messages; this index is the fail-closed source of truth.
--
-- Empty phone ('') is excluded so legacy rows without phone can coexist.
-- Soft-deleted rows are excluded (deleted_at IS NULL).

CREATE UNIQUE INDEX IF NOT EXISTS uk_owners_clinic_phone
    ON owners (clinic_id, phone)
    WHERE deleted_at IS NULL AND phone <> '';
