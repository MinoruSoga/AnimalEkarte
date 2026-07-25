ALTER TABLE exam_types
    ADD CONSTRAINT uq_exam_types_id_clinic UNIQUE (id, clinic_id);

ALTER TABLE exam_type_fields
    ADD COLUMN clinic_id bigint;

UPDATE exam_type_fields AS field
SET clinic_id = exam_type.clinic_id
FROM exam_types AS exam_type
WHERE exam_type.id = field.exam_type_id;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM exam_type_fields AS field
        LEFT JOIN exam_types AS exam_type ON exam_type.id = field.exam_type_id
        WHERE field.clinic_id IS NULL
           OR field.clinic_id IS DISTINCT FROM exam_type.clinic_id
    ) THEN
        RAISE EXCEPTION
            'exam_type_fields clinic backfill is incomplete or inconsistent'
            USING ERRCODE = '23514';
    END IF;
END
$$;

ALTER TABLE exam_type_fields
    ALTER COLUMN clinic_id SET NOT NULL,
    ADD CONSTRAINT fk_exam_type_fields_clinic
        FOREIGN KEY (clinic_id) REFERENCES clinics(id) ON DELETE RESTRICT,
    ADD CONSTRAINT uq_exam_type_fields_id_clinic UNIQUE (id, clinic_id);

ALTER TABLE exam_type_fields
    DROP CONSTRAINT exam_type_fields_exam_type_id_fkey;

ALTER TABLE exam_type_fields
    ADD CONSTRAINT fk_exam_type_fields_type_clinic
    FOREIGN KEY (exam_type_id, clinic_id)
    REFERENCES exam_types (id, clinic_id)
    ON DELETE CASCADE;

CREATE INDEX idx_exam_type_fields_clinic_type_sort
    ON exam_type_fields (clinic_id, exam_type_id, sort_order);

CREATE TABLE exam_reference_ranges (
    id                    BIGSERIAL     PRIMARY KEY,
    clinic_id             bigint        NOT NULL,
    exam_type_field_id    bigint        NOT NULL,
    animal_species_id     bigint        NOT NULL,
    ref_min               decimal(10,4),
    ref_max               decimal(10,4),
    qualitative_min       text,
    qualitative_max       text,
    created_at            timestamptz   NOT NULL DEFAULT now(),
    updated_at            timestamptz   NOT NULL DEFAULT now(),
    CONSTRAINT fk_exam_reference_ranges_clinic
        FOREIGN KEY (clinic_id) REFERENCES clinics(id) ON DELETE RESTRICT,
    CONSTRAINT fk_exam_reference_ranges_field_clinic
        FOREIGN KEY (exam_type_field_id, clinic_id)
        REFERENCES exam_type_fields (id, clinic_id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_exam_reference_ranges_animal_species
        FOREIGN KEY (animal_species_id) REFERENCES animal_species(id) ON DELETE RESTRICT,
    CONSTRAINT uq_exam_reference_ranges_clinic_field_species
        UNIQUE (clinic_id, exam_type_field_id, animal_species_id),
    CONSTRAINT chk_exam_reference_ranges_ref_order
        CHECK (ref_min IS NULL OR ref_max IS NULL OR ref_min <= ref_max)
);

CREATE INDEX idx_exam_reference_ranges_animal_species
    ON exam_reference_ranges (animal_species_id);

-- RLS policies
SELECT app_private.apply_rls_policy(
    'exam_type_fields',
    'tenant_exam_type_fields_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);

SELECT app_private.apply_rls_policy(
    'exam_reference_ranges',
    'tenant_exam_reference_ranges_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);
