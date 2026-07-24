package lstep

// line_link_token_repository_test.go — LineLinkTokenRepository 統合テスト（BE-021）。
//
// 保護する不変条件:
//   - FindByToken は未使用・未失効のトークンのみ返す。
//   - 失効済み・使用済み・存在しないトークンは NotFound を返す。
//   - MarkUsed で used_at を設定すると、以後 FindByToken は該当トークンを返さなくなる。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupLineLinkTokenTestDB は line_link_tokens テーブルを整備する。
func setupLineLinkTokenTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.LineLinkToken{}))
	db.Exec("TRUNCATE TABLE line_link_tokens CASCADE")
	return db
}

func makeLineLinkToken(t *testing.T, db *gorm.DB, clinicID, ownerID uint64, token string, expiresAt time.Time, usedAt *time.Time) *model.LineLinkToken {
	t.Helper()
	tk := &model.LineLinkToken{
		ClinicID:  clinicID,
		OwnerID:   ownerID,
		Token:     token,
		ExpiresAt: expiresAt,
		UsedAt:    usedAt,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(tk).Error)
	return tk
}

func TestLineLinkTokenRepository_Create(t *testing.T) {
	db := setupLineLinkTokenTestDB(t)
	repo := NewLineLinkTokenRepository(db)
	ctx := context.Background()

	tk := &model.LineLinkToken{
		ClinicID:  1,
		OwnerID:   10,
		Token:     "create-token",
		ExpiresAt: time.Now().Add(time.Hour),
	}
	require.NoError(t, repo.Create(ctx, tk))
	assert.NotZero(t, tk.ID)

	var count int64
	require.NoError(t, db.Model(&model.LineLinkToken{}).Where("token = ?", "create-token").Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

func TestLineLinkTokenRepository_FindByToken(t *testing.T) {
	db := setupLineLinkTokenTestDB(t)
	repo := NewLineLinkTokenRepository(db)
	ctx := context.Background()

	now := time.Now()
	makeLineLinkToken(t, db, 1, 10, "valid-token", now.Add(time.Hour), nil)
	makeLineLinkToken(t, db, 1, 10, "expired-token", now.Add(-time.Hour), nil)
	usedAt := now.Add(-time.Minute)
	makeLineLinkToken(t, db, 1, 10, "used-token", now.Add(time.Hour), &usedAt)

	t.Run("finds unused, unexpired token", func(t *testing.T) {
		got, err := repo.FindByToken(ctx, "valid-token")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "valid-token", got.Token)
	})

	t.Run("expired token is not found", func(t *testing.T) {
		got, err := repo.FindByToken(ctx, "expired-token")
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("used token is not found", func(t *testing.T) {
		got, err := repo.FindByToken(ctx, "used-token")
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("unknown token is not found", func(t *testing.T) {
		got, err := repo.FindByToken(ctx, "does-not-exist")
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestLineLinkTokenRepository_MarkUsed(t *testing.T) {
	db := setupLineLinkTokenTestDB(t)
	repo := NewLineLinkTokenRepository(db)
	ctx := context.Background()

	tk := makeLineLinkToken(t, db, 1, 10, "to-be-used", time.Now().Add(time.Hour), nil)

	usedAt := time.Now()
	require.NoError(t, repo.MarkUsed(ctx, tk.ID, usedAt))

	t.Run("persists used_at", func(t *testing.T) {
		var fetched model.LineLinkToken
		require.NoError(t, db.First(&fetched, tk.ID).Error)
		require.NotNil(t, fetched.UsedAt)
		assert.WithinDuration(t, usedAt, *fetched.UsedAt, time.Second)
	})

	t.Run("token is no longer findable after being used", func(t *testing.T) {
		got, err := repo.FindByToken(ctx, "to-be-used")
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})
}
