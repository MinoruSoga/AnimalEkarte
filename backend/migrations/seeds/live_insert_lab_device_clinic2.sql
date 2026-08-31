-- One-shot clinic 2 lab-device master repair.
-- This script resolves exam references by tenant-scoped semantic names. It
-- never relies on environment-specific exam_type/exam_type_field numeric IDs.
-- Run only through the approved staging/live runbook with psql ON_ERROR_STOP=1.

BEGIN;
SELECT pg_advisory_xact_lock(333002);

DO $$
DECLARE n bigint;
BEGIN
  SELECT count(*) INTO n FROM clinics
  WHERE id = 2 AND name = '城東センター病院' AND is_active;
  IF n <> 1 THEN
    RAISE EXCEPTION 'clinic 2 semantic identity mismatch';
  END IF;
END $$;

CREATE TEMP TABLE desired_lab_items (
  clinic_id bigint NOT NULL,
  source_type text NOT NULL,
  device_item_code text NOT NULL,
  field_name text NOT NULL,
  unit text NOT NULL,
  value_shape text NOT NULL,
  sort_order integer NOT NULL,
  PRIMARY KEY (clinic_id, source_type, device_item_code)
) ON COMMIT DROP;
INSERT INTO desired_lab_items VALUES
  (2, 'fuji_nx600', 'Na-P', 'Na', 'mEq/l', 'numeric', 10),
  (2, 'fuji_nx600', 'K-P', 'K', 'mEq/l', 'numeric', 20),
  (2, 'fuji_nx600', 'Cl-P', 'Cl', 'mEq/l', 'numeric', 30),
  (2, 'fuji_nx600', 'LIP-P', 'LIP-P', 'U/l', 'numeric', 40),
  (2, 'fuji_nx600', 'TP-P', 'TP', 'g/dl', 'numeric', 50),
  (2, 'fuji_nx600', 'ALB-P', 'ALB', 'g/dl', 'numeric', 60),
  (2, 'fuji_nx600', 'ALPi-P', 'ALP', 'U/l', 'numeric', 70),
  (2, 'fuji_nx600', 'GLU-P', 'Glu', 'mg/dl', 'numeric', 80),
  (2, 'fuji_nx600', 'TBIL-P', 'TBIL', 'mg/dl', 'numeric', 90),
  (2, 'fuji_nx600', 'IP-P', 'IP', 'mg/dl', 'numeric', 100),
  (2, 'fuji_nx600', 'TCHO-P', 'CHOL', 'mg/dl', 'numeric', 110),
  (2, 'fuji_nx600', 'GGT-P', 'GGT', 'U/l', 'numeric', 120),
  (2, 'fuji_nx600', 'GPT-P', 'GPT', 'U/l', 'numeric', 130),
  (2, 'fuji_nx600', 'Ca-P', 'Ca', 'mg/dl', 'numeric', 140),
  (2, 'fuji_nx600', 'CRE-P', 'Cre', 'mg/dl', 'numeric', 150),
  (2, 'fuji_nx600', 'BUN-P', 'BUN', 'mg/dl', 'numeric', 160),
  (2, 'fuji_au10v', 'vf-SAA', 'vf-SAA', 'ug/mL', 'inequality', 10),
  (2, 'arkray_pu4010', 'GLU', '尿糖', 'mg/dL', 'dash', 10),
  (2, 'arkray_pu4010', 'PRO', '尿蛋白', 'mg/dL', 'qual_and_num', 20),
  (2, 'arkray_pu4010', 'BIL', 'ビリルビン', 'mg/dL', 'dash', 30),
  (2, 'arkray_pu4010', 'URO', 'ウロビリノーゲン', 'mg/dL', 'text', 40),
  (2, 'arkray_pu4010', 'PH', 'pH', '', 'numeric', 50),
  (2, 'arkray_pu4010', 'BLD', '潜血', 'mg/dL', 'qual_and_num', 60),
  (2, 'arkray_pu4010', 'KET', '尿ケトン', 'mg/dL', 'dash', 70),
  (2, 'arkray_pu4010', 'NIT', 'NIT', 'mg/dL', 'dash', 80),
  (2, 'idexx_vetlab', 'WBC', 'WBC', 'K/uL', 'numeric', 10),
  (2, 'idexx_vetlab', 'RBC', 'RBC', 'M/uL', 'numeric', 20),
  (2, 'idexx_vetlab', 'HCT', 'HCT', '%', 'numeric', 30),
  (2, 'idexx_vetlab', 'HGB', 'HGB', 'g/dL', 'numeric', 40),
  (2, 'idexx_vetlab', 'PLT', 'PLT', 'K/uL', 'numeric', 50),
  (2, 'idexx_vetlab', 'NEU', 'NEU', 'K/uL', 'numeric', 60),
  (2, 'idexx_vetlab', 'LYM', 'LYM', 'K/uL', 'numeric', 70),
  (2, 'idexx_vetlab', 'MONO', 'MONO', 'K/uL', 'numeric', 80),
  (2, 'idexx_vetlab', 'EOS', 'EOS', 'K/uL', 'numeric', 90),
  (2, 'idexx_vetlab', 'BASO', 'BASO', 'K/uL', 'numeric', 100),
  (2, 'idexx_vetlab', 'RETIC', 'RETIC', 'K/uL', 'numeric', 110);

CREATE TEMP TABLE resolved_lab_items ON COMMIT DROP AS
SELECT d.*, f.id AS exam_type_field_id, f.exam_type_id
FROM desired_lab_items d
JOIN exam_type_fields f
  ON f.clinic_id = d.clinic_id AND f.name = d.field_name;

DO $$
DECLARE expected_count bigint; resolved_count bigint; ambiguous_count bigint;
BEGIN
  SELECT count(*) INTO expected_count FROM desired_lab_items;
  SELECT count(*) INTO resolved_count FROM resolved_lab_items;
  SELECT count(*) INTO ambiguous_count
  FROM (SELECT d.source_type, d.device_item_code
        FROM desired_lab_items d
        JOIN exam_type_fields f ON f.clinic_id=d.clinic_id AND f.name=d.field_name
        GROUP BY d.source_type, d.device_item_code HAVING count(*) <> 1) x;
  IF resolved_count <> expected_count OR ambiguous_count <> 0 THEN
    RAISE EXCEPTION 'exam field semantic resolution mismatch: expected %, resolved %, ambiguous %', expected_count, resolved_count, ambiguous_count;
  END IF;
  IF EXISTS (SELECT 1 FROM resolved_lab_items GROUP BY source_type HAVING count(DISTINCT exam_type_id) <> 1) THEN
    RAISE EXCEPTION 'device items resolve to multiple exam types';
  END IF;
END $$;

CREATE TEMP TABLE desired_lab_devices ON COMMIT DROP AS
SELECT r.clinic_id, r.source_type,
       CASE r.source_type WHEN 'fuji_nx600' THEN 'NX600' WHEN 'fuji_au10v' THEN 'AU10V'
            WHEN 'arkray_pu4010' THEN '尿（PU-4010）' WHEN 'idexx_vetlab' THEN 'IDEXX VetLab' END AS name,
       min(r.exam_type_id) AS exam_type_id,
       CASE r.source_type WHEN 'fuji_nx600' THEN 10 WHEN 'fuji_au10v' THEN 20
            WHEN 'arkray_pu4010' THEN 30 WHEN 'idexx_vetlab' THEN 40 END AS sort_order
FROM resolved_lab_items r GROUP BY r.clinic_id, r.source_type;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM lab_devices e JOIN desired_lab_devices d USING (clinic_id, source_type)
    WHERE (e.name, e.exam_type_id, e.is_active, e.sort_order)
       IS DISTINCT FROM (d.name, d.exam_type_id, true, d.sort_order)
  ) THEN RAISE EXCEPTION 'existing lab_devices row conflicts with desired semantic mapping'; END IF;
  IF EXISTS (
    SELECT 1 FROM lab_device_item_masters e
    JOIN resolved_lab_items d USING (clinic_id, source_type, device_item_code)
    WHERE (e.unit, e.value_shape, e.sort_order, e.is_active)
       IS DISTINCT FROM (d.unit, d.value_shape, d.sort_order, true)
       OR (e.exam_type_field_id IS NOT NULL AND e.exam_type_field_id <> d.exam_type_field_id)
  ) THEN RAISE EXCEPTION 'existing lab item row conflicts with desired semantic mapping'; END IF;
END $$;

INSERT INTO lab_devices (clinic_id, source_type, name, exam_type_id, is_active, sort_order)
SELECT clinic_id, source_type, name, exam_type_id, true, sort_order
FROM desired_lab_devices d
WHERE NOT EXISTS (SELECT 1 FROM lab_devices e WHERE e.clinic_id=d.clinic_id AND e.source_type=d.source_type);

INSERT INTO lab_device_item_masters
  (clinic_id, source_type, device_item_code, unit, value_shape, exam_type_field_id, sort_order, is_active)
SELECT clinic_id, source_type, device_item_code, unit, value_shape, exam_type_field_id, sort_order, true
FROM resolved_lab_items d
WHERE NOT EXISTS (SELECT 1 FROM lab_device_item_masters e
                  WHERE e.clinic_id=d.clinic_id AND e.source_type=d.source_type AND e.device_item_code=d.device_item_code);

UPDATE lab_device_item_masters e
SET exam_type_field_id=d.exam_type_field_id, updated_at=now()
FROM resolved_lab_items d
WHERE e.clinic_id=d.clinic_id AND e.source_type=d.source_type
  AND e.device_item_code=d.device_item_code AND e.exam_type_field_id IS NULL;

DO $$
BEGIN
  IF (SELECT count(*) FROM lab_devices e JOIN desired_lab_devices d USING (clinic_id, source_type)
      WHERE (e.name,e.exam_type_id,e.is_active,e.sort_order)=(d.name,d.exam_type_id,true,d.sort_order)) <> 4 THEN
    RAISE EXCEPTION 'lab device postcondition mismatch';
  END IF;
  IF (SELECT count(*) FROM lab_device_item_masters e JOIN resolved_lab_items d
      USING (clinic_id,source_type,device_item_code)
      WHERE (e.unit,e.value_shape,e.exam_type_field_id,e.sort_order,e.is_active)
          =(d.unit,d.value_shape,d.exam_type_field_id,d.sort_order,true)) <> (SELECT count(*) FROM desired_lab_items) THEN
    RAISE EXCEPTION 'lab item postcondition mismatch';
  END IF;
END $$;

COMMIT;
