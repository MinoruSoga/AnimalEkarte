package lstep

// repository_test.go — Repository（BE-004 / ISSUE-005 / ISSUE-009）の統合テスト。
//
// 対象: FindCheckupSyncPreview
// 検証観点: 集計値（生存ペット頭数/氏名CSV/来院数）、clinic_id 隔離、種別フィルタ、
//           来院日レンジ、慢性疾患、累計診療費、年間来院回数、年齢レンジ、最終健診日。
//
// makeBilling/makeCheckupTypeMaster はフラット package の同名ヘルパーの複製
// （BE8-4: import cycle を避けるための最小限の重複、移動時の型リネームはしない方針の対象外）。
// setupTestDB/ensureAutoMigrated/makeTestOwner はフラット側と同様 testdb.* に直接委譲する。

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

// setupCheckupSyncTestDB は owners/pets/medical_records/billings/checkups/pet_chronic_conditions を整備する。
func setupCheckupSyncTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.AnimalSpecies{}, &model.Pet{}, &model.CheckupType{}, &model.Checkup{}, &model.PetChronicCondition{},
	))
	db.Exec("TRUNCATE TABLE checkups CASCADE")
	db.Exec("TRUNCATE TABLE checkup_types CASCADE")
	db.Exec("TRUNCATE TABLE pet_chronic_conditions CASCADE")
	db.Exec("TRUNCATE TABLE billings CASCADE")
	db.Exec("TRUNCATE TABLE medical_records CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	return db
}

// makeBilling はテスト用の Billing を作成する（accounting_repository_unpaid_test.go の同名ヘルパーの複製）。
func makeBilling(t *testing.T, db *gorm.DB, clinicID uint64, ownerID, petID *uint64, amount int64, status model.BillingStatus, scheduledDate time.Time) {
	t.Helper()
	b := &model.Billing{
		ClinicID:      clinicID,
		OwnerID:       ownerID,
		PetID:         petID,
		TotalAmount:   amount,
		Status:        status,
		ScheduledDate: scheduledDate,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(b).Error)
}

// makeCheckupTypeMaster はテスト用の CheckupType を作成して返す
// （master_preload_clinic_isolation_test.go の同名ヘルパーの複製）。
func makeCheckupTypeMaster(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.CheckupType {
	t.Helper()
	ct := &model.CheckupType{ClinicID: clinicID, Name: name}
	require.NoError(t, db.WithContext(context.Background()).Create(ct).Error)
	return ct
}

func makeSyncSpecies(t *testing.T, db *gorm.DB, name string) *model.AnimalSpecies {
	t.Helper()
	sp := &model.AnimalSpecies{Name: name}
	require.NoError(t, db.WithContext(context.Background()).Create(sp).Error)
	return sp
}

func makeSyncPet(t *testing.T, db *gorm.DB, clinicID, ownerID, speciesID uint64, name string, birthDate, deceasedAt *time.Time) *model.Pet {
	t.Helper()
	pet := &model.Pet{
		ClinicID:        clinicID,
		OwnerID:         ownerID,
		AnimalSpeciesID: speciesID,
		Name:            name,
		BirthDate:       birthDate,
		DeceasedAt:      deceasedAt,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(pet).Error)
	return pet
}

func makeSyncMedicalRecord(t *testing.T, db *gorm.DB, clinicID, ownerID, petID uint64, date time.Time) *model.MedicalRecord {
	t.Helper()
	oid := ownerID
	pid := petID
	mr := &model.MedicalRecord{
		ClinicID: clinicID,
		RecordNo: fmt.Sprintf("MR-SYNC-%d-%d", ownerID, date.UnixNano()),
		Date:     date,
		OwnerID:  &oid,
		PetID:    &pid,
		Status:   model.MedicalRecordStatusFinalized,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(mr).Error)
	return mr
}

func makeSyncCheckup(t *testing.T, db *gorm.DB, clinicID, mrID, petID, checkupTypeID uint64, date time.Time) uint64 {
	t.Helper()
	pid := petID
	c := &model.Checkup{
		ClinicID:        clinicID,
		MedicalRecordID: mrID,
		PetID:           &pid,
		CheckupTypeID:   checkupTypeID,
		Date:            date,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(c).Error)
	return c.ID
}

func TestCheckupSyncRepository_FindCheckupSyncPreview_NilParamsReturnsInvalidInput(t *testing.T) {
	db := setupCheckupSyncTestDB(t)
	repo := NewCheckupSyncRepository(db)
	ctx := context.Background()

	_, err := repo.FindCheckupSyncPreview(ctx, nil)
	require.Error(t, err)
}

func TestCheckupSyncRepository_FindCheckupSyncPreview_CurrentOwnerAfterTransfer(t *testing.T) {
	db := setupCheckupSyncTestDB(t)
	repo := NewCheckupSyncRepository(db)
	ctx := context.Background()
	const clinicID = uint64(70107)

	previousOwner := testdb.MakeTestOwner(t, db, clinicID, "同期譲渡前飼主")
	currentOwner := testdb.MakeTestOwner(t, db, clinicID, "同期譲渡後飼主")
	species := makeSyncSpecies(t, db, "同期譲渡動物種")
	pet := makeSyncPet(t, db, clinicID, previousOwner.ID, species.ID, "同期譲渡ペット", nil, nil)
	visitDate := time.Now().AddDate(0, 0, -10)
	checkupDate := time.Now().AddDate(0, 0, -5)
	record := makeSyncMedicalRecord(t, db, clinicID, previousOwner.ID, pet.ID, visitDate)
	checkupType := makeCheckupTypeMaster(t, db, clinicID, "同期譲渡健診")
	makeSyncCheckup(t, db, clinicID, record.ID, pet.ID, checkupType.ID, checkupDate)
	require.NoError(t, db.WithContext(ctx).Model(pet).Update("owner_id", currentOwner.ID).Error)

	rows, err := repo.FindCheckupSyncPreview(ctx, &FindCheckupSyncPreviewParams{ClinicID: clinicID})
	require.NoError(t, err)
	byOwner := make(map[uint64]CheckupSyncPreviewRow, len(rows))
	for _, row := range rows {
		byOwner[row.OwnerID] = row
	}

	current := byOwner[currentOwner.ID]
	assert.Equal(t, int64(1), current.TotalVisitCount)
	require.NotNil(t, current.LastVisitDate)
	require.NotNil(t, current.LastCheckupDate)
	previous := byOwner[previousOwner.ID]
	assert.Equal(t, int64(0), previous.TotalVisitCount)
	assert.Nil(t, previous.LastVisitDate)
	assert.Nil(t, previous.LastCheckupDate)
}

func TestCheckupSyncRepository_FindCheckupSyncPreview_AggregationClinicIsolationAndSpecies(t *testing.T) {
	db := setupCheckupSyncTestDB(t)
	repo := NewCheckupSyncRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	dogSpecies := makeSyncSpecies(t, db, "犬")
	catSpecies := makeSyncSpecies(t, db, "猫")

	ownerA := testdb.MakeTestOwner(t, db, clinicA, "同期飼主A")
	deceasedAt := time.Now()
	livingDog := makeSyncPet(t, db, clinicA, ownerA.ID, dogSpecies.ID, "同期ポチ", nil, nil)
	_ = makeSyncPet(t, db, clinicA, ownerA.ID, catSpecies.ID, "同期タマ", nil, &deceasedAt) // 死亡ペット

	makeSyncMedicalRecord(t, db, clinicA, ownerA.ID, livingDog.ID, time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC))
	makeSyncMedicalRecord(t, db, clinicA, ownerA.ID, livingDog.ID, time.Date(2026, 2, 5, 0, 0, 0, 0, time.UTC))

	// 別クリニックの飼主（混入しないこと）
	ownerB := testdb.MakeTestOwner(t, db, clinicB, "同期飼主B")
	petB := makeSyncPet(t, db, clinicB, ownerB.ID, dogSpecies.ID, "同期ポチB", nil, nil)
	makeSyncMedicalRecord(t, db, clinicB, ownerB.ID, petB.ID, time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC))

	rows, err := repo.FindCheckupSyncPreview(ctx, &FindCheckupSyncPreviewParams{ClinicID: clinicA})
	require.NoError(t, err)
	require.Len(t, rows, 1, "clinic A の飼主のみ返る")
	row := rows[0]
	assert.Equal(t, ownerA.ID, row.OwnerID)
	assert.Equal(t, int64(1), row.LivingPetCount, "死亡ペットは生存頭数に含まれない")
	assert.Equal(t, "同期ポチ", row.PetNamesCSV, "死亡ペット名はCSVに含まれない")
	assert.Equal(t, int64(2), row.TotalVisitCount)

	t.Run("種別フィルタは生存ペットのみで判定される", func(t *testing.T) {
		rows, err := repo.FindCheckupSyncPreview(ctx, &FindCheckupSyncPreviewParams{ClinicID: clinicA, Species: "イヌ"})
		require.NoError(t, err)
		assert.Len(t, rows, 0, "犬に一致しない種別名では該当なし")

		rows, err = repo.FindCheckupSyncPreview(ctx, &FindCheckupSyncPreviewParams{ClinicID: clinicA, Species: "犬"})
		require.NoError(t, err)
		require.Len(t, rows, 1)

		rows, err = repo.FindCheckupSyncPreview(ctx, &FindCheckupSyncPreviewParams{ClinicID: clinicA, Species: "猫"})
		require.NoError(t, err)
		assert.Len(t, rows, 0, "唯一の猫は死亡しているため対象外")
	})
}

func TestCheckupSyncRepository_FindCheckupSyncPreview_LastVisitDateRangeFilters(t *testing.T) {
	db := setupCheckupSyncTestDB(t)
	repo := NewCheckupSyncRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	dogSpecies := makeSyncSpecies(t, db, "犬")
	ownerEarly := testdb.MakeTestOwner(t, db, clinicA, "早期飼主")
	petEarly := makeSyncPet(t, db, clinicA, ownerEarly.ID, dogSpecies.ID, "早期ポチ", nil, nil)
	makeSyncMedicalRecord(t, db, clinicA, ownerEarly.ID, petEarly.ID, time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC))

	ownerRecent := testdb.MakeTestOwner(t, db, clinicA, "最近飼主")
	petRecent := makeSyncPet(t, db, clinicA, ownerRecent.ID, dogSpecies.ID, "最近ポチ", nil, nil)
	makeSyncMedicalRecord(t, db, clinicA, ownerRecent.ID, petRecent.ID, time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC))

	before := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	rows, err := repo.FindCheckupSyncPreview(ctx, &FindCheckupSyncPreviewParams{ClinicID: clinicA, LastVisitBefore: &before})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, ownerEarly.ID, rows[0].OwnerID)

	after := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	rows, err = repo.FindCheckupSyncPreview(ctx, &FindCheckupSyncPreviewParams{ClinicID: clinicA, LastVisitAfter: &after})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, ownerRecent.ID, rows[0].OwnerID)
}

func TestCheckupSyncRepository_FindCheckupSyncPreview_ChronicConditionFilter(t *testing.T) {
	db := setupCheckupSyncTestDB(t)
	repo := NewCheckupSyncRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	dogSpecies := makeSyncSpecies(t, db, "犬")
	ownerSick := testdb.MakeTestOwner(t, db, clinicA, "慢性疾患飼主")
	petSick := makeSyncPet(t, db, clinicA, ownerSick.ID, dogSpecies.ID, "疾患ポチ", nil, nil)
	makeSyncMedicalRecord(t, db, clinicA, ownerSick.ID, petSick.ID, time.Now())
	require.NoError(t, db.WithContext(ctx).Create(&model.PetChronicCondition{
		ClinicID: clinicA, PetID: petSick.ID, ConditionCode: "DM", ConditionName: "糖尿病",
		DiagnosedAt: time.Now(), IsActive: true,
	}).Error)

	ownerHealthy := testdb.MakeTestOwner(t, db, clinicA, "健康飼主")
	petHealthy := makeSyncPet(t, db, clinicA, ownerHealthy.ID, dogSpecies.ID, "健康ポチ", nil, nil)
	makeSyncMedicalRecord(t, db, clinicA, ownerHealthy.ID, petHealthy.ID, time.Now())

	trueFlag := true
	rows, err := repo.FindCheckupSyncPreview(ctx, &FindCheckupSyncPreviewParams{ClinicID: clinicA, HasChronicCondition: &trueFlag})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, ownerSick.ID, rows[0].OwnerID)
	assert.True(t, rows[0].HasChronicCondition)

	falseFlag := false
	rows, err = repo.FindCheckupSyncPreview(ctx, &FindCheckupSyncPreviewParams{ClinicID: clinicA, HasChronicCondition: &falseFlag})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, ownerHealthy.ID, rows[0].OwnerID)
}

func TestCheckupSyncRepository_FindCheckupSyncPreview_TotalAmountAndAnnualVisitFilters(t *testing.T) {
	db := setupCheckupSyncTestDB(t)
	repo := NewCheckupSyncRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	dogSpecies := makeSyncSpecies(t, db, "犬")
	ownerHigh := testdb.MakeTestOwner(t, db, clinicA, "高額飼主")
	petHigh := makeSyncPet(t, db, clinicA, ownerHigh.ID, dogSpecies.ID, "高額ポチ", nil, nil)
	makeSyncMedicalRecord(t, db, clinicA, ownerHigh.ID, petHigh.ID, time.Now())
	makeBilling(t, db, clinicA, &ownerHigh.ID, nil, 10000, model.BillingStatusCompleted, time.Now())
	makeBilling(t, db, clinicA, &ownerHigh.ID, nil, 5000, model.BillingStatusWaiting, time.Now()) // 未完了は除外

	ownerLow := testdb.MakeTestOwner(t, db, clinicA, "低額飼主")
	petLow := makeSyncPet(t, db, clinicA, ownerLow.ID, dogSpecies.ID, "低額ポチ", nil, nil)
	makeSyncMedicalRecord(t, db, clinicA, ownerLow.ID, petLow.ID, time.Now())
	makeBilling(t, db, clinicA, &ownerLow.ID, nil, 100, model.BillingStatusCompleted, time.Now())

	minAmount := int64(5000)
	rows, err := repo.FindCheckupSyncPreview(ctx, &FindCheckupSyncPreviewParams{ClinicID: clinicA, MinTotalAmount: &minAmount})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, ownerHigh.ID, rows[0].OwnerID)
	assert.Equal(t, int64(10000), rows[0].TotalAmount, "completed のみ合算される")

	ownerFrequent := testdb.MakeTestOwner(t, db, clinicA, "頻回飼主")
	petFrequent := makeSyncPet(t, db, clinicA, ownerFrequent.ID, dogSpecies.ID, "頻回ポチ", nil, nil)
	makeSyncMedicalRecord(t, db, clinicA, ownerFrequent.ID, petFrequent.ID, time.Now())
	makeSyncMedicalRecord(t, db, clinicA, ownerFrequent.ID, petFrequent.ID, time.Now().AddDate(0, 0, -10))

	ownerRare := testdb.MakeTestOwner(t, db, clinicA, "希少飼主")
	petRare := makeSyncPet(t, db, clinicA, ownerRare.ID, dogSpecies.ID, "希少ポチ", nil, nil)
	makeSyncMedicalRecord(t, db, clinicA, ownerRare.ID, petRare.ID, time.Now().AddDate(-2, 0, 0)) // 2年前=年間対象外

	minVisits := int64(2)
	rows, err = repo.FindCheckupSyncPreview(ctx, &FindCheckupSyncPreviewParams{ClinicID: clinicA, MinAnnualVisitCount: &minVisits})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, ownerFrequent.ID, rows[0].OwnerID)
}

func TestCheckupSyncRepository_FindCheckupSyncPreview_AgeAndLastCheckupFilters(t *testing.T) {
	db := setupCheckupSyncTestDB(t)
	repo := NewCheckupSyncRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	dogSpecies := makeSyncSpecies(t, db, "犬")
	ct := makeCheckupTypeMaster(t, db, clinicA, "同期健診種別")

	youngBirth := time.Now().AddDate(-1, 0, 0)
	ownerYoung := testdb.MakeTestOwner(t, db, clinicA, "若齢飼主")
	petYoung := makeSyncPet(t, db, clinicA, ownerYoung.ID, dogSpecies.ID, "若齢ポチ", &youngBirth, nil)
	mrYoung := makeSyncMedicalRecord(t, db, clinicA, ownerYoung.ID, petYoung.ID, time.Now())
	makeSyncCheckup(t, db, clinicA, mrYoung.ID, petYoung.ID, ct.ID, time.Now().AddDate(0, -1, 0))

	oldBirth := time.Now().AddDate(-10, 0, 0)
	ownerOld := testdb.MakeTestOwner(t, db, clinicA, "高齢飼主")
	petOld := makeSyncPet(t, db, clinicA, ownerOld.ID, dogSpecies.ID, "高齢ポチ", &oldBirth, nil)
	mrOld := makeSyncMedicalRecord(t, db, clinicA, ownerOld.ID, petOld.ID, time.Now())
	makeSyncCheckup(t, db, clinicA, mrOld.ID, petOld.ID, ct.ID, time.Now().AddDate(-1, 0, 0))

	t.Run("MinAgeYears は高齢ペットのみヒット", func(t *testing.T) {
		minAge := 5
		rows, err := repo.FindCheckupSyncPreview(ctx, &FindCheckupSyncPreviewParams{ClinicID: clinicA, MinAgeYears: &minAge})
		require.NoError(t, err)
		require.Len(t, rows, 1)
		assert.Equal(t, ownerOld.ID, rows[0].OwnerID)
	})

	t.Run("MaxAgeYears は若齢ペットのみヒット", func(t *testing.T) {
		maxAge := 3
		rows, err := repo.FindCheckupSyncPreview(ctx, &FindCheckupSyncPreviewParams{ClinicID: clinicA, MaxAgeYears: &maxAge})
		require.NoError(t, err)
		require.Len(t, rows, 1)
		assert.Equal(t, ownerYoung.ID, rows[0].OwnerID)
	})

	t.Run("LastCheckupBefore は健診実施が古い飼主のみヒット", func(t *testing.T) {
		checkupBefore := time.Now().AddDate(0, -6, 0)
		rows, err := repo.FindCheckupSyncPreview(ctx, &FindCheckupSyncPreviewParams{ClinicID: clinicA, LastCheckupBefore: &checkupBefore})
		require.NoError(t, err)
		require.Len(t, rows, 1)
		assert.Equal(t, ownerOld.ID, rows[0].OwnerID)
	})

	t.Run("LastCheckupAfter は健診実施が新しい飼主のみヒット", func(t *testing.T) {
		checkupAfter := time.Now().AddDate(0, -6, 0)
		rows, err := repo.FindCheckupSyncPreview(ctx, &FindCheckupSyncPreviewParams{ClinicID: clinicA, LastCheckupAfter: &checkupAfter})
		require.NoError(t, err)
		require.Len(t, rows, 1)
		assert.Equal(t, ownerYoung.ID, rows[0].OwnerID)
	})
}

func TestCheckupSyncRepository_FindCheckupSyncPreview_RejectsCrossClinicChildRows(t *testing.T) {
	db := setupCheckupSyncTestDB(t)
	repo := NewCheckupSyncRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	dogSpecies := makeSyncSpecies(t, db, "犬")
	ownerA := testdb.MakeTestOwner(t, db, clinicA, "分離対象飼主")
	petA := makeSyncPet(t, db, clinicA, ownerA.ID, dogSpecies.ID, "分離対象ペット", nil, nil)
	makeSyncMedicalRecord(t, db, clinicA, ownerA.ID, petA.ID, time.Now().AddDate(0, 0, -7))

	// Deliberately inconsistent child rows model data that bypassed the normal
	// application write path. A tenant-scoped read must still fail closed.
	require.NoError(t, db.WithContext(ctx).Create(&model.PetChronicCondition{
		ClinicID: clinicB, PetID: petA.ID, ConditionCode: "CROSS", ConditionName: "別clinic疾患",
		DiagnosedAt: time.Now(), IsActive: true,
	}).Error)

	checkupTypeA := makeCheckupTypeMaster(t, db, clinicA, "分離健診")
	crossClinicRecord := makeSyncMedicalRecord(t, db, clinicB, ownerA.ID, petA.ID, time.Now())
	makeSyncCheckup(t, db, clinicA, crossClinicRecord.ID, petA.ID, checkupTypeA.ID, time.Now())

	rows, err := repo.FindCheckupSyncPreview(ctx, &FindCheckupSyncPreviewParams{ClinicID: clinicA})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.False(t, rows[0].HasChronicCondition)
	assert.Nil(t, rows[0].LastCheckupDate)

	trueFlag := true
	rows, err = repo.FindCheckupSyncPreview(ctx, &FindCheckupSyncPreviewParams{
		ClinicID: clinicA, HasChronicCondition: &trueFlag,
	})
	require.NoError(t, err)
	assert.Empty(t, rows)

	falseFlag := false
	rows, err = repo.FindCheckupSyncPreview(ctx, &FindCheckupSyncPreviewParams{
		ClinicID: clinicA, HasChronicCondition: &falseFlag,
	})
	require.NoError(t, err)
	require.Len(t, rows, 1)

	checkupBefore := time.Now().AddDate(0, 1, 0)
	rows, err = repo.FindCheckupSyncPreview(ctx, &FindCheckupSyncPreviewParams{
		ClinicID: clinicA, LastCheckupBefore: &checkupBefore,
	})
	require.NoError(t, err)
	assert.Empty(t, rows)

	checkupAfter := time.Now().AddDate(-1, 0, 0)
	rows, err = repo.FindCheckupSyncPreview(ctx, &FindCheckupSyncPreviewParams{
		ClinicID: clinicA, LastCheckupAfter: &checkupAfter,
	})
	require.NoError(t, err)
	assert.Empty(t, rows)
}

// BUG-030: CTE rewrite must keep hard row cap (formerly only LIMIT after full GROUP BY).
func TestCheckupSyncRepository_FindCheckupSyncPreview_RespectsRowLimit(t *testing.T) {
	db := setupCheckupSyncTestDB(t)
	repo := NewCheckupSyncRepository(db)
	ctx := context.Background()
	const clinicID = uint64(73030)

	dogSpecies := makeSyncSpecies(t, db, "制限テスト種")
	// Cap is 500; seed slightly over so the bound is observable.
	const seed = CheckupSyncPreviewRowLimit + 25
	for i := 0; i < seed; i++ {
		owner := testdb.MakeTestOwner(t, db, clinicID, fmt.Sprintf("制限飼主-%d", i))
		pet := makeSyncPet(t, db, clinicID, owner.ID, dogSpecies.ID, fmt.Sprintf("制限ペット-%d", i), nil, nil)
		makeSyncMedicalRecord(t, db, clinicID, owner.ID, pet.ID, time.Now().AddDate(0, 0, -i))
	}

	rows, err := repo.FindCheckupSyncPreview(ctx, &FindCheckupSyncPreviewParams{ClinicID: clinicID})
	require.NoError(t, err)
	assert.LessOrEqual(t, len(rows), CheckupSyncPreviewRowLimit)
	assert.Equal(t, CheckupSyncPreviewRowLimit, len(rows))
}

func TestBuildCheckupSyncPostJoinFilters_DateAndAmount(t *testing.T) {
	before := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	after := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	minAge := 3
	maxAge := 12
	minAmount := int64(1000)
	minVisits := int64(2)
	checkupBefore := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)
	checkupAfter := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	filters, args := buildCheckupSyncPostJoinFilters(&FindCheckupSyncPreviewParams{
		LastVisitBefore:     &before,
		LastVisitAfter:      &after,
		MinAgeYears:         &minAge,
		MaxAgeYears:         &maxAge,
		MinTotalAmount:      &minAmount,
		MinAnnualVisitCount: &minVisits,
		LastCheckupBefore:   &checkupBefore,
		LastCheckupAfter:    &checkupAfter,
	})
	require.Len(t, filters, 8)
	require.Len(t, args, 8)
	assert.Contains(t, filters[0], "va.last_visit_date <=")
	assert.Equal(t, "2026-03-01", args[0])
	// order: before, after, minAge, maxAge, minVisits, minAmount, checkupBefore, checkupAfter
	assert.Equal(t, minVisits, args[4])
	assert.Equal(t, minAmount, args[5])
}
