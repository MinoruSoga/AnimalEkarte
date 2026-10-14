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

// cutoverMedicineNameBatchSize bounds the names per target lookup so one
// query stays small; no per-row database queries and no source values in
// errors.
const cutoverMedicineNameBatchSize = 256

// validateCutoverMedicinesTargetNames implements the pre-Begin leg of
// Codex 4266290426 for user decision A: before the first table transaction
// of a candidate22 table-per-commit apply, the bundle's medicines CSV is
// re-read (row count and SHA-256 bound to the exact opened bytes; header
// permutation only under the existing rehearsal contract drift) and every
// exact source name is checked against the target clinic's undeleted
// medicines. Soft-deleted rows and other clinics never collide; inactive
// rows do. Legacy21 never calls this helper. Read-only: no insert, update,
// or delete.
func validateCutoverMedicinesTargetNames(ctx context.Context, session cutoverSession, bundle CutoverBundle, seeds CutoverSeedIDs) error {
	table, err := cutoverManifestTableByName(bundle.Manifest, "medicines")
	if err != nil {
		return fmt.Errorf("medicines name preflight: %w", err)
	}
	var spec CutoverTableSpec
	for _, candidate := range CutoverTableSpecs(bundle.Manifest.profile) {
		if candidate.Name == "medicines" {
			spec = candidate
			break
		}
	}
	path := cutoverCSVPath(bundle.SourceDir, table)
	resolved, err := resolveCutoverAccountCSV(path)
	if err != nil {
		return fmt.Errorf("medicines name preflight: resolve csv: %w", err)
	}
	info, err := os.Lstat(resolved)
	if err != nil {
		return fmt.Errorf("medicines name preflight: inspect csv: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("medicines name preflight: csv must not be a symbolic link")
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("medicines name preflight: csv must be a regular file")
	}
	f, err := openStableOwnerOnlyFile(resolved)
	if err != nil {
		return fmt.Errorf("medicines name preflight: open csv: %w", err)
	}
	defer func() { _ = f.Close() }()
	hash := sha256.New()
	reader := csv.NewReader(bufio.NewReader(io.TeeReader(io.LimitReader(f, maxCutoverCSVBytes+1), hash)))
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("medicines name preflight: read csv header: %w", err)
	}
	allowPermutation := rehearsalContractDrift(bundle.Manifest, bundle.Provenance.Mode)
	indexes, _, err := cutoverColumnIndexes(spec, header, allowPermutation)
	if err != nil {
		return fmt.Errorf("medicines name preflight: %w", err)
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
			return fmt.Errorf("medicines name preflight: read csv row %d: %w", rowCount+2, readErr)
		}
		rowCount++
		names = append(names, row[nameIdx])
	}
	if rowCount != table.RowCount {
		return fmt.Errorf("medicines name preflight: source row count mismatch: got %d, want %d", rowCount, table.RowCount)
	}
	if digest := hex.EncodeToString(hash.Sum(nil)); !strings.EqualFold(digest, table.SHA256) {
		return fmt.Errorf("medicines name preflight: source csv sha256 does not match manifest")
	}
	for start := 0; start < len(names); start += cutoverMedicineNameBatchSize {
		end := start + cutoverMedicineNameBatchSize
		if end > len(names) {
			end = len(names)
		}
		var exists bool
		if err := session.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM public.medicines WHERE clinic_id = $1 AND deleted_at IS NULL AND name = ANY($2::text[]))`,
			seeds.ClinicID, names[start:end]).Scan(&exists); err != nil {
			return fmt.Errorf("medicines name preflight: inspect target names: %w", err)
		}
		if exists {
			return fmt.Errorf("%s: candidate medicines name already exists in the target clinic (deleted_at IS NULL)", CutoverRefBandOccupied)
		}
	}
	return nil
}
