-- EMR-306 ケージ一時連結 Phase 1（C案: 代表ケージ参照）。
-- 連結中の構成ケージ（メンバー）が代表ケージを指す nullable な自己参照FKを追加する。
-- NULL = 連結していないケージ（独立ケージ、および代表ケージ自身）。
-- 1ケージが持つ状態は independent / representative / member のいずれか1つで、
-- ネスト（member の member）や循環はアプリケーション層のバリデーションで禁止する
-- （この列は状態の表現のみを担う）。enum 追加は不要。
-- 代表ケージ行の物理削除は構成員の参照が残る限り ON DELETE RESTRICT で拒否する
-- （ケージ削除は deleted_at による soft delete が正規経路。連結解除は明示操作のみで、
--  削除による暗黙の連結解除＝構成員行の書換えを FK レベルで防ぐ）。
-- Do not edit 001_init.sql. User must run: make migrate
-- STG 未適用の 025〜027 をユーザー承認により統合（2026-10-09）。
-- 旧 025〜027 適用済みのローカル DB は再作成が必要（README.md 参照）。
-- 医院越境の参照を DB でも拒否し、検索インデックスは clinic_id を先頭にする。

ALTER TABLE cages
    ADD COLUMN merged_into_cage_id bigint;

ALTER TABLE cages
    ADD CONSTRAINT fk_cages_merged_into_cage
        FOREIGN KEY (merged_into_cage_id, clinic_id)
        REFERENCES cages (id, clinic_id)
        ON DELETE RESTRICT;

COMMENT ON COLUMN cages.merged_into_cage_id IS
    'EMR-306: 連結中の構成ケージが参照する代表ケージID。NULL = 独立ケージまたは代表ケージ自身。';

CREATE INDEX idx_cages_merged_into_cage_id
    ON cages (clinic_id, merged_into_cage_id)
    WHERE merged_into_cage_id IS NOT NULL AND deleted_at IS NULL;
