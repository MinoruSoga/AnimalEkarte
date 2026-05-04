-- FEAT-383-supplement Scope 3: Lステップ自動配信バッチ実行時刻をクリニックごとに設定可能にする。
-- clinic_settings テーブルに lstep_fire_hour_jst カラムを追加する（Case A: ALTER TABLE）。
ALTER TABLE clinic_settings
    ADD COLUMN lstep_fire_hour_jst INT NOT NULL DEFAULT 10
        CHECK (lstep_fire_hour_jst BETWEEN 0 AND 23);

COMMENT ON COLUMN clinic_settings.lstep_fire_hour_jst IS 'Lステップ自動配信バッチを実行する時刻（JST、0–23）。デフォルト 10 時。';
