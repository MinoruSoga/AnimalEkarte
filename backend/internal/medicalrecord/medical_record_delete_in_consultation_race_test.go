package medicalrecord

// medical_record_delete_in_consultation_race_test.go — F-1
//
// Reservation in_consultation は appointments を FOR UPDATE してから medical_records を
// COUNT する。Delete も appointments 行ロックを先に取り、同じロック順序で相互排除する。
// 予約側が先に commit して in_consultation になった場合、Delete は Conflict し live karte を残す。

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
		Status:        model.MedicalRecordStatusDraft,
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

func countLiveRecordsForAppointment(t *testing.T, db *gorm.DB, appointmentID uint64) int64 {
	t.Helper()
	var count int64
	require.NoError(t, db.Model(&model.MedicalRecord{}).
		Where("appointment_id = ? AND deleted_at IS NULL", appointmentID).
		Count(&count).Error)
	return count
}

func TestMedicalRecordService_DeleteWaitsOnAppointmentRowLockBeforeInConsultationCommit(t *testing.T) {
	const clinicID = uint64(1)
	db := setupMedicalRecordDeleteAppointmentRaceDB(t)
	appointment, record := seedDraftRecordLinkedToConfirmedAppointment(t, db, clinicID, "DEL-INCONSULT-RACE")
	repo := NewMedicalRecordRepository(db)
	svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, testTransactor{db: db})

	holdingTx := db.Begin()
	require.NoError(t, holdingTx.Error)
	committed := false
	defer func() {
		if !committed {
			holdingTx.Rollback()
		}
	}()
	lockAppointmentForUpdate(t, holdingTx, clinicID, appointment.ID)

	deleteDone := make(chan error, 1)
	go func() { deleteDone <- svc.Delete(context.Background(), clinicID, record.ID) }()
	select {
	case err := <-deleteDone:
		require.Failf(t, "delete did not wait for appointment row lock", "err=%v", err)
	case <-time.After(100 * time.Millisecond):
	}

	require.NoError(t, holdingTx.Model(&model.Reservation{}).
		Where("id = ? AND clinic_id = ?", appointment.ID, clinicID).
		Update("status", model.ReservationStatusInConsultation).Error)
	require.NoError(t, holdingTx.Commit().Error)
	committed = true

	var deleteErr error
	select {
	case deleteErr = <-deleteDone:
	case <-time.After(2 * time.Second):
		t.Fatal("delete did not finish after appointment lock was released")
	}
	require.Error(t, deleteErr)
	assert.True(t, apperrors.IsConflict(deleteErr), "deleteErr=%v", deleteErr)

	var persisted model.Reservation
	require.NoError(t, db.First(&persisted, appointment.ID).Error)
	liveCount := countLiveRecordsForAppointment(t, db, appointment.ID)
	assert.Equal(t, model.ReservationStatusInConsultation, persisted.Status)
	assert.EqualValues(t, 1, liveCount, "in_consultation appointment must keep its live karte")
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
	case <-time.After(2 * time.Second):
		t.Fatal("appointment lock did not finish after the holding transaction released")
	}
}
