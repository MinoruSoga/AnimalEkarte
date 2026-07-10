package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestEscapeLikePattern(t *testing.T) {
	assert.Equal(t, `100\%\_\\`, escapeLikePattern(`100%_\`))
	assert.Equal(t, `normal`, escapeLikePattern(`normal`))
}

func TestQuotePostgresIdentifier(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		want       string
	}{
		{
			name:       "simple identifier",
			identifier: "ekarte_db_test",
			want:       `"ekarte_db_test"`,
		},
		{
			name:       "embedded quote is escaped",
			identifier: `ekarte"db_test`,
			want:       `"ekarte""db_test"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, quotePostgresIdentifier(tt.identifier))
		})
	}
}

func TestIsDuplicateDatabaseError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "duplicate_database is true",
			err:  &pgconn.PgError{Code: "42P04"},
			want: true,
		},
		{
			name: "wrapped duplicate_database is true",
			err:  fmt.Errorf("create failed: %w", &pgconn.PgError{Code: "42P04"}),
			want: true,
		},
		{
			name: "different pg error is false",
			err:  &pgconn.PgError{Code: "23505"},
			want: false,
		},
		{
			name: "plain error is false",
			err:  errors.New("plain error"),
			want: false,
		},
		{
			name: "nil is false",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isDuplicateDatabaseError(tt.err))
		})
	}
}

func TestFindOwnerLTV_SearchEscapesLikeWildcards(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLtvRepository(db)
	ctx := context.Background()
	clinicID := uint64(1)

	literalPercentOwner := &model.Owner{
		ClinicID: clinicID,
		Name:     "100% literal owner",
	}
	similarOwner := &model.Owner{
		ClinicID: clinicID,
		Name:     "100X wildcard owner",
	}
	require.NoError(t, db.WithContext(ctx).Create(literalPercentOwner).Error)
	require.NoError(t, db.WithContext(ctx).Create(similarOwner).Error)

	rows, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
		ClinicID:       clinicID,
		Search:         "100%",
		IncludeZero:    true,
		IncludeNoVisit: true,
	})

	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, literalPercentOwner.ID, rows[0].OwnerID)
}

func TestFindOwnerLTV_SearchNormalizesKana(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLtvRepository(db)
	ctx := context.Background()
	clinicID := uint64(1)

	// カタカナ名の飼主（DB はカタカナ登録）
	katakanaOwner := &model.Owner{ClinicID: clinicID, Name: "ヤマダタロウ"}
	// ひらがな名の飼主
	hiraganaOwner := &model.Owner{ClinicID: clinicID, Name: "さとうけんじ"}
	require.NoError(t, db.WithContext(ctx).Create(katakanaOwner).Error)
	require.NoError(t, db.WithContext(ctx).Create(hiraganaOwner).Error)

	// ひらがな検索でカタカナ登録データがヒットすること
	rows, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
		ClinicID:       clinicID,
		Search:         "やまだ",
		IncludeZero:    true,
		IncludeNoVisit: true,
	})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, katakanaOwner.ID, rows[0].OwnerID)

	// カタカナ検索でひらがな登録データがヒットすること
	rows2, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
		ClinicID:       clinicID,
		Search:         "サトウ",
		IncludeZero:    true,
		IncludeNoVisit: true,
	})
	require.NoError(t, err)
	require.Len(t, rows2, 1)
	assert.Equal(t, hiraganaOwner.ID, rows2[0].OwnerID)
}

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

	// Test 4: 期間付き net_paid_amount の HAVING でも金額閾値をバインドして絞り込める
	year := time.Now().Year()
	minAmount := int64(6000)
	result4, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
		ClinicID:       clinicID,
		AmountBasis:    "net_paid_amount",
		Year:           &year,
		MinTotalAmount: &minAmount,
		IncludeZero:    true,
	})
	assert.NoError(t, err)
	assert.Len(t, result4, 1)
	assert.Equal(t, int64(7000), *result4[0].AnnualAmount)

	minAmount = 7500
	result5, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
		ClinicID:       clinicID,
		AmountBasis:    "net_paid_amount",
		Year:           &year,
		MinTotalAmount: &minAmount,
		IncludeZero:    true,
	})
	assert.NoError(t, err)
	assert.Len(t, result5, 0)
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
	db := setupTestDB(t)
	repo := NewLtvRepository(db)
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
	db := setupTestDB(t)
	repo := NewLtvRepository(db)
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
	db := setupTestDB(t)
	repo := NewLtvRepository(db)
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
	db := setupTestDB(t)
	repo := NewLtvRepository(db)
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

// テストDB接続・ENUM型・共有ベースモデルの AutoMigrate はプロセス全体（package repository の
// go test 実行単位）で一度だけ実行する（sharedTestSchemaOnce/TestMain）。130+ ファイル・158+ 箇所の
// setupTestDB(t) 呼び出し毎にこれらを繰り返すと、接続確立×2・ENUM存在チェック46回・AutoMigrate
// スキーマ内省クエリが呼び出し回数分積み上がり、repository テストスイート全体の支配的コストになる
// （2026-07 計測: ローカル 191s → 本最適化後は setupTestDB 呼び出し側を一切変更せず短縮）。
// TRUNCATE のみ setupTestDB 内で呼び出し毎に実行し、テスト間データ分離を維持する。
var (
	sharedTestDB         *gorm.DB
	sharedTestDBOnce     sync.Once
	sharedTestDBErr      error
	sharedTestSchemaOnce sync.Once
	sharedTestSchemaErr  error
)

// TestMain は internal/repository package の全テストで共有する DB 接続プールを管理する。
// 個々のテストは接続を閉じず（sharedTestDBOnce で一度だけ確立・全テストで再利用）、プロセス終了時に
// ここで一度だけ閉じる。
func TestMain(m *testing.M) {
	code := m.Run()
	if sharedTestDB != nil {
		if sqlDB, err := sharedTestDB.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}
	os.Exit(code)
}

// setupTestDB はテスト用の DB を返し、共有ベーステーブルを TRUNCATE してクリーンな状態にします。
// DB接続確立・ENUM型作成・ベースモデルの AutoMigrate はプロセス全体で一度だけ実行されます。
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := getTestDatabaseConnection(t)

	sharedTestSchemaOnce.Do(func() {
		sharedTestSchemaErr = setupSharedTestSchema(db)
	})
	if sharedTestSchemaErr != nil {
		t.Fatalf("failed to set up shared test schema: %v", sharedTestSchemaErr)
	}

	// Truncate tables to ensure clean state (data isolation between tests)
	db.Exec("TRUNCATE TABLE billing_refunds CASCADE")
	db.Exec("TRUNCATE TABLE payments CASCADE")
	db.Exec("TRUNCATE TABLE billings CASCADE")
	db.Exec("TRUNCATE TABLE medical_records CASCADE")
	db.Exec("TRUNCATE TABLE owners CASCADE")

	return db
}

// setupIsolatedTestDB は setupTestDB と異なり、プロセス全体で共有しない「呼び出し毎に完全に新しい」
// DB 接続を返す（最適化前の setupTestDB と同じ挙動）。
//
// checkup_field_results/checkup_type_fields は checkup_field_repository_test.go（AutoMigrate 由来
// スキーマ）・checkup_field_cascade_test.go（migration 010 の実 DDL）・
// checkup_field_composite_fk_test.go（010 実 DDL + migration 012 複合 FK）という 3 種の異なる
// ヘルパーが同じテーブルを意図的に毎回 DROP+CREATE し合う（migration drift 検出が目的で、
// 挙動として必須）。この cluster に setupTestDB の共有コネクションプールを使うと、いずれかの
// ヘルパーが DROP TABLE/DROP TYPE した瞬間に、別テストが既に保持していた同一物理コネクション上の
// キャッシュ済み prepared statement（古いテーブル/型 OID 参照）が
// "cache lookup failed"（SQLSTATE XX000）で壊れる。3 ヘルパーの意図的な毎回 DROP+CREATE は
// 統合できないため、この cluster だけは共有プールから外し、テスト毎に使い捨ての新規コネクションを
// 割り当てることでキャッシュ汚染を根本的に回避する（対象は少数のためスループット影響は軽微）。
func setupIsolatedTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := connectTestDatabase()
	if err != nil {
		t.Fatalf("failed to connect to isolated test db: %v", err)
	}
	if sqlDB, derr := db.DB(); derr == nil {
		t.Cleanup(func() { _ = sqlDB.Close() })
	}

	if err := setupSharedTestSchema(db); err != nil {
		t.Fatalf("failed to set up shared test schema (isolated): %v", err)
	}

	db.Exec("TRUNCATE TABLE billing_refunds CASCADE")
	db.Exec("TRUNCATE TABLE payments CASCADE")
	db.Exec("TRUNCATE TABLE billings CASCADE")
	db.Exec("TRUNCATE TABLE medical_records CASCADE")
	db.Exec("TRUNCATE TABLE owners CASCADE")

	return db
}

// testSchemaEnumType is one hand-maintained ENUM type double used by setupSharedTestSchema.
// Kept as a package-level type/var (rather than a func-local literal) so
// test_schema_enum_parity_test.go (G12-2) can compare it against 001_init.sql directly instead
// of re-parsing Go source via go/ast.
type testSchemaEnumType struct {
	name   string
	create string
}

// sharedTestSchemaEnumTypes hand-duplicates every PostgreSQL ENUM type from 001_init.sql
// (54 types total, 2026-07-04 consolidated migration + 009 #201 薬量計算 4 型を含む）。
// model.Medicine が calculation_type を持つため、本 setup を使う全テストの medicines
// AutoMigrate に medicine_calculation_type が必須（欠落で CREATE TABLE 失敗）。
//
// G12-2 (BE-refactor.md): this list previously drifted from 001_init.sql — item_source was
// missing 'trimming' (blocking any integration test that persists a trimming billing_items
// row via billing_item_repository.go's Source: model.ItemSourceTrimming) and 4 whole types
// were absent. test_schema_enum_parity_test.go now gates this list against 001_init.sql on
// every `go test ./internal/repository/...` run so it cannot silently drift again.
//
// checkup_field_type is included here even though checkup_field_repository_test.go (and its
// sibling _cascade_test.go/_composite_fk_test.go) also DROP+CREATE it: those three helpers run
// on setupIsolatedTestDB (a throw-away connection per call, see repository/CLAUDE.md), and no
// setupTestDB-based (shared-connection) test AutoMigrates CheckupTypeField/CheckupFieldResult,
// so there is no column depending on this type via the shared connection. Package tests never
// run in parallel (no t.Parallel() in this package), so the isolated helpers' own DROP+CREATE
// cannot race this one either — adding it here is a no-op once created, not a collision.
var sharedTestSchemaEnumTypes = []testSchemaEnumType{
	// ペット関連
	{"pet_status", "CREATE TYPE pet_status AS ENUM ('alive', 'deceased')"},
	{"pet_gender", "CREATE TYPE pet_gender AS ENUM ('male', 'female', 'unknown')"},
	{"acquisition_type", "CREATE TYPE acquisition_type AS ENUM ('purchased', 'transferred', 'rescued', 'other')"},
	{"danger_level", "CREATE TYPE danger_level AS ENUM ('low', 'medium', 'high')"},
	{"membership_type", "CREATE TYPE membership_type AS ENUM ('non_member', 'member', 'deceased', 'transferred')"},
	// マスタ共通
	{"inventory_category", "CREATE TYPE inventory_category AS ENUM ('medicine', 'consumable', 'food', 'other')"},
	{"inventory_status", "CREATE TYPE inventory_status AS ENUM ('sufficient', 'low', 'out_of_stock')"},
	{"dosage_form", "CREATE TYPE dosage_form AS ENUM ('tablet', 'liquid', 'injection', 'topical', 'powder')"},
	{"medicine_unit", "CREATE TYPE medicine_unit AS ENUM ('per_tablet', 'per_ml', 'per_dose', 'per_gram')"},
	// #201 薬量自動計算（migration 009）: medicines.calculation_type + medicine_dose_params 用
	{"medicine_calculation_type", "CREATE TYPE medicine_calculation_type AS ENUM ('none', 'per_weight')"},
	{"medicine_dose_basis", "CREATE TYPE medicine_dose_basis AS ENUM ('per_administration', 'per_day')"},
	{"medicine_rounding_mode", "CREATE TYPE medicine_rounding_mode AS ENUM ('up', 'down', 'nearest')"},
	{"medicine_dose_species", "CREATE TYPE medicine_dose_species AS ENUM ('dog', 'cat')"},
	{"cage_type", "CREATE TYPE cage_type AS ENUM ('icu', 'dog', 'cat', 'general')"},
	{"cage_size", "CREATE TYPE cage_size AS ENUM ('small', 'medium', 'large')"},
	{"body_size", "CREATE TYPE body_size AS ENUM ('small', 'medium', 'large')"},
	{"billing_unit", "CREATE TYPE billing_unit AS ENUM ('per_day', 'per_night')"},
	{"target_size", "CREATE TYPE target_size AS ENUM ('small', 'medium', 'large', 'cat')"},
	{"anesthesia_type", "CREATE TYPE anesthesia_type AS ENUM ('none', 'local', 'sedation', 'general')"},
	{"vaccine_species", "CREATE TYPE vaccine_species AS ENUM ('dog', 'cat', 'both')"},
	// 電子カルテ関連
	{"medical_record_status", "CREATE TYPE medical_record_status AS ENUM ('draft', 'finalized')"},
	{"treatment_item_type", "CREATE TYPE treatment_item_type AS ENUM ('consultation', 'procedure', 'medicine', 'other')"},
	{"treatment_status", "CREATE TYPE treatment_status AS ENUM ('pending', 'completed', 'not_applicable')"},
	{"exam_status", "CREATE TYPE exam_status AS ENUM ('pending', 'in_progress', 'result_entered', 'completed', 'confirmed')"},
	{"exam_result_status", "CREATE TYPE exam_result_status AS ENUM ('normal', 'high', 'low')"},
	{"next_schedule_type", "CREATE TYPE next_schedule_type AS ENUM ('3weeks', '4weeks', '1year', 'other')"},
	{"appetite_level", "CREATE TYPE appetite_level AS ENUM ('normal', 'increased', 'decreased', 'none')"},
	{"water_intake_level", "CREATE TYPE water_intake_level AS ENUM ('normal', 'increased', 'decreased', 'none')"},
	{"medical_image_type", "CREATE TYPE medical_image_type AS ENUM ('xray', 'echo', 'photo', 'endoscope', 'ct', 'mri', 'microscope', 'other')"},
	{"estimate_status", "CREATE TYPE estimate_status AS ENUM ('draft', 'sent', 'approved', 'rejected')"},
	{"confirmation_status", "CREATE TYPE confirmation_status AS ENUM ('pending', 'confirmed', 'returned')"},
	{"item_category", "CREATE TYPE item_category AS ENUM ('examination', 'test', 'procedure', 'surgery', 'medicine', 'food', 'goods', 'other', 'vaccine', 'trimming', 'hotel', 'training')"},
	// G12-2: 'trimming' was missing — billing_item_repository.go:271 persists
	// Source: model.ItemSourceTrimming, so its integration path was untestable under this schema.
	{"item_source", "CREATE TYPE item_source AS ENUM ('medical_record', 'manual', 'hospitalization', 'trimming')"},
	{"campaign_discount_type", "CREATE TYPE campaign_discount_type AS ENUM ('rate', 'amount')"},
	// 予約・会計・入院関連
	{"visit_type", "CREATE TYPE visit_type AS ENUM ('first', 'revisit')"},
	{"reservation_status", "CREATE TYPE reservation_status AS ENUM ('confirmed', 'pending', 'cancelled', 'checked_in', 'in_consultation', 'accounting', 'completed', 'no_show')"},
	{"staff_type", "CREATE TYPE staff_type AS ENUM ('doctor', 'nurse', 'trimmer', 'resource')"},
	{"reservation_source", "CREATE TYPE reservation_source AS ENUM ('manual', 'line')"},
	{"billing_status", "CREATE TYPE billing_status AS ENUM ('waiting', 'completed', 'cancelled', 'pending')"},
	{"hospitalization_type", "CREATE TYPE hospitalization_type AS ENUM ('hospitalization', 'hotel')"},
	{"hospitalization_status", "CREATE TYPE hospitalization_status AS ENUM ('admitted', 'discharged', 'reserved')"},
	{"care_plan_type", "CREATE TYPE care_plan_type AS ENUM ('food', 'medicine', 'treatment', 'instruction', 'item')"},
	{"care_plan_status", "CREATE TYPE care_plan_status AS ENUM ('active', 'completed', 'discontinued')"},
	{"care_log_type", "CREATE TYPE care_log_type AS ENUM ('food', 'excretion', 'medicine', 'treatment', 'other')"},
	{"care_log_status", "CREATE TYPE care_log_status AS ENUM ('completed', 'partial', 'skipped')"},
	{"plan_timing", "CREATE TYPE plan_timing AS ENUM ('morning', 'noon', 'night')"},
	{"body_weight_unit", "CREATE TYPE body_weight_unit AS ENUM ('Kg', 'g')"},
	// トリミング・シフト関連
	{"reservation_type_category", "CREATE TYPE reservation_type_category AS ENUM ('general', 'trimming')"},
	{"payment_method", "CREATE TYPE payment_method AS ENUM ('cash', 'credit_card', 'electronic_money', 'bank_transfer')"},
	{"shift_type", "CREATE TYPE shift_type AS ENUM ('full', 'morning', 'afternoon', 'off', 'paid_leave')"},
	{"tax_type", "CREATE TYPE tax_type AS ENUM ('included', 'excluded', 'exempt')"},
	// lab_import（検査結果取込ジョブ）関連
	{"lab_import_job_status", "CREATE TYPE lab_import_job_status AS ENUM ('received', 'validated', 'mapped', 'persisted', 'duplicate', 'needs_review', 'failed')"},
	{"lab_import_source_type", "CREATE TYPE lab_import_source_type AS ENUM ('fixture', 'drwan', 'manual')"},
	// #211 健診パッケージ（migration 010 → 001_init.sql 統合済み）
	{"checkup_field_type", "CREATE TYPE checkup_field_type AS ENUM ('number', 'single_select', 'multi_select', 'boolean', 'checklist', 'text')"},
}

// enumValueRe extracts the ordered list of quoted ENUM value literals out of a
// "CREATE TYPE ... AS ENUM ('a', 'b', ...)" definition string.
var enumValueRe = regexp.MustCompile(`'[^']*'`)

// reconcileEnumTypeDefinition self-heals a stale ekarte_db_test so sharedTestSchemaEnumTypes
// edits (G12-2: e.g. item_source lacking 'trimming') take effect without a manual DB reset. A
// previous IF NOT EXISTS guard silently kept stale definitions forever once a type had been
// created once in the test DB.
//
// It prefers the non-destructive path: if the existing value set is an unchanged, order-preserving
// prefix of the new definition (a pure append — the only kind of drift this task actually hit),
// it widens the type in place with ALTER TYPE ... ADD VALUE. An earlier version of this function
// unconditionally did DROP TYPE ... CASCADE + recreate for any mismatch; verified empirically
// (scoped test run) that this transiently breaks any already-provisioned column of the type —
// billing_item_lstep_queries_test.go / billing_item_repository_tx_atomicity_test.go both assume
// billing_items.source already exists and do not themselves AutoMigrate(&model.BillingItem{}),
// so a blanket CASCADE drop leaves them broken until some other test file happens to run its own
// AutoMigrate first. ALTER TYPE ADD VALUE avoids that class of collateral breakage entirely.
//
// DROP+recreate remains the fallback for genuinely incompatible drift (reordered/removed/renamed
// values) — none of the 54 types hit that case for G12-2, but a future migration edit could.
func reconcileEnumTypeDefinition(db *gorm.DB, name, create string) error {
	var existing []string
	if err := db.Raw(`
		SELECT e.enumlabel FROM pg_type t
		JOIN pg_enum e ON t.oid = e.enumtypid
		WHERE t.typname = ?
		ORDER BY e.enumsortorder`, name).Scan(&existing).Error; err != nil {
		return fmt.Errorf("failed to inspect existing ENUM %s: %w", name, err)
	}

	expected := enumValueRe.FindAllString(create, -1)

	if len(existing) == 0 {
		if err := db.Exec(create).Error; err != nil {
			return fmt.Errorf("failed to create ENUM type %s: %w", name, err)
		}
		return nil
	}

	if enumValuesEqual(existing, expected) {
		return nil
	}

	if appended, ok := enumAppendedValues(existing, expected); ok {
		for _, v := range appended {
			if err := db.Exec(fmt.Sprintf("ALTER TYPE %s ADD VALUE IF NOT EXISTS %s", name, v)).Error; err != nil {
				return fmt.Errorf("failed to widen ENUM type %s with value %s: %w", name, v, err)
			}
		}
		return nil
	}

	if err := db.Exec(fmt.Sprintf("DROP TYPE IF EXISTS %s CASCADE", name)).Error; err != nil {
		return fmt.Errorf("failed to drop stale ENUM type %s: %w", name, err)
	}
	if err := db.Exec(create).Error; err != nil {
		return fmt.Errorf("failed to create ENUM type %s: %w", name, err)
	}
	return nil
}

// enumValuesEqual reports whether existing (unquoted labels from pg_enum) exactly matches
// expectedQuoted (quoted literals extracted from a CREATE TYPE string), in the same order.
func enumValuesEqual(existing, expectedQuoted []string) bool {
	if len(existing) != len(expectedQuoted) {
		return false
	}
	for i, v := range expectedQuoted {
		if "'"+existing[i]+"'" != v {
			return false
		}
	}
	return true
}

// enumAppendedValues reports whether expectedQuoted equals existing with one or more values
// appended at the end (an order-preserving prefix match), returning just the appended
// (still-quoted) values in order. Returns ok=false for any reorder/removal/rename.
func enumAppendedValues(existing, expectedQuoted []string) (appended []string, ok bool) {
	if len(expectedQuoted) <= len(existing) {
		return nil, false
	}
	for i, v := range existing {
		if "'"+v+"'" != expectedQuoted[i] {
			return nil, false
		}
	}
	return expectedQuoted[len(existing):], true
}

// TestEnumValuesEqual_TestEnumAppendedValues pins reconcileEnumTypeDefinition's pure helpers —
// the logic that decides between a non-destructive ALTER TYPE ADD VALUE and a destructive
// DROP+recreate (G12-2).
func TestEnumValuesEqual_TestEnumAppendedValues(t *testing.T) {
	t.Run("enumValuesEqual: identical values match", func(t *testing.T) {
		assert.True(t, enumValuesEqual([]string{"manual", "trimming"}, []string{"'manual'", "'trimming'"}))
	})
	t.Run("enumValuesEqual: different length does not match", func(t *testing.T) {
		assert.False(t, enumValuesEqual([]string{"manual"}, []string{"'manual'", "'trimming'"}))
	})
	t.Run("enumValuesEqual: same length different value does not match", func(t *testing.T) {
		assert.False(t, enumValuesEqual([]string{"manual", "hospitalization"}, []string{"'manual'", "'trimming'"}))
	})

	t.Run("enumAppendedValues: pure trailing append is detected (item_source G12-2 case)", func(t *testing.T) {
		appended, ok := enumAppendedValues(
			[]string{"medical_record", "manual", "hospitalization"},
			[]string{"'medical_record'", "'manual'", "'hospitalization'", "'trimming'"},
		)
		assert.True(t, ok)
		assert.Equal(t, []string{"'trimming'"}, appended)
	})
	t.Run("enumAppendedValues: no new values is not an append", func(t *testing.T) {
		_, ok := enumAppendedValues([]string{"a", "b"}, []string{"'a'", "'b'"})
		assert.False(t, ok)
	})
	t.Run("enumAppendedValues: reordered values is not a pure append", func(t *testing.T) {
		_, ok := enumAppendedValues([]string{"a", "b"}, []string{"'b'", "'a'", "'c'"})
		assert.False(t, ok)
	})
	t.Run("enumAppendedValues: removed value is not a pure append", func(t *testing.T) {
		_, ok := enumAppendedValues([]string{"a", "b"}, []string{"'a'"})
		assert.False(t, ok)
	})
}

// setupSharedTestSchema は PostgreSQL カスタム ENUM 型の作成とベースモデルの AutoMigrate を行います。
// setupTestDB から sharedTestSchemaOnce 経由でプロセス全体につき一度だけ呼ばれます。
func setupSharedTestSchema(db *gorm.DB) error {
	// AutoMigrate の前に、PostgreSQL カスタム ENUM 型を作成する（sharedTestSchemaEnumTypes 参照）。
	for _, et := range sharedTestSchemaEnumTypes {
		if err := reconcileEnumTypeDefinition(db, et.name, et.create); err != nil {
			return err
		}
	}

	if err := db.AutoMigrate(
		&model.Owner{},
		&model.MedicalRecord{},
		&model.Billing{},
		&model.Payment{},
		&model.BillingRefund{},
	); err != nil {
		return fmt.Errorf("failed to migrate test db: %w", err)
	}
	return nil
}

// getTestDatabaseConnection はテスト用の DB コネクションを返す。接続確立（テストDB存在確認込み）は
// sharedTestDBOnce によりプロセス全体で一度だけ行われ、以降の呼び出しは共有プールを再利用する。
func getTestDatabaseConnection(t *testing.T) *gorm.DB {
	t.Helper()
	sharedTestDBOnce.Do(func() {
		sharedTestDB, sharedTestDBErr = connectTestDatabase()
	})
	if sharedTestDBErr != nil {
		t.Fatalf("failed to connect to test db: %v", sharedTestDBErr)
	}
	return sharedTestDB
}

// connectTestDatabase はテスト用 DB への接続を確立する（sharedTestDBOnce によりプロセス全体で一度だけ呼ばれる）。
func connectTestDatabase() (*gorm.DB, error) {
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
		fmt.Fprintf(os.Stderr, "warning: failed to connect to main db: %v\n", err)
	} else {
		if err := ensureTestDatabaseExists(mainDB, testDBName); err != nil {
			return nil, err
		}
		// 接続リーク防止: mainDB は CREATE DATABASE 専用。閉じないと接続が漏れ、
		// PostgreSQL の max_connections を使い切る（FATAL: sorry, too many clients already / SQLSTATE 53300）。
		if sqlMainDB, derr := mainDB.DB(); derr == nil {
			_ = sqlMainDB.Close()
		}
	}

	// テストDB接続
	// 共有プール上で ENUM/テーブルの DROP+CREATE を毎テスト実行すると、サーバサイド prepared
	// statement キャッシュ（pgx デフォルト cache_statement モード）が古い型/リレーション OID を
	// 保持し続け "cache lookup failed" (SQLSTATE XX000) で失敗する。この対策としては
	// setupTestDB 内で全 ENUM を一度きり idempotent 作成するのに加え、DROP+CREATE を行う
	// 個別ヘルパー（checkup_field_repository_test.go / medicine_dose_param_clinic_isolation_test.go）
	// 側もプロセス全体で一度だけ実行するよう sync.Once 化した。これによりプロセス起動後は
	// スキーマが不変となるため、接続プロトコルは pgx デフォルト（cache_statement、最速）のままでよい。
	testDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, testDBName)
	if envDSN := os.Getenv("TEST_DATABASE_URL"); envDSN != "" {
		testDSN = envDSN
	}
	db, err := gorm.Open(postgres.Open(testDSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test db: %w", err)
	}
	// 接続枯渇防止: このプールは全テストで共有する唯一の接続で、プロセス終了時に TestMain が一度だけ閉じる
	// （テスト毎に開閉すると、full suite で接続確立オーバーヘッドが呼び出し回数分積み上がる）。
	if sqlDB, derr := db.DB(); derr == nil {
		sqlDB.SetMaxOpenConns(10)
		sqlDB.SetMaxIdleConns(2)
	}
	return db, nil
}

func ensureTestDatabaseExists(mainDB *gorm.DB, testDBName string) error {
	lockKey := "setup_test_database:" + testDBName
	if err := mainDB.Exec("SELECT pg_advisory_lock(hashtext(?))", lockKey).Error; err != nil {
		return fmt.Errorf("failed to acquire test database creation lock: %w", err)
	}
	defer func() {
		_ = mainDB.Exec("SELECT pg_advisory_unlock(hashtext(?))", lockKey).Error
	}()

	var exists bool
	if err := mainDB.Raw("SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = ?)", testDBName).Scan(&exists).Error; err != nil {
		return fmt.Errorf("failed to check test database existence: %w", err)
	}
	if exists {
		return nil
	}

	if err := mainDB.Exec("CREATE DATABASE " + quotePostgresIdentifier(testDBName)).Error; err != nil {
		if isDuplicateDatabaseError(err) {
			return nil
		}
		return fmt.Errorf("failed to create test database %s: %w", testDBName, err)
	}
	return nil
}

func quotePostgresIdentifier(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}

func isDuplicateDatabaseError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "42P04"
}
