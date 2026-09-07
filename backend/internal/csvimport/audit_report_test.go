package csvimport

import "testing"

func TestPathWithinRoot(t *testing.T) {
	if !PathWithinRoot("/migration-reports", "/migration-reports/report.json") {
		t.Fatal("direct child should be inside root")
	}
	for _, candidate := range []string{
		"/migration-reports",
		"/tmp/report.json",
		"report.json",
		"/migration-reports/nested/report.json",
	} {
		if PathWithinRoot("/migration-reports", candidate) {
			t.Fatalf("candidate %q should be outside root", candidate)
		}
	}
}
