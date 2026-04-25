-- LSTEP-BE-012: 慢性疾患フラグ管理テーブル
CREATE TABLE pet_chronic_conditions (
  id            BIGSERIAL     PRIMARY KEY,
  clinic_id     BIGINT        NOT NULL REFERENCES clinics(id),
  pet_id        BIGINT        NOT NULL REFERENCES pets(id),
  condition_code VARCHAR(50)  NOT NULL,
  condition_name VARCHAR(100) NOT NULL,
  diagnosed_at  DATE          NOT NULL,
  notes         TEXT,
  is_active     BOOLEAN       NOT NULL DEFAULT TRUE,
  created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
  deleted_at    TIMESTAMPTZ   NULL
);

CREATE INDEX idx_pet_chronic_conditions_clinic_pet
  ON pet_chronic_conditions (clinic_id, pet_id)
  WHERE deleted_at IS NULL;
