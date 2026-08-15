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
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sys/unix"
)

const cutoverAdvisoryLockKey = int64(0x41454b41525445) // "AEKARTE"

// ErrCutoverCommitOutcomeUnknown means PostgreSQL did not return a definitive
// commit result. Operators must reconcile with VerifyCutover before retrying or
// restoring a backup because the transaction may already be committed.
var ErrCutoverCommitOutcomeUnknown = errors.New("cutover commit outcome unknown")

// ErrCutoverTransactionNotStarted means no target data transaction was opened.
// It lets the CLI report a precise audit state without exposing driver details.
var ErrCutoverTransactionNotStarted = errors.New("begin cutover transaction failed")

type CutoverSeedIDs struct {
	ClinicID                  int64 `json:"clinicId"`
	AnimalSpeciesID           int64 `json:"animalSpeciesId"`
	ExamTypeID                int64 `json:"examTypeId"`
	TrimmingReservationTypeID int64 `json:"trimmingReservationTypeId"`
	CashPaymentMethodID       int64 `json:"cashPaymentMethodId"`
	CreditCardPaymentMethodID int64 `json:"creditCardPaymentMethodId"`
}

type CutoverResult struct {
	CompletedAt time.Time        `json:"completedAt"`
	ClinicCode  string           `json:"clinicCode"`
	RunID       string           `json:"runId"`
	IDBand      CutoverIDBand    `json:"idBand"`
	Counts      map[string]int64 `json:"counts"`
}

type cutoverSeedFacts struct {
	ClinicExists               bool
	SpeciesActive              bool
	ExamTypeClinicID           int64
	ExamTypeName               string
	ExamTypeActive             bool
	ReservationTypeClinicID    int64
	ReservationTypeCategory    string
	ReservationTypeActive      bool
	CashMethodClinicID         int64
	CashMethodSystemKey        string
	CashMethodActive           bool
	CashMethodMatchCount       int64
	CreditCardMethodClinicID   int64
	CreditCardMethodSystemKey  string
	CreditCardMethodActive     bool
	CreditCardMethodMatchCount int64
}

type cutoverQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

type cutoverTransaction interface {
	cutoverQuerier
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	CopyFrom(context.Context, io.Reader, string) (pgconn.CommandTag, error)
	Commit(context.Context) error
	Rollback(context.Context) error
}

type pgxCutoverTransaction struct {
	pgx.Tx
}

func (tx pgxCutoverTransaction) CopyFrom(ctx context.Context, reader io.Reader, copySQL string) (pgconn.CommandTag, error) {
	return tx.Conn().PgConn().CopyFrom(ctx, reader, copySQL)
}

type cutoverForeignKeySpec struct {
	childTable   string
	childColumn  string
	parentTable  string
	parentColumn string
}

type cutoverCompositeForeignKeySpec struct {
	childTable    string
	childColumns  []string
	parentTable   string
	parentColumns []string
}

func cutoverRequiredForeignKeys() []cutoverForeignKeySpec {
	return []cutoverForeignKeySpec{
		{"staffs", "clinic_id", "clinics", "id"},
		{"procedures", "clinic_id", "clinics", "id"},
		{"procedures", "parent_id", "procedures", "id"},
		{"merchandise_items", "clinic_id", "clinics", "id"},
		{"owners", "clinic_id", "clinics", "id"},
		{"pets", "clinic_id", "clinics", "id"},
		{"pets", "owner_id", "owners", "id"},
		{"pets", "animal_species_id", "animal_species", "id"},
		{"medical_records", "clinic_id", "clinics", "id"},
		{"medical_records", "owner_id", "owners", "id"},
		{"medical_records", "pet_id", "pets", "id"},
		{"medical_records", "doctor_id", "staffs", "id"},
		{"medical_records", "entered_by", "staffs", "id"},
		{"inquiries", "medical_record_id", "medical_records", "id"},
		{"inquiries", "staff_id", "staffs", "id"},
		{"clinical_plans", "medical_record_id", "medical_records", "id"},
		{"vital_records", "clinic_id", "clinics", "id"},
		{"vital_records", "medical_record_id", "medical_records", "id"},
		{"vital_records", "pet_id", "pets", "id"},
		{"vital_records", "staff_id", "staffs", "id"},
		{"appointments", "clinic_id", "clinics", "id"},
		{"appointments", "pet_id", "pets", "id"},
		{"appointments", "reservation_type_id", "reservation_types", "id"},
		{"appointments", "doctor_id", "staffs", "id"},
		{"appointment_trimming_details", "clinic_id", "clinics", "id"},
		{"appointment_trimming_details", "appointment_id", "appointments", "id"},
		{"billings", "clinic_id", "clinics", "id"},
		{"billings", "medical_record_id", "medical_records", "id"},
		{"billings", "owner_id", "owners", "id"},
		{"billings", "pet_id", "pets", "id"},
		{"billing_items", "billing_id", "billings", "id"},
		{"payments", "clinic_id", "clinics", "id"},
		{"payments", "billing_id", "billings", "id"},
		{"payments", "payment_method_id", "payment_methods", "id"},
		{"payments", "paid_by", "staffs", "id"},
		{"payment_splits", "clinic_id", "clinics", "id"},
		{"payment_splits", "billing_id", "billings", "id"},
		{"payment_splits", "payment_method_id", "payment_methods", "id"},
		{"payment_splits", "paid_by", "staffs", "id"},
		{"estimates", "clinic_id", "clinics", "id"},
		{"estimates", "medical_record_id", "medical_records", "id"},
		{"estimates", "owner_id", "owners", "id"},
		{"estimates", "created_by", "staffs", "id"},
		{"estimate_items", "estimate_id", "estimates", "id"},
		{"estimate_items", "consultation_id", "consultations", "id"},
		{"estimate_items", "procedure_id", "procedures", "id"},
		{"estimate_items", "medicine_id", "medicines", "id"},
		{"estimate_items", "merchandise_item_id", "merchandise_items", "id"},
		{"exams", "medical_record_id", "medical_records", "id"},
		{"exams", "clinic_id", "clinics", "id"},
		{"exams", "pet_id", "pets", "id"},
		{"exams", "exam_type_id", "exam_types", "id"},
		{"exam_results", "exam_id", "exams", "id"},
		{"vaccines", "clinic_id", "clinics", "id"},
		{"vaccines", "inventory_id", "inventory_items", "id"},
		{"vaccines", "parent_id", "vaccines", "id"},
		{"vaccinations", "clinic_id", "clinics", "id"},
		{"vaccinations", "medical_record_id", "medical_records", "id"},
		{"vaccinations", "pet_id", "pets", "id"},
		{"vaccinations", "vaccine_id", "vaccines", "id"},
		{"vaccinations", "doctor_id", "staffs", "id"},
	}
}

// cutoverRequiredCompositeForeignKeys lists multi-column clinic-axis FKs that
// single-column inventory cannot express. Column order is matched exactly
// against pg_constraint.conkey/confkey ordinals.
func cutoverRequiredCompositeForeignKeys() []cutoverCompositeForeignKeySpec {
	return []cutoverCompositeForeignKeySpec{
		{
			childTable:    "payments",
			childColumns:  []string{"billing_id", "clinic_id"},
			parentTable:   "billings",
			parentColumns: []string{"id", "clinic_id"},
		},
		{
			childTable:    "payments",
			childColumns:  []string{"payment_method_id", "clinic_id"},
			parentTable:   "payment_methods",
			parentColumns: []string{"id", "clinic_id"},
		},
	}
}

// cutoverRequiredTargetColumnExclusions lists target columns that are NOT NULL,
// have no default, and are not identity/generated, yet are intentionally omitted
// from CutoverTableSpecs.Columns. Prefer extending the CSV contract over adding
// exclusions; only document justified exceptions here.
func cutoverRequiredTargetColumnExclusions() map[string]map[string]struct{} {
	// No exclusions currently justified.
	return nil
}

// cutoverCompositeForeignKeyQuery matches a validated+enforced multi-column FK
// by ordered child/parent column name arrays (conkey/confkey ordinal order).
const cutoverCompositeForeignKeyQuery = `SELECT EXISTS (
  SELECT 1
  FROM pg_constraint c
  JOIN pg_class child ON child.oid = c.conrelid
  JOIN pg_namespace child_ns ON child_ns.oid = child.relnamespace
  JOIN pg_class parent ON parent.oid = c.confrelid
  JOIN pg_namespace parent_ns ON parent_ns.oid = parent.relnamespace
  WHERE c.contype = 'f' AND c.convalidated = true AND c.conenforced = true
    AND child_ns.nspname = 'public' AND child.relname = $1
    AND parent_ns.nspname = 'public' AND parent.relname = $2
    AND array_length(c.conkey, 1) = cardinality($3::text[])
    AND array_length(c.confkey, 1) = cardinality($4::text[])
    AND (
      SELECT array_agg(a.attname ORDER BY cols.ordinality)
      FROM unnest(c.conkey) WITH ORDINALITY AS cols(attnum, ordinality)
      JOIN pg_attribute a ON a.attrelid = child.oid AND a.attnum = cols.attnum
    ) = $3::text[]
    AND (
      SELECT array_agg(a.attname ORDER BY cols.ordinality)
      FROM unnest(c.confkey) WITH ORDINALITY AS cols(attnum, ordinality)
      JOIN pg_attribute a ON a.attrelid = parent.oid AND a.attnum = cols.attnum
    ) = $4::text[]
)`

// cutoverRequiredTargetColumnsQuery lists target columns that COPY must supply:
// NOT NULL, no column_default, not identity, not generated. Serial/identity id
// columns are excluded by default/identity filters so they need not appear as a
// "missing required" drift signal when present only as DB-managed keys.
const cutoverRequiredTargetColumnsQuery = `
SELECT COALESCE(array_agg(column_name::text ORDER BY ordinal_position), '{}'::text[])
FROM information_schema.columns
WHERE table_schema = 'public'
  AND table_name = $1
  AND is_nullable = 'NO'
  AND column_default IS NULL
  AND COALESCE(is_identity, 'NO') = 'NO'
  AND COALESCE(is_generated, 'NEVER') = 'NEVER'
`

// PreflightCutoverTarget is read-only. It verifies the explicit target seed
// bindings, BIGINT ID contract, serial sequences, and that the clinic band has
// not already been used. Apply repeats all checks under transaction/table locks.
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
	return applyCutoverWithBegin(ctx, func(ctx context.Context) (cutoverTransaction, error) {
		tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
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
	if err := verifyCutoverRows(ctx, tx, bundle.Manifest, seeds); err != nil {
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
	if err := validateCutoverTarget(ctx, target, manifest, seeds, false); err != nil {
		return err
	}
	if err := verifyCutoverRows(ctx, target, manifest, seeds); err != nil {
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
			return fmt.Errorf("inspect target composite foreign key for %s(%s)",
				foreignKey.childTable,
				strings.Join(foreignKey.childColumns, ", "),
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

func validateCutoverColumnTypes(ctx context.Context, q cutoverQuerier) error {
	for _, spec := range CutoverTableSpecs() {
		bandColumns := make(map[string]struct{}, len(spec.BandColumns))
		for _, column := range spec.BandColumns {
			bandColumns[column] = struct{}{}
		}
		for _, column := range spec.Columns {
			var dataType string
			err := q.QueryRow(ctx, `SELECT data_type FROM information_schema.columns
WHERE table_schema = 'public' AND table_name = $1 AND column_name = $2`, spec.Name, column).Scan(&dataType)
			if err != nil {
				return fmt.Errorf("target schema is missing required column public.%s.%s", spec.Name, column)
			}
			if _, isBandColumn := bandColumns[column]; isBandColumn && dataType != "bigint" {
				return fmt.Errorf("target schema requires public.%s.%s to exist as bigint", spec.Name, column)
			}
		}
	}
	return nil
}

func validateCutoverSequences(ctx context.Context, q cutoverQuerier) error {
	for _, spec := range CutoverTableSpecs() {
		var sequence *string
		if err := q.QueryRow(ctx, `SELECT pg_get_serial_sequence($1, 'id')`, "public."+spec.Name).Scan(&sequence); err != nil || sequence == nil || *sequence == "" {
			return fmt.Errorf("target table %s must have a serial id sequence", spec.Name)
		}
	}
	return nil
}

func validateCutoverBandEmpty(ctx context.Context, q cutoverQuerier, band CutoverIDBand) error {
	for _, spec := range CutoverTableSpecs() {
		floor := band.NonOwnerIDOffset
		if spec.Name == "owners" {
			floor = band.OwnerFloor
		}
		query := fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM %s WHERE id >= $1 AND id < $2)`, pgx.Identifier{spec.Name}.Sanitize())
		var occupied bool
		if err := q.QueryRow(ctx, query, floor, band.EndExclusive).Scan(&occupied); err != nil {
			return fmt.Errorf("inspect target band for table %s: %w", spec.Name, err)
		}
		if occupied {
			// Non-PHI: table name + stable ref only. Rerun/idempotency guard (Issue #250).
			return fmt.Errorf("%s: target clinic band is already occupied in table %s", CutoverRefBandOccupied, spec.Name)
		}
	}
	return nil
}

func lockCutoverTables(ctx context.Context, tx cutoverTransaction) error {
	identifiers := make([]string, 0, len(CutoverTableSpecs()))
	for _, spec := range CutoverTableSpecs() {
		identifiers = append(identifiers, pgx.Identifier{spec.Name}.Sanitize())
	}
	if _, err := tx.Exec(ctx, "LOCK TABLE "+strings.Join(identifiers, ", ")+" IN SHARE ROW EXCLUSIVE MODE"); err != nil {
		return fmt.Errorf("lock cutover tables: %w", err)
	}
	return nil
}

type cutoverTransformResult struct {
	count int64
	err   error
}

func copyCutoverTable(ctx context.Context, tx cutoverTransaction, path string, spec CutoverTableSpec, table CutoverManifestTable, seeds CutoverSeedIDs) (int64, error) {
	copyCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	reader, writer := io.Pipe()
	result := make(chan cutoverTransformResult, 1)
	go func() {
		count, err := transformCutoverCSV(copyCtx, path, writer, seeds, table.SHA256)
		_ = writer.CloseWithError(err)
		result <- cutoverTransformResult{count: count, err: err}
	}()

	columns := make([]string, len(spec.Columns))
	for i, column := range spec.Columns {
		columns[i] = pgx.Identifier{column}.Sanitize()
	}
	options := []string{"FORMAT csv", "HEADER true"}
	if len(spec.ForceNotNullColumns) > 0 {
		forceNotNullColumns := make([]string, len(spec.ForceNotNullColumns))
		for i, column := range spec.ForceNotNullColumns {
			forceNotNullColumns[i] = pgx.Identifier{column}.Sanitize()
		}
		options = append(options, "FORCE_NOT_NULL ("+strings.Join(forceNotNullColumns, ", ")+")")
	}
	copySQL := "COPY " + pgx.Identifier{spec.Name}.Sanitize() + " (" + strings.Join(columns, ", ") + ") FROM STDIN WITH (" + strings.Join(options, ", ") + ")"
	tag, copyErr := tx.CopyFrom(copyCtx, reader, copySQL)
	if copyErr != nil {
		cancel()
		// Never forward the PostgreSQL COPY error through the pipe: it may echo
		// a source value and the transformer error is otherwise returned to the
		// caller. Closing without that error only unblocks the writer.
		_ = reader.Close()
	}
	transformed := <-result
	if err := cutoverCopyResultError(spec.Name, copyErr, transformed.err); err != nil {
		return 0, err
	}
	if tag.RowsAffected() != table.RowCount || transformed.count != table.RowCount {
		return 0, fmt.Errorf("table %s: imported row count does not match manifest", spec.Name)
	}
	return tag.RowsAffected(), nil
}

func cutoverCopyResultError(table string, copyErr, transformErr error) error {
	if copyErr != nil {
		// PostgreSQL COPY errors may echo a source value. Do not propagate them
		// because the CSV contains PHI. This branch intentionally takes priority
		// over a pipe error caused while cancelling the transformer.
		return fmt.Errorf("table %s: target database rejected the CSV COPY", table)
	}
	if transformErr != nil {
		return fmt.Errorf("table %s: source changed or transformation failed: %w", table, transformErr)
	}
	return nil
}

func transformCutoverCSV(ctx context.Context, path string, output io.Writer, seeds CutoverSeedIDs, expectedSHA256 string) (int64, error) {
	file, err := openStableOwnerOnlyFile(path)
	if err != nil {
		return 0, err
	}
	defer func() { _ = file.Close() }()
	info, err := file.Stat()
	if err != nil || info.Size() > maxCutoverCSVBytes {
		return 0, fmt.Errorf("source CSV exceeds the size limit")
	}
	hash := sha256.New()
	reader := csv.NewReader(bufio.NewReader(io.TeeReader(io.LimitReader(file, maxCutoverCSVBytes+1), hash)))
	writer := csv.NewWriter(output)
	replacements := map[string]string{
		"{{CLINIC_ID}}":                     strconv.FormatInt(seeds.ClinicID, 10),
		"{{FALLBACK_ANIMAL_SPECIES_ID}}":    strconv.FormatInt(seeds.AnimalSpeciesID, 10),
		"{{FALLBACK_EXAM_TYPE_ID}}":         strconv.FormatInt(seeds.ExamTypeID, 10),
		"{{TRIMMING_RESERVATION_TYPE_ID}}":  strconv.FormatInt(seeds.TrimmingReservationTypeID, 10),
		"{{PAYMENT_METHOD_CASH_ID}}":        strconv.FormatInt(seeds.CashPaymentMethodID, 10),
		"{{PAYMENT_METHOD_CREDIT_CARD_ID}}": strconv.FormatInt(seeds.CreditCardPaymentMethodID, 10),
	}
	var records int64
	firstRecord := true
	for {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		record, readErr := reader.Read()
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return 0, fmt.Errorf("read CSV record %d", records+1)
		}
		mapped := append([]string(nil), record...)
		for i, value := range mapped {
			if replacement, ok := replacements[value]; ok {
				mapped[i] = replacement
			}
		}
		if err := writer.Write(mapped); err != nil {
			return 0, fmt.Errorf("write transformed CSV record %d: %w", records+1, err)
		}
		if firstRecord {
			firstRecord = false
		} else {
			records++
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return 0, fmt.Errorf("flush transformed CSV: %w", err)
	}
	actual := hex.EncodeToString(hash.Sum(nil))
	if !strings.EqualFold(actual, expectedSHA256) {
		return 0, fmt.Errorf("CSV sha256 changed after preflight")
	}
	return records, nil
}

func openStableOwnerOnlyFile(path string) (*os.File, error) {
	before, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("inspect source CSV: %w", err)
	}
	if before.Mode()&os.ModeSymlink != 0 || !before.Mode().IsRegular() {
		return nil, fmt.Errorf("source CSV must be a regular file and not a symbolic link")
	}
	if err := requireOwnerOnly(before, false); err != nil {
		return nil, err
	}
	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0) //nolint:gosec // operator path; O_NOFOLLOW enforces the source boundary
	if err != nil {
		return nil, fmt.Errorf("open source CSV: %w", err)
	}
	file := os.NewFile(uintptr(fd), filepath.Base(path))
	if file == nil {
		_ = unix.Close(fd)
		return nil, fmt.Errorf("open source CSV: invalid file descriptor")
	}
	after, err := file.Stat()
	if err != nil || !os.SameFile(before, after) {
		_ = file.Close()
		return nil, fmt.Errorf("source CSV changed while it was opened")
	}
	if !after.Mode().IsRegular() {
		_ = file.Close()
		return nil, fmt.Errorf("source CSV must be a regular file")
	}
	if err := requireOwnerOnly(after, false); err != nil {
		_ = file.Close()
		return nil, err
	}
	return file, nil
}

func verifyCutoverRows(ctx context.Context, q cutoverQuerier, manifest CutoverManifest, seeds CutoverSeedIDs) error {
	for i, spec := range CutoverTableSpecs() {
		floor := manifest.IDBand.NonOwnerIDOffset
		if spec.Name == "owners" {
			floor = manifest.IDBand.OwnerFloor
		}
		query := fmt.Sprintf(`SELECT count(*) FROM %s WHERE id >= $1 AND id < $2`, pgx.Identifier{spec.Name}.Sanitize())
		var count int64
		if err := q.QueryRow(ctx, query, floor, manifest.IDBand.EndExclusive).Scan(&count); err != nil {
			return fmt.Errorf("verify table %s row count: %w", spec.Name, err)
		}
		if count != manifest.Tables[i].RowCount {
			// Non-PHI: table name + stable ref only (no row payloads).
			return fmt.Errorf("%s: table %s: committed row count does not match manifest", CutoverRefRowCount, spec.Name)
		}
		if hasColumn(spec.Columns, "clinic_id") {
			clinicQuery := fmt.Sprintf(`SELECT count(*) FROM %s WHERE id >= $1 AND id < $2 AND clinic_id <> $3`, pgx.Identifier{spec.Name}.Sanitize())
			var mismatched int64
			if err := q.QueryRow(ctx, clinicQuery, floor, manifest.IDBand.EndExclusive, seeds.ClinicID).Scan(&mismatched); err != nil {
				return fmt.Errorf("%s: verify table %s clinic isolation: %w", CutoverRefClinicIsolation, spec.Name, err)
			}
			if mismatched != 0 {
				// Non-PHI: table name + stable ref only. Never echo foreign clinic IDs or row values.
				return fmt.Errorf("%s: table %s contains rows assigned to another clinic", CutoverRefClinicIsolation, spec.Name)
			}
		}
	}
	return verifyCutoverPaymentGraph(ctx, q, &manifest, seeds)
}

func hasColumn(columns []string, target string) bool {
	for _, column := range columns {
		if column == target {
			return true
		}
	}
	return false
}

func advanceCutoverSequences(ctx context.Context, tx cutoverTransaction) error {
	// PostgreSQL sequence changes are non-transactional. This function only
	// advances values and runs after all row/count checks, so the only possible
	// rollback residue is a harmless jump to the reserved application range.
	for _, spec := range CutoverTableSpecs() {
		var sequence *string
		if err := tx.QueryRow(ctx, `SELECT pg_get_serial_sequence($1, 'id')`, "public."+spec.Name).Scan(&sequence); err != nil || sequence == nil {
			return fmt.Errorf("resolve sequence for table %s", spec.Name)
		}
		var current int64
		sequenceSQL := pgx.Identifier(strings.Split(*sequence, ".")).Sanitize()
		if err := tx.QueryRow(ctx, "SELECT last_value FROM "+sequenceSQL).Scan(&current); err != nil {
			return fmt.Errorf("read sequence for table %s: %w", spec.Name, err)
		}
		var maxID int64
		maxQuery := fmt.Sprintf(`SELECT COALESCE(max(id), 0) FROM %s`, pgx.Identifier{spec.Name}.Sanitize())
		if err := tx.QueryRow(ctx, maxQuery).Scan(&maxID); err != nil {
			return fmt.Errorf("read max id for table %s: %w", spec.Name, err)
		}
		lastValue := max(current, max(maxID, applicationIDFloor-1))
		if _, err := tx.Exec(ctx, `SELECT setval($1::regclass, $2, true)`, *sequence, lastValue); err != nil {
			return fmt.Errorf("advance sequence for table %s: %w", spec.Name, err)
		}
	}
	return nil
}

func verifyCutoverSequences(ctx context.Context, q cutoverQuerier) error {
	for _, spec := range CutoverTableSpecs() {
		var sequence *string
		if err := q.QueryRow(ctx, `SELECT pg_get_serial_sequence($1, 'id')`, "public."+spec.Name).Scan(&sequence); err != nil || sequence == nil {
			return fmt.Errorf("resolve sequence for table %s", spec.Name)
		}
		sequenceSQL := pgx.Identifier(strings.Split(*sequence, ".")).Sanitize()
		var lastValue int64
		var isCalled bool
		if err := q.QueryRow(ctx, "SELECT last_value, is_called FROM "+sequenceSQL).Scan(&lastValue, &isCalled); err != nil {
			return fmt.Errorf("verify sequence for table %s: %w", spec.Name, err)
		}
		nextValue := lastValue
		if isCalled {
			if lastValue == math.MaxInt64 {
				return fmt.Errorf("table %s sequence exhausted", spec.Name)
			}
			nextValue++
		}
		if nextValue < applicationIDFloor {
			return fmt.Errorf("table %s sequence next value is below application floor", spec.Name)
		}
		var maxID int64
		maxQuery := fmt.Sprintf(`SELECT COALESCE(max(id), 0) FROM %s`, pgx.Identifier{spec.Name}.Sanitize())
		if err := q.QueryRow(ctx, maxQuery).Scan(&maxID); err != nil {
			return fmt.Errorf("read max id for table %s during sequence verification: %w", spec.Name, err)
		}
		if nextValue <= maxID {
			return fmt.Errorf("table %s sequence next value does not exceed the current max id", spec.Name)
		}
	}
	return nil
}
