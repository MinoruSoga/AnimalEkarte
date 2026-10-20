package testdb

import (
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

// TestSetupIsolated001TestDBSchemaInvariants は 001_init.sql を source にした
// fixture が実 001 の不変条件（clinic スコープ FK・RLS policy・deferred
// constraint trigger）を再現し、取引分離（deferred FK 違反は commit 時に失敗 /
// rollback 後の行は残らない）を持つことを最小範囲で確認する。
//
// DB スタックが必要（-short では既存の skipShortDatabase 慣習で skip）。
func TestSetupIsolated001TestDBSchemaInvariants(t *testing.T) {
	db := SetupIsolated001TestDB(t)

	t.Run("clinic scoped foreign keys exist", func(t *testing.T) {
		for _, table := range []string{"owners", "billings", "appointments"} {
			var count int64
			err := db.Raw(`
				SELECT count(*)
				FROM pg_constraint
				WHERE contype = 'f'
				  AND conrelid = to_regclass(?)
				  AND confrelid = 'clinics'::regclass`, table).Scan(&count).Error
			require.NoErrorf(t, err, "query FK constraints for %s", table)
			require.Greaterf(t, count, int64(0), "%s must keep a FK to clinics(id)", table)
		}
	})

	t.Run("row level security is enabled with a policy", func(t *testing.T) {
		var rlsEnabled bool
		err := db.Raw(`SELECT relrowsecurity FROM pg_class WHERE oid = 'owners'::regclass`).Scan(&rlsEnabled).Error
		require.NoError(t, err)
		require.True(t, rlsEnabled, "owners must keep RLS enabled (001 tenant isolation)")

		var policies int64
		err = db.Raw(`
			SELECT count(*)
			FROM pg_policies
			WHERE schemaname = 'public'
			  AND tablename = 'owners'
			  AND policyname = 'tenant_clinic_id_isolation'`).Scan(&policies).Error
		require.NoError(t, err)
		require.EqualValues(t, 1, policies, "owners must keep tenant_clinic_id_isolation policy")
	})

	t.Run("deferred constraint triggers exist", func(t *testing.T) {
		var count int64
		err := db.Raw(`
			SELECT count(*)
			FROM pg_trigger
			WHERE tgname IN ('trimming_booking_snapshot_complete', 'deposit_operation_complete')
			  AND tgdeferrable
			  AND tginitdeferred`).Scan(&count).Error
		require.NoError(t, err)
		require.EqualValues(t, 2, count, "001 deferred constraint triggers must exist")
	})

	t.Run("deferred foreign key violation fails at commit", func(t *testing.T) {
		tx := db.Begin()
		require.NoError(t, tx.Error)
		defer func() { _ = tx.Rollback() }()

		var companyID, clinicID, ownerID, staffID, bindingID int64
		require.NoError(t, tx.Raw(`INSERT INTO companies (name) VALUES ('smoke-001') RETURNING id`).Scan(&companyID).Error)
		require.NoError(t, tx.Raw(`INSERT INTO clinics (company_id, name) VALUES (?, 'smoke-001') RETURNING id`, companyID).Scan(&clinicID).Error)
		require.NoError(t, tx.Raw(`INSERT INTO owners (clinic_id) VALUES (?) RETURNING id`, clinicID).Scan(&ownerID).Error)
		require.NoError(t, tx.Raw(`INSERT INTO staffs (clinic_id, name) VALUES (?, 'smoke-001') RETURNING id`, clinicID).Scan(&staffID).Error)
		require.NoError(t, tx.Raw(`
			INSERT INTO operation_bindings
				(clinic_id, owner_kind, actor_kind, actor_ref, intent, operation_key, fingerprint, fingerprint_key_version)
			VALUES (?, 'owner', 'staff', 'smoke-001', 'smoke-001', gen_random_uuid(), decode(repeat('00', 32), 'hex'), 'v1')
			RETURNING id`, clinicID).Scan(&bindingID).Error)
		// operation_success_required() を満たす receipt を同じ tx に置く。
		require.NoError(t, tx.Exec(`
			INSERT INTO operation_results (clinic_id, binding_id, state, result_refs, committed_versions, effect_summary)
			VALUES (?, ?, 'succeeded', '{}', '{}', '{}')`, clinicID, bindingID).Error)

		// common_owner_id は DEFERRABLE INITIALLY DEFERRED の FK。存在しない UUID を
		// 入れても INSERT 時点では成功し、commit 時に 23503 で失敗する。
		require.NoError(t, tx.Exec(`
			INSERT INTO common_owner_clinic_bindings
				(clinic_id, common_owner_id, local_owner_id, confirmed_by, confirmed_at, operation_id)
			VALUES (?, '00000000-0000-0000-0000-000000000000', ?, ?, now(), ?)`,
			clinicID, ownerID, staffID, bindingID).Error)

		err := tx.Commit().Error
		require.Error(t, err, "deferred FK to a missing common_owner_subjects row must fail at commit")
		var pgErr *pgconn.PgError
		require.ErrorAs(t, err, &pgErr, "commit failure must be a PostgreSQL error")
		require.Equal(t, "23503", pgErr.Code, "commit must fail on foreign_key_violation (deferred FK)")
	})

	t.Run("rollback leaves no row", func(t *testing.T) {
		tx := db.Begin()
		require.NoError(t, tx.Error)
		require.NoError(t, tx.Exec(`INSERT INTO companies (name) VALUES ('smoke-001-rollback')`).Error)
		require.NoError(t, tx.Rollback().Error)

		var count int64
		require.NoError(t, db.Raw(`SELECT count(*) FROM companies WHERE name = 'smoke-001-rollback'`).Scan(&count).Error)
		require.Zero(t, count, "rolled back insert must not survive")
	})
}
