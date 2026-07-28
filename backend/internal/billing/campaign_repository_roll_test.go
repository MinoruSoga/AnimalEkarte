package billing

// repository_test.go — CampaignRepository の統合テスト（内部カバレッジ向上）。
//
// 対象: FindAll / FindByID / Create / Update / ReplaceTargets / Delete / Reorder /
//       FindApplicableForItem / FindAllApplicableForItem
// 検証観点: 正常系、clinic_id 隔離、ソフトデリート除外、NotFound ラップ、
//           適用可能キャンペーン判定（期間・is_active・カテゴリ/個別商品・discount_value 優先度）。

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

// setupCampaignTestDB は campaigns / campaign_target_categories / campaign_target_items を整備する。
// campaign_discount_type ENUM は testdb.SetupTestDB の共通リストに含まれないため、ここで個別に作成する。
func setupCampaignTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, db.Exec(`
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'campaign_discount_type') THEN
        CREATE TYPE campaign_discount_type AS ENUM ('rate', 'amount');
    END IF;
END
$$;`).Error)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.Campaign{}, &model.CampaignTargetCategory{}, &model.CampaignTargetItem{}))
	db.Exec("TRUNCATE TABLE campaign_target_items CASCADE")
	db.Exec("TRUNCATE TABLE campaign_target_categories CASCADE")
	db.Exec("TRUNCATE TABLE campaigns CASCADE")
	return db
}

func TestCampaignRepository_FindAll_ClinicIsolationAndSortOrder(t *testing.T) {
	db := setupCampaignTestDB(t)
	repo := NewCampaignRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	jun := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	jul := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)

	other := &model.Campaign{ClinicID: clinicB, Name: "医院Bのキャンペーン", StartDate: jun, EndDate: jul, SortOrder: 1}
	require.NoError(t, db.WithContext(ctx).Create(other).Error)
	second := &model.Campaign{ClinicID: clinicA, Name: "Bキャンペーン", StartDate: jun, EndDate: jul, SortOrder: 2}
	require.NoError(t, db.WithContext(ctx).Create(second).Error)
	first := &model.Campaign{ClinicID: clinicA, Name: "Aキャンペーン", StartDate: jun, EndDate: jul, SortOrder: 1}
	require.NoError(t, db.WithContext(ctx).Create(first).Error)

	got, err := repo.FindAll(ctx, clinicA)
	require.NoError(t, err)
	require.Len(t, got, 2, "clinic A のキャンペーンのみ返る")
	assert.Equal(t, first.ID, got[0].ID, "sort_order 昇順で先頭に来る")
	assert.Equal(t, second.ID, got[1].ID)
	for _, c := range got {
		assert.NotEqual(t, other.ID, c.ID, "別クリニックのキャンペーンが混入してはならない")
	}
}

func TestCampaignRepository_FindByID(t *testing.T) {
	db := setupCampaignTestDB(t)
	repo := NewCampaignRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	jun := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	jul := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)

	c := &model.Campaign{ClinicID: clinicA, Name: "夏の割引", StartDate: jun, EndDate: jul}
	require.NoError(t, db.WithContext(ctx).Create(c).Error)

	t.Run("同一クリニックで取得できる", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, c.ID)
		require.NoError(t, err)
		assert.Equal(t, "夏の割引", got.Name)
	})

	t.Run("別クリニックからは NotFound", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicB, c.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しない ID は NotFound", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestCampaignRepository_Create(t *testing.T) {
	db := setupCampaignTestDB(t)
	repo := NewCampaignRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)
	jun := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	jul := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)

	c := &model.Campaign{
		ClinicID:      clinicA,
		Name:          "新規キャンペーン",
		StartDate:     jun,
		EndDate:       jul,
		DiscountType:  model.CampaignDiscountTypeRate,
		DiscountValue: 10,
		TargetCategories: []model.CampaignTargetCategory{
			{Category: model.ItemCategoryVaccine},
		},
		TargetItems: []model.CampaignTargetItem{
			{MerchandiseItemID: 555},
		},
	}
	created, err := repo.Create(ctx, c)
	require.NoError(t, err)
	assert.NotZero(t, created.ID)

	got, err := repo.FindByID(ctx, clinicA, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "新規キャンペーン", got.Name)
	require.Len(t, got.TargetCategories, 1, "対象カテゴリが同時作成される")
	assert.Equal(t, model.ItemCategoryVaccine, got.TargetCategories[0].Category)
	require.Len(t, got.TargetItems, 1, "対象商品が同時作成される")
	assert.Equal(t, uint64(555), got.TargetItems[0].MerchandiseItemID)
}

// TestCampaignRepository_Create_IsActiveFalsePersists is BUG-455-S3:
// gorm:"default:true" omits zero bools from INSERT; explicit false must survive Create.
func TestCampaignRepository_Create_IsActiveFalsePersists(t *testing.T) {
	db := setupCampaignTestDB(t)
	repo := NewCampaignRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	jun := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	jul := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)

	inactive := &model.Campaign{
		ClinicID:      clinicID,
		Name:          "create inactive campaign",
		StartDate:     jun,
		EndDate:       jul,
		DiscountType:  model.CampaignDiscountTypeRate,
		DiscountValue: 10,
		IsActive:      false,
	}
	created, err := repo.Create(ctx, inactive)
	require.NoError(t, err)
	require.NotZero(t, created.ID)
	assert.False(t, inactive.IsActive, "in-memory input must keep false after Create")
	assert.False(t, created.IsActive, "Create return value must keep explicit false")

	got, err := repo.FindByID(ctx, clinicID, created.ID)
	require.NoError(t, err)
	assert.False(t, got.IsActive, "DB read-back must keep explicit false")

	var rawActive bool
	require.NoError(t, db.WithContext(ctx).
		Model(&model.Campaign{}).
		Select("is_active").
		Where("id = ?", created.ID).
		Scan(&rawActive).Error)
	assert.False(t, rawActive, "raw is_active column must be false")
}

// TestCampaignRepository_Create_IsActiveTruePersists is true-path regression for BUG-455-S3.
func TestCampaignRepository_Create_IsActiveTruePersists(t *testing.T) {
	db := setupCampaignTestDB(t)
	repo := NewCampaignRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	jun := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	jul := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)

	active := &model.Campaign{
		ClinicID:      clinicID,
		Name:          "create active true campaign",
		StartDate:     jun,
		EndDate:       jul,
		DiscountType:  model.CampaignDiscountTypeRate,
		DiscountValue: 10,
		IsActive:      true,
	}
	created, err := repo.Create(ctx, active)
	require.NoError(t, err)
	assert.True(t, active.IsActive)
	assert.True(t, created.IsActive)

	got, err := repo.FindByID(ctx, clinicID, created.ID)
	require.NoError(t, err)
	assert.True(t, got.IsActive)

	var rawActive bool
	require.NoError(t, db.WithContext(ctx).
		Model(&model.Campaign{}).
		Select("is_active").
		Where("id = ?", created.ID).
		Scan(&rawActive).Error)
	assert.True(t, rawActive)
}

func TestCampaignRepository_Update(t *testing.T) {
	db := setupCampaignTestDB(t)
	repo := NewCampaignRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	jun := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	jul := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)

	c := &model.Campaign{ClinicID: clinicA, Name: "旧名称", StartDate: jun, EndDate: jul}
	require.NoError(t, db.WithContext(ctx).Create(c).Error)

	t.Run("同一クリニックで更新できる", func(t *testing.T) {
		got, err := repo.Update(ctx, clinicA, c.ID, map[string]any{"name": "新名称"})
		require.NoError(t, err)
		assert.Equal(t, "新名称", got.Name)
	})

	t.Run("別クリニックからの更新は NotFound", func(t *testing.T) {
		_, err := repo.Update(ctx, clinicB, c.ID, map[string]any{"name": "乗っ取り"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しない ID の更新は NotFound", func(t *testing.T) {
		_, err := repo.Update(ctx, clinicA, 999999, map[string]any{"name": "x"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestCampaignRepository_ReplaceTargets(t *testing.T) {
	db := setupCampaignTestDB(t)
	repo := NewCampaignRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)
	jun := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	jul := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)

	c := &model.Campaign{ClinicID: clinicA, Name: "対象差し替え", StartDate: jun, EndDate: jul}
	require.NoError(t, db.WithContext(ctx).Create(c).Error)

	require.NoError(t, repo.ReplaceTargets(ctx, c.ID,
		[]model.ItemCategory{model.ItemCategoryVaccine, model.ItemCategoryTrimming},
		[]uint64{100, 200}))

	got, err := repo.FindByID(ctx, clinicA, c.ID)
	require.NoError(t, err)
	assert.Len(t, got.TargetCategories, 2)
	assert.Len(t, got.TargetItems, 2)

	t.Run("再度差し替えると旧対象が消え新対象のみ残る", func(t *testing.T) {
		require.NoError(t, repo.ReplaceTargets(ctx, c.ID, []model.ItemCategory{model.ItemCategoryGoods}, nil))

		got2, err := repo.FindByID(ctx, clinicA, c.ID)
		require.NoError(t, err)
		require.Len(t, got2.TargetCategories, 1)
		assert.Equal(t, model.ItemCategoryGoods, got2.TargetCategories[0].Category)
		assert.Empty(t, got2.TargetItems, "空スライスへの差し替えで対象商品が全削除される")
	})
}

func TestCampaignRepository_Delete(t *testing.T) {
	db := setupCampaignTestDB(t)
	repo := NewCampaignRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	jun := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	jul := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)

	c := &model.Campaign{ClinicID: clinicA, Name: "削除対象", StartDate: jun, EndDate: jul}
	require.NoError(t, db.WithContext(ctx).Create(c).Error)

	t.Run("別クリニックからの削除は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, c.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しない ID の削除は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("同一クリニックで削除でき、ソフトデリートされ FindAll から除外される", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, c.ID))

		_, err := repo.FindByID(ctx, clinicA, c.ID)
		assert.True(t, apperrors.IsNotFound(err))

		all, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		for _, x := range all {
			assert.NotEqual(t, c.ID, x.ID)
		}

		var raw model.Campaign
		require.NoError(t, db.WithContext(ctx).Unscoped().Where("id = ?", c.ID).First(&raw).Error)
		assert.True(t, raw.DeletedAt.Valid, "deleted_at が設定されているべき")
	})
}

func TestCampaignRepository_Reorder(t *testing.T) {
	db := setupCampaignTestDB(t)
	repo := NewCampaignRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	jun := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	jul := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)

	c1 := &model.Campaign{ClinicID: clinicA, Name: "C1", StartDate: jun, EndDate: jul}
	require.NoError(t, db.WithContext(ctx).Create(c1).Error)
	c2 := &model.Campaign{ClinicID: clinicA, Name: "C2", StartDate: jun, EndDate: jul}
	require.NoError(t, db.WithContext(ctx).Create(c2).Error)
	c3 := &model.Campaign{ClinicID: clinicA, Name: "C3", StartDate: jun, EndDate: jul}
	require.NoError(t, db.WithContext(ctx).Create(c3).Error)

	t.Run("指定順に sort_order が振り直される", func(t *testing.T) {
		require.NoError(t, repo.Reorder(ctx, clinicA, []uint64{c3.ID, c1.ID, c2.ID}))

		got, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		require.Len(t, got, 3)
		assert.Equal(t, c3.ID, got[0].ID)
		assert.Equal(t, c1.ID, got[1].ID)
		assert.Equal(t, c2.ID, got[2].ID)
	})

	t.Run("別クリニックの ID を含むと失敗する", func(t *testing.T) {
		other := &model.Campaign{ClinicID: clinicB, Name: "他院", StartDate: jun, EndDate: jul}
		require.NoError(t, db.WithContext(ctx).Create(other).Error)
		err := repo.Reorder(ctx, clinicA, []uint64{c1.ID, other.ID})
		require.Error(t, err)
	})
}

func TestCampaignRepository_FindApplicableForItem(t *testing.T) {
	db := setupCampaignTestDB(t)
	repo := NewCampaignRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	inRange := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	jun := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	jul := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)
	future := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	futureEnd := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)

	t.Run("該当キャンペーンが無ければ nil,nil", func(t *testing.T) {
		got, err := repo.FindApplicableForItem(ctx, clinicA, inRange, model.ItemCategoryVaccine, nil)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	low := &model.Campaign{ClinicID: clinicA, Name: "低割引", StartDate: jun, EndDate: jul, IsActive: true, DiscountValue: 5}
	require.NoError(t, db.WithContext(ctx).Create(low).Error)
	require.NoError(t, repo.ReplaceTargets(ctx, low.ID, []model.ItemCategory{model.ItemCategoryVaccine}, nil))

	high := &model.Campaign{ClinicID: clinicA, Name: "高割引", StartDate: jun, EndDate: jul, IsActive: true, DiscountValue: 20}
	require.NoError(t, db.WithContext(ctx).Create(high).Error)
	require.NoError(t, repo.ReplaceTargets(ctx, high.ID, []model.ItemCategory{model.ItemCategoryVaccine}, nil))

	// BUG-455-S3: repository Create compensates gorm default:true so IsActive:false
	// persists without a separate post-Create Update workaround.
	inactive := &model.Campaign{ClinicID: clinicA, Name: "無効", StartDate: jun, EndDate: jul, IsActive: false, DiscountValue: 99}
	_, err := repo.Create(ctx, inactive)
	require.NoError(t, err)
	require.NoError(t, repo.ReplaceTargets(ctx, inactive.ID, []model.ItemCategory{model.ItemCategoryVaccine}, nil))

	outOfRange := &model.Campaign{ClinicID: clinicA, Name: "期間外", StartDate: future, EndDate: futureEnd, IsActive: true, DiscountValue: 50}
	require.NoError(t, db.WithContext(ctx).Create(outOfRange).Error)
	require.NoError(t, repo.ReplaceTargets(ctx, outOfRange.ID, []model.ItemCategory{model.ItemCategoryVaccine}, nil))

	otherClinic := &model.Campaign{ClinicID: clinicB, Name: "別医院", StartDate: jun, EndDate: jul, IsActive: true, DiscountValue: 90}
	require.NoError(t, db.WithContext(ctx).Create(otherClinic).Error)
	require.NoError(t, repo.ReplaceTargets(ctx, otherClinic.ID, []model.ItemCategory{model.ItemCategoryVaccine}, nil))

	t.Run("discount_value が最大のキャンペーンを返す（無効・期間外・他院は除外）", func(t *testing.T) {
		got, err := repo.FindApplicableForItem(ctx, clinicA, inRange, model.ItemCategoryVaccine, nil)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, high.ID, got.ID)
	})

	t.Run("個別商品指定でカテゴリ不一致でもヒットする", func(t *testing.T) {
		itemOnly := &model.Campaign{ClinicID: clinicA, Name: "商品限定", StartDate: jun, EndDate: jul, IsActive: true, DiscountValue: 30}
		require.NoError(t, db.WithContext(ctx).Create(itemOnly).Error)
		require.NoError(t, repo.ReplaceTargets(ctx, itemOnly.ID, nil, []uint64{777}))

		itemID := uint64(777)
		got, err := repo.FindApplicableForItem(ctx, clinicA, inRange, model.ItemCategoryOther, &itemID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, itemOnly.ID, got.ID)
	})
}

func TestCampaignRepository_FindAllApplicableForItem(t *testing.T) {
	db := setupCampaignTestDB(t)
	repo := NewCampaignRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)
	inRange := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	jun := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	jul := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)

	t.Run("該当なしは空スライス", func(t *testing.T) {
		got, err := repo.FindAllApplicableForItem(ctx, clinicA, inRange, model.ItemCategoryTrimming, nil)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	low := &model.Campaign{ClinicID: clinicA, Name: "低割引", StartDate: jun, EndDate: jul, IsActive: true, DiscountValue: 5}
	require.NoError(t, db.WithContext(ctx).Create(low).Error)
	require.NoError(t, repo.ReplaceTargets(ctx, low.ID, []model.ItemCategory{model.ItemCategoryTrimming}, nil))

	high := &model.Campaign{ClinicID: clinicA, Name: "高割引", StartDate: jun, EndDate: jul, IsActive: true, DiscountValue: 20}
	require.NoError(t, db.WithContext(ctx).Create(high).Error)
	require.NoError(t, repo.ReplaceTargets(ctx, high.ID, []model.ItemCategory{model.ItemCategoryTrimming}, nil))

	got, err := repo.FindAllApplicableForItem(ctx, clinicA, inRange, model.ItemCategoryTrimming, nil)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, high.ID, got[0].ID, "discount_value 降順で先頭に高割引が来る")
	assert.Equal(t, low.ID, got[1].ID)
}
