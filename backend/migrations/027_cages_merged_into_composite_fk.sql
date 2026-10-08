-- EMR-306 レビュー指摘対応（Codex P1）: merged_into_cage_id の自己参照FKを
-- hospitalizations.cage_id と同じ医院スコープ込みの複合参照に揃える。
-- 025 の単一列FK (merged_into_cage_id) REFERENCES cages(id) では、import/修復等の
-- 非API経路で他医院のケージを指す行を DB 層で拒否できず、ON DELETE RESTRICT が
-- 医院越境の削除依存を生み得る。複合参照先の uq_cages_id_clinic UNIQUE (id, clinic_id)
-- は 001_init.sql で既存のため追加の UNIQUE 制約は不要。
-- API 書込み経路（Merge）は常に同一 clinic 内で整合するため既存データとの互換性はある。

ALTER TABLE cages
    DROP CONSTRAINT fk_cages_merged_into_cage;

ALTER TABLE cages
    ADD CONSTRAINT fk_cages_merged_into_cage
        FOREIGN KEY (merged_into_cage_id, clinic_id)
        REFERENCES cages (id, clinic_id)
        ON DELETE RESTRICT;
