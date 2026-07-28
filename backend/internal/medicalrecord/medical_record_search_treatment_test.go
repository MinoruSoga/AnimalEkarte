package medicalrecord

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func setupMedicalRecordTreatmentSearchTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupMedicalRecordListTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.Consultation{},
		&model.Procedure{},
		&model.Medicine{},
		&model.InventoryItem{},
	))
	require.NoError(t, db.Exec(
		"TRUNCATE TABLE treatments, consultations, procedures, medicines, inventory_items CASCADE",
	).Error)
	testdb.SeedClinicsForFK(t, db, 1, 2)
	return db
}

func makeTreatmentSearchRecord(t *testing.T, db *gorm.DB, clinicID uint64, recordNo string, date time.Time) *model.MedicalRecord {
	t.Helper()
	return makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicID,
		RecordNo: recordNo,
		Date:     date,
	})
}

func makeTreatmentSearchTreatment(t *testing.T, db *gorm.DB, treatment *model.Treatment) *model.Treatment {
	t.Helper()
	require.NoError(t, db.WithContext(context.Background()).Create(treatment).Error)
	return treatment
}

func assertTreatmentSearchResult(
	t *testing.T,
	repo MedicalRecordRepository,
	clinicIDs []uint64,
	search string,
	wantRecordID uint64,
) {
	t.Helper()
	got, total, err := repo.FindAll(
		context.Background(),
		clinicIDs,
		MedicalRecordListFilters{Search: search},
		1,
		100,
	)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, got, 1)
	assert.Equal(t, wantRecordID, got[0].ID)
}

func assertTreatmentSearchEmpty(t *testing.T, repo MedicalRecordRepository, clinicIDs []uint64, search string) {
	t.Helper()
	got, total, err := repo.FindAll(
		context.Background(),
		clinicIDs,
		MedicalRecordListFilters{Search: search},
		1,
		100,
	)
	require.NoError(t, err)
	assert.Zero(t, total)
	assert.Empty(t, got)
}

func TestMedicalRecordRepository_FindAll_TreatmentSearch(t *testing.T) {
	db := setupMedicalRecordTreatmentSearchTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	baseDate := time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC)

	targetRecord := makeTreatmentSearchRecord(t, db, clinicA, "TS-TARGET", baseDate)
	makeTreatmentSearchTreatment(t, db, &model.Treatment{
		MedicalRecordID: targetRecord.ID,
		ItemType:        model.TreatmentItemTypeOther,
		Content:         "content-needle-237",
	})
	makeTreatmentSearchTreatment(t, db, &model.Treatment{
		MedicalRecordID: targetRecord.ID,
		ItemType:        model.TreatmentItemTypeOther,
		Memo:            "memo-needle-237",
	})
	targetProcedure := &model.Procedure{ClinicID: clinicA, Name: "procedure-needle-237"}
	targetMedicine := &model.Medicine{ClinicID: clinicA, Name: "medicine-needle-237"}
	targetConsultation := &model.Consultation{ClinicID: clinicA, Name: "consultation-needle-237"}
	targetInventory := &model.InventoryItem{
		ClinicID: clinicA,
		Name:     "inventory-needle-237",
		Category: model.InventoryCategoryOther,
	}
	for _, master := range []any{targetProcedure, targetMedicine, targetConsultation, targetInventory} {
		require.NoError(t, db.WithContext(ctx).Create(master).Error)
	}
	makeTreatmentSearchTreatment(t, db, &model.Treatment{
		MedicalRecordID: targetRecord.ID,
		ItemType:        model.TreatmentItemTypeProcedure,
		ProcedureID:     &targetProcedure.ID,
	})
	makeTreatmentSearchTreatment(t, db, &model.Treatment{
		MedicalRecordID: targetRecord.ID,
		ItemType:        model.TreatmentItemTypeMedicine,
		MedicineID:      &targetMedicine.ID,
	})
	makeTreatmentSearchTreatment(t, db, &model.Treatment{
		MedicalRecordID: targetRecord.ID,
		ItemType:        model.TreatmentItemTypeConsultation,
		ConsultationID:  &targetConsultation.ID,
	})
	makeTreatmentSearchTreatment(t, db, &model.Treatment{
		MedicalRecordID: targetRecord.ID,
		ItemType:        model.TreatmentItemTypeOther,
		InventoryID:     &targetInventory.ID,
	})

	kanaRecord := makeTreatmentSearchRecord(t, db, clinicA, "TS-KANA", baseDate.Add(time.Hour))
	kanaMedicine := &model.Medicine{ClinicID: clinicA, Name: "アモキシシリン"}
	require.NoError(t, db.WithContext(ctx).Create(kanaMedicine).Error)
	makeTreatmentSearchTreatment(t, db, &model.Treatment{
		MedicalRecordID: kanaRecord.ID,
		ItemType:        model.TreatmentItemTypeMedicine,
		MedicineID:      &kanaMedicine.ID,
	})

	duplicateRecord := makeTreatmentSearchRecord(t, db, clinicA, "TS-DUPLICATE", baseDate.Add(2*time.Hour))
	for range 2 {
		makeTreatmentSearchTreatment(t, db, &model.Treatment{
			MedicalRecordID: duplicateRecord.ID,
			ItemType:        model.TreatmentItemTypeOther,
			Content:         "duplicate-needle-237",
		})
	}

	ownerA := makeTestOwner(t, db, clinicA, "汚染テスト自院飼主")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "汚染テスト自院ペット")
	pollutedRecord := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA,
		RecordNo: "TS-POLLUTED",
		Date:     baseDate.Add(3 * time.Hour),
		OwnerID:  &ownerA.ID,
		PetID:    &petA.ID,
	})
	pollutedProcedure := &model.Procedure{ClinicID: clinicB, Name: "foreign-procedure-needle-237"}
	pollutedMedicine := &model.Medicine{ClinicID: clinicB, Name: "foreign-medicine-needle-237"}
	pollutedConsultation := &model.Consultation{ClinicID: clinicB, Name: "foreign-consultation-needle-237"}
	pollutedInventory := &model.InventoryItem{
		ClinicID: clinicB,
		Name:     "foreign-inventory-needle-237",
		Category: model.InventoryCategoryOther,
	}
	for _, master := range []any{pollutedProcedure, pollutedMedicine, pollutedConsultation, pollutedInventory} {
		require.NoError(t, db.WithContext(ctx).Create(master).Error)
	}
	pollutedTreatments := []model.Treatment{
		{
			MedicalRecordID: pollutedRecord.ID,
			ItemType:        model.TreatmentItemTypeProcedure,
			ProcedureID:     &pollutedProcedure.ID,
			Content:         "polluted-procedure-content-237",
			Memo:            "polluted-procedure-memo-237",
		},
		{
			MedicalRecordID: pollutedRecord.ID,
			ItemType:        model.TreatmentItemTypeMedicine,
			MedicineID:      &pollutedMedicine.ID,
			Content:         "polluted-medicine-content-237",
			Memo:            "polluted-medicine-memo-237",
		},
		{
			MedicalRecordID: pollutedRecord.ID,
			ItemType:        model.TreatmentItemTypeConsultation,
			ConsultationID:  &pollutedConsultation.ID,
			Content:         "polluted-consultation-content-237",
			Memo:            "polluted-consultation-memo-237",
		},
		{
			MedicalRecordID: pollutedRecord.ID,
			ItemType:        model.TreatmentItemTypeOther,
			InventoryID:     &pollutedInventory.ID,
			Content:         "polluted-inventory-content-237",
			Memo:            "polluted-inventory-memo-237",
		},
	}
	for i := range pollutedTreatments {
		makeTreatmentSearchTreatment(t, db, &pollutedTreatments[i])
	}

	deletedMasterRecord := makeTreatmentSearchRecord(t, db, clinicA, "TS-DELETED-MASTERS", baseDate.Add(4*time.Hour))
	deletedProcedure := &model.Procedure{ClinicID: clinicA, Name: "deleted-procedure-needle-237"}
	deletedMedicine := &model.Medicine{ClinicID: clinicA, Name: "deleted-medicine-needle-237"}
	deletedConsultation := &model.Consultation{ClinicID: clinicA, Name: "deleted-consultation-needle-237"}
	deletedInventory := &model.InventoryItem{
		ClinicID: clinicA,
		Name:     "deleted-inventory-needle-237",
		Category: model.InventoryCategoryOther,
	}
	for _, master := range []any{deletedProcedure, deletedMedicine, deletedConsultation, deletedInventory} {
		require.NoError(t, db.WithContext(ctx).Create(master).Error)
	}
	deletedMasterTreatments := []model.Treatment{
		{MedicalRecordID: deletedMasterRecord.ID, ItemType: model.TreatmentItemTypeProcedure, ProcedureID: &deletedProcedure.ID},
		{MedicalRecordID: deletedMasterRecord.ID, ItemType: model.TreatmentItemTypeMedicine, MedicineID: &deletedMedicine.ID},
		{MedicalRecordID: deletedMasterRecord.ID, ItemType: model.TreatmentItemTypeConsultation, ConsultationID: &deletedConsultation.ID},
		{MedicalRecordID: deletedMasterRecord.ID, ItemType: model.TreatmentItemTypeOther, InventoryID: &deletedInventory.ID},
	}
	for i := range deletedMasterTreatments {
		makeTreatmentSearchTreatment(t, db, &deletedMasterTreatments[i])
	}
	for _, master := range []any{deletedProcedure, deletedMedicine, deletedConsultation, deletedInventory} {
		require.NoError(t, db.WithContext(ctx).Delete(master).Error)
	}

	deletedTreatmentRecord := makeTreatmentSearchRecord(t, db, clinicA, "TS-DELETED-TREATMENT", baseDate.Add(5*time.Hour))
	deletedTreatmentMedicine := &model.Medicine{ClinicID: clinicA, Name: "deleted-treatment-master-237"}
	require.NoError(t, db.WithContext(ctx).Create(deletedTreatmentMedicine).Error)
	deletedTreatment := makeTreatmentSearchTreatment(t, db, &model.Treatment{
		MedicalRecordID: deletedTreatmentRecord.ID,
		ItemType:        model.TreatmentItemTypeMedicine,
		MedicineID:      &deletedTreatmentMedicine.ID,
		Content:         "deleted-treatment-content-237",
		Memo:            "deleted-treatment-memo-237",
	})
	require.NoError(t, db.WithContext(ctx).Delete(deletedTreatment).Error)

	foreignRecord := makeTreatmentSearchRecord(t, db, clinicB, "TS-FOREIGN-RECORD", baseDate.Add(6*time.Hour))
	makeTreatmentSearchTreatment(t, db, &model.Treatment{
		MedicalRecordID: foreignRecord.ID,
		ItemType:        model.TreatmentItemTypeOther,
		Content:         "foreign-record-content-237",
	})

	type wildcardFixture struct {
		name         string
		search       string
		literalValue string
		decoyValue   string
		record       *model.MedicalRecord
	}
	wildcardFixtures := []wildcardFixture{
		{
			name:         "percent",
			search:       "pct%needle237",
			literalValue: "pct%needle237",
			decoyValue:   "pctXneedle237",
		},
		{
			name:         "underscore",
			search:       "under_score_237",
			literalValue: "under_score_237",
			decoyValue:   "underXscoreX237",
		},
		{
			name:         "backslash",
			search:       `back\slash237`,
			literalValue: `back\slash237`,
			decoyValue:   "backXslash237",
		},
	}
	for i := range wildcardFixtures {
		literalRecord := makeTreatmentSearchRecord(
			t,
			db,
			clinicA,
			"TS-WILDCARD-LITERAL-"+wildcardFixtures[i].name,
			baseDate.Add(time.Duration(7+i)*time.Hour),
		)
		makeTreatmentSearchTreatment(t, db, &model.Treatment{
			MedicalRecordID: literalRecord.ID,
			ItemType:        model.TreatmentItemTypeOther,
			Content:         wildcardFixtures[i].literalValue,
		})
		decoyRecord := makeTreatmentSearchRecord(
			t,
			db,
			clinicA,
			"TS-WILDCARD-DECOY-"+wildcardFixtures[i].name,
			baseDate.Add(time.Duration(10+i)*time.Hour),
		)
		makeTreatmentSearchTreatment(t, db, &model.Treatment{
			MedicalRecordID: decoyRecord.ID,
			ItemType:        model.TreatmentItemTypeOther,
			Content:         wildcardFixtures[i].decoyValue,
		})
		wildcardFixtures[i].record = literalRecord
	}

	pagedRecords := make([]*model.MedicalRecord, 3)
	for i := range pagedRecords {
		pagedRecords[i] = makeTreatmentSearchRecord(
			t,
			db,
			clinicA,
			"TS-PAGE-"+string(rune('A'+i)),
			baseDate.Add(time.Duration(20+i)*time.Hour),
		)
		makeTreatmentSearchTreatment(t, db, &model.Treatment{
			MedicalRecordID: pagedRecords[i].ID,
			ItemType:        model.TreatmentItemTypeOther,
			Content:         "page-needle-237",
		})
	}

	positiveCases := []struct {
		name   string
		search string
	}{
		{name: "治療内容(content)で検索できる", search: "content-needle-237"},
		{name: "治療メモ(memo)で検索できる", search: "memo-needle-237"},
		{name: "処置名で検索できる", search: "procedure-needle-237"},
		{name: "薬剤名で検索できる", search: "medicine-needle-237"},
		{name: "診察名で検索できる", search: "consultation-needle-237"},
		{name: "在庫品名で検索できる", search: "inventory-needle-237"},
	}
	for _, tt := range positiveCases {
		t.Run(tt.name, func(t *testing.T) {
			assertTreatmentSearchResult(t, repo, []uint64{clinicA}, tt.search, targetRecord.ID)
		})
	}

	t.Run("カタカナのマスタ名はカタカナ入力でもひらがな入力でもヒットする", func(t *testing.T) {
		for _, search := range []string{"アモキシ", "あもきし"} {
			assertTreatmentSearchResult(t, repo, []uint64{clinicA}, search, kanaRecord.ID)
		}
	})

	t.Run("一致treatmentが複数でも行が重複しない", func(t *testing.T) {
		assertTreatmentSearchResult(t, repo, []uint64{clinicA}, "duplicate-needle-237", duplicateRecord.ID)
	})

	t.Run("汚染マスタFKの外部clinic名では検索ヒットしない", func(t *testing.T) {
		masterSearches := []string{
			pollutedProcedure.Name,
			pollutedMedicine.Name,
			pollutedConsultation.Name,
			pollutedInventory.Name,
		}
		for _, search := range masterSearches {
			assertTreatmentSearchEmpty(t, repo, []uint64{clinicA, clinicB}, search)
		}
		for i := range pollutedTreatments {
			treatment := &pollutedTreatments[i]
			assertTreatmentSearchResult(t, repo, []uint64{clinicA, clinicB}, treatment.Content, pollutedRecord.ID)
			assertTreatmentSearchResult(t, repo, []uint64{clinicA, clinicB}, treatment.Memo, pollutedRecord.ID)
		}
	})

	t.Run("論理削除済みマスタ名では検索ヒットしない", func(t *testing.T) {
		for _, search := range []string{
			deletedProcedure.Name,
			deletedMedicine.Name,
			deletedConsultation.Name,
			deletedInventory.Name,
		} {
			assertTreatmentSearchEmpty(t, repo, []uint64{clinicA}, search)
		}
	})

	t.Run("論理削除済みtreatmentは検索対象外", func(t *testing.T) {
		for _, search := range []string{
			deletedTreatment.Content,
			deletedTreatment.Memo,
			deletedTreatmentMedicine.Name,
		} {
			assertTreatmentSearchEmpty(t, repo, []uint64{clinicA}, search)
		}
	})

	t.Run("別clinicのカルテは検索対象外", func(t *testing.T) {
		assertTreatmentSearchEmpty(t, repo, []uint64{clinicA}, "foreign-record-content-237")
	})

	t.Run("LIKEワイルドカードはエスケープされる", func(t *testing.T) {
		for _, fixture := range wildcardFixtures {
			t.Run(fixture.name, func(t *testing.T) {
				assertTreatmentSearchResult(t, repo, []uint64{clinicA}, fixture.search, fixture.record.ID)
			})
		}
	})

	t.Run("空文字と長大な検索語の境界", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{Search: ""}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(15), total)
		assert.Len(t, got, 15)

		assertTreatmentSearchEmpty(t, repo, []uint64{clinicA}, strings.Repeat("長", 1000))
	})

	t.Run("countとページングが結果と一致する", func(t *testing.T) {
		firstPage, firstTotal, err := repo.FindAll(
			ctx,
			[]uint64{clinicA},
			MedicalRecordListFilters{Search: "page-needle-237"},
			1,
			2,
		)
		require.NoError(t, err)
		secondPage, secondTotal, err := repo.FindAll(
			ctx,
			[]uint64{clinicA},
			MedicalRecordListFilters{Search: "page-needle-237"},
			2,
			2,
		)
		require.NoError(t, err)

		assert.Equal(t, int64(3), firstTotal)
		assert.Equal(t, firstTotal, secondTotal)
		require.Len(t, firstPage, 2)
		require.Len(t, secondPage, 1)

		gotIDs := map[uint64]struct{}{}
		for _, record := range append(firstPage, secondPage...) {
			gotIDs[record.ID] = struct{}{}
		}
		assert.Len(t, gotIDs, 3)
		for _, record := range pagedRecords {
			_, found := gotIDs[record.ID]
			assert.True(t, found)
		}
	})
}
