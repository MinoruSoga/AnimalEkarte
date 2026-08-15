package reservation

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestValidateTrimmingCreateStatus(t *testing.T) {
	tests := []struct {
		name    string
		status  model.ReservationStatus
		wantErr bool
	}{
		{name: "pending is accepted", status: model.ReservationStatusPending},
		{name: "accounting is accepted", status: model.ReservationStatusAccounting},
		{name: "completed is transition owned", status: model.ReservationStatusCompleted, wantErr: true},
		{name: "no show is transition owned", status: model.ReservationStatusNoShow, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTrimmingCreateStatus(tt.status)
			if !tt.wantErr {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.True(t, apperrors.IsInvalidInput(err))
		})
	}
}

func TestReservationRepository_CreateForTrimmingRejectsCompletedStatus(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	reservationType := &model.ReservationType{
		ClinicID: clinicID,
		Name:     "完了直接作成拒否区分",
		Category: model.ReservationTypeCategoryTrimming,
		IsActive: true,
	}
	require.NoError(t, db.Create(reservationType).Error)
	start := time.Now().UTC().Add(time.Hour)

	_, err := createForTrimmingInReservationRepoTest(ctx, db, repo, clinicID, CreateTrimmingReservationInput{
		ReservationTypeID: reservationType.ID,
		StartTime:         start,
		EndTime:           start.Add(time.Hour),
		Status:            model.ReservationStatusCompleted,
	})
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))

	var count int64
	require.NoError(t, db.Model(&model.Reservation{}).
		Where("clinic_id = ? AND reservation_type_id = ?", clinicID, reservationType.ID).
		Count(&count).Error)
	assert.Zero(t, count)
}

func TestReservationRepository_UpdateForTrimmingRejectsCompletedStatus(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	reservationType := &model.ReservationType{
		ClinicID: clinicID,
		Name:     "完了直接更新拒否区分",
		Category: model.ReservationTypeCategoryTrimming,
		IsActive: true,
	}
	require.NoError(t, db.Create(reservationType).Error)
	appointment := makeReservationForReservationRepoTest(
		t,
		db,
		clinicID,
		reservationType.ID,
		time.Now().UTC().Add(time.Hour),
		model.ReservationStatusPending,
		model.ReservationSourceManual,
		nil,
		nil,
	)
	completed := model.ReservationStatusCompleted

	_, err := updateForTrimmingInReservationRepoTest(
		ctx,
		db,
		repo,
		clinicID,
		appointment.ID,
		UpdateTrimmingReservationInput{Status: &completed},
	)
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Equal(t, model.ReservationStatusPending, reloadReservationIntentStatus(t, db, appointment.ID))
}

func TestValidateTrimmingStatusTransitionRejectsTransitionOwnedStatuses(t *testing.T) {
	current := model.ReservationStatusPending
	statuses := []model.ReservationStatus{
		model.ReservationStatusCompleted,
		model.ReservationStatusNoShow,
	}

	for _, status := range statuses {
		t.Run(string(status), func(t *testing.T) {
			err := validateTrimmingStatusTransition(current, &status)
			require.Error(t, err)
			assert.True(t, apperrors.IsInvalidInput(err))
		})
	}
}
