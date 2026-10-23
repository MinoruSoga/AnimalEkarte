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

		// Create cannot bypass a tombstone: the tombstoned row keeps its PK, so a
		// second INSERT for the same (clinic, type, assignment) is a duplicate key.
		lineTimeRequirePgCode(t,
			db.Exec(`
				INSERT INTO line_staff_start_policies
					(clinic_id, reservation_type_id, assignment_id, staff_id, schema_version, weekly_times, date_overrides)
				VALUES (?, ?, ?, ?, 'trimming-registered-starts/v1', ?, '[]')`,
				clinicA, typeA, assignA, staffA.ID, lineTimePolicyWeeklyTimesJSON).Error,
			"23505")

		// Reconfigure clears the tombstone on the same row; entity_version_guard
		// advances the version once for the changed row (3 -> 4).
		require.NoError(t, db.Exec(`UPDATE line_staff_start_policies SET deleted_at = NULL, date_overrides = '[]' WHERE clinic_id = ?`, clinicA).Error)
		var tombstoned bool
		require.NoError(t, db.Raw(`SELECT deleted_at IS NOT NULL FROM line_staff_start_policies WHERE clinic_id = ?`, clinicA).Scan(&tombstoned).Error)
		require.False(t, tombstoned, "reconfigure clears deleted_at")
		require.NoError(t, db.Raw(`SELECT version FROM line_staff_start_policies WHERE clinic_id = ?`, clinicA).Scan(&version).Error)
		require.EqualValues(t, 4, version, "reconfigure advances the version once")

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

		// Real type/staff/assignment ids for the line_start_policy anchor. The
		// anchor carries no business FK by design (contract §4.3), so these only
		// make the fixture realistic.
		var anchorType, anchorAssign int64
		anchorStaff := testdb.MakeDoctor(t, db, uint64(clinic), "010 anchor staff")
		require.NoError(t, db.Raw(`INSERT INTO reservation_types (clinic_id, name, duration_minutes) VALUES (?, '010 anchor type', 15) RETURNING id`, clinic).Scan(&anchorType).Error)
		require.NoError(t, db.Raw(`INSERT INTO staff_clinic_assignments (staff_id, clinic_id, is_main) VALUES (?, ?, true) RETURNING id`, anchorStaff.ID, clinic).Scan(&anchorAssign).Error)

		// insertAnchor inserts one anchor row on a fresh binding of this clinic;
		// nil arguments are SQL NULL.
		insertAnchor := func(kind string, targetID, staffID, assignmentID any) error {
			return db.Exec(`
				INSERT INTO reservation_master_operation_anchors
					(operation_binding_id, clinic_id, target_kind, target_id, staff_id, assignment_id)
				VALUES (?, ?, ?, ?, ?, ?)`,
				lineTimeSeedBinding(t, db, clinic, "reservation-master.anchor-fixture"), clinic, kind, targetID, staffID, assignmentID).Error
		}

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

		// Valid rows of the second CHECK branch and of the NULL-target type/group
		// branch. CHECK: (type|group AND staff/assignment NULL) OR
		// (line_start_policy AND target/staff/assignment NOT NULL); target_id
		// itself is nullable for type/group (create-family).
		require.NoError(t, insertAnchor("line_start_policy", anchorType, anchorStaff.ID, anchorAssign))
		require.NoError(t, insertAnchor("line_start_policy", int64(999997), int64(999998), int64(999999)),
			"no business FK: the original relation is validated by the app, not the claim row")
		require.NoError(t, insertAnchor("group", nil, nil, nil))
		require.NoError(t, insertAnchor("type", nil, nil, nil))

		// line_start_policy requires each of target/staff/assignment.
		lineTimeRequirePgCode(t, insertAnchor("line_start_policy", int64(1), nil, int64(3)), "23514")
		lineTimeRequirePgCode(t, insertAnchor("line_start_policy", int64(1), int64(2), nil), "23514")
		lineTimeRequirePgCode(t, insertAnchor("line_start_policy", nil, int64(2), int64(3)), "23514")

		// type/group reject staff_id or assignment_id (either one alone).
		lineTimeRequirePgCode(t, insertAnchor("type", nil, int64(2), nil), "23514")
		lineTimeRequirePgCode(t, insertAnchor("type", int64(1), nil, int64(3)), "23514")
		lineTimeRequirePgCode(t, insertAnchor("group", nil, nil, int64(3)), "23514")

		// target_kind is the closed set type | group | line_start_policy.
		lineTimeRequirePgCode(t, insertAnchor("unknown", nil, nil, nil), "23514")

		// Non-positive ids: column CHECKs (target_id/staff_id/assignment_id are
		// `IS NULL OR > 0`) and operation_binding_id/clinic_id `> 0`. CHECKs fire
		// before the binding FK is looked up.
		lineTimeRequirePgCode(t, insertAnchor("type", int64(0), nil, nil), "23514")
		lineTimeRequirePgCode(t, insertAnchor("type", int64(-1), nil, nil), "23514")
		lineTimeRequirePgCode(t, insertAnchor("line_start_policy", int64(1), int64(0), int64(3)), "23514")
		lineTimeRequirePgCode(t, insertAnchor("line_start_policy", int64(1), int64(2), int64(-1)), "23514")
		lineTimeRequirePgCode(t,
			db.Exec(`
				INSERT INTO reservation_master_operation_anchors (operation_binding_id, clinic_id, target_kind)
				VALUES (?, 0, 'type')`, lineTimeSeedBinding(t, db, clinic, "reservation-type.create")).Error,
			"23514")
		lineTimeRequirePgCode(t,
			db.Exec(`
				INSERT INTO reservation_master_operation_anchors (operation_binding_id, clinic_id, target_kind)
				VALUES (0, ?, 'type')`, clinic).Error,
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
