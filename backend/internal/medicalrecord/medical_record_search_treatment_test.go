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

// TestMedicalRecordRepository_FindAll_TreatmentSearchExcluded は EMR-244 で
// 治療内容・治療メモ・治療マスタ名を検索する UNION 腕（旧 arm 5）をカルテ一覧の
// free 検索から外した契約を固定する。残る検索対象は
// 「カルテ番号 / 飼主名・カナ / ペット名・カナ / 主訴」の4条件のみ。
func TestMedicalRecordRepository_FindAll_TreatmentSearchExcluded(t *testing.T) {
	db := setupMedicalRecordTreatmentSearchTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	baseDate := time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC)

	// 4条件すべてに針を持つ対象カルテ。飼主カナ・ペットカナも個別に検証するため
	// name_kana を明示的に設定する（makeTestOwner/makeSpeciesAndPet はセットしない）。
	owner := makeTestOwner(t, db, clinicA, "四条件飼主")
	require.NoError(t, db.Model(owner).Update("name_kana", "ヨンジョウケン").Error)
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "四条件ペット")
	require.NoError(t, db.Model(pet).Update("name_kana", "シカイペット").Error)

	targetRecord := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA,
		RecordNo: "TS-TARGET-237",
		Date:     baseDate,
		OwnerID:  &owner.ID,
		PetID:    &pet.ID,
	})
	makeInquiryForRecord(t, db, targetRecord.ID, "四条件主訴：食欲不振")

	// 旧 arm 5 でだけヒットしていた針: treatment の content/memo と
	// procedure/medicine/consultation/inventory 各マスタ名。
	makeTreatmentSearchTreatment(t, db, &model.Treatment{
		MedicalRecordID: targetRecord.ID,
		ItemType:        model.TreatmentItemTypeOther,
		Content:         "content-needle-237",
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

	kanaMedicine := &model.Medicine{ClinicID: clinicA, Name: "アモキシシリン"}
	require.NoError(t, db.WithContext(ctx).Create(kanaMedicine).Error)
	makeTreatmentSearchTreatment(t, db, &model.Treatment{
		MedicalRecordID: targetRecord.ID,
		ItemType:        model.TreatmentItemTypeMedicine,
		MedicineID:      &kanaMedicine.ID,
	})

	// 同一語が飼主名とペット名の両腕にヒットしても UNION 側で行が重複しないことを見る。
	dedupOwner := makeTestOwner(t, db, clinicA, "同名検索飼主")
	dedupPet := makeSpeciesAndPet(t, db, clinicA, dedupOwner.ID, "同名検索飼主")
	dedupRecord := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA,
		RecordNo: "TS-MULTIARM-237",
		Date:     baseDate.Add(time.Hour),
		OwnerID:  &dedupOwner.ID,
		PetID:    &dedupPet.ID,
	})

	// LIKE ワイルドカードのエスケープは残った腕（カルテ番号 ILIKE）でも維持する。
	type wildcardFixture struct {
		name     string
		search   string
		recordNo string
		decoyNo  string
		record   *model.MedicalRecord
	}
	wildcardFixtures := []wildcardFixture{
		{name: "percent", search: "pct%needle-237", recordNo: "TS-PCT%NEEDLE-237", decoyNo: "TS-PCTXNEEDLE-237"},
		{name: "underscore", search: "under_score-237", recordNo: "TS-UNDER_SCORE-237", decoyNo: "TS-UNDERXSCORE-237"},
		{name: "backslash", search: `back\slash-237`, recordNo: `TS-BACK\SLASH-237`, decoyNo: "TS-BACKXSLASH-237"},
	}
	for i := range wildcardFixtures {
		wildcardFixtures[i].record = makeTreatmentSearchRecord(
			t, db, clinicA, wildcardFixtures[i].recordNo, baseDate.Add(time.Duration(4+i)*time.Hour),
		)
		makeTreatmentSearchRecord(t, db, clinicA, wildcardFixtures[i].decoyNo, baseDate.Add(time.Duration(7+i)*time.Hour))
	}

	deletedRecord := makeTreatmentSearchRecord(t, db, clinicA, "TS-DELETED-237", baseDate.Add(10*time.Hour))
	require.NoError(t, db.WithContext(ctx).Delete(deletedRecord).Error)

	foreignRecord := makeTreatmentSearchRecord(t, db, clinicB, "TS-FOREIGN-237", baseDate.Add(11*time.Hour))
	makeTreatmentSearchTreatment(t, db, &model.Treatment{
		MedicalRecordID: foreignRecord.ID,
		ItemType:        model.TreatmentItemTypeOther,
		Content:         "foreign-record-content-237",
	})

	pagedRecords := make([]*model.MedicalRecord, 3)
	for i := range pagedRecords {
		pagedRecords[i] = makeTreatmentSearchRecord(
			t,
			db,
			clinicA,
			"TS-PAGE-"+string(rune('A'+i))+"-237",
			baseDate.Add(time.Duration(20+i)*time.Hour),
		)
	}

	t.Run("治療内容・メモ・治療マスタ名では検索ヒットしない", func(t *testing.T) {
		for _, search := range []string{
			"content-needle-237",
			"memo-needle-237",
			"procedure-needle-237",
			"medicine-needle-237",
			"consultation-needle-237",
			"inventory-needle-237",
		} {
			assertTreatmentSearchEmpty(t, repo, []uint64{clinicA}, search)
		}
	})

	t.Run("カタカナの治療マスタ名もカタカナ・ひらがなどちらでもヒットしない", func(t *testing.T) {
		for _, search := range []string{"アモキシ", "あもきし"} {
			assertTreatmentSearchEmpty(t, repo, []uint64{clinicA}, search)
		}
	})

	t.Run("検索対象はカルテ番号・飼主名/カナ・ペット名/カナ・主訴の4条件", func(t *testing.T) {
		assertTreatmentSearchResult(t, repo, []uint64{clinicA}, "ts-target-237", targetRecord.ID)
		assertTreatmentSearchResult(t, repo, []uint64{clinicA}, "四条件飼主", targetRecord.ID)
		assertTreatmentSearchResult(t, repo, []uint64{clinicA}, "よんじょうけん", targetRecord.ID)
		assertTreatmentSearchResult(t, repo, []uint64{clinicA}, "四条件ペット", targetRecord.ID)
		assertTreatmentSearchResult(t, repo, []uint64{clinicA}, "しかいぺっと", targetRecord.ID)
		assertTreatmentSearchResult(t, repo, []uint64{clinicA}, "四条件主訴", targetRecord.ID)
	})

	t.Run("複数の検索腕に一致しても行が重複しない", func(t *testing.T) {
		assertTreatmentSearchResult(t, repo, []uint64{clinicA}, "同名検索飼主", dedupRecord.ID)
	})

	t.Run("別clinicのカルテは検索対象外", func(t *testing.T) {
		assertTreatmentSearchResult(t, repo, []uint64{clinicB}, "ts-foreign-237", foreignRecord.ID)
		assertTreatmentSearchEmpty(t, repo, []uint64{clinicA}, "ts-foreign-237")
		assertTreatmentSearchEmpty(t, repo, []uint64{clinicA}, "foreign-record-content-237")
	})

	t.Run("論理削除済みカルテは検索対象外", func(t *testing.T) {
		assertTreatmentSearchEmpty(t, repo, []uint64{clinicA}, "ts-deleted-237")
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
		assert.Equal(t, int64(11), total)
		assert.Len(t, got, 11)

		assertTreatmentSearchEmpty(t, repo, []uint64{clinicA}, strings.Repeat("長", 1000))
	})

	t.Run("countとページングが結果と一致する", func(t *testing.T) {
		firstPage, firstTotal, err := repo.FindAll(
			ctx,
			[]uint64{clinicA},
			MedicalRecordListFilters{Search: "ts-page"},
			1,
			2,
		)
		require.NoError(t, err)
		secondPage, secondTotal, err := repo.FindAll(
			ctx,
			[]uint64{clinicA},
			MedicalRecordListFilters{Search: "ts-page"},
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
