package reservation

// in_consultation_karte_delete_race_test.go — reverse of medicalrecord F-1.
//
// Forward race (medicalrecord): reservation commits in_consultation first, then
// medical-record Delete returns 409 and keeps the live karte.
// Reverse race (this file): GORM soft-delete of the linked karte commits first
// (lane cannot call medicalrecord.Service.Delete). ReservationService.Update to
// in_consultation then runs applyReservationUpdate's LockAndFindByID +
// validateInConsultationHasMedicalRecord COUNT and must Conflict because
// CountMedicalRecordsByReservationID sees deleted_at IS NULL → 0.

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func setupInConsultationKarteDeleteRaceDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.ReservationType{}, &model.Reservation{}, &model.MedicalRecord{}))
	return db
}

// seedConfirmedAppointmentWithLiveKarte creates a clinic-scoped confirmed appointment
// and a live (non-deleted) medical record linked via appointment_id.
// Reservation type / RecordNo must be unique: reservation_types (clinic_id, name) and
// medical_records (clinic_id, record_no) are unique, and this suite does not TRUNCATE
// appointments/reservation_types (shared testdb).
func seedConfirmedAppointmentWithLiveKarte(
	t *testing.T,
	db *gorm.DB,
	clinicID uint64,
	recordNo string,
) (*model.Reservation, *model.MedicalRecord) {
	t.Helper()
	reservationType := &model.ReservationType{
		ClinicID: clinicID,
		Name:     "karte-del-race-" + recordNo,
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

	record := &model.MedicalRecord{
		ClinicID:      clinicID,
		RecordNo:      recordNo,
		Date:          start,
		Status:        model.MedicalRecordStatusDraft,
		AppointmentID: &appointment.ID,
	}
	require.NoError(t, db.Create(record).Error)
	return appointment, record
}

func TestReservationService_Update_InConsultationAfterCommittedKarteDelete(t *testing.T) {
	tests := []struct {
		name          string
		deleteFirst   bool
		wantErr       bool
		wantStatus    model.ReservationStatus
		wantLiveCount int64
	}{
		{
			name:          "カルテ削除が先にコミットされると in_consultation は Conflict",
			deleteFirst:   true,
			wantErr:       true,
			wantStatus:    model.ReservationStatusConfirmed,
			wantLiveCount: 0,
		},
		{
			name:          "有効カルテが残っていれば in_consultation へ遷移できる",
			deleteFirst:   false,
			wantErr:       false,
			wantStatus:    model.ReservationStatusInConsultation,
			wantLiveCount: 1,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupInConsultationKarteDeleteRaceDB(t)
			ctx := context.Background()
			const clinicID = uint64(1)
			recordNo := fmt.Sprintf("MR-RACE-REV-%d-%d", i, time.Now().UnixNano())
			appointment, record := seedConfirmedAppointmentWithLiveKarte(t, db, clinicID, recordNo)

			repo := NewReservationRepository(db)
			svc := NewReservationServiceWithAvailabilityAndType(
				repo,
				nil,
				testNewTransactor(db),
				nil,
				nil,
			)

			liveBefore, err := repo.CountMedicalRecordsByReservationID(ctx, clinicID, appointment.ID)
			require.NoError(t, err)
			require.Equal(t, int64(1), liveBefore, "seed must create one live karte")

			// Transaction 1: committed GORM soft-delete (not medicalrecord.Service.Delete).
			if tt.deleteFirst {
				require.NoError(t, db.WithContext(ctx).Delete(record).Error)
				var deleted model.MedicalRecord
				require.NoError(t, db.WithContext(ctx).Unscoped().First(&deleted, record.ID).Error)
				require.True(t, deleted.DeletedAt.Valid, "transaction 1 must commit a soft-delete")
			}

			// Transaction 2: real ReservationService.Update WithTx
			// (LockAndFindByID then validateInConsultationHasMedicalRecord COUNT).
			status := model.ReservationStatusInConsultation
			updated, err := svc.Update(ctx, clinicID, appointment.ID, &UpdateReservationInput{Status: &status})

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, updated)
				assert.True(t, apperrors.IsConflict(err), "expected conflict but got: %v", err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, updated)
				assert.Equal(t, model.ReservationStatusInConsultation, updated.Status)
			}

			var persisted model.Reservation
			require.NoError(t, db.WithContext(ctx).First(&persisted, appointment.ID).Error)
			assert.Equal(t, tt.wantStatus, persisted.Status)
			if tt.deleteFirst {
				assert.NotEqual(t, model.ReservationStatusInConsultation, persisted.Status)
			}

			count, err := repo.CountMedicalRecordsByReservationID(ctx, clinicID, appointment.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.wantLiveCount, count)
		})
	}
}
