package main

import (
	"os"
	"strings"
	"testing"
)

// EMR-169: 標準健診区分の手動投入スクリプトが fail-closed 構造を維持することを固定する。
// live_insert_reservation_types_contract_test.go と同型の静的 contract test（DB 非依存）。
func TestLiveInsertStandardCheckupTypesFailsClosed(t *testing.T) {
	contents, err := os.ReadFile("../../migrations/seeds/live_insert_standard_checkup_types.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := string(contents)
	for _, required := range []string{
		"pg_advisory_xact_lock(333052)",
		"RAISE EXCEPTION",
		"WHERE NOT EXISTS",
		"deleted_at IS NULL",
		"'年4健診'", "'バレンタイン健診'", "'5月健診'", "'アドプリット検診'",
		"'歯科検診'", "'皮膚検診'", "'耳検診'", "'眼科検診'",
		"CROSS JOIN desired_checkup_types",
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
