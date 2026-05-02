-- FEAT-379: per-clinic code-to-tag mapping table for health/prevention/product sync
CREATE TABLE IF NOT EXISTS lstep_tag_code_mappings (
    id            bigserial    PRIMARY KEY,
    clinic_id     bigint       NOT NULL REFERENCES clinics(id),
    tag_name      text         NOT NULL,
    code_type     text         NOT NULL,
    codes         text[]       NOT NULL DEFAULT '{}',
    species_scope text,
    age_min       int,
    deleted_at    timestamptz,
    created_at    timestamptz  NOT NULL DEFAULT now(),
    updated_at    timestamptz  NOT NULL DEFAULT now()
);

-- (clinic_id, tag_name, code_type) must be unique among active rows
CREATE UNIQUE INDEX IF NOT EXISTS idx_lstep_tag_code_mappings_unique
    ON lstep_tag_code_mappings(clinic_id, tag_name, code_type)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_lstep_tag_code_mappings_clinic_tag
    ON lstep_tag_code_mappings(clinic_id, tag_name)
    WHERE deleted_at IS NULL;
