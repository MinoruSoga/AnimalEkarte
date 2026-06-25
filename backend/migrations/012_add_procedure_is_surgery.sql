-- #159: procedures.is_surgery を追加。手術処置を is_surgery フラグで識別する。
-- additive: DEFAULT false で後方互換。既存処置レコードは is_surgery = false のままになる。
ALTER TABLE procedures ADD COLUMN IF NOT EXISTS is_surgery BOOLEAN NOT NULL DEFAULT false;

CREATE INDEX IF NOT EXISTS idx_procedures_clinic_is_surgery
    ON procedures(clinic_id, is_surgery)
    WHERE is_surgery = true;
