package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type currentAccessAccountService struct {
	account *model.Account
	err     error
}

type fakeCurrentAccessStaffReader struct {
	staff *CurrentAccessStaffIdentity
	err   error
}

func (r fakeCurrentAccessStaffReader) FindCurrentAccessStaff(
	context.Context,
	uint64,
) (*CurrentAccessStaffIdentity, error) {
	return r.staff, r.err
}

func (s currentAccessAccountService) FindByEmail(
	context.Context,
	string,
) (*model.Account, error) {
	return nil, errors.New("unexpected FindByEmail call")
}

func (s currentAccessAccountService) GetByID(
	_ context.Context,
	_ uint64,
) (*model.Account, error) {
	return s.account, s.err
}

func (currentAccessAccountService) UpdatePasswordHash(
	context.Context,
	uint64,
	string,
) error {
	return errors.New("unexpected UpdatePasswordHash call")
}

type currentAccessAssignmentReader struct {
	assignments []model.StaffClinicAssignment
	err         error
}

func (r currentAccessAssignmentReader) FindAllByStaffID(
	context.Context,
	uint64,
) ([]model.StaffClinicAssignment, error) {
	return append([]model.StaffClinicAssignment(nil), r.assignments...), r.err
}

type currentAccessClinicReader struct {
	clinics []model.Clinic
	err     error
}

func (r currentAccessClinicReader) ListClinics(
	context.Context,
) ([]model.Clinic, error) {
	return append([]model.Clinic(nil), r.clinics...), r.err
}

func TestCurrentAccessResolver_UsesCurrentAccountAndActiveAssignments(t *testing.T) {
	accountID := uint64(41)
	updatedAt := time.Unix(1_721_000_000, 123)
	resolver := NewCurrentAccessResolverWithClinics(
		fakeCurrentAccessStaffReader{staff: &CurrentAccessStaffIdentity{
			ID:        17,
			AccountID: &accountID,
			IsActive:  true,
		}},
		currentAccessAccountService{account: &model.Account{
			ID:            accountID,
			IsActive:      true,
			IsSystemAdmin: true,
			UpdatedAt:     updatedAt,
		}},
		currentAccessAssignmentReader{assignments: []model.StaffClinicAssignment{
			{StaffID: 17, ClinicID: 23},
			{StaffID: 17, ClinicID: 24, IsMain: true},
			{StaffID: 17, ClinicID: 23},
			{
				StaffID:   17,
				ClinicID:  25,
				DeletedAt: deletedAtForCurrentAccessTest(),
			},
		}},
		currentAccessClinicReader{clinics: []model.Clinic{
			{ID: 0, IsActive: true},
			{ID: 23, IsActive: true},
			{ID: 24, IsActive: true},
			{ID: 25, IsActive: false},
			{ID: 31, IsActive: true},
		}},
	)

	access, err := resolver.Resolve(context.Background(), 17)

	require.NoError(t, err)
	assert.Equal(t, updatedAt.UnixNano(), access.AccountEpoch)
	assert.True(t, access.IsSystemAdmin)
	assert.Equal(t, []uint64{23, 24, 31}, access.ClinicIDs)
	assert.Equal(t, "24", access.MainClinicID)
}

func TestCurrentAccessResolver_RegularStaffUsesScopedActiveIDsWithoutListClinics(
	t *testing.T,
) {
	accountID := uint64(41)
	scoped := &scopedCurrentAccessClinicReader{activeIDs: []uint64{24}}
	resolver := NewCurrentAccessResolverWithClinics(
		fakeCurrentAccessStaffReader{staff: &CurrentAccessStaffIdentity{
			ID:        17,
			AccountID: &accountID,
			IsActive:  true,
		}},
		currentAccessAccountService{account: &model.Account{
			ID:        accountID,
			IsActive:  true,
			UpdatedAt: time.Unix(1_721_000_000, 0),
		}},
		currentAccessAssignmentReader{assignments: []model.StaffClinicAssignment{
			{StaffID: 17, ClinicID: 23, IsMain: true},
			{StaffID: 17, ClinicID: 24},
		}},
		scoped,
	)

	access, err := resolver.Resolve(context.Background(), 17)

	require.NoError(t, err)
	assert.Equal(t, []uint64{24}, access.ClinicIDs)
	assert.Equal(t, 0, scoped.listAllCalls)
}

type scopedCurrentAccessClinicReader struct {
	activeIDs    []uint64
	listAllCalls int
}

func (r *scopedCurrentAccessClinicReader) ListClinics(
	context.Context,
) ([]model.Clinic, error) {
	r.listAllCalls++
	return nil, errors.New("ListClinics must not run for non-admin staff")
}

func (r *scopedCurrentAccessClinicReader) ListActiveClinicIDs(
	_ context.Context,
	_ []uint64,
) ([]uint64, error) {
	return append([]uint64(nil), r.activeIDs...), nil
}

func TestCurrentAccessResolver_RegularStaffUsesOnlyActiveClinicInventory(
	t *testing.T,
) {
	accountID := uint64(41)
	resolver := NewCurrentAccessResolverWithClinics(
		fakeCurrentAccessStaffReader{staff: &CurrentAccessStaffIdentity{
			ID:        17,
			AccountID: &accountID,
			IsActive:  true,
		}},
		currentAccessAccountService{account: &model.Account{
			ID:        accountID,
			IsActive:  true,
			UpdatedAt: time.Unix(1_721_000_000, 0),
		}},
		currentAccessAssignmentReader{assignments: []model.StaffClinicAssignment{
			{StaffID: 17, ClinicID: 23, IsMain: true},
			{StaffID: 17, ClinicID: 24},
			{StaffID: 17, ClinicID: 25},
			{StaffID: 17, ClinicID: 24},
			{
				StaffID:   17,
				ClinicID:  26,
				DeletedAt: deletedAtForCurrentAccessTest(),
			},
		}},
		currentAccessClinicReader{clinics: []model.Clinic{
			{ID: 23, IsActive: false},
			{ID: 24, IsActive: true},
			{ID: 26, IsActive: true},
			{ID: 31, IsActive: true},
		}},
	)

	access, err := resolver.Resolve(context.Background(), 17)

	require.NoError(t, err)
	assert.False(t, access.IsSystemAdmin)
	assert.Equal(t, []uint64{24}, access.ClinicIDs)
	assert.Equal(t, "24", access.MainClinicID)
}

func TestCurrentAccessResolver_RegularStaffInactiveOnlyIsForbidden(
	t *testing.T,
) {
	accountID := uint64(41)
	resolver := NewCurrentAccessResolverWithClinics(
		fakeCurrentAccessStaffReader{staff: &CurrentAccessStaffIdentity{
			ID:        17,
			AccountID: &accountID,
			IsActive:  true,
		}},
		currentAccessAccountService{account: &model.Account{
			ID:        accountID,
			IsActive:  true,
			UpdatedAt: time.Unix(1_721_000_000, 0),
		}},
		currentAccessAssignmentReader{assignments: []model.StaffClinicAssignment{
			{StaffID: 17, ClinicID: 23, IsMain: true},
			{StaffID: 17, ClinicID: 25},
		}},
		currentAccessClinicReader{clinics: []model.Clinic{
			{ID: 23, IsActive: false},
			{ID: 24, IsActive: true},
		}},
	)

	access, err := resolver.Resolve(context.Background(), 17)

	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
	assert.Nil(t, access)
}

func TestCurrentAccessResolver_RegularStaffClinicAuthorityFailsClosed(
	t *testing.T,
) {
	accountID := uint64(41)
	activeStaff := fakeCurrentAccessStaffReader{staff: &CurrentAccessStaffIdentity{
		ID:        17,
		AccountID: &accountID,
		IsActive:  true,
	}}
	activeAccount := currentAccessAccountService{account: &model.Account{
		ID:        accountID,
		IsActive:  true,
		UpdatedAt: time.Unix(1_721_000_000, 0),
	}}
	activeAssignments := currentAccessAssignmentReader{
		assignments: []model.StaffClinicAssignment{{
			StaffID:  17,
			ClinicID: 23,
			IsMain:   true,
		}},
	}

	for _, test := range []struct {
		name    string
		clinics CurrentAccessClinicReader
	}{
		{name: "missing clinic dependency"},
		{
			name: "clinic lookup failure",
			clinics: currentAccessClinicReader{
				err: errors.New("clinics unavailable"),
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			resolver := NewCurrentAccessResolverWithClinics(
				activeStaff,
				activeAccount,
				activeAssignments,
				test.clinics,
			)

			access, err := resolver.Resolve(context.Background(), 17)

			require.Error(t, err)
			assert.Nil(t, access)
		})
	}
}

func TestCurrentAccessResolver_SystemAdminClinicAuthorityFailsClosed(t *testing.T) {
	accountID := uint64(41)
	activeStaff := fakeCurrentAccessStaffReader{staff: &CurrentAccessStaffIdentity{
		ID:        17,
		AccountID: &accountID,
		IsActive:  true,
	}}
	activeAccount := currentAccessAccountService{account: &model.Account{
		ID:            accountID,
		IsActive:      true,
		IsSystemAdmin: true,
		UpdatedAt:     time.Unix(1_721_000_000, 0),
	}}

	tests := []struct {
		name       string
		clinics    CurrentAccessClinicReader
		wantTarget error
	}{
		{
			name: "missing clinic dependency",
		},
		{
			name: "clinic lookup failure",
			clinics: currentAccessClinicReader{
				err: errors.New("clinics unavailable"),
			},
		},
		{
			name: "no active nonzero clinic",
			clinics: currentAccessClinicReader{clinics: []model.Clinic{
				{ID: 0, IsActive: true},
				{ID: 23, IsActive: false},
			}},
			wantTarget: apperrors.ErrForbidden,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resolver := NewCurrentAccessResolverWithClinics(
				activeStaff,
				activeAccount,
				currentAccessAssignmentReader{},
				test.clinics,
			)

			access, err := resolver.Resolve(context.Background(), 17)

			require.Error(t, err)
			assert.Nil(t, access)
			if test.wantTarget != nil {
				assert.ErrorIs(t, err, test.wantTarget)
			}
		})
	}
}

func TestCurrentAccessResolver_FailsClosedForInvalidIdentityAndDependencies(
	t *testing.T,
) {
	accountID := uint64(41)
	activeAccount := &model.Account{
		ID:        accountID,
		IsActive:  true,
		UpdatedAt: time.Unix(1_721_000_000, 0),
	}
	activeStaff := &CurrentAccessStaffIdentity{
		ID:        17,
		AccountID: &accountID,
		IsActive:  true,
	}
	tests := []struct {
		name            string
		staff           *CurrentAccessStaffIdentity
		staffErr        error
		accounts        CurrentAccessAccountReader
		assignments     CurrentAccessAssignmentReader
		wantTarget      error
		wantStaffLookup bool
	}{
		{name: "staff lookup error remains stage typed", staffErr: errors.New("staff unavailable"), accounts: currentAccessAccountService{account: activeAccount}, assignments: currentAccessAssignmentReader{}, wantStaffLookup: true},
		{name: "nil staff", accounts: currentAccessAccountService{account: activeAccount}, assignments: currentAccessAssignmentReader{}, wantTarget: apperrors.ErrForbidden},
		{name: "mismatched staff identity", staff: &CurrentAccessStaffIdentity{ID: 99, AccountID: &accountID, IsActive: true}, accounts: currentAccessAccountService{account: activeAccount}, assignments: currentAccessAssignmentReader{}, wantTarget: apperrors.ErrForbidden},
		{name: "inactive staff", staff: &CurrentAccessStaffIdentity{ID: 17, AccountID: &accountID}, accounts: currentAccessAccountService{account: activeAccount}, assignments: currentAccessAssignmentReader{}, wantTarget: apperrors.ErrForbidden},
		{name: "deleted staff", staff: &CurrentAccessStaffIdentity{ID: 17, AccountID: &accountID, IsActive: true, IsDeleted: true}, accounts: currentAccessAccountService{account: activeAccount}, assignments: currentAccessAssignmentReader{}, wantTarget: apperrors.ErrForbidden},
		{name: "unlinked staff", staff: &CurrentAccessStaffIdentity{ID: 17, IsActive: true}, accounts: currentAccessAccountService{account: activeAccount}, assignments: currentAccessAssignmentReader{}, wantTarget: apperrors.ErrUnauthorized},
		{name: "missing account dependency", staff: activeStaff, assignments: currentAccessAssignmentReader{}},
		{name: "missing assignment dependency", staff: activeStaff, accounts: currentAccessAccountService{account: activeAccount}},
		{name: "account lookup error", staff: activeStaff, accounts: currentAccessAccountService{err: errors.New("account unavailable")}, assignments: currentAccessAssignmentReader{}},
		{name: "account not found", staff: activeStaff, accounts: currentAccessAccountService{err: apperrors.WrapNotFound("account", "41")}, assignments: currentAccessAssignmentReader{}, wantTarget: apperrors.ErrUnauthorized},
		{name: "nil account", staff: activeStaff, accounts: currentAccessAccountService{}, assignments: currentAccessAssignmentReader{}, wantTarget: apperrors.ErrUnauthorized},
		{name: "mismatched account identity", staff: activeStaff, accounts: currentAccessAccountService{account: &model.Account{ID: 99, IsActive: true, UpdatedAt: activeAccount.UpdatedAt}}, assignments: currentAccessAssignmentReader{}, wantTarget: apperrors.ErrUnauthorized},
		{name: "inactive account", staff: activeStaff, accounts: currentAccessAccountService{account: &model.Account{ID: accountID, UpdatedAt: activeAccount.UpdatedAt}}, assignments: currentAccessAssignmentReader{}, wantTarget: apperrors.ErrUnauthorized},
		{name: "deleted account", staff: activeStaff, accounts: currentAccessAccountService{account: &model.Account{ID: accountID, IsActive: true, UpdatedAt: activeAccount.UpdatedAt, DeletedAt: deletedAtForCurrentAccessTest()}}, assignments: currentAccessAssignmentReader{}, wantTarget: apperrors.ErrUnauthorized},
		{name: "invalid account epoch", staff: activeStaff, accounts: currentAccessAccountService{account: &model.Account{ID: accountID, IsActive: true}}, assignments: currentAccessAssignmentReader{}},
		{name: "assignment lookup error", staff: activeStaff, accounts: currentAccessAccountService{account: activeAccount}, assignments: currentAccessAssignmentReader{err: errors.New("assignments unavailable")}},
		{name: "zero clinic assignment", staff: activeStaff, accounts: currentAccessAccountService{account: activeAccount}, assignments: currentAccessAssignmentReader{assignments: []model.StaffClinicAssignment{{StaffID: 17, ClinicID: 0}}}},
		{name: "mismatched assignment staff", staff: activeStaff, accounts: currentAccessAccountService{account: activeAccount}, assignments: currentAccessAssignmentReader{assignments: []model.StaffClinicAssignment{{StaffID: 99, ClinicID: 23}}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resolver := NewCurrentAccessResolver(
				fakeCurrentAccessStaffReader{staff: test.staff, err: test.staffErr},
				test.accounts,
				test.assignments,
			)

			access, err := resolver.Resolve(context.Background(), 17)

			require.Error(t, err)
			assert.Nil(t, access)
			if test.wantTarget != nil {
				assert.ErrorIs(t, err, test.wantTarget)
			}
			if test.wantStaffLookup {
				var staffLookup *StaffLookupError
				assert.ErrorAs(t, err, &staffLookup)
			}
		})
	}
}

func TestCurrentAccessStaffReader_FailsClosedWithoutDatabase(t *testing.T) {
	reader := NewCurrentAccessStaffReader(nil)

	staff, err := reader.FindCurrentAccessStaff(context.Background(), 17)

	require.Error(t, err)
	assert.Nil(t, staff)
}

func TestStaffLookupError_PreservesCause(t *testing.T) {
	cause := errors.New("temporary staff lookup failure")
	err := &StaffLookupError{cause: cause}

	assert.Contains(t, err.Error(), "failed to resolve current staff")
	assert.ErrorIs(t, err, cause)
}

func deletedAtForCurrentAccessTest() gorm.DeletedAt {
	return gorm.DeletedAt{Time: time.Now(), Valid: true}
}
