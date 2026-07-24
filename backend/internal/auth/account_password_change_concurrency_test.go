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
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

type barrierAccountRepository struct {
	AccountRepository
	readReady chan<- struct{}
	release   <-chan struct{}
}

func (r barrierAccountRepository) FindByID(
	ctx context.Context,
	id uint64,
) (*model.Account, error) {
	account, err := r.AccountRepository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	r.readReady <- struct{}{}
	<-r.release
	return account, nil
}

func TestAccountService_ChangePasswordConcurrentSameCurrentPassword_AllowsOneSuccess(
	t *testing.T,
) {
	db := setupAccountTestDB(t)
	const currentPassword = "oldPassw0rd"
	currentHash, err := bcrypt.GenerateFromPassword(
		[]byte(currentPassword),
		bcrypt.MinCost,
	)
	require.NoError(t, err)
	account := &model.Account{
		Email:        "change-password-race@example.test",
		PasswordHash: string(currentHash),
		IsActive:     true,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(account).Error)
	previousEpoch := account.UpdatedAt

	readReady := make(chan struct{}, 2)
	releaseReads := make(chan struct{})
	accountRepo := barrierAccountRepository{
		AccountRepository: NewAccountRepository(db),
		readReady:         readReady,
		release:           releaseReads,
	}
	service := NewAccountServiceWithCredentialAudit(
		accountRepo,
		NewPasswordResetTokenRepository(db),
		persistence.NewTransactor(db),
		noopCredentialAuditTxLogger{},
	)
	changer, ok := service.(AccountPasswordChanger)
	require.True(t, ok)

	type changeResult struct {
		newPassword string
		err         error
	}
	start := make(chan struct{})
	results := make(chan changeResult, 2)
	newPasswords := []string{"winnerOne1", "winnerTwo2"}
	for _, newPassword := range newPasswords {
		go func(candidate string) {
			<-start
			results <- changeResult{
				newPassword: candidate,
				err: changer.ChangePassword(
					context.Background(),
					account.ID,
					currentPassword,
					candidate,
					testPasswordChangeAudit(17),
				),
			}
		}(newPassword)
	}
	close(start)
	<-readReady
	<-readReady
	close(releaseReads)

	successes := make([]changeResult, 0, 1)
	failures := make([]changeResult, 0, 1)
	for range newPasswords {
		result := <-results
		if result.err == nil {
			successes = append(successes, result)
			continue
		}
		failures = append(failures, result)
	}

	require.Len(t, successes, 1)
	require.Len(t, failures, 1)
	assert.ErrorIs(t, failures[0].err, apperrors.ErrUnauthorized)

	reloaded, err := NewAccountRepository(db).FindByID(
		context.Background(),
		account.ID,
	)
	require.NoError(t, err)
	assert.True(t, reloaded.UpdatedAt.After(previousEpoch))
	assert.NoError(
		t,
		bcrypt.CompareHashAndPassword(
			[]byte(reloaded.PasswordHash),
			[]byte(successes[0].newPassword),
		),
	)
	assert.True(
		t,
		errors.Is(
			bcrypt.CompareHashAndPassword(
				[]byte(reloaded.PasswordHash),
				[]byte(failures[0].newPassword),
			),
			bcrypt.ErrMismatchedHashAndPassword,
		),
	)
}

func TestAccountService_ChangePassword_InvalidatesOutstandingResetToken(
	t *testing.T,
) {
	db := setupAccountTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.PasswordResetToken{}))

	const (
		currentPassword = "oldPassw0rd"
		newPassword     = "newPassw0rd"
		rawResetToken   = "outstanding-reset-token"
	)
	currentHash, err := bcrypt.GenerateFromPassword(
		[]byte(currentPassword),
		bcrypt.MinCost,
	)
	require.NoError(t, err)
	account := &model.Account{
		Email:        "change-password-reset-invalidation@example.test",
		PasswordHash: string(currentHash),
		IsActive:     true,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(account).Error)

	accountRepo := NewAccountRepository(db)
	tokenRepo := NewPasswordResetTokenRepository(db)
	require.NoError(t, tokenRepo.Create(context.Background(), &model.PasswordResetToken{
		AccountID: account.ID,
		TokenHash: hashToken(rawResetToken),
		ExpiresAt: time.Now().Add(time.Hour),
	}))
	transactor := persistence.NewTransactor(db)
	accountService := NewAccountServiceWithCredentialAudit(
		accountRepo,
		tokenRepo,
		transactor,
		noopCredentialAuditTxLogger{},
	)
	changer, ok := accountService.(AccountPasswordChanger)
	require.True(t, ok)

	require.NoError(t, changer.ChangePassword(
		context.Background(),
		account.ID,
		currentPassword,
		newPassword,
		testPasswordChangeAudit(17),
	))

	passwordReset := NewPasswordResetServiceWithCredentialAudit(
		&PasswordResetConfig{},
		accountRepo,
		tokenRepo,
		nil,
		transactor,
		noopCredentialAuditSubjectResolver{},
		noopCredentialAuditTxLogger{},
	)
	resetErr := passwordReset.ResetPassword(
		context.Background(),
		rawResetToken,
		"anotherPassw0rd",
	)
	require.Error(t, resetErr)
	assert.True(t, apperrors.IsInvalidInput(resetErr))
}

func TestAccountService_ChangePassword_TokenInvalidationFailureRollsBackCredential(
	t *testing.T,
) {
	db := setupAccountTestDB(t)
	const (
		currentPassword = "oldPassw0rd"
		newPassword     = "newPassw0rd"
	)
	currentHash, err := bcrypt.GenerateFromPassword(
		[]byte(currentPassword),
		bcrypt.MinCost,
	)
	require.NoError(t, err)
	account := &model.Account{
		Email:        "change-password-real-rollback@example.test",
		PasswordHash: string(currentHash),
		IsActive:     true,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(account).Error)

	accountRepo := NewAccountRepository(db)
	service := NewAccountServiceWithCredentialAudit(
		accountRepo,
		accountResetTokenInvalidatorFunc(func(
			ctx context.Context,
			accountID uint64,
		) error {
			require.NotNil(t, persistence.TxFromContext(ctx))
			assert.Equal(t, account.ID, accountID)
			return errors.New("forced reset-token invalidation failure")
		}),
		persistence.NewTransactor(db),
		noopCredentialAuditTxLogger{},
	)
	changer, ok := service.(AccountPasswordChanger)
	require.True(t, ok)

	err = changer.ChangePassword(
		context.Background(),
		account.ID,
		currentPassword,
		newPassword,
		testPasswordChangeAudit(17),
	)
	require.Error(t, err)

	reloaded, reloadErr := accountRepo.FindByID(context.Background(), account.ID)
	require.NoError(t, reloadErr)
	assert.Equal(t, string(currentHash), reloaded.PasswordHash)
	assert.NoError(
		t,
		bcrypt.CompareHashAndPassword(
			[]byte(reloaded.PasswordHash),
			[]byte(currentPassword),
		),
	)
	assert.Error(
		t,
		bcrypt.CompareHashAndPassword(
			[]byte(reloaded.PasswordHash),
			[]byte(newPassword),
		),
	)
}
