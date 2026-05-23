-- 006: add payment_splits table for mixed payment method support
CREATE TABLE IF NOT EXISTS payment_splits (
    id                BIGSERIAL PRIMARY KEY,
    clinic_id         BIGINT NOT NULL,
    billing_id        BIGINT NOT NULL REFERENCES billings(id) ON DELETE RESTRICT,
    method            payment_method NOT NULL,
    payment_method_id BIGINT REFERENCES payment_methods(id),
    amount            BIGINT NOT NULL DEFAULT 0,
    received_amount   BIGINT NOT NULL DEFAULT 0,
    change_amount     BIGINT NOT NULL DEFAULT 0,
    paid_by           BIGINT REFERENCES staffs(id),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payment_splits_clinic_billing ON payment_splits(clinic_id, billing_id);

-- Backfill: existing paid billings get 1 split row from payments
INSERT INTO payment_splits (clinic_id, billing_id, method, payment_method_id, amount, received_amount, change_amount, paid_by)
SELECT
    b.clinic_id,
    p.billing_id,
    p.method,
    p.payment_method_id,
    p.billing_amount,
    p.received_amount,
    p.change_amount,
    p.paid_by
FROM payments p
JOIN billings b ON b.id = p.billing_id
WHERE p.billing_amount > 0
ON CONFLICT DO NOTHING;
