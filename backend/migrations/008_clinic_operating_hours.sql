-- EMR-446/466: clinic-line-time storage foundation (selected contract
-- docs/work/emr/emr-clinic-line-time-selected-contract-20261022.md).
-- A03 common clinic operating-hours aggregate (contract §2.1).
-- Storage only: no API/service/handler writer is wired here, no defaults or
-- backfills are invented, and no data migration is performed.
-- Apply only through the normal, human-run migration procedure.

-- One row per clinic (clinic_id PK). Absent row = unconfigured; the row is
-- never complemented to all-open/all-closed. The three settings fields carry
-- no DEFAULT: the first write must be fully explicit. version is the CAS
-- counter, starting at 1 and maintained by app_private.entity_version_guard().
CREATE TABLE clinic_operating_hours (
    clinic_id               bigint      PRIMARY KEY REFERENCES clinics(id) ON DELETE RESTRICT,
    schema_version          text        NOT NULL CHECK (schema_version = 'clinic-operating-hours/v1'),
    version                 bigint      NOT NULL DEFAULT 1 CHECK (version > 0),
    weekly_hours            jsonb       NOT NULL CHECK (jsonb_typeof(weekly_hours) = 'array' AND jsonb_array_length(weekly_hours) = 7),
    closed_dates            jsonb       NOT NULL CHECK (jsonb_typeof(closed_dates) = 'array'),
    national_holiday_closed boolean     NOT NULL,
    created_at              timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at              timestamptz NOT NULL DEFAULT clock_timestamp()
);
COMMENT ON TABLE clinic_operating_hours IS 'EMR-446/466 A03 (contract §2.1): common clinic operating-hours aggregate. Absent row = unconfigured (not complemented to all-open or all-closed). Separate from clinics basics, clinic_settings register-day boundary, and LINE-internal settings; no dual-write with the legacy clinic_holidays table.';
COMMENT ON COLUMN clinic_operating_hours.clinic_id IS 'Tenant key and PK; references clinics(id), ON DELETE RESTRICT';
COMMENT ON COLUMN clinic_operating_hours.schema_version IS 'Storage schema marker, fixed to clinic-operating-hours/v1';
COMMENT ON COLUMN clinic_operating_hours.version IS 'Positive CAS counter starting at 1; maintained by app_private.entity_version_guard() (no-op writes keep the same version)';
COMMENT ON COLUMN clinic_operating_hours.weekly_hours IS 'Array of exactly 7 elements, one per weekday 0..6 in ascending order; each element {weekday,intervals:[{start,end}]}, intervals start-ascending and non-overlapping, empty intervals = explicit closed day. Strict element validation is the server strict decoder responsibility, not a DB CHECK';
COMMENT ON COLUMN clinic_operating_hours.closed_dates IS 'Array of {date,reason}, strict Gregorian YYYY-MM-DD, unique ascending; reason may be empty. Strict element validation is the server strict decoder responsibility';
COMMENT ON COLUMN clinic_operating_hours.national_holiday_closed IS 'Explicit boolean national-holiday policy; no default, first write must supply it';
COMMENT ON COLUMN clinic_operating_hours.created_at IS 'Row creation timestamp (clock_timestamp)';
COMMENT ON COLUMN clinic_operating_hours.updated_at IS 'Row update timestamp (clock_timestamp)';

CREATE TRIGGER clinic_operating_hours_storage_version
    BEFORE UPDATE ON clinic_operating_hours
    FOR EACH ROW EXECUTE FUNCTION app_private.entity_version_guard();

SELECT app_private.apply_rls_policy(
    'clinic_operating_hours'::regclass,
    'clinic_operating_hours_clinic_scope',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)');
