package model_test

// line_staff_start_policy_realdb_test.go — real-DB fixture for
// backend/migrations/010_line_staff_start_policy_and_reservation_anchor.sql
// (EMR-446/466 A02 + EMR-492 anchor, contract §4.1/§4.3). Applies 001 via
// testdb.SetupIsolated001TestDB and Execs the 010 body in-test (no dependency
// on 008/009). NOT_RUN yet. Shared helpers live in
// clinic_operating_hours_realdb_test.go (stack base).

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/testdb"
)

// lineTimePolicyWeeklyTimesJSON is the contract §4.1 weekly_times shape:
// [{weekday,times}] with exactly one entry per weekday 0..6 in ascending order;
// times are unique ascending HH:MM (empty = explicitly no registered starts).
// It intentionally differs from lineTimeSevenWeekdayJSON, the {weekday,
// intervals} shape for clinic_operating_hours.weekly_hours (§2.1).
const lineTimePolicyWeeklyTimesJSON = `[{"weekday":0,"times":[]},{"weekday":1,"times":["09:00","09:30"]},{"weekday":2,"times":[]},{"weekday":3,"times":[]},{"weekday":4,"times":[]},{"weekday":5,"times":[]},{"weekday":6,"times":[]}]`

func TestLineStaffStartPolicyAndAnchorRealDB(t *testing.T) {
	db := testdb.SetupIsolated001TestDB(t)
	require.NoError(t, db.Exec(lineTimeReadMigration(t, "010_line_staff_start_policy_and_reservation_anchor.sql")).Error,
		"010 must apply cleanly on top of 001")

	t.Run("staff start policy", func(t *testing.T) {
		clinicA := lineTimeSeedClinic(t, db, "010-policy-a")
		clinicB := lineTimeSeedClinic(t, db, "010-policy-b")

		var typeA, assignA, assignA2, typeB, assignB int64
		staffA := testdb.MakeDoctor(t, db, uint64(clinicA), "010 staff A")
		staffA2 := testdb.MakeDoctor(t, db, uint64(clinicA), "010 staff A2")
		require.NoError(t, db.Raw(`INSERT INTO reservation_types (clinic_id, name, duration_minutes) VALUES (?, '010 type A', 15) RETURNING id`, clinicA).Scan(&typeA).Error)
		require.NoError(t, db.Raw(`INSERT INTO staff_clinic_assignments (staff_id, clinic_id, is_main) VALUES (?, ?, true) RETURNING id`, staffA.ID, clinicA).Scan(&assignA).Error)
		require.NoError(t, db.Raw(`INSERT INTO staff_clinic_assignments (staff_id, clinic_id, is_main) VALUES (?, ?, true) RETURNING id`, staffA2.ID, clinicA).Scan(&assignA2).Error)
		require.NoError(t, db.Raw(`INSERT INTO reservation_types (clinic_id, name, duration_minutes) VALUES (?, '010 type B', 15) RETURNING id`, clinicB).Scan(&typeB).Error)
		staffB := testdb.MakeDoctor(t, db, uint64(clinicB), "010 staff B")
		require.NoError(t, db.Raw(`INSERT INTO staff_clinic_assignments (staff_id, clinic_id, is_main) VALUES (?, ?, true) RETURNING id`, staffB.ID, clinicB).Scan(&assignB).Error)

		require.NoError(t, db.Exec(`
			INSERT INTO line_staff_start_policies
				(clinic_id, reservation_type_id, assignment_id, staff_id, schema_version, weekly_times, date_overrides)
			VALUES (?, ?, ?, ?, 'trimming-registered-starts/v1', ?, '[]')`,
			clinicA, typeA, assignA, staffA.ID, lineTimePolicyWeeklyTimesJSON).Error)

		var version int64
		require.NoError(t, db.Raw(`SELECT version FROM line_staff_start_policies WHERE clinic_id = ?`, clinicA).Scan(&version).Error)
		require.EqualValues(t, 1, version)

		// Change advances the version; tombstoning advances it again.
		require.NoError(t, db.Exec(`UPDATE line_staff_start_policies SET date_overrides = '[{"date":"2026-11-01","times":[]}]' WHERE clinic_id = ?`, clinicA).Error)
		require.NoError(t, db.Exec(`UPDATE line_staff_start_policies SET deleted_at = clock_timestamp() WHERE clinic_id = ?`, clinicA).Error)
		require.NoError(t, db.Raw(`SELECT version FROM line_staff_start_policies WHERE clinic_id = ?`, clinicA).Scan(&version).Error)
		require.EqualValues(t, 3, version, "change + tombstone each advance the version once")

		// Physical delete is forbidden.
		lineTimeRequirePgCode(t,
			db.Exec(`DELETE FROM line_staff_start_policies WHERE clinic_id = ?`, clinicA).Error,
			"23514")

		// An assignment/staff mismatch violates the composite FK.
		lineTimeRequirePgCode(t,
			db.Exec(`
				INSERT INTO line_staff_start_policies
					(clinic_id, reservation_type_id, assignment_id, staff_id, schema_version, weekly_times, date_overrides)
				VALUES (?, ?, ?, ?, 'trimming-registered-starts/v1', ?, '[]')`,
				clinicA, typeA, assignA2, staffA.ID, lineTimePolicyWeeklyTimesJSON).Error,
			"23503")

		// A type/assignment of another clinic violates the composite FKs.
		lineTimeRequirePgCode(t,
			db.Exec(`
				INSERT INTO line_staff_start_policies
					(clinic_id, reservation_type_id, assignment_id, staff_id, schema_version, weekly_times, date_overrides)
				VALUES (?, ?, ?, ?, 'trimming-registered-starts/v1', ?, '[]')`,
				clinicA, typeB, assignA, staffA.ID, lineTimePolicyWeeklyTimesJSON).Error,
			"23503")
		lineTimeRequirePgCode(t,
			db.Exec(`
				INSERT INTO line_staff_start_policies
					(clinic_id, reservation_type_id, assignment_id, staff_id, schema_version, weekly_times, date_overrides)
				VALUES (?, ?, ?, ?, 'trimming-registered-starts/v1', ?, '[]')`,
				clinicA, typeA, assignB, staffA.ID, lineTimePolicyWeeklyTimesJSON).Error,
			"23503")
	})

	t.Run("reservation master operation anchors", func(t *testing.T) {
		clinic := lineTimeSeedClinic(t, db, "010-anchor")
		binding := lineTimeSeedBinding(t, db, clinic, "line-start-policy.create")

		// type/group anchors must have staff/assignment NULL.
		require.NoError(t, db.Exec(`
			INSERT INTO reservation_master_operation_anchors (operation_binding_id, clinic_id, target_kind, target_id)
			VALUES (?, ?, 'type', 1)`, binding, clinic).Error)
		lineTimeRequirePgCode(t,
			db.Exec(`
				INSERT INTO reservation_master_operation_anchors (operation_binding_id, clinic_id, target_kind, staff_id)
				VALUES (?, ?, 'group', 9)`, lineTimeSeedBinding(t, db, clinic, "reservation-type.create"), clinic).Error,
			"23514")

		// line_start_policy anchors require target/staff/assignment ids.
		lineTimeRequirePgCode(t,
			db.Exec(`
				INSERT INTO reservation_master_operation_anchors (operation_binding_id, clinic_id, target_kind, target_id)
				VALUES (?, ?, 'line_start_policy', 1)`, lineTimeSeedBinding(t, db, clinic, "line-start-policy.create"), clinic).Error,
			"23514")

		// UPDATE and DELETE are refused (immutable claim-time anchor).
		lineTimeRequirePgCode(t,
			db.Exec(`UPDATE reservation_master_operation_anchors SET target_id = 2 WHERE operation_binding_id = ?`, binding).Error,
			"23514")
		lineTimeRequirePgCode(t,
			db.Exec(`DELETE FROM reservation_master_operation_anchors WHERE operation_binding_id = ?`, binding).Error,
			"23514")

		// The anchor references a real (binding, clinic) pair.
		lineTimeRequirePgCode(t,
			db.Exec(`
				INSERT INTO reservation_master_operation_anchors (operation_binding_id, clinic_id, target_kind)
				VALUES (?, ?, 'type')`, binding+999999, clinic).Error,
			"23503")
	})
}
