package model_test

// clinic_operating_hours_realdb_test.go — real-DB fixture for the common
// clinic operating-hours storage (the former
// backend/migrations/008_clinic_operating_hours.sql, EMR-446/466 A03,
// contract §2.1). The DDL lives in 001_init.sql since the 2026-10-23
// consolidation; testdb.SetupIsolated001TestDB applies it. NOT_RUN yet in
// this workspace.
//
// Shared realdb helpers (lineTimeSevenWeekdayJSON,
// lineTimeSeedClinic, lineTimeSeedBinding, lineTimeRequirePgCode) live in this
// earliest stack file; ③/④ realdb tests reuse them.
//
// RLS note: apply_rls_policy ENABLEs (not FORCEs) RLS and the fixture connects
// as the table owner, so visibility is checked through an unprivileged role;
// it is skipped when the environment role lacks CREATEROLE.

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/testdb"
)

// lineTimeSevenWeekdayJSON is a contract-shaped 7-element weekday array (all
// days explicit, empty = closed) shared by the common/policy fixtures.
const lineTimeSevenWeekdayJSON = `[{"weekday":0,"intervals":[]},{"weekday":1,"intervals":[]},{"weekday":2,"intervals":[]},{"weekday":3,"intervals":[]},{"weekday":4,"intervals":[]},{"weekday":5,"intervals":[]},{"weekday":6,"intervals":[]}]`

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
	// 001_init.sql carries the former 008 DDL (2026-10-23 consolidation).
	db := testdb.SetupIsolated001TestDB(t)
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

	// Negative storage cases. Every statement must be rejected, and none may
	// leave a row behind. clinicC never receives a valid row, so a stray insert
	// is detectable; the duplicate-PK case targets clinicA's existing row.
	clinicC := lineTimeSeedClinic(t, db, "008-common-c")
	sixDays := `[{"weekday":0,"intervals":[]},{"weekday":1,"intervals":[]},{"weekday":2,"intervals":[]},{"weekday":3,"intervals":[]},{"weekday":4,"intervals":[]},{"weekday":5,"intervals":[]}]`
	eightDays := lineTimeSevenWeekdayJSON[:len(lineTimeSevenWeekdayJSON)-1] + `,{"weekday":7,"intervals":[]}]`
	validWeekly := `'` + lineTimeSevenWeekdayJSON + `'`
	rejectedInserts := []struct {
		name   string
		code   string
		values string
	}{
		{"weekly_hours with 6 elements", "23514", `'clinic-operating-hours/v1', DEFAULT, '` + sixDays + `', '[]', false`},
		{"weekly_hours with 8 elements", "23514", `'clinic-operating-hours/v1', DEFAULT, '` + eightDays + `', '[]', false`},
		{"weekly_hours empty array", "23514", `'clinic-operating-hours/v1', DEFAULT, '[]', '[]', false`},
		{"closed_dates not an array", "23514", `'clinic-operating-hours/v1', DEFAULT, ` + validWeekly + `, '{}', false`},
		{"version zero", "23514", `'clinic-operating-hours/v1', 0, ` + validWeekly + `, '[]', false`},
		{"wrong schema_version marker", "23514", `'clinic-operating-hours/v2', DEFAULT, ` + validWeekly + `, '[]', false`},
		{"NULL weekly_hours", "23502", `'clinic-operating-hours/v1', DEFAULT, NULL, '[]', false`},
		{"NULL closed_dates", "23502", `'clinic-operating-hours/v1', DEFAULT, ` + validWeekly + `, NULL, false`},
		{"NULL national_holiday_closed", "23502", `'clinic-operating-hours/v1', DEFAULT, ` + validWeekly + `, '[]', NULL`},
	}
	for _, tc := range rejectedInserts {
		t.Run(tc.name, func(t *testing.T) {
			lineTimeRequirePgCode(t,
				db.Exec(`
					INSERT INTO clinic_operating_hours
						(clinic_id, schema_version, version, weekly_hours, closed_dates, national_holiday_closed)
					VALUES (?, `+tc.values+`)`, clinicC).Error,
				tc.code)
		})
	}

	// national_holiday_closed has no DEFAULT: omitting it is NOT NULL (23502).
	lineTimeRequirePgCode(t,
		db.Exec(`
			INSERT INTO clinic_operating_hours (clinic_id, schema_version, weekly_hours, closed_dates)
			VALUES (?, 'clinic-operating-hours/v1', ?, '[]')`,
			clinicC, lineTimeSevenWeekdayJSON).Error,
		"23502")

	// clinic_id is the PK: a second row for the same clinic is rejected (23505).
	lineTimeRequirePgCode(t,
		db.Exec(`
			INSERT INTO clinic_operating_hours
				(clinic_id, schema_version, weekly_hours, closed_dates, national_holiday_closed)
			VALUES (?, 'clinic-operating-hours/v1', ?, '[]', false)`,
			clinicA, lineTimeSevenWeekdayJSON).Error,
		"23505")

	require.NoError(t, db.Raw(`SELECT count(*) FROM clinic_operating_hours WHERE clinic_id = ?`, clinicC).Scan(&count).Error)
	require.Zero(t, count, "rejected inserts must leave no row for clinic C")
	require.NoError(t, db.Raw(`SELECT count(*) FROM clinic_operating_hours WHERE clinic_id = ?`, clinicA).Scan(&count).Error)
	require.EqualValues(t, 1, count, "duplicate-PK rejection must keep clinic A at exactly one row")

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
	// Roles are cluster-global (pg_authid), not per-database, so a leftover
	// from an earlier run must not collide: the probe role name is unique per
	// test run (same convention as disposable001DatabaseName).
	role := fmt.Sprintf("rls_probe_008_%d", time.Now().UnixNano())
	if _, err := conn.ExecContext(ctx, fmt.Sprintf(`CREATE ROLE %s NOLOGIN`, role)); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "42501" {
			t.Skipf("CREATEROLE unavailable in this environment; RLS visibility check skipped: %v", err)
		}
		t.Fatalf("create RLS probe role %s: %v", role, err)
	}
	// The role lives in the shared cluster catalog and outlives the disposable
	// DB drop. DROP ROLE fails while the GRANT below still exists, so cleanup
	// must revoke privileges first, then drop; failures are reported, not
	// swallowed.
	t.Cleanup(func() {
		cleanupConn, cerr := sqlDB.Conn(ctx)
		if cerr != nil {
			t.Errorf("RLS probe cleanup: acquire owner connection: %v", cerr)
			return
		}
		defer cleanupConn.Close()
		if _, cerr := cleanupConn.ExecContext(ctx, fmt.Sprintf(`REVOKE ALL ON clinic_operating_hours FROM %s`, role)); cerr != nil {
			t.Errorf("RLS probe cleanup: revoke privileges from %s: %v", role, cerr)
		}
		if _, cerr := cleanupConn.ExecContext(ctx, fmt.Sprintf(`DROP ROLE IF EXISTS %s`, role)); cerr != nil {
			t.Errorf("RLS probe cleanup: drop role %s: %v", role, cerr)
		}
	})
	_, err = conn.ExecContext(ctx, fmt.Sprintf(`GRANT SELECT ON clinic_operating_hours TO %s`, role))
	require.NoError(t, err)
	// Probe inside a dedicated transaction: SET LOCAL ROLE and a local
	// set_config never outlive the rollback, so the pooled session keeps no
	// role/GUC residue when the connection is returned.
	probeTx, err := conn.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer func() { _ = probeTx.Rollback() }()
	_, err = probeTx.ExecContext(ctx, fmt.Sprintf(`SET LOCAL ROLE %s`, role))
	require.NoError(t, err)
	_, err = probeTx.ExecContext(ctx, fmt.Sprintf(`SELECT set_config('app.current_clinic_ids', '%d', true)`, clinicA))
	require.NoError(t, err)
	var visible int64
	require.NoError(t, probeTx.QueryRowContext(ctx, `SELECT count(*) FROM clinic_operating_hours`).Scan(&visible))
	require.EqualValues(t, 1, visible, "other-clinic row must be invisible under the clinic-scope policy")
}
