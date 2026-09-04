package auth

// token_blacklist_repository_test.go — M-8 認証トークン・ライフサイクル回帰テスト
//
// テスト対象: TokenBlacklistRepository
// 保護する不変条件:
//   - 登録済みかつ未失効の JTI は「存在（= 失効済み）」を返す。
//   - 未登録の JTI は「存在しない（= 有効）」を返す。
//   - 期限切れの JTI は存在しないものとして扱う（ExistsByJTI は expires_at > now でフィルタ）。
//   - DeleteExpired は期限切れエントリのみを物理削除し、有効エントリは残す。
//
// この挙動が壊れると、ログアウト済み refresh_token の JTI 照合が
// すり抜ける（失効が効かない）ため、認証セキュリティ上の回帰を検知する。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupTokenBlacklistTestDB は token_blacklist テーブルを用意し、
// クリーンな状態でテストを開始できるようにする。
func setupTokenBlacklistTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.TokenBlacklist{}))
	testdb.Truncate(t, db, "token_blacklist")
	return db
}

// makeBlacklistEntry は指定 JTI・有効期限でブラックリストエントリを作成する。
func makeBlacklistEntry(t *testing.T, repo TokenBlacklistRepository, jti string, expiresAt time.Time) {
	t.Helper()
	entry := &model.TokenBlacklist{
		JTI:       jti,
		ExpiresAt: expiresAt,
	}
	require.NoError(t, repo.Create(context.Background(), entry))
}

// TestTokenBlacklistRepository_ExistsByJTI は
// JTI 失効照会の挙動を検証する。
func TestTokenBlacklistRepository_ExistsByJTI(t *testing.T) {
	db := setupTokenBlacklistTestDB(t)
	repo := NewTokenBlacklistRepository(db)
	ctx := context.Background()

	t.Run("登録済みかつ未失効の JTI は存在（失効済み）を返す", func(t *testing.T) {
		makeBlacklistEntry(t, repo, "revoked-jti", time.Now().Add(1*time.Hour))

		exists, err := repo.ExistsByJTI(ctx, "revoked-jti")
		require.NoError(t, err)
		assert.True(t, exists, "ブラックリスト登録済みの JTI は失効扱いであるべき")
	})

	t.Run("未登録の JTI は存在しない（有効）を返す", func(t *testing.T) {
		exists, err := repo.ExistsByJTI(ctx, "never-registered-jti")
		require.NoError(t, err)
		assert.False(t, exists, "未登録の JTI は失効していないべき")
	})

	// 意味論の固定: expires_at を過ぎたエントリは存在しないものとして扱う。
	// これは DeleteExpired によるメンテ前でも失効照会が誤って残らないための設計。
	t.Run("期限切れの JTI は存在しないものとして扱う", func(t *testing.T) {
		makeBlacklistEntry(t, repo, "expired-jti", time.Now().Add(-1*time.Hour))

		exists, err := repo.ExistsByJTI(ctx, "expired-jti")
		require.NoError(t, err)
		assert.False(t, exists, "期限切れエントリは存在しない扱い（expires_at > now フィルタ）であるべき")
	})
}

// TestTokenBlacklistRepository_DeleteExpired は
// 期限切れエントリのクリーンアップ（有効エントリは残す）を検証する。
func TestTokenBlacklistRepository_DeleteExpired(t *testing.T) {
	db := setupTokenBlacklistTestDB(t)
	repo := NewTokenBlacklistRepository(db)
	ctx := context.Background()

	makeBlacklistEntry(t, repo, "expired-1", time.Now().Add(-2*time.Hour))
	makeBlacklistEntry(t, repo, "expired-2", time.Now().Add(-1*time.Hour))
	makeBlacklistEntry(t, repo, "valid-1", time.Now().Add(1*time.Hour))

	require.NoError(t, repo.DeleteExpired(ctx))

	// ExistsByJTI は expires_at でフィルタするため物理削除の確認には使えない。
	// テーブルを直接 Count して物理削除を検証する。
	countByJTI := func(jti string) int64 {
		var n int64
		require.NoError(t, db.Model(&model.TokenBlacklist{}).Where("jti = ?", jti).Count(&n).Error)
		return n
	}

	t.Run("期限切れエントリは物理削除される", func(t *testing.T) {
		assert.Equal(t, int64(0), countByJTI("expired-1"), "expired-1 は削除されているべき")
		assert.Equal(t, int64(0), countByJTI("expired-2"), "expired-2 は削除されているべき")
	})

	t.Run("有効エントリは残る", func(t *testing.T) {
		assert.Equal(t, int64(1), countByJTI("valid-1"), "有効エントリは削除されてはならない")

		exists, err := repo.ExistsByJTI(ctx, "valid-1")
		require.NoError(t, err)
		assert.True(t, exists, "有効エントリは失効照会でも存在扱いであるべき")
	})
}

// TestTokenBlacklistRepository_Create_DuplicateJTI は jti が主キーであるため、
// 同一 JTI の重複登録が GORM エラーとしてラップされて返ることを検証する
// （apperrors.FromGORM の NotFound 以外の分岐を経由する）。
func TestTokenBlacklistRepository_Create_DuplicateJTI(t *testing.T) {
	db := setupTokenBlacklistTestDB(t)
	repo := NewTokenBlacklistRepository(db)
	ctx := context.Background()

	makeBlacklistEntry(t, repo, "duplicate-jti", time.Now().Add(1*time.Hour))

	err := repo.Create(ctx, &model.TokenBlacklist{JTI: "duplicate-jti", ExpiresAt: time.Now().Add(2 * time.Hour)})
	require.Error(t, err, "主キー重複は作成エラーになるべき")
}
