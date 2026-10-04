-- EMR-285: 未設定スタッフへ対応可能予約区分（staff_reservation_capabilities）を backfill する。
--
-- 背景: TASK-021 Stage B（staff_reservation_capabilities を肯定形の唯一 write SoT とし、
-- 旧 staff_reservation_exclusions は「excluded = 医院の有効 universe \ capable」の逆写像
-- facade へ転換）で、capability 行を持たないスタッフの既定が
--   旧: 除外未設定 = 全予約区分に対応可能（選択肢に出る）
--   新: capability 未設定 = 全予約区分に対応不可（fail-closed・選択肢から消える）
-- へ暗黙に反転した。切替時の backfill が存在せず、CSV cutover（staffs のみ・
-- junction 表は契約外）で投入された全スタッフが未設定のまま残り、
-- 新規予約作成画面の担当者選択肢が 0 件になった（STG で実測確認済み）。
--
-- 本 migration は facade と同じ計算
--   capable = clinic の is_active かつ未削除の reservation_types \ 旧除外表
-- を「capability 行を 1 件も持たない（= 未設定）スタッフ × 有効な所属医院」に限り
-- 一度だけ適用する。1 行でも capability 行を持つスタッフ（明示設定済み）は対象外。
-- 旧除外表の行は staff 単位（clinic 非スコープ）の旧来設定として引き続き尊重する。
-- 実行は INSERT ... SELECT + ON CONFLICT DO NOTHING で冪等。
-- Do not edit 001_init.sql. User must run: make migrate

INSERT INTO staff_reservation_capabilities (clinic_id, staff_id, reservation_type_id)
SELECT sca.clinic_id, sca.staff_id, rt.id
FROM staff_clinic_assignments sca
JOIN staffs s
  ON s.id = sca.staff_id
 AND s.deleted_at IS NULL
JOIN reservation_types rt
  ON rt.clinic_id = sca.clinic_id
 AND rt.is_active = TRUE
 AND rt.deleted_at IS NULL
WHERE sca.deleted_at IS NULL
  -- 未設定スタッフのみ: 同一医院に 1 行でも capability 行を持つ（明示設定済み）場合は触らない
  AND NOT EXISTS (
      SELECT 1
      FROM staff_reservation_capabilities cap
      WHERE cap.clinic_id = sca.clinic_id
        AND cap.staff_id = sca.staff_id
  )
  -- 旧除外表（staff 単位）の指定は legacy excluded として universe から除く
  AND NOT EXISTS (
      SELECT 1
      FROM staff_reservation_exclusions ex
      WHERE ex.staff_id = sca.staff_id
        AND ex.reservation_type_id = rt.id
  )
ON CONFLICT ON CONSTRAINT uq_staff_reservation_capability DO NOTHING;
