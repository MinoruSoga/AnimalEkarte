package model_test

// line_acceptance_version_realdb_test.go — real-DB fixture for
// backend/migrations/009_line_acceptance_version.sql (EMR-446/466 A03 LINE
// acceptance, contract §3.1). Applies 001 via testdb.SetupIsolated001TestDB
// and Execs the 009 body in-test (no dependency on 008). NOT_RUN yet.
// Shared helpers live in clinic_operating_hours_realdb_test.go (stack base).

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/testdb"
)

func TestLineAcceptanceVersionRealDB(t *testing.T) {
	db := testdb.SetupIsolated001TestDB(t)
	require.NoError(t, db.Exec(lineTimeReadMigration(t, "009_line_acceptance_version.sql")).Error,
		"009 must apply cleanly on top of 001")

	clinic := lineTimeSeedClinic(t, db, "009-line")

	// INSERT: writer-supplied versions are not adopted; acceptance starts at 1.
	require.NoError(t, db.Exec(`
		INSERT INTO line_reservation_settings (clinic_id, acceptance_version, reviewed_acceptance_version)
		VALUES (?, 99, NULL)`, clinic).Error)
	var acceptance int64
	var reviewed *int64
	require.NoError(t, db.Raw(`
		SELECT acceptance_version, reviewed_acceptance_version FROM line_reservation_settings WHERE clinic_id = ?`, clinic).
		Row().Scan(&acceptance, &reviewed))
	require.EqualValues(t, 1, acceptance, "INSERT pins acceptance_version to 1")
	require.Nil(t, reviewed, "existing rows stay legacy-unreviewed")

	// reviewed beyond the acceptance version is rejected by the CHECK.
	lineTimeRequirePgCode(t,
		db.Exec(`INSERT INTO line_reservation_settings (clinic_id, reviewed_acceptance_version) VALUES (?, 5)`,
			lineTimeSeedClinic(t, db, "009-line-bad-reviewed")).Error,
		"23514")

	// initialize -> configured (contract §3.2): an INSERT carrying
	// reviewed_acceptance_version = 1 succeeds. The guard pins the resulting
	// acceptance_version to 1 (the writer-supplied 99 is not adopted), so
	// reviewed (1) <= acceptance (1) satisfies the named CHECK and the row is
	// configured from the start.
	clinicInit := lineTimeSeedClinic(t, db, "009-line-init-configured")
	require.NoError(t, db.Exec(`
		INSERT INTO line_reservation_settings (clinic_id, acceptance_version, reviewed_acceptance_version)
		VALUES (?, 99, 1)`, clinicInit).Error)
	require.NoError(t, db.Raw(`
		SELECT acceptance_version, reviewed_acceptance_version FROM line_reservation_settings WHERE clinic_id = ?`, clinicInit).
		Row().Scan(&acceptance, &reviewed))
	require.EqualValues(t, 1, acceptance, "INSERT pins acceptance_version to 1 even with reviewed = 1")
	require.NotNil(t, reviewed)
	require.EqualValues(t, 1, *reviewed, "reviewed = acceptance_version = configured")

	// The guard forces acceptance_version to 1 before the CHECK runs, so a
	// reviewed marker above 1 is rejected even when the writer supplied a
	// matching larger version (99 / 99 -> 1 / 99).
	lineTimeRequirePgCode(t,
		db.Exec(`
			INSERT INTO line_reservation_settings (clinic_id, acceptance_version, reviewed_acceptance_version)
			VALUES (?, 99, 99)`, lineTimeSeedClinic(t, db, "009-line-init-reviewed-ahead")).Error,
		"23514")

	// reviewed must stay positive (CHECK reviewed_acceptance_version > 0).
	lineTimeRequirePgCode(t,
		db.Exec(`INSERT INTO line_reservation_settings (clinic_id, reviewed_acceptance_version) VALUES (?, 0)`,
			lineTimeSeedClinic(t, db, "009-line-init-reviewed-zero")).Error,
		"23514")

	// P change advances the version and leaves reviewed behind (unreviewed).
	require.NoError(t, db.Exec(`UPDATE line_reservation_settings SET status = 'active' WHERE clinic_id = ?`, clinic).Error)
	require.NoError(t, db.Raw(`SELECT acceptance_version, reviewed_acceptance_version FROM line_reservation_settings WHERE clinic_id = ?`, clinic).
		Row().Scan(&acceptance, &reviewed))
	require.EqualValues(t, 2, acceptance)
	require.Nil(t, reviewed, "P change without confirm falls back to legacy-unreviewed")

	// N-only change keeps the version.
	require.NoError(t, db.Exec(`UPDATE line_reservation_settings SET header_text = 'x' WHERE clinic_id = ?`, clinic).Error)
	require.NoError(t, db.Raw(`SELECT acceptance_version FROM line_reservation_settings WHERE clinic_id = ?`, clinic).Scan(&acceptance).Error)
	require.EqualValues(t, 2, acceptance, "non-P display fields must not advance the version")

	// daily_limit is outside P (contract §3.1: retired limits are legacy-held):
	// version and reviewed stay unchanged (2 / NULL).
	require.NoError(t, db.Exec(`UPDATE line_reservation_settings SET daily_limit = 5 WHERE clinic_id = ?`, clinic).Error)
	require.NoError(t, db.Raw(`SELECT acceptance_version, reviewed_acceptance_version FROM line_reservation_settings WHERE clinic_id = ?`, clinic).
		Row().Scan(&acceptance, &reviewed))
	require.EqualValues(t, 2, acceptance, "daily_limit is not in P and must not advance the version")
	require.Nil(t, reviewed, "non-P change leaves reviewed untouched")

	// Setting reviewed to NULL restores OLD (confirmation cannot be cleared).
	require.NoError(t, db.Exec(`UPDATE line_reservation_settings SET reviewed_acceptance_version = NULL WHERE clinic_id = ?`, clinic).Error)
	require.NoError(t, db.Raw(`SELECT acceptance_version, reviewed_acceptance_version FROM line_reservation_settings WHERE clinic_id = ?`, clinic).
		Row().Scan(&acceptance, &reviewed))
	require.EqualValues(t, 2, acceptance)
	require.Nil(t, reviewed)

	// A mismatched reviewed marker is rejected.
	lineTimeRequirePgCode(t,
		db.Exec(`UPDATE line_reservation_settings SET reviewed_acceptance_version = 7 WHERE clinic_id = ?`, clinic).Error,
		"23514")

	// Confirm: reviewed := expected+1 advances the version and marks configured.
	require.NoError(t, db.Exec(`UPDATE line_reservation_settings SET reviewed_acceptance_version = 3 WHERE clinic_id = ?`, clinic).Error)
	require.NoError(t, db.Raw(`SELECT acceptance_version, reviewed_acceptance_version FROM line_reservation_settings WHERE clinic_id = ?`, clinic).
		Row().Scan(&acceptance, &reviewed))
	require.EqualValues(t, 3, acceptance)
	require.NotNil(t, reviewed)
	require.EqualValues(t, 3, *reviewed, "reviewed equals the resulting version = configured")

	// Writer-SET version values are always overwritten (no double increment).
	require.NoError(t, db.Exec(`UPDATE line_reservation_settings SET acceptance_version = 42 WHERE clinic_id = ?`, clinic).Error)
	require.NoError(t, db.Raw(`SELECT acceptance_version FROM line_reservation_settings WHERE clinic_id = ?`, clinic).Scan(&acceptance).Error)
	require.EqualValues(t, 3, acceptance, "writer-supplied version is ignored")

	// Confirmed (acceptance = reviewed): a true no-op keeps both unchanged.
	require.NoError(t, db.Exec(`UPDATE line_reservation_settings SET status = status WHERE clinic_id = ?`, clinic).Error)
	require.NoError(t, db.Raw(`SELECT acceptance_version, reviewed_acceptance_version FROM line_reservation_settings WHERE clinic_id = ?`, clinic).
		Row().Scan(&acceptance, &reviewed))
	require.EqualValues(t, 3, acceptance, "confirmed no-op must not advance the version")
	require.NotNil(t, reviewed)
	require.EqualValues(t, 3, *reviewed)

	// Confirmed + non-P change (additional_fields): the guard sees no P or
	// reviewed change, so acceptance (3) and reviewed (3) both stay put and the
	// row remains configured.
	require.NoError(t, db.Exec(`UPDATE line_reservation_settings SET additional_fields = '[]' WHERE clinic_id = ?`, clinic).Error)
	require.NoError(t, db.Raw(`SELECT acceptance_version, reviewed_acceptance_version FROM line_reservation_settings WHERE clinic_id = ?`, clinic).
		Row().Scan(&acceptance, &reviewed))
	require.EqualValues(t, 3, acceptance, "additional_fields is not in P and must not advance the version")
	require.NotNil(t, reviewed)
	require.EqualValues(t, 3, *reviewed, "non-P change keeps the confirmed marker")

	// Confirmed + P change without reviewed: the version advances once and
	// reviewed stays behind (back to legacy-unreviewed).
	require.NoError(t, db.Exec(`UPDATE line_reservation_settings SET national_holiday_closed = true WHERE clinic_id = ?`, clinic).Error)
	require.NoError(t, db.Raw(`SELECT acceptance_version, reviewed_acceptance_version FROM line_reservation_settings WHERE clinic_id = ?`, clinic).
		Row().Scan(&acceptance, &reviewed))
	require.EqualValues(t, 4, acceptance)
	require.NotNil(t, reviewed)
	require.EqualValues(t, 3, *reviewed)

	// Re-confirm to configured (reviewed := expected+1).
	require.NoError(t, db.Exec(`UPDATE line_reservation_settings SET reviewed_acceptance_version = 5 WHERE clinic_id = ?`, clinic).Error)
	require.NoError(t, db.Raw(`SELECT acceptance_version, reviewed_acceptance_version FROM line_reservation_settings WHERE clinic_id = ?`, clinic).
		Row().Scan(&acceptance, &reviewed))
	require.EqualValues(t, 5, acceptance)
	require.NotNil(t, reviewed)
	require.EqualValues(t, 5, *reviewed)

	// Confirmed + reviewed := NULL: the stored marker is restored (confirmation
	// cannot be cleared) and the version does not move.
	require.NoError(t, db.Exec(`UPDATE line_reservation_settings SET reviewed_acceptance_version = NULL WHERE clinic_id = ?`, clinic).Error)
	require.NoError(t, db.Raw(`SELECT acceptance_version, reviewed_acceptance_version FROM line_reservation_settings WHERE clinic_id = ?`, clinic).
		Row().Scan(&acceptance, &reviewed))
	require.EqualValues(t, 5, acceptance)
	require.NotNil(t, reviewed)
	require.EqualValues(t, 5, *reviewed, "reviewed := NULL restores the stored marker")

	// P change + reviewed := expected+1 in a single UPDATE: one increment with
	// reviewed equal to the resulting version (the confirm/update write path).
	require.NoError(t, db.Exec(`UPDATE line_reservation_settings SET break_hours = '[]', reviewed_acceptance_version = 6 WHERE clinic_id = ?`, clinic).Error)
	require.NoError(t, db.Raw(`SELECT acceptance_version, reviewed_acceptance_version FROM line_reservation_settings WHERE clinic_id = ?`, clinic).
		Row().Scan(&acceptance, &reviewed))
	require.EqualValues(t, 6, acceptance, "P change + confirm in one update increments exactly once")
	require.NotNil(t, reviewed)
	require.EqualValues(t, 6, *reviewed)

	// Each protected field P individually advances the version once. A fresh
	// row starts at acceptance 1 / legacy-unreviewed; values below differ from
	// the 001_init.sql column defaults while satisfying its types (no extra
	// CHECKs exist on these columns) and the app-level enum/JSON shapes.
	clinicP := lineTimeSeedClinic(t, db, "009-line-p-fields")
	require.NoError(t, db.Exec(`INSERT INTO line_reservation_settings (clinic_id) VALUES (?)`, clinicP).Error)
	require.NoError(t, db.Raw(`SELECT acceptance_version, reviewed_acceptance_version FROM line_reservation_settings WHERE clinic_id = ?`, clinicP).
		Row().Scan(&acceptance, &reviewed))
	require.EqualValues(t, 1, acceptance)
	require.Nil(t, reviewed)

	pChanges := []struct {
		name string
		set  string
	}{
		{"status", `status = 'running'`},
		{"closed_weekdays", `closed_weekdays = '[0]'`},
		{"closed_dates", `closed_dates = '["2026-12-25"]'`},
		{"national_holiday_closed", `national_holiday_closed = true`},
		{"business_hours", `business_hours = '{"start":"1000","end":"1800"}'`},
		{"business_hours_by_weekday", `business_hours_by_weekday = '{"1":{"start":"0900","end":"1200"}}'`},
		{"break_hours", `break_hours = '[{"start":"1300","end":"1400"}]'`},
		{"booking_window_max_days", `booking_window_max_days = 60`},
		{"booking_window_min_days", `booking_window_min_days = 1`},
		{"time_slot_mode", `time_slot_mode = 'allow_gaps'`},
		{"time_slot_interval_minutes", `time_slot_interval_minutes = 30`},
		{"show_no_staff_option", `show_no_staff_option = false`},
	}
	for i, pc := range pChanges {
		require.NoError(t, db.Exec(`UPDATE line_reservation_settings SET `+pc.set+` WHERE clinic_id = ?`, clinicP).Error,
			"P field %s must update cleanly", pc.name)
		require.NoError(t, db.Raw(`SELECT acceptance_version, reviewed_acceptance_version FROM line_reservation_settings WHERE clinic_id = ?`, clinicP).
			Row().Scan(&acceptance, &reviewed))
		require.EqualValues(t, i+2, acceptance, "P field %s must advance the version once", pc.name)
		require.Nil(t, reviewed, "unconfirmed P field %s change stays legacy-unreviewed", pc.name)
	}
}
