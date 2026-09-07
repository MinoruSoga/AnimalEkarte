package staff

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
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

func (m *mockClinicMembershipRepo) FindByStaffAndClinic(
	_ context.Context,
	staffID, clinicID uint64,
) (*model.StaffClinicAssignment, error) {
	return &model.StaffClinicAssignment{StaffID: staffID, ClinicID: clinicID}, nil
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

func (m *mockClinicMembershipRepo) DeleteByStaffAndClinicIDs(
	ctx context.Context,
	staffID uint64,
	clinicIDs []uint64,
) error {
	if len(clinicIDs) == 0 {
		return nil
	}
	return m.Delete(ctx, staffID)
}

var _ StaffClinicAssignmentRepository = (*mockClinicMembershipRepo)(nil)

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

func TestService_FindByAccountID(t *testing.T) {
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
			svc := newTestService(repo)

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

func TestService_CreateWithAccount_ValidationErrors(t *testing.T) {
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
			svc := newTestService(repo)

			staff, err := svc.CreateWithAccount(context.Background(), tt.input)

			assert.Error(t, err)
			assert.Nil(t, staff)
			assert.True(t, apperrors.IsInvalidInput(err))
		})
	}
}

func TestService_CreateWithAccount_EmailUniquenessCheckError(t *testing.T) {
	accountRepo := &mockAccountForStaff{
		findByEmailFn: func(_ context.Context, _ string) (*model.Account, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewService(&mockStaffRepository{}, accountRepo, &mockAssignmentForStaff{}, &mockReservationForStaff{}, &mockShiftEntryForStaff{}, &mockPermissionGroupRepository{}, &mockResStaffForStaff{}, nil, nil, noopTransactor{})

	staff, err := svc.CreateWithAccount(context.Background(), &CreateStaffWithAccountInput{
		ClinicID: 1, Name: "スタッフ", Email: "a@example.com", Password: "Passw0rd1",
	})

	assert.Error(t, err)
	assert.Nil(t, staff)
}

func TestService_CreateWithAccount_EmailAlreadyExists(t *testing.T) {
	accountRepo := &mockAccountForStaff{
		findByEmailFn: func(_ context.Context, email string) (*model.Account, error) {
			return &model.Account{ID: 99, Email: email}, nil
		},
	}
	svc := NewService(&mockStaffRepository{}, accountRepo, &mockAssignmentForStaff{}, &mockReservationForStaff{}, &mockShiftEntryForStaff{}, &mockPermissionGroupRepository{}, &mockResStaffForStaff{}, nil, nil, noopTransactor{})

	staff, err := svc.CreateWithAccount(context.Background(), &CreateStaffWithAccountInput{
		ClinicID: 1, Name: "スタッフ", Email: "dup@example.com", Password: "Passw0rd1",
	})

	assert.Error(t, err)
	assert.Nil(t, staff)
	assert.True(t, apperrors.IsAlreadyExists(err))
}

func TestService_CreateWithAccount_PasswordTooLongForBcrypt(t *testing.T) {
	// bcrypt rejects passwords over 72 bytes; this is >90 bytes and still satisfies
	// validatePassword (contains letters and digits, length >= 8).
	longPassword := strings.Repeat("Aa1", 30)
	svc := newTestService(&mockStaffRepository{})

	staff, err := svc.CreateWithAccount(context.Background(), &CreateStaffWithAccountInput{
		ClinicID: 1, Name: "スタッフ", Email: "a@example.com", Password: longPassword,
	})

	assert.Error(t, err)
	assert.Nil(t, staff)
}

func TestService_CreateWithAccount_Success(t *testing.T) {
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
	svc := NewService(staffRepo, accountRepo, assignmentRepo, &mockReservationForStaff{}, &mockShiftEntryForStaff{}, &mockPermissionGroupRepository{}, &mockResStaffForStaff{}, nil, nil, noopTransactor{})

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

func TestService_CreateWithAccount_CustomStaffTypeAndReservationVisible(t *testing.T) {
	visible := false
	staffRepo := &mockStaffRepository{
		createFn: func(_ context.Context, s *model.Staff) error {
			s.ID = 1
			return nil
		},
	}
	svc := newTestService(staffRepo)

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

func TestService_CreateWithAccount_TxFailures(t *testing.T) {
	baseInput := &CreateStaffWithAccountInput{ClinicID: 1, Name: "スタッフ", Email: "a@example.com", Password: "Passw0rd1"}

	t.Run("account create fails", func(t *testing.T) {
		accountRepo := &mockAccountForStaff{
			createFn: func(_ context.Context, _ *model.Account) error { return errors.New("db error") },
		}
		svc := NewService(&mockStaffRepository{}, accountRepo, &mockAssignmentForStaff{}, &mockReservationForStaff{}, &mockShiftEntryForStaff{}, &mockPermissionGroupRepository{}, &mockResStaffForStaff{}, nil, nil, noopTransactor{})

		staff, err := svc.CreateWithAccount(context.Background(), baseInput)

		assert.Error(t, err)
		assert.Nil(t, staff)
	})

	t.Run("staff create fails", func(t *testing.T) {
		staffRepo := &mockStaffRepository{
			createFn: func(_ context.Context, _ *model.Staff) error { return errors.New("db error") },
		}
		svc := NewService(staffRepo, &mockAccountForStaff{}, &mockAssignmentForStaff{}, &mockReservationForStaff{}, &mockShiftEntryForStaff{}, &mockPermissionGroupRepository{}, &mockResStaffForStaff{}, nil, nil, noopTransactor{})

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
		svc := NewService(staffRepo, &mockAccountForStaff{}, assignmentRepo, &mockReservationForStaff{}, &mockShiftEntryForStaff{}, &mockPermissionGroupRepository{}, &mockResStaffForStaff{}, nil, nil, noopTransactor{})

		staff, err := svc.CreateWithAccount(context.Background(), baseInput)

		assert.Error(t, err)
		assert.Nil(t, staff)
	})
}

// ---- SetClinicAssignments (additional error branches) ----
