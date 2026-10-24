package medicalrecord

// medical_record_delete_in_consultation_race_test.go — F-1
//
// Reservation in_consultation は appointments を FOR UPDATE してから medical_records を
// COUNT する。EMR-495 P1: medicalRecordService.Delete は固定エラーを返す（削除経路なし）ため、
// Delete 経由の競合テスト（旧 DeleteWaitsOnAppointmentRowLockBeforeInConsultationCommit）は廃止した。
// リポジトリの LockLinkedAppointmentForUpdate の契約テストのみ残す。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func setupMedicalRecordDeleteAppointmentRaceDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupMedicalRecordListTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.ReservationType{}, &model.Reservation{}))
	require.NoError(t, db.Exec("TRUNCATE TABLE medical_records, appointments, reservation_types CASCADE").Error)
	return db
}

func seedDraftRecordLinkedToConfirmedAppointment(
	t *testing.T,
	db *gorm.DB,
	clinicID uint64,
	recordNo string,
) (*model.Reservation, *model.MedicalRecord) {
	t.Helper()
	reservationType := &model.ReservationType{
		ClinicID: clinicID,
		Name:     "in-consultation race",
		Category: model.ReservationTypeCategoryGeneral,
		IsActive: true,
	}
	require.NoError(t, db.Create(reservationType).Error)
	start := time.Date(2026, 8, 22, 10, 0, 0, 0, time.UTC)
	appointment := &model.Reservation{
		ClinicID:          clinicID,
		StartTime:         start,
		EndTime:           start.Add(30 * time.Minute),
		VisitType:         model.VisitTypeRevisit,
		ReservationTypeID: reservationType.ID,
		Status:            model.ReservationStatusConfirmed,
		Source:            model.ReservationSourceManual,
		CustomerFields:    []byte(`{}`),
	}
	require.NoError(t, db.Create(appointment).Error)
	record := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID:      clinicID,
		RecordNo:      recordNo,
		Date:          start,
		AppointmentID: &appointment.ID,
	})
	return appointment, record
}

func lockAppointmentForUpdate(t *testing.T, tx *gorm.DB, clinicID, appointmentID uint64) {
	t.Helper()
	var locked model.Reservation
	require.NoError(t, tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND clinic_id = ?", appointmentID, clinicID).
		First(&locked).Error)
}

func TestMedicalRecordRepository_LockLinkedAppointmentForUpdate_RequiresAmbientTransaction(t *testing.T) {
	db := setupMedicalRecordDeleteAppointmentRaceDB(t)
	repo := NewMedicalRecordRepository(db)
	_, err := repo.LockLinkedAppointmentForUpdate(context.Background(), 1, 1)
	require.Error(t, err)
	assert.False(t, apperrors.IsNotFound(err))
	assert.False(t, apperrors.IsConflict(err))
}

func TestMedicalRecordRepository_LockLinkedAppointmentForUpdate_WaitsOnAppointmentRowLock(t *testing.T) {
	const clinicID = uint64(1)
	db := setupMedicalRecordDeleteAppointmentRaceDB(t)
	appointment, record := seedDraftRecordLinkedToConfirmedAppointment(t, db, clinicID, "DEL-APPT-LOCK-ORDER")
	repo := NewMedicalRecordRepository(db)

	holdingTx := db.Begin()
	require.NoError(t, holdingTx.Error)
	released := false
	defer func() {
		if !released {
			holdingTx.Rollback()
		}
	}()
	lockAppointmentForUpdate(t, holdingTx, clinicID, appointment.ID)

	lockDone := make(chan error, 1)
	go func() {
		lockDone <- (testTransactor{db: db}).WithTx(context.Background(), func(txCtx context.Context) error {
			_, err := repo.LockLinkedAppointmentForUpdate(txCtx, clinicID, record.ID)
			return err
		})
	}()
	select {
	case err := <-lockDone:
		require.Failf(t, "appointment lock did not wait", "err=%v", err)
	case <-time.After(100 * time.Millisecond):
	}
	require.NoError(t, holdingTx.Rollback().Error)
	released = true

	select {
	case err := <-lockDone:
		require.NoError(t, err)
	case <-time.After(30 * time.Second):
		t.Fatal("appointment lock did not finish after the holding transaction released")
	}
}
