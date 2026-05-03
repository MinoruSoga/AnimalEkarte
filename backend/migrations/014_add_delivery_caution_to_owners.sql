-- FEAT-381-2 Commit 1: delivery_caution flag (Lstep phase 2)
-- 配信注意フラグ: delivery_excluded（配信停止）より軽度の注意レベル

ALTER TABLE owners
    ADD COLUMN IF NOT EXISTS delivery_caution boolean NOT NULL DEFAULT false;

ALTER TABLE owners
    ADD COLUMN IF NOT EXISTS delivery_caution_reason varchar(100);
