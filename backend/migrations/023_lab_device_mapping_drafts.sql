-- EMR-297: one lab result becomes an unconfirmed item-mapping draft.
-- A person confirms it once. Later imports use only that confirmed map.
--
-- Species-specific rows live in lab_device_mapping_specs.
-- Device-level confirms keep the existing lab_device_item_masters unique key
-- (clinic_id, source_type, device_item_code). Do not add species to that key.
--
-- Do not edit 001_init.sql. User must run: make migrate

ALTER TABLE lab_import_jobs
    ADD COLUMN species varchar(32) NOT NULL DEFAULT '';

COMMENT ON COLUMN lab_import_jobs.species IS
    'Empty, canine, or feline. Empty imports use device-level masters only.';

CREATE TABLE lab_device_mapping_drafts (
    id bigserial PRIMARY KEY,
    clinic_id bigint NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    source_type varchar(32) NOT NULL DEFAULT '',
    species varchar(32) NOT NULL DEFAULT '',
    status varchar(32) NOT NULL,
    raw_payload bytea NOT NULL,
    raw_fingerprint varchar(64) NOT NULL DEFAULT '',
    error_message text NOT NULL DEFAULT '',
    confirmed_at timestamptz,
    confirmed_by_staff_id bigint,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    CONSTRAINT uq_lab_device_mapping_drafts_id_clinic UNIQUE (id, clinic_id),
    CONSTRAINT chk_lab_device_mapping_drafts_source_type
        CHECK (source_type IN ('', 'fuji_nx600', 'fuji_au10v', 'arkray_pu4010', 'idexx_vetlab')),
    CONSTRAINT chk_lab_device_mapping_drafts_species
        CHECK (species IN ('', 'canine', 'feline')),
    CONSTRAINT chk_lab_device_mapping_drafts_status
        CHECK (status IN ('unconfirmed', 'confirmed', 'failed'))
);

CREATE INDEX idx_lab_device_mapping_drafts_clinic_source
    ON lab_device_mapping_drafts (clinic_id, source_type, species, created_at);

COMMENT ON TABLE lab_device_mapping_drafts IS
    'One device result kept until a person confirms the item map. Not a chart write.';

CREATE TABLE lab_device_mapping_draft_items (
    id bigserial PRIMARY KEY,
    clinic_id bigint NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    draft_id bigint NOT NULL,
    device_item_code varchar(64) NOT NULL DEFAULT '',
    value_raw text NOT NULL DEFAULT '',
    device_unit varchar(64) NOT NULL DEFAULT '',
    chart_unit varchar(32) NOT NULL DEFAULT '',
    unit_mismatch boolean NOT NULL DEFAULT false,
    exam_type_field_id bigint,
    proposed_field_name varchar(255) NOT NULL DEFAULT '',
    item_source varchar(32) NOT NULL,
    llm_request jsonb,
    llm_response jsonb,
    llm_error varchar(64) NOT NULL DEFAULT '',
    sort_order integer NOT NULL DEFAULT 0,
    confirmed boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    CONSTRAINT fk_lab_device_mapping_draft_items_draft_clinic
        FOREIGN KEY (draft_id, clinic_id)
        REFERENCES lab_device_mapping_drafts (id, clinic_id)
        ON DELETE RESTRICT,
    CONSTRAINT chk_lab_device_mapping_draft_items_source
        CHECK (item_source IN ('dictionary', 'llm', 'blank', 'comment'))
);

ALTER TABLE lab_device_mapping_draft_items
    ADD CONSTRAINT fk_lab_device_mapping_draft_items_field_clinic
    FOREIGN KEY (exam_type_field_id, clinic_id)
    REFERENCES exam_type_fields (id, clinic_id)
    ON DELETE RESTRICT;

CREATE INDEX idx_lab_device_mapping_draft_items_clinic_draft
    ON lab_device_mapping_draft_items (clinic_id, draft_id, sort_order);

COMMENT ON COLUMN lab_device_mapping_draft_items.value_raw IS
    'Device text kept in the app database. Not sent to the LLM and not returned by the API.';
COMMENT ON COLUMN lab_device_mapping_draft_items.llm_request IS
    'LLM request body for this item only. Null when the dictionary resolved it. Never stores the API key.';

CREATE TABLE lab_device_mapping_specs (
    id bigserial PRIMARY KEY,
    clinic_id bigint NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    source_type varchar(32) NOT NULL,
    species varchar(32) NOT NULL,
    device_item_code varchar(64) NOT NULL,
    unit varchar(32) NOT NULL DEFAULT '',
    value_shape varchar(32) NOT NULL,
    exam_type_field_id bigint NOT NULL,
    confirmed_draft_id bigint NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    CONSTRAINT uq_lab_device_mapping_specs
        UNIQUE (clinic_id, source_type, species, device_item_code),
    CONSTRAINT chk_lab_device_mapping_specs_species
        CHECK (species IN ('canine', 'feline')),
    CONSTRAINT chk_lab_device_mapping_specs_source_type
        CHECK (source_type IN ('fuji_nx600', 'fuji_au10v', 'arkray_pu4010', 'idexx_vetlab')),
    CONSTRAINT chk_lab_device_mapping_specs_value_shape
        CHECK (value_shape IN ('numeric', 'inequality', 'qual_and_num', 'dash', 'text')),
    CONSTRAINT fk_lab_device_mapping_specs_draft_clinic
        FOREIGN KEY (confirmed_draft_id, clinic_id)
        REFERENCES lab_device_mapping_drafts (id, clinic_id)
        ON DELETE RESTRICT
);

ALTER TABLE lab_device_mapping_specs
    ADD CONSTRAINT fk_lab_device_mapping_specs_field_clinic
    FOREIGN KEY (exam_type_field_id, clinic_id)
    REFERENCES exam_type_fields (id, clinic_id)
    ON DELETE RESTRICT;

COMMENT ON TABLE lab_device_mapping_specs IS
    'Confirmed device+species item map. Import overlays these onto device-level masters. Species-empty samples do not read this table.';

CREATE TABLE lab_device_mapping_generation (
    clinic_id bigint PRIMARY KEY REFERENCES clinics(id) ON DELETE RESTRICT,
    stopped boolean NOT NULL DEFAULT false,
    updated_at timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE lab_device_mapping_generation IS
    'Per-clinic switch for new mapping drafts. Import of an already confirmed map ignores this row.';

SELECT app_private.apply_rls_policy(
    'lab_device_mapping_drafts',
    'tenant_clinic_id_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);

SELECT app_private.apply_rls_policy(
    'lab_device_mapping_draft_items',
    'tenant_clinic_id_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);

SELECT app_private.apply_rls_policy(
    'lab_device_mapping_specs',
    'tenant_clinic_id_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);

SELECT app_private.apply_rls_policy(
    'lab_device_mapping_generation',
    'tenant_clinic_id_isolation',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)'
);
