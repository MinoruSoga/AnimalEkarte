package csvimport

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/animal-ekarte/backend/internal/dbconn"
)

// This test executes the catalog and payment-graph SQL against real
// PostgreSQL. All fixtures live in temporary tables inside a transaction that
// is always rolled back. It is opt-in because the shared Compose DB requires
// the project-wide DB lease even for rollback-only verification.
func TestCutoverPaymentTargetSQLAgainstPostgres(t *testing.T) {
	if os.Getenv("CSVIMPORT_DB_INTEGRATION") != "1" {
		t.Skip("set CSVIMPORT_DB_INTEGRATION=1 under the shared DB lease")
	}

	params, err := dbconn.FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	database := os.Getenv("DB_NAME")
	if database == "" {
		t.Fatal("DB_NAME is required")
	}
	if !dbconn.IsLocalHost(params.Host) || params.SSLMode != "disable" {
		t.Fatal("integration test only accepts a local/container database with SSL disabled")
	}
	port, err := strconv.ParseUint(params.Port, 10, 16)
	if err != nil || port == 0 {
		t.Fatal("DB_PORT must be an integer between 1 and 65535")
	}

	poolConfig, err := pgxpool.ParseConfig("sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	poolConfig.ConnConfig.Host = params.Host
	poolConfig.ConnConfig.Port = uint16(port)
	poolConfig.ConnConfig.User = params.User
	poolConfig.ConnConfig.Password = params.Password
	poolConfig.ConnConfig.Database = database
	poolConfig.ConnConfig.Fallbacks = nil

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	if err := validateCutoverUniqueIndexes(ctx, tx); err != nil {
		t.Fatal(err)
	}

	fixtureDDL := []string{
		`SET LOCAL search_path = pg_temp, public`,
		`CREATE TEMP TABLE billings (
  id bigint PRIMARY KEY,
  clinic_id bigint NOT NULL,
  total_amount bigint NOT NULL,
  status public.billing_status NOT NULL,
  completed_at timestamptz,
  deleted_at timestamptz
)`,
		`CREATE TEMP TABLE staffs (
  id bigint PRIMARY KEY,
  clinic_id bigint NOT NULL
)`,
		`CREATE TEMP TABLE payments (
  id bigint PRIMARY KEY,
  clinic_id bigint NOT NULL,
  billing_id bigint NOT NULL UNIQUE,
  subtotal bigint,
  tax_total bigint,
  total_amount bigint,
  insurance_ratio numeric(3,2),
  insurance_amount bigint,
  discount_amount bigint,
  billing_amount bigint,
  received_amount bigint,
  change_amount bigint,
  method public.payment_method,
  payment_method_id bigint,
  paid_by bigint,
  created_at timestamptz,
  deleted_at timestamptz
)`,
		`CREATE TEMP TABLE payment_splits (
  id bigint PRIMARY KEY,
  clinic_id bigint NOT NULL,
  billing_id bigint NOT NULL,
  method public.payment_method NOT NULL,
  payment_method_id bigint,
  amount bigint,
  received_amount bigint,
  change_amount bigint,
  paid_by bigint,
  created_at timestamptz
		)`,
		`CREATE INDEX ON payment_splits (billing_id)`,
	}
	for _, statement := range fixtureDDL {
		if _, err := tx.Exec(ctx, statement); err != nil {
			t.Fatalf("create payment graph fixtures: %v", err)
		}
	}

	const (
		bandStart = int64(9_000_000_000_000_000_000)
		bandEnd   = int64(9_000_000_000_000_000_100)
		clinicID  = int64(101)
		cashID    = int64(201)
		cardID    = int64(202)
		staffID   = bandStart + 1
		billingID = bandStart + 2
		paymentID = bandStart + 3
		splitID   = bandStart + 4
	)
	if _, err := tx.Exec(ctx, `INSERT INTO staffs (id, clinic_id) VALUES ($1, $2)`, staffID, clinicID); err != nil {
		t.Fatalf("insert staff fixture: %v", err)
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO billings (id, clinic_id, total_amount, status, completed_at)
VALUES ($1, $2, 1000, 'completed', '2026-07-22T00:00:00Z')
`, billingID, clinicID); err != nil {
		t.Fatalf("insert billing fixture: %v", err)
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO payments (
  id, clinic_id, billing_id, subtotal, tax_total, total_amount, insurance_ratio,
  insurance_amount, discount_amount, billing_amount, received_amount,
  change_amount, method, payment_method_id, paid_by, created_at
) VALUES (
  $1, $2, $3, 1000, 0, 1000, 0, 0, 0, 1000, 1000,
  0, 'cash', $4, $5, '2026-07-22T00:00:00Z'
)
`, paymentID, clinicID, billingID, cashID, staffID); err != nil {
		t.Fatalf("insert payment fixture: %v", err)
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO payment_splits (
  id, clinic_id, billing_id, method, payment_method_id, amount,
  received_amount, change_amount, paid_by, created_at
) VALUES (
  $1, $2, $3, 'cash', $4, 1000,
  1000, 0, $5, '2026-07-22T00:00:00Z'
)
`, splitID, clinicID, billingID, cashID, staffID); err != nil {
		t.Fatalf("insert payment split fixture: %v", err)
	}

	manifest := CutoverManifest{IDBand: CutoverIDBand{
		NonOwnerIDOffset: bandStart,
		EndExclusive:     bandEnd,
	}}
	seeds := CutoverSeedIDs{
		ClinicID: clinicID, CashPaymentMethodID: cashID, CreditCardPaymentMethodID: cardID,
	}
	if err := verifyCutoverPaymentGraph(ctx, tx, &manifest, seeds, CutoverProvenanceContract{Mode: CutoverProvenanceFormal}); err != nil {
		t.Fatal(err)
	}

	assertRejected := func(
		name string,
		mutateSQL string,
		mutateArgs []any,
		restoreSQL string,
		restoreArgs []any,
	) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			if _, err := tx.Exec(ctx, mutateSQL, mutateArgs...); err != nil {
				t.Fatalf("mutate fixture: %v", err)
			}
			if err := verifyCutoverPaymentGraph(ctx, tx, &manifest, seeds, CutoverProvenanceContract{Mode: CutoverProvenanceFormal}); err == nil {
				t.Fatal("payment graph violation was accepted")
			}
			if _, err := tx.Exec(ctx, restoreSQL, restoreArgs...); err != nil {
				t.Fatalf("restore fixture: %v", err)
			}
		})
	}
	assertRejected(
		"cross-clinic payment",
		"UPDATE payments SET clinic_id = $1 WHERE id = $2",
		[]any{clinicID + 1, paymentID},
		"UPDATE payments SET clinic_id = $1 WHERE id = $2",
		[]any{clinicID, paymentID},
	)
	assertRejected(
		"cross-clinic split",
		"UPDATE payment_splits SET clinic_id = $1 WHERE id = $2",
		[]any{clinicID + 1, splitID},
		"UPDATE payment_splits SET clinic_id = $1 WHERE id = $2",
		[]any{clinicID, splitID},
	)
	assertRejected(
		"cross-clinic staff",
		"UPDATE staffs SET clinic_id = $1 WHERE id = $2",
		[]any{clinicID + 1, staffID},
		"UPDATE staffs SET clinic_id = $1 WHERE id = $2",
		[]any{clinicID, staffID},
	)
	assertRejected(
		"payment method mismatch",
		"UPDATE payment_splits SET payment_method_id = $1 WHERE id = $2",
		[]any{cardID, splitID},
		"UPDATE payment_splits SET payment_method_id = $1 WHERE id = $2",
		[]any{cashID, splitID},
	)
	assertRejected(
		"outside-band split",
		"UPDATE payment_splits SET id = $1 WHERE billing_id = $2",
		[]any{bandEnd, billingID},
		"UPDATE payment_splits SET id = $1 WHERE billing_id = $2",
		[]any{splitID, billingID},
	)
	assertRejected(
		"nullable insurance ratio",
		"UPDATE payments SET insurance_ratio = NULL WHERE id = $1",
		[]any{paymentID},
		"UPDATE payments SET insurance_ratio = 0 WHERE id = $1",
		[]any{paymentID},
	)
	assertRejected(
		"negative insurance amount",
		"UPDATE payments SET insurance_amount = -1 WHERE id = $1",
		[]any{paymentID},
		"UPDATE payments SET insurance_amount = 0 WHERE id = $1",
		[]any{paymentID},
	)
	assertRejected(
		"soft-deleted payment",
		"UPDATE payments SET deleted_at = now() WHERE id = $1",
		[]any{paymentID},
		"UPDATE payments SET deleted_at = NULL WHERE id = $1",
		[]any{paymentID},
	)
	assertRejected(
		"completed billing without payment",
		"UPDATE payments SET billing_id = $1 WHERE id = $2",
		[]any{billingID + 1, paymentID},
		"UPDATE payments SET billing_id = $1 WHERE id = $2",
		[]any{billingID, paymentID},
	)

	var planJSON []byte
	if err := tx.QueryRow(
		ctx,
		"EXPLAIN (FORMAT JSON, COSTS FALSE) "+verifyCutoverPaymentGraphQuery,
		manifest.IDBand.NonOwnerIDOffset,
		manifest.IDBand.EndExclusive,
		clinicID,
		cashID,
		cardID,
	).Scan(&planJSON); err != nil {
		t.Fatalf("explain payment graph query: %v", err)
	}
	if len(planJSON) == 0 {
		t.Fatal("payment graph EXPLAIN returned an empty plan")
	}
}
