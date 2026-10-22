package model_test

// clinic_operating_hours_realdb_test.go — real-DB fixture for
// backend/migrations/008_clinic_operating_hours.sql (EMR-446/466 A03,
// contract §2.1). Applies 001 via testdb.SetupIsolated001TestDB and Execs the
// 008 body in-test. NOT_RUN yet in this workspace.
//
// Shared realdb helpers (lineTimeSevenWeekdayJSON, lineTimeReadMigration,
// lineTimeSeedClinic, lineTimeSeedBinding, lineTimeRequirePgCode) live in this
// earliest stack file; ③/④ realdb tests reuse them.
//
// RLS note: apply_rls_policy ENABLEs (not FORCEs) RLS and the fixture connects
// as the table owner, so visibility is checked through an unprivileged role;
// it is skipped when the environment role lacks CREATEROLE.

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/testdb"
)

// lineTimeSevenWeekdayJSON is a contract-shaped 7-element weekday array (all
// days explicit, empty = closed) shared by the common/policy fixtures.
const lineTimeSevenWeekdayJSON = `[{"weekday":0,"intervals":[]},{"weekday":1,"intervals":[]},{"weekday":2,"intervals":[]},{"weekday":3,"intervals":[]},{"weekday":4,"intervals":[]},{"weekday":5,"intervals":[]},{"weekday":6,"intervals":[]}]`

func lineTimeReadMigration(t *testing.T, name string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "migrations", name))
	require.NoError(t, err)
	return string(raw)
}

func lineTimeSeedClinic(t *testing.T, db *gorm.DB, name string) int64 {
	t.Helper()
	var companyID, clinicID int64
	require.NoError(t, db.Raw(`INSERT INTO companies (name) VALUES (?) RETURNING id`, name).Scan(&companyID).Error)
	require.NoError(t, db.Raw(`INSERT INTO clinics (company_id, name) VALUES (?, ?) RETURNING id`, companyID, name).Scan(&clinicID).Error)
	return clinicID
}

func lineTimeSeedBinding(t *testing.T, db *gorm.DB, clinicID int64, intent string) int64 {
	t.Helper()
	var bindingID int64
	require.NoError(t, db.Raw(`
		INSERT INTO operation_bindings
			(clinic_id, owner_kind, actor_kind, actor_ref, intent, operation_key, fingerprint, fingerprint_key_version)
		VALUES (?, 'reservation', 'staff', 'fixture-linetime', ?, gen_random_uuid(), decode(repeat('00', 32), 'hex'), 'v1')
		RETURNING id`, clinicID, intent).Scan(&bindingID).Error)
	return bindingID
}

func lineTimeRequirePgCode(t *testing.T, err error, code string) {
	t.Helper()
	require.Error(t, err)
	var pgErr *pgconn.PgError
	require.ErrorAs(t, err, &pgErr)
	require.Equal(t, code, pgErr.Code)
}

func TestClinicOperatingHoursRealDB(t *testing.T) {
	db := testdb.SetupIsolated001TestDB(t)
	require.NoError(t, db.Exec(lineTimeReadMigration(t, "008_clinic_operating_hours.sql")).Error,
		"008 must apply cleanly on top of 001")
	ctx := context.Background()

	clinicA := lineTimeSeedClinic(t, db, "008-common-a")
	clinicB := lineTimeSeedClinic(t, db, "008-common-b")

	var count int64
	require.NoError(t, db.Raw(`SELECT count(*) FROM clinic_operating_hours WHERE clinic_id = ?`, clinicA).Scan(&count).Error)
	require.Zero(t, count, "absent row = unconfigured")

	// First write must be fully explicit (no settings defaults exist).
	require.NoError(t, db.Exec(`
		INSERT INTO clinic_operating_hours
			(clinic_id, schema_version, weekly_hours, closed_dates, national_holiday_closed)
		VALUES (?, 'clinic-operating-hours/v1', ?, '[]', false)`,
		clinicA, lineTimeSevenWeekdayJSON).Error)

	var version int64
	require.NoError(t, db.Raw(`SELECT version FROM clinic_operating_hours WHERE clinic_id = ?`, clinicA).Scan(&version).Error)
	require.EqualValues(t, 1, version, "first version is the CAS origin")

	// True no-op keeps the same version.
	require.NoError(t, db.Exec(`UPDATE clinic_operating_hours SET closed_dates = closed_dates WHERE clinic_id = ?`, clinicA).Error)
	require.NoError(t, db.Raw(`SELECT version FROM clinic_operating_hours WHERE clinic_id = ?`, clinicA).Scan(&version).Error)
	require.EqualValues(t, 1, version, "no-op must not advance the version")

	// A real change advances it exactly once.
	require.NoError(t, db.Exec(`UPDATE clinic_operating_hours SET national_holiday_closed = true WHERE clinic_id = ?`, clinicA).Error)
	require.NoError(t, db.Raw(`SELECT version FROM clinic_operating_hours WHERE clinic_id = ?`, clinicA).Scan(&version).Error)
	require.EqualValues(t, 2, version, "changed write advances the version once")

	// Storage maintains version: an app SET is rejected (23514).
	lineTimeRequirePgCode(t,
		db.Exec(`UPDATE clinic_operating_hours SET version = 99 WHERE clinic_id = ?`, clinicA).Error,
		"23514")

	// RLS policy exists.
	var policies int64
	require.NoError(t, db.Raw(`
		SELECT count(*) FROM pg_policies
		WHERE schemaname = 'public' AND tablename = 'clinic_operating_hours'
		  AND policyname = 'clinic_operating_hours_clinic_scope'`).Scan(&policies).Error)
	require.EqualValues(t, 1, policies)

	// A row of clinic B is invisible to a role scoped to clinic A.
	require.NoError(t, db.Exec(`
		INSERT INTO clinic_operating_hours
			(clinic_id, schema_version, weekly_hours, closed_dates, national_holiday_closed)
		VALUES (?, 'clinic-operating-hours/v1', ?, '[]', false)`,
		clinicB, lineTimeSevenWeekdayJSON).Error)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	conn, err := sqlDB.Conn(ctx)
	require.NoError(t, err)
	defer conn.Close()
	const role = "rls_probe_008"
	if _, err := conn.ExecContext(ctx, fmt.Sprintf(`CREATE ROLE %s NOLOGIN`, role)); err != nil {
		t.Skipf("CREATEROLE unavailable in this environment; RLS visibility check skipped: %v", err)
	}
	t.Cleanup(func() {
		cleanupConn, cerr := sqlDB.Conn(ctx)
		if cerr == nil {
			defer cleanupConn.Close()
			_, _ = cleanupConn.ExecContext(ctx, fmt.Sprintf(`DROP ROLE IF EXISTS %s`, role))
		}
	})
	_, err = conn.ExecContext(ctx, fmt.Sprintf(`GRANT SELECT ON clinic_operating_hours TO %s`, role))
	require.NoError(t, err)
	_, err = conn.ExecContext(ctx, fmt.Sprintf(`SET ROLE %s`, role))
	require.NoError(t, err)
	_, err = conn.ExecContext(ctx, fmt.Sprintf(`SET app.current_clinic_ids = '%d'`, clinicA))
	require.NoError(t, err)
	var visible int64
	require.NoError(t, conn.QueryRowContext(ctx, `SELECT count(*) FROM clinic_operating_hours`).Scan(&visible))
	require.EqualValues(t, 1, visible, "other-clinic row must be invisible under the clinic-scope policy")
}
