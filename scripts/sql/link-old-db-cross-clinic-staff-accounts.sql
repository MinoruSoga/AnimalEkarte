-- Link old_db staffs that share one unambiguous display name across clinics.
-- One accounts row (login disabled) + shared staffs.account_id + home-clinic
-- staff_clinic_assignments. Does not rewrite staffs.id.
-- Skips: blank names; 2+ staffs of that name in one clinic; conflicting licenses.
-- Never SELECT staff names into notices.

-- password_hash is unusable (not a valid bcrypt match). Login stays staff-attach/demo.
WITH normalized AS (
  SELECT
    s.id,
    s.clinic_id,
    s.account_id,
    s.license_number,
    regexp_replace(btrim(replace(s.name, '　', ' ')), '\s+', ' ', 'g') AS name_key
  FROM staffs s
  WHERE s.deleted_at IS NULL
    AND btrim(s.name) <> ''
),
per_clinic AS (
  SELECT name_key, clinic_id, count(*) AS n
  FROM normalized
  GROUP BY name_key, clinic_id
),
eligible_names AS (
  SELECT name_key
  FROM per_clinic
  GROUP BY name_key
  HAVING bool_and(n = 1)
     AND count(*) >= 2
),
license_conflict AS (
  SELECT n.name_key
  FROM normalized n
  JOIN eligible_names e ON e.name_key = n.name_key
  WHERE btrim(coalesce(n.license_number, '')) <> ''
  GROUP BY n.name_key
  HAVING count(DISTINCT btrim(n.license_number)) > 1
),
members AS (
  SELECT n.id, n.clinic_id, n.account_id, n.name_key
  FROM normalized n
  JOIN eligible_names e ON e.name_key = n.name_key
  WHERE n.name_key NOT IN (SELECT name_key FROM license_conflict)
),
picked AS (
  SELECT
    name_key,
    min(account_id) FILTER (WHERE account_id IS NOT NULL) AS existing_account_id
  FROM members
  GROUP BY name_key
),
created_accounts AS (
  INSERT INTO accounts (email, password_hash, is_active)
  SELECT
    'olddb-link-' || md5(p.name_key) || '@invalid.local',
    '$2b$10$000000000000000000000uUnusableOldDbLinkHashxxxxxx',
    false
  FROM picked p
  WHERE p.existing_account_id IS NULL
  ON CONFLICT (email) DO NOTHING
  RETURNING id, email
),
resolved AS (
  SELECT
    p.name_key,
    coalesce(
      p.existing_account_id,
      (
        SELECT a.id
        FROM accounts a
        WHERE a.email = 'olddb-link-' || md5(p.name_key) || '@invalid.local'
        LIMIT 1
      )
    ) AS account_id
  FROM picked p
),
updated_staffs AS (
  UPDATE staffs s
  SET account_id = r.account_id,
      updated_at = now()
  FROM members m
  JOIN resolved r ON r.name_key = m.name_key
  WHERE s.id = m.id
    AND r.account_id IS NOT NULL
    AND s.account_id IS DISTINCT FROM r.account_id
  RETURNING s.id, s.clinic_id, s.account_id
),
assigned AS (
  INSERT INTO staff_clinic_assignments (staff_id, clinic_id, is_main)
  SELECT m.id, c.clinic_id, (m.clinic_id = c.clinic_id)
  FROM members m
  JOIN resolved r ON r.name_key = m.name_key
  JOIN (SELECT DISTINCT name_key, clinic_id FROM members) c ON c.name_key = m.name_key
  WHERE r.account_id IS NOT NULL
  ON CONFLICT (staff_id, clinic_id) DO NOTHING
  RETURNING staff_id
)
SELECT
  (SELECT count(*) FROM resolved WHERE account_id IS NOT NULL) AS groups_linked,
  (SELECT count(*) FROM members) AS staffs_in_groups;
