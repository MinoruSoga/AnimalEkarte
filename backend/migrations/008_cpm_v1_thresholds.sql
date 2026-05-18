-- CPM V1 判定閾値 (clinic 単位調整可能)
-- CalculateCPMStage のハードコードを clinic_settings 経由に置き換え

ALTER TABLE clinic_settings
    ADD COLUMN IF NOT EXISTS cpm_v1_dormant_days             INT     NOT NULL DEFAULT 240
        CHECK (cpm_v1_dormant_days >= 1),
    ADD COLUMN IF NOT EXISTS cpm_v1_noah_days                INT     NOT NULL DEFAULT 365
        CHECK (cpm_v1_noah_days >= 1),
    ADD COLUMN IF NOT EXISTS cpm_v1_noah_annual_visits       INT     NOT NULL DEFAULT 3
        CHECK (cpm_v1_noah_annual_visits >= 1),
    ADD COLUMN IF NOT EXISTS cpm_v1_noah_ltv                 BIGINT  NOT NULL DEFAULT 80000
        CHECK (cpm_v1_noah_ltv >= 0),
    ADD COLUMN IF NOT EXISTS cpm_v1_core_days                INT     NOT NULL DEFAULT 180
        CHECK (cpm_v1_core_days >= 1),
    ADD COLUMN IF NOT EXISTS cpm_v1_core_annual_visits       INT     NOT NULL DEFAULT 2
        CHECK (cpm_v1_core_annual_visits >= 1),
    ADD COLUMN IF NOT EXISTS cpm_v1_core_ltv                 BIGINT  NOT NULL DEFAULT 50000
        CHECK (cpm_v1_core_ltv >= 0),
    ADD COLUMN IF NOT EXISTS cpm_v1_spot_min_amount          BIGINT  NOT NULL DEFAULT 30000
        CHECK (cpm_v1_spot_min_amount >= 0),
    ADD COLUMN IF NOT EXISTS cpm_v1_spot_inactive_days       INT     NOT NULL DEFAULT 90
        CHECK (cpm_v1_spot_inactive_days >= 1),
    ADD COLUMN IF NOT EXISTS cpm_v1_growing_max_days         INT     NOT NULL DEFAULT 90
        CHECK (cpm_v1_growing_max_days >= 1),
    ADD COLUMN IF NOT EXISTS cpm_v1_growing_min_visits       INT     NOT NULL DEFAULT 2
        CHECK (cpm_v1_growing_min_visits >= 1),
    ADD COLUMN IF NOT EXISTS cpm_v1_growing_max_visits       INT     NOT NULL DEFAULT 3
        CHECK (cpm_v1_growing_max_visits >= 1),
    ADD COLUMN IF NOT EXISTS cpm_v1_ltv_break_low            BIGINT  NOT NULL DEFAULT 20000
        CHECK (cpm_v1_ltv_break_low >= 0);

COMMENT ON COLUMN clinic_settings.cpm_v1_dormant_days          IS 'CPM V1 cpm_dormant: 最終来院からの経過日数 >= この値で dormant 判定 (デフォルト 240)';
COMMENT ON COLUMN clinic_settings.cpm_v1_noah_days             IS 'CPM V1 cpm_noah: 初来院からの経過日数 >= この値 (デフォルト 365)';
COMMENT ON COLUMN clinic_settings.cpm_v1_noah_annual_visits    IS 'CPM V1 cpm_noah: 年間来院回数 >= この値 (デフォルト 3)';
COMMENT ON COLUMN clinic_settings.cpm_v1_noah_ltv              IS 'CPM V1 cpm_noah: 累計金額 >= この値 (デフォルト 80000)';
COMMENT ON COLUMN clinic_settings.cpm_v1_core_days             IS 'CPM V1 cpm_core: 初来院からの経過日数 >= この値 (デフォルト 180)';
COMMENT ON COLUMN clinic_settings.cpm_v1_core_annual_visits    IS 'CPM V1 cpm_core: 年間来院回数 >= この値 (デフォルト 2)';
COMMENT ON COLUMN clinic_settings.cpm_v1_core_ltv              IS 'CPM V1 cpm_core: 累計金額 >= この値; growing 上限にも兼用 (デフォルト 50000)';
COMMENT ON COLUMN clinic_settings.cpm_v1_spot_min_amount       IS 'CPM V1 cpm_spot: 単回最大金額 >= この値 (デフォルト 30000)';
COMMENT ON COLUMN clinic_settings.cpm_v1_spot_inactive_days    IS 'CPM V1 cpm_spot: 最終来院からの経過日数 > この値 (デフォルト 90)';
COMMENT ON COLUMN clinic_settings.cpm_v1_growing_max_days      IS 'CPM V1 cpm_growing: 初来院からの経過日数 <= この値 (デフォルト 90)';
COMMENT ON COLUMN clinic_settings.cpm_v1_growing_min_visits    IS 'CPM V1 cpm_growing: 総来院回数 >= この値 (デフォルト 2)';
COMMENT ON COLUMN clinic_settings.cpm_v1_growing_max_visits    IS 'CPM V1 cpm_growing: 総来院回数 <= この値 (デフォルト 3)';
COMMENT ON COLUMN clinic_settings.cpm_v1_ltv_break_low         IS 'CPM V1 growing 下限 / encounter 上限境界 (デフォルト 20000)';
