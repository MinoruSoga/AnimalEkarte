package staff

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestService_AttachAccount_RejectsNonAdmin(t *testing.T) {
	svc := newCoreService(
		&coreMockStaffRepository{},
		&coreMockAccountRepository{},
		&coreMockStaffClinicAssignmentRepository{},
		&coreMockReservationQueryRepository{},
		&coreMockShiftEntryRepository{},
		&coreFakeTransactor{},
	)
	staff, err := svc.AttachAccount(context.Background(), 1, 10, &AttachStaffAccountInput{
		Email:           "staff@example.test",
		IsSystemAdmin:   false,
		CredentialAudit: testStaffCredentialAudit(1, 10),
	})
	require.Error(t, err)
	assert.Nil(t, staff)
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
}

func TestService_AttachAccount_RejectsCatalogEmail(t *testing.T) {
	svc := newCoreService(
		&coreMockStaffRepository{},
		&coreMockAccountRepository{},
		&coreMockStaffClinicAssignmentRepository{},
		&coreMockReservationQueryRepository{},
		&coreMockShiftEntryRepository{},
		&coreFakeTransactor{},
	)
	staff, err := svc.AttachAccount(context.Background(), 1, 10, &AttachStaffAccountInput{
		Email:           "stg-staff-10000021@example.test",
		IsSystemAdmin:   true,
		CredentialAudit: testStaffCredentialAudit(1, 10),
	})
	require.Error(t, err)
	assert.Nil(t, staff)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestService_AttachAccount_CreatesAccountWithoutReturningSecret(t *testing.T) {
	created := false
	attached := false
	repo := &coreMockStaffRepository{
		lockInClinicFn: func(_ context.Context, clinicID, id uint64) (*model.Staff, error) {
			return &model.Staff{ID: id, ClinicID: clinicID, IsActive: true}, nil
		},
		findByIDInClinicFn: func(_ context.Context, clinicID, id uint64) (*model.Staff, error) {
			accountID := uint64(88)
			return &model.Staff{
				ID:        id,
				ClinicID:  clinicID,
				AccountID: &accountID,
				Account:   &model.Account{ID: accountID, Email: "staff@example.test"},
			}, nil
		},
		attachAccountIDFn: func(_ context.Context, clinicID, staffID, accountID uint64) error {
			attached = true
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(10), staffID)
			assert.Equal(t, uint64(88), accountID)
			return nil
		},
	}
	accountRepo := &coreMockAccountRepository{
		createFn: func(_ context.Context, account *model.Account) error {
			created = true
			assert.Equal(t, "staff@example.test", account.Email)
			assert.NotEmpty(t, account.PasswordHash)
			assert.True(t, account.IsActive)
			assert.False(t, account.IsSystemAdmin)
			account.ID = 88
			return nil
		},
	}
	svc := newCoreService(
		repo,
		accountRepo,
		&coreMockStaffClinicAssignmentRepository{},
		&coreMockReservationQueryRepository{},
		&coreMockShiftEntryRepository{},
		&coreFakeTransactor{},
	)
	staff, err := svc.AttachAccount(context.Background(), 1, 10, &AttachStaffAccountInput{
		Email:           "staff@example.test",
		IsSystemAdmin:   true,
		CredentialAudit: testStaffCredentialAudit(1, 10),
	})
	require.NoError(t, err)
	require.NotNil(t, staff)
	assert.True(t, created)
	assert.True(t, attached)
	assert.Equal(t, uint64(10), staff.ID)
}

func TestService_AttachAccount_ConflictWhenAccountExists(t *testing.T) {
	accountID := uint64(3)
	svc := newCoreService(
		&coreMockStaffRepository{
			lockInClinicFn: func(_ context.Context, clinicID, id uint64) (*model.Staff, error) {
				return &model.Staff{ID: id, ClinicID: clinicID, IsActive: true, AccountID: &accountID}, nil
			},
		},
		&coreMockAccountRepository{},
		&coreMockStaffClinicAssignmentRepository{},
		&coreMockReservationQueryRepository{},
		&coreMockShiftEntryRepository{},
		&coreFakeTransactor{},
	)
	staff, err := svc.AttachAccount(context.Background(), 1, 10, &AttachStaffAccountInput{
		Email:           "staff@example.test",
		IsSystemAdmin:   true,
		CredentialAudit: testStaffCredentialAudit(1, 10),
	})
	require.Error(t, err)
	assert.Nil(t, staff)
	assert.True(t, apperrors.IsConflict(err))
}
