package csvimport

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var tables = []string{"owners", "pets", "medical_records", "exams", "exam_results", "billings", "billing_items"}

type legacyImportTransaction interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	QueryRow(context.Context, string, ...any) pgx.Row
	CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error)
	Commit(context.Context) error
	Rollback(context.Context) error
}

// Import replaces the old demo's owner/clinical/accounting graph in a disposable
// target database. expectedDatabase is the required current_database() name
// (TRM-07 / U-X03-CSVIMPORT-GUARD); empty or mismatched names fail closed before
// any DELETE. The source directory is expected to be a read-only mount.
func Import(ctx context.Context, pool *pgxpool.Pool, sourceDir string, expectedDatabase string, clinicID, speciesID, examTypeID int64) (map[string]int64, error) {
	return importWithBegin(ctx, func(ctx context.Context) (legacyImportTransaction, error) {
		return pool.Begin(ctx)
	}, sourceDir, expectedDatabase, clinicID, speciesID, examTypeID)
}

func importWithBegin(
	ctx context.Context,
	begin func(context.Context) (legacyImportTransaction, error),
	sourceDir string,
	expectedDatabase string,
	clinicID int64,
	speciesID int64,
	examTypeID int64,
) (map[string]int64, error) {
	if sourceDir == "" || !filepath.IsAbs(sourceDir) {
		return nil, fmt.Errorf("source directory must be an absolute path")
	}
	if strings.TrimSpace(expectedDatabase) == "" {
		return nil, fmt.Errorf("expected disposable database name is required")
	}
	tx, err := begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin import tx: target database unavailable")
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := requireDisposableTargetDatabase(ctx, tx, expectedDatabase); err != nil {
		return nil, err
	}
	if err := deleteDemoGraph(ctx, tx); err != nil {
		return nil, err
	}
	counts := make(map[string]int64, len(tables))
	for _, table := range tables {
		count, err := importTable(ctx, tx, filepath.Join(sourceDir, table+".csv"), table, clinicID, speciesID, examTypeID)
		if err != nil {
			return nil, fmt.Errorf("import %s: %w", table, err)
		}
		counts[table] = count
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit import tx: target database rejected transaction")
	}
	return counts, nil
}

// requireDisposableTargetDatabase fail-closes when the open transaction is not
// connected to the caller-attested disposable database name. Runs before any
// unscoped DELETE in Import (TRM-07).
func requireDisposableTargetDatabase(ctx context.Context, tx legacyImportTransaction, expectedDatabase string) error {
	expected := strings.TrimSpace(expectedDatabase)
	if expected == "" {
		return fmt.Errorf("expected disposable database name is required")
	}
	var actual string
	if err := tx.QueryRow(ctx, `SELECT current_database()`).Scan(&actual); err != nil {
		return fmt.Errorf("identify target database: target database rejected operation")
	}
	if actual != expected {
		return fmt.Errorf("target database identity mismatch: refusing destructive import")
	}
	return nil
}

func deleteDemoGraph(ctx context.Context, tx legacyImportTransaction) error {
	// Reverse FK order for all non-master rows owned by 003_demo. This keeps
	// clinics, accounts, staffs, and master catalog rows available for imports.
	for _, table := range []string{
		"billing_items", "payments", "payment_splits", "billing_refunds", "billing_confirmations", "billings",
		"estimate_items", "estimates", "exam_results", "medical_record_images", "treatments", "treatment_plans", "clinical_plans",
		"inquiries", "checkups", "vaccinations", "exams", "prescriptions", "medical_record_addenda",
		"care_logs", "staff_notes", "care_plan_items", "vital_records", "daily_records",
		"appointment_trimming_details", "appointment_trimming_options", "appointments", "hospitalizations",
		"pet_chronic_conditions", "lstep_delivery_trigger_log", "lstep_tag_cache", "shared_files", "line_send_logs", "line_customers", "medical_records", "pets", "owners",
	} {
		if _, err := tx.Exec(ctx, `DELETE FROM "`+table+`"`); err != nil {
			return fmt.Errorf("delete %s: target database rejected operation", table)
		}
	}
	return nil
}

type spec struct {
	columns []string
	values  func([]string, int64, int64, int64) ([]any, error)
}

func importTable(ctx context.Context, tx legacyImportTransaction, path, table string, clinicID, speciesID, examTypeID int64) (int64, error) {
	f, err := os.Open(path) //nolint:gosec // operator-controlled absolute CSV mount; validated IsAbs above
	if err != nil {
		return 0, fmt.Errorf("open %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()
	r := csv.NewReader(f)
	r.ReuseRecord = true
	header, err := r.Read()
	if err != nil {
		return 0, fmt.Errorf("read csv header: %w", err)
	}
	idx := make(map[string]int, len(header))
	for i, col := range header {
		idx[col] = i
	}
	s, err := tableSpec(table)
	if err != nil {
		return 0, err
	}
	for _, col := range s.columns {
		if _, ok := idx[col]; !ok && col != "clinic_id" && col != "animal_species_id" && col != "exam_type_id" {
			return 0, fmt.Errorf("missing source column %s", col)
		}
	}
	copyRows := make([][]any, 0, 1024)
	var count int64
	flush := func() error {
		if len(copyRows) == 0 {
			return nil
		}
		_, e := tx.CopyFrom(ctx, pgx.Identifier{table}, s.columns, pgx.CopyFromRows(copyRows))
		copyRows = nil
		if e != nil {
			return fmt.Errorf("copy %s: target database rejected row batch", table)
		}
		return nil
	}
	for {
		row, e := r.Read()
		if errors.Is(e, io.EOF) {
			break
		}
		if e != nil {
			return 0, fmt.Errorf("read csv: %w", e)
		}
		vals, e := s.values(row, clinicID, speciesID, examTypeID)
		if e != nil {
			return 0, fmt.Errorf("map row: %w", e)
		}
		copyRows = append(copyRows, vals)
		count++
		if len(copyRows) == 1024 {
			if e := flush(); e != nil {
				return 0, e
			}
		}
	}
	if err := flush(); err != nil {
		return 0, err
	}
	return count, nil
}

func tableSpec(table string) (spec, error) {
	// Source and target overlap in the same order for these fields; omitted target
	// columns deliberately retain database defaults (timestamps, nullable staff IDs,
	// and demo-only relations).
	switch table {
	case "owners":
		return mapped([]string{"id", "clinic_id", "name", "birth_date", "company", "postal_code", "address1", "address2", "home_postal_code", "home_address1", "home_address2", "phone", "company_phone", "email", "remarks", "is_dangerous", "discount_rate", "membership_type", "dm_preference"}, []string{"id", "clinic_id", "name", "birth_date", "company", "postal_code", "address1", "address2", "home_postal_code", "home_address1", "home_address2", "phone", "company_phone", "email", "remarks", "is_dangerous", "discount_rate", "membership_type", "dm_preference"}, nil), nil
	case "pets":
		return mapped([]string{"id", "clinic_id", "owner_id", "pet_number", "name", "animal_species_id", "gender", "status", "birth_date", "breed", "color", "weight", "neutered_date", "food", "remarks", "deceased_at"}, nil, nil), nil
	case "medical_records":
		return mapped([]string{"id", "clinic_id", "record_no", "date", "owner_id", "pet_id", "status", "visit_type"}, nil, nil), nil
	case "exams":
		return mapped([]string{"id", "clinic_id", "medical_record_id", "pet_id", "date", "exam_type_id", "result_summary"}, nil, nil), nil
	case "exam_results":
		return mapped([]string{"id", "exam_id", "name", "inspection_value", "normal_value", "sort_order"}, nil, nil), nil
	case "billings":
		return mapped([]string{"id", "clinic_id", "medical_record_id", "owner_id", "pet_id", "total_amount", "status", "scheduled_date"}, nil, nil), nil
	case "billing_items":
		return mapped([]string{"id", "billing_id", "category", "name", "unit_price", "quantity", "tax_type", "is_insurance_applicable", "sort_order"}, nil, nil), nil
	default:
		return spec{}, fmt.Errorf("unsupported table %s", table)
	}
}

func mapped(src, _ []string, _ any) spec {
	return spec{columns: src, values: func(row []string, clinic, species, exam int64) ([]any, error) {
		vals := make([]any, len(src))
		for i := range src {
			col, v := src[i], row[i]
			switch col {
			case "clinic_id":
				v = strconv.FormatInt(clinic, 10)
			case "animal_species_id":
				if v == "" || strings.HasPrefix(v, "{{FALLBACK_") {
					v = strconv.FormatInt(species, 10)
				}
			case "exam_type_id":
				v = strconv.FormatInt(exam, 10)
			}
			if strings.TrimSpace(v) == "" && nullableEmpty(col) {
				vals[i] = nil
				continue
			}
			x, e := typedValue(col, v)
			if e != nil {
				return nil, fmt.Errorf("%s: %w", col, e)
			}
			vals[i] = x
		}
		return vals, nil
	}}
}
func nullableEmpty(col string) bool {
	switch col {
	case "birth_date", "neutered_date", "deceased_at", "medical_record_id", "owner_id", "pet_id", "weight", "visit_type", "dm_preference", "completed_at":
		return true
	}
	return false
}
func typedValue(col, v string) (any, error) {
	if v == "" {
		return "", nil
	}
	for _, c := range []string{"id", "clinic_id", "owner_id", "pet_id", "medical_record_id", "exam_id", "billing_id", "animal_species_id", "exam_type_id", "total_amount", "unit_price", "sort_order"} {
		if col == c {
			n, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("parse %s as integer", col)
			}
			return n, nil
		}
	}
	for _, c := range []string{"weight", "quantity", "discount_rate"} {
		if col == c {
			n, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return nil, fmt.Errorf("parse %s as decimal", col)
			}
			return n, nil
		}
	}
	for _, c := range []string{"is_dangerous", "is_insurance_applicable"} {
		if col == c {
			b, err := strconv.ParseBool(v)
			if err != nil {
				return nil, fmt.Errorf("parse %s as boolean", col)
			}
			return b, nil
		}
	}
	return v, nil
}
