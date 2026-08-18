-- BUG-009: /estimates/new?petId= で作成した見積をペットに永続紐付けする。
-- owner_id と同型（nullable FK + ON DELETE SET NULL）。clinic 所有は service で検証する。
ALTER TABLE estimates
  ADD COLUMN IF NOT EXISTS pet_id bigint REFERENCES pets(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_estimates_clinic_pet
  ON estimates (clinic_id, pet_id)
  WHERE pet_id IS NOT NULL AND deleted_at IS NULL;
