-- LSA-15 / LANE-BE ④: day-grain uniqueness for delivery trigger double-fire defense.
-- Complements Go CreateIfAbsentToday (advisory lock). Expression uses Asia/Tokyo date
-- to match application "today" boundaries used by ExistsTodayByOwnerAndType.

CREATE UNIQUE INDEX IF NOT EXISTS uk_lstep_delivery_trigger_log_clinic_owner_type_day
    ON lstep_delivery_trigger_log (
        clinic_id,
        owner_id,
        trigger_type,
        ((scheduled_at AT TIME ZONE 'Asia/Tokyo')::date)
    );

COMMENT ON INDEX uk_lstep_delivery_trigger_log_clinic_owner_type_day IS
    'LSA-15: at most one trigger log per clinic/owner/type/JST day';
