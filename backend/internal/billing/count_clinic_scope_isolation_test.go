package billing

// count_clinic_scope_isolation_test.go — BE-refactor.md R2-5 (D12) の回帰防止。

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
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

func TestEstimateRepository_ReplaceItems_ClinicIsolation(t *testing.T) {
	db := setupCountEstimateIsolationTestDB(t)
	repo := NewEstimateRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	est := &model.Estimate{ClinicID: clinicA, Title: "本院見積"}
	require.NoError(t, db.WithContext(ctx).Create(est).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.EstimateItem{
		EstimateID: est.ID, Name: "既存明細", Category: model.ItemCategoryOther,
	}).Error)

	err := testNewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
		return repo.ReplaceItems(txCtx, clinicB, est.ID, []model.EstimateItem{{
			Name: "他院からの書込", Category: model.ItemCategoryOther,
		}})
	})
	require.Error(t, err)

	count, err := repo.CountItemsByEstimateID(ctx, clinicA, est.ID)
	require.NoError(t, err)
	require.Equal(t, int64(1), count, "他院 ReplaceItems は本院明細を変えてはならない")
}

func TestEstimateRepository_ReplaceItems_RequiresTransaction(t *testing.T) {
	db := setupCountEstimateIsolationTestDB(t)
	repo := NewEstimateRepository(db)
	est := &model.Estimate{ClinicID: 1, Title: "tx必須"}
	require.NoError(t, db.Create(est).Error)

	err := repo.ReplaceItems(context.Background(), 1, est.ID, nil)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr), "ambient tx なしは fail-closed: %v", err)
	require.Equal(t, "INTERNAL", appErr.Code)
}
