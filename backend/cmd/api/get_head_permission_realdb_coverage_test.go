package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// D3 coverage gate: clinic-fixed ∪ (cross-clinic except GET /api/v1/clinics).
// Full set is 178; remaining without realdb-return-data must stay 0.
const (
	d3CompletedRealDBRouteCount = 178
	d3MaxRemainingWithoutRealDB = 0
)

func d3RealDBTargetEntries(t *testing.T) []getHEADInventoryEntry {
	t.Helper()
	var out []getHEADInventoryEntry
	for _, entry := range readGETHEADInventory(t) {
		switch entry.Class {
		case "clinic-fixed":
			out = append(out, entry)
		case "cross-clinic":
			if entry.Path == "/api/v1/clinics" {
				continue
			}
			out = append(out, entry)
		}
	}
	require.NotEmpty(t, out)
	return out
}

func realDBReturnDataTag(verification string) (string, bool) {
	for _, part := range strings.Split(verification, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "realdb-return-data:") {
			return strings.TrimPrefix(part, "realdb-return-data:"), true
		}
	}
	return "", false
}

func TestGETHEADRealDBReturnDataCoverageStaged(t *testing.T) {
	targets := d3RealDBTargetEntries(t)
	require.Equal(t, 178, len(targets), "D3 target set drifted; recompute allowlist")

	var completed []getHEADInventoryEntry
	var remaining []getHEADInventoryEntry
	for _, entry := range targets {
		tag, ok := realDBReturnDataTag(entry.Verification)
		if !ok {
			remaining = append(remaining, entry)
			continue
		}
		file, symbol, ok := strings.Cut(tag, "#")
		require.True(t, ok, "%s %s: expected file#TestSymbol in %s", entry.Method, entry.Path, tag)
		require.True(t, strings.HasPrefix(symbol, "Test"), "%s %s: missing test symbol in %s", entry.Method, entry.Path, tag)
		base := filepath.Base(file)
		require.True(
			t,
			strings.HasPrefix(base, "realdb_") && strings.Contains(base, "isolation_test.go"),
			"%s %s: realdb-return-data must point at realdb_*isolation_test.go, got %s",
			entry.Method,
			entry.Path,
			base,
		)
		require.False(
			t,
			strings.HasSuffix(base, "_selected_clinic_b_grant_a_test.go"),
			"%s %s: mock handler grant test cannot satisfy realdb-return-data (%s)",
			entry.Method,
			entry.Path,
			base,
		)
		abs := filepath.Join("..", "..", file)
		require.FileExists(t, abs, "%s %s: missing realdb evidence file %s", entry.Method, entry.Path, file)
		completed = append(completed, entry)
	}

	require.Equal(t, d3CompletedRealDBRouteCount, len(completed),
		"completed realdb-return-data routes must cover full D3 set")
	require.Equal(t, d3MaxRemainingWithoutRealDB, len(remaining),
		"remaining D3 routes without realdb-return-data must be 0")
}

func TestGETHEADRealDBReturnDataRejectsMockGrantEvidence(t *testing.T) {
	tag, ok := realDBReturnDataTag(
		"realdb-return-data:internal/staff/staff_handler_selected_clinic_b_grant_a_test.go#TestListStaffs_MembershipABGrantASelectedB",
	)
	require.True(t, ok)
	require.Contains(t, tag, "_selected_clinic_b_grant_a_test.go")
}
