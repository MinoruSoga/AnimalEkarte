-- Supports payment-graph verification and billing-scoped reporting without
-- scanning every clinic's payment splits. The existing
-- (clinic_id, billing_id) index remains the tenant-scoped access path.
CREATE INDEX IF NOT EXISTS idx_payment_splits_billing_id
    ON payment_splits (billing_id);
