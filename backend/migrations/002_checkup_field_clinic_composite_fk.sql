-- 002_checkup_field_clinic_composite_fk.sql
--
-- 背景（重要 — 起草前に必ず読むこと）:
--
-- Issue #211 の 2026-07-07 コメントで「exam_results の DBレベル複合FK → 構造的不可・着手禁止」と
-- 決着している。これは exam_results / exam_type_fields に clinic_id 列自体が存在しない
-- （親 exams / exam_types 経由の間接スコープのみ）ためで、今も変わっていない。
-- 本ファイルの対象はそれとは別の checkup ドメイン（health checkup / 健診パッケージ機能）であり、
-- exam_results を指すものではない。
--
-- 【スコープの訂正 — 起草時に判明した事実】
-- 本ファイルの起草依頼は「checkup_type_fields↔checkup_types」「checkup_field_results↔checkup_type_fields」
-- の2本の複合FKを新設する前提だったが、001_init.sql を実査した結果、後者
-- （checkup_field_results.checkup_type_field_id ↔ checkup_type_fields.clinic_id 複合FK）は
-- 既に 001_init.sql L3678-3691（コメント上は「012_add_clinical_result_composite_fk.sql」相当の
-- 内容が統合スキーマへ折り込み済み）で実装済みであることが判明した:
--   - uq_checkup_type_fields_id_clinic UNIQUE (id, clinic_id)  ← 本ファイルの要件#2相当
--   - fk_checkup_field_results_field_clinic
--       FOREIGN KEY (checkup_type_field_id, clinic_id)
--       REFERENCES checkup_type_fields (id, clinic_id)
--       ON DELETE SET NULL (checkup_type_field_id)           ← 本ファイルの要件#4相当（列指定SET NULLも同一）
-- これを本ファイルで再度追加すると、同一カラム対に対する重複FK制約が2本並存することになり
-- （DROP CONSTRAINT IF EXISTS は旧単一列FKが既に置換済みのため no-op になり、その後の ADD が
-- 冗長な2本目の複合FKとして積み上がる）、意味的重複でしかないため追加しない。
-- 依頼元メモの「checkup_field_results.clinic_id↔checkup_type_fields.clinic_id DBレベル複合FK無
-- （app層+lint+runtime testでカバー済・迂回writeのみ残risk）」という記述は上記の理由により
-- 陳腐化している（実際には既にDBレベル複合FKで保護済み）。この drift はタスク完了報告で
-- 呼び出し元に明示すること。
--
-- 【本ファイルが実際に追加するもの】
-- 実査の結果、checkup_type_fields.checkup_type_id → checkup_types.id の複合FK
-- （clinic_id 込みで越境 INSERT/UPDATE を物理拒否するもの）は未実装だった
-- （001_init.sql L3567 は依然として単一列 REFERENCES checkup_types(id)（ON DELETE は CASCADE 指定）のまま、
-- checkup_types 側にも (id, clinic_id) 相当の UNIQUE が存在しない）。本ファイルはこの1本のみを追加する。
--
-- 列順の規約: 001_init.sql L3678-3691 の既存パターンに合わせ、UNIQUE / FK とも
-- (id, clinic_id) の順で統一する（依頼メモは (clinic_id, id) 順を例示していたが、
-- 同一ファイル内で UNIQUE と FK の列順を揃えれば PostgreSQL の複合FK照合上どちらの順でも
-- 機能的に等価であり、既存実装との対称性を優先した）。
--
-- 挙動保存: checkup_type_id は NOT NULL のため SET NULL は不要。既存の CASCADE 指定（ON DELETE）
-- （親 checkup_types が削除されると checkup_type_fields も連動削除される・純粋従属データ）を
-- 複合FKでもそのまま維持する。
--
-- 冪等性/適用時の注意: DROP CONSTRAINT IF EXISTS で参照する制約名
-- （checkup_type_fields_checkup_type_id_fkey）は 001_init.sql のインライン REFERENCES 由来の
-- PostgreSQL デフォルト命名規則からの推定であり、実DBの実際の制約名がこれと異なる場合がある。
-- 適用前にユーザーが `\d checkup_type_fields` 等で実際の制約名を確認・必要なら本ファイルを
-- 調整すること。本ファイルは起草のみであり、適用（migrate 実行）は行わない。
--
-- 適用前提: 既存データに checkup_type_fields.clinic_id と checkup_types.clinic_id の不整合が
-- 無いことを検証してから適用すること（違反行があると複合FK追加が失敗する）。

-- 親テーブルに複合FKターゲット用の UNIQUE(id, clinic_id) を追加する。
-- id は PK のため (id, clinic_id) は常に一意で、既存データに対して無条件に充足する（挙動非破壊）。
ALTER TABLE checkup_types
    ADD CONSTRAINT uq_checkup_types_id_clinic UNIQUE (id, clinic_id);

-- 既存の単一列FK（001_init.sql の CREATE TABLE インライン FK・自動命名）を複合FKに置換する。
ALTER TABLE checkup_type_fields
    DROP CONSTRAINT IF EXISTS checkup_type_fields_checkup_type_id_fkey;

ALTER TABLE checkup_type_fields
    ADD CONSTRAINT fk_checkup_type_fields_type_clinic
    FOREIGN KEY (checkup_type_id, clinic_id)
    REFERENCES checkup_types (id, clinic_id)
    ON DELETE CASCADE;

-- インデックスについて: 親 checkup_types 削除時の CASCADE 検索（ソフトデリート済み行も含めて
-- 子行を捜索する必要がある）は、001_init.sql の非パーシャル index
-- idx_checkup_type_fields_checkup_type_id (checkup_type_id) で既にカバーされている
-- （WHERE deleted_at IS NULL の部分index である idx_checkup_type_fields_clinic_type_sort は
-- プランナが述語を保証できないため CASCADE 検索には使われない）。よって新規indexは追加しない。
