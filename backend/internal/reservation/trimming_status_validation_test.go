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

// EMR-74 A5: trimming intent の status 遷移バリデータは in_consultation（施術中）を
// 引き続き受理する（汎用 Update のガードとは別経路）。
func TestValidateTrimmingStatusTransitionAllowsInConsultation(t *testing.T) {
	current := model.ReservationStatusCheckedIn
	status := model.ReservationStatusInConsultation
	require.NoError(t, validateTrimmingStatusTransition(current, &status))
}

// EMR-74: 汎用 Update は trimming 予約を in_consultation へ遷移させない。
// in_consultation は受付カンバンの「診療中」（診療行為）であり、trimming の「施術中」は
// trimming intent（UpdateForTrimming）が所有する。
func TestValidateGenericUpdateRejectsTrimmingInConsultation(t *testing.T) {
	inConsultation := model.ReservationStatusInConsultation
	accounting := model.ReservationStatusAccounting
	trimmingType := &model.ReservationType{Category: model.ReservationTypeCategoryTrimming}
	generalType := &model.ReservationType{Category: model.ReservationTypeCategoryGeneral}

	tests := []struct {
		name    string
		current *model.Reservation
		input   *UpdateReservationInput
		wantErr bool
	}{
		{
			name: "trimming checked_in から in_consultation への遷移は拒否",
			current: &model.Reservation{
				Status:          model.ReservationStatusCheckedIn,
				ReservationType: trimmingType,
			},
			input:   &UpdateReservationInput{Status: &inConsultation},
			wantErr: true,
		},
		{
			name: "trimming accounting から in_consultation への遷移は拒否",
			current: &model.Reservation{
				Status:          model.ReservationStatusAccounting,
				ReservationType: trimmingType,
			},
			input:   &UpdateReservationInput{Status: &inConsultation},
			wantErr: true,
		},
		{
			name: "general checked_in から in_consultation への遷移は許可（後段のカルテ検査へ）",
			current: &model.Reservation{
				Status:          model.ReservationStatusCheckedIn,
				ReservationType: generalType,
			},
			input:   &UpdateReservationInput{Status: &inConsultation},
			wantErr: false,
		},
		{
			name: "ReservationType 未 preload は既存挙動（カルテ検査）へフォールスルー",
			current: &model.Reservation{
				Status: model.ReservationStatusCheckedIn,
			},
			input:   &UpdateReservationInput{Status: &inConsultation},
			wantErr: false,
		},
		{
			name: "trimming で status が既に in_consultation の再送は許可",
			current: &model.Reservation{
				Status:          model.ReservationStatusInConsultation,
				ReservationType: trimmingType,
			},
			input:   &UpdateReservationInput{Status: &inConsultation},
			wantErr: false,
		},
		{
			name: "trimming の in_consultation 以外への遷移は許可",
			current: &model.Reservation{
				Status:          model.ReservationStatusCheckedIn,
				ReservationType: trimmingType,
			},
			input:   &UpdateReservationInput{Status: &accounting},
			wantErr: false,
		},
		{
			name: "status 指定なしは対象外",
			current: &model.Reservation{
				Status:          model.ReservationStatusCheckedIn,
				ReservationType: trimmingType,
			},
			input:   &UpdateReservationInput{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGenericUpdateRejectsTrimmingInConsultation(tt.current, tt.input)
			if !tt.wantErr {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.True(t, apperrors.IsInvalidInput(err), "expected InvalidInput but got: %v", err)
		})
	}
}

// EMR-74 A5: trimming intent 経路（UpdateForTrimming）は引き続き in_consultation を受け付ける。
func TestReservationRepository_UpdateForTrimmingAllowsInConsultation(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	reservationType := &model.ReservationType{
		ClinicID: clinicID,
		Name:     "施術中遷移許可区分",
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
		model.ReservationStatusCheckedIn,
		model.ReservationSourceManual,
		nil,
		nil,
	)
	inConsultation := model.ReservationStatusInConsultation

	updated, err := updateForTrimmingInReservationRepoTest(
		ctx,
		db,
		repo,
		clinicID,
		appointment.ID,
		UpdateTrimmingReservationInput{Status: &inConsultation},
	)

	require.NoError(t, err)
	assert.Equal(t, model.ReservationStatusInConsultation, updated.Status)
	assert.Equal(t, model.ReservationStatusInConsultation,
		reloadReservationIntentStatus(t, db, appointment.ID))
}
