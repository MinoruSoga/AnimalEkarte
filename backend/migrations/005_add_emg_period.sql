-- 005_add_emg_period.sql
-- Issue #150: レジ締めに EMG（緊急）区分を追加する。
--
-- cash_register_closes.period を varchar(2) → varchar(3) に拡張し、
-- CHECK 制約に 'emg' を許可する。
-- 既存の 'am' / 'pm' レコードは列拡張・制約緩和のみのため影響を受けない。
--
-- 注意: 001_init.sql のインライン CHECK は自動命名 cash_register_closes_period_check となる。
-- 名前が一致しない環境でも安全に進むよう DROP CONSTRAINT IF EXISTS を用いる。

ALTER TABLE cash_register_closes
    ALTER COLUMN period TYPE varchar(3);

ALTER TABLE cash_register_closes
    DROP CONSTRAINT IF EXISTS cash_register_closes_period_check;

ALTER TABLE cash_register_closes
    ADD CONSTRAINT cash_register_closes_period_check
    CHECK (period IN ('am', 'pm', 'emg'));
