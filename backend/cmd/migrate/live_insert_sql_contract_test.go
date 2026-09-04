package main

import (
	"os"
	"strings"
	"testing"
)

func TestLiveInsertLabDeviceClinic2UsesSemanticReferencesAndFailsClosed(t *testing.T) {
	contents, err := os.ReadFile("../../migrations/seeds/live_insert_lab_device_clinic2.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := string(contents)
	for _, required := range []string{
		"JOIN exam_type_fields f", "f.clinic_id = d.clinic_id", "f.name = d.field_name",
		"RAISE EXCEPTION", "existing lab_devices row conflicts", "existing lab item row conflicts",
		"pg_advisory_xact_lock", "IS DISTINCT FROM",
	} {
		if !strings.Contains(sql, required) {
			t.Errorf("live SQL missing %q", required)
		}
	}
	for _, forbidden := range []string{
		"ON CONFLICT DO NOTHING",
		"53::bigint",
		"exam_type_id, is_active, sort_order, created_at",
		"device items resolve to multiple exam types",
		"HAVING count(DISTINCT exam_type_id) <> 1",
	} {
		if strings.Contains(sql, forbidden) {
			t.Errorf("live SQL contains unstable or silent-conflict pattern %q", forbidden)
		}
	}
}
