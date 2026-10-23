package model

// line_staff_start_policy_migration_test.go — static checks for the per-staff
// start policy and reservation anchor storage DDL (the former
// backend/migrations/010_line_staff_start_policy_and_reservation_anchor.sql,
// now a 001_init.sql section)
// and the matching GORM model tags (EMR-446/466 A02 staff start policy +
// EMR-492 follow-up anchor, contract §4.1/§4.3). NOT_RUN yet in this
// workspace. Shared helpers live in clinic_operating_hours_migration_test.go
// (stack base).

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestLineStaffStartPolicyMigration(t *testing.T) {
	t.Parallel()

	sql := readConsolidatedMigrationSection(t, "010_line_staff_start_policy_and_reservation_anchor.sql")

	requiredSnippets := []string{
		"ADD CONSTRAINT uq_staff_clinic_assignments_id_staff_clinic UNIQUE (id, staff_id, clinic_id)",
		"CREATE TABLE line_staff_start_policies",
		"PRIMARY KEY (clinic_id, reservation_type_id, assignment_id)",
		"FOREIGN KEY (reservation_type_id, clinic_id) REFERENCES reservation_types(id, clinic_id) ON DELETE RESTRICT",
		"FOREIGN KEY (assignment_id, staff_id, clinic_id) REFERENCES staff_clinic_assignments(id, staff_id, clinic_id) ON DELETE RESTRICT",
		"schema_version = 'trimming-registered-starts/v1'",
		"jsonb_typeof(weekly_times) = 'array' AND jsonb_array_length(weekly_times) = 7",
		"jsonb_typeof(date_overrides) = 'array'",
		"deleted_at          timestamptz",
		"CREATE INDEX idx_line_staff_start_policies_clinic_staff ON line_staff_start_policies (clinic_id, staff_id)",
		"CREATE TRIGGER line_staff_start_policies_storage_version\n    BEFORE UPDATE ON line_staff_start_policies\n    FOR EACH ROW EXECUTE FUNCTION app_private.entity_version_guard();",
		"CREATE FUNCTION app_private.line_staff_start_policy_no_delete()",
		"CREATE TRIGGER line_staff_start_policies_no_delete\n    BEFORE DELETE ON line_staff_start_policies\n    FOR EACH ROW EXECUTE FUNCTION app_private.line_staff_start_policy_no_delete();",
		"'line_staff_start_policies_clinic_scope'",
		"CREATE TABLE reservation_master_operation_anchors",
		"operation_binding_id bigint PRIMARY KEY CHECK (operation_binding_id > 0)",
		"target_kind IN ('type', 'group', 'line_start_policy')",
		"FOREIGN KEY (operation_binding_id, clinic_id) REFERENCES operation_bindings (id, clinic_id) ON DELETE RESTRICT",
		"target_kind = 'line_start_policy' AND target_id IS NOT NULL AND staff_id IS NOT NULL AND assignment_id IS NOT NULL",
		"'reservation_master_operation_anchors_clinic_scope'",
		"CREATE TRIGGER reservation_master_operation_anchors_immutable\n    BEFORE UPDATE OR DELETE ON reservation_master_operation_anchors\n    FOR EACH ROW EXECUTE FUNCTION app_private.contract_history_immutable();",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(sql, snippet) {
			t.Fatalf("010 staff policy / anchor migration missing required snippet:\n%s", snippet)
		}
	}
}

func TestReservationMasterOperationAnchorsHaveNoBusinessForeignKeys(t *testing.T) {
	t.Parallel()

	sql := readConsolidatedMigrationSection(t, "010_line_staff_start_policy_and_reservation_anchor.sql")

	// Contract §4.3: the claim-time anchor must not lock business
	// reservation_types/staffs/assignment rows; only operation_bindings is
	// referenced. Scope the assertion to the anchor CREATE TABLE statement.
	anchorDDL := migrationSection(sql,
		"CREATE TABLE reservation_master_operation_anchors", ");")
	if anchorDDL == "" {
		t.Fatal("010 migration must define reservation_master_operation_anchors")
	}
	forbidden := []string{
		"REFERENCES reservation_types",
		"REFERENCES staffs",
		"REFERENCES staff_clinic_assignments",
	}
	for _, snippet := range forbidden {
		if strings.Contains(anchorDDL, snippet) {
			t.Fatalf("reservation_master_operation_anchors must not hold business FK %q (contract §4.3):\n%s", snippet, anchorDDL)
		}
	}
	if !strings.Contains(anchorDDL, "REFERENCES operation_bindings") {
		t.Fatalf("anchor must keep the immediate composite FK to operation_bindings:\n%s", anchorDDL)
	}
}

func TestLineStaffStartPolicyModelTags(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeOf(LineStaffStartPolicy{})
	if (LineStaffStartPolicy{}).TableName() != "line_staff_start_policies" {
		t.Fatalf("unexpected TableName: %s", (LineStaffStartPolicy{}).TableName())
	}
	if TrimmingRegisteredStartsSchemaV1 != "trimming-registered-starts/v1" {
		t.Fatalf("unexpected schema marker: %s", TrimmingRegisteredStartsSchemaV1)
	}
	for _, name := range []string{"ClinicID", "ReservationTypeID", "AssignmentID"} {
		if tag := mustField(t, typ, name).Tag.Get("gorm"); !strings.Contains(tag, "primaryKey") || !strings.Contains(tag, "autoIncrement:false") {
			t.Fatalf("%s must be a non-autoIncrement composite PK member, got %q", name, tag)
		}
	}
	if tag := mustField(t, typ, "Version").Tag.Get("gorm"); !strings.Contains(tag, "default:1") {
		t.Fatalf("Version must carry default:1, got %q", tag)
	}
	for _, name := range []string{"WeeklyTimes", "DateOverrides"} {
		if f := mustField(t, typ, name); f.Type != reflect.TypeOf(json.RawMessage(nil)) {
			t.Fatalf("%s must be json.RawMessage, got %s", name, f.Type)
		}
	}
	deletedAt := mustField(t, typ, "DeletedAt")
	if deletedAt.Type != reflect.TypeOf((*time.Time)(nil)) {
		t.Fatalf("DeletedAt must be *time.Time, not gorm.DeletedAt (tombstone is read by the management GET), got %s", deletedAt.Type)
	}
}

func TestLineStartPolicyContractConstants(t *testing.T) {
	t.Parallel()

	expected := map[string]string{
		"line-start-policy.create":      IntentLineStartPolicyCreate,
		"line-start-policy.replace":     IntentLineStartPolicyReplace,
		"line-start-policy.reconfigure": IntentLineStartPolicyReconfigure,
		"line-start-policy.delete":      IntentLineStartPolicyDelete,
		"type":                          AnchorTargetKindType,
		"group":                         AnchorTargetKindGroup,
		"line_start_policy":             AnchorTargetKindLineStartPolicy,
	}
	for want, got := range expected {
		if got != want {
			t.Fatalf("contract constant mismatch: want %q got %q", want, got)
		}
	}
}
