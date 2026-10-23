-- EMR-446/466: clinic-line-time storage foundation (selected contract
-- docs/work/emr/emr-clinic-line-time-selected-contract-20261022.md).
-- A03 LINE acceptance version columns + storage guard (contract §3.1).
-- Storage only: no API/service/handler writer is wired here, no defaults or
-- backfills are invented, and no data migration is performed.
-- Apply only through the normal, human-run migration procedure.

-- line_reservation_settings stays the existing separate storage; only the
-- acceptance_version CAS counter and reviewed_acceptance_version confirmation
-- marker are added. Existing rows keep acceptance_version=1 and
-- reviewed_acceptance_version=NULL (legacy-unreviewed); no backfill marks old
-- rows as reviewed.
ALTER TABLE line_reservation_settings
    ADD COLUMN acceptance_version bigint NOT NULL DEFAULT 1 CHECK (acceptance_version > 0),
    ADD COLUMN reviewed_acceptance_version bigint,
    ADD CONSTRAINT ck_line_reservation_reviewed_acceptance_version
        CHECK (reviewed_acceptance_version IS NULL OR (reviewed_acceptance_version > 0 AND reviewed_acceptance_version <= acceptance_version));

COMMENT ON COLUMN line_reservation_settings.acceptance_version IS 'EMR-446/466 A03 (contract §3.1): positive CAS counter maintained by app_private.line_acceptance_version_guard(). Advances once per write that changes a protected field P or the reviewed marker; notification/display-only writes and true no-ops keep the same version';
COMMENT ON COLUMN line_reservation_settings.reviewed_acceptance_version IS 'EMR-446/466 A03 (contract §3.1): acceptance_version the clinic confirmed. NULL or <> acceptance_version = legacy-unreviewed; equal = configured. Writers cannot reset it to NULL (guard restores OLD value)';

-- Dedicated guard for the acceptance CAS pair. Protected field set P is closed
-- to the 12 contract columns (contract §3.1):
--   status, closed_weekdays, closed_dates, national_holiday_closed,
--   business_hours, business_hours_by_weekday, break_hours,
--   booking_window_max_days, booking_window_min_days,
--   time_slot_mode, time_slot_interval_minutes, show_no_staff_option
-- daily_limit / monthly_limit / no_staff_mode are NOT in P. N (display fields
-- such as header_text) and additional_fields do not advance the version.
--
-- Consequences:
--   - INSERT: acceptance_version is always 1 (writer values are not adopted);
--     reviewed is NULL or 1 (column CHECK).
--   - UPDATE: the resulting acceptance_version is OLD+1 when any P column or
--     the reviewed marker changed, OLD otherwise; writer-SET version values are
--     always overwritten (no double increments, stale writers cannot Save a
--     zeroed version).
--   - A confirm/update writer SETs reviewed_acceptance_version := expected+1;
--     the guard requires it to equal the resulting acceptance_version.
--   - A stale writer changing P advances the version while reviewed stays
--     behind = fail-closed back to legacy-unreviewed.
--   - Setting reviewed to NULL restores OLD (confirmation state cannot be
--     cleared by a writer that saves a struct without the field).
CREATE FUNCTION app_private.line_acceptance_version_guard()
RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
    p_changed boolean;
    r_changed boolean;
BEGIN
    IF TG_OP = 'INSERT' THEN
        NEW.acceptance_version := 1;
        RETURN NEW;
    END IF;
    IF NEW.reviewed_acceptance_version IS NULL THEN
        NEW.reviewed_acceptance_version := OLD.reviewed_acceptance_version;
    END IF;
    p_changed := (NEW.status, NEW.closed_weekdays, NEW.closed_dates,
                  NEW.national_holiday_closed, NEW.business_hours,
                  NEW.business_hours_by_weekday, NEW.break_hours,
                  NEW.booking_window_max_days, NEW.booking_window_min_days,
                  NEW.time_slot_mode, NEW.time_slot_interval_minutes,
                  NEW.show_no_staff_option)
             IS DISTINCT FROM
                 (OLD.status, OLD.closed_weekdays, OLD.closed_dates,
                  OLD.national_holiday_closed, OLD.business_hours,
                  OLD.business_hours_by_weekday, OLD.break_hours,
                  OLD.booking_window_max_days, OLD.booking_window_min_days,
                  OLD.time_slot_mode, OLD.time_slot_interval_minutes,
                  OLD.show_no_staff_option);
    r_changed := NEW.reviewed_acceptance_version IS DISTINCT FROM OLD.reviewed_acceptance_version;
    NEW.acceptance_version := OLD.acceptance_version + CASE WHEN p_changed OR r_changed THEN 1 ELSE 0 END;
    IF r_changed AND NEW.reviewed_acceptance_version <> NEW.acceptance_version THEN
        RAISE EXCEPTION 'reviewed_acceptance_version must equal the resulting acceptance_version' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
END; $$;
COMMENT ON FUNCTION app_private.line_acceptance_version_guard() IS 'EMR-446/466 A03 (contract §3.1): maintains acceptance_version/reviewed_acceptance_version on line_reservation_settings. State derivation: absent row = unconfigured; reviewed IS NULL or reviewed <> acceptance_version = legacy-unreviewed; reviewed = acceptance_version = configured. True no-op keeps the same version; non-P (header_text etc., additional_fields) changes do not advance it';

CREATE TRIGGER line_reservation_settings_acceptance_guard
    BEFORE INSERT OR UPDATE ON line_reservation_settings
    FOR EACH ROW EXECUTE FUNCTION app_private.line_acceptance_version_guard();
