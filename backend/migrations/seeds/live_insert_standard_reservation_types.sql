-- One-shot standard reservation-type master insert for every clinic.
-- EMR-193 (スプシNo.50 八王子テスト報告): 新規予約作成で予約区分が
-- トリミングしか選べなかったため、標準区分「診察・お手入れ・ワクチン・
-- 健診」を全医院へ投入する。
--
-- 位置づけ:
--   seeds/002_master の CSV bundle は適用後 immutable で、内容変更の正規
--   経路は使い捨て DB への実適用からの cmd/seed-export 再生成のみ
--   （docs/ops/deploy/SEED_MIGRATION_OPERATIONS.md §3 / migrations/CLAUDE.md）。
--   本スクリプトはその正規経路を恒久的に置き換えるものではなく、適用済み
--   DB（STG / local / 本番）へ同一マスタを投入する明示的な手動手順。
--   cmd/migrate の入力ではない（直下 *.sql の DDL フェーズは seed bundle
--   より先に走るため clinics が無い fresh DB では必ず失敗する）。
--   承認済み runbook から psql ON_ERROR_STOP=1 で適用すること
--   （live_insert_lab_device_clinic2.sql と同型）。
--
-- 値の根拠（002_master/reservation_types.csv のトリミング行に揃えた項目）:
--   - is_active=true / duration_minutes=15 / reservation_day_option='none' /
--     is_internal=false / show_short_name=false / short_name=name と同じ。
--   - category='general'（enum は general|trimming の2値）。
--   - reservation_visible=false: 院内予約フォームは is_active のみで絞る
--     （useGetReservationTypesGrouped）ため院内では即選択可能だが、
--     LIFF 飼い主予約へは公開しない保守既定。LINE 公開は医院が
--     設定画面で区分ごとに有効化する（PO 判断を委ねる）。
--   - color: グループ未所属の区分はカレンダーで自身の color が使われる
--     （useReservationTypeColorMap）。識別用に区分ごとに異なる色を付与。
--   - sort_order=1-4: 既存トリミング(9)より前に標準区分を並べる。
--
-- 冪等性・fail-closed:
--   (clinic_id, name) の live 行（deleted_at IS NULL）が既に存在する医院
--   では INSERT をスキップする。idx_reservation_types_clinic_name は
--   deleted_at IS NULL 限定の一意のため既存行との衝突は起きない。
--   既存行（手動作成の「診察」等）は属性を一切上書きしない。
--   事後条件として全医院に 4 区分の live 行が揃うことを検証し、
--   未達なら例外で transaction 全体を rollback する。
--
-- 適用後に fresh DB へも同内容を含めたい場合は、別途 CSV bundle の
-- 正規再生成（cmd/seed-export）をユーザーが行う。本スクリプトの適用だけ
-- では seeds/002_master/reservation_types.csv は変わらない。

BEGIN;
SELECT pg_advisory_xact_lock(333050);

CREATE TEMP TABLE desired_reservation_types (
  name text PRIMARY KEY,
  short_name text NOT NULL,
  duration_minutes integer NOT NULL,
  color text NOT NULL,
  sort_order integer NOT NULL
) ON COMMIT DROP;
INSERT INTO desired_reservation_types (name, short_name, duration_minutes, color, sort_order) VALUES
  ('診察', '診察', 15, '#3B82F6', 1),
  ('お手入れ', 'お手入れ', 15, '#10B981', 2),
  ('ワクチン', 'ワクチン', 15, '#8B5CF6', 3),
  ('健診', '健診', 15, '#F97316', 4);

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM clinics) THEN
    RAISE EXCEPTION 'no clinics found; apply seed bundle 002_master first';
  END IF;
END $$;

INSERT INTO reservation_types
  (clinic_id, name, is_active, description, color, sort_order,
   duration_minutes, short_name, show_short_name, reservation_visible,
   reservation_day_option, is_internal, category)
SELECT c.id, d.name, true, '', d.color, d.sort_order,
       d.duration_minutes, d.short_name, false, false,
       'none', false, 'general'
FROM clinics c
CROSS JOIN desired_reservation_types d
WHERE NOT EXISTS (
  SELECT 1 FROM reservation_types e
  WHERE e.clinic_id = c.id AND e.name = d.name AND e.deleted_at IS NULL
);

DO $$
DECLARE expected bigint; actual bigint;
BEGIN
  SELECT count(*) INTO expected
  FROM clinics c CROSS JOIN desired_reservation_types d;
  SELECT count(*) INTO actual
  FROM clinics c CROSS JOIN desired_reservation_types d
  WHERE EXISTS (
    SELECT 1 FROM reservation_types e
    WHERE e.clinic_id = c.id AND e.name = d.name AND e.deleted_at IS NULL
  );
  IF actual <> expected THEN
    RAISE EXCEPTION 'standard reservation types postcondition mismatch: expected %, got %', expected, actual;
  END IF;
END $$;

COMMIT;
