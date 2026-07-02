package main

import (
	"log/slog"
	"strings"
	"testing"
)

// TestPrintPlanDoesNotPanicForDryRunAndApply exercises both the dry-run and
// apply mode-label branches of printPlan with a representative plan. printPlan
// only logs a pre-computed plan (no DB access), so it is safe to unit test
// directly.
func TestPrintPlanDoesNotPanicForDryRunAndApply(t *testing.T) {
	plan := importPlan{Tables: []tablePlan{
		{
			Table:            "owners",
			Importable:       10,
			SkippedByStatus:  map[string]int64{"needs_review": 2, "blocked": 1},
			SkippedParentGap: 0,
			DeleteFromTarget: 5,
		},
		{
			Table:            "billing_items",
			Importable:       20,
			SkippedByStatus:  map[string]int64{},
			SkippedParentGap: 3,
			DeleteFromTarget: 8,
		},
	}}

	logger := newTestLogger()

	t.Run("dry-run mode", func(t *testing.T) {
		printPlan(logger, plan, options{apply: false})
	})
	t.Run("apply mode", func(t *testing.T) {
		printPlan(logger, plan, options{apply: true, confirmLocal: true})
	})
}

// TestPrintPlanModeLabel pins the human-readable mode string surfaced in the
// log line, since operators rely on it to confirm dry-run vs destructive apply
// before re-running with --apply.
func TestPrintPlanModeLabel(t *testing.T) {
	var buf strings.Builder
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	plan := importPlan{Tables: []tablePlan{
		{Table: "owners", Importable: 1, SkippedByStatus: map[string]int64{}},
	}}

	printPlan(logger, plan, options{apply: false})
	if !strings.Contains(buf.String(), "DRY-RUN") {
		t.Fatalf("dry-run printPlan output must mention DRY-RUN, got: %s", buf.String())
	}

	buf.Reset()
	printPlan(logger, plan, options{apply: true})
	if !strings.Contains(buf.String(), "APPLY") {
		t.Fatalf("apply printPlan output must mention APPLY, got: %s", buf.String())
	}
}
