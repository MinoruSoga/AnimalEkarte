-- CPM V2 来院回数閾値 (clinic 単位調整可能、デフォルト: 2/4/8/13)
-- CalculateCPMStageV2 のハードコードを clinic_settings 経由に置き換え

ALTER TABLE clinic_settings
    ADD COLUMN IF NOT EXISTS cpm_v2_coming_threshold  INT NOT NULL DEFAULT 2
        CHECK (cpm_v2_coming_threshold >= 1),
    ADD COLUMN IF NOT EXISTS cpm_v2_good_threshold    INT NOT NULL DEFAULT 4
        CHECK (cpm_v2_good_threshold >= 1),
    ADD COLUMN IF NOT EXISTS cpm_v2_family_threshold  INT NOT NULL DEFAULT 8
        CHECK (cpm_v2_family_threshold >= 1),
    ADD COLUMN IF NOT EXISTS cpm_v2_noah_threshold    INT NOT NULL DEFAULT 13
        CHECK (cpm_v2_noah_threshold >= 1);

COMMENT ON COLUMN clinic_settings.cpm_v2_coming_threshold  IS 'CPM V2 これから ステージ開始来院回数 (デフォルト 2)';
COMMENT ON COLUMN clinic_settings.cpm_v2_good_threshold    IS 'CPM V2 いいかんじ ステージ開始来院回数 (デフォルト 4)';
COMMENT ON COLUMN clinic_settings.cpm_v2_family_threshold  IS 'CPM V2 ファミリー ステージ開始来院回数 (デフォルト 8)';
COMMENT ON COLUMN clinic_settings.cpm_v2_noah_threshold    IS 'CPM V2 ノア ステージ開始来院回数 (デフォルト 13)';
