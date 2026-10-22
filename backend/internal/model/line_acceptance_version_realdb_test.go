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
}
