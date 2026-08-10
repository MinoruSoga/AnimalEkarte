package billing

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

const (
	firstDay2026June = "2026-06-01"
	lastDay2026June  = "2026-06-30"
)

// setupUnpaidCarryoverTestDB は setupTestDB を拡張し、pets / animal_species テーブルも整備する。
// FindMonthlyUnpaidCarryover は LEFT JOIN pets を使うため pets テーブルが必要。
func setupUnpaidCarryoverTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	// billings はすでに setupTestDB で truncate 済みなので CASCADE 安全
	if err := testdb.EnsureAutoMigrated(db, &model.AnimalSpecies{}, &model.Pet{}); err != nil {
		t.Fatalf("failed to migrate animal_species/pets: %v", err)
	}
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	return db
}

// makeBilling はテスト用の Billing を作成する。makeBillingWith（billing_test_fixtures_test.go）への thin wrapper。
func makeBilling(t *testing.T, db *gorm.DB, clinicID uint64, ownerID, petID *uint64, amount int64, status model.BillingStatus, scheduledDate time.Time) {
	t.Helper()
	makeBillingWith(t, db, billingFixtureOpts{
		ClinicID:      clinicID,
		OwnerID:       ownerID,
		PetID:         petID,
		TotalAmount:   amount,
		Status:        status,
		ScheduledDate: scheduledDate,
	})
}

// makeSpeciesAndPet はテスト用の AnimalSpecies と Pet を作成して返す。
func makeSpeciesAndPet(t *testing.T, db *gorm.DB, clinicID, ownerID uint64, petName string) *model.Pet {
	t.Helper()
	species := &model.AnimalSpecies{Name: "犬"}
	require.NoError(t, db.WithContext(context.Background()).Create(species).Error)
	pet := &model.Pet{
		ClinicID:        clinicID,
		OwnerID:         ownerID,
		AnimalSpeciesID: species.ID,
		Name:            petName,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(pet).Error)
	return pet
}

// TestSumUnpaidByOwner は飼主単位の未納残高が status=waiting のみ・soft-delete 除外・
// クリニック/飼主スコープで正しく集計されることを検証する。#182
func TestSumUnpaidByOwner(t *testing.T) {
	db := testdb.SetupTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	jun := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)

	owner := testdb.MakeTestOwner(t, db, clinicID, "未納 太郎")
	other := testdb.MakeTestOwner(t, db, clinicID, "別 花子")

	// 対象飼主: waiting 1234 + 866（未納=2100）、completed 5000（除外）、cancelled 999（除外）
	makeBilling(t, db, clinicID, &owner.ID, nil, 1234, model.BillingStatusWaiting, jun)
	makeBilling(t, db, clinicID, &owner.ID, nil, 866, model.BillingStatusWaiting, jun)
	makeBilling(t, db, clinicID, &owner.ID, nil, 5000, model.BillingStatusCompleted, jun)
	makeBilling(t, db, clinicID, &owner.ID, nil, 999, model.BillingStatusCancelled, jun)
	// 別飼主の waiting（対象飼主の残高に混入しないこと）
	makeBilling(t, db, clinicID, &other.ID, nil, 7777, model.BillingStatusWaiting, jun)

	t.Run("waiting のみ合算し completed/cancelled/別飼主を除外", func(t *testing.T) {
		got, err := repo.SumUnpaidByOwner(ctx, clinicID, owner.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(2100), got.TotalAmount, "1234+866 の waiting のみ")
		assert.Equal(t, int64(2), got.Count)
	})

	t.Run("未納なしは 0/0（空値ケース）", func(t *testing.T) {
		empty := testdb.MakeTestOwner(t, db, clinicID, "完納 次郎")
		makeBilling(t, db, clinicID, &empty.ID, nil, 3000, model.BillingStatusCompleted, jun)
		got, err := repo.SumUnpaidByOwner(ctx, clinicID, empty.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), got.TotalAmount)
		assert.Equal(t, int64(0), got.Count)
	})

	t.Run("別クリニックスコープでは混入しない", func(t *testing.T) {
		got, err := repo.SumUnpaidByOwner(ctx, 999, owner.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), got.TotalAmount)
		assert.Equal(t, int64(0), got.Count)
	})
}

// TestFindMonthlyUnpaidCarryover_PrevCurrentSplit は
// 前月/当月の分割と対象月以降の除外を検証する。#114
func TestFindMonthlyUnpaidCarryover_PrevCurrentSplit(t *testing.T) {
	db := setupUnpaidCarryoverTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	clinicID := uint64(1)

	owner := testdb.MakeTestOwner(t, db, clinicID, "分割テスト飼主")
	id := owner.ID

	// 前月(5月): 1000円 → prev_month_carryover
	makeBilling(t, db, clinicID, &id, nil, 1000, model.BillingStatusWaiting,
		time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC))
	// 当月(6月): 2000円 → current_month_unpaid
	makeBilling(t, db, clinicID, &id, nil, 2000, model.BillingStatusWaiting,
		time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC))
	// 翌月(7月): 3000円 → scheduled_date > lastDay なので除外
	makeBilling(t, db, clinicID, &id, nil, 3000, model.BillingStatusWaiting,
		time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC))

	items, total, summary, err := repo.FindMonthlyUnpaidCarryover(ctx, clinicID, firstDay2026June, lastDay2026June, 1, 100)
	require.NoError(t, err)

	assert.Equal(t, int64(1000), summary.PrevMonthCarryover, "5月分は前月繰越")
	assert.Equal(t, int64(2000), summary.CurrentMonthUnpaid, "6月分は当月未払い")
	assert.Equal(t, int64(3000), summary.NextMonthCarryover, "次月繰越=前月+当月")
	assert.Equal(t, int64(1), total, "owner+pet組合せ数=1")
	require.Len(t, items, 1)
	assert.Equal(t, int64(1000), items[0].PrevMonthCarryover)
	assert.Equal(t, int64(2000), items[0].CurrentMonthUnpaid)
	assert.Equal(t, int64(3000), items[0].NextMonthCarryover)
	assert.Equal(t, id, items[0].OwnerID)
	assert.Nil(t, items[0].PetID, "pet_id はnil")
	assert.Equal(t, "", items[0].PetName, "pet_name は空文字")
}

// TestFindMonthlyUnpaidCarryover_BoundaryDates は firstDay/lastDay の境界値を検証する。
// firstDay-1 → prev, firstDay → current, lastDay → current, lastDay+1 → 除外
func TestFindMonthlyUnpaidCarryover_BoundaryDates(t *testing.T) {
	db := setupUnpaidCarryoverTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	clinicID := uint64(1)

	owner := testdb.MakeTestOwner(t, db, clinicID, "境界値テスト飼主")
	id := owner.ID

	// 5/31 (firstDay-1): prev
	makeBilling(t, db, clinicID, &id, nil, 100, model.BillingStatusWaiting,
		time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC))
	// 6/1 (firstDay): current
	makeBilling(t, db, clinicID, &id, nil, 200, model.BillingStatusWaiting,
		time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
	// 6/30 (lastDay): current
	makeBilling(t, db, clinicID, &id, nil, 300, model.BillingStatusWaiting,
		time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC))
	// 7/1 (lastDay+1): 除外
	makeBilling(t, db, clinicID, &id, nil, 400, model.BillingStatusWaiting,
		time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC))

	_, _, summary, err := repo.FindMonthlyUnpaidCarryover(ctx, clinicID, firstDay2026June, lastDay2026June, 1, 100)
	require.NoError(t, err)

	assert.Equal(t, int64(100), summary.PrevMonthCarryover, "5/31 は前月繰越に入る")
	assert.Equal(t, int64(500), summary.CurrentMonthUnpaid, "6/1+6/30=500 は当月未払い")
	assert.Equal(t, int64(600), summary.NextMonthCarryover, "次月繰越=100+500、7/1は除外")
}

// TestFindMonthlyUnpaidCarryover_StatusFilter は status=waiting のみが集計対象であることを検証する。
func TestFindMonthlyUnpaidCarryover_StatusFilter(t *testing.T) {
	db := setupUnpaidCarryoverTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	clinicID := uint64(1)

	owner := testdb.MakeTestOwner(t, db, clinicID, "ステータスフィルタ飼主")
	id := owner.ID
	date := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)

	makeBilling(t, db, clinicID, &id, nil, 1000, model.BillingStatusWaiting, date)
	makeBilling(t, db, clinicID, &id, nil, 9000, model.BillingStatusCompleted, date)
	makeBilling(t, db, clinicID, &id, nil, 9000, model.BillingStatusCancelled, date)
	makeBilling(t, db, clinicID, &id, nil, 9000, model.BillingStatusPending, date)

	items, total, summary, err := repo.FindMonthlyUnpaidCarryover(ctx, clinicID, firstDay2026June, lastDay2026June, 1, 100)
	require.NoError(t, err)

	assert.Equal(t, int64(0), summary.PrevMonthCarryover)
	assert.Equal(t, int64(1000), summary.CurrentMonthUnpaid, "waiting の1000円のみ集計")
	assert.Equal(t, int64(1000), summary.NextMonthCarryover)
	assert.Equal(t, int64(1), total)
	assert.Len(t, items, 1)
}

// TestFindMonthlyUnpaidCarryover_NullPetID は pet_id=NULL の billing でも正常に動作することを検証する。
func TestFindMonthlyUnpaidCarryover_NullPetID(t *testing.T) {
	db := setupUnpaidCarryoverTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	clinicID := uint64(1)

	owner := testdb.MakeTestOwner(t, db, clinicID, "ペットなし飼主")
	id := owner.ID
	makeBilling(t, db, clinicID, &id, nil, 5000, model.BillingStatusWaiting,
		time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC))

	items, total, summary, err := repo.FindMonthlyUnpaidCarryover(ctx, clinicID, firstDay2026June, lastDay2026June, 1, 100)
	require.NoError(t, err)

	assert.Equal(t, int64(5000), summary.CurrentMonthUnpaid)
	assert.Equal(t, int64(1), total)
	require.Len(t, items, 1)
	assert.Nil(t, items[0].PetID, "pet_id は nil")
	assert.Equal(t, "", items[0].PetName, "pet_name は空文字 (COALESCE)")
}

// TestFindMonthlyUnpaidCarryover_MultipleOwnersPets は
// 複数飼主・複数ペット・ペットなし billing の正しいグルーピングを検証する。
func TestFindMonthlyUnpaidCarryover_MultipleOwnersPets(t *testing.T) {
	db := setupUnpaidCarryoverTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	clinicID := uint64(1)

	owner1 := testdb.MakeTestOwner(t, db, clinicID, "飼主1")
	owner2 := testdb.MakeTestOwner(t, db, clinicID, "飼主2")
	pet1 := makeSpeciesAndPet(t, db, clinicID, owner1.ID, "ぽち")
	pet2 := makeSpeciesAndPet(t, db, clinicID, owner2.ID, "たま")

	date := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	// owner1+pet1: 2000
	makeBilling(t, db, clinicID, &owner1.ID, &pet1.ID, 2000, model.BillingStatusWaiting, date)
	// owner1+nil: 1000 (同じ飼主でもペットなし = 別グループ)
	makeBilling(t, db, clinicID, &owner1.ID, nil, 1000, model.BillingStatusWaiting, date)
	// owner2+pet2: 3000
	makeBilling(t, db, clinicID, &owner2.ID, &pet2.ID, 3000, model.BillingStatusWaiting, date)

	items, total, summary, err := repo.FindMonthlyUnpaidCarryover(ctx, clinicID, firstDay2026June, lastDay2026June, 1, 100)
	require.NoError(t, err)

	assert.Equal(t, int64(6000), summary.CurrentMonthUnpaid, "全合計=2000+1000+3000")
	assert.Equal(t, int64(3), total, "owner+pet組合せ=3グループ")
	assert.Len(t, items, 3)

	// ペット名を持つグループが正しく入っていることを確認
	var hasPochi, hasTama bool
	for _, it := range items {
		if it.PetName == "ぽち" {
			hasPochi = true
			assert.Equal(t, int64(2000), it.CurrentMonthUnpaid)
		}
		if it.PetName == "たま" {
			hasTama = true
			assert.Equal(t, int64(3000), it.CurrentMonthUnpaid)
		}
	}
	assert.True(t, hasPochi, "ぽち が結果に含まれる")
	assert.True(t, hasTama, "たま が結果に含まれる")
}

// TestFindMonthlyUnpaidCarryover_Pagination は total_count と取得件数の整合を検証する。
func TestFindMonthlyUnpaidCarryover_Pagination(t *testing.T) {
	db := setupUnpaidCarryoverTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	clinicID := uint64(1)

	date := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	for i := range 5 {
		owner := testdb.MakeTestOwner(t, db, clinicID, fmt.Sprintf("ページ飼主%d", i))
		makeBilling(t, db, clinicID, &owner.ID, nil, 1000, model.BillingStatusWaiting, date)
	}

	// page=1, limit=2
	items1, total1, _, err := repo.FindMonthlyUnpaidCarryover(ctx, clinicID, firstDay2026June, lastDay2026June, 1, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(5), total1, "total は全件数")
	assert.Len(t, items1, 2, "page1 limit2 → 2件")

	// page=2, limit=2
	items2, total2, _, err := repo.FindMonthlyUnpaidCarryover(ctx, clinicID, firstDay2026June, lastDay2026June, 2, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(5), total2, "total は変わらず5件")
	assert.Len(t, items2, 2, "page2 limit2 → 2件")

	// page=3, limit=2 (最終ページ: 残り1件)
	items3, total3, _, err := repo.FindMonthlyUnpaidCarryover(ctx, clinicID, firstDay2026June, lastDay2026June, 3, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(5), total3, "total は変わらず5件")
	assert.Len(t, items3, 1, "page3 limit2 → 残り1件")
}

// TestFindMonthlyUnpaidCarryover_ClinicIDIsolation は clinic_id による隔離を検証する。
func TestFindMonthlyUnpaidCarryover_ClinicIDIsolation(t *testing.T) {
	db := setupUnpaidCarryoverTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()

	clinicID1 := uint64(1)
	clinicID2 := uint64(2)
	date := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)

	owner1 := testdb.MakeTestOwner(t, db, clinicID1, "医院1飼主")
	owner2 := testdb.MakeTestOwner(t, db, clinicID2, "医院2飼主")
	makeBilling(t, db, clinicID1, &owner1.ID, nil, 1000, model.BillingStatusWaiting, date)
	makeBilling(t, db, clinicID2, &owner2.ID, nil, 9000, model.BillingStatusWaiting, date)

	items1, total1, summary1, err := repo.FindMonthlyUnpaidCarryover(ctx, clinicID1, firstDay2026June, lastDay2026June, 1, 100)
	require.NoError(t, err)
	assert.Equal(t, int64(1000), summary1.CurrentMonthUnpaid, "医院1は1000のみ")
	assert.Equal(t, int64(1), total1)
	assert.Len(t, items1, 1)

	items2, total2, summary2, err := repo.FindMonthlyUnpaidCarryover(ctx, clinicID2, firstDay2026June, lastDay2026June, 1, 100)
	require.NoError(t, err)
	assert.Equal(t, int64(9000), summary2.CurrentMonthUnpaid, "医院2は9000のみ")
	assert.Equal(t, int64(1), total2)
	assert.Len(t, items2, 1)
}

// TestFindMonthlyUnpaidCarryover_EmptyResult は対象データが存在しない場合にゼロ値を返すことを検証する。
func TestFindMonthlyUnpaidCarryover_EmptyResult(t *testing.T) {
	db := setupUnpaidCarryoverTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	clinicID := uint64(1)

	items, total, summary, err := repo.FindMonthlyUnpaidCarryover(ctx, clinicID, firstDay2026June, lastDay2026June, 1, 100)
	require.NoError(t, err)

	assert.Equal(t, int64(0), summary.PrevMonthCarryover)
	assert.Equal(t, int64(0), summary.CurrentMonthUnpaid)
	assert.Equal(t, int64(0), summary.NextMonthCarryover)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, items)
}

// TestAccountingRepository_FindUnpaidByBilling_ReturnsWaitingWithinRangeOnly は
// status=waiting かつ scheduled_date が範囲内の billing のみが会計単位で返ることを検証する。#120
func TestAccountingRepository_FindUnpaidByBilling_ReturnsWaitingWithinRangeOnly(t *testing.T) {
	db := setupUnpaidCarryoverTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	clinicID := uint64(1)

	owner := testdb.MakeTestOwner(t, db, clinicID, "会計単位テスト飼主")
	id := owner.ID

	makeBilling(t, db, clinicID, &id, nil, 1000, model.BillingStatusWaiting, time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC))
	makeBilling(t, db, clinicID, &id, nil, 2000, model.BillingStatusCompleted, time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC))
	makeBilling(t, db, clinicID, &id, nil, 3000, model.BillingStatusWaiting, time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC))

	billings, total, err := repo.FindUnpaidByBilling(ctx, clinicID, firstDay2026June, lastDay2026June, 1, 100)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total, "waiting かつ範囲内は1件のみ")
	require.Len(t, billings, 1)
	assert.Equal(t, int64(1000), billings[0].TotalAmount)
	assert.NotNil(t, billings[0].Owner, "Owner が Preload される")
}

// TestAccountingRepository_FindUnpaidByBilling_ClinicIsolation は clinic_id 隔離を検証する。
func TestAccountingRepository_FindUnpaidByBilling_ClinicIsolation(t *testing.T) {
	db := setupUnpaidCarryoverTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	clinicA, clinicB := uint64(1), uint64(2)

	ownerA := testdb.MakeTestOwner(t, db, clinicA, "医院A")
	ownerB := testdb.MakeTestOwner(t, db, clinicB, "医院B")
	makeBilling(t, db, clinicA, &ownerA.ID, nil, 1000, model.BillingStatusWaiting, time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC))
	makeBilling(t, db, clinicB, &ownerB.ID, nil, 9000, model.BillingStatusWaiting, time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC))

	billingsA, totalA, err := repo.FindUnpaidByBilling(ctx, clinicA, firstDay2026June, lastDay2026June, 1, 100)
	require.NoError(t, err)
	assert.Equal(t, int64(1), totalA)
	require.Len(t, billingsA, 1)
	assert.Equal(t, int64(1000), billingsA[0].TotalAmount, "clinic B の未納が混入してはならない")
}

// TestAccountingRepository_FindUnpaidByBilling_Pagination はページネーションを検証する。
func TestAccountingRepository_FindUnpaidByBilling_Pagination(t *testing.T) {
	db := setupUnpaidCarryoverTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	clinicID := uint64(1)

	for i := range 3 {
		owner := testdb.MakeTestOwner(t, db, clinicID, fmt.Sprintf("会計ページ飼主%d", i))
		makeBilling(t, db, clinicID, &owner.ID, nil, 1000, model.BillingStatusWaiting, time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC))
	}

	billings, total, err := repo.FindUnpaidByBilling(ctx, clinicID, firstDay2026June, lastDay2026June, 1, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, billings, 2, "page1 limit2 → 2件")
}

// TestAccountingRepository_FindUnpaidByOwner_AggregatesByOwnerAndSummarizes は
// 飼主単位の集約とサマリー（合計・件数・飼主数）が正しく算出されることを検証する。#120
func TestAccountingRepository_FindUnpaidByOwner_AggregatesByOwnerAndSummarizes(t *testing.T) {
	db := setupUnpaidCarryoverTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	clinicID := uint64(1)

	owner1 := testdb.MakeTestOwner(t, db, clinicID, "未納集約飼主1")
	owner2 := testdb.MakeTestOwner(t, db, clinicID, "未納集約飼主2")

	makeBilling(t, db, clinicID, &owner1.ID, nil, 1000, model.BillingStatusWaiting, time.Date(2026, 6, 5, 0, 0, 0, 0, time.UTC))
	makeBilling(t, db, clinicID, &owner1.ID, nil, 500, model.BillingStatusWaiting, time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC))
	makeBilling(t, db, clinicID, &owner2.ID, nil, 3000, model.BillingStatusWaiting, time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC))
	// completed は集計から除外
	makeBilling(t, db, clinicID, &owner1.ID, nil, 9999, model.BillingStatusCompleted, time.Date(2026, 6, 5, 0, 0, 0, 0, time.UTC))

	aggregates, totalOwners, summary, err := repo.FindUnpaidByOwner(ctx, clinicID, firstDay2026June, lastDay2026June, 1, 100)
	require.NoError(t, err)

	assert.Equal(t, int64(2), totalOwners, "未納がある飼主数=2")
	assert.Equal(t, int64(4500), summary.TotalAmount, "1000+500+3000")
	assert.Equal(t, int64(3), summary.BillingCount)
	assert.Equal(t, int64(2), summary.OwnerCount)

	byOwner := make(map[uint64]UnpaidOwnerAggregate, len(aggregates))
	for _, a := range aggregates {
		byOwner[a.OwnerID] = a
	}
	require.Contains(t, byOwner, owner1.ID)
	assert.Equal(t, int64(1500), byOwner[owner1.ID].TotalAmount, "飼主1は1000+500")
	assert.Equal(t, int64(2), byOwner[owner1.ID].Count)
	assert.Equal(t, int64(3000), byOwner[owner2.ID].TotalAmount)
}

// TestAccountingRepository_FindUnpaidByOwner_ClinicIsolation は clinic_id 隔離を検証する。
func TestAccountingRepository_FindUnpaidByOwner_ClinicIsolation(t *testing.T) {
	db := setupUnpaidCarryoverTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	clinicA, clinicB := uint64(1), uint64(2)

	ownerA := testdb.MakeTestOwner(t, db, clinicA, "医院A飼主")
	ownerB := testdb.MakeTestOwner(t, db, clinicB, "医院B飼主")
	makeBilling(t, db, clinicA, &ownerA.ID, nil, 1000, model.BillingStatusWaiting, time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC))
	makeBilling(t, db, clinicB, &ownerB.ID, nil, 9000, model.BillingStatusWaiting, time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC))

	_, totalA, summaryA, err := repo.FindUnpaidByOwner(ctx, clinicA, firstDay2026June, lastDay2026June, 1, 100)
	require.NoError(t, err)
	assert.Equal(t, int64(1), totalA)
	assert.Equal(t, int64(1000), summaryA.TotalAmount, "clinic B の未納が混入してはならない")
}

// TestAccountingRepository_FindUnpaidByOwner_EmptyResult は該当データなしの場合にゼロ値を返すことを検証する。
func TestAccountingRepository_FindUnpaidByOwner_EmptyResult(t *testing.T) {
	db := setupUnpaidCarryoverTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	clinicID := uint64(1)

	aggregates, total, summary, err := repo.FindUnpaidByOwner(ctx, clinicID, firstDay2026June, lastDay2026June, 1, 100)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Equal(t, int64(0), summary.TotalAmount)
	assert.Equal(t, int64(0), summary.BillingCount)
	assert.Equal(t, int64(0), summary.OwnerCount)
	assert.Empty(t, aggregates)
}

// TestUnpaidAggregates_KeepsListingAfterPetTransfer は DEC-27 境界を検証する。
// pets.owner_id が譲渡で変わっても、billings.owner_id スナップショット側の
// 未納一覧・残高・月次繰越が clinic 一致時に落ちないこと。
func TestUnpaidAggregates_KeepsListingAfterPetTransfer(t *testing.T) {
	db := setupUnpaidCarryoverTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	date := time.Date(2026, 6, 12, 0, 0, 0, 0, time.UTC)

	originalOwner := testdb.MakeTestOwner(t, db, clinicID, "未納譲渡前飼主")
	newOwner := testdb.MakeTestOwner(t, db, clinicID, "未納譲渡後飼主")
	pet := makeSpeciesAndPet(t, db, clinicID, originalOwner.ID, "unpaid-transfer-pet")
	makeBilling(t, db, clinicID, &originalOwner.ID, &pet.ID, 4_200, model.BillingStatusWaiting, date)

	require.NoError(t, db.WithContext(ctx).Model(&model.Pet{}).
		Where("id = ?", pet.ID).
		Update("owner_id", newOwner.ID).Error)

	t.Run("FindUnpaidByBilling still lists snapshot owner billing", func(t *testing.T) {
		billings, total, err := repo.FindUnpaidByBilling(ctx, clinicID, firstDay2026June, lastDay2026June, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, billings, 1)
		assert.Equal(t, int64(4_200), billings[0].TotalAmount)
		require.NotNil(t, billings[0].OwnerID)
		assert.Equal(t, originalOwner.ID, *billings[0].OwnerID)
		require.NotNil(t, billings[0].PetID)
		assert.Equal(t, pet.ID, *billings[0].PetID)
	})

	t.Run("FindUnpaidByOwner and SumUnpaidByOwner keep snapshot attribution", func(t *testing.T) {
		aggregates, totalOwners, summary, err := repo.FindUnpaidByOwner(ctx, clinicID, firstDay2026June, lastDay2026June, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(1), totalOwners)
		assert.Equal(t, int64(4_200), summary.TotalAmount)
		require.Len(t, aggregates, 1)
		assert.Equal(t, originalOwner.ID, aggregates[0].OwnerID)
		assert.Equal(t, int64(4_200), aggregates[0].TotalAmount)

		balance, err := repo.SumUnpaidByOwner(ctx, clinicID, originalOwner.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(4_200), balance.TotalAmount)
		assert.Equal(t, int64(1), balance.Count)

		newBalance, err := repo.SumUnpaidByOwner(ctx, clinicID, newOwner.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), newBalance.TotalAmount)
		assert.Equal(t, int64(0), newBalance.Count)
	})

	t.Run("FindMonthlyUnpaidCarryover keeps pet join after transfer", func(t *testing.T) {
		items, total, summary, err := repo.FindMonthlyUnpaidCarryover(ctx, clinicID, firstDay2026June, lastDay2026June, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(4_200), summary.CurrentMonthUnpaid)
		assert.Equal(t, int64(1), total)
		require.Len(t, items, 1)
		assert.Equal(t, originalOwner.ID, items[0].OwnerID)
		require.NotNil(t, items[0].PetID)
		assert.Equal(t, pet.ID, *items[0].PetID)
		assert.Equal(t, "unpaid-transfer-pet", items[0].PetName)
		assert.Equal(t, int64(4_200), items[0].CurrentMonthUnpaid)
	})
}

// TestFindMonthlyUnpaidCarryover_NextMonthCarryoverEquality は
// next_month_carryover = prev_month_carryover + current_month_unpaid の等式を検証する。
func TestFindMonthlyUnpaidCarryover_NextMonthCarryoverEquality(t *testing.T) {
	db := setupUnpaidCarryoverTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	clinicID := uint64(1)

	owner := testdb.MakeTestOwner(t, db, clinicID, "等式検証飼主")
	id := owner.ID

	makeBilling(t, db, clinicID, &id, nil, 1100, model.BillingStatusWaiting,
		time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC))
	makeBilling(t, db, clinicID, &id, nil, 2200, model.BillingStatusWaiting,
		time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC))
	makeBilling(t, db, clinicID, &id, nil, 4400, model.BillingStatusWaiting,
		time.Date(2026, 6, 5, 0, 0, 0, 0, time.UTC))
	makeBilling(t, db, clinicID, &id, nil, 5500, model.BillingStatusWaiting,
		time.Date(2026, 6, 25, 0, 0, 0, 0, time.UTC))

	_, _, summary, err := repo.FindMonthlyUnpaidCarryover(ctx, clinicID, firstDay2026June, lastDay2026June, 1, 100)
	require.NoError(t, err)

	assert.Equal(t, int64(3300), summary.PrevMonthCarryover, "前月繰越=3300")
	assert.Equal(t, int64(9900), summary.CurrentMonthUnpaid, "当月未払い=9900")
	assert.Equal(t, summary.PrevMonthCarryover+summary.CurrentMonthUnpaid, summary.NextMonthCarryover,
		"next_month_carryover は prev + current の和でなければならない")
}

// TestUnpaidIncludesCreditCorrectionResidual_BUG007 はクレジット訂正で生じた
// completed 会計の差額が未納者一覧・飼主残高に載ることを検証する（BUG-007）。
// 通常の全額精算済み completed は依然として未収に出ないこと（回帰）。
func TestUnpaidIncludesCreditCorrectionResidual_BUG007(t *testing.T) {
	db := setupUnpaidCarryoverTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	clinicID := uint64(1)

	owner := testdb.MakeTestOwner(t, db, clinicID, "クレジット訂正未収飼主")
	oid := owner.ID
	date := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)

	// waiting 全額 500（従来経路）
	makeBilling(t, db, clinicID, &oid, nil, 500, model.BillingStatusWaiting, date)

	// completed 全額精算（残差 0 → 未収に出ない）
	full := makeBillingWith(t, db, billingFixtureOpts{
		ClinicID: clinicID, OwnerID: &oid, TotalAmount: 1100,
		Status: model.BillingStatusCompleted, ScheduledDate: date,
	})
	require.NoError(t, db.Create(&model.Payment{
		BillingID: full.ID, ClinicID: clinicID,
		TotalAmount: 1100, BillingAmount: 1100, Method: model.PaymentMethodCreditCard,
	}).Error)

	// completed + クレジット訂正で 1100→900（残差 200 → 未収）
	under := makeBillingWith(t, db, billingFixtureOpts{
		ClinicID: clinicID, OwnerID: &oid, TotalAmount: 1100,
		Status: model.BillingStatusCompleted, ScheduledDate: date,
	})
	require.NoError(t, db.Create(&model.Payment{
		BillingID: under.ID, ClinicID: clinicID,
		TotalAmount: 1100, BillingAmount: 900, Method: model.PaymentMethodCreditCard,
	}).Error)

	t.Run("FindUnpaidByBilling includes residual amount", func(t *testing.T) {
		got, total, err := repo.FindUnpaidByBilling(ctx, clinicID, firstDay2026June, lastDay2026June, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total, "waiting + residual completed の 2 件")
		require.Len(t, got, 2)

		byID := map[uint64]model.Billing{}
		for _, b := range got {
			byID[b.ID] = b
		}
		assert.Contains(t, byID, under.ID)
		assert.NotContains(t, byID, full.ID, "全額精算 completed は未収に出ない")
		assert.Equal(t, int64(200), byID[under.ID].OutstandingAmount)
	})

	t.Run("FindUnpaidByOwner sums residual into owner total", func(t *testing.T) {
		aggs, owners, summary, err := repo.FindUnpaidByOwner(ctx, clinicID, firstDay2026June, lastDay2026June, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(1), owners)
		require.Len(t, aggs, 1)
		// 500 (waiting) + 200 (residual) = 700
		assert.Equal(t, int64(700), aggs[0].TotalAmount)
		assert.Equal(t, int64(2), aggs[0].Count)
		assert.Equal(t, int64(700), summary.TotalAmount)
		assert.Equal(t, int64(2), summary.BillingCount)
	})

	t.Run("SumUnpaidByOwner includes residual", func(t *testing.T) {
		bal, err := repo.SumUnpaidByOwner(ctx, clinicID, oid)
		require.NoError(t, err)
		assert.Equal(t, int64(700), bal.TotalAmount)
		assert.Equal(t, int64(2), bal.Count)
	})
}
