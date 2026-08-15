package billing

// accounting_repository_ltv_test.go — AccountingRepository.SumPaidByOwner /
// MaxSingleVisitAmountByOwner / FindOwnersByAnnualRevenue の統合テスト（実 Postgres テスト DB）。
//
// 注: これらは Lステップタグ同期 / CPM 判定用の集計であり、LtvRepository.FindOwnerLTV とは
// 別の集計経路（ltv_repository_test.go 参照）。本ファイルは AccountingRepository 側を対象とする。
//
// G2F-03: FindOwnersByAnnualRevenue は exact top-20% の bounded SQL ranking。
// 全 clinic owner 売上を Go へ materialize しない。

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func TestAccountingRepository_SumPaidByOwner_SumsCompletedOnlyForOwner(t *testing.T) {
	db := testdb.SetupTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := testdb.MakeTestOwner(t, db, clinicID, "LTV飼主")
	other := testdb.MakeTestOwner(t, db, clinicID, "別飼主")

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
	db := testdb.SetupTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := testdb.MakeTestOwner(t, db, clinicID, "未会計飼主")
	total, err := repo.SumPaidByOwner(ctx, clinicID, owner.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
}

func TestAccountingRepository_SumPaidByOwner_ClinicIsolation(t *testing.T) {
	db := testdb.SetupTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := testdb.MakeTestOwner(t, db, clinicA, "医院A飼主")
	billing := &model.Billing{ClinicID: clinicA, OwnerID: &owner.ID, TotalAmount: 4000, Status: model.BillingStatusCompleted, ScheduledDate: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)}
	require.NoError(t, db.WithContext(ctx).Create(billing).Error)

	total, err := repo.SumPaidByOwner(ctx, clinicB, owner.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total, "別クリニックからは飼主の支払いが見えてはならない")
}

func TestAccountingRepository_MaxSingleVisitAmountByOwner_ReturnsMaxOfCompleted(t *testing.T) {
	db := testdb.SetupTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := testdb.MakeTestOwner(t, db, clinicID, "最大来院飼主")
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
	db := testdb.SetupTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := testdb.MakeTestOwner(t, db, clinicID, "会計なし飼主")
	maxAmount, err := repo.MaxSingleVisitAmountByOwner(ctx, clinicID, owner.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), maxAmount)
}

func TestAccountingRepository_LTVRevenue_OrdersDescendingAndExcludesOldBillings(t *testing.T) {
	db := testdb.SetupTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	now := time.Now()

	// 5 qualifying owners → top 20% = ceil(5*0.2) = 1 → only the highest revenue owner.
	high := testdb.MakeTestOwner(t, db, clinicID, "高額飼主")
	mid := testdb.MakeTestOwner(t, db, clinicID, "中額飼主")
	low := testdb.MakeTestOwner(t, db, clinicID, "低額飼主")
	lower := testdb.MakeTestOwner(t, db, clinicID, "更低額飼主")
	lowest := testdb.MakeTestOwner(t, db, clinicID, "最低額飼主")

	for _, row := range []struct {
		owner  *model.Owner
		amount int64
	}{
		{high, 50_000},
		{mid, 30_000},
		{low, 10_000},
		{lower, 5_000},
		{lowest, 1_000},
	} {
		b := &model.Billing{
			ClinicID: clinicID, OwnerID: &row.owner.ID, TotalAmount: row.amount,
			Status: model.BillingStatusCompleted, ScheduledDate: now.AddDate(0, 0, -5),
			CompletedAt: timePtr(now.AddDate(0, 0, -5)),
		}
		require.NoError(t, db.WithContext(ctx).Create(b).Error)
	}

	// 365日超過（除外対象）— high の売上を追加してしまうと除外検証にならないため別飼主で作成
	tooOld := testdb.MakeTestOwner(t, db, clinicID, "365日超過飼主")
	oldBilling := &model.Billing{ClinicID: clinicID, OwnerID: &tooOld.ID, TotalAmount: 99_999, Status: model.BillingStatusCompleted, ScheduledDate: now.AddDate(0, 0, -400), CompletedAt: timePtr(now.AddDate(0, 0, -400))}
	require.NoError(t, db.WithContext(ctx).Create(oldBilling).Error)

	results, err := repo.FindOwnersByAnnualRevenue(ctx, clinicID)
	require.NoError(t, err)

	require.Len(t, results, 1, "top 20% of 5 qualifying owners is exactly 1")
	assert.Equal(t, high.ID, results[0].OwnerID, "降順ソートで高額飼主のみが top 集合")
	assert.Equal(t, int64(50_000), results[0].Revenue)

	byOwner := make(map[uint64]int64, len(results))
	for _, r := range results {
		byOwner[r.OwnerID] = r.Revenue
	}
	_, oldPresent := byOwner[tooOld.ID]
	assert.False(t, oldPresent, "365日超過の売上は集計から除外される")
	_, midPresent := byOwner[mid.ID]
	assert.False(t, midPresent, "top 20% 外の飼主は返却集合に含まれない")
}

func TestAccountingRepository_LTVRevenue_ExcludesNullOwnerID(t *testing.T) {
	db := testdb.SetupTestDB(t)
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

func TestAccountingRepository_LTVRevenue_ClinicIsolation(t *testing.T) {
	db := testdb.SetupTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	now := time.Now()

	ownerA := testdb.MakeTestOwner(t, db, clinicA, "医院A飼主")
	ownerB := testdb.MakeTestOwner(t, db, clinicB, "医院B飼主")
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

func TestAccountingRepository_LTVRevenue_ExcludesCrossClinicOwnerReference(t *testing.T) {
	db := testdb.SetupTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	now := time.Now()

	ownerA := testdb.MakeTestOwner(t, db, clinicA, "医院A飼主")
	ownerB := testdb.MakeTestOwner(t, db, clinicB, "医院B飼主")
	valid := &model.Billing{
		ClinicID: clinicA, OwnerID: &ownerA.ID, TotalAmount: 1_000,
		Status: model.BillingStatusCompleted, ScheduledDate: now, CompletedAt: timePtr(now),
	}
	malformedCrossClinic := &model.Billing{
		ClinicID: clinicA, OwnerID: &ownerB.ID, TotalAmount: 99_000,
		Status: model.BillingStatusCompleted, ScheduledDate: now, CompletedAt: timePtr(now),
	}
	require.NoError(t, db.WithContext(ctx).Create(valid).Error)
	require.NoError(t, db.WithContext(ctx).Create(malformedCrossClinic).Error)

	results, err := repo.FindOwnersByAnnualRevenue(ctx, clinicA)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, ownerA.ID, results[0].OwnerID)
	assert.Equal(t, int64(1_000), results[0].Revenue)
}

// DEC-27: medical_records.owner_id and billings.owner_id are independent
// snapshots. Mismatched MR snapshot owner must not exclude LTV rows when
// clinic matches; money attribution stays on billings.owner_id.
func TestAccountingRepository_LTVRevenueAggregates_IncludeMismatchedMedicalRecordOwnerSnapshotAndAllowDirectBilling(t *testing.T) {
	db := testdb.SetupTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	now := time.Now()

	owner := testdb.MakeTestOwner(t, db, clinicID, "集計対象飼主")
	otherOwner := testdb.MakeTestOwner(t, db, clinicID, "別飼主")
	medicalRecord := &model.MedicalRecord{
		ClinicID: clinicID, OwnerID: &otherOwner.ID, Date: now, RecordNo: "LTV-OWNER-MISMATCH",
	}
	require.NoError(t, db.WithContext(ctx).Create(medicalRecord).Error)
	direct := &model.Billing{
		ClinicID: clinicID, OwnerID: &owner.ID, TotalAmount: 1_000,
		Status: model.BillingStatusCompleted, ScheduledDate: now, CompletedAt: timePtr(now),
	}
	mismatched := &model.Billing{
		ClinicID: clinicID, OwnerID: &owner.ID, MedicalRecordID: &medicalRecord.ID, TotalAmount: 99_000,
		Status: model.BillingStatusCompleted, ScheduledDate: now, CompletedAt: timePtr(now),
	}
	require.NoError(t, db.WithContext(ctx).Create(direct).Error)
	require.NoError(t, db.WithContext(ctx).Create(mismatched).Error)

	total, err := repo.SumPaidByOwner(ctx, clinicID, owner.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(100_000), total, "direct + MR-linked billing both attribute to billing.owner_id")

	maxAmount, err := repo.MaxSingleVisitAmountByOwner(ctx, clinicID, owner.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(99_000), maxAmount)

	results, err := repo.FindOwnersByAnnualRevenue(ctx, clinicID)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, owner.ID, results[0].OwnerID)
	assert.Equal(t, int64(100_000), results[0].Revenue)
}

// DEC-27 transfer: after pet moves to another owner, LTV for the original
// billing snapshot owner still includes the completed billing (keyed by
// billings.owner_id), even when the linked MR snapshot owner differs from
// the current pets.owner_id.
func TestAccountingRepository_LTVRevenueAggregates_KeepsAttributionAfterPetTransfer(t *testing.T) {
	db := testdb.SetupTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	now := time.Now()

	originalOwner := testdb.MakeTestOwner(t, db, clinicID, "譲渡前飼主")
	newOwner := testdb.MakeTestOwner(t, db, clinicID, "譲渡後飼主")
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.AnimalSpecies{}, &model.Pet{}, &model.MedicalRecord{}))
	species := &model.AnimalSpecies{Name: "犬"}
	require.NoError(t, db.WithContext(ctx).Create(species).Error)
	pet := &model.Pet{
		ClinicID: clinicID, OwnerID: originalOwner.ID, AnimalSpeciesID: species.ID, Name: "ltv-transfer-pet",
	}
	require.NoError(t, db.WithContext(ctx).Create(pet).Error)
	medicalRecord := &model.MedicalRecord{
		ClinicID: clinicID, OwnerID: &originalOwner.ID, PetID: &pet.ID, Date: now, RecordNo: "LTV-TRANSFER",
	}
	require.NoError(t, db.WithContext(ctx).Create(medicalRecord).Error)
	billing := &model.Billing{
		ClinicID: clinicID, OwnerID: &originalOwner.ID, PetID: &pet.ID, MedicalRecordID: &medicalRecord.ID,
		TotalAmount: 7_500, Status: model.BillingStatusCompleted, ScheduledDate: now, CompletedAt: timePtr(now),
	}
	require.NoError(t, db.WithContext(ctx).Create(billing).Error)

	// Pet transfer: current owner changes; billing/MR snapshots stay.
	require.NoError(t, db.WithContext(ctx).Model(&model.Pet{}).
		Where("id = ?", pet.ID).
		Update("owner_id", newOwner.ID).Error)

	total, err := repo.SumPaidByOwner(ctx, clinicID, originalOwner.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(7_500), total, "LTV stays on billing snapshot owner after pet transfer")

	newOwnerTotal, err := repo.SumPaidByOwner(ctx, clinicID, newOwner.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), newOwnerTotal, "current pet owner does not inherit historical LTV")
}

// TestAccountingRepository_LTVRevenue_BoundsTopPercentForLargeClinic
// is the G2F-03 large-clinic regression: with N qualifying owners the repository
// must return only ceil(N*20/100) rows, never the full owner-revenue set.
func TestAccountingRepository_LTVRevenue_BoundsTopPercentForLargeClinic(t *testing.T) {
	db := testdb.SetupTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	const ownerCount = 50
	now := time.Now()

	owners := make([]*model.Owner, 0, ownerCount)
	for i := 0; i < ownerCount; i++ {
		owners = append(owners, testdb.MakeTestOwner(t, db, clinicID, fmt.Sprintf("large-clinic-owner-%02d", i)))
	}
	// Distinct revenues: owners[i] gets (i+1)*1000 so owners[49] is highest.
	for i, owner := range owners {
		amount := int64(i+1) * 1_000
		b := &model.Billing{
			ClinicID: clinicID, OwnerID: &owner.ID, TotalAmount: amount,
			Status: model.BillingStatusCompleted, ScheduledDate: now, CompletedAt: timePtr(now),
		}
		require.NoError(t, db.WithContext(ctx).Create(b).Error)
	}

	results, err := repo.FindOwnersByAnnualRevenue(ctx, clinicID)
	require.NoError(t, err)

	wantTopN := (ownerCount*ltvTopPercent + 99) / 100 // 10 for N=50
	require.Equal(t, wantTopN, len(results),
		"bounded top-percent contract must return ceil(N*20/100)=%d, not full N=%d (Go must not materialize all owner revenues)",
		wantTopN, ownerCount)

	// Highest 10 revenues: owners[49]..owners[40]
	for i, got := range results {
		wantOwner := owners[ownerCount-1-i]
		wantRevenue := int64(ownerCount-i) * 1_000
		assert.Equal(t, wantOwner.ID, got.OwnerID, "rank %d owner", i+1)
		assert.Equal(t, wantRevenue, got.Revenue, "rank %d revenue", i+1)
	}
}

// TestAccountingRepository_LTVRevenue_DeterministicTieBreakByOwnerID
// pins revenue DESC, owner_id ASC when revenues collide at the top-percent boundary.
func TestAccountingRepository_LTVRevenue_DeterministicTieBreakByOwnerID(t *testing.T) {
	db := testdb.SetupTestDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	now := time.Now()

	// 5 owners, identical revenue → topN = 1; lowest owner_id wins the tie.
	owners := make([]*model.Owner, 0, 5)
	for i := 0; i < 5; i++ {
		owners = append(owners, testdb.MakeTestOwner(t, db, clinicID, fmt.Sprintf("tie-owner-%d", i)))
	}
	for _, owner := range owners {
		b := &model.Billing{
			ClinicID: clinicID, OwnerID: &owner.ID, TotalAmount: 10_000,
			Status: model.BillingStatusCompleted, ScheduledDate: now, CompletedAt: timePtr(now),
		}
		require.NoError(t, db.WithContext(ctx).Create(b).Error)
	}

	// Ensure owner IDs are ascending as created.
	for i := 1; i < len(owners); i++ {
		require.Less(t, owners[i-1].ID, owners[i].ID)
	}

	results, err := repo.FindOwnersByAnnualRevenue(ctx, clinicID)
	require.NoError(t, err)
	require.Len(t, results, 1, "top 20% of 5 tied owners is exactly 1")
	assert.Equal(t, owners[0].ID, results[0].OwnerID, "tie-break: lowest owner_id wins at equal revenue")
	assert.Equal(t, int64(10_000), results[0].Revenue)
}

// TestAccountingRepository_LTVRevenue_ExplainPlanIsWindowBounded
// records PostgreSQL EXPLAIN evidence for the exact top-N/count strategy.
// Asserts WindowAgg is present so ranking happens in the database, not in Go.
func TestAccountingRepository_LTVRevenue_ExplainPlanIsWindowBounded(t *testing.T) {
	db := testdb.SetupTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)
	cutoff := time.Now().AddDate(0, 0, -365)

	// EXPLAIN (FORMAT TEXT) returns one row per plan line.
	rows, err := db.WithContext(ctx).Raw(
		"EXPLAIN (FORMAT TEXT, COSTS FALSE) "+ownerAnnualRevenueTopPercentSQL,
		clinicID,
		model.BillingStatusCompleted,
		cutoff,
		ltvTopPercent,
	).Rows()
	require.NoError(t, err)
	defer rows.Close()

	var planLines []string
	for rows.Next() {
		var line string
		require.NoError(t, rows.Scan(&line))
		planLines = append(planLines, line)
	}
	require.NoError(t, rows.Err())
	require.NotEmpty(t, planLines, "EXPLAIN must return a non-empty plan")

	plan := strings.Join(planLines, "\n")
	t.Logf("FindOwnersByAnnualRevenue EXPLAIN plan:\n%s", plan)

	// WindowAgg is PostgreSQL's node for ROW_NUMBER/COUNT(*) OVER ranking.
	// Presence proves top-N selection is planned in SQL, not post-fetched in Go.
	assert.Contains(t, plan, "WindowAgg",
		"plan must use WindowAgg for bounded top-percent ranking (ROW_NUMBER/COUNT OVER)")
}

// TestAccountingRepository_LTVRevenue_SourceContractRejectsFullMaterialization
// fails if the production source regresses to an unbounded grouped Scan without top-N filter.
func TestAccountingRepository_LTVRevenue_SourceContractRejectsFullMaterialization(t *testing.T) {
	src, err := os.ReadFile("accounting_repository_ltv.go")
	require.NoError(t, err)
	body := string(src)

	assert.Contains(t, body, "ROW_NUMBER()", "ranking must use SQL ROW_NUMBER for exact top-N")
	assert.Contains(t, body, "COUNT(*) OVER ()", "topN requires SQL total_owners window count")
	assert.Contains(t, body, "total_owners * ? + 99", "exact ceil(N*percent/100) must live in SQL")
	assert.Contains(t, body, "ORDER BY revenue DESC, owner_id ASC", "deterministic tie policy required")
	assert.NotContains(t, body, `Order("revenue DESC").
		Scan(&results)`,
		"must not Scan the full unbounded owner-revenue set into Go")
}

func timePtr(t time.Time) *time.Time { return &t }
