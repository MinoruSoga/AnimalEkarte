package auth

import (
	"bytes"
	"context"
	"errors"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

func TestAccountRepository_Create_SuppressesCredentialSQLLogging(
	t *testing.T,
) {
	var logOutput bytes.Buffer
	captureLogger := logger.New(
		log.New(&logOutput, "", 0),
		logger.Config{
			LogLevel:             logger.Info,
			ParameterizedQueries: false,
		},
	)
	db, err := gorm.Open(
		postgres.Open("host=localhost user=test dbname=test sslmode=disable"),
		&gorm.Config{
			DisableAutomaticPing:   true,
			DryRun:                 true,
			Logger:                 captureLogger,
			SkipDefaultTransaction: true,
		},
	)
	require.NoError(t, err)

	forcedFailure := errors.New("forced account create failure")
	require.NoError(t, db.Callback().Create().After("gorm:create").Register(
		"test:force_account_create_failure",
		func(tx *gorm.DB) {
			_ = tx.AddError(forcedFailure)
		},
	))

	const controlHash = "control-create-password-hash-must-be-visible"
	controlResult := db.Create(&model.Account{
		Email:        "control-create@example.test",
		PasswordHash: controlHash,
		IsActive:     true,
	})
	require.ErrorIs(t, controlResult.Error, forcedFailure)
	require.Contains(
		t,
		logOutput.String(),
		controlHash,
		"capture logger must expose interpolated values when it is not silenced",
	)

	logOutput.Reset()
	const credentialHash = "$2a$12$create-credential-hash-must-never-reach-logs"
	repository := NewAccountRepository(db)
	err = repository.Create(context.Background(), &model.Account{
		Email:        "credential-create@example.test",
		PasswordHash: credentialHash,
		IsActive:     true,
	})

	require.ErrorIs(t, err, forcedFailure)
	assert.NotContains(t, logOutput.String(), credentialHash)
	assert.Empty(t, logOutput.String(), "credential create must silence the SQL trace entirely")
}

func TestAccountRepository_UpdatePasswordHash_SuppressesCredentialSQLLogging(
	t *testing.T,
) {
	var logOutput bytes.Buffer
	captureLogger := logger.New(
		log.New(&logOutput, "", 0),
		logger.Config{
			LogLevel:             logger.Info,
			ParameterizedQueries: false,
		},
	)
	db, err := gorm.Open(
		postgres.Open("host=localhost user=test dbname=test sslmode=disable"),
		&gorm.Config{
			DisableAutomaticPing:   true,
			DryRun:                 true,
			Logger:                 captureLogger,
			SkipDefaultTransaction: true,
		},
	)
	require.NoError(t, err)

	forcedFailure := errors.New("forced password update failure")
	require.NoError(t, db.Callback().Update().After("gorm:update").Register(
		"test:force_password_update_failure",
		func(tx *gorm.DB) {
			_ = tx.AddError(forcedFailure)
		},
	))

	const controlHash = "control-password-hash-must-be-visible"
	controlResult := db.Model(&model.Account{}).
		Where("id = ?", uint64(9)).
		Updates(map[string]any{"password_hash": controlHash})
	require.ErrorIs(t, controlResult.Error, forcedFailure)
	require.Contains(
		t,
		logOutput.String(),
		controlHash,
		"capture logger must expose interpolated values when it is not silenced",
	)

	logOutput.Reset()
	const credentialHash = "$2a$12$credential-hash-must-never-reach-logs"
	repository := NewAccountRepository(db)
	updater, ok := repository.(PasswordCredentialUpdater)
	require.True(t, ok)
	ctx := persistence.WithTxValue(context.Background(), db)

	err = updater.UpdatePasswordHash(
		ctx,
		9,
		credentialHash,
		time.Now(),
	)

	require.ErrorIs(t, err, forcedFailure)
	assert.NotContains(t, logOutput.String(), credentialHash)
	assert.Empty(t, logOutput.String(), "credential update must silence the SQL trace entirely")
}

func TestPasswordResetTokenRepository_DeleteIssuedSuppressesTokenHashLogging(
	t *testing.T,
) {
	var logOutput bytes.Buffer
	captureLogger := logger.New(
		log.New(&logOutput, "", 0),
		logger.Config{
			LogLevel:             logger.Info,
			ParameterizedQueries: false,
		},
	)
	db, err := gorm.Open(
		postgres.Open("host=localhost user=test dbname=test sslmode=disable"),
		&gorm.Config{
			DisableAutomaticPing:   true,
			DryRun:                 true,
			Logger:                 captureLogger,
			SkipDefaultTransaction: true,
		},
	)
	require.NoError(t, err)

	forcedFailure := errors.New("forced reset-token cleanup failure")
	require.NoError(t, db.Callback().Delete().After("gorm:delete").Register(
		"test:force_reset_token_cleanup_failure",
		func(tx *gorm.DB) {
			_ = tx.AddError(forcedFailure)
		},
	))

	const tokenHash = "reset-token-hash-must-never-reach-logs"
	err = NewPasswordResetTokenRepository(db).DeleteIssued(
		context.Background(),
		9,
		tokenHash,
	)

	require.ErrorIs(t, err, forcedFailure)
	assert.NotContains(t, logOutput.String(), tokenHash)
	assert.Empty(t, logOutput.String(), "reset token cleanup must silence the SQL trace entirely")
}

func TestPasswordResetTokenRepository_SuppressesTokenHashLogging(
	t *testing.T,
) {
	var logOutput bytes.Buffer
	captureLogger := logger.New(
		log.New(&logOutput, "", 0),
		logger.Config{
			LogLevel:             logger.Info,
			ParameterizedQueries: false,
		},
	)
	db, err := gorm.Open(
		postgres.Open("host=localhost user=test dbname=test sslmode=disable"),
		&gorm.Config{
			DisableAutomaticPing:   true,
			DryRun:                 true,
			Logger:                 captureLogger,
			SkipDefaultTransaction: true,
		},
	)
	require.NoError(t, err)
	forcedFailure := errors.New("forced reset-token persistence failure")
	require.NoError(t, db.Callback().Create().After("gorm:create").Register(
		"test:force_reset_token_create_failure",
		func(tx *gorm.DB) {
			_ = tx.AddError(forcedFailure)
		},
	))
	require.NoError(t, db.Callback().Query().After("gorm:query").Register(
		"test:force_reset_token_query_failure",
		func(tx *gorm.DB) {
			_ = tx.AddError(forcedFailure)
		},
	))

	const tokenHash = "reset-token-hash-must-never-reach-logs"
	repository := NewPasswordResetTokenRepository(db)
	tests := []struct {
		name string
		call func() error
	}{
		{
			name: "create",
			call: func() error {
				return repository.Create(context.Background(), &model.PasswordResetToken{
					AccountID: 1,
					TokenHash: tokenHash,
					ExpiresAt: time.Now().Add(time.Hour),
				})
			},
		},
		{
			name: "lookup",
			call: func() error {
				_, lookupErr := repository.FindByTokenHash(
					context.Background(),
					tokenHash,
				)
				return lookupErr
			},
		},
		{
			name: "locked lookup",
			call: func() error {
				ctx := persistence.WithTxValue(context.Background(), db)
				_, lookupErr := repository.FindByTokenHashForUpdate(
					ctx,
					tokenHash,
				)
				return lookupErr
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logOutput.Reset()

			err := test.call()

			require.Error(t, err)
			assert.NotContains(t, err.Error(), tokenHash)
			assert.NotContains(t, logOutput.String(), tokenHash)
			assert.Empty(t, logOutput.String())
		})
	}
}

func TestTokenBlacklistRepository_SuppressesTokenIdentifierLogging(
	t *testing.T,
) {
	var logOutput bytes.Buffer
	captureLogger := logger.New(
		log.New(&logOutput, "", 0),
		logger.Config{
			LogLevel:             logger.Info,
			ParameterizedQueries: false,
		},
	)
	db, err := gorm.Open(
		postgres.Open("host=localhost user=test dbname=test sslmode=disable"),
		&gorm.Config{
			DisableAutomaticPing:   true,
			DryRun:                 true,
			Logger:                 captureLogger,
			SkipDefaultTransaction: true,
		},
	)
	require.NoError(t, err)
	forcedFailure := errors.New("forced token-blacklist persistence failure")
	require.NoError(t, db.Callback().Create().After("gorm:create").Register(
		"test:force_token_blacklist_create_failure",
		func(tx *gorm.DB) {
			_ = tx.AddError(forcedFailure)
		},
	))
	require.NoError(t, db.Callback().Query().After("gorm:query").Register(
		"test:force_token_blacklist_query_failure",
		func(tx *gorm.DB) {
			_ = tx.AddError(forcedFailure)
		},
	))

	const identifier = "refresh-family:identifier-must-never-reach-logs"
	repository := NewTokenBlacklistRepository(db)
	tests := []struct {
		name string
		call func() error
	}{
		{
			name: "create",
			call: func() error {
				return repository.Create(context.Background(), &model.TokenBlacklist{
					JTI:       identifier,
					ExpiresAt: time.Now().Add(time.Hour),
				})
			},
		},
		{
			name: "lookup",
			call: func() error {
				_, lookupErr := repository.ExistsByJTI(
					context.Background(),
					identifier,
				)
				return lookupErr
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logOutput.Reset()

			err := test.call()

			require.Error(t, err)
			assert.NotContains(t, err.Error(), identifier)
			assert.NotContains(t, logOutput.String(), identifier)
			assert.Empty(t, logOutput.String())
		})
	}
}

func TestAccountRepository_UpdatePasswordHash_RequiresAmbientTransaction(
	t *testing.T,
) {
	repository := NewAccountRepository(nil)
	updater, ok := repository.(PasswordCredentialUpdater)
	require.True(t, ok)

	err := updater.UpdatePasswordHash(
		context.Background(),
		1,
		"credential-hash",
		time.Now(),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ambient transaction")
}

func TestAccountRepository_UpdatePasswordHash_RequiresExactlyOneAffectedRow(
	t *testing.T,
) {
	db, err := gorm.Open(
		postgres.Open("host=localhost user=test dbname=test sslmode=disable"),
		&gorm.Config{
			DisableAutomaticPing:   true,
			DryRun:                 true,
			SkipDefaultTransaction: true,
		},
	)
	require.NoError(t, err)
	repository := NewAccountRepository(db)
	ctx := persistence.WithTxValue(context.Background(), db)

	err = repository.UpdatePasswordHash(
		ctx,
		9,
		"$2a$12$credential-hash",
		time.Now(),
	)

	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}
