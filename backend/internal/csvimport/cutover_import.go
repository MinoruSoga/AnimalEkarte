package csvimport

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
		// doctor_id / entered_by are clinic-scoped composites (see cutoverRequiredCompositeForeignKeys).
		{"inquiries", "medical_record_id", "medical_records", "id"},
		{"inquiries", "staff_id", "staffs", "id"},
		{"clinical_plans", "medical_record_id", "medical_records", "id"},
		{"vital_records", "clinic_id", "clinics", "id"},
		{"vital_records", "medical_record_id", "medical_records", "id"},
		{"vital_records", "pet_id", "pets", "id"},
		{"vital_records", "staff_id", "staffs", "id"},
		{"appointments", "clinic_id", "clinics", "id"},
		// owner_id / pet_id / doctor_id are clinic-scoped composites.
		{"appointments", "reservation_type_id", "reservation_types", "id"},
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
			childTable:    "medical_records",
			childColumns:  []string{"doctor_id", "clinic_id"},
			parentTable:   "staffs",
			parentColumns: []string{"id", "clinic_id"},
		},
		{
			childTable:    "medical_records",
			childColumns:  []string{"entered_by", "clinic_id"},
			parentTable:   "staffs",
			parentColumns: []string{"id", "clinic_id"},
		},
		{
			childTable:    "appointments",
			childColumns:  []string{"clinic_id", "owner_id"},
			parentTable:   "owners",
			parentColumns: []string{"clinic_id", "id"},
		},
		{
			childTable:    "appointments",
			childColumns:  []string{"clinic_id", "pet_id"},
			parentTable:   "pets",
			parentColumns: []string{"clinic_id", "id"},
		},
		{
			childTable:    "appointments",
			childColumns:  []string{"doctor_id", "clinic_id"},
			parentTable:   "staffs",
			parentColumns: []string{"id", "clinic_id"},
		},
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
    )::text[] = $3::text[]
    AND (
      SELECT array_agg(a.attname ORDER BY cols.ordinality)
      FROM unnest(c.confkey) WITH ORDINALITY AS cols(attnum, ordinality)
      JOIN pg_attribute a ON a.attrelid = parent.oid AND a.attnum = cols.attnum
    )::text[] = $4::text[]
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
