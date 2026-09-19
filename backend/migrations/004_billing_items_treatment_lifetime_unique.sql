-- Unique provenance for billed treatments, matching vaccination/exam lifetime unique.
-- Partial unique on treatment_id (including soft-deleted rows) so a billed treatment
-- cannot be claimed again on another billing.
CREATE UNIQUE INDEX uq_billing_items_treatment_lifetime
    ON billing_items (treatment_id)
    WHERE treatment_id IS NOT NULL;

COMMENT ON INDEX uq_billing_items_treatment_lifetime IS
    '同一 treatment_id の会計明細を生涯1件に制限する（未請求治療の二重確定防止）';
