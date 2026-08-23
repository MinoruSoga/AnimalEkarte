package staff

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/auth"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

type staffResetInvalidationAccountStore struct {
	auth.AccountRepository
	resetTokens auth.PasswordResetTokenInvalidator
}

type staffResetInvalidationAudit struct{}

func (staffResetInvalidationAudit) LogEntryTx(
	ctx context.Context,
	_ CredentialAuditEntry,
) error {
	if persistence.TxFromContext(ctx) == nil {
		return errors.New("staff credential audit is outside ambient transaction")
	}
	return nil
}

type authResetInvalidationAudit struct{}

func (authResetInvalidationAudit) LogEntryTx(
	ctx context.Context,
	_ auth.AuthAuditEntry,
) error {
	if persistence.TxFromContext(ctx) == nil {
		return errors.New("auth credential audit is outside ambient transaction")
	}
	return nil
}

type authResetInvalidationSubject struct {
	clinicID uint64
	staffID  uint64
}

func (s authResetInvalidationSubject) ResolveCredentialAuditSubject(
	ctx context.Context,
	_ uint64,
) (auth.CredentialAuditSubject, error) {
	if persistence.TxFromContext(ctx) == nil {
		return auth.CredentialAuditSubject{}, errors.New(
			"credential audit subject is outside ambient transaction",
		)
	}
	return auth.CredentialAuditSubject{
		ClinicID: s.clinicID,
		StaffID:  s.staffID,
	}, nil
}

func staffResetTokenHash(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}

func (s staffResetInvalidationAccountStore) DeletePasswordResetTokens(
	ctx context.Context,
	accountID uint64,
) error {
	return s.resetTokens.DeleteByAccountID(ctx, accountID)
}

func TestStaffService_UpdateAuthorizedPassword_InvalidatesOutstandingResetToken(
	t *testing.T,
) {
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.Account{},
		&model.PasswordResetToken{},
	))
	require.NoError(t, db.Exec("TRUNCATE TABLE password_reset_tokens, accounts CASCADE").Error)

	const (
		staffID       = uint64(9)
		clinicA       = uint64(10)
		clinicB       = uint64(20)
		newPassword   = "newPassw0rd"
		rawResetToken = "staff-admin-outstanding-reset-token"
	)
	account := &model.Account{
		Email:        "staff-password-reset-invalidation@example.test",
		PasswordHash: "old-password-hash",
		IsActive:     true,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(account).Error)

	accountRepo := auth.NewAccountRepository(db)
	tokenRepo := auth.NewPasswordResetTokenRepository(db)
	require.NoError(t, tokenRepo.Create(context.Background(), &model.PasswordResetToken{
		AccountID: account.ID,
		TokenHash: staffResetTokenHash(rawResetToken),
		ExpiresAt: time.Now().Add(time.Hour),
	}))
	transactor := persistence.NewTransactor(db)
	staff := &model.Staff{
		ID:        staffID,
		ClinicID:  clinicA,
		AccountID: &account.ID,
	}
	service := NewStaffServiceWithCredentialAudit(
		&coreMockStaffRepository{
			lockInClinicFn: func(
				context.Context,
				uint64,
				uint64,
			) (*model.Staff, error) {
				return staff, nil
			},
			lockForUpdateFn: func(
				_ context.Context,
				_ uint64,
			) (*model.Staff, error) {
				return staff, nil
			},
			findByIDFn: func(
				_ context.Context,
				_ uint64,
			) (*model.Staff, error) {
				return staff, nil
			},
			findByIDInClinicFn: func(
				context.Context,
				uint64,
				uint64,
			) (*model.Staff, error) {
				return staff, nil
			},
		},
		staffResetInvalidationAccountStore{
			AccountRepository: accountRepo,
			resetTokens:       tokenRepo,
		},
		&coreMockStaffClinicAssignmentRepository{
			lockActiveFn: func(
				context.Context,
				uint64,
			) ([]model.StaffClinicAssignment, error) {
				return []model.StaffClinicAssignment{
					{StaffID: staffID, ClinicID: clinicA},
					{StaffID: staffID, ClinicID: clinicB},
				}, nil
			},
		},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		transactor,
		staffResetInvalidationAudit{},
	)

	password := newPassword
	updated, err := service.Update(
		context.Background(),
		clinicA,
		staffID,
		&UpdateStaffInput{
			Password:        &password,
			IsSystemAdmin:   true,
			CredentialAudit: testStaffCredentialAudit(clinicA, staffID),
		},
	)
	require.NoError(t, err)
	require.NotNil(t, updated)

	passwordReset := auth.NewPasswordResetServiceWithCredentialAudit(
		&auth.PasswordResetConfig{},
		accountRepo,
		tokenRepo,
		nil,
		transactor,
		authResetInvalidationSubject{
			clinicID: clinicA,
			staffID:  staffID,
		},
		authResetInvalidationAudit{},
	)
	resetErr := passwordReset.ResetPassword(
		context.Background(),
		rawResetToken,
		"anotherPassw0rd",
	)
	require.Error(t, resetErr)
	assert.True(t, apperrors.IsInvalidInput(resetErr))
}
