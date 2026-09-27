package main

import (
	"os"
	"strings"
	"testing"
)

// EMR-193: 標準予約区分の手動投入スクリプトが fail-closed 構造を維持することを固定する。
// live_insert_lab_device_clinic2.sql の contract test と同型。
func TestLiveInsertStandardReservationTypesFailsClosed(t *testing.T) {
	contents, err := os.ReadFile("../../migrations/seeds/live_insert_standard_reservation_types.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := string(contents)
	for _, required := range []string{
		"pg_advisory_xact_lock",
		"RAISE EXCEPTION",
		"WHERE NOT EXISTS",
		"deleted_at IS NULL",
		"'診察'", "'お手入れ'", "'ワクチン'", "'健診'",
		"'general'",
		"CROSS JOIN desired_reservation_types",
		"postcondition mismatch",
	} {
		if !strings.Contains(sql, required) {
			t.Errorf("live SQL missing %q", required)
		}
	}
	for _, forbidden := range []string{
		"ON CONFLICT DO NOTHING",
		"clinic_id = 1",
		"clinic_id = 2",
		"clinic_id = 3",
		"clinic_id = 4",
	} {
		if strings.Contains(sql, forbidden) {
			t.Errorf("live SQL contains unstable or clinic-hardcoded pattern %q", forbidden)
		}
	}
}
