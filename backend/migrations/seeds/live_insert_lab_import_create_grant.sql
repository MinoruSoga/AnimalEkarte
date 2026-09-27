-- One-shot lab-import:create grant for default-named permission groups at
-- existing clinics. EMR-176: 既定値変更前に作成されたクリニックでは
-- 「執行」「一般」グループが旧既定 (lab-import view のみ) のままで、
-- 検査受信が誰もできない状態が残る。本スクリプトは新既定を既存の
-- 既定名グループへ反映する手動手順。
--
-- 位置づけ:
--   seeds/002_master の CSV bundle は適用後 immutable（本変更で lab-import
--   行の can_create を f→t 更新済み。fresh DB は seed が新既定を適用する）。
--   seed bundle 所有行 (group_id 1–8) への適用済み DB 反映は migration
--   010_grant_lab_import_create_to_default_groups.sql が担う。
--   本スクリプトはその対象外、すなわち CreateClinic / 権限グループ設定 UI で
--   後から作られた既定名グループ (「執行」「一般」、id > 9) を対象とする
--   補完手順であり、cmd/migrate の入力ではない。
--   承認済み runbook から psql ON_ERROR_STOP=1 で適用すること
--   （live_insert_standard_reservation_types.sql と同型）。
--
-- 対象と値の根拠:
--   - 対象: permission_groups.name が '執行' または '一般' の live 行が持つ
--     lab-import ルール。新規クリニックの CreateClinic が生成する既定名と
--     一致させ、個別リネーム済みグループは（既定意図が不明なため）触れない。
--   - 値: can_create='t' のみ立てる。view 権限を持たない行には付与しない
--     （create-only の fail-closed 違反を作らない）。can_edit（受信済み結果の
--     カルテ紐付け/解除/切戻し）は従来値を維持する — 新既定でも一般に
--     edit を付与しないため、edit を立てないことが新既定と一致する。
--   - 閲覧専用・その他名称のグループは対象外。
--
-- 冪等性・fail-closed:
--   can_view='t' かつ can_create='f' の行だけを更新するため再実行は no-op。
--   事後条件として、対象グループに「view あり・create なし」の lab-import 行が
--   残っていないことを検証し、未達なら例外で transaction 全体を rollback する。
--   lab-import ルール行そのものが存在しない既定名グループ（古い経路で作られ
--   た等）には権限行を新規作成しない — 手動付与は権限グループ設定 UI で行い、
--   その旨を NOTICE で報告する。

BEGIN;
SELECT pg_advisory_xact_lock(333051);

CREATE TEMP TABLE desired_lab_import_groups (
  name text PRIMARY KEY
) ON COMMIT DROP;
INSERT INTO desired_lab_import_groups (name) VALUES ('執行'), ('一般');

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM clinics) THEN
    RAISE EXCEPTION 'no clinics found; apply seed bundle 002_master first';
  END IF;
END $$;

DO $$
DECLARE missing_count bigint;
BEGIN
  SELECT count(*) INTO missing_count
  FROM permission_groups g
  JOIN desired_lab_import_groups d ON d.name = g.name
  WHERE g.deleted_at IS NULL
    AND NOT EXISTS (
      SELECT 1 FROM permission_group_rules r
      WHERE r.group_id = g.id
        AND r.resource = 'lab-import'
        AND r.deleted_at IS NULL
    );
  IF missing_count > 0 THEN
    RAISE NOTICE '% default-named group(s) have no lab-import rule row; grant manually via permission-groups UI', missing_count;
  END IF;
END $$;

UPDATE permission_group_rules r
SET can_create = 't'
FROM permission_groups g
JOIN desired_lab_import_groups d ON d.name = g.name
WHERE r.group_id = g.id
  AND g.deleted_at IS NULL
  AND r.resource = 'lab-import'
  AND r.deleted_at IS NULL
  AND r.can_view = 't'
  AND r.can_create = 'f';

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM permission_groups g
    JOIN desired_lab_import_groups d ON d.name = g.name
    JOIN permission_group_rules r ON r.group_id = g.id
    WHERE g.deleted_at IS NULL
      AND r.resource = 'lab-import'
      AND r.deleted_at IS NULL
      AND r.can_view = 't'
      AND r.can_create = 'f'
  ) THEN
    RAISE EXCEPTION 'lab-import:create grant postcondition mismatch';
  END IF;
END $$;

COMMIT;
