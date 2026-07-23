package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// mockClinicMembershipRepo is a StaffClinicAssignmentRepository mock with a fully configurable
// CountByStaffAndClinic. The mockAssignmentForStaff defined in staff_service_test.go hardcodes
// CountByStaffAndClinic to always return 1, which cannot exercise VerifyClinicMembership's
// not-found branch (count == 0), so this file defines its own variant instead of touching that one.
type mockClinicMembershipRepo struct {
	countFn  func(ctx context.Context, staffID, clinicID uint64) (int64, error)
	createFn func(ctx context.Context, a *model.StaffClinicAssignment) error
	deleteFn func(ctx context.Context, staffID uint64) error
}

func (m *mockClinicMembershipRepo) FindByStaffID(_ context.Context, _ uint64) ([]model.StaffClinicAssignment, error) {
	return nil, nil
}

func (m *mockClinicMembershipRepo) CountByStaffAndClinic(ctx context.Context, staffID, clinicID uint64) (int64, error) {
	return m.countFn(ctx, staffID, clinicID)
}

func (m *mockClinicMembershipRepo) LockActiveByStaffAndClinic(
	_ context.Context,
	staffID, clinicID uint64,
) (*model.StaffClinicAssignment, error) {
	return &model.StaffClinicAssignment{StaffID: staffID, ClinicID: clinicID}, nil
}

func (m *mockClinicMembershipRepo) LockActiveByStaff(
	_ context.Context,
	_ uint64,
) ([]model.StaffClinicAssignment, error) {
	return nil, nil
}

func (m *mockClinicMembershipRepo) Create(ctx context.Context, a *model.StaffClinicAssignment) error {
	if m.createFn != nil {
		return m.createFn(ctx, a)
	}
	return nil
}

func (m *mockClinicMembershipRepo) RestoreOrCreate(ctx context.Context, a *model.StaffClinicAssignment) error {
	return m.Create(ctx, a)
}

func (m *mockClinicMembershipRepo) Delete(ctx context.Context, staffID uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, staffID)
	}
	return nil
}

var _ repository.StaffClinicAssignmentRepository = (*mockClinicMembershipRepo)(nil)

type mockClinicLookupForStaffAssignments struct {
	lockActiveByIDFn func(ctx context.Context, id uint64) (*model.Clinic, error)
}

func (m *mockClinicLookupForStaffAssignments) LockActiveByID(ctx context.Context, id uint64) (*model.Clinic, error) {
	return m.lockActiveByIDFn(ctx, id)
}

func existingClinicLookupForStaffAssignments() *mockClinicLookupForStaffAssignments {
	return &mockClinicLookupForStaffAssignments{
		lockActiveByIDFn: func(_ context.Context, id uint64) (*model.Clinic, error) {
			return &model.Clinic{ID: id}, nil
		},
	}
}

type staffAssignmentTxState struct {
	assignments     []model.StaffClinicAssignment
	primaryClinicID uint64
}

type rollbackStaffAssignmentTransactor struct {
	state *staffAssignmentTxState
}

func (t *rollbackStaffAssignmentTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	assignmentsBefore := append([]model.StaffClinicAssignment(nil), t.state.assignments...)
	primaryClinicIDBefore := t.state.primaryClinicID
	if err := fn(ctx); err != nil {
		t.state.assignments = assignmentsBefore
		t.state.primaryClinicID = primaryClinicIDBefore
		return err
	}
	return nil
}

type staffSecurityTxMarkerKey struct{}

type markedStaffSecurityTransactor struct{}

func (markedStaffSecurityTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(context.WithValue(ctx, staffSecurityTxMarkerKey{}, true))
}

func requireStaffSecurityTxContext(t *testing.T, ctx context.Context) {
	t.Helper()
	assert.Equal(t, true, ctx.Value(staffSecurityTxMarkerKey{}), "repository call escaped transaction context")
}

// ---- FindByAccountID ----

func TestStaffService_FindByAccountID(t *testing.T) {
	tests := []struct {
		name              string
		findByAccountIDFn func(ctx context.Context, accountID uint64) (*model.Staff, error)
		wantErr           bool
		wantNotFound      bool
	}{
		{
			name: "returns staff when found",
			findByAccountIDFn: func(_ context.Context, accountID uint64) (*model.Staff, error) {
				return &model.Staff{ID: 1, AccountID: &accountID}, nil
			},
		},
		{
			name: "propagates not found error",
			findByAccountIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
				return nil, apperrors.WrapNotFound("staff", "account_id=999")
			},
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name: "propagates repository error",
			findByAccountIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockStaffRepository{findByAccountIDFn: tt.findByAccountIDFn}
			svc := newTestStaffService(repo)

			staff, err := svc.FindByAccountID(context.Background(), 1)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, staff)
				if tt.wantNotFound {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, staff)
			}
		})
	}
}

// ---- CreateWithAccount ----

func TestStaffService_CreateWithAccount_ValidationErrors(t *testing.T) {
	tests := []struct {
		name  string
		input *CreateStaffWithAccountInput
	}{
		{name: "missing name", input: &CreateStaffWithAccountInput{ClinicID: 1, Name: "   "}},
		{name: "missing clinic_id", input: &CreateStaffWithAccountInput{Name: "スタッフ"}},
		{name: "email without password", input: &CreateStaffWithAccountInput{ClinicID: 1, Name: "スタッフ", Email: "a@example.com"}},
		{name: "weak password", input: &CreateStaffWithAccountInput{ClinicID: 1, Name: "スタッフ", Email: "a@example.com", Password: "short"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockStaffRepository{
				createFn: func(_ context.Context, _ *model.Staff) error {
					t.Fatal("repository must not be called when validation fails")
					return nil
				},
			}
			svc := newTestStaffService(repo)

			staff, err := svc.CreateWithAccount(context.Background(), tt.input)

			assert.Error(t, err)
			assert.Nil(t, staff)
			assert.True(t, apperrors.IsInvalidInput(err))
		})
	}
}

func TestStaffService_CreateWithAccount_EmailUniquenessCheckError(t *testing.T) {
	accountRepo := &mockAccountForStaff{
		findByEmailFn: func(_ context.Context, _ string) (*model.Account, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewStaffService(&mockStaffRepository{}, accountRepo, &mockAssignmentForStaff{}, &mockReservationForStaff{}, &mockShiftEntryForStaff{}, &mockPermissionGroupRepository{}, &mockResStaffForStaff{}, nil, nil, noopTransactor{})

	staff, err := svc.CreateWithAccount(context.Background(), &CreateStaffWithAccountInput{
		ClinicID: 1, Name: "スタッフ", Email: "a@example.com", Password: "Passw0rd1",
	})

	assert.Error(t, err)
	assert.Nil(t, staff)
}

func TestStaffService_CreateWithAccount_EmailAlreadyExists(t *testing.T) {
	accountRepo := &mockAccountForStaff{
		findByEmailFn: func(_ context.Context, email string) (*model.Account, error) {
			return &model.Account{ID: 99, Email: email}, nil
		},
	}
	svc := NewStaffService(&mockStaffRepository{}, accountRepo, &mockAssignmentForStaff{}, &mockReservationForStaff{}, &mockShiftEntryForStaff{}, &mockPermissionGroupRepository{}, &mockResStaffForStaff{}, nil, nil, noopTransactor{})

	staff, err := svc.CreateWithAccount(context.Background(), &CreateStaffWithAccountInput{
		ClinicID: 1, Name: "スタッフ", Email: "dup@example.com", Password: "Passw0rd1",
	})

	assert.Error(t, err)
	assert.Nil(t, staff)
	assert.True(t, apperrors.IsAlreadyExists(err))
}

func TestStaffService_CreateWithAccount_PasswordTooLongForBcrypt(t *testing.T) {
	// bcrypt rejects passwords over 72 bytes; this is >90 bytes and still satisfies
	// validatePassword (contains letters and digits, length >= 8).
	longPassword := strings.Repeat("Aa1", 30)
	svc := newTestStaffService(&mockStaffRepository{})

	staff, err := svc.CreateWithAccount(context.Background(), &CreateStaffWithAccountInput{
		ClinicID: 1, Name: "スタッフ", Email: "a@example.com", Password: longPassword,
	})

	assert.Error(t, err)
	assert.Nil(t, staff)
}

func TestStaffService_CreateWithAccount_Success(t *testing.T) {
	var createdAccount *model.Account
	var createdStaff *model.Staff
	var createdAssignment *model.StaffClinicAssignment

	accountRepo := &mockAccountForStaff{
		createFn: func(_ context.Context, a *model.Account) error {
			a.ID = 42
			createdAccount = a
			return nil
		},
	}
	staffRepo := &mockStaffRepository{
		createFn: func(_ context.Context, s *model.Staff) error {
			s.ID = 7
			createdStaff = s
			return nil
		},
	}
	assignmentRepo := &mockAssignmentForStaff{
		createFn: func(_ context.Context, a *model.StaffClinicAssignment) error {
			createdAssignment = a
			return nil
		},
	}
	svc := NewStaffService(staffRepo, accountRepo, assignmentRepo, &mockReservationForStaff{}, &mockShiftEntryForStaff{}, &mockPermissionGroupRepository{}, &mockResStaffForStaff{}, nil, nil, noopTransactor{})

	staff, err := svc.CreateWithAccount(context.Background(), &CreateStaffWithAccountInput{
		ClinicID: 1,
		Name:     "  新規 スタッフ  ",
		Email:    "new@example.com",
		Password: "Passw0rd1",
	})

	assert.NoError(t, err)
	if assert.NotNil(t, staff) {
		assert.Equal(t, "新規 スタッフ", staff.Name)
		assert.Equal(t, model.StaffTypeDoctor, staff.StaffType)
		assert.True(t, staff.ReservationVisible)
		assert.NotNil(t, staff.AccountID)
	}
	if assert.NotNil(t, createdAccount) {
		assert.True(t, createdAccount.IsActive)
		assert.NotEmpty(t, createdAccount.PasswordHash)
	}
	assert.NotNil(t, createdStaff)
	if assert.NotNil(t, createdAssignment) {
		assert.True(t, createdAssignment.IsMain)
		assert.Equal(t, uint64(1), createdAssignment.ClinicID)
	}
}

func TestStaffService_CreateWithAccount_CustomStaffTypeAndReservationVisible(t *testing.T) {
	visible := false
	staffRepo := &mockStaffRepository{
		createFn: func(_ context.Context, s *model.Staff) error {
			s.ID = 1
			return nil
		},
	}
	svc := newTestStaffService(staffRepo)

	staff, err := svc.CreateWithAccount(context.Background(), &CreateStaffWithAccountInput{
		ClinicID:           1,
		Name:               "スタッフ",
		StaffType:          string(model.StaffTypeTrimmer),
		ReservationVisible: &visible,
	})

	assert.NoError(t, err)
	if assert.NotNil(t, staff) {
		assert.Equal(t, model.StaffTypeTrimmer, staff.StaffType)
		assert.False(t, staff.ReservationVisible)
	}
}

func TestStaffService_CreateWithAccount_TxFailures(t *testing.T) {
	baseInput := &CreateStaffWithAccountInput{ClinicID: 1, Name: "スタッフ", Email: "a@example.com", Password: "Passw0rd1"}

	t.Run("account create fails", func(t *testing.T) {
		accountRepo := &mockAccountForStaff{
			createFn: func(_ context.Context, _ *model.Account) error { return errors.New("db error") },
		}
		svc := NewStaffService(&mockStaffRepository{}, accountRepo, &mockAssignmentForStaff{}, &mockReservationForStaff{}, &mockShiftEntryForStaff{}, &mockPermissionGroupRepository{}, &mockResStaffForStaff{}, nil, nil, noopTransactor{})

		staff, err := svc.CreateWithAccount(context.Background(), baseInput)

		assert.Error(t, err)
		assert.Nil(t, staff)
	})

	t.Run("staff create fails", func(t *testing.T) {
		staffRepo := &mockStaffRepository{
			createFn: func(_ context.Context, _ *model.Staff) error { return errors.New("db error") },
		}
		svc := NewStaffService(staffRepo, &mockAccountForStaff{}, &mockAssignmentForStaff{}, &mockReservationForStaff{}, &mockShiftEntryForStaff{}, &mockPermissionGroupRepository{}, &mockResStaffForStaff{}, nil, nil, noopTransactor{})

		staff, err := svc.CreateWithAccount(context.Background(), baseInput)

		assert.Error(t, err)
		assert.Nil(t, staff)
	})

	t.Run("clinic assignment create fails", func(t *testing.T) {
		staffRepo := &mockStaffRepository{
			createFn: func(_ context.Context, s *model.Staff) error {
				s.ID = 1
				return nil
			},
		}
		assignmentRepo := &mockAssignmentForStaff{
			createFn: func(_ context.Context, _ *model.StaffClinicAssignment) error { return errors.New("db error") },
		}
		svc := NewStaffService(staffRepo, &mockAccountForStaff{}, assignmentRepo, &mockReservationForStaff{}, &mockShiftEntryForStaff{}, &mockPermissionGroupRepository{}, &mockResStaffForStaff{}, nil, nil, noopTransactor{})

		staff, err := svc.CreateWithAccount(context.Background(), baseInput)

		assert.Error(t, err)
		assert.Nil(t, staff)
	})
}

// ---- UpdatePassword ----

func TestStaffService_UpdatePassword(t *testing.T) {
	tests := []struct {
		name           string
		newPassword    string
		updateFieldsFn func(ctx context.Context, id uint64, fields map[string]any) error
		wantErr        bool
	}{
		{
			name:        "updates password successfully",
			newPassword: "Passw0rd1",
		},
		{
			name:        "propagates repository error",
			newPassword: "Passw0rd1",
			updateFieldsFn: func(_ context.Context, _ uint64, _ map[string]any) error {
				return errors.New("db error")
			},
			wantErr: true,
		},
		{
			name:        "returns error when password exceeds bcrypt max length",
			newPassword: strings.Repeat("Aa1", 30),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedFields map[string]any
			accountRepo := &mockAccountForStaff{
				updateFieldsFn: func(ctx context.Context, id uint64, fields map[string]any) error {
					capturedFields = fields
					if tt.updateFieldsFn != nil {
						return tt.updateFieldsFn(ctx, id, fields)
					}
					return nil
				},
			}
			svc := NewStaffService(&mockStaffRepository{}, accountRepo, &mockAssignmentForStaff{}, &mockReservationForStaff{}, &mockShiftEntryForStaff{}, &mockPermissionGroupRepository{}, &mockResStaffForStaff{}, nil, nil, noopTransactor{})

			err := svc.UpdatePassword(context.Background(), 1, tt.newPassword)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, capturedFields, "password_hash")
			}
		})
	}
}

// ---- SetClinicAssignments (additional error branches) ----

func TestStaffService_SetClinicAssignments_ErrorBranches(t *testing.T) {
	t.Run("delete existing assignments fails", func(t *testing.T) {
		assignmentRepo := &mockAssignmentForStaff{
			deleteByStaffIDFn: func(_ context.Context, _ uint64) error { return errors.New("db error") },
		}
		svc := NewStaffService(&mockStaffRepository{}, &mockAccountForStaff{}, assignmentRepo, &mockReservationForStaff{}, &mockShiftEntryForStaff{}, &mockPermissionGroupRepository{}, &mockResStaffForStaff{}, nil, existingClinicLookupForStaffAssignments(), noopTransactor{})

		err := svc.SetClinicAssignments(context.Background(), &SetClinicAssignmentsInput{
			StaffID:             10,
			ClinicIDs:           []uint64{1},
			AuthorizedClinicIDs: []uint64{1},
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
		assert.Equal(t, []string{"find:2", "find:4", "delete", "create:2", "create:4", "primary:2"}, events)
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

func TestStaffService_SetClinicAssignments_RejectsExistingUnauthorizedAssignmentBeforeMutation(t *testing.T) {
	events := make([]string, 0, 2)
	staffRepo := &mockStaffRepository{
		lockForUpdateFn: func(ctx context.Context, id uint64) (*model.Staff, error) {
			requireStaffSecurityTxContext(t, ctx)
			assert.Equal(t, uint64(10), id)
			events = append(events, "lock-staff")
			return &model.Staff{ID: id}, nil
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
			t.Fatal("authorization failure must preserve every existing assignment")
			return nil
		},
		restoreOrCreateFn: func(_ context.Context, _ *model.StaffClinicAssignment) error {
			t.Fatal("authorization failure must not restore or create assignments")
			return nil
		},
	}
	clinicRepo := &mockClinicLookupForStaffAssignments{
		lockActiveByIDFn: func(_ context.Context, _ uint64) (*model.Clinic, error) {
			t.Fatal("existing-scope authorization must fail before target clinic lookup")
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
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
	assert.Equal(t, []string{"lock-staff", "lock-assignments"}, events)
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
		"lock-clinic:1",
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
		"lock-clinic:1",
		"lock-clinic:2",
		"delete",
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
