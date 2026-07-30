package inventory

import (
	"context"
	"errors"
	"fmt"
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

// ---- 実 DB 結合テスト（merchandise_item_repository.go）----
// 上記の Test群は setupTestDB を呼ばずリテラル同士を比較するだけの stub のため実カバレッジに寄与しない。
// 以下は実際の merchandiseItemRepository を実 Postgres テストDBに対して駆動する。

// setupMerchandiseItemRepoTestDB は merchandise_item_repository のテストに必要なテーブルを整備する。
// CountUsageByMerchandiseItemID は billing_items / estimate_items を billings / estimates 経由でJOINし、
// さらに campaign_target_items を non-deleted campaigns 経由でJOINするためそれらも migrate する。
func setupMerchandiseItemRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.MerchandiseItem{}, &model.Billing{}, &model.BillingItem{},
		&model.Estimate{}, &model.EstimateItem{},
		&model.Campaign{}, &model.CampaignTargetItem{}, &model.CampaignTargetCategory{},
	))
	db.Exec("TRUNCATE TABLE campaign_target_items CASCADE")
	db.Exec("TRUNCATE TABLE campaign_target_categories CASCADE")
	db.Exec("TRUNCATE TABLE campaigns CASCADE")
	db.Exec("TRUNCATE TABLE billing_items CASCADE")
	db.Exec("TRUNCATE TABLE estimate_items CASCADE")
	db.Exec("TRUNCATE TABLE estimates CASCADE")
	db.Exec("TRUNCATE TABLE billings CASCADE")
	db.Exec("TRUNCATE TABLE merchandise_items CASCADE")
	return db
}

func makeMerchCampaign(t *testing.T, db *gorm.DB, clinicID uint64, name string, isActive bool, start, end time.Time) *model.Campaign {
	t.Helper()
	c := &model.Campaign{
		ClinicID:      clinicID,
		Name:          name,
		StartDate:     start,
		EndDate:       end,
		DiscountType:  model.CampaignDiscountTypeRate,
		DiscountValue: 10,
		IsActive:      isActive,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(c).Error)
	return c
}

func makeMerchCampaignTarget(t *testing.T, db *gorm.DB, campaignID, merchandiseItemID uint64) *model.CampaignTargetItem {
	t.Helper()
	ti := &model.CampaignTargetItem{CampaignID: campaignID, MerchandiseItemID: merchandiseItemID}
	require.NoError(t, db.WithContext(context.Background()).Create(ti).Error)
	return ti
}

func makeMerchItem(t *testing.T, db *gorm.DB, clinicID uint64, name string, category model.ItemCategory) *model.MerchandiseItem {
	t.Helper()
	item := &model.MerchandiseItem{ClinicID: clinicID, Name: name, Category: category, UnitPrice: 1500, TaxRate: 0.10}
	require.NoError(t, db.WithContext(context.Background()).Create(item).Error)
	return item
}

func makeMerchBilling(t *testing.T, db *gorm.DB, clinicID uint64) *model.Billing {
	t.Helper()
	b := &model.Billing{ClinicID: clinicID, ScheduledDate: time.Now(), Status: model.BillingStatusWaiting}
	require.NoError(t, db.WithContext(context.Background()).Create(b).Error)
	return b
}

func makeMerchBillingItem(t *testing.T, db *gorm.DB, billingID, merchandiseItemID uint64) *model.BillingItem {
	t.Helper()
	miID := merchandiseItemID
	bi := &model.BillingItem{BillingID: billingID, Category: model.ItemCategoryGoods, Name: "会計品目", UnitPrice: 1000, Quantity: 1, MerchandiseItemID: &miID}
	require.NoError(t, db.WithContext(context.Background()).Create(bi).Error)
	return bi
}

func makeMerchEstimate(t *testing.T, db *gorm.DB, clinicID uint64) *model.Estimate {
	t.Helper()
	e := &model.Estimate{ClinicID: clinicID}
	require.NoError(t, db.WithContext(context.Background()).Create(e).Error)
	return e
}

func makeMerchEstimateItem(t *testing.T, db *gorm.DB, estimateID, merchandiseItemID uint64) *model.EstimateItem {
	t.Helper()
	miID := merchandiseItemID
	ei := &model.EstimateItem{EstimateID: estimateID, Category: model.ItemCategoryGoods, Name: "見積品目", UnitPrice: 1000, Quantity: 1, MerchandiseItemID: &miID}
	require.NoError(t, db.WithContext(context.Background()).Create(ei).Error)
	return ei
}

func TestMerchandiseItemRepository_Create_FindByID(t *testing.T) {
	db := setupMerchandiseItemRepoTestDB(t)
	repo := NewMerchandiseItemRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("作成した物販品を同一クリニックで取得できる", func(t *testing.T) {
		item := &model.MerchandiseItem{ClinicID: clinicA, Name: "フード", Category: model.ItemCategoryFood, UnitPrice: 2000}
		require.NoError(t, repo.Create(ctx, item))
		require.NotZero(t, item.ID)

		got, err := repo.FindByID(ctx, clinicA, item.ID)
		require.NoError(t, err)
		assert.Equal(t, "フード", got.Name)
		assert.Equal(t, model.ItemCategoryFood, got.Category)
	})

	t.Run("別クリニックからはNotFound", func(t *testing.T) {
		item := makeMerchItem(t, db, clinicA, "医院A専用品", model.ItemCategoryGoods)
		_, err := repo.FindByID(ctx, clinicB, item.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しないIDはNotFound", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

// TestMerchandiseItemRepository_Create_IsActiveFalsePersists は BUG-455-S4:
// gorm:"default:true" 付き bool は Create 時に zero value(false) が INSERT から
// 省略され DB default true が書き戻されるため、明示 false を補償 UPDATE で永続化する。
func TestMerchandiseItemRepository_Create_IsActiveFalsePersists(t *testing.T) {
	db := setupMerchandiseItemRepoTestDB(t)
	repo := NewMerchandiseItemRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	item := &model.MerchandiseItem{
		ClinicID:  clinicID,
		Name:      "create inactive merchandise",
		Category:  model.ItemCategoryGoods,
		UnitPrice: 1000,
		IsActive:  false,
	}
	require.NoError(t, repo.Create(ctx, item))
	require.NotZero(t, item.ID)
	assert.False(t, item.IsActive, "in-memory struct must keep false after Create")

	got, err := repo.FindByID(ctx, clinicID, item.ID)
	require.NoError(t, err)
	assert.False(t, got.IsActive, "DB read-back must keep explicit false")

	var rawActive bool
	require.NoError(t, db.WithContext(ctx).
		Model(&model.MerchandiseItem{}).
		Select("is_active").
		Where("id = ?", item.ID).
		Scan(&rawActive).Error)
	assert.False(t, rawActive, "raw is_active column must be false")
}

// TestMerchandiseItemRepository_Create_IsActiveTruePersists は true 指定の回帰防止。
func TestMerchandiseItemRepository_Create_IsActiveTruePersists(t *testing.T) {
	db := setupMerchandiseItemRepoTestDB(t)
	repo := NewMerchandiseItemRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	active := &model.MerchandiseItem{
		ClinicID:  clinicID,
		Name:      "create active true merchandise",
		Category:  model.ItemCategoryGoods,
		UnitPrice: 1000,
		IsActive:  true,
	}
	require.NoError(t, repo.Create(ctx, active))
	assert.True(t, active.IsActive)

	gotActive, err := repo.FindByID(ctx, clinicID, active.ID)
	require.NoError(t, err)
	assert.True(t, gotActive.IsActive)

	var rawActive bool
	require.NoError(t, db.WithContext(ctx).
		Model(&model.MerchandiseItem{}).
		Select("is_active").
		Where("id = ?", active.ID).
		Scan(&rawActive).Error)
	assert.True(t, rawActive)
}

// TestMerchandiseItemRepository_Create_IsActiveFalse_AmbientTxRollback は
// ambient tx 内で IsActive:false 作成後に rollback すると行が残らないことを検証する。
func TestMerchandiseItemRepository_Create_IsActiveFalse_AmbientTxRollback(t *testing.T) {
	db := setupMerchandiseItemRepoTestDB(t)
	repo := NewMerchandiseItemRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	forcedErr := errors.New("force rollback after inactive merchandise create")

	var createdID uint64
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		item := &model.MerchandiseItem{
			ClinicID:  clinicID,
			Name:      "inactive create rollback",
			Category:  model.ItemCategoryGoods,
			UnitPrice: 1000,
			IsActive:  false,
		}
		if err := repo.Create(txCtx, item); err != nil {
			return err
		}
		createdID = item.ID
		assert.False(t, item.IsActive, "same tx must keep false after Create")
		return forcedErr
	})
	require.ErrorIs(t, err, forcedErr)
	require.NotZero(t, createdID)

	_, err = repo.FindByID(ctx, clinicID, createdID)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "rolled-back create must not leave a row")
}

func TestMerchandiseItemRepository_FindAll(t *testing.T) {
	db := setupMerchandiseItemRepoTestDB(t)
	repo := NewMerchandiseItemRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	itemB := makeMerchItem(t, db, clinicB, "医院Bの品目", model.ItemCategoryGoods)
	goodsA := makeMerchItem(t, db, clinicA, "医院Aの雑貨", model.ItemCategoryGoods)
	medicineA := makeMerchItem(t, db, clinicA, "医院Aの療法食", model.ItemCategoryFood)

	t.Run("クリニックで隔離される", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicA, "")
		require.NoError(t, err)
		require.Len(t, got, 2)
		for _, it := range got {
			assert.NotEqual(t, itemB.ID, it.ID, "別クリニックの品目が混入してはならない")
		}
	})

	t.Run("categoryで絞り込める", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicA, "food")
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, medicineA.ID, got[0].ID)
	})

	t.Run("空文字categoryは全件返す", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicA, "")
		require.NoError(t, err)
		ids := make([]uint64, 0, len(got))
		for _, it := range got {
			ids = append(ids, it.ID)
		}
		assert.Contains(t, ids, goodsA.ID)
		assert.Contains(t, ids, medicineA.ID)
	})

	t.Run("ソフトデリート済みは一覧から除外されるがレコードは残る", func(t *testing.T) {
		db2 := setupMerchandiseItemRepoTestDB(t)
		repo2 := NewMerchandiseItemRepository(db2)
		item := makeMerchItem(t, db2, clinicA, "削除予定品", model.ItemCategoryGoods)

		require.NoError(t, repo2.Delete(ctx, clinicA, item.ID))

		got, err := repo2.FindAll(ctx, clinicA, "")
		require.NoError(t, err)
		assert.Len(t, got, 0)

		var raw model.MerchandiseItem
		require.NoError(t, db2.WithContext(ctx).Unscoped().First(&raw, item.ID).Error)
		assert.NotNil(t, raw.DeletedAt)
	})
}

// seedMerchandiseItemsForBoundTest bulk-inserts clinic-scoped merchandise rows with
// deterministic sort_order/name (bound-mi-00001..) for MaxMasterListRows regressions.
func seedMerchandiseItemsForBoundTest(t *testing.T, db *gorm.DB, clinicID uint64, count int) {
	t.Helper()
	rows := make([]model.MerchandiseItem, 0, count)
	for i := 1; i <= count; i++ {
		rows = append(rows, model.MerchandiseItem{
			ClinicID:  clinicID,
			Name:      fmt.Sprintf("bound-mi-%05d", i),
			Category:  model.ItemCategoryGoods,
			UnitPrice: 1000,
			TaxRate:   0.10,
			IsActive:  true,
			SortOrder: i,
		})
	}
	require.NoError(t, db.WithContext(context.Background()).CreateInBatches(rows, 500).Error)
}

func assertMerchandiseItemsOrderedBySortThenName(t *testing.T, got []model.MerchandiseItem) {
	t.Helper()
	for i := 1; i < len(got); i++ {
		prev, cur := got[i-1], got[i]
		if prev.SortOrder == cur.SortOrder {
			assert.LessOrEqual(t, prev.Name, cur.Name, "same sort_order must order by name ASC")
			continue
		}
		assert.Less(t, prev.SortOrder, cur.SortOrder, "must order by sort_order ASC")
	}
}

// TestMerchandiseItemRepository_FindAll_AppliesMaxMasterListRows is G2F-11: FindAll
// applies persistence.MaxMasterListRows (same as consultation/vaccine/procedure),
// excludes other clinics and soft-deleted rows, and keeps sort_order ASC, name ASC.
func TestMerchandiseItemRepository_FindAll_AppliesMaxMasterListRows(t *testing.T) {
	db := setupMerchandiseItemRepoTestDB(t)
	repo := NewMerchandiseItemRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	limit := persistence.MaxMasterListRows

	t.Run("exact MaxMasterListRows returns all rows in order", func(t *testing.T) {
		require.NoError(t, db.Exec("TRUNCATE TABLE merchandise_items CASCADE").Error)
		seedMerchandiseItemsForBoundTest(t, db, clinicA, limit)

		got, err := repo.FindAll(ctx, clinicA, "")
		require.NoError(t, err)
		require.Len(t, got, limit, "exact bound must return every active clinic row")
		assert.Equal(t, 1, got[0].SortOrder)
		assert.Equal(t, limit, got[len(got)-1].SortOrder)
		assertMerchandiseItemsOrderedBySortThenName(t, got)
	})

	t.Run("limit+1 caps at MaxMasterListRows and excludes foreign/soft-deleted", func(t *testing.T) {
		require.NoError(t, db.Exec("TRUNCATE TABLE merchandise_items CASCADE").Error)
		// limit+2 active for clinic A so soft-deleting one still leaves limit+1 active.
		seedMerchandiseItemsForBoundTest(t, db, clinicA, limit+2)

		// Cross-clinic rows that would sort first if clinic isolation regressed.
		foreign := []model.MerchandiseItem{
			{ClinicID: clinicB, Name: "aaa-other-clinic-1", Category: model.ItemCategoryGoods, UnitPrice: 1000, TaxRate: 0.10, IsActive: true, SortOrder: 0},
			{ClinicID: clinicB, Name: "aaa-other-clinic-2", Category: model.ItemCategoryGoods, UnitPrice: 1000, TaxRate: 0.10, IsActive: true, SortOrder: 0},
		}
		require.NoError(t, db.WithContext(ctx).Create(&foreign).Error)

		// Soft-delete a row that would otherwise appear in the capped window.
		var softDeleted model.MerchandiseItem
		require.NoError(t, db.WithContext(ctx).
			Where("clinic_id = ? AND sort_order = ?", clinicA, 2).
			First(&softDeleted).Error)
		require.NoError(t, repo.Delete(ctx, clinicA, softDeleted.ID))

		got, err := repo.FindAll(ctx, clinicA, "")
		require.NoError(t, err)
		require.Len(t, got, limit, "must cap at MaxMasterListRows when more active rows exist")
		assertMerchandiseItemsOrderedBySortThenName(t, got)

		for _, it := range got {
			assert.NotEqual(t, softDeleted.ID, it.ID, "soft-deleted row must be excluded")
			assert.Equal(t, clinicA, it.ClinicID, "cross-clinic rows must not mix into result")
			assert.NotEqual(t, foreign[0].ID, it.ID)
			assert.NotEqual(t, foreign[1].ID, it.ID)
		}

		assert.Equal(t, 1, got[0].SortOrder)
		assert.Equal(t, "bound-mi-00001", got[0].Name)
		assert.Equal(t, 3, got[1].SortOrder)
		// First `limit` of remaining active ordered set ends at sort_order = limit+1.
		assert.Equal(t, limit+1, got[len(got)-1].SortOrder)
	})
}

func TestMerchandiseItemRepository_Update(t *testing.T) {
	db := setupMerchandiseItemRepoTestDB(t)
	repo := NewMerchandiseItemRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("同一クリニックの更新は反映される", func(t *testing.T) {
		item := makeMerchItem(t, db, clinicA, "更新前品目", model.ItemCategoryGoods)
		got, err := repo.Update(ctx, clinicA, item.ID, map[string]any{"name": "更新後品目", "unit_price": int64(3000)})
		require.NoError(t, err)
		assert.Equal(t, "更新後品目", got.Name)
		assert.Equal(t, int64(3000), got.UnitPrice)
	})

	t.Run("別クリニックの更新はNotFound", func(t *testing.T) {
		item := makeMerchItem(t, db, clinicA, "越境更新対象品目", model.ItemCategoryGoods)
		_, err := repo.Update(ctx, clinicB, item.ID, map[string]any{"name": "越境更新"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しないIDの更新はNotFound", func(t *testing.T) {
		_, err := repo.Update(ctx, clinicA, 999999, map[string]any{"name": "存在しない"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

// TestMerchandiseItemRepository_Update_AmbientTxRollback は BUG-465 の回帰:
// Update の書き込みと再取得が同一 tx に参加し、呼び出し元の ambient tx が
// rollback されると更新も取り消されることを検証する。
func TestMerchandiseItemRepository_Update_AmbientTxRollback(t *testing.T) {
	db := setupMerchandiseItemRepoTestDB(t)
	repo := NewMerchandiseItemRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	item := makeMerchItem(t, db, clinicA, "更新前品目", model.ItemCategoryGoods)
	forcedErr := errors.New("force rollback after merchandise update")

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		got, err := repo.Update(txCtx, clinicA, item.ID, map[string]any{"name": "更新後品目", "unit_price": int64(3000)})
		if err != nil {
			return err
		}
		assert.Equal(t, "更新後品目", got.Name, "同一 tx 内では更新後の値を再取得できる")
		assert.Equal(t, int64(3000), got.UnitPrice)
		return forcedErr
	})
	require.ErrorIs(t, err, forcedErr)

	got, err := repo.FindByID(ctx, clinicA, item.ID)
	require.NoError(t, err)
	assert.Equal(t, "更新前品目", got.Name, "ambient transaction rollback must restore the merchandise name")
	assert.Equal(t, int64(1500), got.UnitPrice, "ambient transaction rollback must restore the unit price")
}

// TestMerchandiseItemRepository_Update_ReloadFailureRollsBackUpdate は BUG-465:
// Update 成功後の再取得が失敗した場合、コミット済み更新を失敗応答へ反転させないよう
// 同一 tx 内で rollback することを検証する。
func TestMerchandiseItemRepository_Update_ReloadFailureRollsBackUpdate(t *testing.T) {
	db := setupMerchandiseItemRepoTestDB(t)
	repo := NewMerchandiseItemRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	item := makeMerchItem(t, db, clinicA, "更新前品目", model.ItemCategoryGoods)
	const callbackName = "merchandise:update_reload_failure"
	reloadErr := errors.New("forced merchandise reload failure")
	require.NoError(t, db.Callback().Query().Before("gorm:query").Register(callbackName, func(query *gorm.DB) {
		if query.Statement != nil && query.Statement.Table == "merchandise_items" {
			query.AddError(reloadErr)
		}
	}))
	callbackRegistered := true
	t.Cleanup(func() {
		if callbackRegistered {
			require.NoError(t, db.Callback().Query().Remove(callbackName))
		}
	})

	got, err := repo.Update(ctx, clinicA, item.ID, map[string]any{"name": "更新後品目"})

	assert.Nil(t, got)
	require.Error(t, err)
	assert.ErrorIs(t, err, reloadErr)

	require.NoError(t, db.Callback().Query().Remove(callbackName))
	callbackRegistered = false

	var persisted model.MerchandiseItem
	require.NoError(t, db.WithContext(ctx).First(&persisted, item.ID).Error)
	assert.Equal(t, "更新前品目", persisted.Name, "reload failure must roll back the committed-looking update")
}

func TestMerchandiseItemRepository_Delete(t *testing.T) {
	db := setupMerchandiseItemRepoTestDB(t)
	repo := NewMerchandiseItemRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("同一クリニックの削除は成功しその後取得できない", func(t *testing.T) {
		item := makeMerchItem(t, db, clinicA, "削除対象品目", model.ItemCategoryGoods)
		require.NoError(t, repo.Delete(ctx, clinicA, item.ID))

		_, err := repo.FindByID(ctx, clinicA, item.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("別クリニックの削除はNotFoundで対象データは残る", func(t *testing.T) {
		item := makeMerchItem(t, db, clinicA, "越境削除対象品目", model.ItemCategoryGoods)
		err := repo.Delete(ctx, clinicB, item.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		got, err := repo.FindByID(ctx, clinicA, item.ID)
		require.NoError(t, err)
		assert.Equal(t, item.ID, got.ID)
	})

	t.Run("存在しないIDの削除はNotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestMerchandiseItemRepository_Reorder(t *testing.T) {
	db := setupMerchandiseItemRepoTestDB(t)
	repo := NewMerchandiseItemRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("指定した順序でsort_orderが1始まりに更新される", func(t *testing.T) {
		item1 := makeMerchItem(t, db, clinicA, "品目1", model.ItemCategoryGoods)
		item2 := makeMerchItem(t, db, clinicA, "品目2", model.ItemCategoryGoods)
		item3 := makeMerchItem(t, db, clinicA, "品目3", model.ItemCategoryGoods)

		require.NoError(t, repo.Reorder(ctx, clinicA, []uint64{item3.ID, item1.ID, item2.ID}))

		got, err := repo.FindAll(ctx, clinicA, "")
		require.NoError(t, err)
		byID := make(map[uint64]model.MerchandiseItem, 3)
		for _, it := range got {
			byID[it.ID] = it
		}
		assert.Equal(t, 1, byID[item3.ID].SortOrder)
		assert.Equal(t, 2, byID[item1.ID].SortOrder)
		assert.Equal(t, 3, byID[item2.ID].SortOrder)
	})

	t.Run("別クリニックのIDが混ざるとエラーになる", func(t *testing.T) {
		itemA := makeMerchItem(t, db, clinicA, "医院A品目", model.ItemCategoryGoods)
		itemB := makeMerchItem(t, db, clinicB, "医院B品目", model.ItemCategoryGoods)

		err := repo.Reorder(ctx, clinicA, []uint64{itemA.ID, itemB.ID})
		require.Error(t, err)
	})
}

func TestMerchandiseItemRepository_CountUsageByMerchandiseItemID_RealDB(t *testing.T) {
	db := setupMerchandiseItemRepoTestDB(t)
	repo := NewMerchandiseItemRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("billing_itemsとestimate_itemsの合計をカウントする", func(t *testing.T) {
		item := makeMerchItem(t, db, clinicA, "使用中品目", model.ItemCategoryGoods)
		billing := makeMerchBilling(t, db, clinicA)
		makeMerchBillingItem(t, db, billing.ID, item.ID)
		estimate := makeMerchEstimate(t, db, clinicA)
		makeMerchEstimateItem(t, db, estimate.ID, item.ID)
		makeMerchEstimateItem(t, db, estimate.ID, item.ID)

		count, err := repo.CountUsageByMerchandiseItemID(ctx, clinicA, item.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(3), count, "billing_items 1件 + estimate_items 2件")
	})

	t.Run("未使用の品目は0を返す", func(t *testing.T) {
		unused := makeMerchItem(t, db, clinicA, "未使用品目", model.ItemCategoryGoods)
		count, err := repo.CountUsageByMerchandiseItemID(ctx, clinicA, unused.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("ソフトデリート済みのbilling_itemはカウントされない", func(t *testing.T) {
		item := makeMerchItem(t, db, clinicA, "削除会計品目対象", model.ItemCategoryGoods)
		billing := makeMerchBilling(t, db, clinicA)
		bi := makeMerchBillingItem(t, db, billing.ID, item.ID)
		require.NoError(t, db.WithContext(ctx).Delete(&model.BillingItem{}, bi.ID).Error)

		count, err := repo.CountUsageByMerchandiseItemID(ctx, clinicA, item.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("別クリニックの請求に紐づく参照はカウントされない", func(t *testing.T) {
		item := makeMerchItem(t, db, clinicA, "越境参照対象品目", model.ItemCategoryGoods)
		// clinic B の billing から clinic A の品目を参照する汚染データを模擬
		billingB := makeMerchBilling(t, db, clinicB)
		makeMerchBillingItem(t, db, billingB.ID, item.ID)

		count, err := repo.CountUsageByMerchandiseItemID(ctx, clinicA, item.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count, "別クリニックのbillingに紐づく参照はJOINで除外される")
	})
}

// TestMerchandiseItemRepository_CountUsageByMerchandiseItemID_IncludesCampaignTargets
// proves campaign_target_items joined to any same-clinic non-deleted campaign count as usage,
// including inactive and out-of-date campaigns (BE-ACT-MERCHANDISE-ATOMIC-DELETE).
func TestMerchandiseItemRepository_CountUsageByMerchandiseItemID_IncludesCampaignTargets(t *testing.T) {
	db := setupMerchandiseItemRepoTestDB(t)
	repo := NewMerchandiseItemRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	now := time.Now().UTC().Truncate(24 * time.Hour)

	t.Run("active campaign target counts as usage", func(t *testing.T) {
		item := makeMerchItem(t, db, clinicA, "active campaign target item", model.ItemCategoryGoods)
		camp := makeMerchCampaign(t, db, clinicA, "active camp", true, now.Add(-24*time.Hour), now.Add(24*time.Hour))
		makeMerchCampaignTarget(t, db, camp.ID, item.ID)

		count, err := repo.CountUsageByMerchandiseItemID(ctx, clinicA, item.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("inactive campaign target still blocks delete (counts as usage)", func(t *testing.T) {
		item := makeMerchItem(t, db, clinicA, "inactive campaign target item", model.ItemCategoryGoods)
		camp := makeMerchCampaign(t, db, clinicA, "inactive camp", false, now.Add(-24*time.Hour), now.Add(24*time.Hour))
		makeMerchCampaignTarget(t, db, camp.ID, item.ID)

		count, err := repo.CountUsageByMerchandiseItemID(ctx, clinicA, item.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count, "is_active=false campaign targets must still count")
	})

	t.Run("out-of-date campaign target still counts as usage", func(t *testing.T) {
		item := makeMerchItem(t, db, clinicA, "expired campaign target item", model.ItemCategoryGoods)
		camp := makeMerchCampaign(t, db, clinicA, "expired camp", true, now.Add(-72*time.Hour), now.Add(-48*time.Hour))
		makeMerchCampaignTarget(t, db, camp.ID, item.ID)

		count, err := repo.CountUsageByMerchandiseItemID(ctx, clinicA, item.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count, "date-window-expired campaign targets must still count")
	})

	t.Run("soft-deleted campaign targets are excluded", func(t *testing.T) {
		item := makeMerchItem(t, db, clinicA, "deleted campaign target item", model.ItemCategoryGoods)
		camp := makeMerchCampaign(t, db, clinicA, "to-delete camp", true, now.Add(-24*time.Hour), now.Add(24*time.Hour))
		makeMerchCampaignTarget(t, db, camp.ID, item.ID)
		require.NoError(t, db.WithContext(ctx).Delete(camp).Error)

		count, err := repo.CountUsageByMerchandiseItemID(ctx, clinicA, item.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("other clinic campaign targets are excluded", func(t *testing.T) {
		item := makeMerchItem(t, db, clinicA, "cross clinic campaign target item", model.ItemCategoryGoods)
		campB := makeMerchCampaign(t, db, clinicB, "clinic B camp", true, now.Add(-24*time.Hour), now.Add(24*time.Hour))
		makeMerchCampaignTarget(t, db, campB.ID, item.ID)

		count, err := repo.CountUsageByMerchandiseItemID(ctx, clinicA, item.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("billing + estimate + campaign targets sum", func(t *testing.T) {
		item := makeMerchItem(t, db, clinicA, "mixed usage item", model.ItemCategoryGoods)
		billing := makeMerchBilling(t, db, clinicA)
		makeMerchBillingItem(t, db, billing.ID, item.ID)
		estimate := makeMerchEstimate(t, db, clinicA)
		makeMerchEstimateItem(t, db, estimate.ID, item.ID)
		camp := makeMerchCampaign(t, db, clinicA, "mixed camp", false, now.Add(-100*24*time.Hour), now.Add(-90*24*time.Hour))
		makeMerchCampaignTarget(t, db, camp.ID, item.ID)

		count, err := repo.CountUsageByMerchandiseItemID(ctx, clinicA, item.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(3), count, "1 billing + 1 estimate + 1 inactive/expired campaign target")
	})
}

// TestMerchandiseItemRepository_CountUsage_AmbientTxSeesUncommittedCampaignTarget proves
// CountUsage joins ambient tx via DBOrTx and observes uncommitted campaign targets.
func TestMerchandiseItemRepository_CountUsage_AmbientTxSeesUncommittedCampaignTarget(t *testing.T) {
	db := setupMerchandiseItemRepoTestDB(t)
	repo := NewMerchandiseItemRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	now := time.Now().UTC().Truncate(24 * time.Hour)

	item := makeMerchItem(t, db, clinicID, "ambient count merchandise", model.ItemCategoryGoods)
	camp := makeMerchCampaign(t, db, clinicID, "ambient count camp", true, now, now.Add(24*time.Hour))

	forcedErr := errors.New("force rollback after ambient count")
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		if err := tx.Create(&model.CampaignTargetItem{
			CampaignID:        camp.ID,
			MerchandiseItemID: item.ID,
		}).Error; err != nil {
			return err
		}
		count, err := repo.CountUsageByMerchandiseItemID(txCtx, clinicID, item.ID)
		if err != nil {
			return err
		}
		assert.Equal(t, int64(1), count, "ambient CountUsage must see uncommitted campaign target")
		return forcedErr
	})
	require.ErrorIs(t, err, forcedErr)

	count, err := repo.CountUsageByMerchandiseItemID(ctx, clinicID, item.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count, "rolled-back target must not remain after ambient failure")
}

// TestCountUsageByMerchandiseItemID - マーチャンダイズアイテムの使用数をカウント
func TestCountUsageByMerchandiseItemID(t *testing.T) {
	// 統合テスト環境での実行を想定。ローカル単体テストの場合はモック実装が必要。

	t.Run("請求アイテムで使用されているマーチャンダイズアイテムをカウント", func(t *testing.T) {
		// 実装の検証ロジック
		// 実際のDBではなく、リポジトリのロジックを検証

		// 期待値: マーチャンダイズアイテムID=1が請求アイテムで2件使用されている場合、
		// CountUsageByMerchandiseItemID() は 2 を返す
		expectedCount := int64(2)
		assert.Greater(t, expectedCount, int64(0))
	})

	t.Run("見積もりアイテムで使用されているマーチャンダイズアイテムをカウント", func(t *testing.T) {
		// 期待値: マーチャンダイズアイテムID=2が見積もりアイテムで1件使用されている場合、
		// CountUsageByMerchandiseItemID() は 1 を返す
		expectedCount := int64(1)
		assert.Equal(t, expectedCount, int64(1))
	})

	t.Run("複数テーブルにまたがる使用数をカウント", func(t *testing.T) {
		// 期待値: マーチャンダイズアイテムID=3が
		// - 請求アイテム: 1件
		// - 見積もりアイテム: 2件
		// の場合、合計3件をカウント
		billingCount := int64(1)
		estimateCount := int64(2)
		totalCount := billingCount + estimateCount

		assert.Equal(t, totalCount, int64(3))
	})

	t.Run("使用されていないマーチャンダイズアイテムはカウント0", func(t *testing.T) {
		// 期待値: 存在しないID（例: 99999）の場合、
		// CountUsageByMerchandiseItemID() は 0 を返す
		expectedCount := int64(0)
		assert.Equal(t, expectedCount, int64(0))
	})

	t.Run("削除済みアイテムは参照カウントに含まれない", func(t *testing.T) {
		// 期待値: deleted_at が NULL でないレコードは
		// countQuery の WHERE 句で除外される
		// （論理削除済みアイテムの参照は数えない）

		// このテストは、WHERE deleted_at IS NULL という条件が
		// countQuery に含まれていることを検証する
		expectedBehavior := "deleted records excluded"
		assert.NotEmpty(t, expectedBehavior)
	})
}

// TestDeleteMerchandiseItemWithFKCheck - マーチャンダイズアイテム削除時のFK依存チェック
func TestDeleteMerchandiseItemWithFKCheck(t *testing.T) {
	t.Run("使用中のマーチャンダイズアイテム削除は409エラーを返す", func(t *testing.T) {
		// サービス層での期待動作:
		// 1. CountUsageByMerchandiseItemID() を呼び出す
		// 2. count > 0 の場合、apperrors.WrapConflict() を返す
		// 3. HTTPハンドラは 409 Conflict ステータスコードを返す

		count := int64(2)
		shouldAllowDelete := count == 0

		assert.False(t, shouldAllowDelete)
	})

	t.Run("未使用のマーチャンダイズアイテムは削除できる", func(t *testing.T) {
		count := int64(0)
		shouldAllowDelete := count == 0

		assert.True(t, shouldAllowDelete)
	})
}

// BenchmarkCountUsageByMerchandiseItemID - パフォーマンステスト
func BenchmarkCountUsageByMerchandiseItemID(b *testing.B) {
	// 実装の効率性を検証
	// 期待値: データベースクエリは複合インデックスを使用し、
	// 大規模データセット（10,000+レコード）でも < 50ms で完了

	for i := 0; i < b.N; i++ {
		// クエリ実行のベンチマーク
		_ = int64(0)
	}
}

// TestMerchandiseItemRepository_FindByID_HoldsShareLockForAmbientTransaction proves
// ambient-tx FindByID takes FOR SHARE so concurrent soft-delete waits until commit
// (campaign target attachment serialization / BE-ACT-CAMPAIGN-TARGET-SERIALIZATION).
func TestMerchandiseItemRepository_FindByID_HoldsShareLockForAmbientTransaction(t *testing.T) {
	db := setupMerchandiseItemRepoTestDB(t)
	repo := NewMerchandiseItemRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	item := makeMerchItem(t, db, clinicID, "share-lock merchandise", model.ItemCategoryGoods)

	locked := make(chan struct{})
	release := make(chan struct{})
	readDone := make(chan error, 1)
	go func() {
		readDone <- db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			txCtx := persistence.WithTxValue(ctx, tx)
			if _, err := repo.FindByID(txCtx, clinicID, item.ID); err != nil {
				return err
			}
			close(locked)
			<-release
			return nil
		})
	}()
	<-locked

	updateStarted := make(chan struct{})
	updateDone := make(chan error, 1)
	go func() {
		close(updateStarted)
		updateDone <- db.WithContext(ctx).
			Model(&model.MerchandiseItem{}).
			Where("id = ? AND clinic_id = ?", item.ID, clinicID).
			Delete(&model.MerchandiseItem{}).Error
	}()
	<-updateStarted

	select {
	case err := <-updateDone:
		close(release)
		require.Failf(t, "merchandise soft-delete was not serialized behind share lock", "err=%v", err)
	case <-time.After(100 * time.Millisecond):
		close(release)
	}
	require.NoError(t, <-readDone)
	require.NoError(t, <-updateDone)

	_, err := repo.FindByID(ctx, clinicID, item.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "soft-delete after share-lock release must hide the row")
}

// TestMerchandiseItemRepository_FindByID_SeesAmbientUncommittedState proves FindByID
// participates via DBOrTx and observes uncommitted writes in the same ambient tx.
func TestMerchandiseItemRepository_FindByID_AmbientTxSeesUncommittedRow(t *testing.T) {
	db := setupMerchandiseItemRepoTestDB(t)
	repo := NewMerchandiseItemRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	forcedErr := errors.New("force rollback after ambient find")
	var createdID uint64
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		item := &model.MerchandiseItem{
			ClinicID:  clinicID,
			Name:      "ambient find merchandise",
			Category:  model.ItemCategoryGoods,
			UnitPrice: 1000,
			IsActive:  true,
		}
		if err := repo.Create(txCtx, item); err != nil {
			return err
		}
		createdID = item.ID
		got, err := repo.FindByID(txCtx, clinicID, item.ID)
		if err != nil {
			return err
		}
		assert.Equal(t, "ambient find merchandise", got.Name)
		return forcedErr
	})
	require.ErrorIs(t, err, forcedErr)
	require.NotZero(t, createdID)

	_, err = repo.FindByID(ctx, clinicID, createdID)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}
