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

func TestStaffService_SetClinicAssignments_ErrorBranches(t *testing.T) {
	t.Run("delete existing assignments fails", func(t *testing.T) {
		assignmentRepo := &mockAssignmentForStaff{
			lockActiveFn: func(_ context.Context, staffID uint64) ([]model.StaffClinicAssignment, error) {
				return []model.StaffClinicAssignment{{StaffID: staffID, ClinicID: 2, IsMain: true}}, nil
			},
			deleteByClinicIDsFn: func(_ context.Context, _ uint64, clinicIDs []uint64) error {
				assert.Equal(t, []uint64{2}, clinicIDs)
				return errors.New("db error")
			},
		}
		svc := NewStaffService(&mockStaffRepository{}, &mockAccountForStaff{}, assignmentRepo, &mockReservationForStaff{}, &mockShiftEntryForStaff{}, &mockPermissionGroupRepository{}, &mockResStaffForStaff{}, nil, existingClinicLookupForStaffAssignments(), noopTransactor{})

		err := svc.SetClinicAssignments(context.Background(), &SetClinicAssignmentsInput{
			StaffID:             10,
			ClinicIDs:           []uint64{1},
			AuthorizedClinicIDs: []uint64{1, 2},
		})

		assert.Error(t, err)
	})

	t.Run("create assignment fails", func(t *testing.T) {
		assignmentRepo := &mockAssignmentForStaff{
			createFn: func(_ context.Context, _ *model.StaffClinicAssignment) error { return errors.New("db error") },
		}
		svc := NewStaffService(&mockStaffRepository{}, &mockAccountForStaff{}, assignmentRepo, &mockReservationForStaff{}, &mockShiftEntryForStaff{}, &mockPermissionGroupRepository{}, &mockResStaffForStaff{}, nil, existingClinicLookupForStaffAssignments(), noopTransactor{})

		err := svc.SetClinicAssignments(context.Background(), &SetClinicAssignmentsInput{
			StaffID:             10,
			ClinicIDs:           []uint64{1, 2},
			AuthorizedClinicIDs: []uint64{1, 2},
		})

		assert.Error(t, err)
	})

	t.Run("update primary clinic id fails", func(t *testing.T) {
		staffRepo := &mockStaffRepository{
			updatePrimaryFn: func(_ context.Context, _, _ uint64) error { return errors.New("db error") },
		}
		svc := NewStaffService(staffRepo, &mockAccountForStaff{}, &mockAssignmentForStaff{}, &mockReservationForStaff{}, &mockShiftEntryForStaff{}, &mockPermissionGroupRepository{}, &mockResStaffForStaff{}, nil, existingClinicLookupForStaffAssignments(), noopTransactor{})

		err := svc.SetClinicAssignments(context.Background(), &SetClinicAssignmentsInput{
			StaffID:             10,
			ClinicIDs:           []uint64{1},
			AuthorizedClinicIDs: []uint64{1},
		})

		assert.Error(t, err)
	})
}

func TestStaffService_SetClinicAssignments_AuthorizationAndValidationBeforeWrites(t *testing.T) {
	t.Run("more than 50 requested clinics fails before dependencies", func(t *testing.T) {
		clinicIDs := make([]uint64, maxStaffClinicAssignments+1)
		for i := range clinicIDs {
			clinicIDs[i] = uint64(i + 1)
		}
		svc := NewStaffService(
			&mockStaffRepository{},
			&mockAccountForStaff{},
			&mockAssignmentForStaff{},
			&mockReservationForStaff{},
			&mockShiftEntryForStaff{},
			&mockPermissionGroupRepository{},
			&mockResStaffForStaff{},
			nil,
			nil,
			nil,
		)

		err := svc.SetClinicAssignments(context.Background(), &SetClinicAssignmentsInput{
			StaffID:             10,
			ClinicIDs:           clinicIDs,
			AuthorizedClinicIDs: clinicIDs,
		})

		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("non-admin cannot assign a clinic outside authenticated clinic ids", func(t *testing.T) {
		clinicRepo := &mockClinicLookupForStaffAssignments{
			lockActiveByIDFn: func(_ context.Context, _ uint64) (*model.Clinic, error) {
				t.Fatal("clinic lookup must not reveal an unauthorized clinic")
				return nil, nil
			},
		}
		assignmentRepo := &mockAssignmentForStaff{
			deleteByStaffIDFn: func(_ context.Context, _ uint64) error {
				t.Fatal("assignments must remain unchanged after authorization failure")
				return nil
			},
		}
		svc := NewStaffService(
			&mockStaffRepository{},
			&mockAccountForStaff{},
			assignmentRepo,
			&mockReservationForStaff{},
			&mockShiftEntryForStaff{},
			&mockPermissionGroupRepository{},
			&mockResStaffForStaff{},
			nil,
			clinicRepo,
			noopTransactor{},
		)

		err := svc.SetClinicAssignments(context.Background(), &SetClinicAssignmentsInput{
			StaffID:             10,
			ClinicIDs:           []uint64{1, 99},
			AuthorizedClinicIDs: []uint64{1, 2},
			IsSystemAdmin:       false,
		})

		require.Error(t, err)
		assert.ErrorIs(t, err, apperrors.ErrForbidden)
	})

	t.Run("non-admin cannot remove a clinic outside authorized clinic ids", func(t *testing.T) {
		assignmentRepo := &mockAssignmentForStaff{
			lockActiveFn: func(_ context.Context, staffID uint64) ([]model.StaffClinicAssignment, error) {
				return []model.StaffClinicAssignment{
					{StaffID: staffID, ClinicID: 1, IsMain: true},
					{StaffID: staffID, ClinicID: 2, IsMain: false},
				}, nil
			},
			deleteByClinicIDsFn: func(_ context.Context, _ uint64, _ []uint64) error {
				t.Fatal("assignments must remain unchanged after authorization failure")
				return nil
			},
		}
		svc := NewStaffService(
			&mockStaffRepository{},
			&mockAccountForStaff{},
			assignmentRepo,
			&mockReservationForStaff{},
			&mockShiftEntryForStaff{},
			&mockPermissionGroupRepository{},
			&mockResStaffForStaff{},
			nil,
			existingClinicLookupForStaffAssignments(),
			noopTransactor{},
		)

		err := svc.SetClinicAssignments(context.Background(), &SetClinicAssignmentsInput{
			StaffID:             10,
			ClinicIDs:           []uint64{1},
			AuthorizedClinicIDs: []uint64{1},
			IsSystemAdmin:       false,
		})

		require.Error(t, err)
		assert.ErrorIs(t, err, apperrors.ErrForbidden)
	})

	t.Run("all unique clinics are validated before delete and duplicates are removed", func(t *testing.T) {
		events := make([]string, 0, 7)
		created := make([]model.StaffClinicAssignment, 0, 2)
		clinicRepo := &mockClinicLookupForStaffAssignments{
			lockActiveByIDFn: func(_ context.Context, id uint64) (*model.Clinic, error) {
				events = append(events, fmt.Sprintf("find:%d", id))
				return &model.Clinic{ID: id}, nil
			},
		}
		assignmentRepo := &mockAssignmentForStaff{
			deleteByStaffIDFn: func(_ context.Context, staffID uint64) error {
				assert.Equal(t, uint64(10), staffID)
				events = append(events, "delete")
				return nil
			},
			createFn: func(_ context.Context, assignment *model.StaffClinicAssignment) error {
				events = append(events, fmt.Sprintf("create:%d", assignment.ClinicID))
				created = append(created, *assignment)
				return nil
			},
		}
		staffRepo := &mockStaffRepository{
			updatePrimaryFn: func(_ context.Context, staffID, clinicID uint64) error {
				assert.Equal(t, uint64(10), staffID)
				assert.Equal(t, uint64(2), clinicID)
				events = append(events, "primary:2")
				return nil
			},
		}
		svc := NewStaffService(
			staffRepo,
			&mockAccountForStaff{},
			assignmentRepo,
			&mockReservationForStaff{},
			&mockShiftEntryForStaff{},
			&mockPermissionGroupRepository{},
			&mockResStaffForStaff{},
			nil,
			clinicRepo,
			noopTransactor{},
		)

		err := svc.SetClinicAssignments(context.Background(), &SetClinicAssignmentsInput{
			StaffID:             10,
			ClinicIDs:           []uint64{2, 2, 4, 2},
			AuthorizedClinicIDs: []uint64{1, 2, 4},
			IsSystemAdmin:       false,
		})

		require.NoError(t, err)
		assert.Equal(t, []string{"find:2", "find:4", "create:2", "create:4", "primary:2"}, events)
		require.Len(t, created, 2)
		assert.True(t, created[0].IsMain)
		assert.False(t, created[1].IsMain)
	})

	t.Run("missing clinic leaves existing assignments untouched", func(t *testing.T) {
		clinicRepo := &mockClinicLookupForStaffAssignments{
			lockActiveByIDFn: func(_ context.Context, id uint64) (*model.Clinic, error) {
				if id == 4 {
					return nil, apperrors.WrapNotFound("clinic", "4")
				}
				return &model.Clinic{ID: id}, nil
			},
		}
		assignmentRepo := &mockAssignmentForStaff{
			deleteByStaffIDFn: func(_ context.Context, _ uint64) error {
				t.Fatal("all clinics must be validated before deleting existing assignments")
				return nil
			},
		}
		svc := NewStaffService(
			&mockStaffRepository{},
			&mockAccountForStaff{},
			assignmentRepo,
			&mockReservationForStaff{},
			&mockShiftEntryForStaff{},
			&mockPermissionGroupRepository{},
			&mockResStaffForStaff{},
			nil,
			clinicRepo,
			noopTransactor{},
		)

		err := svc.SetClinicAssignments(context.Background(), &SetClinicAssignmentsInput{
			StaffID:             10,
			ClinicIDs:           []uint64{2, 4},
			AuthorizedClinicIDs: []uint64{2, 4},
			IsSystemAdmin:       false,
		})

		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("system admin can assign any existing clinic", func(t *testing.T) {
		var createdClinicID uint64
		clinicRepo := &mockClinicLookupForStaffAssignments{
			lockActiveByIDFn: func(_ context.Context, id uint64) (*model.Clinic, error) {
				return &model.Clinic{ID: id}, nil
			},
		}
		assignmentRepo := &mockAssignmentForStaff{
			createFn: func(_ context.Context, assignment *model.StaffClinicAssignment) error {
				createdClinicID = assignment.ClinicID
				return nil
			},
		}
		svc := NewStaffService(
			&mockStaffRepository{},
			&mockAccountForStaff{},
			assignmentRepo,
			&mockReservationForStaff{},
			&mockShiftEntryForStaff{},
			&mockPermissionGroupRepository{},
			&mockResStaffForStaff{},
			nil,
			clinicRepo,
			noopTransactor{},
		)

		err := svc.SetClinicAssignments(context.Background(), &SetClinicAssignmentsInput{
			StaffID:       10,
			ClinicIDs:     []uint64{99},
			IsSystemAdmin: true,
		})

		require.NoError(t, err)
		assert.Equal(t, uint64(99), createdClinicID)
	})
}

func TestStaffService_SetClinicAssignments_RollsBackReplacementOnWriteError(t *testing.T) {
	state := &staffAssignmentTxState{
		assignments: []model.StaffClinicAssignment{
			{StaffID: 10, ClinicID: 1, IsMain: true},
		},
		primaryClinicID: 1,
	}
	clinicRepo := &mockClinicLookupForStaffAssignments{
		lockActiveByIDFn: func(_ context.Context, id uint64) (*model.Clinic, error) {
			return &model.Clinic{ID: id}, nil
		},
	}
	assignmentRepo := &mockAssignmentForStaff{
		deleteByStaffIDFn: func(_ context.Context, _ uint64) error {
			state.assignments = nil
			return nil
		},
		createFn: func(_ context.Context, assignment *model.StaffClinicAssignment) error {
			if assignment.ClinicID == 3 {
				return errors.New("create failed")
			}
			state.assignments = append(state.assignments, *assignment)
			return nil
		},
	}
	staffRepo := &mockStaffRepository{
		updatePrimaryFn: func(_ context.Context, _, clinicID uint64) error {
			state.primaryClinicID = clinicID
			return nil
		},
	}
	svc := NewStaffService(
		staffRepo,
		&mockAccountForStaff{},
		assignmentRepo,
		&mockReservationForStaff{},
		&mockShiftEntryForStaff{},
		&mockPermissionGroupRepository{},
		&mockResStaffForStaff{},
		nil,
		clinicRepo,
		&rollbackStaffAssignmentTransactor{state: state},
	)

	err := svc.SetClinicAssignments(context.Background(), &SetClinicAssignmentsInput{
		StaffID:             10,
		ClinicIDs:           []uint64{2, 3},
		AuthorizedClinicIDs: []uint64{1, 2, 3},
		IsSystemAdmin:       false,
	})

	require.Error(t, err)
	assert.Equal(t, []model.StaffClinicAssignment{{StaffID: 10, ClinicID: 1, IsMain: true}}, state.assignments)
	assert.Equal(t, uint64(1), state.primaryClinicID)
}

func TestStaffService_SetClinicAssignments_RejectsNonAdminWhenExistingAssignmentOutsideAuthorizedClinics(t *testing.T) {
	t.Run("non-admin PUT fails closed when existing assignment is outside authorized clinics", func(t *testing.T) {
		events := make([]string, 0, 2)
		staffRepo := &mockStaffRepository{
			lockForUpdateFn: func(ctx context.Context, id uint64) (*model.Staff, error) {
				requireStaffSecurityTxContext(t, ctx)
				assert.Equal(t, uint64(10), id)
				events = append(events, "lock-staff")
				return &model.Staff{ID: id, ClinicID: 1}, nil
			},
			updatePrimaryFn: func(_ context.Context, _, _ uint64) error {
				t.Fatal("primary must not mutate after unauthorized existing-assignment rejection")
				return nil
			},
		}
		assignmentRepo := &mockAssignmentForStaff{
			lockActiveFn: func(ctx context.Context, staffID uint64) ([]model.StaffClinicAssignment, error) {
				requireStaffSecurityTxContext(t, ctx)
				assert.Equal(t, uint64(10), staffID)
				events = append(events, "lock-assignments")
				return []model.StaffClinicAssignment{
					{StaffID: staffID, ClinicID: 1, IsMain: true},
					{StaffID: staffID, ClinicID: 2},
				}, nil
			},
			deleteByStaffIDFn: func(_ context.Context, _ uint64) error {
				t.Fatal("delete must not run after unauthorized existing-assignment rejection")
				return nil
			},
			deleteByClinicIDsFn: func(_ context.Context, _ uint64, _ []uint64) error {
				t.Fatal("delete must not run after unauthorized existing-assignment rejection")
				return nil
			},
			restoreOrCreateFn: func(_ context.Context, _ *model.StaffClinicAssignment) error {
				t.Fatal("restore must not run after unauthorized existing-assignment rejection")
				return nil
			},
			createFn: func(_ context.Context, _ *model.StaffClinicAssignment) error {
				t.Fatal("create must not run after unauthorized existing-assignment rejection")
				return nil
			},
		}
		clinicRepo := &mockClinicLookupForStaffAssignments{
			lockActiveByIDFn: func(_ context.Context, _ uint64) (*model.Clinic, error) {
				t.Fatal("clinic lookup must not run after unauthorized existing-assignment rejection")
				return nil, nil
			},
		}
		svc := NewStaffService(
			staffRepo,
			&mockAccountForStaff{},
			assignmentRepo,
			&mockReservationForStaff{},
			&mockShiftEntryForStaff{},
			&mockPermissionGroupRepository{},
			&mockResStaffForStaff{},
			nil,
			clinicRepo,
			markedStaffSecurityTransactor{},
		)

		err := svc.SetClinicAssignments(context.Background(), &SetClinicAssignmentsInput{
			StaffID:             10,
			ClinicIDs:           []uint64{1},
			AuthorizedClinicIDs: []uint64{1},
			IsSystemAdmin:       false,
		})

		require.Error(t, err)
		assert.ErrorIs(t, err, apperrors.ErrForbidden)
		assert.Equal(t, []string{"lock-staff", "lock-assignments"}, events)
	})
}

func TestStaffService_SetClinicAssignments_AdminInactiveClinicDelta(t *testing.T) {
	inactiveNotFound := apperrors.WrapNotFound("clinic", "30")

	t.Run("admin PUT that includes inactive GET ids keeps them without LockActiveByID", func(t *testing.T) {
		restored := make([]uint64, 0, 2)
		locked := make([]uint64, 0, 1)
		staffRepo := &mockStaffRepository{
			lockForUpdateFn: func(_ context.Context, id uint64) (*model.Staff, error) {
				return &model.Staff{ID: id, ClinicID: 20}, nil
			},
			updatePrimaryFn: func(_ context.Context, _, clinicID uint64) error {
				assert.Equal(t, uint64(20), clinicID)
				return nil
			},
		}
		assignmentRepo := &mockAssignmentForStaff{
			lockActiveFn: func(_ context.Context, staffID uint64) ([]model.StaffClinicAssignment, error) {
				return []model.StaffClinicAssignment{
					{StaffID: staffID, ClinicID: 20, IsMain: true},
					{StaffID: staffID, ClinicID: 30},
				}, nil
			},
			deleteByStaffIDFn: func(_ context.Context, _ uint64) error {
				t.Fatal("including inactive assignment ids must not full-replace/delete them")
				return nil
			},
			restoreOrCreateFn: func(_ context.Context, assignment *model.StaffClinicAssignment) error {
				restored = append(restored, assignment.ClinicID)
				return nil
			},
		}
		clinicRepo := &mockClinicLookupForStaffAssignments{
			lockActiveByIDFn: func(_ context.Context, id uint64) (*model.Clinic, error) {
				locked = append(locked, id)
				if id == 30 {
					return nil, inactiveNotFound
				}
				return &model.Clinic{ID: id, IsActive: true}, nil
			},
		}
		svc := NewStaffService(
			staffRepo,
			&mockAccountForStaff{},
			assignmentRepo,
			&mockReservationForStaff{},
			&mockShiftEntryForStaff{},
			&mockPermissionGroupRepository{},
			&mockResStaffForStaff{},
			nil,
			clinicRepo,
			noopTransactor{},
		)

		err := svc.SetClinicAssignments(context.Background(), &SetClinicAssignmentsInput{
			StaffID:             10,
			ClinicIDs:           []uint64{20, 30},
			AuthorizedClinicIDs: []uint64{20},
			IsSystemAdmin:       true,
		})

		require.NoError(t, err)
		assert.NotContains(t, locked, uint64(30))
		assert.Contains(t, restored, uint64(20))
		assert.Contains(t, restored, uint64(30))
	})

	t.Run("admin PUT that omits an inactive id may remove it", func(t *testing.T) {
		assigned := map[uint64]struct{}{20: {}, 30: {}}
		restored := make([]uint64, 0, 1)
		locked := make([]uint64, 0, 1)
		staffRepo := &mockStaffRepository{
			lockForUpdateFn: func(_ context.Context, id uint64) (*model.Staff, error) {
				return &model.Staff{ID: id, ClinicID: 20}, nil
			},
			updatePrimaryFn: func(_ context.Context, _, clinicID uint64) error {
				assert.Equal(t, uint64(20), clinicID)
				return nil
			},
		}
		assignmentRepo := &mockAssignmentForStaff{
			lockActiveFn: func(_ context.Context, staffID uint64) ([]model.StaffClinicAssignment, error) {
				return []model.StaffClinicAssignment{
					{StaffID: staffID, ClinicID: 20, IsMain: true},
					{StaffID: staffID, ClinicID: 30},
				}, nil
			},
			deleteByStaffIDFn: func(_ context.Context, _ uint64) error {
				assigned = map[uint64]struct{}{}
				return nil
			},
			deleteByClinicIDsFn: func(_ context.Context, _ uint64, clinicIDs []uint64) error {
				for _, clinicID := range clinicIDs {
					delete(assigned, clinicID)
				}
				return nil
			},
			restoreOrCreateFn: func(_ context.Context, assignment *model.StaffClinicAssignment) error {
				assigned[assignment.ClinicID] = struct{}{}
				restored = append(restored, assignment.ClinicID)
				return nil
			},
		}
		clinicRepo := &mockClinicLookupForStaffAssignments{
			lockActiveByIDFn: func(_ context.Context, id uint64) (*model.Clinic, error) {
				locked = append(locked, id)
				if id == 30 {
					t.Fatal("omitted inactive clinic must not be LockActiveByID")
					return nil, inactiveNotFound
				}
				return &model.Clinic{ID: id, IsActive: true}, nil
			},
		}
		svc := NewStaffService(
			staffRepo,
			&mockAccountForStaff{},
			assignmentRepo,
			&mockReservationForStaff{},
			&mockShiftEntryForStaff{},
			&mockPermissionGroupRepository{},
			&mockResStaffForStaff{},
			nil,
			clinicRepo,
			noopTransactor{},
		)

		err := svc.SetClinicAssignments(context.Background(), &SetClinicAssignmentsInput{
			StaffID:             10,
			ClinicIDs:           []uint64{20},
			AuthorizedClinicIDs: []uint64{20},
			IsSystemAdmin:       true,
		})

		require.NoError(t, err)
		assert.NotContains(t, locked, uint64(30))
		assert.NotContains(t, restored, uint64(30))
		_, stillAssigned := assigned[30]
		assert.False(t, stillAssigned, "admin mutable scope may remove the omitted inactive assignment")
		_, kept := assigned[20]
		assert.True(t, kept)
	})
}

func TestStaffService_SetClinicAssignments_RejectsRemovingClinicWithExistingShiftBeforeMutation(t *testing.T) {
	events := make([]string, 0, 4)
	staffRepo := &mockStaffRepository{
		lockForUpdateFn: func(ctx context.Context, id uint64) (*model.Staff, error) {
			requireStaffSecurityTxContext(t, ctx)
			events = append(events, "lock-staff")
			return &model.Staff{ID: id}, nil
		},
	}
	assignmentRepo := &mockAssignmentForStaff{
		lockActiveFn: func(ctx context.Context, staffID uint64) ([]model.StaffClinicAssignment, error) {
			requireStaffSecurityTxContext(t, ctx)
			events = append(events, "lock-assignments")
			return []model.StaffClinicAssignment{
				{StaffID: staffID, ClinicID: 1, IsMain: true},
				{StaffID: staffID, ClinicID: 2},
			}, nil
		},
		deleteByStaffIDFn: func(_ context.Context, _ uint64) error {
			t.Fatal("shift dependency must reject before deleting assignments")
			return nil
		},
		restoreOrCreateFn: func(_ context.Context, _ *model.StaffClinicAssignment) error {
			t.Fatal("shift dependency must reject before restoring assignments")
			return nil
		},
	}
	clinicRepo := &mockClinicLookupForStaffAssignments{
		lockActiveByIDFn: func(ctx context.Context, id uint64) (*model.Clinic, error) {
			requireStaffSecurityTxContext(t, ctx)
			events = append(events, fmt.Sprintf("lock-clinic:%d", id))
			return &model.Clinic{ID: id, IsActive: true}, nil
		},
	}
	shiftRepo := &mockShiftEntryForStaff{
		existsByStaffIDFn: func(ctx context.Context, clinicID, staffID uint64) (bool, error) {
			requireStaffSecurityTxContext(t, ctx)
			assert.Equal(t, uint64(2), clinicID)
			assert.Equal(t, uint64(10), staffID)
			events = append(events, "check-shift:2")
			return true, nil
		},
	}
	svc := NewStaffService(
		staffRepo,
		&mockAccountForStaff{},
		assignmentRepo,
		&mockReservationForStaff{},
		shiftRepo,
		&mockPermissionGroupRepository{},
		&mockResStaffForStaff{},
		nil,
		clinicRepo,
		markedStaffSecurityTransactor{},
	)

	err := svc.SetClinicAssignments(context.Background(), &SetClinicAssignmentsInput{
		StaffID:             10,
		ClinicIDs:           []uint64{1},
		AuthorizedClinicIDs: []uint64{1, 2},
	})

	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err), "unexpected error: %v", err)
	assert.Equal(t, []string{
		"lock-staff",
		"lock-assignments",
		"check-shift:2",
	}, events)
}

func TestStaffService_SetClinicAssignments_PropagatesRemovedClinicShiftCheckError(t *testing.T) {
	dependencyErr := errors.New("shift dependency failed")
	assignmentRepo := &mockAssignmentForStaff{
		lockActiveFn: func(_ context.Context, staffID uint64) ([]model.StaffClinicAssignment, error) {
			return []model.StaffClinicAssignment{{StaffID: staffID, ClinicID: 1, IsMain: true}}, nil
		},
		deleteByStaffIDFn: func(_ context.Context, _ uint64) error {
			t.Fatal("dependency read failure must preserve assignments")
			return nil
		},
	}
	shiftRepo := &mockShiftEntryForStaff{
		existsByStaffIDFn: func(_ context.Context, clinicID, staffID uint64) (bool, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(10), staffID)
			return false, dependencyErr
		},
	}
	svc := NewStaffService(
		&mockStaffRepository{},
		&mockAccountForStaff{},
		assignmentRepo,
		&mockReservationForStaff{},
		shiftRepo,
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
	assert.ErrorIs(t, err, dependencyErr)
}

func TestStaffService_SetClinicAssignments_UsesCanonicalLockOrderAndTransactionContext(t *testing.T) {
	events := make([]string, 0, 8)
	input := &SetClinicAssignmentsInput{
		StaffID:             10,
		ClinicIDs:           []uint64{1, 2},
		AuthorizedClinicIDs: []uint64{1, 2},
	}
	originalClinicIDs := append([]uint64(nil), input.ClinicIDs...)
	originalAuthorizedClinicIDs := append([]uint64(nil), input.AuthorizedClinicIDs...)
	staffRepo := &mockStaffRepository{
		lockForUpdateFn: func(ctx context.Context, id uint64) (*model.Staff, error) {
			requireStaffSecurityTxContext(t, ctx)
			events = append(events, "lock-staff")
			return &model.Staff{ID: id}, nil
		},
		updatePrimaryFn: func(ctx context.Context, staffID, clinicID uint64) error {
			requireStaffSecurityTxContext(t, ctx)
			assert.Equal(t, uint64(10), staffID)
			assert.Equal(t, uint64(1), clinicID)
			events = append(events, "primary:1")
			return nil
		},
	}
	assignmentRepo := &mockAssignmentForStaff{
		lockActiveFn: func(ctx context.Context, staffID uint64) ([]model.StaffClinicAssignment, error) {
			requireStaffSecurityTxContext(t, ctx)
			events = append(events, "lock-assignments")
			return []model.StaffClinicAssignment{{StaffID: staffID, ClinicID: 1, IsMain: true}}, nil
		},
		deleteByStaffIDFn: func(ctx context.Context, staffID uint64) error {
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
	clinicRepo := &mockClinicLookupForStaffAssignments{
		lockActiveByIDFn: func(ctx context.Context, id uint64) (*model.Clinic, error) {
			requireStaffSecurityTxContext(t, ctx)
			events = append(events, fmt.Sprintf("lock-clinic:%d", id))
			return &model.Clinic{ID: id, IsActive: true}, nil
		},
	}
	svc := NewStaffService(
		staffRepo,
		&mockAccountForStaff{},
		assignmentRepo,
		&mockReservationForStaff{},
		&mockShiftEntryForStaff{},
		&mockPermissionGroupRepository{},
		&mockResStaffForStaff{},
		nil,
		clinicRepo,
		markedStaffSecurityTransactor{},
	)

	err := svc.SetClinicAssignments(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, []string{
		"lock-staff",
		"lock-assignments",
		"lock-clinic:2",
		"restore:1",
		"restore:2",
		"primary:1",
	}, events)
	assert.Equal(t, originalClinicIDs, input.ClinicIDs)
	assert.Equal(t, originalAuthorizedClinicIDs, input.AuthorizedClinicIDs)
}

// ---- VerifyClinicMembership ----

func TestStaffService_VerifyClinicMembership(t *testing.T) {
	tests := []struct {
		name         string
		count        int64
		countErr     error
		wantErr      bool
		wantNotFound bool
	}{
		{name: "member found", count: 1},
		{name: "not a member → not found", count: 0, wantErr: true, wantNotFound: true},
		{name: "propagates repository error", countErr: errors.New("db error"), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assignmentRepo := &mockClinicMembershipRepo{
				countFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return tt.count, tt.countErr
				},
			}
			svc := NewStaffService(&mockStaffRepository{}, &mockAccountForStaff{}, assignmentRepo, &mockReservationForStaff{}, &mockShiftEntryForStaff{}, &mockPermissionGroupRepository{}, &mockResStaffForStaff{}, nil, nil, noopTransactor{})

			err := svc.VerifyClinicMembership(context.Background(), 10, 1)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNotFound {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
