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
// PostgreSQL without changing data. It is opt-in because the shared Compose DB
// requires the project-wide DB lease even for read-only verification.
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

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	if err := validateCutoverUniqueIndexes(ctx, tx); err != nil {
		t.Fatal(err)
	}

	emptyBand := CutoverManifest{IDBand: CutoverIDBand{
		NonOwnerIDOffset: 9_000_000_000_000_000_000,
		EndExclusive:     9_000_000_000_000_000_001,
	}}
	if err := verifyCutoverPaymentGraph(ctx, tx, emptyBand, CutoverSeedIDs{
		ClinicID: 1, CashPaymentMethodID: 1, CreditCardPaymentMethodID: 2,
	}); err != nil {
		t.Fatal(err)
	}

	var planJSON []byte
	if err := tx.QueryRow(
		ctx,
		"EXPLAIN (FORMAT JSON, COSTS FALSE) "+verifyCutoverPaymentGraphQuery,
		emptyBand.IDBand.NonOwnerIDOffset,
		emptyBand.IDBand.EndExclusive,
		int64(1),
		int64(1),
		int64(2),
	).Scan(&planJSON); err != nil {
		t.Fatalf("explain payment graph query: %v", err)
	}
	if len(planJSON) == 0 {
		t.Fatal("payment graph EXPLAIN returned an empty plan")
	}
}
