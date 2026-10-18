package csvimport

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

type cutoverNamePreflightError struct {
	message string
	cause   error
}

// Error returns only the fixed safe diagnostic message, never the wrapped cause.
func (e *cutoverNamePreflightError) Error() string { return e.message }

// Unwrap returns the preserved underlying cause for errors.Is/As callers.
func (e *cutoverNamePreflightError) Unwrap() error { return e.cause }

const cutoverProcedureNameQuery = `SELECT EXISTS (SELECT 1 FROM public.procedures WHERE clinic_id = $1 AND deleted_at IS NULL AND name = ANY($2::text[]))`

const cutoverVaccineNameQuery = `SELECT EXISTS (SELECT 1 FROM public.vaccines WHERE clinic_id = $1 AND deleted_at IS NULL AND name = ANY($2::text[]))`

// validateCutoverMasterTargetNames is the compatibility wrapper probing one master table's name column with its exact static query.
func validateCutoverMasterTargetNames(ctx context.Context, q cutoverQuerier, bundle CutoverBundle, seeds CutoverSeedIDs, tableName string, nameQuery string) error {
	return validateCutoverTargetUniqueValues(ctx, q, bundle, seeds, cutoverTargetUniqueRule{table: tableName, column: "name", query: nameQuery})
}

// cutoverTargetUniqueContext returns the fixed diagnostic label, preserving the original name-preflight wording only for medicines/procedures/vaccines.
func cutoverTargetUniqueContext(rule cutoverTargetUniqueRule) string {
	if rule.column == "name" && (rule.table == "medicines" || rule.table == "procedures" || rule.table == "vaccines") {
		return fmt.Sprintf("table %s name preflight", rule.table)
	}
	return fmt.Sprintf("table %s column %s unique preflight", rule.table, rule.column)
}

// validateCutoverTargetUniqueValues verifies the complete opened CSV before probing selected keys for seeds.ClinicID. It retains raw selected values and binds batches of at most 256; filtered-out rows still participate in hashing.
func validateCutoverTargetUniqueValues(ctx context.Context, q cutoverQuerier, bundle CutoverBundle, seeds CutoverSeedIDs, rule cutoverTargetUniqueRule) error {
	context := cutoverTargetUniqueContext(rule)
	table, err := cutoverManifestTableByName(bundle.Manifest, rule.table)
	if err != nil {
		return &cutoverNamePreflightError{message: context + ": manifest entry is missing", cause: err}
	}
	var spec CutoverTableSpec
	for _, candidate := range CutoverTableSpecs(bundle.Manifest.profile) {
		if candidate.Name == rule.table {
			spec = candidate
			break
		}
	}
	path := cutoverCSVPath(bundle.SourceDir, table)
	resolved, err := resolveCutoverAccountCSV(path)
	if err != nil {
		return &cutoverNamePreflightError{message: context + ": source csv path is not resolvable", cause: err}
	}
	info, err := os.Lstat(resolved)
	if err != nil {
		return &cutoverNamePreflightError{message: context + ": source csv is not inspectable", cause: err}
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("%s: csv must not be a symbolic link", context)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("%s: csv must be a regular file", context)
	}
	f, err := openStableOwnerOnlyFile(resolved)
	if err != nil {
		return &cutoverNamePreflightError{message: context + ": source csv is not readable", cause: err}
	}
	defer func() { _ = f.Close() }()
	hash := sha256.New()
	reader := csv.NewReader(bufio.NewReader(io.TeeReader(io.LimitReader(f, maxCutoverCSVBytes+1), hash)))
	header, err := reader.Read()
	if err != nil {
		return &cutoverNamePreflightError{message: context + ": source csv header is not readable", cause: err}
	}
	allowPermutation := rehearsalContractDrift(bundle.Manifest, bundle.Provenance.Mode)
	indexes, _, err := cutoverColumnIndexes(spec, header, allowPermutation)
	if err != nil {
		return &cutoverNamePreflightError{message: context + ": source csv header does not match the expected columns", cause: err}
	}
	valueIdx := indexes[rule.column]
	var values []string
	rowCount := int64(0)
	for {
		row, readErr := reader.Read()
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return &cutoverNamePreflightError{message: fmt.Sprintf("%s: source csv row %d is not parseable", context, rowCount+2), cause: readErr}
		}
		rowCount++
		value := row[valueIdx]
		if rule.nonemptyOnly && value == "" {
			continue
		}
		if rule.activeOnly {
			active, ok := cutoverPGBoolean(row[indexes["is_active"]])
			if !ok {
				return fmt.Errorf("table %s column is_active row %d: value is not a PostgreSQL boolean", rule.table, rowCount+1)
			}
			if !active {
				continue
			}
		}
		if rule.tentativeOnly && row[indexes["status"]] != "tentative" {
			continue
		}
		values = append(values, value)
	}
	if rowCount != table.RowCount {
		return fmt.Errorf("%s: source row count mismatch: got %d, want %d", context, rowCount, table.RowCount)
	}
	if digest := hex.EncodeToString(hash.Sum(nil)); !strings.EqualFold(digest, table.SHA256) {
		return fmt.Errorf("%s: source csv sha256 does not match manifest", context)
	}
	for start := 0; start < len(values); start += cutoverMedicineNameBatchSize {
		end := start + cutoverMedicineNameBatchSize
		if end > len(values) {
			end = len(values)
		}
		var exists bool
		if err := q.QueryRow(ctx, rule.query, seeds.ClinicID, values[start:end]).Scan(&exists); err != nil {
			if rule.column == "name" && (rule.table == "medicines" || rule.table == "procedures" || rule.table == "vaccines") {
				return &cutoverNamePreflightError{message: context + ": target name check failed", cause: err}
			}
			return &cutoverNamePreflightError{message: context + ": target check failed", cause: err}
		}
		if exists {
			if rule.column == "name" && (rule.table == "medicines" || rule.table == "procedures" || rule.table == "vaccines") {
				return fmt.Errorf("%s: candidate %s %s already exists in the target clinic (deleted_at IS NULL)", CutoverRefBandOccupied, rule.table, rule.column)
			}
			return fmt.Errorf("%s: table %s column %s unique check failed: candidate value already exists in the target clinic", CutoverRefBandOccupied, rule.table, rule.column)
		}
	}
	return nil
}

// validateCutoverUniqueMasterTargetNames runs every target-unique rule in order, optionally skipping complete legacy band tables.
func validateCutoverUniqueMasterTargetNames(ctx context.Context, q cutoverQuerier, bundle CutoverBundle, seeds CutoverSeedIDs, skipComplete bool) error {
	for _, rule := range cutoverTargetUniqueRules {
		if skipComplete {
			manifestTable, err := cutoverManifestTableByName(bundle.Manifest, rule.table)
			if err != nil {
				return err
			}
			var spec CutoverTableSpec
			for _, candidate := range CutoverTableSpecs(bundle.Manifest.profile) {
				if candidate.Name == rule.table {
					spec = candidate
					break
				}
			}
			state, _, err := inspectCutoverBand(ctx, q, spec, manifestTable, bundle.Manifest.IDBand, seeds)
			if err != nil {
				return err
			}
			if state == cutoverBandComplete {
				continue
			}
		}
		if err := validateCutoverTargetUniqueValues(ctx, q, bundle, seeds, rule); err != nil {
			return err
		}
	}
	return nil
}
