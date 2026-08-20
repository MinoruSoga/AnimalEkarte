-- ============================================================
-- 稼働DB向け: 城東(clinic_id=2) 検査機器マスタ投入
-- 実行方法:
--   docker compose exec -T db psql -U "$DB_USER" -d "$DB_NAME" \
--     < backend/migrations/seeds/live_insert_lab_device_clinic2.sql
--
-- 冪等: ON CONFLICT DO NOTHING のため再実行可。clinic_id=1 は触れない。
-- ============================================================

BEGIN;

-- ── lab_devices ────────────────────────────────────────────
-- UNIQUE (clinic_id, source_type) に衝突したらスキップ
INSERT INTO lab_devices
  (clinic_id, source_type, name, exam_type_id, is_active, sort_order, created_at, updated_at)
VALUES
  (2, 'fuji_nx600',    'NX600',          7,  true, 10, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  (2, 'fuji_au10v',    'AU10V',          7,  true, 20, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  (2, 'arkray_pu4010', '尿（PU-4010）',  13, true, 30, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  (2, 'idexx_vetlab',  'IDEXX VetLab',   6,  true, 40, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09')
ON CONFLICT (clinic_id, source_type) DO NOTHING;

-- ── lab_device_item_masters ────────────────────────────────
-- UNIQUE (clinic_id, source_type, device_item_code) に衝突したらスキップ
-- EnsureDefaults 済みの環境では既存行を保護（exam_type_field_id が NULL の行は UPDATE で埋める）
INSERT INTO lab_device_item_masters
  (clinic_id, source_type, device_item_code, unit, value_shape, exam_type_field_id, sort_order, is_active, created_at, updated_at)
VALUES
  -- fuji_nx600 (血液化学検査 exam_type_id=7)
  (2, 'fuji_nx600', 'Na-P',   'mEq/l',  'numeric',      53, 10,  true, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  (2, 'fuji_nx600', 'K-P',    'mEq/l',  'numeric',      54, 20,  true, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  (2, 'fuji_nx600', 'Cl-P',   'mEq/l',  'numeric',      55, 30,  true, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  (2, 'fuji_nx600', 'LIP-P',  'U/l',    'numeric',      56, 40,  true, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  (2, 'fuji_nx600', 'TP-P',   'g/dl',   'numeric',      57, 50,  true, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  (2, 'fuji_nx600', 'ALB-P',  'g/dl',   'numeric',      58, 60,  true, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  (2, 'fuji_nx600', 'ALPi-P', 'U/l',    'numeric',      59, 70,  true, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  (2, 'fuji_nx600', 'GLU-P',  'mg/dl',  'numeric',      60, 80,  true, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  (2, 'fuji_nx600', 'TBIL-P', 'mg/dl',  'numeric',      61, 90,  true, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  (2, 'fuji_nx600', 'IP-P',   'mg/dl',  'numeric',      62, 100, true, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  (2, 'fuji_nx600', 'TCHO-P', 'mg/dl',  'numeric',      63, 110, true, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  (2, 'fuji_nx600', 'GGT-P',  'U/l',    'numeric',      64, 120, true, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  (2, 'fuji_nx600', 'GPT-P',  'U/l',    'numeric',      21, 130, true, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  (2, 'fuji_nx600', 'Ca-P',   'mg/dl',  'numeric',      65, 140, true, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  (2, 'fuji_nx600', 'CRE-P',  'mg/dl',  'numeric',      23, 150, true, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  (2, 'fuji_nx600', 'BUN-P',  'mg/dl',  'numeric',      22, 160, true, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  -- fuji_au10v (SAA 血液化学検査 exam_type_id=7)
  (2, 'fuji_au10v', 'vf-SAA', 'ug/mL',  'inequality',   66, 10,  true, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  -- arkray_pu4010 (尿検査 exam_type_id=13)
  (2, 'arkray_pu4010', 'GLU', 'mg/dL',  'dash',         67, 10,  true, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  (2, 'arkray_pu4010', 'PRO', 'mg/dL',  'qual_and_num', 38, 20,  true, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  (2, 'arkray_pu4010', 'BIL', 'mg/dL',  'dash',         68, 30,  true, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  (2, 'arkray_pu4010', 'URO', 'mg/dL',  'text',         69, 40,  true, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  (2, 'arkray_pu4010', 'PH',  '',        'numeric',      37, 50,  true, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  (2, 'arkray_pu4010', 'BLD', 'mg/dL',  'qual_and_num', 70, 60,  true, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  (2, 'arkray_pu4010', 'KET', 'mg/dL',  'dash',         71, 70,  true, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  (2, 'arkray_pu4010', 'NIT', 'mg/dL',  'dash',         72, 80,  true, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  -- idexx_vetlab (CBC 血液検査 exam_type_id=6)
  (2, 'idexx_vetlab', 'WBC',   'K/uL',  'numeric',      18, 10,  true, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  (2, 'idexx_vetlab', 'RBC',   'M/uL',  'numeric',      19, 20,  true, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  (2, 'idexx_vetlab', 'HCT',   '%',     'numeric',      20, 30,  true, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  (2, 'idexx_vetlab', 'HGB',   'g/dL',  'numeric',      45, 40,  true, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  (2, 'idexx_vetlab', 'PLT',   'K/uL',  'numeric',      46, 50,  true, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  (2, 'idexx_vetlab', 'NEU',   'K/uL',  'numeric',      47, 60,  true, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  (2, 'idexx_vetlab', 'LYM',   'K/uL',  'numeric',      48, 70,  true, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  (2, 'idexx_vetlab', 'MONO',  'K/uL',  'numeric',      49, 80,  true, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  (2, 'idexx_vetlab', 'EOS',   'K/uL',  'numeric',      50, 90,  true, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  (2, 'idexx_vetlab', 'BASO',  'K/uL',  'numeric',      51, 100, true, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09'),
  (2, 'idexx_vetlab', 'RETIC', 'K/uL',  'numeric',      52, 110, true, '2026-08-20 17:00:00+09', '2026-08-20 17:00:00+09')
ON CONFLICT (clinic_id, source_type, device_item_code) DO NOTHING;

-- EnsureDefaults が先行して NULL のまま残した行を埋める（exam_type_field_id が NULL のみ更新）
UPDATE lab_device_item_masters dst
SET
  exam_type_field_id = src.etf_id,
  updated_at         = now()
FROM (VALUES
  (2::bigint, 'fuji_nx600'::varchar,    'Na-P'::varchar,   53::bigint),
  (2, 'fuji_nx600',    'K-P',     54),
  (2, 'fuji_nx600',    'Cl-P',    55),
  (2, 'fuji_nx600',    'LIP-P',   56),
  (2, 'fuji_nx600',    'TP-P',    57),
  (2, 'fuji_nx600',    'ALB-P',   58),
  (2, 'fuji_nx600',    'ALPi-P',  59),
  (2, 'fuji_nx600',    'GLU-P',   60),
  (2, 'fuji_nx600',    'TBIL-P',  61),
  (2, 'fuji_nx600',    'IP-P',    62),
  (2, 'fuji_nx600',    'TCHO-P',  63),
  (2, 'fuji_nx600',    'GGT-P',   64),
  (2, 'fuji_nx600',    'GPT-P',   21),
  (2, 'fuji_nx600',    'Ca-P',    65),
  (2, 'fuji_nx600',    'CRE-P',   23),
  (2, 'fuji_nx600',    'BUN-P',   22),
  (2, 'fuji_au10v',    'vf-SAA',  66),
  (2, 'arkray_pu4010', 'GLU',     67),
  (2, 'arkray_pu4010', 'PRO',     38),
  (2, 'arkray_pu4010', 'BIL',     68),
  (2, 'arkray_pu4010', 'URO',     69),
  (2, 'arkray_pu4010', 'PH',      37),
  (2, 'arkray_pu4010', 'BLD',     70),
  (2, 'arkray_pu4010', 'KET',     71),
  (2, 'arkray_pu4010', 'NIT',     72),
  (2, 'idexx_vetlab',  'WBC',     18),
  (2, 'idexx_vetlab',  'RBC',     19),
  (2, 'idexx_vetlab',  'HCT',     20),
  (2, 'idexx_vetlab',  'HGB',     45),
  (2, 'idexx_vetlab',  'PLT',     46),
  (2, 'idexx_vetlab',  'NEU',     47),
  (2, 'idexx_vetlab',  'LYM',     48),
  (2, 'idexx_vetlab',  'MONO',    49),
  (2, 'idexx_vetlab',  'EOS',     50),
  (2, 'idexx_vetlab',  'BASO',    51),
  (2, 'idexx_vetlab',  'RETIC',   52)
) AS src(cid, st, code, etf_id)
WHERE dst.clinic_id         = src.cid
  AND dst.source_type       = src.st
  AND dst.device_item_code  = src.code
  AND dst.exam_type_field_id IS NULL;

-- 確認クエリ
SELECT 'lab_devices clinic_id=2 count=' || count(*) FROM lab_devices WHERE clinic_id = 2;
SELECT 'lab_device_item_masters clinic_id=2 count=' || count(*) FROM lab_device_item_masters WHERE clinic_id = 2;
SELECT 'null exam_type_field_id count=' || count(*) FROM lab_device_item_masters WHERE clinic_id = 2 AND exam_type_field_id IS NULL;

COMMIT;
