package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

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

func (m *mockClinicMembershipRepo) Create(ctx context.Context, a *model.StaffClinicAssignment) error {
	if m.createFn != nil {
		return m.createFn(ctx, a)
	}
	return nil
}

func (m *mockClinicMembershipRepo) Delete(ctx context.Context, staffID uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, staffID)
	}
	return nil
}

var _ repository.StaffClinicAssignmentRepository = (*mockClinicMembershipRepo)(nil)

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
	svc := NewStaffService(&mockStaffRepository{}, accountRepo, &mockAssignmentForStaff{}, &mockReservationForStaff{}, &mockShiftEntryForStaff{}, &mockPermissionGroupRepository{}, &mockResStaffForStaff{}, nil, noopTransactor{})

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
	svc := NewStaffService(&mockStaffRepository{}, accountRepo, &mockAssignmentForStaff{}, &mockReservationForStaff{}, &mockShiftEntryForStaff{}, &mockPermissionGroupRepository{}, &mockResStaffForStaff{}, nil, noopTransactor{})

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
	svc := NewStaffService(staffRepo, accountRepo, assignmentRepo, &mockReservationForStaff{}, &mockShiftEntryForStaff{}, &mockPermissionGroupRepository{}, &mockResStaffForStaff{}, nil, noopTransactor{})

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
		svc := NewStaffService(&mockStaffRepository{}, accountRepo, &mockAssignmentForStaff{}, &mockReservationForStaff{}, &mockShiftEntryForStaff{}, &mockPermissionGroupRepository{}, &mockResStaffForStaff{}, nil, noopTransactor{})

		staff, err := svc.CreateWithAccount(context.Background(), baseInput)

		assert.Error(t, err)
		assert.Nil(t, staff)
	})

	t.Run("staff create fails", func(t *testing.T) {
		staffRepo := &mockStaffRepository{
			createFn: func(_ context.Context, _ *model.Staff) error { return errors.New("db error") },
		}
		svc := NewStaffService(staffRepo, &mockAccountForStaff{}, &mockAssignmentForStaff{}, &mockReservationForStaff{}, &mockShiftEntryForStaff{}, &mockPermissionGroupRepository{}, &mockResStaffForStaff{}, nil, noopTransactor{})

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
		svc := NewStaffService(staffRepo, &mockAccountForStaff{}, assignmentRepo, &mockReservationForStaff{}, &mockShiftEntryForStaff{}, &mockPermissionGroupRepository{}, &mockResStaffForStaff{}, nil, noopTransactor{})

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
			svc := NewStaffService(&mockStaffRepository{}, accountRepo, &mockAssignmentForStaff{}, &mockReservationForStaff{}, &mockShiftEntryForStaff{}, &mockPermissionGroupRepository{}, &mockResStaffForStaff{}, nil, noopTransactor{})

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
		svc := NewStaffService(&mockStaffRepository{}, &mockAccountForStaff{}, assignmentRepo, &mockReservationForStaff{}, &mockShiftEntryForStaff{}, &mockPermissionGroupRepository{}, &mockResStaffForStaff{}, nil, noopTransactor{})

		err := svc.SetClinicAssignments(context.Background(), 10, []uint64{1})

		assert.Error(t, err)
	})

	t.Run("create assignment fails", func(t *testing.T) {
		assignmentRepo := &mockAssignmentForStaff{
			createFn: func(_ context.Context, _ *model.StaffClinicAssignment) error { return errors.New("db error") },
		}
		svc := NewStaffService(&mockStaffRepository{}, &mockAccountForStaff{}, assignmentRepo, &mockReservationForStaff{}, &mockShiftEntryForStaff{}, &mockPermissionGroupRepository{}, &mockResStaffForStaff{}, nil, noopTransactor{})

		err := svc.SetClinicAssignments(context.Background(), 10, []uint64{1, 2})

		assert.Error(t, err)
	})

	t.Run("update primary clinic id fails", func(t *testing.T) {
		staffRepo := &mockStaffRepository{
			updatePrimaryFn: func(_ context.Context, _, _ uint64) error { return errors.New("db error") },
		}
		svc := NewStaffService(staffRepo, &mockAccountForStaff{}, &mockAssignmentForStaff{}, &mockReservationForStaff{}, &mockShiftEntryForStaff{}, &mockPermissionGroupRepository{}, &mockResStaffForStaff{}, nil, noopTransactor{})

		err := svc.SetClinicAssignments(context.Background(), 10, []uint64{1})

		assert.Error(t, err)
	})
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
			svc := NewStaffService(&mockStaffRepository{}, &mockAccountForStaff{}, assignmentRepo, &mockReservationForStaff{}, &mockShiftEntryForStaff{}, &mockPermissionGroupRepository{}, &mockResStaffForStaff{}, nil, noopTransactor{})

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
