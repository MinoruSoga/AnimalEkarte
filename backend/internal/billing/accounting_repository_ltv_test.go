package billing

// accounting_repository_ltv_test.go — AccountingRepository.SumPaidByOwner /
// MaxSingleVisitAmountByOwner / FindOwnersByAnnualRevenue の統合テスト（実 Postgres テスト DB）。
//
// 注: これらは Lステップタグ同期 / CPM 判定用の集計であり、LtvRepository.FindOwnerLTV とは
// 別の集計経路（ltv_repository_test.go 参照）。本ファイルは AccountingRepository 側を対象とする。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repotest"
)

func TestAccountingRepository_SumPaidByOwner_SumsCompletedOnlyForOwner(t *testing.T) {
	db := repotest.SetupTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := repotest.MakeTestOwner(t, db, clinicID, "LTV飼主")
	other := repotest.MakeTestOwner(t, db, clinicID, "別飼主")

	completed1 := &model.Billing{ClinicID: clinicID, OwnerID: &owner.ID, TotalAmount: 3000, Status: model.BillingStatusCompleted, ScheduledDate: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)}
	completed2 := &model.Billing{ClinicID: clinicID, OwnerID: &owner.ID, TotalAmount: 5000, Status: model.BillingStatusCompleted, ScheduledDate: time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC)}
	waiting := &model.Billing{ClinicID: clinicID, OwnerID: &owner.ID, TotalAmount: 9000, Status: model.BillingStatusWaiting, ScheduledDate: time.Date(2026, 6, 3, 0, 0, 0, 0, time.UTC)}
	otherOwnerBilling := &model.Billing{ClinicID: clinicID, OwnerID: &other.ID, TotalAmount: 7000, Status: model.BillingStatusCompleted, ScheduledDate: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)}
	require.NoError(t, db.WithContext(ctx).Create(completed1).Error)
	require.NoError(t, db.WithContext(ctx).Create(completed2).Error)
	require.NoError(t, db.WithContext(ctx).Create(waiting).Error)
	require.NoError(t, db.WithContext(ctx).Create(otherOwnerBilling).Error)

	total, err := repo.SumPaidByOwner(ctx, clinicID, owner.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(8000), total, "completed の3000+5000のみ・waiting/別飼主は除外")
}

func TestAccountingRepository_SumPaidByOwner_ZeroWhenNoCompletedBillings(t *testing.T) {
	db := repotest.SetupTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := repotest.MakeTestOwner(t, db, clinicID, "未会計飼主")
	total, err := repo.SumPaidByOwner(ctx, clinicID, owner.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
}

func TestAccountingRepository_SumPaidByOwner_ClinicIsolation(t *testing.T) {
	db := repotest.SetupTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := repotest.MakeTestOwner(t, db, clinicA, "医院A飼主")
	billing := &model.Billing{ClinicID: clinicA, OwnerID: &owner.ID, TotalAmount: 4000, Status: model.BillingStatusCompleted, ScheduledDate: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)}
	require.NoError(t, db.WithContext(ctx).Create(billing).Error)

	total, err := repo.SumPaidByOwner(ctx, clinicB, owner.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total, "別クリニックからは飼主の支払いが見えてはならない")
}

func TestAccountingRepository_MaxSingleVisitAmountByOwner_ReturnsMaxOfCompleted(t *testing.T) {
	db := repotest.SetupTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := repotest.MakeTestOwner(t, db, clinicID, "最大来院飼主")
	for _, amt := range []int64{10_000, 35_000, 8_000} {
		b := &model.Billing{ClinicID: clinicID, OwnerID: &owner.ID, TotalAmount: amt, Status: model.BillingStatusCompleted, ScheduledDate: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)}
		require.NoError(t, db.WithContext(ctx).Create(b).Error)
	}
	// pending は含まれない
	pending := &model.Billing{ClinicID: clinicID, OwnerID: &owner.ID, TotalAmount: 50_000, Status: model.BillingStatusPending, ScheduledDate: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)}
	require.NoError(t, db.WithContext(ctx).Create(pending).Error)

	maxAmount, err := repo.MaxSingleVisitAmountByOwner(ctx, clinicID, owner.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(35_000), maxAmount)
}

func TestAccountingRepository_MaxSingleVisitAmountByOwner_ZeroWhenNoCompletedBillings(t *testing.T) {
	db := repotest.SetupTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := repotest.MakeTestOwner(t, db, clinicID, "会計なし飼主")
	maxAmount, err := repo.MaxSingleVisitAmountByOwner(ctx, clinicID, owner.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), maxAmount)
}

func TestAccountingRepository_FindOwnersByAnnualRevenue_OrdersDescendingAndExcludesOldBillings(t *testing.T) {
	db := repotest.SetupTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	now := time.Now()

	high := repotest.MakeTestOwner(t, db, clinicID, "高額飼主")
	low := repotest.MakeTestOwner(t, db, clinicID, "低額飼主")

	// 直近365日以内（対象）
	b1 := &model.Billing{ClinicID: clinicID, OwnerID: &high.ID, TotalAmount: 30_000, Status: model.BillingStatusCompleted, ScheduledDate: now.AddDate(0, 0, -10), CompletedAt: timePtr(now.AddDate(0, 0, -10))}
	b2 := &model.Billing{ClinicID: clinicID, OwnerID: &low.ID, TotalAmount: 5_000, Status: model.BillingStatusCompleted, ScheduledDate: now.AddDate(0, 0, -5), CompletedAt: timePtr(now.AddDate(0, 0, -5))}
	require.NoError(t, db.WithContext(ctx).Create(b1).Error)
	require.NoError(t, db.WithContext(ctx).Create(b2).Error)

	// 365日超過（除外対象）— high の売上を追加してしまうと除外検証にならないため別飼主で作成
	tooOld := repotest.MakeTestOwner(t, db, clinicID, "365日超過飼主")
	oldBilling := &model.Billing{ClinicID: clinicID, OwnerID: &tooOld.ID, TotalAmount: 99_999, Status: model.BillingStatusCompleted, ScheduledDate: now.AddDate(0, 0, -400), CompletedAt: timePtr(now.AddDate(0, 0, -400))}
	require.NoError(t, db.WithContext(ctx).Create(oldBilling).Error)

	results, err := repo.FindOwnersByAnnualRevenue(ctx, clinicID)
	require.NoError(t, err)

	byOwner := make(map[uint64]int64, len(results))
	for _, r := range results {
		byOwner[r.OwnerID] = r.Revenue
	}
	assert.Equal(t, int64(30_000), byOwner[high.ID])
	assert.Equal(t, int64(5_000), byOwner[low.ID])
	_, oldPresent := byOwner[tooOld.ID]
	assert.False(t, oldPresent, "365日超過の売上は集計から除外される")

	require.Len(t, results, 2)
	assert.Equal(t, high.ID, results[0].OwnerID, "降順ソートで高額飼主が先頭")
}

func TestAccountingRepository_FindOwnersByAnnualRevenue_ExcludesNullOwnerID(t *testing.T) {
	db := repotest.SetupTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	now := time.Now()

	// owner_id が NULL の直接会計（除外対象）
	noOwner := &model.Billing{ClinicID: clinicID, TotalAmount: 12_000, Status: model.BillingStatusCompleted, ScheduledDate: now.AddDate(0, 0, -1), CompletedAt: timePtr(now.AddDate(0, 0, -1))}
	require.NoError(t, db.WithContext(ctx).Create(noOwner).Error)

	results, err := repo.FindOwnersByAnnualRevenue(ctx, clinicID)
	require.NoError(t, err)
	assert.Empty(t, results, "owner_id が NULL の会計は集計対象外")
}

func TestAccountingRepository_FindOwnersByAnnualRevenue_ClinicIsolation(t *testing.T) {
	db := repotest.SetupTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	now := time.Now()

	ownerA := repotest.MakeTestOwner(t, db, clinicA, "医院A飼主")
	ownerB := repotest.MakeTestOwner(t, db, clinicB, "医院B飼主")
	bA := &model.Billing{ClinicID: clinicA, OwnerID: &ownerA.ID, TotalAmount: 1_000, Status: model.BillingStatusCompleted, ScheduledDate: now.AddDate(0, 0, -1), CompletedAt: timePtr(now.AddDate(0, 0, -1))}
	bB := &model.Billing{ClinicID: clinicB, OwnerID: &ownerB.ID, TotalAmount: 9_000, Status: model.BillingStatusCompleted, ScheduledDate: now.AddDate(0, 0, -1), CompletedAt: timePtr(now.AddDate(0, 0, -1))}
	require.NoError(t, db.WithContext(ctx).Create(bA).Error)
	require.NoError(t, db.WithContext(ctx).Create(bB).Error)

	resultsA, err := repo.FindOwnersByAnnualRevenue(ctx, clinicA)
	require.NoError(t, err)
	require.Len(t, resultsA, 1)
	assert.Equal(t, ownerA.ID, resultsA[0].OwnerID)
	assert.Equal(t, int64(1_000), resultsA[0].Revenue)
}

func timePtr(t time.Time) *time.Time { return &t }
