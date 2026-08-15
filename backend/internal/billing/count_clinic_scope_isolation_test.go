package billing

// count_clinic_scope_isolation_test.go — BE-refactor.md R2-5 (D12) の回帰防止。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func setupCountEstimateIsolationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.Estimate{}, &model.EstimateItem{}))
	db.Exec("TRUNCATE TABLE estimates CASCADE")
	return db
}

func TestEstimateRepository_CountItemsByEstimateID_ClinicIsolation(t *testing.T) {
	db := setupCountEstimateIsolationTestDB(t)
	repo := NewEstimateRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	est := &model.Estimate{ClinicID: clinicA}
	require.NoError(t, db.WithContext(ctx).Create(est).Error)
	item := &model.EstimateItem{EstimateID: est.ID, Name: "テスト項目", Category: model.ItemCategoryOther}
	require.NoError(t, db.WithContext(ctx).Create(item).Error)

	t.Run("同一クリニックIDでは件数が見える", func(t *testing.T) {
		count, err := repo.CountItemsByEstimateID(ctx, clinicA, est.ID)
		require.NoError(t, err)
		require.Equal(t, int64(1), count)
	})

	t.Run("別クリニックIDでは0件を返す（JOIN述語がないと漏洩しうる）", func(t *testing.T) {
		count, err := repo.CountItemsByEstimateID(ctx, clinicB, est.ID)
		require.NoError(t, err)
		require.Equal(t, int64(0), count)
	})
}
