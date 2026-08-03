-- 004_examination_revisions.sql
-- TASK-027 Slice A: append-only examination revision foundation and first-confirm pointer CAS.
-- Apply only through the USER-operated migration workflow. This file is additive; 001-003 remain unchanged.

ALTER TABLE exams
    ADD COLUMN current_revision_version BIGINT NULL
        CHECK (current_revision_version >= 1),
    ADD CONSTRAINT uq_exams_clinic_id_id
        UNIQUE (clinic_id, id);

CREATE TABLE examination_revisions (
    id                      BIGSERIAL   PRIMARY KEY,
    clinic_id               BIGINT      NOT NULL,
    examination_id          BIGINT      NOT NULL,
    version                 BIGINT      NOT NULL CHECK (version >= 1),
    kind                    TEXT        NOT NULL CHECK (kind IN ('working', 'official')),
    status                  exam_status NOT NULL,
    medical_record_id       BIGINT,
    pet_id                  BIGINT,
    medical_record_owner_id BIGINT,
    pet_owner_id            BIGINT,
    animal_species_id       BIGINT,
    exam_type_id            BIGINT      NOT NULL,
    doctor_id               BIGINT,
    job_id                  UUID,
    actor_id                BIGINT      NOT NULL,
    date                    DATE        NOT NULL,
    result_summary          TEXT        NOT NULL DEFAULT '',
    machine                 TEXT        NOT NULL DEFAULT '',
    display_snapshot        JSONB       NOT NULL,
    schema_version          SMALLINT    NOT NULL DEFAULT 1,
    change_reason           TEXT        NULL,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_examination_revisions_clinic_exam_version
        UNIQUE (clinic_id, examination_id, version),
    CONSTRAINT ck_examination_revisions_schema_version
        CHECK (schema_version = 1),
    CONSTRAINT ck_examination_revisions_change_reason
        CHECK (change_reason IS NULL OR btrim(change_reason) <> ''),
    CONSTRAINT ck_examination_revisions_kind_status
        CHECK (
            (kind = 'official' AND status = 'confirmed') OR
            (kind = 'working' AND status IN ('pending', 'in_progress', 'result_entered', 'completed'))
        ),
    CONSTRAINT ck_examination_revisions_display_snapshot
        CHECK (
            jsonb_typeof(display_snapshot) = 'object' AND
            COALESCE(jsonb_typeof(display_snapshot -> 'medical_record_no'), '') = 'string' AND
            COALESCE(jsonb_typeof(display_snapshot -> 'pet_name'), '') = 'string' AND
            COALESCE(jsonb_typeof(display_snapshot -> 'medical_record_owner_name'), '') = 'string' AND
            COALESCE(jsonb_typeof(display_snapshot -> 'pet_owner_name'), '') = 'string' AND
            COALESCE(jsonb_typeof(display_snapshot -> 'species_name'), '') = 'string' AND
            COALESCE(jsonb_typeof(display_snapshot -> 'exam_type_name'), '') = 'string' AND
            COALESCE(jsonb_typeof(display_snapshot -> 'doctor_name'), '') = 'string'
        )
);

CREATE TABLE examination_revision_items (
    id                   BIGSERIAL          PRIMARY KEY,
    clinic_id            BIGINT             NOT NULL,
    examination_id       BIGINT             NOT NULL,
    version              BIGINT             NOT NULL CHECK (version >= 1),
    exam_type_field_id   BIGINT,
    name                 TEXT               NOT NULL DEFAULT '',
    inspection_value     TEXT               NOT NULL DEFAULT '',
    normal_value         TEXT               NOT NULL DEFAULT '',
    result               TEXT               NOT NULL DEFAULT '',
    unit                 TEXT               NOT NULL DEFAULT '',
    reference_value      TEXT               NOT NULL DEFAULT '',
    ref_min              DECIMAL(10,4),
    ref_max              DECIMAL(10,4),
    qualitative_min      TEXT,
    qualitative_max      TEXT,
    is_assessed          BOOLEAN            NOT NULL,
    is_abnormal          BOOLEAN            NOT NULL,
    status               exam_result_status NOT NULL,
    sort_order           INTEGER            NOT NULL DEFAULT 0,
    created_at           TIMESTAMPTZ        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT ck_examination_revision_items_reference_range
        CHECK (ref_min IS NULL OR ref_max IS NULL OR ref_min <= ref_max)
);

ALTER TABLE examination_revisions
    ADD CONSTRAINT fk_examination_revisions_exam
        FOREIGN KEY (clinic_id, examination_id)
        REFERENCES exams (clinic_id, id)
        ON DELETE RESTRICT NOT DEFERRABLE;

ALTER TABLE examination_revision_items
    ADD CONSTRAINT fk_examination_revision_items_revision
        FOREIGN KEY (clinic_id, examination_id, version)
        REFERENCES examination_revisions (clinic_id, examination_id, version)
        ON DELETE RESTRICT NOT DEFERRABLE;

ALTER TABLE exams
    ADD CONSTRAINT fk_exams_current_revision
        FOREIGN KEY (clinic_id, id, current_revision_version)
        REFERENCES examination_revisions (clinic_id, examination_id, version)
        ON DELETE RESTRICT NOT DEFERRABLE;

CREATE INDEX idx_examination_revisions_latest_kind
    ON examination_revisions (clinic_id, examination_id, kind, version DESC);

CREATE INDEX idx_examination_revision_items_revision_sort
    ON examination_revision_items (clinic_id, examination_id, version, sort_order, id);

CREATE INDEX idx_exams_current_revision_pointer
    ON exams (clinic_id, id, current_revision_version);

CREATE OR REPLACE FUNCTION app_private.prevent_examination_revision_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'examination_revisions is append-only; UPDATE/DELETE are not allowed'
        USING ERRCODE = 'check_violation';
END;
$$;

CREATE TRIGGER trg_examination_revisions_immutable
    BEFORE UPDATE OR DELETE ON examination_revisions
    FOR EACH ROW
    EXECUTE FUNCTION app_private.prevent_examination_revision_mutation();

CREATE OR REPLACE FUNCTION app_private.prevent_examination_revision_item_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'examination_revision_items is append-only; UPDATE/DELETE are not allowed'
        USING ERRCODE = 'check_violation';
END;
$$;

CREATE TRIGGER trg_examination_revision_items_immutable
    BEFORE UPDATE OR DELETE ON examination_revision_items
    FOR EACH ROW
    EXECUTE FUNCTION app_private.prevent_examination_revision_item_mutation();

CREATE OR REPLACE FUNCTION app_private.enforce_examination_revision_item_insert_version()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
    selected_version BIGINT;
BEGIN
    SELECT current_revision_version
    INTO selected_version
    FROM exams
    WHERE clinic_id = NEW.clinic_id
      AND id = NEW.examination_id
    FOR UPDATE;

    IF NOT FOUND THEN
        RAISE EXCEPTION
            'examination revision item has no matching examination: clinic_id=%, examination_id=%',
            NEW.clinic_id,
            NEW.examination_id
            USING ERRCODE = 'check_violation';
    END IF;

    IF NEW.version <> COALESCE(selected_version, 0) + 1 THEN
        RAISE EXCEPTION
            'examination revision item version must be the next unselected version: selected=%, attempted=%',
            selected_version,
            NEW.version
            USING ERRCODE = 'check_violation';
    END IF;

    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_examination_revision_items_insert_version
    BEFORE INSERT ON examination_revision_items
    FOR EACH ROW
    EXECUTE FUNCTION app_private.enforce_examination_revision_item_insert_version();

SELECT app_private.apply_rls_policy(
    'examination_revisions',
    'tenant_examination_revisions_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);

SELECT app_private.apply_rls_policy(
    'examination_revision_items',
    'tenant_examination_revision_items_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);
