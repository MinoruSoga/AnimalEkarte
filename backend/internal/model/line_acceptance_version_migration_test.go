package model

// line_acceptance_version_migration_test.go — static checks for
// backend/migrations/009_line_acceptance_version.sql and the new model fields
// (EMR-446/466 A03 LINE acceptance, contract §3.1). NOT_RUN yet in this
// workspace. Shared helpers live in clinic_operating_hours_migration_test.go
// (stack base).

import (
	"reflect"
	"strings"
	"testing"
)

func TestLineAcceptanceVersionMigration(t *testing.T) {
	t.Parallel()

	sql := readMigrationFile(t, "../../migrations/009_line_acceptance_version.sql")

	requiredSnippets := []string{
		"ADD COLUMN acceptance_version bigint NOT NULL DEFAULT 1 CHECK (acceptance_version > 0)",
		"ADD COLUMN reviewed_acceptance_version bigint",
		"ck_line_reservation_reviewed_acceptance_version",
		"CREATE FUNCTION app_private.line_acceptance_version_guard()",
		"CREATE TRIGGER line_reservation_settings_acceptance_guard\n    BEFORE INSERT OR UPDATE ON line_reservation_settings\n    FOR EACH ROW EXECUTE FUNCTION app_private.line_acceptance_version_guard();",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(sql, snippet) {
			t.Fatalf("009 line acceptance migration missing required snippet:\n%s", snippet)
		}
	}
}

func TestLineAcceptanceGuardProtectsOnlyTheTwelveContractColumns(t *testing.T) {
	t.Parallel()

	sql := readMigrationFile(t, "../../migrations/009_line_acceptance_version.sql")

	guard := migrationSection(sql,
		"CREATE FUNCTION app_private.line_acceptance_version_guard()", "END; $$;")
	if guard == "" {
		t.Fatal("009 migration must define app_private.line_acceptance_version_guard()")
	}

	protected := []string{
		"status",
		"closed_weekdays",
		"closed_dates",
		"national_holiday_closed",
		"business_hours",
		"business_hours_by_weekday",
		"break_hours",
		"booking_window_max_days",
		"booking_window_min_days",
		"time_slot_mode",
		"time_slot_interval_minutes",
		"show_no_staff_option",
	}
	for _, col := range protected {
		if !strings.Contains(guard, "NEW."+col) || !strings.Contains(guard, "OLD."+col) {
			t.Fatalf("line_acceptance_version_guard must compare protected column %q (contract §3.1 P)", col)
		}
	}

	// N (display/notification fields), the registration-fields column, and the
	// retired/legacy columns must not participate in the version decision
	// (contract §3.1).
	notProtected := []string{
		"daily_limit",
		"monthly_limit",
		"no_staff_mode",
		"header_text",
		"reservation_notice",
		"cancel_notice",
		"privacy_policy",
		"calendar_months",
		"phone_number",
		"notification_email",
		"request_example",
		"additional_fields",
	}
	for _, col := range notProtected {
		if strings.Contains(guard, "NEW."+col) || strings.Contains(guard, "OLD."+col) {
			t.Fatalf("line_acceptance_version_guard must not compare non-protected column %q (contract §3.1)", col)
		}
	}

	requiredLogic := []string{
		"NEW.acceptance_version := 1;",
		"NEW.reviewed_acceptance_version := OLD.reviewed_acceptance_version;",
		"NEW.acceptance_version := OLD.acceptance_version + CASE WHEN p_changed OR r_changed THEN 1 ELSE 0 END;",
		"reviewed_acceptance_version must equal the resulting acceptance_version",
	}
	for _, snippet := range requiredLogic {
		if !strings.Contains(guard, snippet) {
			t.Fatalf("line_acceptance_version_guard missing required logic:\n%s", snippet)
		}
	}
}

func TestLineReservationSettingAcceptanceColumns(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeOf(LineReservationSetting{})
	av := mustField(t, typ, "AcceptanceVersion")
	if av.Type.Kind() != reflect.Int64 {
		t.Fatalf("AcceptanceVersion must be int64, got %s", av.Type)
	}
	if tag := av.Tag.Get("gorm"); !strings.Contains(tag, "default:1") || !strings.Contains(tag, "not null") {
		t.Fatalf("AcceptanceVersion must be not null with default:1, got %q", tag)
	}
	if jsonTag := av.Tag.Get("json"); jsonTag != "-" {
		t.Fatalf("AcceptanceVersion must not reach the legacy wire (json:\"-\"), got %q", jsonTag)
	}
	rv := mustField(t, typ, "ReviewedAcceptanceVersion")
	if rv.Type.Kind() != reflect.Ptr {
		t.Fatalf("ReviewedAcceptanceVersion must be *int64 (nullable), got %s", rv.Type)
	}
	if jsonTag := rv.Tag.Get("json"); jsonTag != "-" {
		t.Fatalf("ReviewedAcceptanceVersion must not reach the legacy wire (json:\"-\"), got %q", jsonTag)
	}
}

func TestLineAcceptanceContractConstants(t *testing.T) {
	t.Parallel()

	expected := map[string]string{
		"reservation":                LineTimeOwnerKindReservation,
		"line-acceptance/v1":         LineAcceptanceSchemaV1,
		"line-acceptance.initialize": IntentLineAcceptanceInitialize,
		"line-acceptance.confirm":    IntentLineAcceptanceConfirm,
		"line-acceptance.update":     IntentLineAcceptanceUpdate,
	}
	for want, got := range expected {
		if got != want {
			t.Fatalf("contract constant mismatch: want %q got %q", want, got)
		}
	}
}
