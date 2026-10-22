-- EMR-446/466: clinic-line-time storage foundation (selected contract
-- docs/work/emr/emr-clinic-line-time-selected-contract-20261022.md).
-- A02 per-staff trimming start policy + EMR-492 follow-up anchor
-- (contract §4.1, §4.3).
-- Storage only: no API/service/handler writer is wired here, no defaults or
-- backfills are invented, and no data migration is performed.
-- Apply only through the normal, human-run migration procedure.

-- The new UNIQUE exists only to let the policy's composite FK pin the same
-- staff to its assignment row (id alone is already unique).
ALTER TABLE staff_clinic_assignments
    ADD CONSTRAINT uq_staff_clinic_assignments_id_staff_clinic UNIQUE (id, staff_id, clinic_id);

-- (clinic_id, reservation_type_id, assignment_id) PK. staff_id must match the
-- assignment's staff, enforced by the composite FK. Absent row = unconfigured;
-- explicit empty = weekday times:[]; date overrides replace (not union) the
-- weekly set; deleted_at is the tombstone marker (rows are never physically
-- deleted). Duty/capability/duration/capacity data is not duplicated here.
CREATE TABLE line_staff_start_policies (
    clinic_id           bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    reservation_type_id bigint      NOT NULL,
    assignment_id       bigint      NOT NULL,
    staff_id            bigint      NOT NULL,
    schema_version      text        NOT NULL CHECK (schema_version = 'trimming-registered-starts/v1'),
    version             bigint      NOT NULL DEFAULT 1 CHECK (version > 0),
    weekly_times        jsonb       NOT NULL CHECK (jsonb_typeof(weekly_times) = 'array' AND jsonb_array_length(weekly_times) = 7),
    date_overrides      jsonb       NOT NULL CHECK (jsonb_typeof(date_overrides) = 'array'),
    deleted_at          timestamptz,
    created_at          timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at          timestamptz NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (clinic_id, reservation_type_id, assignment_id),
    FOREIGN KEY (reservation_type_id, clinic_id) REFERENCES reservation_types(id, clinic_id) ON DELETE RESTRICT,
    FOREIGN KEY (assignment_id, staff_id, clinic_id) REFERENCES staff_clinic_assignments(id, staff_id, clinic_id) ON DELETE RESTRICT
);
CREATE INDEX idx_line_staff_start_policies_clinic_staff ON line_staff_start_policies (clinic_id, staff_id);
COMMENT ON TABLE line_staff_start_policies IS 'EMR-446/466 A02 (contract §4.1): per-staff trimming registered start times. Absent row = unconfigured; explicit empty = weekday times:[]; a date override replaces the weekly set for that date (removing it falls back to weekly); deleted_at is the tombstone (reconfigure clears it and bumps version). Duty/capability/duration/capacity are not duplicated; strict JSON validation (weekday 0..6 unique ascending, HH:MM unique ascending times, unique ascending dates) is the server strict decoder responsibility';
COMMENT ON COLUMN line_staff_start_policies.clinic_id IS 'Tenant key and PK member; references clinics(id), ON DELETE RESTRICT';
COMMENT ON COLUMN line_staff_start_policies.reservation_type_id IS 'PK member; composite FK (reservation_type_id, clinic_id) -> reservation_types(id, clinic_id)';
COMMENT ON COLUMN line_staff_start_policies.assignment_id IS 'PK member; composite FK (assignment_id, staff_id, clinic_id) -> staff_clinic_assignments(id, staff_id, clinic_id) pins staff_id to this assignment';
COMMENT ON COLUMN line_staff_start_policies.staff_id IS 'Staff of the assignment; consistency with assignment_id is enforced by the composite FK, not by the app alone';
COMMENT ON COLUMN line_staff_start_policies.schema_version IS 'Storage schema marker, fixed to trimming-registered-starts/v1';
COMMENT ON COLUMN line_staff_start_policies.version IS 'Positive CAS counter starting at 1; maintained by app_private.entity_version_guard(). Tombstoning and reconfigure also advance it once';
COMMENT ON COLUMN line_staff_start_policies.weekly_times IS 'Array of exactly 7 elements, one per weekday 0..6 in ascending order; each element {weekday,times}, times are unique ascending HH:MM 00:00..23:59, empty = explicit no starts';
COMMENT ON COLUMN line_staff_start_policies.date_overrides IS 'Array of {date,times}, strict Gregorian YYYY-MM-DD unique ascending; a present date replaces the weekly set, times:[] means zero starts that day';
COMMENT ON COLUMN line_staff_start_policies.deleted_at IS 'Tombstone marker; rows are never physically deleted (line_staff_start_policy_no_delete). NULL = configured';
COMMENT ON COLUMN line_staff_start_policies.created_at IS 'Row creation timestamp (clock_timestamp)';
COMMENT ON COLUMN line_staff_start_policies.updated_at IS 'Row update timestamp (clock_timestamp)';

CREATE TRIGGER line_staff_start_policies_storage_version
    BEFORE UPDATE ON line_staff_start_policies
    FOR EACH ROW EXECUTE FUNCTION app_private.entity_version_guard();

-- Physical delete is forbidden: DELETE only tombstones (deleted_at + version+1).
-- app_private.contract_history_immutable() has a different message/semantics
-- ('contract history cannot be changed or deleted'), so a dedicated function
-- carries the contract-mandated message (precedent: revisit_intake_reject_mutation).
CREATE FUNCTION app_private.line_staff_start_policy_no_delete()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'line staff start policies are tombstoned, never deleted' USING ERRCODE = '23514';
END; $$;

CREATE TRIGGER line_staff_start_policies_no_delete
    BEFORE DELETE ON line_staff_start_policies
    FOR EACH ROW EXECUTE FUNCTION app_private.line_staff_start_policy_no_delete();

SELECT app_private.apply_rls_policy(
    'line_staff_start_policies'::regclass,
    'line_staff_start_policies_clinic_scope',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)');

-- EMR-492 follow-up anchor (contract §4.3): main has no
-- reservation_master_operation_anchors yet, so the base table from the 492
-- selected contract and the line_start_policy follow-up delta are defined in
-- one CREATE. Written only by the claim winner hook in the same short
-- transaction as the binding; never repaired or backfilled. No FKs to
-- business reservation_types/staffs/assignment rows so the short claim does
-- not disturb lock ordering (contract §4.3): the original relation is
-- validated at preflight/current-authorization/Apply. Whether target_id may
-- be NULL for type/group depends on the intent and is checked by the app
-- (create-family NULL, existing-target family required).
CREATE TABLE reservation_master_operation_anchors (
    operation_binding_id bigint PRIMARY KEY CHECK (operation_binding_id > 0),
    clinic_id            bigint NOT NULL CHECK (clinic_id > 0),
    target_kind          text   NOT NULL CHECK (target_kind IN ('type', 'group', 'line_start_policy')),
    target_id            bigint CHECK (target_id IS NULL OR target_id > 0),
    staff_id             bigint CHECK (staff_id IS NULL OR staff_id > 0),
    assignment_id        bigint CHECK (assignment_id IS NULL OR assignment_id > 0),
    FOREIGN KEY (operation_binding_id, clinic_id) REFERENCES operation_bindings (id, clinic_id) ON DELETE RESTRICT,
    CHECK (
        (target_kind IN ('type', 'group') AND staff_id IS NULL AND assignment_id IS NULL)
        OR (target_kind = 'line_start_policy' AND target_id IS NOT NULL AND staff_id IS NOT NULL AND assignment_id IS NOT NULL)
    )
);
COMMENT ON TABLE reservation_master_operation_anchors IS 'EMR-492 base + §4.3 follow-up: claim-time original target of a reservation-master binding, written only in the claim transaction. Immutable. target_id NULL-ability for type/group depends on the intent and is app-checked; line_start_policy requires target_id (the reservation type), staff_id and assignment_id';
COMMENT ON COLUMN reservation_master_operation_anchors.operation_binding_id IS 'operation_bindings row claimed in the same transaction; PK';
COMMENT ON COLUMN reservation_master_operation_anchors.clinic_id IS 'Authorization-scope clinic copied from the binding';
COMMENT ON COLUMN reservation_master_operation_anchors.target_kind IS 'type | group | line_start_policy';
COMMENT ON COLUMN reservation_master_operation_anchors.target_id IS 'Original target id; NULL-ability for type/group is intent-dependent and app-checked, required (the reservation type id) for line_start_policy';
COMMENT ON COLUMN reservation_master_operation_anchors.staff_id IS 'NULL for type/group; required original staff for line_start_policy. No business FK by design (short claim lock order)';
COMMENT ON COLUMN reservation_master_operation_anchors.assignment_id IS 'NULL for type/group; required original assignment for line_start_policy. No business FK by design (short claim lock order)';

SELECT app_private.apply_rls_policy(
    'reservation_master_operation_anchors'::regclass,
    'reservation_master_operation_anchors_clinic_scope',
    'app_private.has_clinic_access(clinic_id)',
    'app_private.has_clinic_access(clinic_id)');

CREATE TRIGGER reservation_master_operation_anchors_immutable
    BEFORE UPDATE OR DELETE ON reservation_master_operation_anchors
    FOR EACH ROW EXECUTE FUNCTION app_private.contract_history_immutable();
