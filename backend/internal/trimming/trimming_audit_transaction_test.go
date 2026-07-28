package trimming

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/reservation"
)

// TestTrimmingService_DeleteAuditFailureRollsBackDatabase is intentionally DB-backed.
// Run only while holding the repository-wide global DB test lease.
func TestTrimmingService_DeleteAuditFailureRollsBackDatabase(t *testing.T) {
	db := setupIsolatedTestDB(t)
	require.NoError(t, db.Exec("TRUNCATE TABLE appointment_trimming_options CASCADE").Error)
	require.NoError(t, db.Exec("TRUNCATE TABLE appointment_trimming_details CASCADE").Error)
	require.NoError(t, db.Exec("TRUNCATE TABLE reservation_types CASCADE").Error)

	const clinicID, actorID = uint64(1), uint64(42)
	reservationType := &model.ReservationType{
		ClinicID: clinicID,
		Name:     "監査rollback用トリミング",
		Category: model.ReservationTypeCategoryTrimming,
		IsActive: true,
	}
	require.NoError(t, db.Create(reservationType).Error)
	start := time.Now().UTC().Add(time.Hour).Truncate(time.Minute)
	appointment := &model.Reservation{
		ClinicID:          clinicID,
		StartTime:         start,
		EndTime:           start.Add(time.Hour),
		ReservationTypeID: reservationType.ID,
		VisitType:         model.VisitTypeRevisit,
		Status:            model.ReservationStatusPending,
		Source:            model.ReservationSourceManual,
		CustomerFields:    json.RawMessage(`{}`),
	}
	require.NoError(t, db.Create(appointment).Error)
	detail := &model.AppointmentTrimmingDetail{
		ClinicID:      clinicID,
		AppointmentID: appointment.ID,
		StyleRequest:  "監査失敗でも残る",
		BWUnit:        model.BodyWeightUnitKg,
	}
	require.NoError(t, db.Create(detail).Error)

	sentinel := errors.New("forced durable audit failure")
	svc := NewTrimmingServiceWithAudit(
		reservation.NewReservationRepository(db),
		nil,
		nil,
		nil,
		NewAppointmentTrimmingDetailRepository(db),
		nil,
		nil,
		newTestTransactor(db),
		&trimmingAuditRecorder{err: sentinel},
	)

	err := svc.Delete(context.Background(), clinicID, appointment.ID, ptrUint64(actorID))

	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	var persistedAppointment model.Reservation
	require.NoError(t, db.Unscoped().First(&persistedAppointment, appointment.ID).Error)
	assert.False(t, persistedAppointment.DeletedAt.Valid, "audit failure must roll back the appointment soft delete")
	var detailCount int64
	require.NoError(t, db.Model(&model.AppointmentTrimmingDetail{}).
		Where("clinic_id = ? AND appointment_id = ?", clinicID, appointment.ID).
		Count(&detailCount).Error)
	assert.Equal(t, int64(1), detailCount, "audit failure must preserve the clinical detail")
}
