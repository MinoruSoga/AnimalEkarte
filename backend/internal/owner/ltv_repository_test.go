package owner

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func setupLTVTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.AnimalSpecies{}, &model.Pet{}))
	return db
}

type ltvTestRepository struct {
	t    *testing.T
	db   *gorm.DB
	repo LtvRepository
}

func newLTVTestRepository(t *testing.T, db *gorm.DB) LtvRepository {
	t.Helper()
	return &ltvTestRepository{t: t, db: db, repo: NewLtvRepository(db)}
}

func (r *ltvTestRepository) FindOwnerLTV(ctx context.Context, params *FindOwnerLTVParams) ([]OwnerLTVRow, error) {
	r.t.Helper()
	var records []model.MedicalRecord
	require.NoError(r.t, r.db.WithContext(ctx).
		Where("pet_id IS NULL AND owner_id IS NOT NULL").
		Find(&records).Error)
	if len(records) > 0 {
		species := &model.AnimalSpecies{Name: "LTV fixture species"}
		require.NoError(r.t, r.db.WithContext(ctx).Create(species).Error)
		for i := range records {
			var owner model.Owner
			err := r.db.WithContext(ctx).
				Where("id = ? AND clinic_id = ?", *records[i].OwnerID, records[i].ClinicID).
				First(&owner).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			require.NoError(r.t, err)
			pet := &model.Pet{
				ClinicID:        records[i].ClinicID,
				OwnerID:         owner.ID,
				AnimalSpeciesID: species.ID,
				Name:            fmt.Sprintf("LTV fixture pet %d", records[i].ID),
			}
			require.NoError(r.t, r.db.WithContext(ctx).Create(pet).Error)
			require.NoError(r.t, r.db.WithContext(ctx).
				Model(&records[i]).
				Update("pet_id", pet.ID).Error)
		}
	}
	return r.repo.FindOwnerLTV(ctx, params)
}

func TestFindOwnerLTV_NilPetRecordIsNotCurrentOwnerVisit(t *testing.T) {
	db := setupLTVTestDB(t)
	repo := NewLtvRepository(db)
	ctx := context.Background()
	const clinicID = uint64(70108)

	owner := testdb.MakeTestOwner(t, db, clinicID, "pet未設定LTV飼主")
	record := &model.MedicalRecord{
		ClinicID: clinicID,
		RecordNo: "MR-LTV-NIL-PET",
		Date:     time.Now(),
		OwnerID:  &owner.ID,
	}
	require.NoError(t, db.WithContext(ctx).Create(record).Error)

	rows, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
		ClinicID:       clinicID,
		IncludeZero:    true,
		IncludeNoVisit: true,
	})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, owner.ID, rows[0].OwnerID)
	assert.Zero(t, rows[0].TotalVisitCount)
	assert.Nil(t, rows[0].LastVisitDate)
}

func TestFindOwnerLTV_CurrentOwnerAfterTransfer(t *testing.T) {
	db := setupLTVTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.AnimalSpecies{}, &model.Pet{}))
	repo := newLTVTestRepository(t, db)
	ctx := context.Background()
	const clinicID = uint64(70106)

	previousOwner := testdb.MakeTestOwner(t, db, clinicID, "LTV譲渡前飼主")
	currentOwner := testdb.MakeTestOwner(t, db, clinicID, "LTV譲渡後飼主")
	species := &model.AnimalSpecies{Name: "LTV譲渡動物種"}
	require.NoError(t, db.WithContext(ctx).Create(species).Error)
	pet := &model.Pet{
		ClinicID:        clinicID,
		OwnerID:         previousOwner.ID,
		AnimalSpeciesID: species.ID,
		Name:            "LTV譲渡ペット",
	}
	require.NoError(t, db.WithContext(ctx).Create(pet).Error)
	record := &model.MedicalRecord{
		ClinicID: clinicID,
		RecordNo: "MR-LTV-CURRENT-OWNER",
		Date:     time.Now(),
		OwnerID:  &previousOwner.ID,
		PetID:    &pet.ID,
	}
	require.NoError(t, db.WithContext(ctx).Create(record).Error)
	require.NoError(t, db.WithContext(ctx).Model(pet).Update("owner_id", currentOwner.ID).Error)

	rows, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
		ClinicID:       clinicID,
		IncludeZero:    true,
		IncludeNoVisit: true,
	})
	require.NoError(t, err)

	byOwner := make(map[uint64]OwnerLTVRow, len(rows))
	for _, row := range rows {
		byOwner[row.OwnerID] = row
	}
	assert.Equal(t, int64(1), byOwner[currentOwner.ID].TotalVisitCount)
	assert.Equal(t, int64(0), byOwner[previousOwner.ID].TotalVisitCount)
}

func TestFindOwnerLTV_SearchEscapesLikeWildcards(t *testing.T) {
	db := setupLTVTestDB(t)
	repo := newLTVTestRepository(t, db)
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
	db := setupLTVTestDB(t)
	repo := newLTVTestRepository(t, db)
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
	db := setupLTVTestDB(t)
	repo := newLTVTestRepository(t, db)
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
	db := setupLTVTestDB(t)
	repo := newLTVTestRepository(t, db)
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
	db := setupLTVTestDB(t)
	repo := newLTVTestRepository(t, db)
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
	db := setupLTVTestDB(t)
	repo := newLTVTestRepository(t, db)
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
	db := setupLTVTestDB(t)
	repo := newLTVTestRepository(t, db)
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

	// Test 3: 売上タブ相当（year）でも include_zero=false なら除外
	year := time.Now().Year()
	result3, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
		ClinicID:    clinicID,
		Year:        &year,
		AmountBasis: "gross_total_amount",
		Sort:        "annual_amount",
		IncludeZero: false,
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, len(result3), "revenue tab params should still exclude zero when include_zero=false")
}

// TestShouldExcludeZeroAnnualAmount_BUG012 は 0 円除外の適用範囲（売上軸のみ）を表で固定する。
func TestShouldExcludeZeroAnnualAmount_BUG012(t *testing.T) {
	year := 2026
	minAmt := int64(1)
	minVisit := int64(1)

	tests := []struct {
		name   string
		params FindOwnerLTVParams
		want   bool
	}{
		{
			name:   "default LTV list excludes zero",
			params: FindOwnerLTVParams{},
			want:   true,
		},
		{
			name:   "include_zero true never excludes",
			params: FindOwnerLTVParams{IncludeZero: true, Year: &year, Sort: "annual_amount"},
			want:   false,
		},
		{
			name:   "revenue year+amount_basis excludes",
			params: FindOwnerLTVParams{Year: &year, AmountBasis: "gross_total_amount", Sort: "annual_amount"},
			want:   true,
		},
		{
			name:   "visit tab period_preset does not exclude",
			params: FindOwnerLTVParams{PeriodPreset: "last_12_months", Sort: "period_visit_count"},
			want:   false,
		},
		{
			name:   "last_visit bucket does not exclude",
			params: FindOwnerLTVParams{LastVisitBucket: "over_3m", Sort: "last_visit_date"},
			want:   false,
		},
		{
			name:   "amount range still excludes on visit sort",
			params: FindOwnerLTVParams{MinTotalAmount: &minAmt, Sort: "period_visit_count"},
			want:   true,
		},
		{
			name:   "min visit count alone does not exclude",
			params: FindOwnerLTVParams{MinVisitCount: &minVisit, Sort: "period_visit_count"},
			want:   false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, shouldExcludeZeroAnnualAmount(&tc.params))
		})
	}
}

// TestFindOwnerLTV_VisitAndLastVisitTabsKeepZeroAmountOwners_BUG012
// 来院回数・最終来院タブ相当の params では annual_amount=0 でも飼主が残る（BUG-012）。
func TestFindOwnerLTV_VisitAndLastVisitTabsKeepZeroAmountOwners_BUG012(t *testing.T) {
	db := setupLTVTestDB(t)
	repo := newLTVTestRepository(t, db)
	ctx := context.Background()
	clinicID := uint64(8012)

	owner := &model.Owner{
		ClinicID: clinicID,
		Name:     "Owner Visit Zero Yen BUG012",
	}
	require.NoError(t, db.WithContext(ctx).Create(owner).Error)

	// 120 日前の来院 → last_visit_bucket = over_3m、会計なし → annual_amount=0
	visitDate := time.Now().AddDate(0, 0, -120)
	mr := &model.MedicalRecord{
		ClinicID: clinicID,
		OwnerID:  &owner.ID,
		Date:     visitDate,
	}
	require.NoError(t, db.WithContext(ctx).Create(mr).Error)

	// 来院回数タブ既定（include_zero 未指定 = false）
	visitRows, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
		ClinicID:     clinicID,
		PeriodPreset: "last_12_months",
		Sort:         "period_visit_count",
		Order:        "desc",
		IncludeZero:  false,
	})
	require.NoError(t, err)
	require.Len(t, visitRows, 1, "visit tab must list owners with visits even when annual_amount=0")
	assert.Equal(t, owner.ID, visitRows[0].OwnerID)
	require.NotNil(t, visitRows[0].AnnualAmount)
	assert.Equal(t, int64(0), *visitRows[0].AnnualAmount)
	require.NotNil(t, visitRows[0].PeriodVisitCount)
	assert.Equal(t, int64(1), *visitRows[0].PeriodVisitCount)

	// 最終来院タブ既定
	lastRows, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
		ClinicID:        clinicID,
		LastVisitBucket: "over_3m",
		Sort:            "last_visit_date",
		Order:           "asc",
		IncludeZero:     false,
	})
	require.NoError(t, err)
	require.Len(t, lastRows, 1, "last_visit tab must list matching bucket even when annual_amount=0")
	assert.Equal(t, owner.ID, lastRows[0].OwnerID)

	// 売上タブ相当は依然として除外
	year := time.Now().Year()
	revenueRows, err := repo.FindOwnerLTV(ctx, &FindOwnerLTVParams{
		ClinicID:    clinicID,
		Year:        &year,
		AmountBasis: "gross_total_amount",
		Sort:        "annual_amount",
		IncludeZero: false,
	})
	require.NoError(t, err)
	assert.Equal(t, 0, len(revenueRows), "revenue ranking must still exclude zero-amount owners")
}

// TestFindOwnerLTV_AmountBasisSwitching
// ISSUE-001: amount_basis 切り替え（gross_total_amount / paid_amount / net_paid_amount）
func TestFindOwnerLTV_AmountBasisSwitching(t *testing.T) {
	db := setupLTVTestDB(t)
	repo := newLTVTestRepository(t, db)
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

	// Refund: 同一医院の 1000 + 500 円。複数返金でも請求・支払いを重複集計しない。
	refund := &model.BillingRefund{
		ClinicID:  clinicID,
		BillingID: billing.ID,
		Amount:    1000,
	}
	if err := db.WithContext(ctx).Create(refund).Error; err != nil {
		t.Fatalf("failed to create refund: %v", err)
	}
	secondRefund := &model.BillingRefund{
		ClinicID:  clinicID,
		BillingID: billing.ID,
		Amount:    500,
	}
	if err := db.WithContext(ctx).Create(secondRefund).Error; err != nil {
		t.Fatalf("failed to create second refund: %v", err)
	}
	// 壊れた外部参照があっても別医院の返金を医院1へ混入させない。
	crossClinicRefund := &model.BillingRefund{
		ClinicID:  2,
		BillingID: billing.ID,
		Amount:    99_999,
	}
	if err := db.WithContext(ctx).Create(crossClinicRefund).Error; err != nil {
		t.Fatalf("failed to create cross-clinic refund: %v", err)
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
	assert.Equal(t, int64(6500), *result3[0].AnnualAmount, "net_paid_amount should be 8000 - 1500 = 6500")

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
	require.Len(t, result4, 1)
	assert.Equal(t, int64(6500), *result4[0].AnnualAmount)

	minAmount = 7000
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
