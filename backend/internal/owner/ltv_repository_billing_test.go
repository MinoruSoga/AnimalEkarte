package owner

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestFindOwnerLTV_ExcludesBillingOwnedByAnotherOwner(t *testing.T) {
	db := setupLTVTestDB(t)
	repo := newLTVTestRepository(t, db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := &model.Owner{ClinicID: clinicID, Name: "Owner A"}
	otherOwner := &model.Owner{ClinicID: clinicID, Name: "Owner B"}
	require.NoError(t, db.WithContext(ctx).Create(owner).Error)
	require.NoError(t, db.WithContext(ctx).Create(otherOwner).Error)

	mr := &model.MedicalRecord{ClinicID: clinicID, OwnerID: &owner.ID, Date: time.Now()}
	require.NoError(t, db.WithContext(ctx).Create(mr).Error)
	valid := &model.Billing{
		ClinicID: clinicID, MedicalRecordID: &mr.ID, OwnerID: &owner.ID,
		TotalAmount: 1_000, Status: model.BillingStatusCompleted,
	}
	malformedOwnerReference := &model.Billing{
		ClinicID: clinicID, MedicalRecordID: &mr.ID, OwnerID: &otherOwner.ID,
		TotalAmount: 99_000, Status: model.BillingStatusCompleted,
	}
	require.NoError(t, db.WithContext(ctx).Create(valid).Error)
	require.NoError(t, db.WithContext(ctx).Create(malformedOwnerReference).Error)

	rows, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
		ClinicID: clinicID, IncludeZero: true, IncludeNoVisit: true,
	})
	require.NoError(t, err)
	require.Len(t, rows, 2)
	byOwner := make(map[uint64]OwnerLTVRow, len(rows))
	for _, row := range rows {
		byOwner[row.OwnerID] = row
	}
	assert.Equal(t, int64(1_000), byOwner[owner.ID].TotalAmount)
	assert.Equal(t, int64(0), byOwner[otherOwner.ID].TotalAmount)
	assert.Equal(t, int64(0), byOwner[otherOwner.ID].MaxSingleVisitAmount)
}

// TestFindOwnerLTV_OnlyCompletedBillings
// ISSUE-001: status != completed の会計は集計に含まれない
func TestFindOwnerLTV_OnlyCompletedBillings(t *testing.T) {
	db := setupLTVTestDB(t)
	repo := newLTVTestRepository(t, db)
	ctx := context.Background()

	clinicID := uint64(1)

	owner := &model.Owner{
		ClinicID: clinicID,
		Name:     "Owner Billing Status Test",
	}
	if err := db.WithContext(ctx).Create(owner).Error; err != nil {
		t.Fatalf("failed to create owner: %v", err)
	}

	mr := &model.MedicalRecord{
		ClinicID: clinicID,
		OwnerID:  &owner.ID,
		Date:     time.Now(),
	}
	if err := db.WithContext(ctx).Create(mr).Error; err != nil {
		t.Fatalf("failed to create medical record: %v", err)
	}

	// completed: 5000 円
	completed := &model.Billing{
		ClinicID:        clinicID,
		MedicalRecordID: &mr.ID,
		TotalAmount:     5000,
		Status:          "completed",
	}
	if err := db.WithContext(ctx).Create(completed).Error; err != nil {
		t.Fatalf("failed to create completed billing: %v", err)
	}

	// pending: 3000 円（含まれるべきではない）
	pending := &model.Billing{
		ClinicID:        clinicID,
		MedicalRecordID: &mr.ID,
		TotalAmount:     3000,
		Status:          "pending",
	}
	if err := db.WithContext(ctx).Create(pending).Error; err != nil {
		t.Fatalf("failed to create pending billing: %v", err)
	}

	result, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
		ClinicID:    clinicID,
		IncludeZero: true,
	})
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	// completed のみ = 5000
	assert.Equal(t, int64(5000), result[0].TotalAmount, "only completed billings should be included")
}

// TestFindOwnerLTV_MaxSingleVisitAmount
// ISSUE-006: max_single_visit_amount は完了済み請求の単一最大額（CPMスポット判定用）。
// タグ同期側 AccountingRepository.MaxSingleVisitAmountByOwner と同じ集計範囲（owner_id 直接 + status='completed' + deleted_at IS NULL）を返す。
func TestFindOwnerLTV_MaxSingleVisitAmount(t *testing.T) {
	db := setupLTVTestDB(t)
	repo := newLTVTestRepository(t, db)
	ctx := context.Background()

	clinicID := uint64(1)

	// Owner + 来院 + 複数請求（completed/pending を混在）
	owner := &model.Owner{
		ClinicID: clinicID,
		Name:     "Owner Max Single Visit",
	}
	if err := db.WithContext(ctx).Create(owner).Error; err != nil {
		t.Fatalf("failed to create owner: %v", err)
	}

	mr := &model.MedicalRecord{
		ClinicID: clinicID,
		OwnerID:  &owner.ID,
		Date:     time.Now(),
	}
	if err := db.WithContext(ctx).Create(mr).Error; err != nil {
		t.Fatalf("failed to create medical record: %v", err)
	}

	// completed: 10,000 / 35,000 / 8,000
	for _, amount := range []int64{10_000, 35_000, 8_000} {
		b := &model.Billing{
			ClinicID:        clinicID,
			MedicalRecordID: &mr.ID,
			OwnerID:         &owner.ID,
			TotalAmount:     amount,
			Status:          model.BillingStatusCompleted,
		}
		if err := db.WithContext(ctx).Create(b).Error; err != nil {
			t.Fatalf("failed to create completed billing: %v", err)
		}
	}

	// pending: 50,000 — 含まれてはならない
	pending := &model.Billing{
		ClinicID:        clinicID,
		MedicalRecordID: &mr.ID,
		OwnerID:         &owner.ID,
		TotalAmount:     50_000,
		Status:          "pending",
	}
	if err := db.WithContext(ctx).Create(pending).Error; err != nil {
		t.Fatalf("failed to create pending billing: %v", err)
	}

	result, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
		ClinicID:    clinicID,
		IncludeZero: true,
	})
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, int64(35_000), result[0].MaxSingleVisitAmount,
		"max_single_visit_amount は completed 請求の最大額（35,000）であるべき")
}

// TestFindOwnerLTV_MaxSingleVisitAmountWithoutMedicalRecord
// ISSUE-006: medical_record_id を持たない billing も MaxSingleVisitAmount に含まれること
// （タグ同期側と集計範囲を一致させるため）。
func TestFindOwnerLTV_MaxSingleVisitAmountWithoutMedicalRecord(t *testing.T) {
	db := setupLTVTestDB(t)
	repo := newLTVTestRepository(t, db)
	ctx := context.Background()

	clinicID := uint64(1)

	owner := &model.Owner{
		ClinicID: clinicID,
		Name:     "Owner Direct Billing",
	}
	if err := db.WithContext(ctx).Create(owner).Error; err != nil {
		t.Fatalf("failed to create owner: %v", err)
	}

	// medical_record_id NULL の completed billing — タグ同期側はカウントするため一覧側も合わせる
	directBilling := &model.Billing{
		ClinicID:    clinicID,
		OwnerID:     &owner.ID,
		TotalAmount: 40_000,
		Status:      model.BillingStatusCompleted,
	}
	if err := db.WithContext(ctx).Create(directBilling).Error; err != nil {
		t.Fatalf("failed to create direct billing: %v", err)
	}

	result, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
		ClinicID:       clinicID,
		IncludeZero:    true,
		IncludeNoVisit: true,
	})
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, int64(40_000), result[0].MaxSingleVisitAmount,
		"medical_record_id NULL の請求も MaxSingleVisitAmount に含めるべき（タグ同期側との集計範囲一致）")
}

func TestFindOwnerLTV_IncludesCompletedBillingWithoutMedicalRecordInRevenueButNotVisitCount(t *testing.T) {
	db := setupLTVTestDB(t)
	repo := newLTVTestRepository(t, db)
	ctx := context.Background()
	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)
	now := time.Now()

	ownerA := &model.Owner{ClinicID: clinicA, Name: "Item D Revenue Owner"}
	otherOwnerA := &model.Owner{ClinicID: clinicA, Name: "Item D Other Owner"}
	ownerB := &model.Owner{ClinicID: clinicB, Name: "Item D Clinic B Owner"}
	require.NoError(t, db.WithContext(ctx).Create(ownerA).Error)
	require.NoError(t, db.WithContext(ctx).Create(otherOwnerA).Error)
	require.NoError(t, db.WithContext(ctx).Create(ownerB).Error)

	recordA := &model.MedicalRecord{ClinicID: clinicA, OwnerID: &ownerA.ID, Date: now}
	secondRecordA := &model.MedicalRecord{ClinicID: clinicA, OwnerID: &ownerA.ID, Date: now.AddDate(0, 0, -1)}
	historicalRecordA := &model.MedicalRecord{ClinicID: clinicA, OwnerID: &ownerA.ID, Date: now.AddDate(0, 0, -2)}
	otherRecordA := &model.MedicalRecord{ClinicID: clinicA, OwnerID: &otherOwnerA.ID, Date: now}
	recordB := &model.MedicalRecord{ClinicID: clinicB, OwnerID: &ownerB.ID, Date: now}
	require.NoError(t, db.WithContext(ctx).Create(recordA).Error)
	require.NoError(t, db.WithContext(ctx).Create(secondRecordA).Error)
	require.NoError(t, db.WithContext(ctx).Create(historicalRecordA).Error)
	require.NoError(t, db.WithContext(ctx).Create(otherRecordA).Error)
	require.NoError(t, db.WithContext(ctx).Create(recordB).Error)

	linkedCompleted := &model.Billing{
		ClinicID: clinicA, MedicalRecordID: &recordA.ID,
		TotalAmount: 1_000, Status: model.BillingStatusCompleted,
		ScheduledDate: now, CompletedAt: &now,
	}
	manualCompleted := &model.Billing{
		ClinicID: clinicA, OwnerID: &ownerA.ID,
		TotalAmount: 4_000, Status: model.BillingStatusCompleted,
		ScheduledDate: now, CompletedAt: &now,
	}
	historicalCompleted := &model.Billing{
		ClinicID: clinicA, OwnerID: &ownerA.ID, MedicalRecordID: &historicalRecordA.ID,
		TotalAmount: 2_000, Status: model.BillingStatusCompleted,
		ScheduledDate: historicalRecordA.Date, CompletedAt: &now,
	}
	require.NoError(t, db.WithContext(ctx).Create(linkedCompleted).Error)
	require.NoError(t, db.WithContext(ctx).Create(manualCompleted).Error)
	require.NoError(t, db.WithContext(ctx).Create(historicalCompleted).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.Payment{
		BillingID: linkedCompleted.ID, BillingAmount: 800,
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.Payment{
		BillingID: manualCompleted.ID, BillingAmount: 3_000,
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.Payment{
		BillingID: historicalCompleted.ID, BillingAmount: 1_000,
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.BillingRefund{
		ClinicID: clinicA, BillingID: linkedCompleted.ID, Amount: 100,
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.BillingRefund{
		ClinicID: clinicA, BillingID: manualCompleted.ID, Amount: 500,
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.BillingRefund{
		ClinicID: clinicA, BillingID: historicalCompleted.ID, Amount: 200,
	}).Error)
	require.NoError(t, db.WithContext(ctx).Delete(historicalRecordA).Error)

	excludedBillings := []model.Billing{
		{
			ClinicID: clinicA, OwnerID: &ownerA.ID,
			TotalAmount: 50_000, Status: model.BillingStatusPending, ScheduledDate: now,
		},
		{
			ClinicID: clinicA, OwnerID: &ownerA.ID,
			TotalAmount: 60_000, Status: model.BillingStatusCancelled, ScheduledDate: now,
		},
		{
			ClinicID: clinicB, OwnerID: &ownerA.ID,
			TotalAmount: 70_000, Status: model.BillingStatusCompleted,
			ScheduledDate: now, CompletedAt: &now,
		},
		{
			ClinicID: clinicA, OwnerID: &ownerA.ID, MedicalRecordID: &recordB.ID,
			TotalAmount: 80_000, Status: model.BillingStatusCompleted,
			ScheduledDate: now, CompletedAt: &now,
		},
		{
			ClinicID: clinicA, OwnerID: &ownerA.ID, MedicalRecordID: &otherRecordA.ID,
			TotalAmount: 90_000, Status: model.BillingStatusCompleted,
			ScheduledDate: now, CompletedAt: &now,
		},
	}
	for i := range excludedBillings {
		require.NoError(t, db.WithContext(ctx).Create(&excludedBillings[i]).Error)
	}

	from := now.AddDate(0, 0, -7).Format(time.DateOnly)
	to := now.AddDate(0, 0, 1).Format(time.DateOnly)
	tests := []struct {
		name        string
		amountBasis string
		expected    int64
	}{
		{name: "gross total", amountBasis: "gross_total_amount", expected: 7_000},
		{name: "paid amount", amountBasis: "paid_amount", expected: 4_800},
		{name: "net paid amount", amountBasis: "net_paid_amount", expected: 4_000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
				ClinicID:       clinicA,
				Search:         "Revenue Owner",
				From:           &from,
				To:             &to,
				AmountBasis:    tt.amountBasis,
				IncludeZero:    true,
				IncludeNoVisit: true,
			})
			require.NoError(t, err)
			require.Len(t, rows, 1)
			require.NotNil(t, rows[0].AnnualAmount)
			require.NotNil(t, rows[0].BillingCount)
			require.NotNil(t, rows[0].PeriodVisitCount)
			assert.Equal(t, int64(7_000), rows[0].TotalAmount)
			assert.Equal(t, tt.expected, *rows[0].AnnualAmount)
			assert.Equal(t, int64(3), *rows[0].BillingCount)
			assert.Equal(t, int64(2), rows[0].TotalVisitCount)
			assert.Equal(t, int64(2), *rows[0].PeriodVisitCount)
		})
	}
}

// TestFindOwnerLTV_ClinicIDIsolation
// clinic_id による分離を検証
func TestFindOwnerLTV_ClinicIDIsolation(t *testing.T) {
	db := setupLTVTestDB(t)
	repo := newLTVTestRepository(t, db)
	ctx := context.Background()

	clinicID1 := uint64(1)
	clinicID2 := uint64(2)

	// Clinic 1 の Owner
	owner1 := &model.Owner{
		ClinicID: clinicID1,
		Name:     "Owner Clinic 1",
	}
	if err := db.WithContext(ctx).Create(owner1).Error; err != nil {
		t.Fatalf("failed to create owner1: %v", err)
	}

	// Clinic 2 の Owner
	owner2 := &model.Owner{
		ClinicID: clinicID2,
		Name:     "Owner Clinic 2",
	}
	if err := db.WithContext(ctx).Create(owner2).Error; err != nil {
		t.Fatalf("failed to create owner2: %v", err)
	}

	// Clinic 1 の Clinic 2 の Owner の医療記録
	mr2 := &model.MedicalRecord{
		ClinicID: clinicID2,
		OwnerID:  &owner2.ID,
		Date:     time.Now(),
	}
	if err := db.WithContext(ctx).Create(mr2).Error; err != nil {
		t.Fatalf("failed to create medical record: %v", err)
	}

	// Clinic 1 でクエリ
	result1, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
		ClinicID:       clinicID1,
		IncludeZero:    true,
		IncludeNoVisit: true,
	})
	assert.NoError(t, err)
	// Clinic 1 には Owner 1 のみ
	assert.Len(t, result1, 1)
	assert.Equal(t, "Owner Clinic 1", result1[0].OwnerName)

	// Clinic 2 でクエリ
	result2, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
		ClinicID:       clinicID2,
		IncludeZero:    true,
		IncludeNoVisit: true,
	})
	assert.NoError(t, err)
	// Clinic 2 には Owner 2 のみ
	assert.Len(t, result2, 1)
	assert.Equal(t, "Owner Clinic 2", result2[0].OwnerName)
	assert.Equal(t, int64(1), result2[0].TotalVisitCount)
}

// TestFindOwnerLTV_SameDayMultipleVisitsCountAsOne
// ISSUE-005 / 仕様書 §3.3: 同一飼い主が同じ日に複数カルテを持つ場合は来院1回として数える。
// total_visit_count / period_visit_count は COUNT(DISTINCT mr.date) で算出されるため、
// 同日複数カルテは1回扱いになるべき。
func TestFindOwnerLTV_SameDayMultipleVisitsCountAsOne(t *testing.T) {
	db := setupLTVTestDB(t)
	repo := newLTVTestRepository(t, db)
	ctx := context.Background()

	clinicID := uint64(1)

	owner := &model.Owner{
		ClinicID: clinicID,
		Name:     "Owner Same Day Visits",
	}
	if err := db.WithContext(ctx).Create(owner).Error; err != nil {
		t.Fatalf("failed to create owner: %v", err)
	}

	// 同じ日 (-10) に 3 件のカルテ
	sameDay := time.Now().AddDate(0, 0, -10)
	for i := range 3 {
		mr := &model.MedicalRecord{
			ClinicID: clinicID,
			OwnerID:  &owner.ID,
			Date:     sameDay,
		}
		if err := db.WithContext(ctx).Create(mr).Error; err != nil {
			t.Fatalf("failed to create medical record %d: %v", i, err)
		}
	}
	// 別の日 (-30) に 1 件
	differentDay := time.Now().AddDate(0, 0, -30)
	mr2 := &model.MedicalRecord{
		ClinicID: clinicID,
		OwnerID:  &owner.ID,
		Date:     differentDay,
	}
	if err := db.WithContext(ctx).Create(mr2).Error; err != nil {
		t.Fatalf("failed to create medical record: %v", err)
	}

	result, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
		ClinicID:    clinicID,
		IncludeZero: true,
	})
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	// 4 件のカルテだが日付ベースでは 2 日分
	assert.Equal(t, int64(2), result[0].TotalVisitCount,
		"同日複数カルテは1回として数える (DISTINCT mr.date)")

	// period_preset 指定時も同様
	result2, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
		ClinicID:     clinicID,
		PeriodPreset: "last_3_months",
		IncludeZero:  true,
	})
	assert.NoError(t, err)
	assert.Len(t, result2, 1)
	assert.NotNil(t, result2[0].PeriodVisitCount)
	assert.Equal(t, int64(2), *result2[0].PeriodVisitCount,
		"period_visit_count も同日複数カルテを1回として数える")
}

// TestFindOwnerLTV_FromToBoundaryInclusive
// ISSUE-005 / 仕様書 §10.1: from / to の境界日が集計対象に含まれる。
// SQL は `mr.date >= from AND mr.date <= to` で両端を含む。
func TestFindOwnerLTV_FromToBoundaryInclusive(t *testing.T) {
	db := setupLTVTestDB(t)
	repo := newLTVTestRepository(t, db)
	ctx := context.Background()

	clinicID := uint64(1)

	owner := &model.Owner{
		ClinicID: clinicID,
		Name:     "Owner Boundary Test",
	}
	if err := db.WithContext(ctx).Create(owner).Error; err != nil {
		t.Fatalf("failed to create owner: %v", err)
	}

	// 境界日と前後 1 日に来院
	dates := []time.Time{
		time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC), // from-1
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),   // from
		time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC),  // 範囲内
		time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC), // to
		time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),   // to+1
	}
	for i, d := range dates {
		mr := &model.MedicalRecord{
			ClinicID: clinicID,
			OwnerID:  &owner.ID,
			Date:     d,
		}
		if err := db.WithContext(ctx).Create(mr).Error; err != nil {
			t.Fatalf("failed to create medical record %d: %v", i, err)
		}
		billing := &model.Billing{
			ClinicID:        clinicID,
			MedicalRecordID: &mr.ID,
			OwnerID:         &owner.ID,
			TotalAmount:     1000,
			Status:          model.BillingStatusCompleted,
			ScheduledDate:   d,
		}
		if err := db.WithContext(ctx).Create(billing).Error; err != nil {
			t.Fatalf("failed to create billing %d: %v", i, err)
		}
	}

	from := "2026-01-01"
	to := "2026-12-31"
	result, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
		ClinicID:    clinicID,
		From:        &from,
		To:          &to,
		IncludeZero: true,
	})
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	require.NotNil(t, result[0].PeriodVisitCount)
	// 範囲内 3 件 (from / 範囲内 / to)、範囲外 2 件 (from-1 / to+1) は除外
	assert.Equal(t, int64(3), *result[0].PeriodVisitCount,
		"from / to の境界日は集計に含まれる (両端 inclusive)")
	require.NotNil(t, result[0].AnnualAmount)
	assert.Equal(t, int64(3000), *result[0].AnnualAmount,
		"annual_amount は from / to 境界を含む合計 (1000 * 3)")
	// total_visit_count は全期間 = 5 日
	assert.Equal(t, int64(5), result[0].TotalVisitCount,
		"total_visit_count は全期間で from / to の影響を受けない")
}

// TestFindOwnerLTV_SearchByName
// ISSUE-005 / 仕様書 §10.1: search パラメータが owner.name の部分一致で効く。
// SQL は ILIKE '%search%' で大文字小文字を区別しない。
func TestFindOwnerLTV_SearchByName(t *testing.T) {
	db := setupLTVTestDB(t)
	repo := newLTVTestRepository(t, db)
	ctx := context.Background()

	clinicID := uint64(1)

	names := []string{"山田 太郎", "山田 花子", "佐藤 一郎", "Tanaka Smith"}
	for _, name := range names {
		o := &model.Owner{ClinicID: clinicID, Name: name}
		if err := db.WithContext(ctx).Create(o).Error; err != nil {
			t.Fatalf("failed to create owner: %v", err)
		}
	}

	t.Run("partial match (japanese)", func(t *testing.T) {
		result, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
			ClinicID:       clinicID,
			Search:         "山田",
			IncludeZero:    true,
			IncludeNoVisit: true,
		})
		assert.NoError(t, err)
		assert.Len(t, result, 2, "山田 を含む 2 件のみ")
		for _, r := range result {
			assert.Contains(t, r.OwnerName, "山田")
		}
	})

	t.Run("partial match (case-insensitive ascii)", func(t *testing.T) {
		result, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
			ClinicID:       clinicID,
			Search:         "tanaka",
			IncludeZero:    true,
			IncludeNoVisit: true,
		})
		assert.NoError(t, err)
		assert.Len(t, result, 1, "ILIKE で大文字小文字を区別しない")
		assert.Equal(t, "Tanaka Smith", result[0].OwnerName)
	})

	t.Run("no match returns empty", func(t *testing.T) {
		result, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
			ClinicID:       clinicID,
			Search:         "存在しない名前",
			IncludeZero:    true,
			IncludeNoVisit: true,
		})
		assert.NoError(t, err)
		assert.Len(t, result, 0)
	})

	t.Run("empty search returns all", func(t *testing.T) {
		result, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
			ClinicID:       clinicID,
			Search:         "",
			IncludeZero:    true,
			IncludeNoVisit: true,
		})
		assert.NoError(t, err)
		assert.Len(t, result, 4, "search 未指定は全件")
	})
}

// TestLtvRepository_BuildOrderBy はソートフィールド×順序の組み合わせで ORDER BY 句が
// 期待通りに構築されることを検証する（DB 非依存の純粋関数のためテーブル駆動で直接検証）。
func TestLtvRepository_BuildOrderBy(t *testing.T) {
	repo := &ltvRepository{}

	tests := []struct {
		name   string
		sort   string
		order  string
		expect string
	}{
		{"annual_amount asc", "annual_amount", "asc", "annual_amount ASC NULLS LAST"},
		{"annual_amount desc", "annual_amount", "desc", "annual_amount DESC NULLS LAST"},
		{"total_amount asc", "total_amount", "asc", "total_amount ASC NULLS LAST"},
		{"visit_count desc", "visit_count", "desc", "period_visit_count DESC NULLS LAST"},
		{"total_visit_count asc", "total_visit_count", "asc", "total_visit_count ASC NULLS LAST"},
		{"annual_visit_count desc", "annual_visit_count", "desc", "annual_visit_count DESC NULLS LAST"},
		{"last_visit_date asc", "last_visit_date", "asc", "last_visit_date ASC NULLS LAST"},
		{"days_since_last_visit desc", "days_since_last_visit", "desc", "days_since_last_visit DESC NULLS LAST"},
		{"owner_name asc (no NULLS LAST)", "owner_name", "asc", "owner_name ASC"},
		{"unknown sort falls back to total_amount", "unknown_field", "desc", "total_amount DESC NULLS LAST"},
		{"empty sort falls back to total_amount", "", "asc", "total_amount ASC NULLS LAST"},
		{"invalid order defaults to desc", "total_amount", "sideways", "total_amount DESC NULLS LAST"},
		{"empty order defaults to desc", "total_amount", "", "total_amount DESC NULLS LAST"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expect, repo.buildOrderBy(tc.sort, tc.order))
		})
	}
}

// TestLtvRepository_CalculateDateRange_InvalidFormats は from/to のパース失敗時に
// エラーが伝播することを検証する（DB 非依存）。
func TestLtvRepository_CalculateDateRange_InvalidFormats(t *testing.T) {
	repo := &ltvRepository{}

	t.Run("invalid From format returns an error", func(t *testing.T) {
		from := "not-a-date"
		to := "2026-01-01"
		fromDate, toDate, err := repo.calculateDateRange(&FindOwnerLTVParams{From: &from, To: &to})
		assert.Error(t, err)
		assert.Nil(t, fromDate)
		assert.Nil(t, toDate)
	})

	t.Run("invalid To format returns an error", func(t *testing.T) {
		from := "2026-01-01"
		to := "not-a-date"
		fromDate, toDate, err := repo.calculateDateRange(&FindOwnerLTVParams{From: &from, To: &to})
		assert.Error(t, err)
		assert.Nil(t, fromDate)
		assert.Nil(t, toDate)
	})

	t.Run("year takes priority over period_preset", func(t *testing.T) {
		year := 2025
		fromDate, toDate, err := repo.calculateDateRange(&FindOwnerLTVParams{Year: &year, PeriodPreset: "last_3_months"})
		require.NoError(t, err)
		require.NotNil(t, fromDate)
		require.NotNil(t, toDate)
		assert.Equal(t, 2025, fromDate.Year())
		assert.Equal(t, 2025, toDate.Year())
	})

	t.Run("no filters returns nil range (all time)", func(t *testing.T) {
		fromDate, toDate, err := repo.calculateDateRange(&FindOwnerLTVParams{})
		require.NoError(t, err)
		assert.Nil(t, fromDate)
		assert.Nil(t, toDate)
	})
}

// TestFindOwnerLTV_InvalidFromDateFormatPropagatesError
// calculateDateRange のエラーが公開 API である FindOwnerLTV から呼び出し元へ伝播することを検証する。
func TestFindOwnerLTV_InvalidFromDateFormatPropagatesError(t *testing.T) {
	db := setupLTVTestDB(t)
	repo := newLTVTestRepository(t, db)
	ctx := context.Background()

	badFrom := "20260101"
	to := "2026-12-31"
	rows, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
		ClinicID: uint64(1),
		From:     &badFrom,
		To:       &to,
	})
	assert.Error(t, err)
	assert.Nil(t, rows)
}

// TestFindOwnerLTV_MinVisitCountAndMaxVisitCountFilter
// AGG-BE-002: min_visit_count / max_visit_count による HAVING 絞り込みを検証する。
func TestFindOwnerLTV_MinVisitCountAndMaxVisitCountFilter(t *testing.T) {
	db := setupLTVTestDB(t)
	repo := newLTVTestRepository(t, db)
	ctx := context.Background()

	clinicID := uint64(1)

	fewVisits := &model.Owner{ClinicID: clinicID, Name: "Owner Few Visits"}
	manyVisits := &model.Owner{ClinicID: clinicID, Name: "Owner Many Visits"}
	require.NoError(t, db.WithContext(ctx).Create(fewVisits).Error)
	require.NoError(t, db.WithContext(ctx).Create(manyVisits).Error)

	now := time.Now()
	for i := 0; i < 2; i++ {
		mr := &model.MedicalRecord{ClinicID: clinicID, OwnerID: &fewVisits.ID, Date: now.AddDate(0, 0, -i)}
		require.NoError(t, db.WithContext(ctx).Create(mr).Error)
	}
	for i := 0; i < 5; i++ {
		mr := &model.MedicalRecord{ClinicID: clinicID, OwnerID: &manyVisits.ID, Date: now.AddDate(0, 0, -i)}
		require.NoError(t, db.WithContext(ctx).Create(mr).Error)
	}

	t.Run("min_visit_count excludes owners with fewer visits", func(t *testing.T) {
		minVisits := int64(3)
		rows, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
			ClinicID:      clinicID,
			MinVisitCount: &minVisits,
			IncludeZero:   true,
		})
		require.NoError(t, err)
		require.Len(t, rows, 1)
		assert.Equal(t, manyVisits.ID, rows[0].OwnerID)
	})

	t.Run("max_visit_count excludes owners with more visits", func(t *testing.T) {
		maxVisits := int64(3)
		rows, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
			ClinicID:      clinicID,
			MaxVisitCount: &maxVisits,
			IncludeZero:   true,
		})
		require.NoError(t, err)
		require.Len(t, rows, 1)
		assert.Equal(t, fewVisits.ID, rows[0].OwnerID)
	})
}

// TestFindOwnerLTV_LastVisitBucketFilterExcludesOtherBuckets
// AGG-BE-003: last_visit_bucket 指定時、他バケットのオーナーは除外されることを検証する。
func TestFindOwnerLTV_LastVisitBucketFilterExcludesOtherBuckets(t *testing.T) {
	db := setupLTVTestDB(t)
	repo := newLTVTestRepository(t, db)
	ctx := context.Background()

	clinicID := uint64(1)
	now := time.Now()

	recentOwner := &model.Owner{ClinicID: clinicID, Name: "Owner Recent"}
	oldOwner := &model.Owner{ClinicID: clinicID, Name: "Owner Old"}
	require.NoError(t, db.WithContext(ctx).Create(recentOwner).Error)
	require.NoError(t, db.WithContext(ctx).Create(oldOwner).Error)

	require.NoError(t, db.WithContext(ctx).Create(&model.MedicalRecord{ClinicID: clinicID, OwnerID: &recentOwner.ID, Date: now.AddDate(0, 0, -1)}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.MedicalRecord{ClinicID: clinicID, OwnerID: &oldOwner.ID, Date: now.AddDate(0, 0, -400)}).Error)

	rows, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
		ClinicID:        clinicID,
		LastVisitBucket: "within_3m",
		IncludeZero:     true,
	})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, recentOwner.ID, rows[0].OwnerID)
}

// TestFindOwnerLTV_SortOrdering
// sort/order パラメータの組み合わせで total_amount の昇順・降順が反転することを検証する。
func TestFindOwnerLTV_SortOrdering(t *testing.T) {
	db := setupLTVTestDB(t)
	repo := newLTVTestRepository(t, db)
	ctx := context.Background()

	clinicID := uint64(1)

	low := &model.Owner{ClinicID: clinicID, Name: "Owner Low Amount"}
	high := &model.Owner{ClinicID: clinicID, Name: "Owner High Amount"}
	require.NoError(t, db.WithContext(ctx).Create(low).Error)
	require.NoError(t, db.WithContext(ctx).Create(high).Error)

	mrLow := &model.MedicalRecord{ClinicID: clinicID, OwnerID: &low.ID, Date: time.Now()}
	require.NoError(t, db.WithContext(ctx).Create(mrLow).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.Billing{ClinicID: clinicID, MedicalRecordID: &mrLow.ID, OwnerID: &low.ID, TotalAmount: 1000, Status: model.BillingStatusCompleted}).Error)

	mrHigh := &model.MedicalRecord{ClinicID: clinicID, OwnerID: &high.ID, Date: time.Now()}
	require.NoError(t, db.WithContext(ctx).Create(mrHigh).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.Billing{ClinicID: clinicID, MedicalRecordID: &mrHigh.ID, OwnerID: &high.ID, TotalAmount: 9000, Status: model.BillingStatusCompleted}).Error)

	t.Run("ascending order returns the lowest total_amount first", func(t *testing.T) {
		rows, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{ClinicID: clinicID, Sort: "total_amount", Order: "asc", IncludeZero: true})
		require.NoError(t, err)
		require.Len(t, rows, 2)
		assert.Equal(t, low.ID, rows[0].OwnerID)
		assert.Equal(t, high.ID, rows[1].OwnerID)
	})

	t.Run("descending order returns the highest total_amount first", func(t *testing.T) {
		rows, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{ClinicID: clinicID, Sort: "total_amount", Order: "desc", IncludeZero: true})
		require.NoError(t, err)
		require.Len(t, rows, 2)
		assert.Equal(t, high.ID, rows[0].OwnerID)
		assert.Equal(t, low.ID, rows[1].OwnerID)
	})
}
