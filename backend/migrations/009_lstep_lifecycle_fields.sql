-- LSTEP-BE-017 (DB part): ペット死亡記録・飼い主オプトアウトフィールド追加
-- 動物医療系EMRとして「死亡ペットへのリマインダー配信」を防ぐ絶対的品質要件。

-- ペット死亡記録
ALTER TABLE pets
    ADD COLUMN deceased_at     TIMESTAMPTZ NULL,
    ADD COLUMN deceased_reason TEXT        NULL;

CREATE INDEX idx_pets_deceased ON pets (clinic_id, deceased_at)
    WHERE deceased_at IS NOT NULL;

-- 飼い主オプトアウト
ALTER TABLE owners
    ADD COLUMN lstep_opt_out        BOOLEAN     NOT NULL DEFAULT false,
    ADD COLUMN lstep_opt_out_at     TIMESTAMPTZ NULL,
    ADD COLUMN lstep_opt_out_reason TEXT        NULL;

COMMENT ON COLUMN pets.deceased_at IS 'ペット死亡日。NULL = 生存中。';
COMMENT ON COLUMN pets.deceased_reason IS 'ペット死亡理由（任意記録）。';
COMMENT ON COLUMN owners.lstep_opt_out IS 'Lステップ配信オプトアウトフラグ。true = すべてのタグ付与をスキップ。';
COMMENT ON COLUMN owners.lstep_opt_out_at IS 'オプトアウト設定日時。';
COMMENT ON COLUMN owners.lstep_opt_out_reason IS 'オプトアウト理由（監査ログ用）。';
