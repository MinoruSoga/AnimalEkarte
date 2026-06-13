BEGIN;

ALTER TABLE billing_refunds
    DROP CONSTRAINT IF EXISTS billing_refunds_billing_id_fkey;

ALTER TABLE billing_refunds
    ADD CONSTRAINT billing_refunds_billing_id_fkey
    FOREIGN KEY (billing_id)
    REFERENCES billings(id)
    ON DELETE RESTRICT;

COMMIT;
