-- 010_add_clinical_result_composite_fk.sql
--
-- 由来: 旧 012_add_clinical_result_composite_fk.sql（2026-07-04 に 001_init.sql §7 へ畳み込み済み）。
-- 目的: schema_migrations に薄い/統合前の 001_init.sql のみが記録されている DB へ、
--       畳み込み DDL を additive incremental として適用する upgrade path。
-- 冪等: fresh DB（001 統合スキーマ適用済み）でも IF NOT EXISTS / duplicate_object ガードで通る。
-- 注意: 001_init.sql 本文は変更しない（checksum 維持）。002_checkup_field_clinic_composite_fk.sql
--       （checkup_types↔checkup_type_fields 複合 FK）とは重複しない。
-- スコープ: checkup_field_results↔checkup_type_fields 複合 FK のみ。
-- 002_checkup_field_clinic_composite_fk.sql（checkup_type_fields↔checkup_types）とは別物 — 二重追加しない。

-- 012: 臨床結果テーブルの DB レベル複合 FK（clinic_id 込み）で越境 INSERT/UPDATE を物理拒否する
-- （BE-refactor.md R3-7 / D13・defense-in-depth）。
--
-- 対象は checkup_field_results のみ。exam_results は clinic_id 列を持たず、参照先の exam_type_fields も
-- clinic_id 列を持たない（clinic は exam_type_fields→exam_types→clinics と2段先）ため、(id, clinic_id) の
-- 複合 FK が構造的に張れない。exam_results への同等防御は clinic_id 列の追加 + backfill という
-- 非 additive なスキーマ拡張を要し、behavior-preserving リファクタの範囲外（別タスク）。
--
-- 挙動保存: migration 010 の患者結果値保護（フィールド定義削除時に結果値を残す ON DELETE SET NULL）を
-- 列指定 SET NULL（PostgreSQL 15+ 機能・本番は PG18）で維持する。親 checkup_type_fields を削除すると
-- checkup_type_field_id のみ NULL 化され（MATCH SIMPLE で FK チェックがスキップされる）、NOT NULL の
-- clinic_id と結果値スナップショットは保持される。単一列 SET NULL FK と挙動は完全に一致する。
--
-- 適用前提（手順1・必須）: 既存データに親子 clinic_id 不整合が無いことを検証してから適用すること
-- （違反行があると複合 FK 追加が失敗する）。STG 適用は db_reset 運用ルールに従う。

-- 親テーブルに複合 FK ターゲット用の UNIQUE(id, clinic_id) を追加する。
-- id は PK のため (id, clinic_id) は常に一意で、既存データに対して無条件に充足する（挙動非破壊）。
DO $$
BEGIN
    ALTER TABLE checkup_type_fields
        ADD CONSTRAINT uq_checkup_type_fields_id_clinic UNIQUE (id, clinic_id);
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

-- 既存の単一列 FK（010 の CREATE TABLE インライン FK・自動命名）を複合 FK に置換する。
ALTER TABLE checkup_field_results
    DROP CONSTRAINT IF EXISTS checkup_field_results_checkup_type_field_id_fkey;

DO $$
BEGIN
    ALTER TABLE checkup_field_results
        ADD CONSTRAINT fk_checkup_field_results_field_clinic
        FOREIGN KEY (checkup_type_field_id, clinic_id)
        REFERENCES checkup_type_fields (id, clinic_id)
        ON DELETE SET NULL (checkup_type_field_id);
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;
