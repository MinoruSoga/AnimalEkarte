package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
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

// setupTestDB はテスト用の DB を初期化してマイグレーションを実行します
func setupTestDB(t *testing.T) *gorm.DB {
	// テスト DB コネクション（docker compose で起動）
	// 実装は環境に合わせて調整
	db := getTestDatabaseConnection(t)

	// AutoMigrate の前に、PostgreSQL カスタム ENUM 型を作成
	// （001_init.sql の 46 型 + 009 #201 薬量計算の 4 型）。
	// model.Medicine が calculation_type を持つため、本 setup を使う全テストの
	// medicines AutoMigrate に medicine_calculation_type が必須（欠落で CREATE TABLE 失敗）。
	// DROP TYPE IF EXISTS → CREATE TYPE の順序で、既存型を削除してから再作成
	enumTypes := []struct {
		name   string
		create string
	}{
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
		{"item_source", "CREATE TYPE item_source AS ENUM ('medical_record', 'manual', 'hospitalization')"},
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
	}
	for _, et := range enumTypes {
		query := fmt.Sprintf(`
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = '%s') THEN
        %s;
    END IF;
END
$$;`, et.name, et.create)
		if err := db.Exec(query).Error; err != nil {
			t.Fatalf("failed to create ENUM type %s: %v", et.name, err)
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
		ensureTestDatabaseExists(t, mainDB, testDBName)
		// 接続リーク防止: mainDB は CREATE DATABASE 専用。閉じないと setupTestDB 呼び出し毎に
		// 1 接続が漏れ、full suite で PostgreSQL の max_connections を使い切る
		// （FATAL: sorry, too many clients already / SQLSTATE 53300）。
		if sqlMainDB, derr := mainDB.DB(); derr == nil {
			_ = sqlMainDB.Close()
		}
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
	// 接続枯渇防止: setupTestDB は多数のテストから呼ばれる。プールを上限し、テスト終了時に
	// 必ず閉じることで接続の累積を防ぐ（閉じないと full suite で max_connections を超え 53300）。
	if sqlDB, derr := db.DB(); derr == nil {
		sqlDB.SetMaxOpenConns(10)
		sqlDB.SetMaxIdleConns(2)
		t.Cleanup(func() { _ = sqlDB.Close() })
	}
	return db
}

func ensureTestDatabaseExists(t *testing.T, mainDB *gorm.DB, testDBName string) {
	t.Helper()

	lockKey := "setup_test_database:" + testDBName
	if err := mainDB.Exec("SELECT pg_advisory_lock(hashtext(?))", lockKey).Error; err != nil {
		t.Fatalf("failed to acquire test database creation lock: %v", err)
	}
	defer func() {
		if err := mainDB.Exec("SELECT pg_advisory_unlock(hashtext(?))", lockKey).Error; err != nil {
			t.Logf("warning: failed to release test database creation lock: %v", err)
		}
	}()

	var exists bool
	if err := mainDB.Raw("SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = ?)", testDBName).Scan(&exists).Error; err != nil {
		t.Fatalf("failed to check test database existence: %v", err)
	}
	if exists {
		return
	}

	if err := mainDB.Exec("CREATE DATABASE " + quotePostgresIdentifier(testDBName)).Error; err != nil {
		if isDuplicateDatabaseError(err) {
			return
		}
		t.Fatalf("failed to create test database %s: %v", testDBName, err)
	}
}

func quotePostgresIdentifier(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}

func isDuplicateDatabaseError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "42P04"
}
