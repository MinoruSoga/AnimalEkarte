-- EMR-446/466: persisted claim anchors and lab mapping version storage.
-- EMR-493: original-target anchor for staff.leave_period.* bindings.
-- EMR-297-A: versioned lab mapping storage (heads + immutable revisions) and
-- the original-target recovery anchor for lab_mapping.* bindings.
-- Apply only through the normal, human-run migration procedure.

-- (1) EMR-493 original-target anchor. create keeps leave_period_id NULL and
-- authorizes off the staff root; change requires an existing leave. The
-- binding's owner_kind='staff' and intent pairing
-- (staff.leave_period.create | staff.leave_period.change) is verified by the
-- app at claim/resolve time. Never repaired or backfilled.
CREATE TABLE staff_leave_operation_targets (
    clinic_id       bigint NOT NULL CHECK (clinic_id > 0),
    binding_id      bigint NOT NULL CHECK (binding_id > 0),
    staff_id        bigint NOT NULL CHECK (staff_id > 0),
    leave_period_id bigint CHECK (leave_period_id IS NULL OR leave_period_id > 0),
    PRIMARY KEY (clinic_id, binding_id),
    FOREIGN KEY (binding_id, clinic_id) REFERENCES operation_bindings (id, clinic_id) ON DELETE RESTRICT,
    FOREIGN KEY (staff_id) REFERENCES staffs (id) ON DELETE RESTRICT,
    FOREIGN KEY (leave_period_id, clinic_id) REFERENCES staff_leave_periods (id, clinic_id) ON DELETE RESTRICT
);
COMMENT ON TABLE staff_leave_operation_targets IS 'EMR-493: claim-time original target of a staff.leave_period.* binding. Written only in the claim transaction. Never repaired or backfilled.';
COMMENT ON COLUMN staff_leave_operation_targets.clinic_id IS 'Authorization-scope clinic copied from the binding';
COMMENT ON COLUMN staff_leave_operation_targets.binding_id IS 'operation_bindings row claimed in the same transaction (owner_kind=staff; intent pairing verified by the app at claim/resolve)';
COMMENT ON COLUMN staff_leave_operation_targets.staff_id IS 'Staff root the claim authorizes against; does not prove current employment';
COMMENT ON COLUMN staff_leave_operation_targets.leave_period_id IS 'NULL on create (target not yet numbered); existing leave id is required on change';

SELECT app_private.apply_rls_policy(
    'staff_leave_operation_targets'::regclass,
    'staff_leave_operation_targets_clinic_scope',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)');

CREATE TRIGGER staff_leave_operation_targets_immutable
    BEFORE UPDATE OR DELETE ON staff_leave_operation_targets
    FOR EACH ROW EXECUTE FUNCTION app_private.contract_history_immutable();

-- (2) EMR-297-A current mapping revision per scope. Mutable head row; every
-- settings writer (master PATCH / configuration / spec / draft confirm)
-- increments revision in the same transaction. New readers stay inactive until
-- all writers participate.
CREATE TABLE lab_device_mapping_heads (
    clinic_id     bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    source_type   varchar(32) NOT NULL,
    species_scope varchar(32) NOT NULL,
    revision      bigint      NOT NULL CHECK (revision > 0),
    updated_at    timestamptz NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (clinic_id, source_type, species_scope),
    CHECK (source_type IN ('fuji_nx600', 'fuji_au10v', 'arkray_pu4010', 'idexx_vetlab')),
    CHECK (species_scope IN ('', 'canine', 'feline'))
);
COMMENT ON TABLE lab_device_mapping_heads IS 'EMR-297-A: current mapping revision per (clinic, source_type, species_scope). species_scope '''' is generic. All settings writers increment revision in the same transaction; new readers stay inactive until all writers participate.';
COMMENT ON COLUMN lab_device_mapping_heads.source_type IS 'Same value set as lab_device_item_masters';
COMMENT ON COLUMN lab_device_mapping_heads.species_scope IS 'Normalized scope; '''' = generic, then canine / feline';
COMMENT ON COLUMN lab_device_mapping_heads.revision IS 'Positive counter bumped by every mapping write in the same transaction';

SELECT app_private.apply_rls_policy(
    'lab_device_mapping_heads'::regclass,
    'lab_device_mapping_heads_clinic_scope',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)');

-- (3) EMR-297-A immutable applied-version snapshot. Past revisions cannot be
-- updated or deleted.
CREATE TABLE lab_device_mapping_revisions (
    id                         bigserial   PRIMARY KEY,
    clinic_id                  bigint      NOT NULL,
    source_type                varchar(32) NOT NULL,
    species_scope              varchar(32) NOT NULL,
    revision                   bigint      NOT NULL CHECK (revision > 0),
    snapshot_json              jsonb       NOT NULL CHECK (jsonb_typeof(snapshot_json) = 'object'),
    snapshot_digest            bytea       NOT NULL CHECK (octet_length(snapshot_digest) = 32),
    field_definition_snapshots jsonb       NOT NULL CHECK (jsonb_typeof(field_definition_snapshots) = 'array'),
    confirmed_draft_id         bigint,
    operation_id               bigint,
    recorded_by_staff_id       bigint,
    recorded_at                timestamptz NOT NULL DEFAULT clock_timestamp(),
    UNIQUE (clinic_id, source_type, species_scope, revision),
    UNIQUE (id, clinic_id),
    FOREIGN KEY (clinic_id, source_type, species_scope) REFERENCES lab_device_mapping_heads (clinic_id, source_type, species_scope) ON DELETE RESTRICT,
    FOREIGN KEY (confirmed_draft_id, clinic_id) REFERENCES lab_device_mapping_drafts (id, clinic_id) ON DELETE RESTRICT,
    FOREIGN KEY (operation_id, clinic_id) REFERENCES operation_bindings (id, clinic_id) ON DELETE RESTRICT,
    FOREIGN KEY (recorded_by_staff_id) REFERENCES staffs (id) ON DELETE RESTRICT
);
COMMENT ON TABLE lab_device_mapping_revisions IS 'EMR-297-A: immutable applied mapping version per scope. Rows are never updated or deleted.';
COMMENT ON COLUMN lab_device_mapping_revisions.snapshot_json IS 'Schema "lab-mapping-revision/v1": code-ascending applied mapping rows. Does not embed its own digest';
COMMENT ON COLUMN lab_device_mapping_revisions.snapshot_digest IS 'SHA-256(JCS(snapshot_json))';
COMMENT ON COLUMN lab_device_mapping_revisions.field_definition_snapshots IS 'Array of "lab-field-definition/v1" projections of the adopted field definitions; remains the digest base even when the field has no version';
COMMENT ON COLUMN lab_device_mapping_revisions.confirmed_draft_id IS 'Set only for draft_confirm provenance';
COMMENT ON COLUMN lab_device_mapping_revisions.operation_id IS 'NULL on legacy_baseline; past approvers are not fabricated';
COMMENT ON COLUMN lab_device_mapping_revisions.recorded_by_staff_id IS 'NULL on legacy_baseline; past timestamps are not fabricated';

SELECT app_private.apply_rls_policy(
    'lab_device_mapping_revisions'::regclass,
    'lab_device_mapping_revisions_clinic_scope',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)');

CREATE TRIGGER lab_device_mapping_revisions_immutable
    BEFORE UPDATE OR DELETE ON lab_device_mapping_revisions
    FOR EACH ROW EXECUTE FUNCTION app_private.contract_history_immutable();

-- (4) EMR-297-A original-target recovery anchor for lab_mapping.* bindings.
-- Written in the same transaction as the claim. No FKs to draft/job/master/
-- field rows: the claim never locks rows that may not exist or may still
-- change; Apply validates them.
CREATE TABLE lab_mapping_operation_targets (
    clinic_id           bigint      NOT NULL CHECK (clinic_id > 0),
    binding_id          bigint      NOT NULL CHECK (binding_id > 0),
    target_kind         text        NOT NULL CHECK (target_kind IN ('sample_registration', 'draft', 'job')),
    intent              text        NOT NULL CHECK (intent IN ('lab_mapping.sample_register', 'lab_mapping.confirm', 'lab_mapping.resolve')),
    draft_id            bigint      CHECK (draft_id IS NULL OR draft_id > 0),
    job_id              bigint      CHECK (job_id IS NULL OR job_id > 0),
    source_type         varchar(32),
    species_scope       varchar(32),
    command_json        jsonb       NOT NULL CHECK (jsonb_typeof(command_json) = 'object'),
    command_sha256      bytea       NOT NULL CHECK (octet_length(command_sha256) = 32),
    private_raw_payload bytea,
    PRIMARY KEY (clinic_id, binding_id),
    FOREIGN KEY (binding_id, clinic_id) REFERENCES operation_bindings (id, clinic_id) ON DELETE RESTRICT,
    CHECK (source_type IS NULL OR source_type IN ('fuji_nx600', 'fuji_au10v', 'arkray_pu4010', 'idexx_vetlab')),
    CHECK (species_scope IS NULL OR species_scope IN ('', 'canine', 'feline')),
    CHECK (
        (intent = 'lab_mapping.sample_register' AND target_kind = 'sample_registration'
            AND draft_id IS NULL AND job_id IS NULL
            AND private_raw_payload IS NOT NULL AND octet_length(private_raw_payload) <= 8192)
        OR (intent = 'lab_mapping.confirm' AND target_kind = 'draft'
            AND draft_id IS NOT NULL AND job_id IS NULL
            AND private_raw_payload IS NULL)
        OR (intent = 'lab_mapping.resolve' AND target_kind = 'job'
            AND job_id IS NOT NULL AND draft_id IS NULL
            AND private_raw_payload IS NULL)
    )
);
COMMENT ON TABLE lab_mapping_operation_targets IS 'EMR-297-A: claim-time original target of a lab_mapping.* binding, added in the same transaction as the claim. Immutable.';
COMMENT ON COLUMN lab_mapping_operation_targets.clinic_id IS 'Authorization-scope clinic copied from the binding';
COMMENT ON COLUMN lab_mapping_operation_targets.binding_id IS 'operation_bindings row claimed in the same transaction; actor_ref on the binding is the actor of record and is not duplicated here';
COMMENT ON COLUMN lab_mapping_operation_targets.target_kind IS 'sample_registration | draft | job, paired with intent by the combination CHECK';
COMMENT ON COLUMN lab_mapping_operation_targets.intent IS 'lab_mapping.sample_register | lab_mapping.confirm | lab_mapping.resolve';
COMMENT ON COLUMN lab_mapping_operation_targets.draft_id IS 'Required for lab_mapping.confirm only; no FK (draft may not exist or may change at claim time)';
COMMENT ON COLUMN lab_mapping_operation_targets.job_id IS 'Required for lab_mapping.resolve only; no FK (job rows change)';
COMMENT ON COLUMN lab_mapping_operation_targets.source_type IS 'Immutable original scope resolved at claim time; NULL allowed for sample_register';
COMMENT ON COLUMN lab_mapping_operation_targets.species_scope IS 'Immutable original scope resolved at claim time; NULL allowed for sample_register';
COMMENT ON COLUMN lab_mapping_operation_targets.command_json IS 'Owner-private canonical command (schema lab-mapping-*/v1), not the C06 Content flat map';
COMMENT ON COLUMN lab_mapping_operation_targets.command_sha256 IS 'Same value as command_sha256 in the C06 Content flat map';
COMMENT ON COLUMN lab_mapping_operation_targets.private_raw_payload IS 'Required only for sample_register, max 8KiB; NULL for confirm/resolve';

SELECT app_private.apply_rls_policy(
    'lab_mapping_operation_targets'::regclass,
    'lab_mapping_operation_targets_clinic_scope',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)');

CREATE TRIGGER lab_mapping_operation_targets_immutable
    BEFORE UPDATE OR DELETE ON lab_mapping_operation_targets
    FOR EACH ROW EXECUTE FUNCTION app_private.contract_history_immutable();

-- (5) EMR-297-A draft_version for generation/confirm CAS.
ALTER TABLE lab_device_mapping_drafts
    ADD COLUMN version bigint NOT NULL DEFAULT 1 CHECK (version > 0);
COMMENT ON COLUMN lab_device_mapping_drafts.version IS 'EMR-297-A draft_version: positive CAS counter read by generation transitions and confirm preconditions';
