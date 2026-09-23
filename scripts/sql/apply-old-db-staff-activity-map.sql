-- Activate migrated staffs using the authoritative old_db staff-activity map
-- (old_db artifact schemaVersion: staff-activity-map-manifest-v1).
-- The runner loads TEMP table tmp_staff_activity_map
--   (csv_staff_id bigint, clinic_code text, classification text,
--    basis text, last_activity_bucket text)
-- before this file and wraps everything in one transaction.
--
-- Policy (old_db handoff conditions): set is_active=true for ACTIVE_CONFIRMED
-- rows only. ACTIVE_LIKELY / LIKELY_RETIRED / NO_EVIDENCE rows are never
-- touched (already-inactive by import policy; this script never deactivates).
-- Does not touch accounts, assignments, staffs.id, or reservation_visible.
-- Never SELECT staff names into output.

-- Fail-closed input validation. Any violation aborts the transaction.
DO $$
DECLARE
  bad bigint;
BEGIN
  -- every ACTIVE_CONFIRMED csv_staff_id must resolve to a live staff row
  SELECT count(*) INTO bad
    FROM tmp_staff_activity_map m
   WHERE m.classification = 'ACTIVE_CONFIRMED'
     AND NOT EXISTS (
       SELECT 1 FROM staffs s
        WHERE s.id = m.csv_staff_id AND s.deleted_at IS NULL);
  IF bad > 0 THEN
    RAISE EXCEPTION 'activity-map apply: % ACTIVE_CONFIRMED csv_staff_id do not resolve to live staffs', bad;
  END IF;

  -- a staff row must not appear more than once in the map
  IF EXISTS (
    SELECT 1 FROM (
      SELECT csv_staff_id
        FROM tmp_staff_activity_map
       GROUP BY csv_staff_id
      HAVING count(*) > 1
    ) dup
  ) THEN
    RAISE EXCEPTION 'activity-map apply: duplicate csv_staff_id in map';
  END IF;

  -- classification values must be within the known contract
  IF EXISTS (
    SELECT 1 FROM tmp_staff_activity_map
     WHERE classification NOT IN
       ('ACTIVE_CONFIRMED', 'ACTIVE_LIKELY', 'LIKELY_RETIRED', 'NO_EVIDENCE')
  ) THEN
    RAISE EXCEPTION 'activity-map apply: unknown classification value';
  END IF;
END $$;

-- Activate confirmed-current staff rows. Idempotent: already-active rows are
-- skipped by the predicate.
UPDATE staffs s
   SET is_active = true,
       updated_at = now()
  FROM tmp_staff_activity_map m
 WHERE m.classification = 'ACTIVE_CONFIRMED'
   AND s.id = m.csv_staff_id
   AND s.deleted_at IS NULL
   AND NOT s.is_active;

SELECT
  (SELECT count(*) FROM tmp_staff_activity_map WHERE classification = 'ACTIVE_CONFIRMED')
    AS map_active_confirmed,
  (SELECT count(*) FROM staffs s JOIN tmp_staff_activity_map m
     ON s.id = m.csv_staff_id AND m.classification = 'ACTIVE_CONFIRMED'
    WHERE s.deleted_at IS NULL AND s.is_active) AS staffs_now_active;
