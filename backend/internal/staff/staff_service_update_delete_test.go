package staff

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestStaffService_Update(t *testing.T) {
	name := "更新後 スタッフ"
	tests := []struct {
		name     string
		clinicID uint64
		id       uint64
		input    *UpdateStaffInput
		repoErr  error
		wantErr  bool
		wantNF   bool
	}{
		{
			name:     "updates staff successfully",
			clinicID: 1,
			id:       1,
			input:    &UpdateStaffInput{Name: &name},
			repoErr:  nil,
			wantErr:  false,
		},
		{
			name:     "returns not found error when staff does not exist",
			clinicID: 1,
			id:       999,
			input:    &UpdateStaffInput{Name: &name},
			repoErr:  apperrors.WrapNotFound("staff", "999"),
			wantErr:  true,
			wantNF:   true,
		},
		{
			name:     "returns error when no field provided",
			clinicID: 1,
			id:       1,
			input:    &UpdateStaffInput{},
			repoErr:  nil,
			wantErr:  true,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			id:       1,
			input:    &UpdateStaffInput{Name: &name},
			repoErr:  errors.New("db error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockStaffRepository{
				updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
					return tt.repoErr
				},
				findByIDFn: func(_ context.Context, id uint64) (*model.Staff, error) {
					return &model.Staff{ID: id}, nil
				},
			}
			assignmentRepo := &mockAssignmentForStaff{
				lockActiveFn: func(
					_ context.Context,
					staffID uint64,
				) ([]model.StaffClinicAssignment, error) {
					return []model.StaffClinicAssignment{{
						StaffID:  staffID,
						ClinicID: tt.clinicID,
						IsMain:   true,
					}}, nil
				},
			}
			svc := newTestStaffServiceWithAssignmentRepo(repo, assignmentRepo)

			staff, err := svc.Update(
				context.Background(),
				tt.clinicID,
				tt.id,
				authorizedStaffUpdate(tt.input, tt.clinicID),
			)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, staff)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, staff)
			}
		})
	}
}

func TestStaffService_Reorder(t *testing.T) {
	tests := []struct {
		name             string
		ids              []uint64
		repoErr          error
		wantErr          bool
		wantInvalidInput bool
	}{
		{
			name:    "reorders successfully",
			ids:     []uint64{3, 1, 2},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:             "returns invalid input when ids is empty",
			ids:              []uint64{},
			wantErr:          true,
			wantInvalidInput: true,
		},
		{
			name:    "propagates repository error",
			ids:     []uint64{1, 2},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockStaffRepository{reorderErr: tt.repoErr}
			svc := newTestStaffService(repo)

			err := svc.Reorder(context.Background(), 1, tt.ids)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantInvalidInput {
					assert.True(t, apperrors.IsInvalidInput(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStaffService_Delete(t *testing.T) {
	tests := []struct {
		name                string
		clinicID            uint64
		id                  uint64
		findByIDErr         error
		reservationExists   bool
		shiftExists         bool
		checkReservationErr error
		checkShiftErr       error
		blockingRefs        []StaffDependencyCount
		blockingErr         error
		repoErr             error
		wantErr             bool
		wantNF              bool
		wantConflict        bool
	}{
		{
			name:                "deletes staff successfully when no dependencies exist",
			clinicID:            1,
			id:                  10,
			reservationExists:   false,
			shiftExists:         false,
			checkReservationErr: nil,
			checkShiftErr:       nil,
			repoErr:             nil,
			wantErr:             false,
		},
		{
			name:        "returns not found error when staff does not exist",
			clinicID:    1,
			id:          999,
			findByIDErr: apperrors.WrapNotFound("staff", "999"),
			wantErr:     true,
			wantNF:      true,
		},
		{
			name:                "returns conflict error when staff has reservations",
			clinicID:            1,
			id:                  10,
			reservationExists:   true,
			shiftExists:         false,
			checkReservationErr: nil,
			checkShiftErr:       nil,
			repoErr:             nil,
			wantErr:             true,
			wantConflict:        true,
		},
		{
			name:                "returns conflict error when staff has shift entries",
			clinicID:            1,
			id:                  10,
			reservationExists:   false,
			shiftExists:         true,
			checkReservationErr: nil,
			checkShiftErr:       nil,
			repoErr:             nil,
			wantErr:             true,
			wantConflict:        true,
		},
		{
			name:                "returns conflict error when staff has both reservations and shifts",
			clinicID:            1,
			id:                  10,
			reservationExists:   true,
			shiftExists:         true,
			checkReservationErr: nil,
			checkShiftErr:       nil,
			repoErr:             nil,
			wantErr:             true,
			wantConflict:        true,
		},
		{
			name:                "returns error when reservation check fails",
			clinicID:            1,
			id:                  10,
			reservationExists:   false,
			shiftExists:         false,
			checkReservationErr: errors.New("db error"),
			checkShiftErr:       nil,
			repoErr:             nil,
			wantErr:             true,
		},
		{
			name:                "returns error when shift check fails",
			clinicID:            1,
			id:                  10,
			reservationExists:   false,
			shiftExists:         false,
			checkReservationErr: nil,
			checkShiftErr:       errors.New("db error"),
			repoErr:             nil,
			wantErr:             true,
		},
		{
			name:         "returns conflict error when staff has blocking dependencies",
			clinicID:     1,
			id:           10,
			blockingRefs: []StaffDependencyCount{{Label: "カルテ追記", Count: 1}},
			wantErr:      true,
			wantConflict: true,
		},
		{
			name:        "returns error when dependency check fails",
			clinicID:    1,
			id:          10,
			blockingErr: errors.New("db error"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockStaffRepository{
				findByIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
					if tt.findByIDErr != nil {
						return nil, tt.findByIDErr
					}
					return &model.Staff{ID: tt.id}, nil
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
				countBlockingRefsFn: func(_ context.Context, _, _ uint64) ([]StaffDependencyCount, error) {
					return tt.blockingRefs, tt.blockingErr
				},
			}
			reservationRepo := &mockReservationForStaff{
				existsByStaffIDFn: func(_ context.Context, _, _ uint64) (bool, error) {
					return tt.reservationExists, tt.checkReservationErr
				},
			}
			shiftRepo := &mockShiftEntryForStaff{
				existsByStaffIDFn: func(_ context.Context, _, _ uint64) (bool, error) {
					return tt.shiftExists, tt.checkShiftErr
				},
			}
			assignmentRepo := &mockAssignmentForStaff{
				lockActiveFn: func(_ context.Context, staffID uint64) ([]model.StaffClinicAssignment, error) {
					return []model.StaffClinicAssignment{{
						StaffID:  staffID,
						ClinicID: tt.clinicID,
						IsMain:   true,
					}}, nil
				},
			}
			svc := NewStaffService(repo, &mockAccountForStaff{}, assignmentRepo, reservationRepo, shiftRepo, &mockPermissionGroupRepository{}, &mockResStaffForStaff{}, nil, nil, noopTransactor{})

			err := svc.Delete(context.Background(), tt.clinicID, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
				if tt.wantConflict {
					assert.True(t, apperrors.IsConflict(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
