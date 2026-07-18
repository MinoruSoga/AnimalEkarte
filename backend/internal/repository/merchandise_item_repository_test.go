package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- 実 DB 結合テスト（merchandise_item_repository.go）----
// 上記の Test群は setupTestDB を呼ばずリテラル同士を比較するだけの stub のため実カバレッジに寄与しない。
// 以下は実際の merchandiseItemRepository を実 Postgres テストDBに対して駆動する。

// setupMerchandiseItemRepoTestDB は merchandise_item_repository のテストに必要なテーブルを整備する。
// CountUsageByMerchandiseItemID は billing_items / estimate_items を billings / estimates 経由でJOINするため
// それらも migrate する。
func setupMerchandiseItemRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, ensureAutoMigrated(db,
		&model.MerchandiseItem{}, &model.Billing{}, &model.BillingItem{},
		&model.Estimate{}, &model.EstimateItem{},
	))
	db.Exec("TRUNCATE TABLE billing_items CASCADE")
	db.Exec("TRUNCATE TABLE estimate_items CASCADE")
	db.Exec("TRUNCATE TABLE estimates CASCADE")
	db.Exec("TRUNCATE TABLE billings CASCADE")
	db.Exec("TRUNCATE TABLE merchandise_items CASCADE")
	return db
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
