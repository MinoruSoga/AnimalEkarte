package lstep

const checkupSyncPreviewSQL = `
WITH visit_agg AS (
  SELECT
    p.owner_id,
    MAX(mr.date) AS last_visit_date,
    MIN(mr.date) AS first_visit_date,
    COUNT(DISTINCT mr.date) AS total_visit_count,
    COUNT(DISTINCT CASE WHEN mr.date >= NOW() - INTERVAL '365 days' THEN mr.date END) AS annual_visit_count
  FROM medical_records mr
  INNER JOIN pets p
    ON p.id = mr.pet_id
   AND p.clinic_id = mr.clinic_id
  WHERE mr.clinic_id = ?
    AND mr.deleted_at IS NULL
  GROUP BY p.owner_id
),
pet_agg AS (
  SELECT
    p.owner_id,
    COALESCE(STRING_AGG(DISTINCT p.name, ',' ORDER BY p.name) FILTER (WHERE p.deceased_at IS NULL), '') AS pet_names,
    COUNT(DISTINCT p.id) FILTER (WHERE p.deceased_at IS NULL) AS living_pet_count,
    MIN(EXTRACT(YEAR FROM AGE(NOW(), p.birth_date))::int) FILTER (WHERE p.deceased_at IS NULL AND p.birth_date IS NOT NULL) AS min_pet_age_years,
    MAX(EXTRACT(YEAR FROM AGE(NOW(), p.birth_date))::int) FILTER (WHERE p.deceased_at IS NULL AND p.birth_date IS NOT NULL) AS max_pet_age_years
  FROM pets p
  WHERE p.clinic_id = ?
    AND p.deleted_at IS NULL
  GROUP BY p.owner_id
),
billing_agg AS (
  SELECT
    b.owner_id,
    COALESCE(SUM(b.total_amount), 0) AS total_amount,
    COALESCE(MAX(b.total_amount), 0) AS max_single_visit_amount
  FROM billings b
  WHERE b.clinic_id = ?
    AND b.status = ?
    AND b.deleted_at IS NULL
    AND b.owner_id IS NOT NULL
  GROUP BY b.owner_id
),
checkup_agg AS (
  SELECT
    p.owner_id,
    MAX(c.date) AS last_checkup_date
  FROM checkups c
  INNER JOIN medical_records mrc
    ON mrc.id = c.medical_record_id
   AND mrc.deleted_at IS NULL
   AND mrc.clinic_id = c.clinic_id
  INNER JOIN pets p
    ON p.id = mrc.pet_id
   AND p.clinic_id = mrc.clinic_id
  WHERE c.clinic_id = ?
    AND c.deleted_at IS NULL
  GROUP BY p.owner_id
),
chronic_owners AS (
  SELECT DISTINCT pp.owner_id
  FROM pet_chronic_conditions pcc
  INNER JOIN pets pp
    ON pp.id = pcc.pet_id
   AND pp.deleted_at IS NULL
   AND pp.deceased_at IS NULL
   AND pp.clinic_id = pcc.clinic_id
  WHERE pcc.clinic_id = ?
    AND pcc.is_active = TRUE
    AND pcc.deleted_at IS NULL
)
SELECT
  o.id          AS owner_id,
  o.name        AS owner_name,
  o.line_user_id,
  o.lstep_opt_out,
  COALESCE(pa.pet_names, '') AS pet_names,
  COALESCE(pa.living_pet_count, 0) AS living_pet_count,
  va.last_visit_date,
  va.first_visit_date,
  COALESCE(va.total_visit_count, 0) AS total_visit_count,
  COALESCE(va.annual_visit_count, 0) AS annual_visit_count,
  pa.min_pet_age_years,
  pa.max_pet_age_years,
  (co.owner_id IS NOT NULL) AS has_chronic_condition,
  COALESCE(ba.total_amount, 0) AS total_amount,
  COALESCE(ba.max_single_visit_amount, 0) AS max_single_visit_amount,
  ca.last_checkup_date
FROM owners o
LEFT JOIN visit_agg va ON va.owner_id = o.id
LEFT JOIN pet_agg pa ON pa.owner_id = o.id
LEFT JOIN billing_agg ba ON ba.owner_id = o.id
LEFT JOIN checkup_agg ca ON ca.owner_id = o.id
LEFT JOIN chronic_owners co ON co.owner_id = o.id
WHERE %s
%s
ORDER BY va.last_visit_date DESC NULLS LAST
LIMIT ?
`
