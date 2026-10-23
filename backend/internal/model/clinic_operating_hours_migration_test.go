package model

// clinic_operating_hours_migration_test.go — static checks for the common
// clinic operating-hours storage DDL (the former
// backend/migrations/008_clinic_operating_hours.sql, now a 001_init.sql
// section) and the matching GORM
// model tags (EMR-446/466 A03 common clinic operating hours, contract §2.1).
// NOT_RUN yet in this workspace; same static-migration-test form as
// rls_migration_test.go. Shared helpers (migrationSection, mustField) live in
// this earliest stack file; ③/④ static tests reuse them.

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// migrationSection extracts one statement span starting at startMarker up to
// (but not including) endMarker, or to EOF when endMarker is empty. Used to
// scope presence/absence assertions to a single statement.
func migrationSection(sql, startMarker, endMarker string) string {
	start := strings.Index(sql, startMarker)
	if start < 0 {
		return ""
	}
	rest := sql[start:]
	if endMarker == "" {
		return rest
	}
	if end := strings.Index(rest, endMarker); end >= 0 {
		return rest[:end]
	}
	return rest
}

// mustField returns the named struct field of typ, or fails the test.
func mustField(t *testing.T, typ reflect.Type, name string) reflect.StructField {
	t.Helper()
	f, ok := typ.FieldByName(name)
	if !ok {
		t.Fatalf("model %s has no field %s", typ.Name(), name)
	}
	return f
}

func TestClinicOperatingHoursMigration(t *testing.T) {
	t.Parallel()

	sql := readConsolidatedMigrationSection(t, "008_clinic_operating_hours.sql")

	requiredSnippets := []string{
		"CREATE TABLE clinic_operating_hours",
		"clinic_id               bigint      PRIMARY KEY REFERENCES clinics(id) ON DELETE RESTRICT",
		"schema_version = 'clinic-operating-hours/v1'",
		"version                 bigint      NOT NULL DEFAULT 1 CHECK (version > 0)",
		"jsonb_typeof(weekly_hours) = 'array' AND jsonb_array_length(weekly_hours) = 7",
		"jsonb_typeof(closed_dates) = 'array'",
		"national_holiday_closed boolean     NOT NULL",
		"CREATE TRIGGER clinic_operating_hours_storage_version\n    BEFORE UPDATE ON clinic_operating_hours\n    FOR EACH ROW EXECUTE FUNCTION app_private.entity_version_guard();",
		"'clinic_operating_hours'::regclass",
		"'clinic_operating_hours_clinic_scope'",
		"'app_private.has_clinic_access(clinic_id)'",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(sql, snippet) {
			t.Fatalf("008 clinic_operating_hours migration missing required snippet:\n%s", snippet)
		}
	}

	// The settings fields must carry no DEFAULT (first write fully explicit).
	createDDL := migrationSection(sql, "CREATE TABLE clinic_operating_hours", ");")
	if createDDL == "" {
		t.Fatal("008 must define clinic_operating_hours")
	}
	for _, col := range []string{"weekly_hours", "closed_dates", "national_holiday_closed"} {
		seg := migrationSection(createDDL, col, ",")
		if strings.Contains(seg, "DEFAULT") {
			t.Fatalf("%s must not carry a DEFAULT (contract §2.1): %s", col, seg)
		}
	}
}

func TestClinicOperatingHoursModelTags(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeOf(ClinicOperatingHours{})
	if (ClinicOperatingHours{}).TableName() != "clinic_operating_hours" {
		t.Fatalf("unexpected TableName: %s", (ClinicOperatingHours{}).TableName())
	}
	if ClinicOperatingHoursSchemaV1 != "clinic-operating-hours/v1" {
		t.Fatalf("unexpected schema marker: %s", ClinicOperatingHoursSchemaV1)
	}
	if tag := mustField(t, typ, "ClinicID").Tag.Get("gorm"); !strings.Contains(tag, "primaryKey") || !strings.Contains(tag, "autoIncrement:false") {
		t.Fatalf("ClinicID must be a non-autoIncrement primary key, got %q", tag)
	}
	if tag := mustField(t, typ, "Version").Tag.Get("gorm"); !strings.Contains(tag, "default:1") {
		t.Fatalf("Version must carry default:1, got %q", tag)
	}
	for _, name := range []string{"WeeklyHours", "ClosedDates"} {
		if f := mustField(t, typ, name); f.Type != reflect.TypeOf(json.RawMessage(nil)) {
			t.Fatalf("%s must be json.RawMessage, got %s", name, f.Type)
		}
	}
	if f := mustField(t, typ, "NationalHolidayClosed"); f.Type.Kind() != reflect.Bool {
		t.Fatalf("NationalHolidayClosed must be a non-pointer bool (explicit, no default), got %s", f.Type)
	}
}

func TestClinicOperatingHoursContractConstants(t *testing.T) {
	t.Parallel()

	expected := map[string]string{
		"clinic":                                            LineTimeOwnerKindClinic,
		"clinic.operating-hours.initialize":                 IntentClinicOperatingHoursInitialize,
		"clinic.operating-hours.update":                     IntentClinicOperatingHoursUpdate,
		"clinic.operating-hours.holiday-set-from-shifts":    IntentClinicOperatingHoursHolidaySetFromShifts,
		"clinic.operating-hours.holiday-unset-from-shifts":  IntentClinicOperatingHoursHolidayUnsetFromShifts,
		"clinic.operating-hours.holiday-set-from-closing":   IntentClinicOperatingHoursHolidaySetFromClosing,
		"clinic.operating-hours.holiday-unset-from-closing": IntentClinicOperatingHoursHolidayUnsetFromClosing,
	}
	for want, got := range expected {
		if got != want {
			t.Fatalf("contract constant mismatch: want %q got %q", want, got)
		}
	}
}
