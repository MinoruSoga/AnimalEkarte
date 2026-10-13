package billing

// emr354_close_billing_details_test.go — EMR-354 個別会計明細 category_amounts 回帰テスト
//
// テスト対象: AccountingRepository.GetCloseAggregate が返す BillingDetails の各行。
// EMR-354 の要件: split 明細行ごとに「その支払額を billing_items 重みで按分した
// GROSS 部門内訳（category_amounts）」を transport すること。
// 既存の allocation helpers（AllocateBillingPayments + MatrixRowTotals + 合成 method key）の
// 振る舞いを按分期待値の算定根拠とする。allocation 本体・集計マトリクスの改変は禁止。
//
// category_amounts は新規フィールドのため、本テストは Go struct ではなく
// 行を JSON marshal した wire shape に対して断言する。
// → ベースラインではキーが存在せず振る舞いとして失敗する（コンパイルエラーではない RED）。

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// emr354CategoryAmounts は明細行を JSON marshal し category_amounts を取り出す。
// フィールド未実装（JSON キー欠落）なら require で失敗する。
func emr354CategoryAmounts(t *testing.T, row CloseBillingDetailRow) map[string]int64 {
	t.Helper()
	raw, err := json.Marshal(row)
	require.NoError(t, err)
	var decoded map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(raw, &decoded))
	require.Contains(t, decoded, "category_amounts",
		"EMR-354: 明細行 JSON に category_amounts が必要（split 支払額の部門別 GROSS 按分）")
	var amounts map[string]int64
	require.NoError(t, json.Unmarshal(decoded["category_amounts"], &amounts))
	require.NotNil(t, amounts, "category_amounts は null ではなく map であるべき")
	return amounts
}

// emr354MakeCompletedBilling は completedAt で会計済みの最小 Billing を作成する。
// 既存 makeBillingWith は CompletedAt を扱えないため（completed クエリ条件を外す）、
// close テスト系と同じ直接作成パターンで fixtures を立てる。
func emr354MakeCompletedBilling(t *testing.T, db *gorm.DB, clinicID uint64, ownerID *uint64, total int64, completedAt time.Time) *model.Billing {
	t.Helper()
	billing := &model.Billing{
		ClinicID:      clinicID,
		OwnerID:       ownerID,
		TotalAmount:   total,
		Status:        model.BillingStatusCompleted,
		ScheduledDate: completedAt,
		CompletedAt:   &completedAt,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(billing).Error)
	return billing
}

func emr354AddItem(t *testing.T, db *gorm.DB, billingID uint64, category model.ItemCategory, unitPrice, discount int64) *model.BillingItem {
	t.Helper()
	item := &model.BillingItem{
		BillingID:      billingID,
		Category:       category,
		Name:           string(category) + "項目",
		UnitPrice:      unitPrice,
		Quantity:       1,
		DiscountAmount: discount,
		TaxType:        model.TaxTypeIncluded,
		TaxRate:        0.1,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(item).Error)
	return item
}

func emr354AddSplit(t *testing.T, db *gorm.DB, clinicID, billingID uint64, methodID *uint64, method model.PaymentMethod, amount int64) {
	t.Helper()
	split := &model.PaymentSplit{
		ClinicID:        clinicID,
		BillingID:       billingID,
		PaymentMethodID: methodID,
		Method:          method,
		Amount:          amount,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(split).Error)
}

func emr354MakePayMethod(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.PaymentMethodMaster {
	t.Helper()
	pm := &model.PaymentMethodMaster{ClinicID: clinicID, Name: name, IsActive: true}
	require.NoError(t, db.WithContext(context.Background()).Create(pm).Error)
	return pm
}

func emr354DetailByMethod(t *testing.T, details []CloseBillingDetailRow, methodID uint64) CloseBillingDetailRow {
	t.Helper()
	for _, d := range details {
		if d.PaymentMethodID != nil && *d.PaymentMethodID == methodID {
			return d
		}
	}
	t.Fatalf("payment_method_id=%d の明細行が見つからない（全%d行）", methodID, len(details))
	return CloseBillingDetailRow{}
}

// TestEMR354_GetCloseAggregate_CategoryAmounts_SplitLargestRemainder は
// 同一 billing の複数 split 各行に「行の支払額を item 重みで按分した GROSS 内訳」が乗ること、
// および割引後重み・最大剰余丸め・各行合計=行支払額を検証する。
//
// 重み: examination 3000, food (2000-500割引)=1500 → 合計4500。
// cash 1000円 split: 666.7/333.3 → 最大剰余で examination 667 / food 333。
// card 3000円 split: 2000/1000（割り切れ）。
func TestEMR354_GetCloseAggregate_CategoryAmounts_SplitLargestRemainder(t *testing.T) {
	db := setupAccountingIsolationTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	completedAt := midJune(1)
	owner := testdb.MakeTestOwner(t, db, clinicID, "山田花子")
	billing := emr354MakeCompletedBilling(t, db, clinicID, &owner.ID, 4000, completedAt)

	emr354AddItem(t, db, billing.ID, model.ItemCategoryExamination, 3000, 0)
	emr354AddItem(t, db, billing.ID, model.ItemCategoryFood, 2000, 500)

	cashMethod := emr354MakePayMethod(t, db, clinicID, "現金")
	cardMethod := emr354MakePayMethod(t, db, clinicID, "クレジットカード")
	emr354AddSplit(t, db, clinicID, billing.ID, &cashMethod.ID, model.PaymentMethodCash, 1000)
	emr354AddSplit(t, db, clinicID, billing.ID, &cardMethod.ID, model.PaymentMethodCreditCard, 3000)

	got, err := repo.GetCloseAggregate(ctx, GetCloseAggregateInput{
		ClinicID:    clinicID,
		PeriodStart: midJune(1),
		PeriodEnd:   midJune(2),
	})
	require.NoError(t, err)

	// 既存仕様の維持: split ごとに1行（行数を潰さない）。
	require.Len(t, got.BillingDetails, 2, "split ごとの明細行仕様は維持する")

	cashRow := emr354DetailByMethod(t, got.BillingDetails, cashMethod.ID)
	cardRow := emr354DetailByMethod(t, got.BillingDetails, cardMethod.ID)

	// GROSS 按分: 行支払額を item 重みで按分（最大剰余丸めは既存 helper の仕様）。
	assert.Equal(t, map[string]int64{
		"examination": 667,
		"food":        333,
	}, emr354CategoryAmounts(t, cashRow), "cash split の category_amounts 合計は行支払額1000に一致")
	assert.Equal(t, map[string]int64{
		"examination": 2000,
		"food":        1000,
	}, emr354CategoryAmounts(t, cardRow), "card split の category_amounts 合計は行支払額3000に一致")
}

// TestEMR354_GetCloseAggregate_CategoryAmounts_SoftDeletedItemExcluded は
// soft-delete 済み billing_items が重みに混入しないことを検証する。
func TestEMR354_GetCloseAggregate_CategoryAmounts_SoftDeletedItemExcluded(t *testing.T) {
	db := setupAccountingIsolationTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	completedAt := midJune(1)
	billing := emr354MakeCompletedBilling(t, db, clinicID, nil, 3000, completedAt)

	emr354AddItem(t, db, billing.ID, model.ItemCategoryExamination, 2000, 0)
	emr354AddItem(t, db, billing.ID, model.ItemCategorySurgery, 1000, 0)
	deleted := emr354AddItem(t, db, billing.ID, model.ItemCategoryFood, 900, 0)
	require.NoError(t, db.WithContext(ctx).Delete(deleted).Error)

	cashMethod := emr354MakePayMethod(t, db, clinicID, "現金")
	emr354AddSplit(t, db, clinicID, billing.ID, &cashMethod.ID, model.PaymentMethodCash, 3000)

	got, err := repo.GetCloseAggregate(ctx, GetCloseAggregateInput{
		ClinicID:    clinicID,
		PeriodStart: midJune(1),
		PeriodEnd:   midJune(2),
	})
	require.NoError(t, err)
	require.Len(t, got.BillingDetails, 1)

	// food(900) が重みに残ると 2000:1000:900 → 1538:769:692 に化ける。削除除外で 2000:1000。
	amounts := emr354CategoryAmounts(t, got.BillingDetails[0])
	assert.Equal(t, map[string]int64{
		"examination": 2000,
		"surgery":     1000,
	}, amounts)
	assert.NotContains(t, amounts, "food", "soft-delete 済み item の部門が出てはならない")
}

// TestEMR354_GetCloseAggregate_CategoryAmounts_ClinicIsolationAndBlankedNames は
// 他医院データの混入防止と、他医院 owner/pet 名のブランク化が category_amounts 追加後も維持される
// ことを検証する。
func TestEMR354_GetCloseAggregate_CategoryAmounts_ClinicIsolationAndBlankedNames(t *testing.T) {
	db := setupAccountingIsolationTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	completedAt := midJune(1)

	// clinic B: 別医院の billing（clinicA 結果に混入してはならない）
	ownerB := testdb.MakeTestOwner(t, db, clinicB, "別医院飼主")
	billingB := emr354MakeCompletedBilling(t, db, clinicB, &ownerB.ID, 5000, completedAt)
	emr354AddItem(t, db, billingB.ID, model.ItemCategoryExamination, 5000, 0)
	emr354AddSplit(t, db, clinicB, billingB.ID, nil, model.PaymentMethodCash, 5000)

	// clinic A: owner/pet が別医院所属（corrupt 参照）の billing。
	// 名前はブランク化されるが category_amounts 自体は自医院 item 重みで計算される。
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.Pet{}, &model.AnimalSpecies{}))
	species := &model.AnimalSpecies{Name: "cat"}
	require.NoError(t, db.WithContext(ctx).Create(species).Error)
	foreignPet := &model.Pet{ClinicID: clinicB, OwnerID: ownerB.ID, AnimalSpeciesID: species.ID, Name: "foreign-cat"}
	require.NoError(t, db.WithContext(ctx).Create(foreignPet).Error)

	billingA := emr354MakeCompletedBilling(t, db, clinicA, &ownerB.ID, 1500, completedAt)
	billingA.PetID = &foreignPet.ID
	require.NoError(t, db.WithContext(ctx).Save(billingA).Error)
	emr354AddItem(t, db, billingA.ID, model.ItemCategoryExamination, 1500, 0)
	emr354AddSplit(t, db, clinicA, billingA.ID, nil, model.PaymentMethodCash, 1500)

	got, err := repo.GetCloseAggregate(ctx, GetCloseAggregateInput{
		ClinicID:    clinicA,
		PeriodStart: midJune(1),
		PeriodEnd:   midJune(2),
	})
	require.NoError(t, err)

	require.Len(t, got.BillingDetails, 1, "clinicB の billing は clinicA の明細に出てはならない")
	row := got.BillingDetails[0]
	assert.Equal(t, billingA.ID, row.BillingID)
	assert.Equal(t, "", row.OwnerName, "他医院 owner 名はブランク化される（EMR-354 でも維持）")
	assert.Equal(t, "", row.PetName, "他医院 pet 名はブランク化される（EMR-354 でも維持）")
	assert.Equal(t, map[string]int64{"examination": 1500}, emr354CategoryAmounts(t, row),
		"corrupt 参照でも category_amounts は自医院 item 重みで計算される")
}

// TestEMR354_GetCloseAggregate_CategoryAmounts_RefundKeepsGross は
// 全期間返金スカラ（refund_amount/net_amount）と GROSS 按分が混ざらないことを検証する。
func TestEMR354_GetCloseAggregate_CategoryAmounts_RefundKeepsGross(t *testing.T) {
	db := setupAccountingIsolationTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	completedAt := midJune(1)
	billing := emr354MakeCompletedBilling(t, db, clinicID, nil, 3000, completedAt)
	emr354AddItem(t, db, billing.ID, model.ItemCategoryExamination, 3000, 0)
	cashMethod := emr354MakePayMethod(t, db, clinicID, "現金")
	emr354AddSplit(t, db, clinicID, billing.ID, &cashMethod.ID, model.PaymentMethodCash, 3000)

	refund := &model.BillingRefund{
		ClinicID:   clinicID,
		BillingID:  billing.ID,
		Amount:     1000,
		RefundedAt: midJune(1),
	}
	require.NoError(t, db.WithContext(ctx).Create(refund).Error)

	got, err := repo.GetCloseAggregate(ctx, GetCloseAggregateInput{
		ClinicID:    clinicID,
		PeriodStart: midJune(1),
		PeriodEnd:   midJune(2),
	})
	require.NoError(t, err)
	require.Len(t, got.BillingDetails, 1)

	row := got.BillingDetails[0]
	assert.Equal(t, int64(1000), row.RefundAmount, "全期間返金スカラは維持")
	assert.Equal(t, int64(2000), row.NetAmount, "net = 支払額 - 全期間返金（既存仕様）")
	assert.Equal(t, map[string]int64{"examination": 3000}, emr354CategoryAmounts(t, row),
		"category_amounts は GROSS（返金で減算しない）")
}

// TestEMR354_GetCloseAggregate_CategoryAmounts_NoItemsFallback は
// items を持たない billing の category_amounts が既存 helper の未分類フォールバック
// （other に全額）になることを確認する。
// ※ items なし時の扱いは Lead 確認事項。AllocateBillingPayments 既存仕様の踏襲を前提に置く。
func TestEMR354_GetCloseAggregate_CategoryAmounts_NoItemsFallback(t *testing.T) {
	db := setupAccountingIsolationTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	completedAt := midJune(1)
	billing := emr354MakeCompletedBilling(t, db, clinicID, nil, 500, completedAt)
	cashMethod := emr354MakePayMethod(t, db, clinicID, "現金")
	emr354AddSplit(t, db, clinicID, billing.ID, &cashMethod.ID, model.PaymentMethodCash, 500)

	got, err := repo.GetCloseAggregate(ctx, GetCloseAggregateInput{
		ClinicID:    clinicID,
		PeriodStart: midJune(1),
		PeriodEnd:   midJune(2),
	})
	require.NoError(t, err)
	require.Len(t, got.BillingDetails, 1)

	assert.Equal(t, map[string]int64{"other": 500}, emr354CategoryAmounts(t, got.BillingDetails[0]),
		"重み0件のとき既存 helper と同じく other へフォールバック（全額を MIN(category) へは寄せない）")
}

// TestEMR354_GetCloseAggregate_EmptyPeriod は対象期間に会計済み billing がないとき
// BillingDetails が空で返る既存の空ケースを維持することを確認する。
func TestEMR354_GetCloseAggregate_EmptyPeriod(t *testing.T) {
	db := setupAccountingIsolationTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	got, err := repo.GetCloseAggregate(ctx, GetCloseAggregateInput{
		ClinicID:    clinicID,
		PeriodStart: midJune(1),
		PeriodEnd:   midJune(2),
	})
	require.NoError(t, err)
	assert.Empty(t, got.BillingDetails)
}
