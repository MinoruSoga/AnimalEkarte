-- EMR-174: pets に「名前の由来」と「出逢いのストーリー」の任意記録カラムを追加する。
-- blood_type / microchip_number / deceased_reason（001_init.sql pets 定義）と同型の
-- text NULL で、NULL=未記録・空文字列は登録しない（API 層で空白を NULL へ正規化）。
-- 業務参照・検索の対象ではないためインデックスは張らない。
-- Do not edit 001_init.sql. User must run: make migrate

ALTER TABLE pets
    ADD COLUMN name_origin   text NULL,   -- 名前の由来。NULL=未記録。
    ADD COLUMN meeting_story text NULL;   -- 出逢いのストーリー。NULL=未記録。

COMMENT ON COLUMN pets.name_origin   IS '名前の由来。NULL=未記録。';
COMMENT ON COLUMN pets.meeting_story IS '出逢いのストーリー。NULL=未記録。';
