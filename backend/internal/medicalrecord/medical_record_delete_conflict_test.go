package medicalrecord

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestMedicalRecordService_DeleteConflictKindsPreservePublicContract(t *testing.T) {
	tests := []struct {
		name              string
		lockedStatus      model.MedicalRecordStatus
		estimateCount     int64
		appointmentStatus model.ReservationStatus
		wantKind          medicalRecordDeleteConflictKind
		wantMessage       string
	}{
		{
			name:         "locked record state changed",
			lockedStatus: model.MedicalRecordStatusFinalized,
			wantKind:     medicalRecordDeleteStateConflict,
			wantMessage:  "確定済みまたは下書き以外の診療記録は削除できません",
		},
		{
			name:              "linked appointment already in consultation",
			lockedStatus:      model.MedicalRecordStatusDraft,
			appointmentStatus: model.ReservationStatusInConsultation,
			wantKind:          medicalRecordDeleteStateConflict,
			wantMessage:       "診療中の予約に紐づくカルテは削除できません",
		},
		{
			name:          "estimate dependency exists",
			lockedStatus:  model.MedicalRecordStatusDraft,
			estimateCount: 1,
			wantKind:      medicalRecordDeleteDependencyConflict,
			wantMessage:   "この項目は使用中のため削除できません",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockMedicalRecordRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{
						ID:       91,
						ClinicID: 3,
						Status:   tt.lockedStatus,
					}, nil
				},
				lockLinkedAppointmentForUpdateFn: func(_ context.Context, _, _ uint64) (*linkedAppointmentLock, error) {
					if tt.appointmentStatus == "" {
						return &linkedAppointmentLock{}, nil
					}
					return &linkedAppointmentLock{Appointment: &model.Reservation{ID: 7, ClinicID: 3, Status: tt.appointmentStatus}}, nil
				},
				countEstimatesByMedicalRecordIDFn: func(_ context.Context, _ uint64) (int64, error) {
					return tt.estimateCount, nil
				},
			}
			svc := NewMedicalRecordServiceWithTxAudit(
				repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, &mockTransactor{})

			err := svc.Delete(context.Background(), 3, 91)

			require.Error(t, err)
			assert.True(t, apperrors.IsConflict(err))
			kind, ok := medicalRecordDeleteConflictKindFromError(err)
			require.True(t, ok)
			assert.Equal(t, tt.wantKind, kind)
			var appErr *apperrors.AppError
			require.ErrorAs(t, err, &appErr)
			assert.Equal(t, "CONFLICT", appErr.Code)
			assert.Equal(t, tt.wantMessage, appErr.Message)
			assert.True(t, errors.Is(appErr, apperrors.ErrConflict))
		})
	}
}
