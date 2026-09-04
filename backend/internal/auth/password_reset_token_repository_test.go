package auth

// password_reset_token_repository_test.go — M-8 認証トークン・ライフサイクル回帰テスト
//
// テスト対象: PasswordResetTokenRepository
// 保護する不変条件:
//   - 有効なトークンはハッシュで取得できる。
//   - 使用済み（= DeleteByID で物理削除）トークンは再取得できない（単回使用）。
//   - 存在しないトークンハッシュは NotFound を返す。
//   - DeleteByAccountID は該当アカウントの既存トークンのみを一括削除する。
//
// 実装の意味論に関する重要な注意（テストで明示的に固定する）:
//   本 repository の FindByTokenHash は expires_at を一切検証しない。
//   期限判定は service 層の責務であり、期限切れトークンでも repository は返す。
//   単回使用は used_at フラグではなく DeleteByID による物理削除で担保される
//   （「使用済みマーク」メソッドは存在しない）。

import (
	"context"
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

// setupPasswordResetTokenTestDB は password_reset_tokens テーブルを用意し、
// クリーンな状態でテストを開始できるようにする。
func setupPasswordResetTokenTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.PasswordResetToken{}))
	testdb.Truncate(t, db, "password_reset_tokens")
	return db
}

// makePasswordResetToken は指定アカウント・ハッシュ・有効期限でトークンを作成する。
func makePasswordResetToken(t *testing.T, repo PasswordResetTokenRepository, accountID uint64, hash string, expiresAt time.Time) *model.PasswordResetToken {
	t.Helper()
	token := &model.PasswordResetToken{
		AccountID: accountID,
		TokenHash: hash,
		ExpiresAt: expiresAt,
	}
	require.NoError(t, repo.Create(context.Background(), token))
	require.NotZero(t, token.ID, "Create 後は自動採番された ID が入る")
	return token
}

// TestPasswordResetTokenRepository_FindByTokenHash は
// ハッシュによるトークン取得の挙動を検証する。
func TestPasswordResetTokenRepository_FindByTokenHash(t *testing.T) {
	db := setupPasswordResetTokenTestDB(t)
	repo := NewPasswordResetTokenRepository(db)
	ctx := context.Background()

	const accountID = uint64(1)

	t.Run("有効期限内のトークンをハッシュで取得できる", func(t *testing.T) {
		created := makePasswordResetToken(t, repo, accountID, "valid-hash-001", time.Now().Add(1*time.Hour))

		got, err := repo.FindByTokenHash(ctx, "valid-hash-001")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, created.ID, got.ID)
		assert.Equal(t, accountID, got.AccountID)
		assert.Equal(t, "valid-hash-001", got.TokenHash)
	})

	t.Run("存在しないトークンハッシュでは NotFound を返す", func(t *testing.T) {
		got, err := repo.FindByTokenHash(ctx, "does-not-exist")
		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	// 意味論の固定: repository は期限を検証しない。期限判定は service 層の責務。
	// この不変条件が壊れる（repository が期限フィルタを追加する）と本テストが失敗し、
	// 期限判定ロジックの移動をレビューで検知できる。
	t.Run("期限切れトークンも repository は取得できる（期限判定は呼び出し側の責務）", func(t *testing.T) {
		makePasswordResetToken(t, repo, accountID, "expired-hash-001", time.Now().Add(-1*time.Hour))

		got, err := repo.FindByTokenHash(ctx, "expired-hash-001")
		require.NoError(t, err, "repository は expires_at を検証しないため取得できる")
		require.NotNil(t, got)
		assert.True(t, got.ExpiresAt.Before(time.Now()), "取得したトークンは期限切れ（呼び出し側が拒否すべき）")
	})
}

// TestPasswordResetTokenRepository_DeleteByID_SingleUse は
// 単回使用（使用後 DeleteByID で物理削除 → 再取得不可）を検証する。
func TestPasswordResetTokenRepository_DeleteByID_SingleUse(t *testing.T) {
	db := setupPasswordResetTokenTestDB(t)
	repo := NewPasswordResetTokenRepository(db)
	ctx := context.Background()

	const accountID = uint64(1)
	created := makePasswordResetToken(t, repo, accountID, "single-use-hash", time.Now().Add(1*time.Hour))

	t.Run("使用前はハッシュで取得できる", func(t *testing.T) {
		got, err := repo.FindByTokenHash(ctx, "single-use-hash")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, created.ID, got.ID)
	})

	t.Run("DeleteByID による使用後は再取得できない（単回使用）", func(t *testing.T) {
		require.NoError(t, repo.DeleteByID(ctx, created.ID))

		got, err := repo.FindByTokenHash(ctx, "single-use-hash")
		require.Error(t, err, "削除済みトークンは再利用できてはならない")
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})
}

// TestPasswordResetTokenRepository_DeleteByAccountID は
// トークン発行前クリーンアップ（該当アカウントの既存トークン一括削除）を検証する。
func TestPasswordResetTokenRepository_DeleteByAccountID(t *testing.T) {
	db := setupPasswordResetTokenTestDB(t)
	repo := NewPasswordResetTokenRepository(db)
	ctx := context.Background()

	const (
		targetAccountID = uint64(1)
		otherAccountID  = uint64(2)
	)

	// 対象アカウントに 2 件、別アカウントに 1 件のトークンを作成する。
	makePasswordResetToken(t, repo, targetAccountID, "target-hash-1", time.Now().Add(1*time.Hour))
	makePasswordResetToken(t, repo, targetAccountID, "target-hash-2", time.Now().Add(1*time.Hour))
	makePasswordResetToken(t, repo, otherAccountID, "other-hash-1", time.Now().Add(1*time.Hour))

	require.NoError(t, persistence.NewTransactor(db).WithTx(
		ctx,
		func(txCtx context.Context) error {
			return repo.DeleteByAccountID(txCtx, targetAccountID)
		},
	))

	t.Run("対象アカウントの既存トークンはすべて削除される", func(t *testing.T) {
		for _, hash := range []string{"target-hash-1", "target-hash-2"} {
			got, err := repo.FindByTokenHash(ctx, hash)
			require.Error(t, err, "%s は削除されているべき", hash)
			assert.Nil(t, got)
			assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
		}
	})

	t.Run("別アカウントのトークンは削除されない（アカウント単位スコープ）", func(t *testing.T) {
		got, err := repo.FindByTokenHash(ctx, "other-hash-1")
		require.NoError(t, err, "別アカウントのトークンは巻き込み削除されてはならない")
		require.NotNil(t, got)
		assert.Equal(t, otherAccountID, got.AccountID)
	})
}

func TestPasswordResetTokenRepository_DeleteByAccountIDRequiresAmbientTransaction(
	t *testing.T,
) {
	repo := NewPasswordResetTokenRepository(nil)

	err := repo.DeleteByAccountID(context.Background(), 1)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ambient transaction")
}

func TestPasswordResetTokenRepository_FindLatestByAccountIDForUpdate(
	t *testing.T,
) {
	db := setupPasswordResetTokenTestDB(t)
	repo := NewPasswordResetTokenRepository(db)
	ctx := context.Background()
	const accountID = uint64(1)

	older := &model.PasswordResetToken{
		AccountID: accountID,
		TokenHash: "latest-lock-older",
		ExpiresAt: time.Now().Add(time.Hour),
		CreatedAt: time.Now().Add(-time.Minute),
	}
	newer := &model.PasswordResetToken{
		AccountID: accountID,
		TokenHash: "latest-lock-newer",
		ExpiresAt: time.Now().Add(time.Hour),
		CreatedAt: time.Now(),
	}
	require.NoError(t, repo.Create(ctx, older))
	require.NoError(t, repo.Create(ctx, newer))

	var got *model.PasswordResetToken
	require.NoError(t, persistence.NewTransactor(db).WithTx(
		ctx,
		func(txCtx context.Context) error {
			var err error
			got, err = repo.FindLatestByAccountIDForUpdate(txCtx, accountID)
			return err
		},
	))

	require.NotNil(t, got)
	assert.Equal(t, newer.ID, got.ID)
}

// TestPasswordResetTokenRepository_Create_DuplicateTokenHash は token_hash に
// uniqueIndex が張られているため、同一ハッシュの重複作成がエラーになることを検証する
// （apperrors.FromGORM の NotFound 以外の分岐を経由する）。
func TestPasswordResetTokenRepository_Create_DuplicateTokenHash(t *testing.T) {
	db := setupPasswordResetTokenTestDB(t)
	repo := NewPasswordResetTokenRepository(db)
	ctx := context.Background()

	makePasswordResetToken(t, repo, uint64(1), "dup-hash", time.Now().Add(1*time.Hour))

	err := repo.Create(ctx, &model.PasswordResetToken{
		AccountID: uint64(2),
		TokenHash: "dup-hash",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	})
	require.Error(t, err, "token_hash の重複は作成エラーになるべき")
}

// TestPasswordResetTokenRepository_DeleteByID_NonexistentID_NoError は、
// 存在しない ID に対する DeleteByID が RowsAffected を検証しないため
// エラーを返さない（no-op）ことを固定する。
func TestPasswordResetTokenRepository_DeleteByID_NonexistentID_NoError(t *testing.T) {
	db := setupPasswordResetTokenTestDB(t)
	repo := NewPasswordResetTokenRepository(db)
	ctx := context.Background()

	err := repo.DeleteByID(ctx, 9999999)
	require.NoError(t, err, "DeleteByID は RowsAffected==0 でもエラーを返さない実装")
}
