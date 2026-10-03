package medicalrecord

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// TestMedicalRecordRepository_FindAll_ColumnTextFilters は EMR-245 の表示列
// フィルタ（OwnerName / PetName / ChiefComplaint）を検証する。
// いずれも server-side の ILIKE 部分一致で、互いに・既存フィルタとも AND 結合。
func TestMedicalRecordRepository_FindAll_ColumnTextFilters(t *testing.T) {
	db := setupMedicalRecordTreatmentSearchTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	baseDate := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	// 対象カルテ: 飼主「山田花子」・ペット「モモタロウ」・主訴「食欲低下と嘔吐」。
	// name_kana は表示列フィルタでも畳込み一致の対象（横断検索と同じ仕様）。
	ownerA := makeTestOwner(t, db, clinicA, "山田花子")
	require.NoError(t, db.Model(ownerA).Update("name_kana", "ヤマダハナコ").Error)
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "モモタロウ")
	require.NoError(t, db.Model(petA).Update("name_kana", "モモタロウ").Error)
	recordA := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA,
		RecordNo: "CF-001",
		Date:     baseDate,
		OwnerID:  &ownerA.ID,
		PetID:    &petA.ID,
	})
	makeInquiryForRecord(t, db, recordA.ID, "食欲低下と嘔吐")

	// 同医院の非対象カルテ: 全列が別値。
	ownerB := makeTestOwner(t, db, clinicA, "佐藤一郎")
	petB := makeSpeciesAndPet(t, db, clinicA, ownerB.ID, "チョコ")
	recordB := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA,
		RecordNo: "CF-002",
		Date:     baseDate,
		OwnerID:  &ownerB.ID,
		PetID:    &petB.ID,
	})
	makeInquiryForRecord(t, db, recordB.ID, "定期健診")

	// 空白を含む飼主名（全角/半角空白の正規化一致を見る）。
	ownerC := makeTestOwner(t, db, clinicA, "鈴木 一郎")
	petC := makeSpeciesAndPet(t, db, clinicA, ownerC.ID, "コタロウ")
	recordC := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA,
		RecordNo: "CF-003",
		Date:     baseDate,
		OwnerID:  &ownerC.ID,
		PetID:    &petC.ID,
	})

	// 他医院の同名カルテ（clinic 隔離と行ごとの EXISTS 相関を見る）。
	ownerForeign := makeTestOwner(t, db, clinicB, "山田次郎")
	petForeign := makeSpeciesAndPet(t, db, clinicB, ownerForeign.ID, "モモジロウ")
	recordForeign := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicB,
		RecordNo: "CF-B01",
		Date:     baseDate,
		OwnerID:  &ownerForeign.ID,
		PetID:    &petForeign.ID,
	})
	makeInquiryForRecord(t, db, recordForeign.ID, "食欲低下と嘔吐")

	// 削除済み飼主のカルテは飼主名フィルタにヒットしない。
	ownerDeleted := makeTestOwner(t, db, clinicA, "削除対象飼主")
	petDeletedOwner := makeSpeciesAndPet(t, db, clinicA, ownerDeleted.ID, "ケイタイ")
	makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA,
		RecordNo: "CF-004",
		Date:     baseDate,
		OwnerID:  &ownerDeleted.ID,
		PetID:    &petDeletedOwner.ID,
	})
	require.NoError(t, db.Delete(ownerDeleted).Error)

	// 削除済みペットのカルテはペット名フィルタにヒットしない。
	ownerD := makeTestOwner(t, db, clinicA, "高橋次郎")
	petDeleted := makeSpeciesAndPet(t, db, clinicA, ownerD.ID, "デリート")
	makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA,
		RecordNo: "CF-005",
		Date:     baseDate,
		OwnerID:  &ownerD.ID,
		PetID:    &petDeleted.ID,
	})
	require.NoError(t, db.Delete(petDeleted).Error)

	// 削除済みカルテは列フィルタでも除外される。
	ownerE := makeTestOwner(t, db, clinicA, "削除済カルテ飼主")
	petE := makeSpeciesAndPet(t, db, clinicA, ownerE.ID, "サクラ")
	recordDeleted := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA,
		RecordNo: "CF-006",
		Date:     baseDate,
		OwnerID:  &ownerE.ID,
		PetID:    &petE.ID,
	})
	require.NoError(t, db.Delete(recordDeleted).Error)

	// 既存のマスタIDフィルタとの併用: 対象カルテにだけ診察「CF再診」を処置。
	consultation := &model.Consultation{ClinicID: clinicA, Name: "CF再診"}
	require.NoError(t, db.WithContext(ctx).Create(consultation).Error)
	makeTreatmentSearchTreatment(t, db, &model.Treatment{
		MedicalRecordID: recordA.ID,
		ItemType:        model.TreatmentItemTypeConsultation,
		ConsultationID:  &consultation.ID,
	})

	findIDs := func(t *testing.T, clinicIDs []uint64, filters MedicalRecordListFilters, page, limit int) ([]uint64, int64) {
		t.Helper()
		got, total, err := repo.FindAll(ctx, clinicIDs, filters, page, limit)
		require.NoError(t, err)
		ids := make([]uint64, 0, len(got))
		for i := range got {
			ids = append(ids, got[i].ID)
		}
		return ids, total
	}

	tests := []struct {
		name      string
		clinicIDs []uint64
		filters   MedicalRecordListFilters
		wantIDs   []uint64
	}{
		{
			name:      "飼主名の部分一致",
			clinicIDs: []uint64{clinicA},
			filters:   MedicalRecordListFilters{OwnerName: "山田"},
			wantIDs:   []uint64{recordA.ID},
		},
		{
			name:      "飼主名は name_kana でも一致（横断検索と同じカナ畳込み）",
			clinicIDs: []uint64{clinicA},
			filters:   MedicalRecordListFilters{OwnerName: "やまだ"},
			wantIDs:   []uint64{recordA.ID},
		},
		{
			name:      "ペット名の部分一致",
			clinicIDs: []uint64{clinicA},
			filters:   MedicalRecordListFilters{PetName: "モモ"},
			wantIDs:   []uint64{recordA.ID},
		},
		{
			name:      "ペット名は name_kana でも一致",
			clinicIDs: []uint64{clinicA},
			filters:   MedicalRecordListFilters{PetName: "ももたろう"},
			wantIDs:   []uint64{recordA.ID},
		},
		{
			name:      "主訴の部分一致",
			clinicIDs: []uint64{clinicA},
			filters:   MedicalRecordListFilters{ChiefComplaint: "嘔吐"},
			wantIDs:   []uint64{recordA.ID},
		},
		{
			name:      "3条件は独立して AND 結合される",
			clinicIDs: []uint64{clinicA},
			filters:   MedicalRecordListFilters{OwnerName: "山田", PetName: "モモ", ChiefComplaint: "嘔吐"},
			wantIDs:   []uint64{recordA.ID},
		},
		{
			name:      "1条件でも不一致なら0件",
			clinicIDs: []uint64{clinicA},
			filters:   MedicalRecordListFilters{OwnerName: "山田", PetName: "チョコ"},
			wantIDs:   []uint64{},
		},
		{
			name:      "横断検索 Search とも AND で併用できる",
			clinicIDs: []uint64{clinicA},
			filters:   MedicalRecordListFilters{Search: "山田", ChiefComplaint: "定期"},
			wantIDs:   []uint64{},
		},
		{
			name:      "全角空白を含む入力は正規化して一致する",
			clinicIDs: []uint64{clinicA},
			filters:   MedicalRecordListFilters{OwnerName: "鈴木　一郎"},
			wantIDs:   []uint64{recordC.ID},
		},
		{
			name:      "正規化後に空になる入力はヒットしない",
			clinicIDs: []uint64{clinicA},
			filters:   MedicalRecordListFilters{OwnerName: "　"},
			wantIDs:   []uint64{},
		},
		{
			name:      "LIKE ワイルドカード % はエスケープされる",
			clinicIDs: []uint64{clinicA},
			filters:   MedicalRecordListFilters{OwnerName: "%"},
			wantIDs:   []uint64{},
		},
		{
			name:      "LIKE ワイルドカード _ はエスケープされる",
			clinicIDs: []uint64{clinicA},
			filters:   MedicalRecordListFilters{PetName: "モモタロ_"},
			wantIDs:   []uint64{},
		},
		{
			name:      "既存のマスタIDフィルタと AND 結合される",
			clinicIDs: []uint64{clinicA},
			filters:   MedicalRecordListFilters{OwnerName: "山田", ConsultationID: &consultation.ID},
			wantIDs:   []uint64{recordA.ID},
		},
		{
			name:      "マスタIDフィルタと列フィルタのどちらか不一致なら0件",
			clinicIDs: []uint64{clinicA},
			filters:   MedicalRecordListFilters{OwnerName: "佐藤", ConsultationID: &consultation.ID},
			wantIDs:   []uint64{},
		},
		{
			name:      "他医院の同名飼主は隔離される",
			clinicIDs: []uint64{clinicA},
			filters:   MedicalRecordListFilters{OwnerName: "山田"},
			wantIDs:   []uint64{recordA.ID},
		},
		{
			name:      "複数医院スコープでは行ごとの EXISTS 相関で各自の clinic に一致",
			clinicIDs: []uint64{clinicA, clinicB},
			filters:   MedicalRecordListFilters{OwnerName: "山田"},
			wantIDs:   []uint64{recordA.ID, recordForeign.ID},
		},
		{
			name:      "削除済み飼主のカルテはヒットしない",
			clinicIDs: []uint64{clinicA},
			filters:   MedicalRecordListFilters{OwnerName: "削除対象"},
			wantIDs:   []uint64{},
		},
		{
			name:      "削除済みペットのカルテはヒットしない",
			clinicIDs: []uint64{clinicA},
			filters:   MedicalRecordListFilters{PetName: "デリート"},
			wantIDs:   []uint64{},
		},
		{
			name:      "削除済みカルテは列フィルタでも除外される",
			clinicIDs: []uint64{clinicA},
			filters:   MedicalRecordListFilters{OwnerName: "削除済カルテ"},
			wantIDs:   []uint64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIDs, total := findIDs(t, tt.clinicIDs, tt.filters, 1, 100)
			assert.ElementsMatch(t, tt.wantIDs, gotIDs)
			assert.Equal(t, int64(len(tt.wantIDs)), total)
		})
	}

	t.Run("count は行数と一致しページングは server-side のまま", func(t *testing.T) {
		gotIDs, total := findIDs(t, []uint64{clinicA, clinicB},
			MedicalRecordListFilters{OwnerName: "山田"}, 1, 1)
		require.Len(t, gotIDs, 1)
		assert.Equal(t, int64(2), total)
	})
}
