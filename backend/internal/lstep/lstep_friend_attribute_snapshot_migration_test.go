package lstep

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

const lstepSnapshotImportClinicFKMigrationSQL = `-- Enforce tenant ownership across LSTEP friend snapshots and their CSV import.
-- Existing mismatches abort the migration before either constraint is changed.

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM lstep_friend_attribute_snapshots AS snapshot
        JOIN lstep_csv_imports AS csv_import
          ON csv_import.id = snapshot.csv_import_id
        WHERE snapshot.csv_import_id IS NOT NULL
          AND snapshot.clinic_id <> csv_import.clinic_id
    ) THEN
        RAISE EXCEPTION
            'cross-clinic lstep friend snapshot csv_import reference exists'
            USING ERRCODE = '23514';
    END IF;
END
$$;

ALTER TABLE lstep_csv_imports
    ADD CONSTRAINT uq_lstep_csv_imports_clinic_id_id
    UNIQUE (clinic_id, id);

ALTER TABLE lstep_friend_attribute_snapshots
    DROP CONSTRAINT IF EXISTS lstep_friend_attribute_snapshots_csv_import_id_fkey;

ALTER TABLE lstep_friend_attribute_snapshots
    ADD CONSTRAINT fk_lstep_snapshots_clinic_csv_import
    FOREIGN KEY (clinic_id, csv_import_id)
    REFERENCES lstep_csv_imports (clinic_id, id)
    ON DELETE RESTRICT;
`

func TestLstepSnapshotImportClinicFKMigration_MatchesArchivedInitialMigration(t *testing.T) {
	raw, err := os.ReadFile("../../migrations/001_init.sql")
	require.NoError(t, err)
	initial := string(raw)

	const sourceMarker = "-- Source file: 002_lstep_snapshot_import_clinic_fk.sql"
	const nextSourceMarker = "-- Source file: 003_medical_records_appointment_id_index.sql"
	start := strings.Index(initial, sourceMarker)
	require.GreaterOrEqual(t, start, 0, "001_init.sql must contain the archived LSTEP migration")

	endOffset := strings.Index(initial[start:], "\n"+nextSourceMarker)
	require.Greater(t, endOffset, 0, "archived LSTEP migration must end at the 003 source marker")
	block := initial[start : start+endOffset]

	const sourceSHA = "10222c570054a80a5d47cf4b66e4235e92ca35643c9c23ef00a9ce8bca0086b6"
	const shaHeader = "-- Source SHA-256: " + sourceSHA + "\n"
	shaOffset := strings.Index(block, shaHeader)
	require.GreaterOrEqual(t, shaOffset, 0, "archived LSTEP migration must contain its exact SHA-256 header")

	body := block[shaOffset+len(shaHeader):]
	require.Equal(t, lstepSnapshotImportClinicFKMigrationSQL, body)

	checksum := sha256.Sum256([]byte(body))
	require.Equal(t, sourceSHA, fmt.Sprintf("%x", checksum))
}

func TestLstepSnapshotImportClinicFKMigration(t *testing.T) {
	db := setupLstepCsvImportServiceTestDB(t)
	applyLstepSnapshotImportClinicFKMigration(t, db)

	clinicA := seedLstepCsvImportClinic(t, db)
	clinicB := seedLstepCsvImportClinic(t, db)
	importA := &model.LstepCsvImport{ClinicID: clinicA, CsvType: csvTypeFriendAttribute, FileName: "a.csv"}
	importB := &model.LstepCsvImport{ClinicID: clinicB, CsvType: csvTypeFriendAttribute, FileName: "b.csv"}
	require.NoError(t, db.Create(importA).Error)
	require.NoError(t, db.Create(importB).Error)

	takenAt := time.Date(2026, 7, 22, 0, 0, 0, 0, time.UTC)
	crossClinic := &model.LstepFriendAttributeSnapshot{
		ClinicID: clinicA, LineUserID: "U-cross-clinic", SnapshotTakenAt: takenAt, CsvImportID: &importB.ID,
	}
	err := db.Create(crossClinic).Error
	require.Error(t, err)
	assert.True(t, isPostgresFKViolation(err), "cross-clinic import reference must fail with 23503: %v", err)

	valid := &model.LstepFriendAttributeSnapshot{
		ClinicID: clinicA, LineUserID: "U-same-clinic", SnapshotTakenAt: takenAt, CsvImportID: &importA.ID,
	}
	require.NoError(t, db.Create(valid).Error)

	withoutImport := &model.LstepFriendAttributeSnapshot{
		ClinicID: clinicA, LineUserID: "U-no-import", SnapshotTakenAt: takenAt,
	}
	require.NoError(t, db.Create(withoutImport).Error)
}

func TestLstepSnapshotImportClinicFKMigration_RejectsExistingCrossClinicRows(t *testing.T) {
	db := setupLstepCsvImportServiceTestDB(t)
	clinicA := seedLstepCsvImportClinic(t, db)
	clinicB := seedLstepCsvImportClinic(t, db)
	importB := &model.LstepCsvImport{ClinicID: clinicB, CsvType: csvTypeFriendAttribute, FileName: "b.csv"}
	require.NoError(t, db.Create(importB).Error)

	crossClinic := &model.LstepFriendAttributeSnapshot{
		ClinicID: clinicA, LineUserID: "U-existing-cross-clinic",
		SnapshotTakenAt: time.Date(2026, 7, 22, 0, 0, 0, 0, time.UTC), CsvImportID: &importB.ID,
	}
	require.NoError(t, db.Create(crossClinic).Error)

	err := db.Exec(lstepSnapshotImportClinicFKMigrationSQL).Error
	require.Error(t, err)
	assert.True(t, isPostgresCode(err, "23514"), "precheck must reject existing cross-clinic rows: %v", err)
}

func applyLstepSnapshotImportClinicFKMigration(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(lstepSnapshotImportClinicFKMigrationSQL).Error)
}

func isPostgresFKViolation(err error) bool {
	return isPostgresCode(err, "23503")
}

func isPostgresCode(err error, code string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == code
}
