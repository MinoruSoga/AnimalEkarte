package lstep

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

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

	migration, err := os.ReadFile("../../migrations/002_lstep_snapshot_import_clinic_fk.sql")
	require.NoError(t, err)
	err = db.Exec(string(migration)).Error
	require.Error(t, err)
	assert.True(t, isPostgresCode(err, "23514"), "precheck must reject existing cross-clinic rows: %v", err)
}

func applyLstepSnapshotImportClinicFKMigration(t *testing.T, db *gorm.DB) {
	t.Helper()
	migration, err := os.ReadFile("../../migrations/002_lstep_snapshot_import_clinic_fk.sql")
	require.NoError(t, err)
	require.NoError(t, db.Exec(string(migration)).Error)
}

func isPostgresFKViolation(err error) bool {
	return isPostgresCode(err, "23503")
}

func isPostgresCode(err error, code string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == code
}
