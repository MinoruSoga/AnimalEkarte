package staff

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestService_SetClinicAssignments_ChecksReservationDependencyBeforeRemoval(t *testing.T) {
	dependencyErr := errors.New("reservation dependency lookup failed")
	tests := []struct {
		name              string
		reservationExists bool
		reservationErr    error
		wantConflict      bool
		wantErr           error
		wantMutated       bool
		wantEvents        []string
	}{
		{
			name:              "予約が存在する所属は解除しない",
			reservationExists: true,
			wantConflict:      true,
			wantMutated:       false,
			wantEvents: []string{
				"lock-staff",
				"lock-assignments",
				"lock-clinic:2",
				"check-reservation:1",
			},
		},
		{
			name:           "予約依存照会エラーは変更前にfail-closed",
			reservationErr: dependencyErr,
			wantErr:        dependencyErr,
			wantMutated:    false,
			wantEvents: []string{
				"lock-staff",
				"lock-assignments",
				"lock-clinic:2",
				"check-reservation:1",
			},
		},
		{
			name:        "予約もシフトもなければ所属を置換する",
			wantMutated: true,
			wantEvents: []string{
				"lock-staff",
				"lock-assignments",
				"lock-clinic:2",
				"check-reservation:1",
				"check-shift:1",
				"delete",
				"restore:2",
				"primary:2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events := make([]string, 0, len(tt.wantEvents))
			mutated := false
			staffRepo := &mockStaffRepository{
				lockForUpdateFn: func(ctx context.Context, id uint64) (*model.Staff, error) {
					requireStaffSecurityTxContext(t, ctx)
					events = append(events, "lock-staff")
					return &model.Staff{ID: id}, nil
				},
				updatePrimaryFn: func(ctx context.Context, _, clinicID uint64) error {
					requireStaffSecurityTxContext(t, ctx)
					mutated = true
					events = append(events, fmt.Sprintf("primary:%d", clinicID))
					return nil
				},
			}
			assignmentRepo := &mockAssignmentForStaff{
				lockActiveFn: func(ctx context.Context, staffID uint64) ([]model.StaffClinicAssignment, error) {
					requireStaffSecurityTxContext(t, ctx)
					events = append(events, "lock-assignments")
					return []model.StaffClinicAssignment{
						{StaffID: staffID, ClinicID: 1, IsMain: true},
					}, nil
				},
				deleteByStaffIDFn: func(ctx context.Context, _ uint64) error {
					requireStaffSecurityTxContext(t, ctx)
					mutated = true
					events = append(events, "delete")
					return nil
				},
				restoreOrCreateFn: func(ctx context.Context, assignment *model.StaffClinicAssignment) error {
					requireStaffSecurityTxContext(t, ctx)
					mutated = true
					events = append(events, fmt.Sprintf("restore:%d", assignment.ClinicID))
					return nil
				},
			}
			reservationUsage := &mockReservationForStaff{
				existsByStaffIDFn: func(ctx context.Context, clinicID, staffID uint64) (bool, error) {
					requireStaffSecurityTxContext(t, ctx)
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(10), staffID)
					events = append(events, "check-reservation:1")
					return tt.reservationExists, tt.reservationErr
				},
			}
			shiftRepo := &mockShiftEntryForStaff{
				existsByStaffIDFn: func(ctx context.Context, clinicID, staffID uint64) (bool, error) {
					requireStaffSecurityTxContext(t, ctx)
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(10), staffID)
					events = append(events, "check-shift:1")
					return false, nil
				},
			}
			clinicRepo := &mockClinicLookupForStaffAssignments{
				lockActiveByIDFn: func(ctx context.Context, clinicID uint64) (*model.Clinic, error) {
					requireStaffSecurityTxContext(t, ctx)
					events = append(events, fmt.Sprintf("lock-clinic:%d", clinicID))
					return &model.Clinic{ID: clinicID, IsActive: true}, nil
				},
			}
			svc := NewService(
				staffRepo,
				&mockAccountForStaff{},
				assignmentRepo,
				reservationUsage,
				shiftRepo,
				&mockPermissionGroupRepository{},
				&mockResStaffForStaff{},
				nil,
				clinicRepo,
				markedStaffSecurityTransactor{},
			)

			err := svc.SetClinicAssignments(context.Background(), &SetClinicAssignmentsInput{
				StaffID:             10,
				ClinicIDs:           []uint64{2},
				AuthorizedClinicIDs: []uint64{1, 2},
			})

			switch {
			case tt.wantConflict:
				require.Error(t, err)
				assert.True(t, apperrors.IsConflict(err), "unexpected error: %v", err)
			case tt.wantErr != nil:
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			default:
				require.NoError(t, err)
			}
			assert.Equal(t, tt.wantMutated, mutated)
			assert.Equal(t, tt.wantEvents, events)
		})
	}
}

func TestService_SetClinicAssignments_MissingReservationUsageFailsClosed(t *testing.T) {
	mutated := false
	assignmentRepo := &mockAssignmentForStaff{
		lockActiveFn: func(_ context.Context, staffID uint64) ([]model.StaffClinicAssignment, error) {
			return []model.StaffClinicAssignment{{StaffID: staffID, ClinicID: 1, IsMain: true}}, nil
		},
		deleteByStaffIDFn: func(_ context.Context, _ uint64) error {
			mutated = true
			return nil
		},
	}
	svc := NewService(
		&mockStaffRepository{},
		&mockAccountForStaff{},
		assignmentRepo,
		nil,
		&mockShiftEntryForStaff{},
		&mockPermissionGroupRepository{},
		&mockResStaffForStaff{},
		nil,
		existingClinicLookupForStaffAssignments(),
		noopTransactor{},
	)

	err := svc.SetClinicAssignments(context.Background(), &SetClinicAssignmentsInput{
		StaffID:             10,
		ClinicIDs:           []uint64{2},
		AuthorizedClinicIDs: []uint64{1, 2},
	})

	require.Error(t, err)
	assert.False(t, mutated)
}

func TestService_SetClinicAssignments_UsesDeterministicClinicAndDependencyOrder(t *testing.T) {
	events := make([]string, 0, 16)
	input := &SetClinicAssignmentsInput{
		StaffID:             10,
		ClinicIDs:           []uint64{3, 2},
		AuthorizedClinicIDs: []uint64{1, 2, 3, 4},
	}
	staffRepo := &mockStaffRepository{
		lockForUpdateFn: func(_ context.Context, id uint64) (*model.Staff, error) {
			events = append(events, "lock-staff")
			return &model.Staff{ID: id}, nil
		},
		updatePrimaryFn: func(_ context.Context, _, clinicID uint64) error {
			events = append(events, fmt.Sprintf("primary:%d", clinicID))
			return nil
		},
	}
	assignmentRepo := &mockAssignmentForStaff{
		lockActiveFn: func(_ context.Context, staffID uint64) ([]model.StaffClinicAssignment, error) {
			events = append(events, "lock-assignments")
			return []model.StaffClinicAssignment{
				{StaffID: staffID, ClinicID: 4},
				{StaffID: staffID, ClinicID: 1, IsMain: true},
			}, nil
		},
		deleteByStaffIDFn: func(_ context.Context, _ uint64) error {
			events = append(events, "delete")
			return nil
		},
		restoreOrCreateFn: func(_ context.Context, assignment *model.StaffClinicAssignment) error {
			events = append(events, fmt.Sprintf("restore:%d", assignment.ClinicID))
			return nil
		},
	}
	reservationUsage := &mockReservationForStaff{
		existsByStaffIDFn: func(_ context.Context, clinicID, _ uint64) (bool, error) {
			events = append(events, fmt.Sprintf("check-reservation:%d", clinicID))
			return false, nil
		},
	}
	shiftRepo := &mockShiftEntryForStaff{
		existsByStaffIDFn: func(_ context.Context, clinicID, _ uint64) (bool, error) {
			events = append(events, fmt.Sprintf("check-shift:%d", clinicID))
			return false, nil
		},
	}
	clinicRepo := &mockClinicLookupForStaffAssignments{
		lockActiveByIDFn: func(_ context.Context, clinicID uint64) (*model.Clinic, error) {
			events = append(events, fmt.Sprintf("lock-clinic:%d", clinicID))
			return &model.Clinic{ID: clinicID, IsActive: true}, nil
		},
	}
	svc := NewService(
		staffRepo,
		&mockAccountForStaff{},
		assignmentRepo,
		reservationUsage,
		shiftRepo,
		&mockPermissionGroupRepository{},
		&mockResStaffForStaff{},
		nil,
		clinicRepo,
		noopTransactor{},
	)

	err := svc.SetClinicAssignments(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, []string{
		"lock-staff",
		"lock-assignments",
		"lock-clinic:2",
		"lock-clinic:3",
		"check-reservation:1",
		"check-reservation:4",
		"check-shift:1",
		"check-shift:4",
		"delete",
		"restore:3",
		"restore:2",
		"primary:3",
	}, events)
	assert.Equal(t, []uint64{3, 2}, input.ClinicIDs, "caller-owned order must remain unchanged")
}

func TestService_SetClinicAssignments_BatchesRemovedClinicDependencyChecks(t *testing.T) {
	events := make([]string, 0, 12)
	staffRepo := &mockStaffRepository{
		lockForUpdateFn: func(ctx context.Context, id uint64) (*model.Staff, error) {
			requireStaffSecurityTxContext(t, ctx)
			events = append(events, "lock-staff")
			return &model.Staff{ID: id}, nil
		},
		updatePrimaryFn: func(ctx context.Context, _, clinicID uint64) error {
			requireStaffSecurityTxContext(t, ctx)
			events = append(events, fmt.Sprintf("primary:%d", clinicID))
			return nil
		},
	}
	assignmentRepo := &mockAssignmentForStaff{
		lockActiveFn: func(ctx context.Context, staffID uint64) ([]model.StaffClinicAssignment, error) {
			requireStaffSecurityTxContext(t, ctx)
			events = append(events, "lock-assignments")
			return []model.StaffClinicAssignment{
				{StaffID: staffID, ClinicID: 4},
				{StaffID: staffID, ClinicID: 1, IsMain: true},
			}, nil
		},
		deleteByStaffIDFn: func(ctx context.Context, _ uint64) error {
			requireStaffSecurityTxContext(t, ctx)
			events = append(events, "delete")
			return nil
		},
		restoreOrCreateFn: func(ctx context.Context, assignment *model.StaffClinicAssignment) error {
			requireStaffSecurityTxContext(t, ctx)
			events = append(events, fmt.Sprintf("restore:%d", assignment.ClinicID))
			return nil
		},
	}
	reservationUsage := &mockReservationForStaff{
		existsByStaffIDFn: func(_ context.Context, _, _ uint64) (bool, error) {
			t.Fatal("removed-clinic reservation checks must be batched")
			return false, nil
		},
		findClinicIDsByStaffIDFn: func(ctx context.Context, clinicIDs []uint64, staffID uint64) ([]uint64, error) {
			requireStaffSecurityTxContext(t, ctx)
			assert.Equal(t, []uint64{1, 4}, clinicIDs)
			assert.Equal(t, uint64(10), staffID)
			events = append(events, "check-reservations:[1 4]")
			return nil, nil
		},
	}
	shiftRepo := &mockShiftEntryForStaff{
		existsByStaffIDFn: func(_ context.Context, _, _ uint64) (bool, error) {
			t.Fatal("removed-clinic shift checks must be batched")
			return false, nil
		},
		findClinicIDsByStaffIDFn: func(ctx context.Context, clinicIDs []uint64, staffID uint64) ([]uint64, error) {
			requireStaffSecurityTxContext(t, ctx)
			assert.Equal(t, []uint64{1, 4}, clinicIDs)
			assert.Equal(t, uint64(10), staffID)
			events = append(events, "check-shifts:[1 4]")
			return nil, nil
		},
	}
	clinicRepo := &mockClinicLookupForStaffAssignments{
		lockActiveByIDFn: func(ctx context.Context, clinicID uint64) (*model.Clinic, error) {
			requireStaffSecurityTxContext(t, ctx)
			events = append(events, fmt.Sprintf("lock-clinic:%d", clinicID))
			return &model.Clinic{ID: clinicID, IsActive: true}, nil
		},
	}
	svc := NewService(
		staffRepo,
		&mockAccountForStaff{},
		assignmentRepo,
		reservationUsage,
		shiftRepo,
		&mockPermissionGroupRepository{},
		&mockResStaffForStaff{},
		nil,
		clinicRepo,
		markedStaffSecurityTransactor{},
	)

	err := svc.SetClinicAssignments(context.Background(), &SetClinicAssignmentsInput{
		StaffID:             10,
		ClinicIDs:           []uint64{3, 2},
		AuthorizedClinicIDs: []uint64{1, 2, 3, 4},
	})

	require.NoError(t, err)
	assert.Equal(t, []string{
		"lock-staff",
		"lock-assignments",
		"lock-clinic:2",
		"lock-clinic:3",
		"check-reservations:[1 4]",
		"check-shifts:[1 4]",
		"delete",
		"restore:3",
		"restore:2",
		"primary:3",
	}, events)
}
