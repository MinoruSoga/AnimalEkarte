package auth

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

type passwordResetGORMTransactor struct {
	db *gorm.DB
}

func (t passwordResetGORMTransactor) WithTx(
	ctx context.Context,
	fn func(context.Context) error,
) error {
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(persistence.WithTxValue(ctx, tx))
	})
}

type synchronizedPasswordResetTokenRepository struct {
	PasswordResetTokenRepository
	waiting atomic.Int32
	release chan struct{}
}

type synchronizedForgotPasswordAccountRepository struct {
	AccountRepository
	waiting atomic.Int32
	release chan struct{}
}

type passwordResetCrossOperationBarrier struct {
	waiting atomic.Int32
	release chan struct{}
}

func (b *passwordResetCrossOperationBarrier) wait() {
	if b.waiting.Add(1) == 2 {
		close(b.release)
	}
	<-b.release
}

type crossOperationAccountRepository struct {
	AccountRepository
	barrier *passwordResetCrossOperationBarrier
}

func (r *crossOperationAccountRepository) FindByEmail(
	ctx context.Context,
	email string,
) (*model.Account, error) {
	account, err := r.AccountRepository.FindByEmail(ctx, email)
	r.barrier.wait()
	return account, err
}

func (r *crossOperationAccountRepository) UpdatePasswordHash(
	ctx context.Context,
	accountID uint64,
	newHash string,
	updatedAt time.Time,
) error {
	return r.AccountRepository.(PasswordCredentialUpdater).
		UpdatePasswordHash(ctx, accountID, newHash, updatedAt)
}

type crossOperationTokenRepository struct {
	PasswordResetTokenRepository
	barrier *passwordResetCrossOperationBarrier
}

func (r *crossOperationTokenRepository) FindByTokenHash(
	ctx context.Context,
	hash string,
) (*model.PasswordResetToken, error) {
	token, err := r.PasswordResetTokenRepository.FindByTokenHash(ctx, hash)
	r.barrier.wait()
	return token, err
}

func (r *synchronizedForgotPasswordAccountRepository) FindByEmail(
	ctx context.Context,
	email string,
) (*model.Account, error) {
	account, err := r.AccountRepository.FindByEmail(ctx, email)
	if r.waiting.Add(1) == 2 {
		close(r.release)
	}
	<-r.release
	return account, err
}

func (r *synchronizedPasswordResetTokenRepository) FindByTokenHash(
	ctx context.Context,
	hash string,
) (*model.PasswordResetToken, error) {
	token, err := r.PasswordResetTokenRepository.FindByTokenHash(ctx, hash)
	if r.waiting.Add(1) == 2 {
		close(r.release)
	}
	<-r.release
	return token, err
}

func TestPasswordResetService_ResetPassword_ConcurrentSameTokenOnlyOneSucceeds(t *testing.T) {
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.Account{},
		&model.PasswordResetToken{},
	))
	ctx := context.Background()

	accountRepo := NewAccountRepository(db)
	tokenRepo := NewPasswordResetTokenRepository(db)
	account := &model.Account{
		Email:        fmt.Sprintf("password-reset-concurrency-%d@example.com", time.Now().UnixNano()),
		PasswordHash: "old-password-hash",
	}
	require.NoError(t, accountRepo.Create(ctx, account))

	rawToken := fmt.Sprintf("same-token-%d", time.Now().UnixNano())
	resetToken := &model.PasswordResetToken{
		AccountID: account.ID,
		TokenHash: hashToken(rawToken),
		ExpiresAt: time.Now().Add(time.Hour),
	}
	require.NoError(t, tokenRepo.Create(ctx, resetToken))

	synchronizedRepo := &synchronizedPasswordResetTokenRepository{
		PasswordResetTokenRepository: tokenRepo,
		release:                      make(chan struct{}),
	}
	svc := NewPasswordResetServiceWithCredentialAudit(
		&PasswordResetConfig{},
		accountRepo,
		synchronizedRepo,
		nil,
		passwordResetGORMTransactor{db: db},
		noopCredentialAuditSubjectResolver{},
		noopCredentialAuditTxLogger{},
	)

	type resetResult struct {
		password string
		err      error
	}
	results := make(chan resetResult, 2)
	for _, password := range []string{"first-new-password", "second-new-password"} {
		go func(password string) {
			results <- resetResult{
				password: password,
				err:      svc.ResetPassword(ctx, rawToken, password),
			}
		}(password)
	}

	first := <-results
	second := <-results
	close(results)

	successes := make([]resetResult, 0, 1)
	failures := make([]resetResult, 0, 1)
	for _, result := range []resetResult{first, second} {
		if result.err == nil {
			successes = append(successes, result)
		} else {
			failures = append(failures, result)
		}
	}
	require.Len(t, successes, 1)
	require.Len(t, failures, 1)
	assert.True(t, apperrors.IsInvalidInput(failures[0].err), "replayed token must be rejected: %v", failures[0].err)

	reloadedAccount, err := accountRepo.FindByID(ctx, account.ID)
	require.NoError(t, err)
	require.NoError(t, bcrypt.CompareHashAndPassword(
		[]byte(reloadedAccount.PasswordHash),
		[]byte(successes[0].password),
	))

	consumed, err := tokenRepo.FindByTokenHash(ctx, hashToken(rawToken))
	assert.Nil(t, consumed)
	assert.True(t, apperrors.IsNotFound(err), "successful reset must consume the token: %v", err)
}

func TestPasswordResetService_ForgotPassword_ConcurrentRequestsLeaveOneLiveToken(
	t *testing.T,
) {
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.Account{},
		&model.PasswordResetToken{},
	))
	ctx := context.Background()

	baseAccountRepo := NewAccountRepository(db)
	account := &model.Account{
		Email:        fmt.Sprintf("forgot-password-concurrency-%d@example.com", time.Now().UnixNano()),
		PasswordHash: "current-password-hash",
		IsActive:     true,
	}
	require.NoError(t, baseAccountRepo.Create(ctx, account))

	accountRepo := &synchronizedForgotPasswordAccountRepository{
		AccountRepository: baseAccountRepo,
		release:           make(chan struct{}),
	}
	tokenRepo := NewPasswordResetTokenRepository(db)
	svc := NewPasswordResetServiceWithCredentialAudit(
		&PasswordResetConfig{},
		accountRepo,
		tokenRepo,
		nil,
		passwordResetGORMTransactor{db: db},
		noopCredentialAuditSubjectResolver{},
		noopCredentialAuditTxLogger{},
	)

	results := make(chan error, 2)
	for range 2 {
		go func() {
			results <- svc.ForgotPassword(ctx, account.Email)
		}()
	}

	require.NoError(t, <-results)
	require.NoError(t, <-results)
	close(results)

	var liveTokenCount int64
	require.NoError(t, db.WithContext(ctx).
		Model(&model.PasswordResetToken{}).
		Where("account_id = ?", account.ID).
		Count(&liveTokenCount).Error)
	assert.Equal(t, int64(1), liveTokenCount)
}

func TestPasswordResetService_ForgotAndResetUseSameAccountFirstLockOrder(
	t *testing.T,
) {
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.Account{},
		&model.PasswordResetToken{},
	))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	baseAccountRepo := NewAccountRepository(db)
	baseTokenRepo := NewPasswordResetTokenRepository(db)
	oldHash, err := bcrypt.GenerateFromPassword(
		[]byte("old-password-1"),
		bcrypt.MinCost,
	)
	require.NoError(t, err)
	account := &model.Account{
		Email: fmt.Sprintf(
			"forgot-reset-cross-operation-%d@example.com",
			time.Now().UnixNano(),
		),
		PasswordHash: string(oldHash),
		IsActive:     true,
	}
	require.NoError(t, baseAccountRepo.Create(ctx, account))

	rawToken := fmt.Sprintf("cross-operation-token-%d", time.Now().UnixNano())
	require.NoError(t, baseTokenRepo.Create(ctx, &model.PasswordResetToken{
		AccountID: account.ID,
		TokenHash: hashToken(rawToken),
		ExpiresAt: time.Now().Add(time.Hour),
	}))

	barrier := &passwordResetCrossOperationBarrier{release: make(chan struct{})}
	accountRepo := &crossOperationAccountRepository{
		AccountRepository: baseAccountRepo,
		barrier:           barrier,
	}
	tokenRepo := &crossOperationTokenRepository{
		PasswordResetTokenRepository: baseTokenRepo,
		barrier:                      barrier,
	}
	svc := NewPasswordResetServiceWithCredentialAudit(
		&PasswordResetConfig{},
		accountRepo,
		tokenRepo,
		nil,
		passwordResetGORMTransactor{db: db},
		noopCredentialAuditSubjectResolver{},
		noopCredentialAuditTxLogger{},
	)
	defer svc.Wait()

	forgotResult := make(chan error, 1)
	resetResult := make(chan error, 1)
	go func() {
		forgotResult <- svc.ForgotPassword(ctx, account.Email)
	}()
	go func() {
		resetResult <- svc.ResetPassword(ctx, rawToken, "new-password-2")
	}()

	var forgotErr error
	select {
	case forgotErr = <-forgotResult:
	case <-ctx.Done():
		t.Fatal("forgot password did not complete; possible cross-operation deadlock")
	}
	var resetErr error
	select {
	case resetErr = <-resetResult:
	case <-ctx.Done():
		t.Fatal("reset password did not complete; possible cross-operation deadlock")
	}

	require.NoError(t, forgotErr)
	if resetErr != nil {
		assert.True(
			t,
			apperrors.IsInvalidInput(resetErr),
			"the only valid losing reset outcome is token invalidation: %v",
			resetErr,
		)
	}

	var liveTokenCount int64
	require.NoError(t, db.WithContext(ctx).
		Model(&model.PasswordResetToken{}).
		Where("account_id = ?", account.ID).
		Count(&liveTokenCount).Error)
	assert.LessOrEqual(
		t,
		liveTokenCount,
		int64(1),
		"account-first serialization must never leave multiple live reset tokens",
	)

	reloaded, err := baseAccountRepo.FindByID(ctx, account.ID)
	require.NoError(t, err)
	if resetErr == nil {
		require.NoError(t, bcrypt.CompareHashAndPassword(
			[]byte(reloaded.PasswordHash),
			[]byte("new-password-2"),
		))
	} else {
		assert.Equal(t, string(oldHash), reloaded.PasswordHash)
	}
}
