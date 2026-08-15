package lstep

// line_link_token_repository_test.go — LineLinkTokenRepository integration tests.
//
// These tests require the shared PostgreSQL test database. They are authored here
// but executed only by the root run while holding the repository-wide DB lease.

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func setupLineLinkTokenTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.LineLinkToken{}))
	require.NoError(t, db.Exec("TRUNCATE TABLE line_link_tokens CASCADE").Error)
	return db
}

func makeLineLinkToken(
	t *testing.T,
	db *gorm.DB,
	clinicID, ownerID uint64,
	token string,
	expiresAt time.Time,
	usedAt *time.Time,
) *model.LineLinkToken {
	t.Helper()
	linkToken := &model.LineLinkToken{
		ClinicID:  clinicID,
		OwnerID:   ownerID,
		Token:     token,
		ExpiresAt: expiresAt,
		UsedAt:    usedAt,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(linkToken).Error)
	return linkToken
}

func TestLineLinkTokenRepository_Create(t *testing.T) {
	db := setupLineLinkTokenTestDB(t)
	repo := NewLineLinkTokenRepository(db)
	owner := testdb.MakeTestOwner(t, db, 1, "LINE token owner")
	token := &model.LineLinkToken{
		ClinicID:  1,
		OwnerID:   owner.ID,
		Token:     digestLineLinkToken("raw-token"),
		ExpiresAt: time.Now().Add(time.Hour),
	}

	require.NoError(t, repo.Create(context.Background(), token))
	assert.NotZero(t, token.ID)

	var persisted model.LineLinkToken
	require.NoError(t, db.First(&persisted, token.ID).Error)
	assert.Equal(t, digestLineLinkToken("raw-token"), persisted.Token)
	assert.Len(t, persisted.Token, 43)
}

func TestLineLinkTokenRepository_CreateParticipatesInAmbientTransaction(t *testing.T) {
	db := setupLineLinkTokenTestDB(t)
	repo := NewLineLinkTokenRepository(db)
	owner := testdb.MakeTestOwner(t, db, 1, "LINE token rollback owner")
	rollbackErr := errors.New("force rollback")
	token := &model.LineLinkToken{
		ClinicID:  owner.ClinicID,
		OwnerID:   owner.ID,
		Token:     digestLineLinkToken("rollback-token"),
		ExpiresAt: time.Now().Add(time.Hour),
	}

	err := persistence.NewTransactor(db).WithTx(context.Background(), func(ctx context.Context) error {
		require.NoError(t, repo.Create(ctx, token))
		return rollbackErr
	})

	require.ErrorIs(t, err, rollbackErr)
	var count int64
	require.NoError(t, db.Model(&model.LineLinkToken{}).Where("token = ?", token.Token).Count(&count).Error)
	assert.Zero(t, count)
}

func TestLineLinkTokenRepository_LockUsableByRawToken(t *testing.T) {
	db := setupLineLinkTokenTestDB(t)
	repo := NewLineLinkTokenRepository(db)
	transactor := persistence.NewTransactor(db)
	now := time.Now()
	rawToken := "new-base64url-token"
	owner := testdb.MakeTestOwner(t, db, 1, "LINE digest owner")
	makeLineLinkToken(t, db, 1, owner.ID, digestLineLinkToken(rawToken), now.Add(time.Hour), nil)
	makeLineLinkToken(t, db, 1, owner.ID, digestLineLinkToken("expired-token"), now.Add(-time.Hour), nil)
	usedAt := now.Add(-time.Minute)
	makeLineLinkToken(t, db, 1, owner.ID, digestLineLinkToken("used-token"), now.Add(time.Hour), &usedAt)

	t.Run("finds the digest without exposing the raw token", func(t *testing.T) {
		require.NoError(t, transactor.WithTx(context.Background(), func(ctx context.Context) error {
			got, err := repo.LockUsableByRawToken(ctx, rawToken, now)
			require.NoError(t, err)
			assert.Equal(t, digestLineLinkToken(rawToken), got.Token)
			assert.NotEqual(t, rawToken, got.Token)
			return nil
		}))
	})

	for _, candidate := range []string{"expired-token", "used-token", "unknown-token"} {
		t.Run(candidate+" is rejected", func(t *testing.T) {
			err := transactor.WithTx(context.Background(), func(ctx context.Context) error {
				got, err := repo.LockUsableByRawToken(ctx, candidate, now)
				assert.Nil(t, got)
				return err
			})
			require.Error(t, err)
			assert.True(t, apperrors.IsNotFound(err))
		})
	}
}

func TestLineLinkTokenRepository_LegacyFallbackIsLimitedTo64HexTokens(t *testing.T) {
	db := setupLineLinkTokenTestDB(t)
	repo := NewLineLinkTokenRepository(db)
	transactor := persistence.NewTransactor(db)
	now := time.Now()
	legacyRawToken := strings.Repeat("ab", 32)
	legacyOwner := testdb.MakeTestOwner(t, db, 1, "legacy LINE token owner")
	plaintextOwner := testdb.MakeTestOwner(t, db, 1, "plaintext LINE token owner")
	makeLineLinkToken(t, db, 1, legacyOwner.ID, legacyRawToken, now.Add(time.Hour), nil)
	makeLineLinkToken(t, db, 1, plaintextOwner.ID, "legacy-plaintext-token", now.Add(time.Hour), nil)

	require.NoError(t, transactor.WithTx(context.Background(), func(ctx context.Context) error {
		got, err := repo.LockUsableByRawToken(ctx, legacyRawToken, now)
		require.NoError(t, err)
		assert.Equal(t, legacyRawToken, got.Token)
		return nil
	}))

	err := transactor.WithTx(context.Background(), func(ctx context.Context) error {
		got, err := repo.LockUsableByRawToken(ctx, "legacy-plaintext-token", now)
		assert.Nil(t, got)
		return err
	})
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestLineLinkTokenRepository_ConsumeIsSingleUseCAS(t *testing.T) {
	db := setupLineLinkTokenTestDB(t)
	repo := NewLineLinkTokenRepository(db)
	transactor := persistence.NewTransactor(db)
	now := time.Now()
	owner := testdb.MakeTestOwner(t, db, 1, "single-use LINE token owner")
	token := makeLineLinkToken(t, db, 1, owner.ID, digestLineLinkToken("single-use"), now.Add(time.Hour), nil)

	require.NoError(t, transactor.WithTx(context.Background(), func(ctx context.Context) error {
		require.NoError(t, repo.Consume(ctx, token.ID, now))
		err := repo.Consume(ctx, token.ID, now.Add(time.Second))
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
		return nil
	}))

	var persisted model.LineLinkToken
	require.NoError(t, db.First(&persisted, token.ID).Error)
	require.NotNil(t, persisted.UsedAt)
	assert.WithinDuration(t, now, *persisted.UsedAt, time.Second)
}

func TestLineLinkTokenRepository_ConcurrentDoubleConsumeOnlyOneSucceeds(t *testing.T) {
	db := setupLineLinkTokenTestDB(t)
	repo := NewLineLinkTokenRepository(db)
	now := time.Now()
	rawToken := "concurrent-single-use"
	owner := testdb.MakeTestOwner(t, db, 1, "concurrent LINE token owner")
	makeLineLinkToken(t, db, 1, owner.ID, digestLineLinkToken(rawToken), now.Add(time.Hour), nil)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	start := make(chan struct{})
	var successes atomic.Int32
	var waitGroup sync.WaitGroup
	for i := 0; i < 2; i++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-start
			err := persistence.NewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
				token, err := repo.LockUsableByRawToken(txCtx, rawToken, now)
				if err != nil {
					return err
				}
				return repo.Consume(txCtx, token.ID, now)
			})
			if err == nil {
				successes.Add(1)
			}
		}()
	}
	close(start)
	done := make(chan struct{})
	go func() {
		waitGroup.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-ctx.Done():
		t.Fatal("concurrent token consume did not terminate")
	}

	assert.Equal(t, int32(1), successes.Load())
}

func TestLineLinkTokenRepository_TransactionRequiredForLockAndConsume(t *testing.T) {
	db := setupLineLinkTokenTestDB(t)
	repo := NewLineLinkTokenRepository(db)

	_, lockErr := repo.LockUsableByRawToken(context.Background(), "token", time.Now())
	consumeErr := repo.Consume(context.Background(), 1, time.Now())

	require.Error(t, lockErr)
	require.Error(t, consumeErr)
}
