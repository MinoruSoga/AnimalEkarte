package repository

// billings_hospitalization_unique_migration_test.go — 004 partial UNIQUE の静的ゲート。
//
// 退院会計の二重永続化を防ぐ idx_billings_hospitalization_id_unique が、連番 migration に
// 正しい述語（hospitalization_id IS NOT NULL AND deleted_at IS NULL）で存在するかを固定する。
// DB 不要（os.ReadFile）。適用可否・重複データ検出は別（適用前手動 SQL）。

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const billingsHospitalizationUniqueMigration = "004_add_billings_hospitalization_id_unique_index.sql"

func TestBillingsHospitalizationUniqueMigration_PartialUniqueIndexSQL(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(migrationsDir, billingsHospitalizationUniqueMigration))
	if err != nil {
		t.Fatalf("read %s: %v", billingsHospitalizationUniqueMigration, err)
	}
	// Strip -- line comments so header docs cannot false-pass predicate assertions.
	ddl := stripSQLLineComments(string(raw))
	normalized := strings.Join(strings.Fields(ddl), " ")

	if !strings.Contains(normalized, "CREATE UNIQUE INDEX idx_billings_hospitalization_id_unique") {
		t.Error("expected CREATE UNIQUE INDEX idx_billings_hospitalization_id_unique")
	}
	if !strings.Contains(normalized, "ON billings(hospitalization_id)") {
		t.Error("expected ON billings(hospitalization_id)")
	}
	if !strings.Contains(normalized, "hospitalization_id IS NOT NULL") {
		t.Error("expected WHERE predicate hospitalization_id IS NOT NULL")
	}
	if !strings.Contains(normalized, "deleted_at IS NULL") {
		t.Error("expected WHERE predicate deleted_at IS NULL")
	}
	if strings.Contains(strings.ToUpper(ddl), "ON DELETE CASCADE") {
		t.Error("004 must not introduce ON DELETE CASCADE")
	}
}

func stripSQLLineComments(s string) string {
	var b strings.Builder
	for _, line := range strings.Split(s, "\n") {
		if i := strings.Index(line, "--"); i >= 0 {
			line = line[:i]
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String()
}
