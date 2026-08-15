package clinic

// company_repository_test.go — 法人情報 data access の統合テスト（内部カバレッジ向上）。
//
// 対象: FindSingleton / Update
// 検証観点: 正常系（id=1 固定シングルトン）、レコード不在時の NotFound。
//
// 実装上の注意（テストで明示的に固定する）:
//   company は P4 clinicScope 例外テーブルであり、FindSingleton/Update はいずれも
//   常に id=1 を対象とする（マルチテナントではなく法人単位のグローバル設定）。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupCompanyTestDB は companies テーブルを整備する（依存する他テーブルなし）。
func setupCompanyTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.Company{}))
	db.Exec("TRUNCATE TABLE companies CASCADE")
	return db
}

// TestCompanyRepository_FindSingleton はシングルトン取得の正常系・NotFound を検証する。
func TestCompanyRepository_FindSingleton(t *testing.T) {
	db := setupCompanyTestDB(t)
	repo := NewCompanyRepository(db)
	ctx := context.Background()

	t.Run("レコードが無ければ NotFound を返す", func(t *testing.T) {
		_, err := repo.FindSingleton(ctx)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("id=1 のレコードを返す", func(t *testing.T) {
		require.NoError(t, db.WithContext(ctx).Create(&model.Company{ID: 1, Name: "ノア動物病院グループ"}).Error)

		got, err := repo.FindSingleton(ctx)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, uint64(1), got.ID)
		assert.Equal(t, "ノア動物病院グループ", got.Name)
	})
}

// TestCompanyRepository_Update は部分更新の正常系・NotFound を検証する。
func TestCompanyRepository_Update(t *testing.T) {
	db := setupCompanyTestDB(t)
	repo := NewCompanyRepository(db)
	ctx := context.Background()

	t.Run("レコードが無ければ NotFound を返す", func(t *testing.T) {
		err := repo.Update(ctx, map[string]any{"name": "存在しない法人"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("正常系: name/phone_number が更新される", func(t *testing.T) {
		require.NoError(t, db.WithContext(ctx).Create(&model.Company{ID: 1, Name: "更新前法人"}).Error)

		require.NoError(t, repo.Update(ctx, map[string]any{
			"name":         "更新後法人",
			"phone_number": "03-1234-5678",
		}))

		got, err := repo.FindSingleton(ctx)
		require.NoError(t, err)
		assert.Equal(t, "更新後法人", got.Name)
		assert.Equal(t, "03-1234-5678", got.PhoneNumber)
	})
}
