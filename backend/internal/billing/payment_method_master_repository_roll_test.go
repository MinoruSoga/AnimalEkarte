package billing

// repository_test.go — BE8-3: repotest 抽出後の最初のパッケージローカル DB 統合テスト雛形。
// 先行 14 サブパッケージのローカルテストが 0 件だった構造的原因（setupTestDB が _test.go 内定義で
// import 不能）は repotest への抽出で解消済み。以降の新規サブパッケージはこの形を先例とする。

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

func setupRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.PaymentMethodMaster{}))
	db.Exec("TRUNCATE TABLE payment_methods CASCADE")
	return db
}

func TestPaymentMethodMasterRepository_Create_FindByID(t *testing.T) {
	db := setupRepoTestDB(t)
	repo := NewPaymentMethodMasterRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(9001), uint64(9002)

	t.Run("作成した支払方法を同一クリニックで取得できる", func(t *testing.T) {
		m := &model.PaymentMethodMaster{ClinicID: clinicA, Name: "現金(repotest雛形)", IsActive: true}
		created, err := repo.Create(ctx, m)
		require.NoError(t, err)
		require.NotZero(t, created.ID)

		got, err := repo.FindByID(ctx, clinicA, created.ID)
		require.NoError(t, err)
		assert.Equal(t, "現金(repotest雛形)", got.Name)
	})

	t.Run("別クリニックからはNotFound", func(t *testing.T) {
		m := &model.PaymentMethodMaster{ClinicID: clinicA, Name: "クレジットカード(repotest雛形)", IsActive: true}
		created, err := repo.Create(ctx, m)
		require.NoError(t, err)

		_, err = repo.FindByID(ctx, clinicB, created.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}
