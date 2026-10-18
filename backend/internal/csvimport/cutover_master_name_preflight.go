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

func (e *cutoverNamePreflightError) Error() string { return e.message }

func (e *cutoverNamePreflightError) Unwrap() error { return e.cause }

const cutoverProcedureNameQuery = `SELECT EXISTS (SELECT 1 FROM public.procedures WHERE clinic_id = $1 AND deleted_at IS NULL AND name = ANY($2::text[]))`

const cutoverVaccineNameQuery = `SELECT EXISTS (SELECT 1 FROM public.vaccines WHERE clinic_id = $1 AND deleted_at IS NULL AND name = ANY($2::text[]))`

func validateCutoverMasterTargetNames(ctx context.Context, q cutoverQuerier, bundle CutoverBundle, seeds CutoverSeedIDs, tableName string, nameQuery string) error {
	table, err := cutoverManifestTableByName(bundle.Manifest, tableName)
	if err != nil {
		return &cutoverNamePreflightError{message: fmt.Sprintf("table %s name preflight: manifest entry is missing", tableName), cause: err}
	}
	var spec CutoverTableSpec
	for _, candidate := range CutoverTableSpecs(bundle.Manifest.profile) {
		if candidate.Name == tableName {
			spec = candidate
			break
		}
	}
	path := cutoverCSVPath(bundle.SourceDir, table)
	resolved, err := resolveCutoverAccountCSV(path)
	if err != nil {
		return &cutoverNamePreflightError{message: fmt.Sprintf("table %s name preflight: source csv path is not resolvable", tableName), cause: err}
	}
	info, err := os.Lstat(resolved)
	if err != nil {
		return &cutoverNamePreflightError{message: fmt.Sprintf("table %s name preflight: source csv is not inspectable", tableName), cause: err}
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("table %s name preflight: csv must not be a symbolic link", tableName)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("table %s name preflight: csv must be a regular file", tableName)
	}
	f, err := openStableOwnerOnlyFile(resolved)
	if err != nil {
		return &cutoverNamePreflightError{message: fmt.Sprintf("table %s name preflight: source csv is not readable", tableName), cause: err}
	}
	defer func() { _ = f.Close() }()
	hash := sha256.New()
	reader := csv.NewReader(bufio.NewReader(io.TeeReader(io.LimitReader(f, maxCutoverCSVBytes+1), hash)))
	header, err := reader.Read()
	if err != nil {
		return &cutoverNamePreflightError{message: fmt.Sprintf("table %s name preflight: source csv header is not readable", tableName), cause: err}
	}
	allowPermutation := rehearsalContractDrift(bundle.Manifest, bundle.Provenance.Mode)
	indexes, _, err := cutoverColumnIndexes(spec, header, allowPermutation)
	if err != nil {
		return &cutoverNamePreflightError{message: fmt.Sprintf("table %s name preflight: source csv header does not match the expected columns", tableName), cause: err}
	}
	nameIdx := indexes["name"]
	var names []string
	rowCount := int64(0)
	for {
		row, readErr := reader.Read()
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return &cutoverNamePreflightError{message: fmt.Sprintf("table %s name preflight: source csv row %d is not parseable", tableName, rowCount+2), cause: readErr}
		}
		rowCount++
		names = append(names, row[nameIdx])
	}
	if rowCount != table.RowCount {
		return fmt.Errorf("table %s name preflight: source row count mismatch: got %d, want %d", tableName, rowCount, table.RowCount)
	}
	if digest := hex.EncodeToString(hash.Sum(nil)); !strings.EqualFold(digest, table.SHA256) {
		return fmt.Errorf("table %s name preflight: source csv sha256 does not match manifest", tableName)
	}
	for start := 0; start < len(names); start += cutoverMedicineNameBatchSize {
		end := start + cutoverMedicineNameBatchSize
		if end > len(names) {
			end = len(names)
		}
		var exists bool
		if err := q.QueryRow(ctx, nameQuery, seeds.ClinicID, names[start:end]).Scan(&exists); err != nil {
			return &cutoverNamePreflightError{message: fmt.Sprintf("table %s name preflight: target name check failed", tableName), cause: err}
		}
		if exists {
			return fmt.Errorf("%s: candidate %s name already exists in the target clinic (deleted_at IS NULL)", CutoverRefBandOccupied, tableName)
		}
	}
	return nil
}

func validateCutoverUniqueMasterTargetNames(ctx context.Context, q cutoverQuerier, bundle CutoverBundle, seeds CutoverSeedIDs, skipComplete bool) error {
	for _, probe := range []struct {
		table string
		query string
	}{
		{"procedures", cutoverProcedureNameQuery},
		{"vaccines", cutoverVaccineNameQuery},
	} {
		if skipComplete {
			manifestTable, err := cutoverManifestTableByName(bundle.Manifest, probe.table)
			if err != nil {
				return err
			}
			var spec CutoverTableSpec
			for _, candidate := range CutoverTableSpecs(bundle.Manifest.profile) {
				if candidate.Name == probe.table {
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
		if err := validateCutoverMasterTargetNames(ctx, q, bundle, seeds, probe.table, probe.query); err != nil {
			return err
		}
	}
	return nil
}
