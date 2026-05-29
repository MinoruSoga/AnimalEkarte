-- ------------------------------------
-- staff_reservation_capabilities（スタッフ × 対応可能予約区分 M:N）
-- ------------------------------------
CREATE TABLE IF NOT EXISTS staff_reservation_capabilities (
    id                  BIGSERIAL PRIMARY KEY,
    clinic_id           bigint      NOT NULL REFERENCES clinics(id),
    staff_id            bigint      NOT NULL REFERENCES staffs(id) ON DELETE CASCADE,
    reservation_type_id bigint      NOT NULL REFERENCES reservation_types(id) ON DELETE CASCADE,
    created_at          timestamptz NOT NULL DEFAULT now(),
    UNIQUE(clinic_id, staff_id, reservation_type_id)
);

CREATE INDEX IF NOT EXISTS idx_staff_reservation_capabilities_clinic_staff
    ON staff_reservation_capabilities(clinic_id, staff_id);
CREATE INDEX IF NOT EXISTS idx_staff_reservation_capabilities_clinic_type
    ON staff_reservation_capabilities(clinic_id, reservation_type_id);

COMMENT ON TABLE staff_reservation_capabilities IS 'スタッフ対応可能予約区分';

-- 既存の「非対応予約区分」設定を、正の「対応可能予約区分」へ移行する。
INSERT INTO staff_reservation_capabilities (clinic_id, staff_id, reservation_type_id)
SELECT
    sca.clinic_id,
    s.id,
    rt.id
FROM staffs s
JOIN staff_clinic_assignments sca
    ON sca.staff_id = s.id
JOIN reservation_types rt
    ON rt.clinic_id = sca.clinic_id
    AND rt.deleted_at IS NULL
WHERE s.deleted_at IS NULL
  AND NOT EXISTS (
      SELECT 1
      FROM staff_reservation_exclusions sre
      WHERE sre.staff_id = s.id
        AND sre.reservation_type_id = rt.id
  )
ON CONFLICT (clinic_id, staff_id, reservation_type_id) DO NOTHING;

-- 会計明細を発生元の処置・トリミング appointment と紐付け、未請求候補の再表示を防ぐ。
ALTER TYPE item_source ADD VALUE IF NOT EXISTS 'trimming';

ALTER TABLE billing_items
    ADD COLUMN IF NOT EXISTS treatment_id bigint REFERENCES treatments(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS appointment_id bigint REFERENCES appointments(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS trimming_course_id bigint REFERENCES trimming_courses(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS trimming_option_id bigint REFERENCES trimming_options(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_billing_items_treatment_id
    ON billing_items(treatment_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_billing_items_appointment_id
    ON billing_items(appointment_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_billing_items_trimming_course_id
    ON billing_items(trimming_course_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_billing_items_trimming_option_id
    ON billing_items(trimming_option_id)
    WHERE deleted_at IS NULL;
