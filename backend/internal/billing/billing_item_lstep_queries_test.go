package billing

// billing_item_lstep_queries_test.go — G11-5 テスト負債解消
//
// テスト対象: BillingItemRepository.HasItemByOwnerSince / HasFoodPurchaseByOwnerSince
//   — Lstep 配信ターゲティング用の購買クエリ2本（FEAT-379）
//   （FindOwnersByCategoryPurchaseDate はF6+H-3で未配線のため削除・本テストも削除済み）
// 保護する不変条件:
//   - すべて billings.completed_at を基準とする（billings.issued_at ではない。ドキュメント
//     コメントの乖離を修正した根拠を固定する）
//   - HasFoodPurchaseByOwnerSince は names 指定時は name IN、未指定時は category=food
//
// このテストは completed_at を issued_at 等の別カラムに差し替えると必ず失敗するよう設計されている。

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

func TestBillingItemRepository_HasItemByOwnerSince(t *testing.T) {
	ctx := context.Background()
	db := testdb.SetupTestDB(t)
	repo := NewBillingItemRepository(db)
	const clinicA, clinicB = uint64(1), uint64(2)
	since := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	owner := testdb.MakeTestOwner(t, db, clinicA, "G11-5飼主1")
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
		ownerEarly := testdb.MakeTestOwner(t, db, clinicA, "G11-5飼主early")
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
		otherOwner := testdb.MakeTestOwner(t, db, clinicA, "G11-5別飼主")
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
		deletedOwner := testdb.MakeTestOwner(t, db, clinicA, "G11-5削除済み飼主")
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

func TestBillingItemRepository_HasFoodPurchaseByOwnerSince(t *testing.T) {
	ctx := context.Background()
	db := testdb.SetupTestDB(t)
	repo := NewBillingItemRepository(db)
	const clinicA = uint64(1)
	since := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	completedAt := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)

	t.Run("names指定時はname INで判定する", func(t *testing.T) {
		owner := testdb.MakeTestOwner(t, db, clinicA, "G11-5フード飼主1")
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
		owner := testdb.MakeTestOwner(t, db, clinicA, "G11-5フード飼主2")
		billing := makeTrimmingBillingWithCompletedAt(t, db, clinicA, model.BillingStatusCompleted, completedAt)
		billing.OwnerID = &owner.ID
		require.NoError(t, db.WithContext(ctx).Save(billing).Error)
		require.NoError(t, db.WithContext(ctx).Create(&model.BillingItem{BillingID: billing.ID, Category: model.ItemCategoryFood, Name: "任意フード名"}).Error)

		ok, err := repo.HasFoodPurchaseByOwnerSince(ctx, clinicA, owner.ID, since, nil)
		require.NoError(t, err)
		assert.True(t, ok, "names未指定時はcategory=foodで判定")

		nonFoodOwner := testdb.MakeTestOwner(t, db, clinicA, "G11-5非フード飼主")
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
// makeBillingWith（billing_test_fixtures_test.go）への thin wrapper。
func makeTrimmingBillingWithCompletedAt(t *testing.T, db *gorm.DB, clinicID uint64, status model.BillingStatus, completedAt time.Time) *model.Billing {
	t.Helper()
	return makeBillingWith(t, db, billingFixtureOpts{
		ClinicID:      clinicID,
		Status:        status,
		ScheduledDate: completedAt,
		CompletedAt:   &completedAt,
	})
}
