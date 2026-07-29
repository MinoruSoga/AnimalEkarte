-- POC-05 / LANE-BE ④: DB-level non-overlap for closing_special_periods.
-- Complements Go CreateCheckingOverlap/UpdateCheckingOverlap (clinic advisory lock).
-- btree_gist enables equality on clinic_id together with daterange overlap.

CREATE EXTENSION IF NOT EXISTS btree_gist;

ALTER TABLE closing_special_periods
    ADD CONSTRAINT excl_closing_special_periods_clinic_daterange
    EXCLUDE USING gist (
        clinic_id WITH =,
        daterange(start_date, end_date, '[]') WITH &&
    );

COMMENT ON CONSTRAINT excl_closing_special_periods_clinic_daterange ON closing_special_periods IS
    'POC-05: no overlapping special periods within the same clinic';
