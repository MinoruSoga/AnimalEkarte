package billing

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Opt-in disposable integration: MIGRATE_SQL_INTEGRATION=1 + DATABASE_URL.
// EMR-210 — runs the teardown against a migrate-built disposable database so
// the real RESTRICT/composite FKs and append-only triggers are exercised.
// testdb (AutoMigrate) has neither, so only this path proves the delete series
// is closed under the production FK graph.
//
// DATABASE_URL must be a postgres:// URL whose database name matches the
// disposable csvimport21_* / animalekarte_f8_g4_* rehearsal pattern — the same
// safety convention cmd/migrate's disposable integration tests enforce.

var disposableTeardownDBNameRe = regexp.MustCompile(`^(?:csvimport21|animalekarte_f8_g4)_[a-z0-9_]+$`)

func openTeardownIntegrationDB(t *testing.T) *gorm.DB {
	t.Helper()
	if os.Getenv("MIGRATE_SQL_INTEGRATION") != "1" {
		t.Skip("set MIGRATE_SQL_INTEGRATION=1 + DATABASE_URL for the disposable teardown test")
	}
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Fatal("DATABASE_URL is required for the disposable teardown test")
	}
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("parse DATABASE_URL: %v", err)
	}
	dbName := strings.TrimPrefix(parsed.Path, "/")
	if !disposableTeardownDBNameRe.MatchString(dbName) {
		t.Fatalf("DATABASE_URL dbname %q is not a disposable csvimport21_* / animalekarte_f8_g4_* database", dbName)
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect: %v", err)
	}

	// Prove the schema is migrate-built: the append-only trigger only exists
	// after the real DDL path ran. Without it the assertions below cannot
	// distinguish "FK graph satisfied" from "FK graph absent".
	var hasImmutableTrigger bool
	require.NoError(t, db.Raw(
		"SELECT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_cash_register_closes_immutable')",
	).Scan(&hasImmutableTrigger).Error)
	require.True(t, hasImmutableTrigger,
		"DATABASE_URL must point at a migrate-built disposable database (trg_cash_register_closes_immutable missing)")
	return db
}

func teardownIntegrationFixture(t *testing.T, db *gorm.DB) *SyntheticClosingResult {
	t.Helper()
	ctx := context.Background()
	jst, err := time.LoadLocation("Asia/Tokyo")
	require.NoError(t, err)
	got, err := CreateSyntheticClosingFixture(ctx, db, SyntheticClosingRequest{
		AppEnv: "development", DBHost: "db",
		TargetDate: time.Date(2026, 9, 7, 0, 0, 0, 0, jst), PasswordHash: "x",
	})
	require.NoError(t, err)
	return got
}

// The S09 close ledger UAT produced: one append-only close plus an adjustment
// whose composite RESTRICT FKs (close_id+clinic_id, billing_id+clinic_id,
// actor_id+clinic_id) aborted teardown before EMR-210.
func TestSyntheticClosingTeardown_CompletesWithCashRegisterCloseGraph(t *testing.T) {
	db := openTeardownIntegrationDB(t)
	ctx := context.Background()
	got := teardownIntegrationFixture(t, db)

	var staffID uint64
	require.NoError(t, db.Raw("SELECT id FROM staffs WHERE clinic_id = ? LIMIT 1", got.ClinicID).Scan(&staffID).Error)
	require.NotZero(t, staffID)

	var closeID uint64
	require.NoError(t, db.Raw(`
INSERT INTO cash_register_closes (clinic_id, close_date, period, closed_by)
VALUES (?, CURRENT_DATE, 'am', ?)
RETURNING id`, got.ClinicID, staffID).Scan(&closeID).Error)
	require.NotZero(t, closeID)
	require.NoError(t, db.Exec(`
INSERT INTO cash_register_close_adjustments (clinic_id, close_id, billing_id, reason, actor_id)
VALUES (?, ?, ?, 's09 teardown regression', ?)`,
		got.ClinicID, closeID, got.BillingIDs[0], staffID).Error)

	require.NoError(t, DeleteSyntheticClosingFixture(ctx, db, "development", "db", got.ClinicID, got.CleanupToken))

	for _, table := range []string{
		"cash_register_close_adjustments", "cash_register_closes",
		"payment_splits", "payments", "billing_items", "billings",
		"payment_methods", "pets", "owners", "staffs",
	} {
		var remaining int64
		require.NoError(t, db.Raw(
			fmt.Sprintf("SELECT count(*) FROM %s WHERE clinic_id = ?", table), got.ClinicID,
		).Scan(&remaining).Error)
		assert.Zero(t, remaining, "%s rows must be removed", table)
	}
	var clinics int64
	require.NoError(t, db.Raw("SELECT count(*) FROM clinics WHERE id = ?", got.ClinicID).Scan(&clinics).Error)
	assert.Zero(t, clinics)

	// The append-only triggers must be live again after teardown commits —
	// DISABLE TRIGGER USER inside the tx must not leak. tgenabled 'O' means the
	// trigger fires in the origin (normal) session role again.
	var stillDisabled []string
	require.NoError(t, db.Raw(`
SELECT tgname FROM pg_trigger
WHERE tgname IN (
	'trg_cash_register_closes_immutable',
	'trg_cash_register_close_adjustments_immutable',
	'trg_examination_revisions_immutable',
	'trg_examination_revision_items_immutable',
	'trg_lab_import_exam_retractions_immutable',
	'trg_lab_import_exam_retraction_items_immutable',
	'trg_lab_import_usage_receipts_immutable',
	'trg_lab_import_revert_receipts_immutable'
) AND tgenabled <> 'O'`).Scan(&stillDisabled).Error)
	assert.Empty(t, stillDisabled, "append-only triggers must be re-enabled after teardown")
}

// audit_logs handling is an EMR-211 product decision: with the default policy
// the RESTRICT FKs (actor_id → staffs, clinic_id → clinics) keep teardown
// fail-closed — this test pins that BLOCKED contract — while the injected
// policy seam must be able to resolve the rows inside the same transaction.
func TestSyntheticClosingTeardown_AuditRowsStayBlockingByDefault(t *testing.T) {
	db := openTeardownIntegrationDB(t)
	ctx := context.Background()
	got := teardownIntegrationFixture(t, db)

	var staffID uint64
	require.NoError(t, db.Raw("SELECT id FROM staffs WHERE clinic_id = ? LIMIT 1", got.ClinicID).Scan(&staffID).Error)
	require.NoError(t, db.Exec(`
INSERT INTO audit_logs (clinic_id, actor_id, actor_type, action, resource)
VALUES (?, ?, 'staff', 'login', 'session')`, got.ClinicID, staffID).Error)

	// Default policy: audit rows are preserved and teardown stays fail-closed.
	err := DeleteSyntheticClosingFixture(ctx, db, "development", "db", got.ClinicID, got.CleanupToken)
	require.Error(t, err, "audit rows must keep teardown fail-closed until EMR-211 decides their handling")

	// Rollback left the fixture intact: staff, audit row, and clinic remain.
	var remaining int64
	require.NoError(t, db.Raw("SELECT count(*) FROM audit_logs WHERE clinic_id = ?", got.ClinicID).Scan(&remaining).Error)
	assert.Equal(t, int64(1), remaining)
	require.NoError(t, db.Raw("SELECT count(*) FROM staffs WHERE clinic_id = ?", got.ClinicID).Scan(&remaining).Error)
	assert.Positive(t, remaining)
	require.NoError(t, db.Raw("SELECT count(*) FROM clinics WHERE id = ?", got.ClinicID).Scan(&remaining).Error)
	assert.Equal(t, int64(1), remaining)

	// The EMR-211 seam: a policy that deletes the audit rows lets the same
	// teardown complete — without shipping a default that decides audit fate.
	require.NoError(t, DeleteSyntheticClosingFixtureWithAuditPolicy(ctx, db, "development", "db", got.ClinicID, got.CleanupToken,
		func(_ context.Context, tx *gorm.DB, clinicID uint64) error {
			return tx.Exec("DELETE FROM audit_logs WHERE clinic_id = ?", clinicID).Error
		}))
	require.NoError(t, db.Raw("SELECT count(*) FROM clinics WHERE id = ?", got.ClinicID).Scan(&remaining).Error)
	assert.Zero(t, remaining)
}
