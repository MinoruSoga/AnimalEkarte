package auth

// repository_test.go — Repository の統合テスト。
//
// 移動元: account_repository_test.go（BE8-4 batch23）。
//
// account の repository.go は P4 (clinicScope) の例外対象（accounts テーブルは clinic_id を
// 持たないグローバルなログインアカウント）のため、clinic_id 隔離テストは対象外。
//
// 保護する不変条件:
//   - FindByID / FindByEmail はソフトデリート済みアカウントを除外し、NotFound ラップされたエラーを返す。
//   - Create は email の一意制約違反を AlreadyExists エラーに変換する（apperrors.FromGORM の23505ハンドリング）。
//   - credential や権限を更新する汎用 map API を公開しない。

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

// setupAccountTestDB は accounts テーブルを用意し、クリーンな状態でテストを開始できるようにする。
func setupAccountTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.Account{}))
	db.Exec("TRUNCATE TABLE accounts CASCADE")
	return db
}

// makeAccount はテスト用アカウントを作成して返す。
func makeAccount(t *testing.T, db *gorm.DB, email string) *model.Account {
	t.Helper()
	a := &model.Account{
		Email:        email,
		PasswordHash: "hashed-password",
		IsActive:     true,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(a).Error)
	return a
}

func TestRepository_FindByID(t *testing.T) {
	db := setupAccountTestDB(t)
	repo := NewAccountRepository(db)
	ctx := context.Background()

	t.Run("存在するIDでは取得できる", func(t *testing.T) {
		a := makeAccount(t, db, "find-by-id@example.test")
		got, err := repo.FindByID(ctx, a.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, a.Email, got.Email)
	})

	t.Run("存在しないIDはNotFoundを返す", func(t *testing.T) {
		got, err := repo.FindByID(ctx, uint64(9999999))
		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("ソフトデリート済みアカウントはNotFoundを返す", func(t *testing.T) {
		a := makeAccount(t, db, "soft-deleted@example.test")
		require.NoError(t, db.WithContext(ctx).Delete(a).Error)

		got, err := repo.FindByID(ctx, a.ID)
		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestAccountRepository_FindByIDForUpdateRequiresAmbientTransaction(t *testing.T) {
	repo := NewAccountRepository(nil)

	account, err := repo.FindByIDForUpdate(context.Background(), 1)

	assert.Nil(t, account)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ambient transaction")
}

func TestAccountRepository_CompareAndSwapPasswordHashRequiresAmbientTransaction(
	t *testing.T,
) {
	repo := NewAccountRepository(nil)

	swapped, err := repo.CompareAndSwapPasswordHash(
		context.Background(),
		1,
		"old-hash",
		"new-hash",
		time.Now(),
	)

	assert.False(t, swapped)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ambient transaction")
}

func TestRepository_FindByEmail(t *testing.T) {
	db := setupAccountTestDB(t)
	repo := NewAccountRepository(db)
	ctx := context.Background()

	t.Run("存在するメールアドレスでは取得できる", func(t *testing.T) {
		a := makeAccount(t, db, "find-by-email@example.test")
		got, err := repo.FindByEmail(ctx, a.Email)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, a.ID, got.ID)
	})

	t.Run("存在しないメールアドレスはNotFoundを返す", func(t *testing.T) {
		const privateEmail = "nonexistent@example.test"
		got, err := repo.FindByEmail(ctx, privateEmail)
		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
		assert.NotContains(t, err.Error(), privateEmail)
	})

	t.Run("ソフトデリート済みアカウントのメールアドレスはNotFoundを返す", func(t *testing.T) {
		a := makeAccount(t, db, "soft-deleted-email@example.test")
		require.NoError(t, db.WithContext(ctx).Delete(a).Error)

		got, err := repo.FindByEmail(ctx, a.Email)
		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
		assert.NotContains(t, err.Error(), a.Email)
	})
}

func TestRepository_Create(t *testing.T) {
	db := setupAccountTestDB(t)
	repo := NewAccountRepository(db)
	ctx := context.Background()

	t.Run("新規アカウントを作成できる", func(t *testing.T) {
		a := &model.Account{
			Email:         "create-new@example.test",
			PasswordHash:  "hashed",
			IsActive:      true,
			IsSystemAdmin: false,
		}
		require.NoError(t, repo.Create(ctx, a))
		assert.NotZero(t, a.ID)

		var stored model.Account
		require.NoError(t, db.First(&stored, a.ID).Error)
		assert.Equal(t, "create-new@example.test", stored.Email)
	})

	t.Run("既存のメールアドレスとの重複はAlreadyExistsエラーを返す", func(t *testing.T) {
		existing := makeAccount(t, db, "duplicate@example.test")
		dup := &model.Account{
			Email:        existing.Email,
			PasswordHash: "another-hash",
			IsActive:     true,
		}
		err := repo.Create(ctx, dup)
		require.Error(t, err)
		assert.True(t, apperrors.IsAlreadyExists(err))
	})
}
