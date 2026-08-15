package medicalrecord

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestMedicalRecordRepository_FindAll_KanaNameSearch(t *testing.T) {
	db := setupMedicalRecordTreatmentSearchTestDB(t)
	repo := NewMedicalRecordRepository(db)
	const clinicID = uint64(1)
	baseDate := time.Date(2026, 7, 28, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		recordNo  string
		ownerName string
		petName   string
		search    string
	}{
		{
			name:      "owner_stored_katakana_query_katakana",
			recordNo:  "MR-KANA-OWNER-KK",
			ownerName: "アサギ",
			petName:   "対象外一号",
			search:    "アサギ",
		},
		{
			name:      "owner_stored_katakana_query_hiragana",
			recordNo:  "MR-KANA-OWNER-KH",
			ownerName: "コハク",
			petName:   "対象外二号",
			search:    "こはく",
		},
		{
			name:      "owner_stored_hiragana_query_katakana",
			recordNo:  "MR-KANA-OWNER-HK",
			ownerName: "すみれ",
			petName:   "対象外三号",
			search:    "スミレ",
		},
		{
			name:      "owner_stored_hiragana_query_hiragana",
			recordNo:  "MR-KANA-OWNER-HH",
			ownerName: "なずな",
			petName:   "対象外四号",
			search:    "なずな",
		},
		{
			name:      "pet_stored_katakana_query_katakana",
			recordNo:  "MR-KANA-PET-KK",
			ownerName: "対象外五号",
			petName:   "ラピス",
			search:    "ラピス",
		},
		{
			name:      "pet_stored_katakana_query_hiragana",
			recordNo:  "MR-KANA-PET-KH",
			ownerName: "対象外六号",
			petName:   "ミント",
			search:    "みんと",
		},
		{
			name:      "pet_stored_hiragana_query_katakana",
			recordNo:  "MR-KANA-PET-HK",
			ownerName: "対象外七号",
			petName:   "こむぎ",
			search:    "コムギ",
		},
		{
			name:      "pet_stored_hiragana_query_hiragana",
			recordNo:  "MR-KANA-PET-HH",
			ownerName: "対象外八号",
			petName:   "つばき",
			search:    "つばき",
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner := makeTestOwner(t, db, clinicID, tt.ownerName)
			pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, tt.petName)
			require.Empty(t, owner.NameKana)
			require.Empty(t, pet.NameKana)

			// Keep every non-name search surface unable to match: record_no is ASCII,
			// and this fixture creates no inquiry, treatment, or treatment master rows.
			record := makeFullMedicalRecord(t, db, &model.MedicalRecord{
				ClinicID: clinicID,
				RecordNo: tt.recordNo,
				Date:     baseDate.AddDate(0, 0, i),
				OwnerID:  &owner.ID,
				PetID:    &pet.ID,
			})

			records, total, err := repo.FindAll(
				context.Background(),
				[]uint64{clinicID},
				MedicalRecordListFilters{Search: tt.search},
				1,
				100,
			)
			require.NoError(t, err)
			require.Equal(t, int64(1), total)
			require.Len(t, records, 1)
			require.Equal(t, record.ID, records[0].ID)
		})
	}
}
