package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type credentialAuditTxMarker struct{}

type noopCredentialAuditTxLogger struct{}

func (noopCredentialAuditTxLogger) LogEntryTx(
	context.Context,
	AuthAuditEntry,
) error {
	return nil
}

type noopCredentialAuditSubjectResolver struct{}

func (noopCredentialAuditSubjectResolver) ResolveCredentialAuditSubject(
	context.Context,
	uint64,
) (CredentialAuditSubject, error) {
	return CredentialAuditSubject{ClinicID: 1, StaffID: 1}, nil
}

func testPasswordChangeAudit(staffID uint64) CredentialMutationAudit {
	actorID := staffID
	return CredentialMutationAudit{
		ClinicID:      1,
		ActorID:       &actorID,
		ActorType:     model.AuditActorTypeStaff,
		Action:        model.AuditActionAuthPasswordChange,
		TargetStaffID: staffID,
	}
}

func testPasswordResetAudit() CredentialMutationAudit {
	return CredentialMutationAudit{
		ActorType: model.AuditActorTypeSystem,
		Action:    model.AuditActionAuthPasswordReset,
	}
}

type credentialAuditTxCapture struct {
	entry AuthAuditEntry
	err   error
}

func (a *credentialAuditTxCapture) LogEntryTx(
	ctx context.Context,
	entry AuthAuditEntry,
) error {
	if ctx.Value(credentialAuditTxMarker{}) != true {
		return errors.New("credential audit did not join the ambient transaction")
	}
	a.entry = entry
	return a.err
}

type credentialAuditSubjectResolverStub struct {
	subject CredentialAuditSubject
	err     error
}

func (r credentialAuditSubjectResolverStub) ResolveCredentialAuditSubject(
	ctx context.Context,
	accountID uint64,
) (CredentialAuditSubject, error) {
	if ctx.Value(credentialAuditTxMarker{}) != true {
		return CredentialAuditSubject{}, errors.New(
			"credential audit subject lookup did not join the ambient transaction",
		)
	}
	return r.subject, r.err
}

func TestAccountService_ChangePassword_AuditFailureRollsBackCredential(t *testing.T) {
	const (
		accountID       = uint64(41)
		staffID         = uint64(17)
		clinicID        = uint64(23)
		currentPassword = "oldPassw0rd"
		newPassword     = "newPassw0rd"
	)
	currentHash, err := bcrypt.GenerateFromPassword(
		[]byte(currentPassword),
		bcrypt.MinCost,
	)
	require.NoError(t, err)
	committedHash := string(currentHash)
	stagedHash := ""
	rolledBack := false
	auditErr := errors.New("credential audit unavailable")
	audit := &credentialAuditTxCapture{err: auditErr}
	repo := &mockAccountRepository{
		findByIDFn: func(context.Context, uint64) (*model.Account, error) {
			return &model.Account{
				ID:           accountID,
				PasswordHash: committedHash,
				UpdatedAt:    time.Now().Add(-time.Minute),
			}, nil
		},
		compareAndSwapFn: func(
			ctx context.Context,
			_ uint64,
			_ string,
			newHash string,
			_ time.Time,
		) (bool, error) {
			assert.Equal(t, true, ctx.Value(credentialAuditTxMarker{}))
			stagedHash = newHash
			return true, nil
		},
	}
	transactor := accountChangeTransactorFunc(func(
		ctx context.Context,
		fn func(context.Context) error,
	) error {
		txCtx := context.WithValue(ctx, credentialAuditTxMarker{}, true)
		if txErr := fn(txCtx); txErr != nil {
			rolledBack = true
			stagedHash = ""
			return txErr
		}
		committedHash = stagedHash
		return nil
	})
	service := NewAccountServiceWithCredentialAudit(
		repo,
		accountResetTokenInvalidatorFunc(func(
			ctx context.Context,
			id uint64,
		) error {
			assert.Equal(t, true, ctx.Value(credentialAuditTxMarker{}))
			assert.Equal(t, accountID, id)
			return nil
		}),
		transactor,
		audit,
	)
	changer := service.(AccountPasswordChanger)
	actorID := staffID

	changeErr := changer.ChangePassword(
		context.Background(),
		accountID,
		currentPassword,
		newPassword,
		CredentialMutationAudit{
			ClinicID:      clinicID,
			ActorID:       &actorID,
			ActorType:     model.AuditActorTypeStaff,
			Action:        model.AuditActionAuthPasswordChange,
			TargetStaffID: staffID,
			IPAddress:     "192.0.2.1",
			UserAgent:     "credential-audit-test",
		},
	)

	require.Error(t, changeErr)
	assert.ErrorIs(t, changeErr, auditErr)
	assert.True(t, rolledBack)
	assert.Empty(t, stagedHash)
	assert.Equal(t, string(currentHash), committedHash)
	require.NotNil(t, audit.entry.ResourceID)
	assert.Equal(t, accountID, *audit.entry.ResourceID)
	assert.Equal(t, map[string]any{"staff_id": staffID}, audit.entry.NewValue)
}

func TestAccountService_ChangePasswordFailsClosedWithoutAuditBeforeMutation(
	t *testing.T,
) {
	repositoryCalled := false
	repo := &mockAccountRepository{
		findByIDFn: func(context.Context, uint64) (*model.Account, error) {
			repositoryCalled = true
			return nil, errors.New("must not be reached")
		},
	}
	service := NewAccountServiceWithCredentialInvalidation(
		repo,
		accountResetTokenInvalidatorFunc(func(context.Context, uint64) error {
			return nil
		}),
		accountChangeTransactorFunc(func(
			ctx context.Context,
			fn func(context.Context) error,
		) error {
			return fn(ctx)
		}),
	)

	changeErr := service.(AccountPasswordChanger).ChangePassword(
		context.Background(),
		41,
		"oldPassw0rd",
		"newPassw0rd",
		testPasswordChangeAudit(17),
	)

	require.Error(t, changeErr)
	var appErr *apperrors.AppError
	require.ErrorAs(t, changeErr, &appErr)
	assert.Equal(t, "INTERNAL", appErr.Code)
	assert.False(t, repositoryCalled)
}

func TestPasswordResetService_AuditFailureRollsBackPasswordAndTokenConsumption(
	t *testing.T,
) {
	const (
		accountID = uint64(41)
		staffID   = uint64(17)
		clinicID  = uint64(23)
		rawToken  = "credential-audit-reset-token"
	)
	now := time.Now()
	token := &model.PasswordResetToken{
		ID:        71,
		AccountID: accountID,
		ExpiresAt: now.Add(time.Hour),
		CreatedAt: now,
	}
	committedHash := "old-password-hash"
	stagedHash := ""
	tokenConsumed := false
	rolledBack := false
	auditErr := errors.New("credential audit unavailable")
	audit := &credentialAuditTxCapture{err: auditErr}
	transactor := passwordResetTransactorFunc(func(
		ctx context.Context,
		fn func(context.Context) error,
	) error {
		txCtx := context.WithValue(ctx, credentialAuditTxMarker{}, true)
		if txErr := fn(txCtx); txErr != nil {
			rolledBack = true
			stagedHash = ""
			tokenConsumed = false
			return txErr
		}
		committedHash = stagedHash
		return nil
	})
	accountRepo := &mockAccountRepository{
		findByIDForUpdateFn: func(
			ctx context.Context,
			id uint64,
		) (*model.Account, error) {
			assert.Equal(t, true, ctx.Value(credentialAuditTxMarker{}))
			return &model.Account{
				ID:        id,
				UpdatedAt: now.Add(-time.Minute),
			}, nil
		},
		updatePasswordHashFn: func(
			ctx context.Context,
			id uint64,
			newHash string,
			_ time.Time,
		) error {
			assert.Equal(t, true, ctx.Value(credentialAuditTxMarker{}))
			assert.Equal(t, accountID, id)
			stagedHash = newHash
			return nil
		},
	}
	tokenRepo := &mockPasswordResetTokenRepository{
		findByTokenHashFn: func(context.Context, string) (*model.PasswordResetToken, error) {
			return token, nil
		},
		findByTokenHashForUpdateFn: func(
			ctx context.Context,
			_ string,
		) (*model.PasswordResetToken, error) {
			assert.Equal(t, true, ctx.Value(credentialAuditTxMarker{}))
			return token, nil
		},
		consumeByIDFn: func(ctx context.Context, id uint64) error {
			assert.Equal(t, true, ctx.Value(credentialAuditTxMarker{}))
			assert.Equal(t, token.ID, id)
			tokenConsumed = true
			return nil
		},
	}
	service := NewPasswordResetServiceWithCredentialAudit(
		&PasswordResetConfig{},
		accountRepo,
		tokenRepo,
		nil,
		transactor,
		credentialAuditSubjectResolverStub{
			subject: CredentialAuditSubject{
				ClinicID: clinicID,
				StaffID:  staffID,
			},
		},
		audit,
	)
	completionService := service.(PasswordResetCompletionService)

	result, resetErr := completionService.ResetPasswordWithResult(
		context.Background(),
		rawToken,
		"newPassw0rd",
		CredentialMutationAudit{
			ActorType: model.AuditActorTypeSystem,
			Action:    model.AuditActionAuthPasswordReset,
			IPAddress: "192.0.2.2",
			UserAgent: "credential-reset-audit-test",
		},
	)

	require.Error(t, resetErr)
	assert.ErrorIs(t, resetErr, auditErr)
	assert.Nil(t, result)
	assert.True(t, rolledBack)
	assert.False(t, tokenConsumed)
	assert.Empty(t, stagedHash)
	assert.Equal(t, "old-password-hash", committedHash)
	require.NotNil(t, audit.entry.ResourceID)
	assert.Equal(t, accountID, *audit.entry.ResourceID)
	assert.Equal(t, map[string]any{"staff_id": staffID}, audit.entry.NewValue)
}

func TestPasswordResetService_ResetPasswordFailsClosedWithoutAuditBeforeMutation(
	t *testing.T,
) {
	repositoryCalled := false
	tokenRepo := &mockPasswordResetTokenRepository{
		findByTokenHashFn: func(
			context.Context,
			string,
		) (*model.PasswordResetToken, error) {
			repositoryCalled = true
			return nil, errors.New("must not be reached")
		},
	}
	service := NewPasswordResetServiceWithTransactor(
		&PasswordResetConfig{},
		&mockAccountRepository{},
		tokenRepo,
		nil,
		immediatePasswordResetTransactor(),
	)

	resetErr := service.ResetPassword(
		context.Background(),
		"raw-token",
		"newPassw0rd",
	)

	require.Error(t, resetErr)
	var appErr *apperrors.AppError
	require.ErrorAs(t, resetErr, &appErr)
	assert.Equal(t, "INTERNAL", appErr.Code)
	assert.False(t, repositoryCalled)
}
