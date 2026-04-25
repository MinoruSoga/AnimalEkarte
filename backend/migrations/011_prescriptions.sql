-- 011_prescriptions.sql: 処方薬記録テーブル（LSTEP-BE-009）
CREATE TABLE prescriptions (
    id                BIGSERIAL       PRIMARY KEY,
    clinic_id         BIGINT          NOT NULL REFERENCES clinics(id),
    owner_id          BIGINT          NOT NULL REFERENCES owners(id),
    pet_id            BIGINT                   REFERENCES pets(id),
    medical_record_id BIGINT                   REFERENCES medical_records(id),
    prescribed_at     DATE            NOT NULL,
    duration_days     INT             NOT NULL DEFAULT 0,
    deleted_at        TIMESTAMPTZ,
    created_at        TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_prescriptions_clinic_owner ON prescriptions(clinic_id, owner_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_prescriptions_medical_record ON prescriptions(medical_record_id) WHERE deleted_at IS NULL;
