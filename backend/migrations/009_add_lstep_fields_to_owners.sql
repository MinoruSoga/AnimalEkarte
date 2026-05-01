-- FEAT-381 Phase 1: add LINE/delivery/transfer fields to owners

ALTER TABLE owners ADD COLUMN IF NOT EXISTS line_id_confirmed_at TIMESTAMPTZ;
ALTER TABLE owners ADD COLUMN IF NOT EXISTS delivery_excluded BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE owners ADD COLUMN IF NOT EXISTS delivery_excluded_reason VARCHAR(100);
ALTER TABLE owners ADD COLUMN IF NOT EXISTS is_transferred BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE owners ADD COLUMN IF NOT EXISTS transfer_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_owners_delivery_excluded
    ON owners (clinic_id, delivery_excluded)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_owners_is_transferred
    ON owners (clinic_id, is_transferred)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_owners_line_id_confirmed
    ON owners (clinic_id, line_id_confirmed_at)
    WHERE deleted_at IS NULL;
