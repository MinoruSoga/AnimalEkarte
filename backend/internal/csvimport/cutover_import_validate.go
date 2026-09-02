package csvimport

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func PreflightCutoverTarget(ctx context.Context, target cutoverQuerier, manifest CutoverManifest, seeds CutoverSeedIDs) error {
	if err := validateCutoverTarget(ctx, target, manifest, seeds, true); err != nil {
		return err
	}
	return nil
}

// ApplyCutover imports all twenty-one CSVs in one transaction. It never deletes
// existing rows: a non-empty clinic band fails closed, making retries explicit
// and preventing a cutover from silently replacing unrelated data.
func ApplyCutover(ctx context.Context, pool *pgxpool.Pool, bundle CutoverBundle, seeds CutoverSeedIDs) (CutoverResult, error) {
	return ApplyCutoverWithIsolation(ctx, pool, bundle, seeds, pgx.Serializable)
}

// ApplyCutoverWithIsolation is ApplyCutover with an explicit transaction
// isolation level. Formal cutover keeps Serializable. Local rehearsal may use
// RepeatableRead so multi-million-row COPYs do not exhaust SSI predicate locks
// ("out of shared memory").
func ApplyCutoverWithIsolation(
	ctx context.Context,
	pool *pgxpool.Pool,
	bundle CutoverBundle,
	seeds CutoverSeedIDs,
	iso pgx.TxIsoLevel,
) (CutoverResult, error) {
	if iso == "" {
		iso = pgx.Serializable
	}
	return applyCutoverWithBegin(ctx, func(ctx context.Context) (cutoverTransaction, error) {
		tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: iso})
		if err != nil {
			return nil, err
		}
		return pgxCutoverTransaction{Tx: tx}, nil
	}, bundle, seeds)
}

func applyCutoverWithBegin(
	ctx context.Context,
	begin func(context.Context) (cutoverTransaction, error),
	bundle CutoverBundle,
	seeds CutoverSeedIDs,
) (CutoverResult, error) {
	tx, err := begin(ctx)
	if err != nil {
		return CutoverResult{}, ErrCutoverTransactionNotStarted
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `SET LOCAL lock_timeout = '10s'`); err != nil {
		return CutoverResult{}, fmt.Errorf("configure cutover lock timeout: %w", err)
	}
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, cutoverAdvisoryLockKey); err != nil {
		return CutoverResult{}, fmt.Errorf("acquire cutover lock: %w", err)
	}
	if err := lockCutoverTables(ctx, tx); err != nil {
		return CutoverResult{}, err
	}
	if err := validateCutoverTarget(ctx, tx, bundle.Manifest, seeds, true); err != nil {
		return CutoverResult{}, err
	}

	counts := make(map[string]int64, len(bundle.Manifest.Tables))
	for i, spec := range CutoverTableSpecs() {
		manifestTable := bundle.Manifest.Tables[i]
		path := filepath.Join(bundle.SourceDir, manifestTable.File)
		count, err := copyCutoverTable(ctx, tx, path, spec, manifestTable, seeds)
		if err != nil {
			return CutoverResult{}, err
		}
		counts[spec.Name] = count
	}
	if err := verifyCutoverRows(ctx, tx, bundle.Manifest, seeds, bundle.Provenance); err != nil {
		return CutoverResult{}, err
	}
	if err := advanceCutoverSequences(ctx, tx); err != nil {
		return CutoverResult{}, err
	}
	if err := verifyCutoverSequences(ctx, tx); err != nil {
		return CutoverResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		if errors.Is(err, pgx.ErrTxCommitRollback) {
			return CutoverResult{}, fmt.Errorf("commit rejected and transaction rolled back: %w", err)
		}
		// Do not wrap the driver error: a server-side error can include imported
		// values, and any non-definitive commit response must be treated as an
		// indeterminate outcome rather than a safe rollback.
		return CutoverResult{}, fmt.Errorf("%w: run read-only verify before any retry or restore", ErrCutoverCommitOutcomeUnknown)
	}
	return CutoverResult{
		CompletedAt: time.Now().UTC(),
		ClinicCode:  bundle.Manifest.ClinicCode,
		RunID:       bundle.Manifest.SourceRunID,
		IDBand:      bundle.Manifest.IDBand,
		Counts:      counts,
	}, nil
}

// VerifyCutover checks the committed database against the trusted manifest.
// It is read-only and intentionally does not treat a populated band as an error.
func VerifyCutover(ctx context.Context, target cutoverQuerier, manifest CutoverManifest, seeds CutoverSeedIDs) error {
	return VerifyCutoverWithProvenance(ctx, target, manifest, seeds, CutoverProvenanceContract{Mode: CutoverProvenanceFormal})
}

func VerifyCutoverWithProvenance(ctx context.Context, target cutoverQuerier, manifest CutoverManifest, seeds CutoverSeedIDs, provenance CutoverProvenanceContract) error {
	if err := validateCutoverTarget(ctx, target, manifest, seeds, false); err != nil {
		return err
	}
	if err := verifyCutoverRows(ctx, target, manifest, seeds, provenance); err != nil {
		return err
	}
	return verifyCutoverSequences(ctx, target)
}

func validateCutoverTarget(ctx context.Context, q cutoverQuerier, manifest CutoverManifest, seeds CutoverSeedIDs, requireEmptyBand bool) error {
	facts, err := queryCutoverSeedFacts(ctx, q, seeds)
	if err != nil {
		return err
	}
	if err := validateCutoverSeedFacts(seeds, facts); err != nil {
		return err
	}
	if err := validateRequiredAnimalSpecies(ctx, q); err != nil {
		return err
	}
	if err := validateCutoverColumnTypes(ctx, q); err != nil {
		return err
	}
	if err := validateCutoverRequiredTargetColumns(ctx, q); err != nil {
		return err
	}
	if err := validateCutoverForeignKeys(ctx, q); err != nil {
		return err
	}
	if err := validateCutoverCompositeForeignKeys(ctx, q); err != nil {
		return err
	}
	if err := validateCutoverUniqueIndexes(ctx, q); err != nil {
		return err
	}
	if err := validateCutoverSequences(ctx, q); err != nil {
		return err
	}
	if requireEmptyBand {
		return validateCutoverBandEmpty(ctx, q, manifest.IDBand)
	}
	return nil
}

func validateCutoverForeignKeys(ctx context.Context, q cutoverQuerier) error {
	const query = `SELECT EXISTS (
  SELECT 1
  FROM pg_constraint c
  JOIN pg_class child ON child.oid = c.conrelid
  JOIN pg_namespace child_ns ON child_ns.oid = child.relnamespace
  JOIN pg_class parent ON parent.oid = c.confrelid
  JOIN pg_namespace parent_ns ON parent_ns.oid = parent.relnamespace
  JOIN pg_attribute child_col ON child_col.attrelid = child.oid AND child_col.attnum = c.conkey[1]
  JOIN pg_attribute parent_col ON parent_col.attrelid = parent.oid AND parent_col.attnum = c.confkey[1]
	  WHERE c.contype = 'f' AND c.convalidated = true AND c.conenforced = true
    AND array_length(c.conkey, 1) = 1 AND array_length(c.confkey, 1) = 1
    AND child_ns.nspname = 'public' AND child.relname = $1 AND child_col.attname = $2
    AND parent_ns.nspname = 'public' AND parent.relname = $3 AND parent_col.attname = $4
)`
	for _, foreignKey := range cutoverRequiredForeignKeys() {
		var exists bool
		if err := q.QueryRow(ctx, query,
			foreignKey.childTable,
			foreignKey.childColumn,
			foreignKey.parentTable,
			foreignKey.parentColumn,
		).Scan(&exists); err != nil {
			return fmt.Errorf("inspect target foreign key for %s.%s", foreignKey.childTable, foreignKey.childColumn)
		}
		if !exists {
			return fmt.Errorf("target requires a validated foreign key from %s.%s to %s.%s",
				foreignKey.childTable,
				foreignKey.childColumn,
				foreignKey.parentTable,
				foreignKey.parentColumn,
			)
		}
	}
	return nil
}

func validateCutoverCompositeForeignKeys(ctx context.Context, q cutoverQuerier) error {
	for _, foreignKey := range cutoverRequiredCompositeForeignKeys() {
		var exists bool
		if err := q.QueryRow(ctx, cutoverCompositeForeignKeyQuery,
			foreignKey.childTable,
			foreignKey.parentTable,
			foreignKey.childColumns,
			foreignKey.parentColumns,
		).Scan(&exists); err != nil {
			return fmt.Errorf("inspect target composite foreign key for %s(%s): %w",
				foreignKey.childTable,
				strings.Join(foreignKey.childColumns, ", "),
				err,
			)
		}
		if !exists {
			return fmt.Errorf("target requires a validated composite foreign key from %s(%s) to %s(%s)",
				foreignKey.childTable,
				strings.Join(foreignKey.childColumns, ", "),
				foreignKey.parentTable,
				strings.Join(foreignKey.parentColumns, ", "),
			)
		}
	}
	return nil
}

// validateCutoverRequiredTargetColumns fail-closes when the target has a column
// that COPY cannot leave unset (NOT NULL, no default, not identity/generated)
// but CutoverTableSpecs.Columns omits it. This catches Z1-class drift such as
// payments.clinic_id becoming NOT NULL on the target while remaining absent from
// the CSV contract.
func validateCutoverRequiredTargetColumns(ctx context.Context, q cutoverQuerier) error {
	exclusions := cutoverRequiredTargetColumnExclusions()
	for _, spec := range CutoverTableSpecs() {
		declared := make(map[string]struct{}, len(spec.Columns))
		for _, column := range spec.Columns {
			declared[column] = struct{}{}
		}
		var required []string
		if err := q.QueryRow(ctx, cutoverRequiredTargetColumnsQuery, spec.Name).Scan(&required); err != nil {
			return fmt.Errorf("inspect required target columns for %s: %w", spec.Name, err)
		}
		tableExclusions := exclusions[spec.Name]
		for _, column := range required {
			if _, ok := declared[column]; ok {
				continue
			}
			if _, excluded := tableExclusions[column]; excluded {
				continue
			}
			return fmt.Errorf("cutover CSV contract is missing required target column %s.%s", spec.Name, column)
		}
	}
	return nil
}

func queryCutoverSeedFacts(ctx context.Context, q cutoverQuerier, seeds CutoverSeedIDs) (cutoverSeedFacts, error) {
	const query = `
WITH
  clinic_seed AS MATERIALIZED (
    SELECT id, is_active FROM clinics WHERE id = $1 FOR SHARE
  ),
  species_seed AS MATERIALIZED (
    SELECT id, is_active FROM animal_species WHERE id = $2 FOR SHARE
  ),
  exam_seed AS MATERIALIZED (
    SELECT clinic_id, name, is_active FROM exam_types WHERE id = $3 AND deleted_at IS NULL FOR SHARE
  ),
  reservation_seed AS MATERIALIZED (
    SELECT clinic_id, category, is_active FROM reservation_types WHERE id = $4 AND deleted_at IS NULL FOR SHARE
  ),
  cash_method_seed AS MATERIALIZED (
    SELECT clinic_id, system_key, is_active FROM payment_methods WHERE id = $5 AND deleted_at IS NULL FOR SHARE
  ),
  credit_card_method_seed AS MATERIALIZED (
    SELECT clinic_id, system_key, is_active FROM payment_methods WHERE id = $6 AND deleted_at IS NULL FOR SHARE
  )
SELECT
  EXISTS (SELECT 1 FROM clinic_seed WHERE is_active = true),
  EXISTS (SELECT 1 FROM species_seed WHERE is_active = true),
  COALESCE((SELECT clinic_id FROM exam_seed), 0),
  COALESCE((SELECT name FROM exam_seed), ''),
  COALESCE((SELECT is_active FROM exam_seed), false),
  COALESCE((SELECT clinic_id FROM reservation_seed), 0),
  COALESCE((SELECT category::text FROM reservation_seed), ''),
  COALESCE((SELECT is_active FROM reservation_seed), false),
  COALESCE((SELECT clinic_id FROM cash_method_seed), 0),
  COALESCE((SELECT system_key FROM cash_method_seed), ''),
  COALESCE((SELECT is_active FROM cash_method_seed), false),
  (SELECT count(*) FROM payment_methods
    WHERE clinic_id = $1 AND system_key = 'cash' AND deleted_at IS NULL),
  COALESCE((SELECT clinic_id FROM credit_card_method_seed), 0),
  COALESCE((SELECT system_key FROM credit_card_method_seed), ''),
  COALESCE((SELECT is_active FROM credit_card_method_seed), false),
  (SELECT count(*) FROM payment_methods
    WHERE clinic_id = $1 AND system_key = 'credit_card' AND deleted_at IS NULL)`
	var facts cutoverSeedFacts
	if err := q.QueryRow(ctx, query,
		seeds.ClinicID,
		seeds.AnimalSpeciesID,
		seeds.ExamTypeID,
		seeds.TrimmingReservationTypeID,
		seeds.CashPaymentMethodID,
		seeds.CreditCardPaymentMethodID,
	).Scan(
		&facts.ClinicExists,
		&facts.SpeciesActive,
		&facts.ExamTypeClinicID,
		&facts.ExamTypeName,
		&facts.ExamTypeActive,
		&facts.ReservationTypeClinicID,
		&facts.ReservationTypeCategory,
		&facts.ReservationTypeActive,
		&facts.CashMethodClinicID,
		&facts.CashMethodSystemKey,
		&facts.CashMethodActive,
		&facts.CashMethodMatchCount,
		&facts.CreditCardMethodClinicID,
		&facts.CreditCardMethodSystemKey,
		&facts.CreditCardMethodActive,
		&facts.CreditCardMethodMatchCount,
	); err != nil {
		return cutoverSeedFacts{}, fmt.Errorf("inspect target seed bindings: %w", err)
	}
	return facts, nil
}

// requiredAnimalSpeciesRow is the fixed producer crosswalk / 002_master contract.
// Labels are master-data values (not PHI). Errors must not echo live DB names.
type requiredAnimalSpeciesRow struct {
	ID   int64
	Name string
}

func requiredAnimalSpeciesRows() []requiredAnimalSpeciesRow {
	return []requiredAnimalSpeciesRow{
		{1, "犬"},
		{2, "猫"},
		{3, "鳥"},
		{4, "うさぎ"},
		{5, "ハムスター"},
		{6, "その他"},
	}
}

type requiredAnimalSpeciesFacts struct {
	MissingCount          int64
	InactiveCount         int64
	RenamedCount          int64
	ExactActiveMatchCount int64
}

const requiredAnimalSpeciesQuery = `
-- required animal_species master
WITH expected(id, name) AS (
  VALUES
    ($1::bigint, $2::text),
    ($3::bigint, $4::text),
    ($5::bigint, $6::text),
    ($7::bigint, $8::text),
    ($9::bigint, $10::text),
    ($11::bigint, $12::text)
),
locked AS MATERIALIZED (
  SELECT s.id, s.name, s.is_active
  FROM animal_species s
  WHERE s.id IN (SELECT e.id FROM expected e)
  FOR SHARE
)
SELECT
  (SELECT COALESCE(SUM(1), 0)::bigint
     FROM expected e
    WHERE NOT EXISTS (SELECT 1 FROM locked l WHERE l.id = e.id)),
  (SELECT COALESCE(SUM(1), 0)::bigint
     FROM expected e
     JOIN locked l ON l.id = e.id
    WHERE COALESCE(l.is_active, false) = false),
  (SELECT COALESCE(SUM(1), 0)::bigint
     FROM expected e
     JOIN locked l ON l.id = e.id
    WHERE COALESCE(l.is_active, false) = true
      AND l.name IS DISTINCT FROM e.name),
  (SELECT COALESCE(SUM(1), 0)::bigint
     FROM expected e
     JOIN locked l ON l.id = e.id
    WHERE COALESCE(l.is_active, false) = true
      AND l.name = e.name)`

func queryRequiredAnimalSpeciesFacts(ctx context.Context, q cutoverQuerier) (requiredAnimalSpeciesFacts, error) {
	rows := requiredAnimalSpeciesRows()
	args := make([]any, 0, len(rows)*2)
	for _, row := range rows {
		args = append(args, row.ID, row.Name)
	}
	var facts requiredAnimalSpeciesFacts
	if err := q.QueryRow(ctx, requiredAnimalSpeciesQuery, args...).Scan(
		&facts.MissingCount,
		&facts.InactiveCount,
		&facts.RenamedCount,
		&facts.ExactActiveMatchCount,
	); err != nil {
		return requiredAnimalSpeciesFacts{}, fmt.Errorf("inspect required animal_species master rows: %w", err)
	}
	return facts, nil
}

func validateRequiredAnimalSpeciesFacts(facts requiredAnimalSpeciesFacts) error {
	const want = int64(6)
	if facts.MissingCount > 0 {
		return fmt.Errorf("required animal_species master rows are missing (expected exactly 6 fixed id/name pairs)")
	}
	if facts.InactiveCount > 0 {
		return fmt.Errorf("required animal_species master rows are inactive")
	}
	if facts.RenamedCount > 0 {
		return fmt.Errorf("required animal_species master rows have unexpected names")
	}
	if facts.ExactActiveMatchCount != want {
		return fmt.Errorf("required animal_species master rows are incomplete or inconsistent")
	}
	return nil
}

func validateRequiredAnimalSpecies(ctx context.Context, q cutoverQuerier) error {
	facts, err := queryRequiredAnimalSpeciesFacts(ctx, q)
	if err != nil {
		return err
	}
	return validateRequiredAnimalSpeciesFacts(facts)
}

func validateCutoverSeedFacts(seeds CutoverSeedIDs, facts cutoverSeedFacts) error {
	if seeds.ClinicID <= 0 ||
		seeds.AnimalSpeciesID <= 0 ||
		seeds.ExamTypeID <= 0 ||
		seeds.TrimmingReservationTypeID <= 0 ||
		seeds.CashPaymentMethodID <= 0 ||
		seeds.CreditCardPaymentMethodID <= 0 {
		return fmt.Errorf("all six explicit target seed IDs must be positive")
	}
	if seeds.CashPaymentMethodID == seeds.CreditCardPaymentMethodID {
		return fmt.Errorf("cash and credit-card payment method seed IDs must be different")
	}
	if !facts.ClinicExists {
		return fmt.Errorf("target clinic seed is missing or inactive")
	}
	if !facts.SpeciesActive {
		return fmt.Errorf("fallback animal species seed is missing or inactive")
	}
	if facts.ExamTypeClinicID != seeds.ClinicID || facts.ExamTypeName != "検査" || !facts.ExamTypeActive {
		return fmt.Errorf("fallback exam type must be an active, non-deleted '検査' row for the target clinic")
	}
	if facts.ReservationTypeClinicID != seeds.ClinicID || facts.ReservationTypeCategory != "trimming" || !facts.ReservationTypeActive {
		return fmt.Errorf("trimming reservation type must be active, non-deleted, category=trimming, and belong to the target clinic")
	}
	if facts.CashMethodClinicID != seeds.ClinicID ||
		facts.CashMethodSystemKey != "cash" ||
		!facts.CashMethodActive {
		return fmt.Errorf("cash payment method must be active, non-deleted, system_key=cash, and belong to the target clinic")
	}
	if facts.CashMethodMatchCount != 1 {
		return fmt.Errorf("target clinic must have exactly one unique cash payment method")
	}
	if facts.CreditCardMethodClinicID != seeds.ClinicID ||
		facts.CreditCardMethodSystemKey != "credit_card" ||
		!facts.CreditCardMethodActive {
		return fmt.Errorf("credit-card payment method must be active, non-deleted, system_key=credit_card, and belong to the target clinic")
	}
	if facts.CreditCardMethodMatchCount != 1 {
		return fmt.Errorf("target clinic must have exactly one unique credit-card payment method")
	}
	return nil
}
