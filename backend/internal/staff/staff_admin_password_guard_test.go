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

func TestService_UpdatePassword_RejectsNonAdminChangingSystemAdmin(t *testing.T) {
	accountID := uint64(5)
	password := "NewPassw0rd"
	passwordUpdated := false
	repo := &coreMockStaffRepository{
		findByIDFn: func(_ context.Context, id uint64) (*model.Staff, error) {
			return &model.Staff{ID: id, ClinicID: 1, AccountID: &accountID, IsActive: true}, nil
		},
	}
	accountRepo := &coreMockAccountRepository{
		updatePasswordHashFn: func(context.Context, uint64, string, time.Time) error {
			passwordUpdated = true
			return nil
		},
	}
	accountRepo.findByIDForUpdate = func(_ context.Context, id uint64) (*model.Account, error) {
		return &model.Account{ID: id, IsActive: true, IsSystemAdmin: true}, nil
	}
	svc := newCoreService(
		repo,
		accountRepo,
		&coreMockStaffClinicAssignmentRepository{},
		&coreMockReservationQueryRepository{},
		&coreMockShiftEntryRepository{},
		&coreFakeTransactor{},
	)
	input := authorizedStaffUpdate(&UpdateStaffInput{
		Password:        &password,
		IsSystemAdmin:   false,
		CredentialAudit: testStaffCredentialAudit(1, 1),
	}, 1)

	staff, err := svc.Update(context.Background(), 1, 1, input)
	require.Error(t, err)
	assert.Nil(t, staff)
	assert.True(t, errors.Is(err, apperrors.ErrForbidden))
	assert.False(t, passwordUpdated)
}

func TestService_UpdatePassword_AllowsAdminChangingSystemAdmin(t *testing.T) {
	accountID := uint64(5)
	password := "NewPassw0rd"
	passwordUpdated := false
	repo := &coreMockStaffRepository{
		findByIDFn: func(_ context.Context, id uint64) (*model.Staff, error) {
			return &model.Staff{ID: id, ClinicID: 1, AccountID: &accountID, IsActive: true}, nil
		},
		lockForUpdateFn: func(_ context.Context, id uint64) (*model.Staff, error) {
			return &model.Staff{ID: id, ClinicID: 1, AccountID: &accountID, IsActive: true}, nil
		},
	}
	accountRepo := &coreMockAccountRepository{
		updatePasswordHashFn: func(context.Context, uint64, string, time.Time) error {
			passwordUpdated = true
			return nil
		},
	}
	accountRepo.findByIDForUpdate = func(_ context.Context, id uint64) (*model.Account, error) {
		return &model.Account{ID: id, IsActive: true, IsSystemAdmin: true}, nil
	}
	svc := newCoreService(
		repo,
		accountRepo,
		&coreMockStaffClinicAssignmentRepository{},
		&coreMockReservationQueryRepository{},
		&coreMockShiftEntryRepository{},
		&coreFakeTransactor{},
	)
	input := authorizedStaffUpdate(&UpdateStaffInput{
		Password:        &password,
		IsSystemAdmin:   true,
		CredentialAudit: testStaffCredentialAudit(1, 1),
	}, 1)

	staff, err := svc.Update(context.Background(), 1, 1, input)
	require.NoError(t, err)
	assert.NotNil(t, staff)
	assert.True(t, passwordUpdated)
}

func TestService_UpdatePassword_DoesNotMutateTokensOrAuditWhenForbidden(t *testing.T) {
	accountID := uint64(5)
	password := "NewPassw0rd"
	tokensDeleted := false
	auditLogged := false
	accountRepo := &coreMockAccountRepository{
		updatePasswordHashFn: func(context.Context, uint64, string, time.Time) error {
			t.Fatal("password hash must not change")
			return nil
		},
		deletePasswordResetTokensFn: func(context.Context, uint64) error {
			tokensDeleted = true
			return nil
		},
		findByIDForUpdate: func(_ context.Context, id uint64) (*model.Account, error) {
			return &model.Account{ID: id, IsActive: true, IsSystemAdmin: true}, nil
		},
	}
	svc := NewServiceWithCredentialAudit(
		&coreMockStaffRepository{
			findByIDFn: func(_ context.Context, id uint64) (*model.Staff, error) {
				return &model.Staff{ID: id, ClinicID: 1, AccountID: &accountID, IsActive: true}, nil
			},
		},
		accountRepo,
		&coreMockStaffClinicAssignmentRepository{},
		&coreMockReservationQueryRepository{},
		&coreMockShiftEntryRepository{},
		nil,
		nil,
		nil,
		nil,
		&coreFakeTransactor{},
		recordingStaffCredentialAuditTxLogger{logged: &auditLogged},
	)
	input := authorizedStaffUpdate(&UpdateStaffInput{
		Password:        &password,
		IsSystemAdmin:   false,
		CredentialAudit: testStaffCredentialAudit(1, 1),
	}, 1)
	staff, err := svc.Update(context.Background(), 1, 1, input)
	require.Error(t, err)
	assert.Nil(t, staff)
	assert.True(t, errors.Is(err, apperrors.ErrForbidden))
	assert.False(t, tokensDeleted)
	assert.False(t, auditLogged)
}

func TestService_UpdatePassword_AccountLockFailureIsRejected(t *testing.T) {
	accountID := uint64(5)
	password := "NewPassw0rd"
	accountRepo := &coreMockAccountRepository{
		findByIDForUpdate: func(context.Context, uint64) (*model.Account, error) {
			return nil, errors.New("lock failed")
		},
	}
	svc := newCoreService(
		&coreMockStaffRepository{
			findByIDFn: func(_ context.Context, id uint64) (*model.Staff, error) {
				return &model.Staff{ID: id, ClinicID: 1, AccountID: &accountID, IsActive: true}, nil
			},
		},
		accountRepo,
		&coreMockStaffClinicAssignmentRepository{},
		&coreMockReservationQueryRepository{},
		&coreMockShiftEntryRepository{},
		&coreFakeTransactor{},
	)
	input := authorizedStaffUpdate(&UpdateStaffInput{
		Password:        &password,
		IsSystemAdmin:   true,
		CredentialAudit: testStaffCredentialAudit(1, 1),
	}, 1)
	staff, err := svc.Update(context.Background(), 1, 1, input)
	require.Error(t, err)
	assert.Nil(t, staff)
}

type recordingStaffCredentialAuditTxLogger struct {
	logged *bool
}

func (l recordingStaffCredentialAuditTxLogger) LogEntryTx(
	context.Context,
	CredentialAuditEntry,
) error {
	if l.logged != nil {
		*l.logged = true
	}
	return nil
}
