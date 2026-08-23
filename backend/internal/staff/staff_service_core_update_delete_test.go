package staff

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestStaffServiceCore_Update(t *testing.T) {
	name := "更新太郎"
	emptyName := ""
	password := "newpassword1"
	weakPassword := "short"

	tests := []struct {
		name           string
		input          *UpdateStaffInput
		findByIDFn     func(ctx context.Context, id uint64) (*model.Staff, error)
		updateFn       func(ctx context.Context, clinicID, id uint64, fields map[string]any) error
		accountUpdate  func(ctx context.Context, id uint64, newHash string, updatedAt time.Time) error
		tokenDelete    func(ctx context.Context, id uint64) error
		wantDelete     bool
		wantErr        bool
		wantErrInvalid bool
	}{
		{
			name:    "updates profile fields successfully",
			input:   &UpdateStaffInput{Name: &name},
			wantErr: false,
		},
		{
			name:           "returns invalid input when no fields are set",
			input:          &UpdateStaffInput{},
			wantErr:        true,
			wantErrInvalid: true,
		},
		{
			name:  "returns error when staff not found",
			input: &UpdateStaffInput{Name: &name},
			findByIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
				return nil, apperrors.WrapNotFound("staff", "1")
			},
			wantErr: true,
		},
		{
			name:    "returns error when name is empty",
			input:   &UpdateStaffInput{Name: &emptyName},
			wantErr: true,
		},
		{
			name:  "returns wrapped error when repo.Update fails",
			input: &UpdateStaffInput{Name: &name},
			updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
				return errors.New("db error")
			},
			wantErr: true,
		},
		{
			name:       "updates password when account exists",
			input:      &UpdateStaffInput{Password: &password},
			wantDelete: true,
			wantErr:    false,
		},
		{
			name:    "returns error when password is too weak",
			input:   &UpdateStaffInput{Password: &weakPassword},
			wantErr: true,
		},
		{
			name:  "returns wrapped error when account password update fails",
			input: &UpdateStaffInput{Password: &password},
			accountUpdate: func(_ context.Context, _ uint64, _ string, _ time.Time) error {
				return errors.New("account db error")
			},
			wantErr: true,
		},
		{
			name:  "returns wrapped error when reset token invalidation fails",
			input: &UpdateStaffInput{Password: &password},
			tokenDelete: func(context.Context, uint64) error {
				return errors.New("token delete failed")
			},
			wantDelete: true,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountID := uint64(5)
			findByIDFn := tt.findByIDFn
			if findByIDFn == nil {
				findByIDFn = func(_ context.Context, id uint64) (*model.Staff, error) {
					return &model.Staff{ID: id, ClinicID: 1, AccountID: &accountID}, nil
				}
			}
			repo := &coreMockStaffRepository{
				findByIDFn: findByIDFn,
				updateFn:   tt.updateFn,
			}
			deleteCalled := false
			accountRepo := &coreMockAccountRepository{
				updatePasswordHashFn: tt.accountUpdate,
				deletePasswordResetTokensFn: func(
					ctx context.Context,
					id uint64,
				) error {
					deleteCalled = true
					if tt.tokenDelete != nil {
						return tt.tokenDelete(ctx, id)
					}
					return nil
				},
			}
			tx := &coreFakeTransactor{}
			svc := newCoreStaffService(repo, accountRepo, &coreMockStaffClinicAssignmentRepository{}, &coreMockReservationQueryRepository{}, &coreMockShiftEntryRepository{}, tx)

			input := authorizedStaffUpdate(tt.input, 1)
			if input != nil && input.Password != nil && *input.Password != "" {
				input.CredentialAudit = testStaffCredentialAudit(1, 1)
			}
			staff, err := svc.Update(context.Background(), 1, 1, input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, staff)
				if tt.wantErrInvalid {
					assert.True(t, apperrors.IsInvalidInput(err))
				}
				assert.Equal(t, tt.wantDelete, deleteCalled)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, staff)
			assert.Equal(t, tt.wantDelete, deleteCalled)
		})
	}
}

func TestStaffServiceCore_Update_FindByIDErrorAfterProfileUpdate(t *testing.T) {
	// covers the clinic-scoped reload inside the update transaction failing.
	name := "更新太郎"
	repo := &coreMockStaffRepository{
		findByIDFn: func(_ context.Context, id uint64) (*model.Staff, error) {
			return &model.Staff{ID: id, ClinicID: 1}, nil
		},
		findByIDInClinicFn: func(_ context.Context, _, _ uint64) (*model.Staff, error) {
			return nil, errors.New("reload failed")
		},
	}
	tx := &coreFakeTransactor{}
	svc := newCoreStaffService(repo, &coreMockAccountRepository{}, &coreMockStaffClinicAssignmentRepository{}, &coreMockReservationQueryRepository{}, &coreMockShiftEntryRepository{}, tx)

	staff, err := svc.Update(
		context.Background(),
		1,
		1,
		authorizedStaffUpdate(&UpdateStaffInput{Name: &name}, 1),
	)
	assert.Error(t, err)
	assert.Nil(t, staff)
}

func TestStaffServiceCore_Update_RequiresScopedLockBeforePasswordMutation(t *testing.T) {
	accountID := uint64(5)
	password := "newpassword1"
	accountUpdated := false
	repo := &coreMockStaffRepository{
		findByIDFn: func(_ context.Context, id uint64) (*model.Staff, error) {
			return &model.Staff{ID: id, ClinicID: 2, AccountID: &accountID}, nil
		},
		lockInClinicFn: func(_ context.Context, _, _ uint64) (*model.Staff, error) {
			return nil, apperrors.WrapNotFound("staff", "1")
		},
	}
	accountRepo := &coreMockAccountRepository{
		updatePasswordHashFn: func(_ context.Context, _ uint64, _ string, _ time.Time) error {
			accountUpdated = true
			return nil
		},
	}
	tx := &coreFakeTransactor{}
	svc := newCoreStaffService(
		repo,
		accountRepo,
		&coreMockStaffClinicAssignmentRepository{},
		&coreMockReservationQueryRepository{},
		&coreMockShiftEntryRepository{},
		tx,
	)

	updated, err := svc.Update(
		context.Background(),
		1,
		1,
		authorizedStaffUpdate(&UpdateStaffInput{
			Password:        &password,
			CredentialAudit: testStaffCredentialAudit(1, 1),
		}, 1),
	)

	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Nil(t, updated)
	assert.False(t, accountUpdated)
	assert.Equal(t, 1, tx.calls)
}

func TestStaffServiceCore_Update_RejectsPasswordForStaffWithoutAccount(t *testing.T) {
	password := "newpassword1"
	repo := &coreMockStaffRepository{
		findByIDFn: func(_ context.Context, id uint64) (*model.Staff, error) {
			return &model.Staff{ID: id, ClinicID: 1}, nil
		},
	}
	svc := newCoreStaffService(
		repo,
		&coreMockAccountRepository{},
		&coreMockStaffClinicAssignmentRepository{},
		&coreMockReservationQueryRepository{},
		&coreMockShiftEntryRepository{},
		&coreFakeTransactor{},
	)

	updated, err := svc.Update(
		context.Background(),
		1,
		1,
		authorizedStaffUpdate(&UpdateStaffInput{
			Password:        &password,
			CredentialAudit: testStaffCredentialAudit(1, 1),
		}, 1),
	)

	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, updated)
}

func TestStaffServiceCore_Update_ValidatesPasswordBeforeProfileWrite(t *testing.T) {
	name := "更新太郎"
	weakPassword := "password"
	profileUpdated := false
	accountID := uint64(5)
	repo := &coreMockStaffRepository{
		findByIDFn: func(_ context.Context, id uint64) (*model.Staff, error) {
			return &model.Staff{ID: id, ClinicID: 1, AccountID: &accountID}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
			profileUpdated = true
			return nil
		},
	}
	svc := newCoreStaffService(
		repo,
		&coreMockAccountRepository{},
		&coreMockStaffClinicAssignmentRepository{},
		&coreMockReservationQueryRepository{},
		&coreMockShiftEntryRepository{},
		&coreFakeTransactor{},
	)

	updated, err := svc.Update(
		context.Background(),
		1,
		1,
		authorizedStaffUpdate(&UpdateStaffInput{
			Name:            &name,
			Password:        &weakPassword,
			CredentialAudit: testStaffCredentialAudit(1, 1),
		}, 1),
	)

	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, updated)
	assert.False(t, profileUpdated)
}

func TestStaffServiceCore_Update_ProfileAndPasswordRunInsideOneTransaction(t *testing.T) {
	name := "更新太郎"
	password := "newpassword1"
	accountID := uint64(5)
	repo := &coreMockStaffRepository{
		findByIDFn: func(_ context.Context, id uint64) (*model.Staff, error) {
			return &model.Staff{ID: id, ClinicID: 1, AccountID: &accountID}, nil
		},
	}
	tx := &coreFakeTransactor{}
	svc := newCoreStaffService(
		repo,
		&coreMockAccountRepository{},
		&coreMockStaffClinicAssignmentRepository{},
		&coreMockReservationQueryRepository{},
		&coreMockShiftEntryRepository{},
		tx,
	)

	updated, err := svc.Update(
		context.Background(),
		1,
		1,
		authorizedStaffUpdate(&UpdateStaffInput{
			Name:            &name,
			Password:        &password,
			CredentialAudit: testStaffCredentialAudit(1, 1),
		}, 1),
	)

	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, 1, tx.calls)
}

func TestStaffServiceCore_Update_UsesStaffAssignmentOccupationLockOrder(t *testing.T) {
	occupationID := uint64(30)
	events := make([]string, 0, 3)
	repo := &coreMockStaffRepository{
		lockInClinicFn: func(_ context.Context, clinicID, id uint64) (*model.Staff, error) {
			events = append(events, "staff")
			return &model.Staff{ID: id, ClinicID: clinicID}, nil
		},
		findByIDInClinicFn: func(_ context.Context, clinicID, id uint64) (*model.Staff, error) {
			return &model.Staff{ID: id, ClinicID: clinicID, OccupationID: &occupationID}, nil
		},
	}
	assignments := &coreMockStaffClinicAssignmentRepository{
		lockActiveFn: func(_ context.Context, staffID uint64) ([]model.StaffClinicAssignment, error) {
			events = append(events, "assignment")
			return []model.StaffClinicAssignment{{StaffID: staffID, ClinicID: 10}}, nil
		},
	}
	occupations := &mockOccupationRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Occupation, error) {
			t.Fatal("update must not use an unlocked occupation ownership read")
			return nil, nil
		},
		lockForShareFn: func(_ context.Context, clinicID, id uint64) (*model.Occupation, error) {
			events = append(events, "occupation")
			return &model.Occupation{ID: id, ClinicID: clinicID}, nil
		},
	}
	svc := NewStaffService(
		repo,
		&coreMockAccountRepository{},
		assignments,
		&coreMockReservationQueryRepository{},
		&coreMockShiftEntryRepository{},
		nil,
		nil,
		occupations,
		nil,
		&coreFakeTransactor{},
	)

	updated, err := svc.Update(
		context.Background(),
		10,
		7,
		authorizedStaffUpdate(&UpdateStaffInput{OccupationID: &occupationID}, 10),
	)

	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, []string{"staff", "assignment", "occupation"}, events)
}

// ---- Delete ----

func TestStaffServiceCore_Delete(t *testing.T) {
	tests := []struct {
		name                    string
		findByIDFn              func(ctx context.Context, id uint64) (*model.Staff, error)
		existsByStaffIDFn       func(ctx context.Context, clinicID, staffID uint64) (bool, error)
		shiftExistsByStaffIDFn  func(ctx context.Context, clinicID, staffID uint64) (bool, error)
		countBlockingReferences func(ctx context.Context, clinicID, staffID uint64) ([]StaffDependencyCount, error)
		deleteFn                func(ctx context.Context, clinicID, id uint64) error
		wantErr                 bool
		wantConflict            bool
	}{
		{
			name:    "deletes staff with no dependencies",
			wantErr: false,
		},
		{
			name: "returns error when staff not found",
			findByIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
				return nil, apperrors.WrapNotFound("staff", "1")
			},
			wantErr: true,
		},
		{
			name: "returns conflict when staff has reservations",
			existsByStaffIDFn: func(_ context.Context, _, _ uint64) (bool, error) {
				return true, nil
			},
			wantErr:      true,
			wantConflict: true,
		},
		{
			name: "returns wrapped error when reservation dependency check fails",
			existsByStaffIDFn: func(_ context.Context, _, _ uint64) (bool, error) {
				return false, errors.New("db error")
			},
			wantErr: true,
		},
		{
			name: "returns conflict when staff has shift entries",
			shiftExistsByStaffIDFn: func(_ context.Context, _, _ uint64) (bool, error) {
				return true, nil
			},
			wantErr:      true,
			wantConflict: true,
		},
		{
			name: "returns wrapped error when shift dependency check fails",
			shiftExistsByStaffIDFn: func(_ context.Context, _, _ uint64) (bool, error) {
				return false, errors.New("db error")
			},
			wantErr: true,
		},
		{
			name: "returns wrapped error when blocking-reference check fails",
			countBlockingReferences: func(_ context.Context, _, _ uint64) ([]StaffDependencyCount, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
		{
			name: "returns conflict when staff has blocking references",
			countBlockingReferences: func(_ context.Context, _, _ uint64) ([]StaffDependencyCount, error) {
				return []StaffDependencyCount{{Label: "カルテ", Count: 3}}, nil
			},
			wantErr:      true,
			wantConflict: true,
		},
		{
			name: "returns wrapped error when repo.Delete fails",
			deleteFn: func(_ context.Context, _, _ uint64) error {
				return errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &coreMockStaffRepository{
				findByIDFn:                         tt.findByIDFn,
				countBlockingReferencesByStaffIDFn: tt.countBlockingReferences,
				deleteFn:                           tt.deleteFn,
			}
			reservationRepo := &coreMockReservationQueryRepository{existsByStaffIDFn: tt.existsByStaffIDFn}
			shiftRepo := &coreMockShiftEntryRepository{existsByStaffIDFn: tt.shiftExistsByStaffIDFn}
			svc := newCoreStaffService(repo, &coreMockAccountRepository{}, &coreMockStaffClinicAssignmentRepository{}, reservationRepo, shiftRepo, &coreFakeTransactor{})

			err := svc.Delete(context.Background(), 1, 1, false)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantConflict {
					assert.True(t, apperrors.IsConflict(err))
				}
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestStaffService_Delete_UsesCanonicalLockOrderAndTransactionContext(t *testing.T) {
	events := make([]string, 0, 7)
	repo := &coreMockStaffRepository{
		lockInClinicFn: func(ctx context.Context, clinicID, id uint64) (*model.Staff, error) {
			requireStaffSecurityTxContext(t, ctx)
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(7), id)
			events = append(events, "lock-staff")
			return &model.Staff{ID: id}, nil
		},
		countBlockingReferencesByStaffIDFn: func(ctx context.Context, clinicID, staffID uint64) ([]StaffDependencyCount, error) {
			requireStaffSecurityTxContext(t, ctx)
			events = append(events, "dependencies")
			return nil, nil
		},
		deleteFn: func(ctx context.Context, clinicID, id uint64) error {
			requireStaffSecurityTxContext(t, ctx)
			events = append(events, "delete")
			return nil
		},
	}
	assignmentRepo := &coreMockStaffClinicAssignmentRepository{
		lockActiveFn: func(ctx context.Context, staffID uint64) ([]model.StaffClinicAssignment, error) {
			requireStaffSecurityTxContext(t, ctx)
			events = append(events, "lock-assignments")
			return []model.StaffClinicAssignment{{StaffID: staffID, ClinicID: 1, IsMain: true}}, nil
		},
	}
	reservationRepo := &coreMockReservationQueryRepository{
		existsByStaffIDFn: func(ctx context.Context, clinicID, staffID uint64) (bool, error) {
			requireStaffSecurityTxContext(t, ctx)
			events = append(events, "reservations")
			return false, nil
		},
	}
	shiftRepo := &coreMockShiftEntryRepository{
		existsByStaffIDFn: func(ctx context.Context, clinicID, staffID uint64) (bool, error) {
			requireStaffSecurityTxContext(t, ctx)
			events = append(events, "shifts")
			return false, nil
		},
	}
	svc := newCoreStaffService(
		repo,
		&coreMockAccountRepository{},
		assignmentRepo,
		reservationRepo,
		shiftRepo,
		markedStaffSecurityTransactor{},
	)

	err := svc.Delete(context.Background(), 1, 7, false)

	require.NoError(t, err)
	assert.Equal(t, []string{
		"lock-staff",
		"lock-assignments",
		"reservations",
		"shifts",
		"dependencies",
		"delete",
	}, events)
}

func TestStaffService_Delete_RejectsInvalidOrMultiClinicAssignmentStateBeforeDependencies(t *testing.T) {
	tests := []struct {
		name         string
		assignments  []model.StaffClinicAssignment
		wantNotFound bool
		wantConflict bool
	}{
		{
			name:         "no active assignments",
			assignments:  nil,
			wantNotFound: true,
		},
		{
			name: "only another clinic assignment",
			assignments: []model.StaffClinicAssignment{
				{StaffID: 7, ClinicID: 2},
			},
			wantNotFound: true,
		},
		{
			name: "multiple active assignments",
			assignments: []model.StaffClinicAssignment{
				{StaffID: 7, ClinicID: 1, IsMain: true},
				{StaffID: 7, ClinicID: 2},
			},
			wantConflict: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &coreMockStaffRepository{
				lockInClinicFn: func(_ context.Context, _, id uint64) (*model.Staff, error) {
					return &model.Staff{ID: id}, nil
				},
				countBlockingReferencesByStaffIDFn: func(_ context.Context, _, _ uint64) ([]StaffDependencyCount, error) {
					t.Fatal("dependency state must not be disclosed before assignment validation")
					return nil, nil
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					t.Fatal("invalid assignment state must not delete staff")
					return nil
				},
			}
			assignmentRepo := &coreMockStaffClinicAssignmentRepository{
				lockActiveFn: func(_ context.Context, _ uint64) ([]model.StaffClinicAssignment, error) {
					return tt.assignments, nil
				},
			}
			reservationRepo := &coreMockReservationQueryRepository{
				existsByStaffIDFn: func(_ context.Context, _, _ uint64) (bool, error) {
					t.Fatal("dependency state must not be disclosed before assignment validation")
					return false, nil
				},
			}
			shiftRepo := &coreMockShiftEntryRepository{
				existsByStaffIDFn: func(_ context.Context, _, _ uint64) (bool, error) {
					t.Fatal("dependency state must not be disclosed before assignment validation")
					return false, nil
				},
			}
			svc := newCoreStaffService(
				repo,
				&coreMockAccountRepository{},
				assignmentRepo,
				reservationRepo,
				shiftRepo,
				&coreFakeTransactor{},
			)

			err := svc.Delete(context.Background(), 1, 7, false)

			require.Error(t, err)
			if tt.wantNotFound {
				assert.True(t, apperrors.IsNotFound(err), "unexpected error: %v", err)
			}
			if tt.wantConflict {
				assert.True(t, apperrors.IsConflict(err), "unexpected error: %v", err)
			}
		})
	}
}

// ---- List / GetByID / Reorder (already high coverage elsewhere — smoke tests only) ----

func TestStaffServiceCore_List(t *testing.T) {
	repo := &coreMockStaffRepository{
		findAllFn: func(_ context.Context, _ uint64, _, _ int) ([]model.Staff, int64, error) {
			return []model.Staff{{ID: 1}}, 1, nil
		},
	}
	svc := newCoreStaffService(repo, &coreMockAccountRepository{}, &coreMockStaffClinicAssignmentRepository{}, &coreMockReservationQueryRepository{}, &coreMockShiftEntryRepository{}, &coreFakeTransactor{})
	staffs, total, err := svc.List(context.Background(), 1, 1, 20)
	assert.NoError(t, err)
	assert.Len(t, staffs, 1)
	assert.Equal(t, int64(1), total)

	repo.findAllFn = func(_ context.Context, _ uint64, _, _ int) ([]model.Staff, int64, error) {
		return nil, 0, errors.New("db error")
	}
	_, _, err = svc.List(context.Background(), 1, 1, 20)
	assert.Error(t, err)
}

func TestStaffServiceCore_GetByID(t *testing.T) {
	repo := &coreMockStaffRepository{
		findByIDFn: func(_ context.Context, id uint64) (*model.Staff, error) {
			return &model.Staff{ID: id}, nil
		},
	}
	svc := newCoreStaffService(repo, &coreMockAccountRepository{}, &coreMockStaffClinicAssignmentRepository{}, &coreMockReservationQueryRepository{}, &coreMockShiftEntryRepository{}, &coreFakeTransactor{})
	staff, err := svc.GetByID(context.Background(), 1)
	assert.NoError(t, err)
	assert.NotNil(t, staff)

	repo.findByIDFn = func(_ context.Context, _ uint64) (*model.Staff, error) {
		return nil, apperrors.WrapNotFound("staff", "1")
	}
	_, err = svc.GetByID(context.Background(), 1)
	assert.Error(t, err)
}

func TestStaffServiceCore_Reorder(t *testing.T) {
	repo := &coreMockStaffRepository{
		reorderFn: func(_ context.Context, _ uint64, _ []uint64) error { return nil },
	}
	svc := newCoreStaffService(repo, &coreMockAccountRepository{}, &coreMockStaffClinicAssignmentRepository{}, &coreMockReservationQueryRepository{}, &coreMockShiftEntryRepository{}, &coreFakeTransactor{})

	assert.NoError(t, svc.Reorder(context.Background(), 1, []uint64{2, 1}))
	assert.Error(t, svc.Reorder(context.Background(), 1, []uint64{}))

	repo.reorderFn = func(_ context.Context, _ uint64, _ []uint64) error { return errors.New("db error") }
	assert.Error(t, svc.Reorder(context.Background(), 1, []uint64{1}))
}
