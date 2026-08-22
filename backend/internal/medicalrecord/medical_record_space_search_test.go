package medicalrecord

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestMedicalRecordRepository_FindAll_OwnerNameIdeographicSpaceFourWay(t *testing.T) {
	db := setupMedicalRecordTreatmentSearchTestDB(t)
	repo := NewMedicalRecordRepository(db)
	const clinicA, clinicB = uint64(1), uint64(2)
	baseDate := time.Date(2026, 7, 28, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		storedName string
		query      string
	}{
		{name: "DB fullwidth × query fullwidth", storedName: "カルテ全角全角姓　カルテ全角全角名", query: "カルテ全角全角姓　カルテ全角全角名"},
		{name: "DB fullwidth × query halfwidth", storedName: "カルテ全角半角姓　カルテ全角半角名", query: "カルテ全角半角姓 カルテ全角半角名"},
		{name: "DB halfwidth × query fullwidth", storedName: "カルテ半角全角姓 カルテ半角全角名", query: "カルテ半角全角姓　カルテ半角全角名"},
		{name: "DB halfwidth × query halfwidth", storedName: "カルテ半角半角姓 カルテ半角半角名", query: "カルテ半角半角姓 カルテ半角半角名"},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner := makeTestOwner(t, db, clinicA, tt.storedName)
			pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "space-search-pet-"+tt.name)
			foreignOwner := makeTestOwner(t, db, clinicB, tt.storedName)
			foreignPet := makeSpeciesAndPet(t, db, clinicB, foreignOwner.ID, "space-search-foreign-pet-"+tt.name)
			record := makeFullMedicalRecord(t, db, &model.MedicalRecord{
				ClinicID: clinicA,
				RecordNo: fmt.Sprintf("MR-SPACE-A-%d", i),
				Date:     baseDate.AddDate(0, 0, i),
				OwnerID:  &owner.ID,
				PetID:    &pet.ID,
			})
			_ = makeFullMedicalRecord(t, db, &model.MedicalRecord{
				ClinicID: clinicB,
				RecordNo: fmt.Sprintf("MR-SPACE-B-%d", i),
				Date:     baseDate.AddDate(0, 0, i),
				OwnerID:  &foreignOwner.ID,
				PetID:    &foreignPet.ID,
			})

			got, total, err := repo.FindAll(
				context.Background(),
				[]uint64{clinicA},
				MedicalRecordListFilters{Search: tt.query},
				1,
				100,
			)
			require.NoError(t, err)
			require.Equal(t, int64(1), total)
			require.Len(t, got, 1)
			require.Equal(t, record.ID, got[0].ID)
		})
	}
}
