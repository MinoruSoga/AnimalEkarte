package repository

// billing_item_lstep_queries_test.go — G11-5 テスト負債解消
//
// テスト対象: BillingItemRepository.HasItemByOwnerSince / FindOwnersByCategoryPurchaseDate /
//   HasFoodPurchaseByOwnerSince — Lstep 配信ターゲティング用の購買クエリ3本（FEAT-379/383）
// 保護する不変条件:
//   - すべて billings.completed_at を基準とする（billings.issued_at ではない。ドキュメント
//     コメントの乖離を修正した根拠を固定する）
//   - FindOwnersByCategoryPurchaseDate は HAVING MAX(completed_at::date) = purchaseDate、
//     つまり「その日が最終購入日」を意味する（それ以降により新しい購入があると除外される）
//   - HasFoodPurchaseByOwnerSince は names 指定時は name IN、未指定時は category=food
//
// このテストは completed_at を issued_at 等の別カラムに差し替える、または
// HAVING MAX の意味論を崩すと必ず失敗するよう設計されている。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestBillingItemRepository_HasItemByOwnerSince(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	repo := NewBillingItemRepository(db)
	const clinicA, clinicB = uint64(1), uint64(2)
	since := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	owner := makeOwner(t, db, clinicA, "G11-5飼主1")
	completedAt := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	billing := makeTrimmingBillingWithCompletedAt(t, db, clinicA, model.BillingStatusCompleted, completedAt)
	billing.OwnerID = &owner.ID
	require.NoError(t, db.WithContext(ctx).Save(billing).Error)
	item := &model.BillingItem{BillingID: billing.ID, Category: model.ItemCategoryGoods, Name: "対象商品A"}
	require.NoError(t, db.WithContext(ctx).Create(item).Error)

	t.Run("names一致+completed_at>=sinceでtrue", func(t *testing.T) {
		ok, err := repo.HasItemByOwnerSince(ctx, clinicA, owner.ID, since, []string{"対象商品A"})
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("since前の購入のみならfalse", func(t *testing.T) {
		earlyBilling := makeTrimmingBillingWithCompletedAt(t, db, clinicA, model.BillingStatusCompleted, time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC))
		earlyItem := &model.BillingItem{BillingID: earlyBilling.ID, Category: model.ItemCategoryGoods, Name: "対象商品早期"}
		require.NoError(t, db.WithContext(ctx).Create(earlyItem).Error)
		ownerEarly := makeOwner(t, db, clinicA, "G11-5飼主early")
		earlyBilling.OwnerID = &ownerEarly.ID
		require.NoError(t, db.WithContext(ctx).Save(earlyBilling).Error)

		ok, err := repo.HasItemByOwnerSince(ctx, clinicA, ownerEarly.ID, since, []string{"対象商品早期"})
		require.NoError(t, err)
		assert.False(t, ok, "since より前の購入のみでは対象外")
	})

	t.Run("名前不一致ならfalse", func(t *testing.T) {
		ok, err := repo.HasItemByOwnerSince(ctx, clinicA, owner.ID, since, []string{"存在しない商品名"})
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("別ownerならfalse", func(t *testing.T) {
		otherOwner := makeOwner(t, db, clinicA, "G11-5別飼主")
		ok, err := repo.HasItemByOwnerSince(ctx, clinicA, otherOwner.ID, since, []string{"対象商品A"})
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("別クリニックならfalse", func(t *testing.T) {
		ok, err := repo.HasItemByOwnerSince(ctx, clinicB, owner.ID, since, []string{"対象商品A"})
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("soft-deleted billingならfalse", func(t *testing.T) {
		deletedOwner := makeOwner(t, db, clinicA, "G11-5削除済み飼主")
		deletedBilling := makeTrimmingBillingWithCompletedAt(t, db, clinicA, model.BillingStatusCompleted, completedAt)
		deletedBilling.OwnerID = &deletedOwner.ID
		require.NoError(t, db.WithContext(ctx).Save(deletedBilling).Error)
		deletedItem := &model.BillingItem{BillingID: deletedBilling.ID, Category: model.ItemCategoryGoods, Name: "削除済み商品"}
		require.NoError(t, db.WithContext(ctx).Create(deletedItem).Error)
		require.NoError(t, db.Delete(&model.Billing{}, deletedBilling.ID).Error)

		ok, err := repo.HasItemByOwnerSince(ctx, clinicA, deletedOwner.ID, since, []string{"削除済み商品"})
		require.NoError(t, err)
		assert.False(t, ok, "soft-deleteされたbillingの購買は対象外")
	})

	t.Run("names空ならクエリ発行なしでfalse", func(t *testing.T) {
		ok, err := repo.HasItemByOwnerSince(ctx, clinicA, owner.ID, since, nil)
		require.NoError(t, err)
		assert.False(t, ok)
	})
}

func TestBillingItemRepository_FindOwnersByCategoryPurchaseDate(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	repo := NewBillingItemRepository(db)
	const clinicA = uint64(1)
	targetDate := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)

	t.Run("MAX(completed_at::date)が指定日に一致するownerのみ返る", func(t *testing.T) {
		owner := makeOwner(t, db, clinicA, "G11-5購買飼主1")
		billing := makeTrimmingBillingWithCompletedAt(t, db, clinicA, model.BillingStatusCompleted, targetDate)
		billing.OwnerID = &owner.ID
		require.NoError(t, db.WithContext(ctx).Save(billing).Error)
		item := &model.BillingItem{BillingID: billing.ID, Category: model.ItemCategoryFood, Name: "フードA"}
		require.NoError(t, db.WithContext(ctx).Create(item).Error)

		ids, err := repo.FindOwnersByCategoryPurchaseDate(ctx, clinicA, string(model.ItemCategoryFood), targetDate)
		require.NoError(t, err)
		assert.Contains(t, ids, owner.ID)
	})

	t.Run("同ownerがより新しい購買を持つと除外される(HAVING MAXの本質)", func(t *testing.T) {
		owner := makeOwner(t, db, clinicA, "G11-5購買飼主2")
		oldBilling := makeTrimmingBillingWithCompletedAt(t, db, clinicA, model.BillingStatusCompleted, targetDate)
		oldBilling.OwnerID = &owner.ID
		require.NoError(t, db.WithContext(ctx).Save(oldBilling).Error)
		require.NoError(t, db.WithContext(ctx).Create(&model.BillingItem{BillingID: oldBilling.ID, Category: model.ItemCategoryFood, Name: "フード旧"}).Error)

		newerBilling := makeTrimmingBillingWithCompletedAt(t, db, clinicA, model.BillingStatusCompleted, targetDate.AddDate(0, 0, 5))
		newerBilling.OwnerID = &owner.ID
		require.NoError(t, db.WithContext(ctx).Save(newerBilling).Error)
		require.NoError(t, db.WithContext(ctx).Create(&model.BillingItem{BillingID: newerBilling.ID, Category: model.ItemCategoryFood, Name: "フード新"}).Error)

		ids, err := repo.FindOwnersByCategoryPurchaseDate(ctx, clinicA, string(model.ItemCategoryFood), targetDate)
		require.NoError(t, err)
		assert.NotContains(t, ids, owner.ID, "MAXがtargetDateより後にずれたownerはtargetDateの結果から除外される")

		idsNewer, err := repo.FindOwnersByCategoryPurchaseDate(ctx, clinicA, string(model.ItemCategoryFood), targetDate.AddDate(0, 0, 5))
		require.NoError(t, err)
		assert.Contains(t, idsNewer, owner.ID, "最新購入日を指定すればownerが返る")
	})

	t.Run("別カテゴリ/別クリニック/別日は除外", func(t *testing.T) {
		owner := makeOwner(t, db, clinicA, "G11-5購買飼主3")
		billing := makeTrimmingBillingWithCompletedAt(t, db, clinicA, model.BillingStatusCompleted, targetDate)
		billing.OwnerID = &owner.ID
		require.NoError(t, db.WithContext(ctx).Save(billing).Error)
		require.NoError(t, db.WithContext(ctx).Create(&model.BillingItem{BillingID: billing.ID, Category: model.ItemCategoryGoods, Name: "雑貨"}).Error)

		idsWrongCategory, err := repo.FindOwnersByCategoryPurchaseDate(ctx, clinicA, string(model.ItemCategoryFood), targetDate)
		require.NoError(t, err)
		assert.NotContains(t, idsWrongCategory, owner.ID, "カテゴリ不一致は除外")

		const clinicB = uint64(2)
		idsWrongClinic, err := repo.FindOwnersByCategoryPurchaseDate(ctx, clinicB, string(model.ItemCategoryGoods), targetDate)
		require.NoError(t, err)
		assert.NotContains(t, idsWrongClinic, owner.ID, "別クリニックは除外")

		idsWrongDate, err := repo.FindOwnersByCategoryPurchaseDate(ctx, clinicA, string(model.ItemCategoryGoods), targetDate.AddDate(0, 0, 1))
		require.NoError(t, err)
		assert.NotContains(t, idsWrongDate, owner.ID, "別日は除外")
	})
}

func TestBillingItemRepository_HasFoodPurchaseByOwnerSince(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	repo := NewBillingItemRepository(db)
	const clinicA = uint64(1)
	since := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	completedAt := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)

	t.Run("names指定時はname INで判定する", func(t *testing.T) {
		owner := makeOwner(t, db, clinicA, "G11-5フード飼主1")
		billing := makeTrimmingBillingWithCompletedAt(t, db, clinicA, model.BillingStatusCompleted, completedAt)
		billing.OwnerID = &owner.ID
		require.NoError(t, db.WithContext(ctx).Save(billing).Error)
		// category は food ではないが、names 指定時は category を見ないため対象になるはず。
		require.NoError(t, db.WithContext(ctx).Create(&model.BillingItem{BillingID: billing.ID, Category: model.ItemCategoryGoods, Name: "指定フード"}).Error)

		ok, err := repo.HasFoodPurchaseByOwnerSince(ctx, clinicA, owner.ID, since, []string{"指定フード"})
		require.NoError(t, err)
		assert.True(t, ok, "names指定時はcategoryを問わずname一致で判定")

		okMiss, err := repo.HasFoodPurchaseByOwnerSince(ctx, clinicA, owner.ID, since, []string{"別の名前"})
		require.NoError(t, err)
		assert.False(t, okMiss)
	})

	t.Run("names未指定時はcategory=foodへフォールバックする", func(t *testing.T) {
		owner := makeOwner(t, db, clinicA, "G11-5フード飼主2")
		billing := makeTrimmingBillingWithCompletedAt(t, db, clinicA, model.BillingStatusCompleted, completedAt)
		billing.OwnerID = &owner.ID
		require.NoError(t, db.WithContext(ctx).Save(billing).Error)
		require.NoError(t, db.WithContext(ctx).Create(&model.BillingItem{BillingID: billing.ID, Category: model.ItemCategoryFood, Name: "任意フード名"}).Error)

		ok, err := repo.HasFoodPurchaseByOwnerSince(ctx, clinicA, owner.ID, since, nil)
		require.NoError(t, err)
		assert.True(t, ok, "names未指定時はcategory=foodで判定")

		nonFoodOwner := makeOwner(t, db, clinicA, "G11-5非フード飼主")
		nonFoodBilling := makeTrimmingBillingWithCompletedAt(t, db, clinicA, model.BillingStatusCompleted, completedAt)
		nonFoodBilling.OwnerID = &nonFoodOwner.ID
		require.NoError(t, db.WithContext(ctx).Save(nonFoodBilling).Error)
		require.NoError(t, db.WithContext(ctx).Create(&model.BillingItem{BillingID: nonFoodBilling.ID, Category: model.ItemCategoryGoods, Name: "雑貨"}).Error)

		okNonFood, err := repo.HasFoodPurchaseByOwnerSince(ctx, clinicA, nonFoodOwner.ID, since, nil)
		require.NoError(t, err)
		assert.False(t, okNonFood, "category=food以外はnames未指定フォールバックで対象外")
	})
}

// makeTrimmingBillingWithCompletedAt は completed_at を指定した billing を作成する（owner_id は未設定）。
func makeTrimmingBillingWithCompletedAt(t *testing.T, db *gorm.DB, clinicID uint64, status model.BillingStatus, completedAt time.Time) *model.Billing {
	t.Helper()
	b := &model.Billing{
		ClinicID:      clinicID,
		Status:        status,
		ScheduledDate: completedAt,
		CompletedAt:   &completedAt,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(b).Error)
	return b
}
