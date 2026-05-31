CREATE TABLE IF NOT EXISTS reservation_type_available_slots (
    id                  BIGSERIAL   PRIMARY KEY,
    clinic_id           bigint      NOT NULL REFERENCES clinics(id),
    reservation_type_id bigint      NOT NULL REFERENCES reservation_types(id) ON DELETE CASCADE,
    available_type      text        NOT NULL CHECK (available_type IN ('weekly', 'specific')),
    day_of_week         smallint    CHECK (
                            (available_type = 'weekly' AND day_of_week BETWEEN 0 AND 6)
                            OR (available_type = 'specific' AND day_of_week IS NULL)
                        ),
    specific_date       date        CHECK (
                            (available_type = 'specific' AND specific_date IS NOT NULL)
                            OR (available_type = 'weekly' AND specific_date IS NULL)
                        ),
    start_time          varchar(5)  NOT NULL CHECK (start_time ~ '^\d{2}:\d{2}$'),
    is_active           boolean     NOT NULL DEFAULT true,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_rtype_available_slot_clinic_type
    ON reservation_type_available_slots(clinic_id, reservation_type_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_rtype_available_slot_weekly_unique
    ON reservation_type_available_slots(reservation_type_id, day_of_week, start_time)
    WHERE available_type = 'weekly';
CREATE UNIQUE INDEX IF NOT EXISTS idx_rtype_available_slot_specific_unique
    ON reservation_type_available_slots(reservation_type_id, specific_date, start_time)
    WHERE available_type = 'specific';

COMMENT ON TABLE reservation_type_available_slots IS '予約区分予約可能開始時刻';
