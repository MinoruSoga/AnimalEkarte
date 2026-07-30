package billing

// accounting_repository_reports_close_test.go — AccountingRepository.GetCloseAggregate の
// 統合テスト（実 Postgres テスト DB）。
//
// setupAccountingIsolationTestDB は accounting_repository_clinic_isolation_test.go で定義済みのため再利用する。

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

// setupCloseAggregateTestDB は GetCloseAggregate 系テスト用の DB を用意する。
func setupCloseAggregateTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	return setupAccountingIsolationTestDB(t)
}

// makeCloseBilling は完了済み会計を明細1件+支払い内訳1件付きで作成して返す。
func makeCloseBilling(t *testing.T, db *gorm.DB, clinicID uint64, completedAt time.Time, category model.ItemCategory, unitPrice int64, taxRate float64, method model.PaymentMethod, splitAmount int64) *model.Billing {
	t.Helper()
	ctx := context.Background()
	ca := completedAt
	b := &model.Billing{
		ClinicID:      clinicID,
		Status:        model.BillingStatusCompleted,
		CompletedAt:   &ca,
		ScheduledDate: time.Date(completedAt.Year(), completedAt.Month(), completedAt.Day(), 0, 0, 0, 0, time.UTC),
		TotalAmount:   splitAmount,
	}
	require.NoError(t, db.WithContext(ctx).Create(b).Error)

	item := &model.BillingItem{
		BillingID: b.ID,
		Category:  category,
		Name:      string(category),
		UnitPrice: unitPrice,
		Quantity:  1,
		TaxType:   model.TaxTypeExcluded,
		TaxRate:   taxRate,
	}
	require.NoError(t, db.WithContext(ctx).Create(item).Error)

	split := &model.PaymentSplit{
		ClinicID:  clinicID,
		BillingID: b.ID,
		Method:    method,
		Amount:    splitAmount,
	}
	require.NoError(t, db.WithContext(ctx).Create(split).Error)

	return b
}

func TestAccountingRepository_GetCloseAggregate_AggregatesPaymentsAndCategories(t *testing.T) {
	db := setupCloseAggregateTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	makeCloseBilling(t, db, clinicID, midJune(10), model.ItemCategoryExamination, 1000, 0.10, model.PaymentMethodCash, 1000)
	makeCloseBilling(t, db, clinicID, midJune(15), model.ItemCategoryMedicine, 2000, 0.08, model.PaymentMethodCreditCard, 2000)

	result, err := repo.GetCloseAggregate(ctx, GetCloseAggregateInput{
		ClinicID:    clinicID,
		PeriodStart: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)

	var paymentTotal int64
	for _, p := range result.PaymentRows {
		paymentTotal += p.Amount
	}
	assert.Equal(t, int64(3000), paymentTotal)

	var examTotal, medTotal int64
	for _, c := range result.CategoryRows {
		switch c.Category {
		case string(model.ItemCategoryExamination):
			examTotal += c.Amount
		case string(model.ItemCategoryMedicine):
			medTotal += c.Amount
		}
	}
	assert.Equal(t, int64(1000), examTotal)
	assert.Equal(t, int64(2000), medTotal)
	assert.Equal(t, int64(0), result.TotalRefund)

	var tax10, tax8 int64
	for _, tr := range result.TaxBreakdown {
		switch tr.TaxRate {
		case 10:
			tax10 = tr.TaxAmount
		case 8:
			tax8 = tr.TaxAmount
		}
	}
	assert.Equal(t, int64(100), tax10)
	assert.Equal(t, int64(160), tax8)
}

func TestAccountingRepository_GetCloseAggregate_RefundReducesNetAmountNotBillingAmount(t *testing.T) {
	db := setupCloseAggregateTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	b := makeCloseBilling(t, db, clinicID, midJune(10), model.ItemCategoryExamination, 5000, 0.10, model.PaymentMethodCash, 5000)
	refund := &model.BillingRefund{ClinicID: clinicID, BillingID: b.ID, Amount: 1200, RefundedAt: midJune(11)}
	require.NoError(t, db.WithContext(ctx).Create(refund).Error)

	result, err := repo.GetCloseAggregate(ctx, GetCloseAggregateInput{
		ClinicID:    clinicID,
		PeriodStart: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)

	assert.Equal(t, int64(1200), result.TotalRefund)
	require.Len(t, result.BillingDetails, 1)
	detail := result.BillingDetails[0]
	assert.Equal(t, b.ID, detail.BillingID)
	assert.Equal(t, int64(5000), detail.BillingAmount, "BillingAmount は返金前の会計額")
	assert.Equal(t, int64(1200), detail.RefundAmount)
	assert.Equal(t, int64(3800), detail.NetAmount, "NetAmount = BillingAmount - RefundAmount")
	assert.False(t, detail.IsHospitalization)
}

func TestAccountingRepository_GetCloseAggregate_MixedPaymentSplitsProduceMultipleDetailRows(t *testing.T) {
	db := setupCloseAggregateTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	ca := midJune(10)
	b := &model.Billing{
		ClinicID:      clinicID,
		Status:        model.BillingStatusCompleted,
		CompletedAt:   &ca,
		ScheduledDate: time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
		TotalAmount:   3000,
	}
	require.NoError(t, db.WithContext(ctx).Create(b).Error)
	item := &model.BillingItem{BillingID: b.ID, Category: model.ItemCategoryExamination, Name: "diagnosis", UnitPrice: 3000, Quantity: 1, TaxType: model.TaxTypeExcluded, TaxRate: 0.10}
	require.NoError(t, db.WithContext(ctx).Create(item).Error)

	splitCash := &model.PaymentSplit{ClinicID: clinicID, BillingID: b.ID, Method: model.PaymentMethodCash, Amount: 1000}
	splitCard := &model.PaymentSplit{ClinicID: clinicID, BillingID: b.ID, Method: model.PaymentMethodCreditCard, Amount: 2000}
	require.NoError(t, db.WithContext(ctx).Create(splitCash).Error)
	require.NoError(t, db.WithContext(ctx).Create(splitCard).Error)

	result, err := repo.GetCloseAggregate(ctx, GetCloseAggregateInput{
		ClinicID:    clinicID,
		PeriodStart: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)

	// 混在会計は payment_splits ベースで複数行になる（1会計だが支払い方法2件で明細2行）
	require.Len(t, result.BillingDetails, 2, "混在支払いは billing_id が同じ2明細行になる")
	var sum int64
	for _, d := range result.BillingDetails {
		assert.Equal(t, b.ID, d.BillingID)
		sum += d.BillingAmount
	}
	assert.Equal(t, int64(3000), sum)
}

func TestAccountingRepository_GetCloseAggregate_ClinicIsolation(t *testing.T) {
	db := setupCloseAggregateTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	makeCloseBilling(t, db, clinicA, midJune(10), model.ItemCategoryExamination, 1000, 0.10, model.PaymentMethodCash, 1000)
	makeCloseBilling(t, db, clinicB, midJune(10), model.ItemCategoryExamination, 9999, 0.10, model.PaymentMethodCash, 9999)

	input := GetCloseAggregateInput{
		PeriodStart: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}
	input.ClinicID = clinicA
	resultA, err := repo.GetCloseAggregate(ctx, input)
	require.NoError(t, err)
	var totalA int64
	for _, p := range resultA.PaymentRows {
		totalA += p.Amount
	}
	assert.Equal(t, int64(1000), totalA, "clinic B の会計が混入してはならない")

	input.ClinicID = clinicB
	resultB, err := repo.GetCloseAggregate(ctx, input)
	require.NoError(t, err)
	var totalB int64
	for _, p := range resultB.PaymentRows {
		totalB += p.Amount
	}
	assert.Equal(t, int64(9999), totalB)
}

func TestAccountingRepository_GetCloseAggregate_EmptyPeriodReturnsZeroValues(t *testing.T) {
	db := setupCloseAggregateTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	result, err := repo.GetCloseAggregate(ctx, GetCloseAggregateInput{
		ClinicID:    clinicID,
		PeriodStart: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)

	assert.Empty(t, result.PaymentRows)
	assert.Empty(t, result.CategoryRows)
	assert.Equal(t, int64(0), result.TotalRefund)
	assert.Empty(t, result.BillingDetails)
	assert.Empty(t, result.TaxBreakdown)
}

// TestAccountingRepository_GetCloseAggregate_BlanksForeignPetParentNames
// SEC-SWEEP-02-BILL-B1a-FIX: billings.pet_id が他院 pets を指す行は締め集計に残し、
// pet 名のみ clinic 相関 LEFT JOIN で空にする。金額: 1000+2000+5000=8000。
func TestAccountingRepository_GetCloseAggregate_BlanksForeignPetParentNames(t *testing.T) {
	db := setupCloseAggregateTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.AnimalSpecies{}, &model.Pet{}))
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	periodStart := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)

	// 正常系: pet_id 無しの完了会計 1000
	makeCloseBilling(t, db, clinicA, midJune(10), model.ItemCategoryExamination, 1000, 0.10, model.PaymentMethodCash, 1000)

	// 同一 clinic の pet を持つ完了会計 2000
	ownerA := testdb.MakeTestOwner(t, db, clinicA, "飼主A")
	species := &model.AnimalSpecies{Name: "犬-close"}
	require.NoError(t, db.WithContext(ctx).Create(species).Error)
	petA := &model.Pet{ClinicID: clinicA, OwnerID: ownerA.ID, AnimalSpeciesID: species.ID, Name: "同一院"}
	require.NoError(t, db.WithContext(ctx).Create(petA).Error)
	legit := makeCloseBilling(t, db, clinicA, midJune(12), model.ItemCategoryMedicine, 2000, 0.08, model.PaymentMethodCash, 2000)
	require.NoError(t, db.WithContext(ctx).Model(legit).Update("pet_id", petA.ID).Error)

	// clinicA の billing が clinicB の pet を pet_id に持つ（金額は残し、名前は空）
	ownerB := testdb.MakeTestOwner(t, db, clinicB, "飼主B")
	petB := &model.Pet{ClinicID: clinicB, OwnerID: ownerB.ID, AnimalSpeciesID: species.ID, Name: "他院pet"}
	require.NoError(t, db.WithContext(ctx).Create(petB).Error)
	crossClinicPet := makeCloseBilling(t, db, clinicA, midJune(15), model.ItemCategoryProcedure, 5000, 0.10, model.PaymentMethodCash, 5000)
	require.NoError(t, db.WithContext(ctx).Model(crossClinicPet).Update("pet_id", petB.ID).Error)

	result, err := repo.GetCloseAggregate(ctx, GetCloseAggregateInput{
		ClinicID: clinicA, PeriodStart: periodStart, PeriodEnd: periodEnd,
	})
	require.NoError(t, err)

	var paymentTotal int64
	for _, p := range result.PaymentRows {
		paymentTotal += p.Amount
	}
	// 1000 (no pet) + 2000 (same-clinic pet) + 5000 (foreign pet_id) = 8000
	assert.Equal(t, int64(8000), paymentTotal, "own-clinic billing with foreign pet_id remains in payment totals")

	byID := make(map[uint64]CloseBillingDetailRow, len(result.BillingDetails))
	for _, d := range result.BillingDetails {
		byID[d.BillingID] = d
	}
	require.Contains(t, byID, legit.ID, "same-clinic pet billing must remain in close details")
	assert.Equal(t, petA.Name, byID[legit.ID].PetName)

	require.Contains(t, byID, crossClinicPet.ID, "own-clinic billing with foreign pet_id must remain in close details")
	assert.Empty(t, byID[crossClinicPet.ID].PetName, "foreign-clinic pet name must be blanked")
}

// TestAccountingRepository_GetCloseAggregate_UnclassifiedOtherCountDistinct
// DEC-40: 未分類・要確認件数は category=other 明細を持つ会計の distinct 数。
// - 混在会計（examination + other）も 1 件として数える（MIN(category) では other が落ちる）
// - 同一会計の複数 other 明細 / 複数 payment_splits でも 1 件
// - 明細行数・detail 行数とは一致しない
func TestAccountingRepository_GetCloseAggregate_UnclassifiedOtherCountDistinct(t *testing.T) {
	db := setupCloseAggregateTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	periodStart := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)

	// 1) pure other × 2 payment splits → accounting distinct = 1（detail 行は 2）
	ca1 := midJune(10)
	pureOther := &model.Billing{
		ClinicID: clinicID, Status: model.BillingStatusCompleted, CompletedAt: &ca1,
		ScheduledDate: time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC), TotalAmount: 3000,
	}
	require.NoError(t, db.WithContext(ctx).Create(pureOther).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.BillingItem{
		BillingID: pureOther.ID, Category: model.ItemCategoryOther, Name: "その他A",
		UnitPrice: 3000, Quantity: 1, TaxType: model.TaxTypeExcluded, TaxRate: 0.10,
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.PaymentSplit{
		ClinicID: clinicID, BillingID: pureOther.ID, Method: model.PaymentMethodCash, Amount: 1000,
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.PaymentSplit{
		ClinicID: clinicID, BillingID: pureOther.ID, Method: model.PaymentMethodCreditCard, Amount: 2000,
	}).Error)

	// 2) mixed examination + other → MIN(category)='examination' だが other 件数に 1 を加算
	ca2 := midJune(11)
	mixed := &model.Billing{
		ClinicID: clinicID, Status: model.BillingStatusCompleted, CompletedAt: &ca2,
		ScheduledDate: time.Date(2026, 6, 11, 0, 0, 0, 0, time.UTC), TotalAmount: 4000,
	}
	require.NoError(t, db.WithContext(ctx).Create(mixed).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.BillingItem{
		BillingID: mixed.ID, Category: model.ItemCategoryExamination, Name: "診察",
		UnitPrice: 3000, Quantity: 1, TaxType: model.TaxTypeExcluded, TaxRate: 0.10,
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.BillingItem{
		BillingID: mixed.ID, Category: model.ItemCategoryOther, Name: "その他B",
		UnitPrice: 1000, Quantity: 1, TaxType: model.TaxTypeExcluded, TaxRate: 0.10,
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.PaymentSplit{
		ClinicID: clinicID, BillingID: mixed.ID, Method: model.PaymentMethodCash, Amount: 4000,
	}).Error)

	// 3) pure examination only → other 件数に含めない
	makeCloseBilling(t, db, clinicID, midJune(12), model.ItemCategoryExamination, 500, 0.10, model.PaymentMethodCash, 500)

	// 4) pure other with two other line items → still 1 accounting
	ca4 := midJune(13)
	multiOtherItems := &model.Billing{
		ClinicID: clinicID, Status: model.BillingStatusCompleted, CompletedAt: &ca4,
		ScheduledDate: time.Date(2026, 6, 13, 0, 0, 0, 0, time.UTC), TotalAmount: 700,
	}
	require.NoError(t, db.WithContext(ctx).Create(multiOtherItems).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.BillingItem{
		BillingID: multiOtherItems.ID, Category: model.ItemCategoryOther, Name: "その他C1",
		UnitPrice: 300, Quantity: 1, TaxType: model.TaxTypeExcluded, TaxRate: 0.10,
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.BillingItem{
		BillingID: multiOtherItems.ID, Category: model.ItemCategoryOther, Name: "その他C2",
		UnitPrice: 400, Quantity: 1, TaxType: model.TaxTypeExcluded, TaxRate: 0.10,
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.PaymentSplit{
		ClinicID: clinicID, BillingID: multiOtherItems.ID, Method: model.PaymentMethodCash, Amount: 700,
	}).Error)

	result, err := repo.GetCloseAggregate(ctx, GetCloseAggregateInput{
		ClinicID: clinicID, PeriodStart: periodStart, PeriodEnd: periodEnd,
	})
	require.NoError(t, err)

	// pureOther + mixed + multiOtherItems = 3（exam-only は除外）
	assert.Equal(t, int64(3), result.UnclassifiedOtherCount,
		"accounting distinct with any other item; not billing_items rows and not MIN(category) detail rows")

	// detail 行: pureOther の 2 splits + mixed 1 + exam 1 + multiOther 1 = 5
	// other として MIN(category) で分類される detail は pureOther の 2 行 + multiOther 1 行 = 3
	// （mixed は MIN='examination' のため other detail から落ちる）
	// 偶然 3==3 にならないよう、detail の other 行数は split 展開込みで数える:
	// pureOther 2 + multiOther 1 = 3、だが distinct は mixed を含む 3 で同数になり得る。
	// そこで「mixed を含む distinct」と「billing_items other 行数」の乖離を主に検証し、
	// detail 側は mixed が other に出ていないこと（MIN 歪み）を確認する。
	require.Len(t, result.BillingDetails, 5)
	var otherDetailRows int
	var mixedAppearsAsOther bool
	for _, d := range result.BillingDetails {
		if d.Category == string(model.ItemCategoryOther) {
			otherDetailRows++
			if d.BillingID == mixed.ID {
				mixedAppearsAsOther = true
			}
		}
	}
	assert.Equal(t, 3, otherDetailRows, "MIN(category) detail rows for other (excludes mixed; includes pureOther 2 splits)")
	assert.False(t, mixedAppearsAsOther, "mixed billing must not appear as other under MIN(category)")
	// mixed は distinct に含まれるが detail other には出ない → 件数の意味が異なる
	assert.Contains(t, []uint64{pureOther.ID, mixed.ID, multiOtherItems.ID}, mixed.ID)

	// billing_items の other 行数 = pureOther1 + mixed1 + multiOther2 = 4 ≠ distinct 3
	var otherItemRows int64
	require.NoError(t, db.WithContext(ctx).Raw(`
		SELECT COUNT(*) FROM billing_items
		WHERE category = ? AND deleted_at IS NULL
		  AND billing_id IN (SELECT id FROM billings WHERE clinic_id = ? AND deleted_at IS NULL)
	`, model.ItemCategoryOther, clinicID).Scan(&otherItemRows).Error)
	assert.Equal(t, int64(4), otherItemRows)
	assert.NotEqual(t, otherItemRows, result.UnclassifiedOtherCount,
		"distinct count must not equal billing_items row count")

	// payment-split 展開: pureOther の detail は 2 行だが accounting は 1
	var pureOtherDetailRows int
	for _, d := range result.BillingDetails {
		if d.BillingID == pureOther.ID {
			pureOtherDetailRows++
		}
	}
	assert.Equal(t, 2, pureOtherDetailRows, "mixed payment splits expand detail rows")
	assert.Equal(t, int64(3), result.UnclassifiedOtherCount, "distinct remains accounting-grain")
}

func TestAccountingRepository_GetCloseAggregate_UnclassifiedOtherCountEmpty(t *testing.T) {
	db := setupCloseAggregateTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	makeCloseBilling(t, db, clinicID, midJune(10), model.ItemCategoryExamination, 1000, 0.10, model.PaymentMethodCash, 1000)

	result, err := repo.GetCloseAggregate(ctx, GetCloseAggregateInput{
		ClinicID:    clinicID,
		PeriodStart: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)
	assert.Equal(t, int64(0), result.UnclassifiedOtherCount)
}
