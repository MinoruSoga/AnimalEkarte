package model

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// TestBillDiscountExcluded007Migration_Structure pins the EMR-446/466 (A01)
// storage contract in 007: a nullable is_bill_discount_excluded column with no
// DB default and no backfill on all 7 master tables, plus a storage-managed
// positive bigint version (entity_version_guard trigger) on the 3 tables that
// lacked one. The existing 4 versioned tables must not be re-versioned.
func TestBillDiscountExcluded007Migration_Structure(t *testing.T) {
	t.Parallel()

	sql := readBillDiscountExcluded007Migration(t)

	flagTables := []string{
		"consultations",
		"exam_types",
		"procedures",
		"vaccines",
		"checkup_types",
		"medicines",
		"merchandise_items",
	}
	newlyVersionedTables := []string{"exam_types", "medicines", "merchandise_items"}
	preVersionedTables := []string{"consultations", "procedures", "vaccines", "checkup_types"}

	// Statement checks run on comment-free SQL so header prose cannot satisfy
	// or trip them.
	statements := strings.Split(stripSQLLineComments(sql), ";")

	// (a) nullable flag on all 7 tables, no DEFAULT / NOT NULL / backfill.
	for _, table := range flagTables {
		var stmt string
		for _, s := range statements {
			if strings.Contains(s, "ALTER TABLE "+table) && strings.Contains(s, "is_bill_discount_excluded") {
				stmt = s
				break
			}
		}
		if stmt == "" {
			t.Fatalf("007 migration missing ALTER TABLE %s ADD COLUMN is_bill_discount_excluded", table)
		}
		if !strings.Contains(stmt, "ADD COLUMN is_bill_discount_excluded boolean") {
			t.Fatalf("007 migration %s must declare is_bill_discount_excluded as boolean:\n%s", table, stmt)
		}
		if strings.Contains(stmt, "DEFAULT") {
			t.Fatalf("007 migration %s is_bill_discount_excluded must not carry a DB default (raw NULL is contractual):\n%s", table, stmt)
		}
		if strings.Contains(stmt, "NOT NULL") {
			t.Fatalf("007 migration %s is_bill_discount_excluded must stay nullable:\n%s", table, stmt)
		}
	}

	// No backfill: no UPDATE or INSERT statement. "BEFORE UPDATE ON" inside the
	// trigger DDL is not a data write, so only statement heads are checked.
	for _, s := range statements {
		head := strings.ToUpper(strings.TrimSpace(s))
		if strings.HasPrefix(head, "UPDATE ") || strings.HasPrefix(head, "INSERT ") {
			t.Fatalf("007 migration must not backfill or rewrite existing rows (UPDATE/INSERT is forbidden):\n%s", s)
		}
	}

	// (b) version + CHECK + entity_version_guard trigger on the 3 tables only.
	for _, table := range newlyVersionedTables {
		required := []string{
			"ALTER TABLE " + table,
			"ADD COLUMN version bigint NOT NULL DEFAULT 1",
			"ck_" + table + "_positive_version",
			"CHECK (version > 0)",
			"CREATE TRIGGER " + table + "_storage_version",
			"BEFORE UPDATE ON " + table + " FOR EACH ROW",
			"EXECUTE FUNCTION app_private.entity_version_guard()",
		}
		for _, snippet := range required {
			if !strings.Contains(sql, snippet) {
				t.Fatalf("007 migration %s missing required version snippet:\n%s", table, snippet)
			}
		}
	}
	for _, table := range preVersionedTables {
		for _, s := range statements {
			if strings.Contains(s, "ALTER TABLE "+table) && strings.Contains(s, "version") {
				t.Fatalf("007 migration must not add a second version to already versioned %s:\n%s", table, s)
			}
		}
		if strings.Contains(sql, "CREATE TRIGGER "+table+"_storage_version") {
			t.Fatalf("007 migration must not recreate the existing %s_storage_version trigger", table)
		}
	}
}

// TestBillDiscountExcludedModelTags pins the Go-side storage shape: all 7
// models expose IsBillDiscountExcluded *bool (raw NULL preserved, not
// serialized), and the 3 newly versioned models expose Version int64 with the
// same tags as the existing versioned masters.
func TestBillDiscountExcludedModelTags(t *testing.T) {
	t.Parallel()

	type modelEntry struct {
		name  string
		typ   reflect.Type
		table string
	}
	models := []modelEntry{
		{"Consultation", reflect.TypeOf(Consultation{}), "consultations"},
		{"ExaminationType", reflect.TypeOf(ExaminationType{}), "exam_types"},
		{"Procedure", reflect.TypeOf(Procedure{}), "procedures"},
		{"Vaccine", reflect.TypeOf(Vaccine{}), "vaccines"},
		{"CheckupType", reflect.TypeOf(CheckupType{}), "checkup_types"},
		{"Medicine", reflect.TypeOf(Medicine{}), "medicines"},
		{"MerchandiseItem", reflect.TypeOf(MerchandiseItem{}), "merchandise_items"},
	}

	for _, m := range models {
		field, ok := m.typ.FieldByName("IsBillDiscountExcluded")
		if !ok {
			t.Fatalf("%s (%s) must expose IsBillDiscountExcluded", m.name, m.table)
		}
		if field.Type.Kind() != reflect.Ptr || field.Type.Elem().Kind() != reflect.Bool {
			t.Fatalf("%s.IsBillDiscountExcluded must be *bool to preserve raw NULL, got %s", m.name, field.Type)
		}
		gormTag := field.Tag.Get("gorm")
		for _, tag := range []string{"column:is_bill_discount_excluded", "type:boolean"} {
			if !strings.Contains(gormTag, tag) {
				t.Fatalf("%s.IsBillDiscountExcluded gorm tag must contain %q, got %q", m.name, tag, gormTag)
			}
		}
		if got := field.Tag.Get("json"); got != "-" {
			t.Fatalf("%s.IsBillDiscountExcluded must not be serialized by the model (response DTO owner handles it), got json tag %q", m.name, got)
		}

		version, ok := m.typ.FieldByName("Version")
		if !ok {
			t.Fatalf("%s (%s) must expose Version (all 7 masters are versioned after 007)", m.name, m.table)
		}
		if version.Type.Kind() != reflect.Int64 {
			t.Fatalf("%s.Version must be int64, got %s", m.name, version.Type)
		}
		versionTag := version.Tag.Get("gorm")
		for _, tag := range []string{"column:version", "type:bigint", "not null", "default:1"} {
			if !strings.Contains(versionTag, tag) {
				t.Fatalf("%s.Version gorm tag must contain %q, got %q", m.name, tag, versionTag)
			}
		}
		if got := version.Tag.Get("json"); got != "-" {
			t.Fatalf("%s.Version must not be serialized by the model, got json tag %q", m.name, got)
		}
	}
}

// readBillDiscountExcluded007Migration loads the 007 DDL from
// backend/migrations relative to this package directory.
func readBillDiscountExcluded007Migration(t *testing.T) string {
	t.Helper()
	path := filepath.Join("..", "..", "migrations", "007_bill_discount_excluded_and_master_versions.sql")
	b, err := os.ReadFile(path) //nolint:gosec // fixed relative migration path
	if err != nil {
		t.Fatalf("read 007 migration: %v", err)
	}
	return string(b)
}

// stripSQLLineComments removes `--` line comments so statement checks are not
// fooled by prose inside the migration header.
func stripSQLLineComments(sql string) string {
	var out strings.Builder
	for _, line := range strings.Split(sql, "\n") {
		if i := strings.Index(line, "--"); i >= 0 {
			line = line[:i]
		}
		out.WriteString(line)
		out.WriteByte('\n')
	}
	return out.String()
}
