-- EMR-308 cleanup: retire medical_records.next_visit_recommended_date as an
-- independent source of truth. Surviving future-dated recommendations are
-- backfilled into tentative appointments (次回予定: date known; time, slot and
-- doctor undecided; no capacity consumption), then the column is dropped.
--
-- Backfill contract mirrors reservationRepository.SetNextVisitIntent:
--   * only the LATEST medical record per (clinic_id, pet_id) can carry intent.
--     DISTINCT ON selects the newest non-deleted record FIRST (same MAX(id)
--     basis the retired FindOwnersByNextVisitRecommended batch used), and only
--     THEN is that record's date checked in `candidates` — a newer record with
--     a NULL/past date supersedes an older qualifying one and yields no intent;
--   * only dates still in the future (JST calendar date) survive — past dates
--     are expired intent and are not recreated;
--   * deceased pets are excluded the same way runtime readers do
--     (pets.status = 'deceased' OR pets.deceased_at IS NOT NULL — the
--     EMR-290/BUG-407 dual-field OR contract shared by nextVisitLivingPetClause);
--   * sentinel window: start_time = intended_date 00:00 JST,
--     end_time = +15 minutes (nextVisitIntentTimeBounds / 15 min);
--   * reservation_route = 'record_intent', status = 'tentative',
--     doctor_id = NULL, visit_type = 'revisit', source = 'manual',
--     customer_fields = '{}'::jsonb;
--   * reservation_type_id = the clinic's general type (same resolution as
--     findGeneralReservationTypeID: category='general', not deleted,
--     is_active DESC, id ASC). Clinics without a general reservation type are
--     skipped — they cannot express tentative intent in the current model;
--   * uq_appointments_one_active_tentative (026) allows at most one active
--     tentative per (clinic_id, pet_id): NOT EXISTS skips pets already holding
--     one and ON CONFLICT DO NOTHING absorbs any concurrent insert.
--
-- cmd/migrate wraps this file in one transaction — no BEGIN/COMMIT here.
-- 'tentative' was committed to reservation_status by 025. Apply is manual
-- (make migrate). Do not edit 001_init.sql.

WITH latest AS (
    -- 最新カルテ基準: select the latest non-deleted record per (clinic_id, pet_id)
    -- FIRST; the date check lives in `candidates` so a newer record without a
    -- qualifying date supersedes an older one instead of being skipped over.
    SELECT DISTINCT ON (mr.clinic_id, mr.pet_id)
        mr.id                AS medical_record_id,
        mr.clinic_id,
        mr.pet_id,
        mr.next_visit_recommended_date AS intended_date
    FROM medical_records AS mr
    WHERE mr.deleted_at IS NULL
      AND mr.pet_id IS NOT NULL
    ORDER BY mr.clinic_id, mr.pet_id, mr.id DESC
),
candidates AS (
    SELECT
        l.clinic_id,
        l.pet_id,
        l.medical_record_id,
        l.intended_date,
        p.owner_id,
        rt.id AS reservation_type_id
    FROM latest AS l
    JOIN pets AS p
      ON p.id = l.pet_id
     AND p.clinic_id = l.clinic_id
     AND p.deleted_at IS NULL
     -- EMR-290/BUG-407 dual-field deceased contract (see nextVisitLivingPetClause):
     -- both fields are managed independently, so both must be checked.
     AND p.status <> 'deceased'
     AND p.deceased_at IS NULL
    JOIN owners AS o
      ON o.id = p.owner_id
     AND o.clinic_id = l.clinic_id
     AND o.deleted_at IS NULL
    LEFT JOIN LATERAL (
        SELECT rt.id
        FROM reservation_types AS rt
        WHERE rt.clinic_id = l.clinic_id
          AND rt.category = 'general'
          AND rt.deleted_at IS NULL
        ORDER BY rt.is_active DESC, rt.id ASC
        LIMIT 1
    ) AS rt ON true
    WHERE rt.id IS NOT NULL
      AND l.intended_date IS NOT NULL
      AND l.intended_date >= (now() AT TIME ZONE 'Asia/Tokyo')::date
      AND NOT EXISTS (
          SELECT 1
          FROM appointments AS a
          WHERE a.clinic_id = l.clinic_id
            AND a.pet_id = l.pet_id
            AND a.status = 'tentative'
            AND a.deleted_at IS NULL
      )
)
INSERT INTO appointments (
    clinic_id,
    start_time,
    end_time,
    owner_id,
    pet_id,
    visit_type,
    reservation_type_id,
    doctor_id,
    is_designated,
    status,
    notes,
    source,
    created_by,
    is_staff_delegated,
    customer_fields,
    reservation_route,
    intended_date,
    source_medical_record_id,
    created_at,
    updated_at
)
SELECT
    clinic_id,
    intended_date::timestamp AT TIME ZONE 'Asia/Tokyo',
    intended_date::timestamp AT TIME ZONE 'Asia/Tokyo' + interval '15 minutes',
    owner_id,
    pet_id,
    'revisit'::visit_type,
    reservation_type_id,
    NULL,
    false,
    'tentative'::reservation_status,
    '',
    'manual'::reservation_source,
    NULL,
    false,
    '{}'::jsonb,
    'record_intent',
    intended_date,
    medical_record_id,
    now(),
    now()
FROM candidates
ON CONFLICT (clinic_id, pet_id)
    WHERE status = 'tentative' AND deleted_at IS NULL
    DO NOTHING;

ALTER TABLE medical_records
    DROP COLUMN next_visit_recommended_date;
