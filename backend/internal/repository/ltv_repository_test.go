package repository

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

// TestFindOwnerLTV_PeriodVisitCountDoesNotAffectTotalVisitCount
// ISSUE-002: period_visit_count の絞り込みが total_visit_count に波及しないこと
func TestFindOwnerLTV_PeriodVisitCountDoesNotAffectTotalVisitCount(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLtvRepository(db)
	ctx := context.Background()

	clinicID := uint64(1)

	// Setup: Owner と 390 日間にわたる 10 回の来院を作成
	owner := &model.Owner{
		ClinicID: clinicID,
		Name:     "Owner A - 10 visits over 390 days",
	}
	if err := db.WithContext(ctx).Create(owner).Error; err != nil {
		t.Fatalf("failed to create owner: %v", err)
	}

	// ===== DEBUG 1: owner.ID が生成されているか =====
	t.Logf("[DEBUG] owner.ID=%d (should not be 0)", owner.ID)
	if owner.ID == 0 {
		t.Fatal("owner.ID is 0: Create() failed to generate primary key")
	}

	// ===== DEBUG 2: 直接 SELECT COUNT(*) で Owner が DB に入っているか =====
	var ownerCount int64
	db.WithContext(ctx).Model(&model.Owner{}).Where("clinic_id = ?", clinicID).Count(&ownerCount)
	t.Logf("[DEBUG] Owner count in DB (clinic_id=%d): %d", clinicID, ownerCount)
	if ownerCount == 0 {
		t.Fatal("no owners found in DB after Create()")
	}

	now := time.Now()
	visitDates := []time.Time{
		now.AddDate(0, 0, -5),   // within 3m
		now.AddDate(0, 0, -30),  // within 3m
		now.AddDate(0, 0, -95),  // over 3m
		now.AddDate(0, 0, -120), // over 3m
		now.AddDate(0, 0, -190), // over 6m
		now.AddDate(0, 0, -220), // over 6m
		now.AddDate(0, 0, -350), // over 1y
		now.AddDate(0, 0, -380), // over 1y
		now.AddDate(-2, 0, 0),   // 2 years ago
		now.AddDate(-3, 0, 0),   // 3 years ago
	}

	for i, visitDate := range visitDates {
		mr := &model.MedicalRecord{
			ClinicID: clinicID,
			OwnerID:  &owner.ID,
			Date:     visitDate,
		}
		if err := db.WithContext(ctx).Create(mr).Error; err != nil {
			t.Fatalf("failed to create medical record %d: %v", i, err)
		}

		// Create billing record for each medical record so annual_amount is non-zero
		billing := &model.Billing{
			ClinicID:        clinicID,
			MedicalRecordID: &mr.ID,
			OwnerID:         &owner.ID,
			TotalAmount:     1, // ¥1 per visit (so AnnualAmount = visit count)
			Status:          model.BillingStatusCompleted,
			ScheduledDate:   visitDate,
		}
		if err := db.WithContext(ctx).Create(billing).Error; err != nil {
			t.Fatalf("failed to create billing %d: %v", i, err)
		}
	}

	// ===== DEBUG 3: MedicalRecord が DB に入っているか =====
	var mrCount int64
	db.WithContext(ctx).Model(&model.MedicalRecord{}).Where("clinic_id = ? AND owner_id = ?", clinicID, owner.ID).Count(&mrCount)
	t.Logf("[DEBUG] MedicalRecord count in DB (clinic_id=%d, owner_id=%d): %d", clinicID, owner.ID, mrCount)

	// Test 1: 全期間で total_visit_count = 10
	// ===== DEBUG 4a: Manual SQL to verify LEFT JOIN working =====
	var manualResult []struct {
		OwnerID     uint64
		OwnerName   string
		MRCount     int64
		FirstMRDate *time.Time
		LastMRDate  *time.Time
	}
	manualQuery := `
		SELECT o.id AS owner_id, o.name AS owner_name,
		       COUNT(DISTINCT mr.date) AS mr_count,
		       MIN(mr.date) AS first_mr_date,
		       MAX(mr.date) AS last_mr_date
		FROM owners o
		LEFT JOIN medical_records mr ON mr.owner_id = o.id AND mr.clinic_id = o.clinic_id AND mr.deleted_at IS NULL
		WHERE o.clinic_id = ? AND o.deleted_at IS NULL
		GROUP BY o.id, o.name
	`
	if err := db.WithContext(ctx).Raw(manualQuery, clinicID).Scan(&manualResult).Error; err != nil {
		t.Fatalf("manual query failed: %v", err)
	}
	t.Logf("[DEBUG 4a] Manual LEFT JOIN result count: %d", len(manualResult))
	for i, row := range manualResult {
		t.Logf("[DEBUG 4a] Row %d: OwnerID=%d, Name=%s, MRCount=%d, FirstDate=%v, LastDate=%v",
			i, row.OwnerID, row.OwnerName, row.MRCount, row.FirstMRDate, row.LastMRDate)
	}

	// ===== DEBUG 4b: FindOwnerLTV 実行 =====
	t.Logf("[DEBUG 4b] About to call FindOwnerLTV with ClinicID=%d", clinicID)
	result1, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
		ClinicID: clinicID,
	})
	t.Logf("[DEBUG 4b] FindOwnerLTV result count: %d (expected 1)", len(result1))
	if len(result1) > 0 {
		t.Logf("[DEBUG] result1[0]: OwnerID=%d, OwnerName=%s, TotalVisitCount=%d", result1[0].OwnerID, result1[0].OwnerName, result1[0].TotalVisitCount)
	}
	assert.NoError(t, err)
	assert.Len(t, result1, 1)
	assert.Equal(t, int64(10), result1[0].TotalVisitCount, "全期間での来院回数は 10 であるべき")

	// Test 2: last_12_months 指定時
	// total_visit_count は 10 のまま（period_visit_count が変わるべきで total_visit_count は変わらない）
	result2, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
		ClinicID:     clinicID,
		PeriodPreset: "last_12_months",
	})
	assert.NoError(t, err)
	assert.Len(t, result2, 1)
	// last_12_months: 最後の 365 日 = 7 回（-5, -30, -95, -120, -190, -220, -350）、-380 は 365 日超なので除外
	assert.Equal(t, int64(10), result2[0].TotalVisitCount, "period_preset 指定時も total_visit_count = 10 であるべき")
	assert.NotNil(t, result2[0].PeriodVisitCount)
	assert.Equal(t, int64(7), *result2[0].PeriodVisitCount, "last_12_months での period_visit_count = 7")

	// Test 3: last_3_months 指定時
	result3, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
		ClinicID:     clinicID,
		PeriodPreset: "last_3_months",
	})
	assert.NoError(t, err)
	assert.Len(t, result3, 1)
	// last_3_months: 最後の 90 日 = 2 回（5, 30）
	assert.Equal(t, int64(10), result3[0].TotalVisitCount, "total_visit_count は常に 10 であるべき")
	assert.NotNil(t, result3[0].PeriodVisitCount)
	assert.Equal(t, int64(2), *result3[0].PeriodVisitCount, "last_3_months での period_visit_count = 2")
}

// TestFindOwnerLTV_PeriodPriority
// ISSUE-001: from/to > year > period_preset > default の優先順位を検証
func TestFindOwnerLTV_PeriodPriority(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLtvRepository(db)
	ctx := context.Background()

	clinicID := uint64(1)

	// Setup: 複数年にわたるデータ
	owner := &model.Owner{
		ClinicID: clinicID,
		Name:     "Owner B - multi-year",
	}
	if err := db.WithContext(ctx).Create(owner).Error; err != nil {
		t.Fatalf("failed to create owner: %v", err)
	}

	// 2024: 5回、2025: 8回、2026: 3回
	visitDates := []time.Time{
		// 2024: 5 records
		time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 20, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 6, 10, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 9, 30, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC),
		// 2025: 8 records
		time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 2, 5, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 4, 12, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 5, 25, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 6, 20, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 8, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 10, 30, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 11, 10, 0, 0, 0, 0, time.UTC),
		// 2026: 3 records
		time.Date(2026, 2, 5, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC),
	}

	for i, visitDate := range visitDates {
		mr := &model.MedicalRecord{
			ClinicID: clinicID,
			OwnerID:  &owner.ID,
			Date:     visitDate,
		}
		if err := db.WithContext(ctx).Create(mr).Error; err != nil {
			t.Fatalf("failed to create medical record %d: %v", i, err)
		}

		// Create billing record for each medical record so annual_amount is non-zero
		billing := &model.Billing{
			ClinicID:        clinicID,
			MedicalRecordID: &mr.ID,
			OwnerID:         &owner.ID,
			TotalAmount:     1, // ¥1 per visit (so AnnualAmount = visit count)
			Status:          model.BillingStatusCompleted,
			ScheduledDate:   visitDate,
		}
		if err := db.WithContext(ctx).Create(billing).Error; err != nil {
			t.Fatalf("failed to create billing record %d: %v", i, err)
		}
	}

	// Test 1: from/to 指定時（優先度最高）
	from := "2025-01-01"
	to := "2025-12-31"
	result1, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
		ClinicID:     clinicID,
		From:         &from,
		To:           &to,
		PeriodPreset: "last_3_months", // これは無視されるべき
	})
	assert.NoError(t, err)
	assert.Len(t, result1, 1)
	// 2025 年は 8 回（period_preset は無視）
	assert.Equal(t, int64(8), *result1[0].AnnualAmount)

	// Test 2: year 指定（from/to なし）
	year := 2024
	result2, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
		ClinicID:     clinicID,
		Year:         &year,
		PeriodPreset: "last_3_months", // これは無視されるべき
	})
	assert.NoError(t, err)
	assert.Len(t, result2, 1)
	// 2024 年は 5 回（period_preset は無視）
	assert.Equal(t, int64(5), *result2[0].AnnualAmount)

	// Test 3: period_preset 指定（year/from/to なし）
	result3, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
		ClinicID:     clinicID,
		PeriodPreset: "last_12_months",
	})
	assert.NoError(t, err)
	assert.Len(t, result3, 1)
	// last_12_months: 最後の 365 日（今の状態では 2026 年のほぼすべてと 2025 年の一部）
	// = 4 回（2026 年 4 回）※ 近い日付次第
	assert.NotNil(t, result3[0].PeriodVisitCount)
}

// TestFindOwnerLTV_LastVisitBucketBoundaries
// ISSUE-003: last_visit_bucket の境界値 90 / 180 / 365 日を検証
func TestFindOwnerLTV_LastVisitBucketBoundaries(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLtvRepository(db)
	ctx := context.Background()

	clinicID := uint64(1)
	now := time.Now()

	testCases := []struct {
		name           string
		ownerName      string
		lastVisitDays  int // 負の日数: 過去何日前か
		expectedBucket string
	}{
		{"within_3m (5 days ago)", "Owner Within3m", -5, "within_3m"},
		{"within_3m (89 days ago)", "Owner Within3m_89", -89, "within_3m"},
		{"over_3m (90 days ago)", "Owner Over3m_90", -90, "over_3m"},
		{"over_3m (179 days ago)", "Owner Over3m_179", -179, "over_3m"},
		{"over_6m (180 days ago)", "Owner Over6m_180", -180, "over_6m"},
		{"over_6m (364 days ago)", "Owner Over6m_364", -364, "over_6m"},
		{"over_1y (365 days ago)", "Owner Over1y_365", -365, "over_1y"},
		{"over_1y (400 days ago)", "Owner Over1y_400", -400, "over_1y"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			owner := &model.Owner{
				ClinicID: clinicID,
				Name:     tc.ownerName,
			}
			if err := db.WithContext(ctx).Create(owner).Error; err != nil {
				t.Fatalf("failed to create owner: %v", err)
			}

			visitDate := now.AddDate(0, 0, tc.lastVisitDays)
			mr := &model.MedicalRecord{
				ClinicID: clinicID,
				OwnerID:  &owner.ID,
				Date:     visitDate,
			}
			if err := db.WithContext(ctx).Create(mr).Error; err != nil {
				t.Fatalf("failed to create medical record: %v", err)
			}

			result, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
				ClinicID:    clinicID,
				IncludeZero: true,
			})
			assert.NoError(t, err)
			// 複数 Owner があるので、該当オーナーを検索
			var ownerResult *OwnerLTVRow
			for i := range result {
				if result[i].OwnerName == tc.ownerName {
					ownerResult = &result[i]
					break
				}
			}
			assert.NotNil(t, ownerResult, "owner %s not found in results", tc.ownerName)
			assert.NotNil(t, ownerResult.LastVisitBucket, "last_visit_bucket should not be nil")
			assert.Equal(t, tc.expectedBucket, *ownerResult.LastVisitBucket, "incorrect bucket for %s", tc.ownerName)
		})
	}
}

// TestFindOwnerLTV_NoVisitBucket
// ISSUE-003: no_visit 分類を検証
func TestFindOwnerLTV_NoVisitBucket(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLtvRepository(db)
	ctx := context.Background()

	clinicID := uint64(1)

	// 来院なしの Owner
	owner := &model.Owner{
		ClinicID: clinicID,
		Name:     "Owner No Visit",
	}
	if err := db.WithContext(ctx).Create(owner).Error; err != nil {
		t.Fatalf("failed to create owner: %v", err)
	}

	// Test 1: include_no_visit = false (デフォルト)
	result1, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
		ClinicID:       clinicID,
		IncludeNoVisit: false,
	})
	assert.NoError(t, err)
	// Owner が来院なしで include_no_visit = false なら除外される
	assert.Equal(t, 0, len(result1), "no_visit owner should be excluded when include_no_visit=false")

	// Test 2: include_no_visit = true
	result2, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
		ClinicID:       clinicID,
		IncludeNoVisit: true,
		IncludeZero:    true,
	})
	assert.NoError(t, err)
	assert.Len(t, result2, 1)
	assert.NotNil(t, result2[0].LastVisitBucket)
	assert.Equal(t, "no_visit", *result2[0].LastVisitBucket)
}

// TestFindOwnerLTV_IncludeZero
// ISSUE-001: include_zero フラグを検証
func TestFindOwnerLTV_IncludeZero(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLtvRepository(db)
	ctx := context.Background()

	clinicID := uint64(1)

	// 売上 0 円の Owner（来院あり、会計なし）
	owner := &model.Owner{
		ClinicID: clinicID,
		Name:     "Owner Zero Amount",
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

	// Test 1: include_zero = false
	result1, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
		ClinicID:    clinicID,
		IncludeZero: false,
	})
	assert.NoError(t, err)
	// 売上 0 で include_zero=false なら除外
	assert.Equal(t, 0, len(result1), "zero amount owner should be excluded when include_zero=false")

	// Test 2: include_zero = true
	result2, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
		ClinicID:    clinicID,
		IncludeZero: true,
	})
	assert.NoError(t, err)
	assert.Len(t, result2, 1)
	assert.Equal(t, int64(0), *result2[0].AnnualAmount)
}

// TestFindOwnerLTV_AmountBasisSwitching
// ISSUE-001: amount_basis 切り替え（gross_total_amount / paid_amount / net_paid_amount）
func TestFindOwnerLTV_AmountBasisSwitching(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLtvRepository(db)
	ctx := context.Background()

	clinicID := uint64(1)

	// Owner + Medical Record + Billing + Payment
	owner := &model.Owner{
		ClinicID: clinicID,
		Name:     "Owner Amount Basis Test",
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

	// Billing: 10000 円
	billing := &model.Billing{
		ClinicID:        clinicID,
		MedicalRecordID: &mr.ID,
		TotalAmount:     10000,
		Status:          "completed",
	}
	if err := db.WithContext(ctx).Create(billing).Error; err != nil {
		t.Fatalf("failed to create billing: %v", err)
	}

	// Payment: 8000 円
	payment := &model.Payment{
		BillingID:     billing.ID,
		BillingAmount: 8000,
	}
	if err := db.WithContext(ctx).Create(payment).Error; err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	// Refund: 1000 円
	refund := &model.BillingRefund{
		BillingID: billing.ID,
		Amount:    1000,
	}
	if err := db.WithContext(ctx).Create(refund).Error; err != nil {
		t.Fatalf("failed to create refund: %v", err)
	}

	// Test 1: gross_total_amount (default)
	result1, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
		ClinicID:    clinicID,
		AmountBasis: "gross_total_amount",
		IncludeZero: true,
	})
	assert.NoError(t, err)
	assert.Len(t, result1, 1)
	assert.Equal(t, int64(10000), *result1[0].AnnualAmount, "gross_total_amount should be 10000")

	// Test 2: paid_amount
	result2, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
		ClinicID:    clinicID,
		AmountBasis: "paid_amount",
		IncludeZero: true,
	})
	assert.NoError(t, err)
	assert.Len(t, result2, 1)
	assert.Equal(t, int64(8000), *result2[0].AnnualAmount, "paid_amount should be 8000")

	// Test 3: net_paid_amount
	result3, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
		ClinicID:    clinicID,
		AmountBasis: "net_paid_amount",
		IncludeZero: true,
	})
	assert.NoError(t, err)
	assert.Len(t, result3, 1)
	assert.Equal(t, int64(7000), *result3[0].AnnualAmount, "net_paid_amount should be 8000 - 1000 = 7000")
}

// TestFindOwnerLTV_OnlyCompletedBillings
// ISSUE-001: status != completed の会計は集計に含まれない
func TestFindOwnerLTV_OnlyCompletedBillings(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLtvRepository(db)
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
	db := setupTestDB(t)
	repo := NewLtvRepository(db)
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
	db := setupTestDB(t)
	repo := NewLtvRepository(db)
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

// TestFindOwnerLTV_ClinicIDIsolation
// clinic_id による分離を検証
func TestFindOwnerLTV_ClinicIDIsolation(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLtvRepository(db)
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
	db := setupTestDB(t)
	repo := NewLtvRepository(db)
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
	db := setupTestDB(t)
	repo := NewLtvRepository(db)
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
	db := setupTestDB(t)
	repo := NewLtvRepository(db)
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

// setupTestDB はテスト用の DB を初期化してマイグレーションを実行します
func setupTestDB(t *testing.T) *gorm.DB {
	// テスト DB コネクション（docker compose で起動）
	// 実装は環境に合わせて調整
	db := getTestDatabaseConnection(t)

	// AutoMigrate の前に、PostgreSQL カスタム ENUM 型を作成
	// （マイグレーション 001_init.sql で定義されている全 46 型）
	// DROP TYPE IF EXISTS → CREATE TYPE の順序で、既存型を削除してから再作成
	enumTypes := []struct {
		drop   string
		create string
	}{
		// ペット関連
		{"DROP TYPE IF EXISTS pet_status CASCADE", "CREATE TYPE pet_status AS ENUM ('alive', 'deceased')"},
		{"DROP TYPE IF EXISTS pet_gender CASCADE", "CREATE TYPE pet_gender AS ENUM ('male', 'female', 'unknown')"},
		{"DROP TYPE IF EXISTS acquisition_type CASCADE", "CREATE TYPE acquisition_type AS ENUM ('purchased', 'transferred', 'rescued', 'other')"},
		{"DROP TYPE IF EXISTS danger_level CASCADE", "CREATE TYPE danger_level AS ENUM ('low', 'medium', 'high')"},
		{"DROP TYPE IF EXISTS membership_type CASCADE", "CREATE TYPE membership_type AS ENUM ('non_member', 'member', 'deceased', 'transferred')"},
		// マスタ共通
		{"DROP TYPE IF EXISTS inventory_category CASCADE", "CREATE TYPE inventory_category AS ENUM ('medicine', 'consumable', 'food', 'other')"},
		{"DROP TYPE IF EXISTS inventory_status CASCADE", "CREATE TYPE inventory_status AS ENUM ('sufficient', 'low', 'out_of_stock')"},
		{"DROP TYPE IF EXISTS dosage_form CASCADE", "CREATE TYPE dosage_form AS ENUM ('tablet', 'liquid', 'injection', 'topical', 'powder')"},
		{"DROP TYPE IF EXISTS medicine_unit CASCADE", "CREATE TYPE medicine_unit AS ENUM ('per_tablet', 'per_ml', 'per_dose', 'per_gram')"},
		{"DROP TYPE IF EXISTS cage_type CASCADE", "CREATE TYPE cage_type AS ENUM ('icu', 'dog', 'cat', 'general')"},
		{"DROP TYPE IF EXISTS cage_size CASCADE", "CREATE TYPE cage_size AS ENUM ('small', 'medium', 'large')"},
		{"DROP TYPE IF EXISTS body_size CASCADE", "CREATE TYPE body_size AS ENUM ('small', 'medium', 'large')"},
		{"DROP TYPE IF EXISTS billing_unit CASCADE", "CREATE TYPE billing_unit AS ENUM ('per_day', 'per_night')"},
		{"DROP TYPE IF EXISTS target_size CASCADE", "CREATE TYPE target_size AS ENUM ('small', 'medium', 'large', 'cat')"},
		{"DROP TYPE IF EXISTS anesthesia_type CASCADE", "CREATE TYPE anesthesia_type AS ENUM ('none', 'local', 'sedation', 'general')"},
		{"DROP TYPE IF EXISTS vaccine_species CASCADE", "CREATE TYPE vaccine_species AS ENUM ('dog', 'cat', 'both')"},
		// 電子カルテ関連
		{"DROP TYPE IF EXISTS medical_record_status CASCADE", "CREATE TYPE medical_record_status AS ENUM ('draft', 'finalized')"},
		{"DROP TYPE IF EXISTS treatment_item_type CASCADE", "CREATE TYPE treatment_item_type AS ENUM ('consultation', 'procedure', 'medicine', 'other')"},
		{"DROP TYPE IF EXISTS treatment_status CASCADE", "CREATE TYPE treatment_status AS ENUM ('pending', 'completed', 'not_applicable')"},
		{"DROP TYPE IF EXISTS exam_status CASCADE", "CREATE TYPE exam_status AS ENUM ('pending', 'in_progress', 'result_entered', 'completed', 'confirmed')"},
		{"DROP TYPE IF EXISTS exam_result_status CASCADE", "CREATE TYPE exam_result_status AS ENUM ('normal', 'high', 'low')"},
		{"DROP TYPE IF EXISTS next_schedule_type CASCADE", "CREATE TYPE next_schedule_type AS ENUM ('3weeks', '4weeks', '1year', 'other')"},
		{"DROP TYPE IF EXISTS appetite_level CASCADE", "CREATE TYPE appetite_level AS ENUM ('normal', 'increased', 'decreased', 'none')"},
		{"DROP TYPE IF EXISTS water_intake_level CASCADE", "CREATE TYPE water_intake_level AS ENUM ('normal', 'increased', 'decreased', 'none')"},
		{"DROP TYPE IF EXISTS medical_image_type CASCADE", "CREATE TYPE medical_image_type AS ENUM ('xray', 'echo', 'photo', 'endoscope', 'ct', 'mri', 'microscope', 'other')"},
		{"DROP TYPE IF EXISTS estimate_status CASCADE", "CREATE TYPE estimate_status AS ENUM ('draft', 'sent', 'approved', 'rejected')"},
		{"DROP TYPE IF EXISTS confirmation_status CASCADE", "CREATE TYPE confirmation_status AS ENUM ('pending', 'confirmed', 'returned')"},
		{"DROP TYPE IF EXISTS item_category CASCADE", "CREATE TYPE item_category AS ENUM ('examination', 'test', 'procedure', 'surgery', 'medicine', 'food', 'goods', 'other', 'vaccine', 'trimming', 'hotel', 'training')"},
		{"DROP TYPE IF EXISTS item_source CASCADE", "CREATE TYPE item_source AS ENUM ('medical_record', 'manual', 'hospitalization')"},
		// 予約・会計・入院関連
		{"DROP TYPE IF EXISTS visit_type CASCADE", "CREATE TYPE visit_type AS ENUM ('first', 'revisit')"},
		{"DROP TYPE IF EXISTS reservation_status CASCADE", "CREATE TYPE reservation_status AS ENUM ('confirmed', 'pending', 'cancelled', 'checked_in', 'in_consultation', 'accounting', 'completed', 'no_show')"},
		{"DROP TYPE IF EXISTS staff_type CASCADE", "CREATE TYPE staff_type AS ENUM ('doctor', 'nurse', 'trimmer', 'resource')"},
		{"DROP TYPE IF EXISTS reservation_source CASCADE", "CREATE TYPE reservation_source AS ENUM ('manual', 'line')"},
		{"DROP TYPE IF EXISTS billing_status CASCADE", "CREATE TYPE billing_status AS ENUM ('waiting', 'completed', 'cancelled', 'pending')"},
		{"DROP TYPE IF EXISTS hospitalization_type CASCADE", "CREATE TYPE hospitalization_type AS ENUM ('hospitalization', 'hotel')"},
		{"DROP TYPE IF EXISTS hospitalization_status CASCADE", "CREATE TYPE hospitalization_status AS ENUM ('admitted', 'discharged', 'reserved')"},
		{"DROP TYPE IF EXISTS care_plan_type CASCADE", "CREATE TYPE care_plan_type AS ENUM ('food', 'medicine', 'treatment', 'instruction', 'item')"},
		{"DROP TYPE IF EXISTS care_plan_status CASCADE", "CREATE TYPE care_plan_status AS ENUM ('active', 'completed', 'discontinued')"},
		{"DROP TYPE IF EXISTS care_log_type CASCADE", "CREATE TYPE care_log_type AS ENUM ('food', 'excretion', 'medicine', 'treatment', 'other')"},
		{"DROP TYPE IF EXISTS care_log_status CASCADE", "CREATE TYPE care_log_status AS ENUM ('completed', 'partial', 'skipped')"},
		{"DROP TYPE IF EXISTS plan_timing CASCADE", "CREATE TYPE plan_timing AS ENUM ('morning', 'noon', 'night')"},
		{"DROP TYPE IF EXISTS body_weight_unit CASCADE", "CREATE TYPE body_weight_unit AS ENUM ('Kg', 'g')"},
		// トリミング・シフト関連
		{"DROP TYPE IF EXISTS reservation_type_category CASCADE", "CREATE TYPE reservation_type_category AS ENUM ('general', 'trimming')"},
		{"DROP TYPE IF EXISTS payment_method CASCADE", "CREATE TYPE payment_method AS ENUM ('cash', 'credit_card', 'electronic_money')"},
		{"DROP TYPE IF EXISTS shift_type CASCADE", "CREATE TYPE shift_type AS ENUM ('full', 'morning', 'afternoon', 'off', 'paid_leave')"},
		{"DROP TYPE IF EXISTS tax_type CASCADE", "CREATE TYPE tax_type AS ENUM ('included', 'excluded', 'exempt')"},
	}
	for _, et := range enumTypes {
		db.Exec(et.drop) // ignore errors on DROP
		if err := db.Exec(et.create).Error; err != nil {
			t.Fatalf("failed to create ENUM type: %v", err)
		}
	}

	// Truncate tables to ensure clean state (data isolation between tests)
	db.Exec("TRUNCATE TABLE billing_refunds CASCADE")
	db.Exec("TRUNCATE TABLE payments CASCADE")
	db.Exec("TRUNCATE TABLE billings CASCADE")
	db.Exec("TRUNCATE TABLE medical_records CASCADE")
	db.Exec("TRUNCATE TABLE owners CASCADE")

	db.AutoMigrate(
		&model.Owner{},
		&model.MedicalRecord{},
		&model.Billing{},
		&model.Payment{},
		&model.BillingRefund{},
	)
	if db.Error != nil {
		t.Fatalf("failed to migrate test db: %v", db.Error)
	}
	return db
}

// getTestDatabaseConnection はテスト用の DB コネクションを取得（環境変数から）
func getTestDatabaseConnection(t *testing.T) *gorm.DB {
	// 環境変数から DB パラメータを取得（デフォルト: ekarte_db）
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "db"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "ekarte_user"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "ekarte_password"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "ekarte_db"
	}

	testDBName := dbName + "_test"
	mainDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)

	// まず本番DBに接続してテストDBを作成
	mainDB, err := gorm.Open(postgres.Open(mainDSN), &gorm.Config{})
	if err != nil {
		t.Logf("warning: failed to connect to main db: %v", err)
	} else {
		mainDB.Exec("CREATE DATABASE " + testDBName)
	}

	// テストDB接続
	testDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, testDBName)
	if envDSN := os.Getenv("TEST_DATABASE_URL"); envDSN != "" {
		testDSN = envDSN
	}
	db, err := gorm.Open(postgres.Open(testDSN), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}
	return db
}
