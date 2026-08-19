-- Clinic-owned lab devices. Display names and exam binding are stored here.

CREATE TABLE lab_devices (
    id            bigserial     PRIMARY KEY,
    clinic_id     bigint        NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    source_type   varchar(32)   NOT NULL,
    name          varchar(100)  NOT NULL,
    exam_type_id  bigint,
    is_active     boolean       NOT NULL DEFAULT true,
    sort_order    integer       NOT NULL DEFAULT 0,
    created_at    timestamptz   NOT NULL DEFAULT now(),
    updated_at    timestamptz   NOT NULL DEFAULT now(),
    CONSTRAINT chk_lab_devices_source_type
        CHECK (source_type IN ('fuji_nx600', 'fuji_au10v', 'arkray_pu4010')),
    CONSTRAINT uq_lab_devices_clinic_source
        UNIQUE (clinic_id, source_type),
    CONSTRAINT uq_lab_devices_clinic_name
        UNIQUE (clinic_id, name)
);

ALTER TABLE lab_devices
    ADD CONSTRAINT uq_lab_devices_id_clinic UNIQUE (id, clinic_id);

ALTER TABLE lab_devices
    ADD CONSTRAINT fk_lab_devices_exam_type_clinic
    FOREIGN KEY (exam_type_id, clinic_id)
    REFERENCES exam_types (id, clinic_id)
    ON DELETE RESTRICT;

CREATE INDEX idx_lab_devices_clinic_sort
    ON lab_devices (clinic_id, sort_order);

COMMENT ON TABLE lab_devices IS
    '検査機器マスタ。表示名と検査種別の対応。source_type は電文プロトコル';
COMMENT ON COLUMN lab_devices.name IS
    '医院が付ける機器名。フロント定数には置かない';
COMMENT ON COLUMN lab_devices.exam_type_id IS
    'この機器が載せる検査種別。未設定可';
