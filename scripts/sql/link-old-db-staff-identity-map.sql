-- Link old_db staffs using the authoritative cross-clinic identity map
-- (old_db artifact schemaVersion: staff-identity-map-v1).
-- The runner loads TEMP table tmp_staff_identity_map
--   (person_group_id text, classification text, basis text,
--    clinic_code text, legacy_staff_no text, csv_staff_id bigint)
-- before this file and wraps everything in one transaction.
--
-- Policy (old_db handoff conditions): apply CONFIRMED groups only.
-- NEEDS_REVIEW / UNRESOLVED rows keep separate accounts.
-- One accounts row per group + shared staffs.account_id + every member row
-- gets staff_clinic_assignments to all member clinics (home clinic is_main).
-- Accounts that lose their last staff link to a shared account are
-- deactivated. Does not rewrite staffs.id (clinical FKs stay clinic-banded).
-- Never SELECT staff names into output.

-- Fail-closed input validation. Any violation aborts the transaction.
DO $$
DECLARE
  bad bigint;
BEGIN
  -- every CONFIRMED member must resolve to a live staff row
  SELECT count(*) INTO bad
    FROM tmp_staff_identity_map m
   WHERE m.classification = 'CONFIRMED'
     AND NOT EXISTS (
       SELECT 1 FROM staffs s
        WHERE s.id = m.csv_staff_id AND s.deleted_at IS NULL);
  IF bad > 0 THEN
    RAISE EXCEPTION 'identity-map link: % CONFIRMED csv_staff_id do not resolve to live staffs', bad;
  END IF;

  -- a staff row must not appear in two different CONFIRMED groups, and no
  -- group may list the same staff id twice
  SELECT count(*) INTO bad
    FROM (
      SELECT csv_staff_id
        FROM tmp_staff_identity_map
       WHERE classification = 'CONFIRMED'
       GROUP BY csv_staff_id
      HAVING count(DISTINCT person_group_id) > 1 OR count(*) > 1
    ) dup;
  IF bad > 0 THEN
    RAISE EXCEPTION 'identity-map link: % csv_staff_id are duplicated across CONFIRMED rows', bad;
  END IF;

  -- hachioji has no deterministic cross-clinic key in this artifact version;
  -- a CONFIRMED hachioji row means the input is not what we think it is
  IF EXISTS (
    SELECT 1 FROM tmp_staff_identity_map
     WHERE classification = 'CONFIRMED' AND clinic_code = 'hachioji'
  ) THEN
    RAISE EXCEPTION 'identity-map link: hachioji rows must not be CONFIRMED';
  END IF;

  -- CONFIRMED groups must have >=2 members at distinct clinics
  SELECT count(*) INTO bad
    FROM (
      SELECT m.person_group_id
        FROM tmp_staff_identity_map m
        JOIN staffs s ON s.id = m.csv_staff_id
       WHERE m.classification = 'CONFIRMED'
       GROUP BY m.person_group_id
      HAVING count(*) < 2
          OR count(DISTINCT s.clinic_id) < 2
          OR count(*) <> count(DISTINCT m.csv_staff_id)
    ) malformed;
  IF bad > 0 THEN
    RAISE EXCEPTION 'identity-map link: % CONFIRMED groups are malformed (members/clinics)', bad;
  END IF;
END $$;

-- Snapshot of member rows' previous account links, used to deactivate
-- accounts orphaned by repointing.
CREATE TEMP TABLE tmp_identity_prev_links ON COMMIT DROP AS
SELECT s.id AS staff_id, s.account_id
  FROM staffs s
  JOIN tmp_staff_identity_map m
    ON m.csv_staff_id = s.id AND m.classification = 'CONFIRMED'
 WHERE s.account_id IS NOT NULL;

-- Snapshot of member rows' pre-existing assignment count, for reporting.
CREATE TEMP TABLE tmp_identity_prev_assignments ON COMMIT DROP AS
SELECT count(*) AS n
  FROM staff_clinic_assignments a
  JOIN tmp_staff_identity_map m
    ON m.csv_staff_id = a.staff_id AND m.classification = 'CONFIRMED';

WITH members AS (
  SELECT m.person_group_id, s.id AS staff_id, s.clinic_id, s.account_id
    FROM tmp_staff_identity_map m
    JOIN staffs s ON s.id = m.csv_staff_id AND s.deleted_at IS NULL
   WHERE m.classification = 'CONFIRMED'
),
picked AS (
  SELECT
    person_group_id,
    min(account_id) FILTER (WHERE account_id IS NOT NULL) AS existing_account_id
  FROM members
  GROUP BY person_group_id
),
created_accounts AS (
  INSERT INTO accounts (email, password_hash, is_active)
  SELECT
    'olddb-map-' || md5(p.person_group_id) || '@invalid.local',
    '$2b$10$000000000000000000000uUnusableOldDbLinkHashxxxxxx',
    false
  FROM picked p
  WHERE p.existing_account_id IS NULL
  ON CONFLICT (email) DO NOTHING
  RETURNING id, email
),
resolved AS (
  SELECT
    p.person_group_id,
    coalesce(
      p.existing_account_id,
      (
        SELECT a.id
        FROM accounts a
        WHERE a.email = 'olddb-map-' || md5(p.person_group_id) || '@invalid.local'
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
  JOIN resolved r ON r.person_group_id = m.person_group_id
  WHERE s.id = m.staff_id
    AND r.account_id IS NOT NULL
    AND s.account_id IS DISTINCT FROM r.account_id
  RETURNING s.id, s.clinic_id, s.account_id
),
assigned AS (
  INSERT INTO staff_clinic_assignments (staff_id, clinic_id, is_main)
  SELECT m.staff_id, c.clinic_id, (m.clinic_id = c.clinic_id)
  FROM members m
  JOIN resolved r ON r.person_group_id = m.person_group_id
  JOIN (SELECT DISTINCT person_group_id, clinic_id FROM members) c
    ON c.person_group_id = m.person_group_id
  WHERE r.account_id IS NOT NULL
  ON CONFLICT (staff_id, clinic_id) DO NOTHING
  RETURNING staff_id
)
SELECT
  (SELECT count(*) FROM picked) AS groups_linked,
  (SELECT count(*) FROM members) AS staffs_in_groups,
  (SELECT count(*) FROM updated_staffs) AS staffs_repointed,
  (SELECT count(*) FROM assigned) AS assignments_inserted;

-- Deactivate accounts that lost their last live staff link by repointing.
-- Restricted to accounts that were member-linked before this run; the shared
-- (kept) account still has staff links and is untouched. Placeholder accounts
-- are already is_active=false so this is idempotent.
UPDATE accounts a
   SET is_active = false,
       updated_at = now()
  FROM tmp_identity_prev_links p
 WHERE a.id = p.account_id
   AND a.is_active
   AND NOT EXISTS (
     SELECT 1 FROM staffs s
      WHERE s.account_id = a.id AND s.deleted_at IS NULL);

SELECT
  (SELECT count(DISTINCT a.id)
     FROM staff_clinic_assignments a
     JOIN tmp_staff_identity_map m
       ON m.csv_staff_id = a.staff_id AND m.classification = 'CONFIRMED')
    - (SELECT n FROM tmp_identity_prev_assignments) AS assignments_net_added,
  (SELECT count(*)
     FROM accounts a
     JOIN tmp_identity_prev_links p ON p.account_id = a.id
    WHERE NOT a.is_active
      AND NOT EXISTS (
        SELECT 1 FROM staffs s
         WHERE s.account_id = a.id AND s.deleted_at IS NULL)) AS accounts_orphaned_disabled;
