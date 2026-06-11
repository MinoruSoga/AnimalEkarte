-- 007_add_payment_splits_fk.sql
-- payment_splits.clinic_id の外部キー制約追加

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint c
    JOIN pg_class r ON r.oid = c.conrelid
    WHERE c.contype = 'f'
      AND c.conname = 'fk_payment_splits_clinic_id'
      AND r.relname = 'payment_splits'
  ) THEN
    ALTER TABLE payment_splits
      ADD CONSTRAINT fk_payment_splits_clinic_id
      FOREIGN KEY (clinic_id)
      REFERENCES clinics (id)
      ON DELETE RESTRICT;
  END IF;
END
$$;
