package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

// setupTreatmentHistoryTestDB は setupTestDB を拡張し pets / treatments 周りを整備する。
// FindHistoryByPetID は medical_records JOIN で clinic 隔離するため medical_records も使う。
// medical_records.pet_id は pets への FK があるため pets/animal_species/owners も用意する。
func setupTreatmentHistoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	if err := db.AutoMigrate(&model.AnimalSpecies{}, &model.Pet{}, &model.Treatment{}); err != nil {
		t.Fatalf("failed to migrate pets/treatments: %v", err)
	}
	db.Exec("TRUNCATE TABLE treatments CASCADE")
	db.Exec("TRUNCATE TABLE medical_records CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	return db
}

func makeHistoryMedicalRecord(t *testing.T, db *gorm.DB, clinicID, petID uint64, recordNo string, date time.Time) *model.MedicalRecord {
	t.Helper()
	pet := petID
	mr := &model.MedicalRecord{
		ClinicID: clinicID,
		RecordNo: recordNo,
		Date:     date,
		PetID:    &pet,
		Status:   model.MedicalRecordStatusFinalized,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(mr).Error)
	return mr
}

func makeHistoryTreatment(t *testing.T, db *gorm.DB, medicalRecordID uint64, itemType model.TreatmentItemType, content string, sortOrder int) {
	t.Helper()
	tr := &model.Treatment{
		MedicalRecordID: medicalRecordID,
		ItemType:        itemType,
		Content:         content,
		SortOrder:       sortOrder,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(tr).Error)
}

func TestTreatmentRepository_FindHistoryByPetID(t *testing.T) {
	db := setupTreatmentHistoryTestDB(t)
	repo := NewTreatmentRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	owner := makeOwner(t, db, clinicA, "飼主A")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "ポチ")

	older := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	newer := time.Date(2026, 6, 3, 0, 0, 0, 0, time.UTC)

	// clinic A, pet: 2 カルテ・3 treatment（投薬2 + 手術1）
	mrOld := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "A-OLD", older)
	mrNew := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "A-NEW", newer)
	makeHistoryTreatment(t, db, mrOld.ID, model.TreatmentItemTypeMedicine, "旧・投薬", 0)
	makeHistoryTreatment(t, db, mrNew.ID, model.TreatmentItemTypeMedicine, "新・投薬", 0)
	makeHistoryTreatment(t, db, mrNew.ID, model.TreatmentItemTypeProcedure, "新・処置", 1)

	// 同一 pet だが clinic B のカルテ（clinic_id フィルタが効くかを検証）→ 混入してはならない
	mrOther := makeHistoryMedicalRecord(t, db, clinicB, pet.ID, "B-1", newer)
	makeHistoryTreatment(t, db, mrOther.ID, model.TreatmentItemTypeMedicine, "別clinic・投薬", 0)

	t.Run("returns only same-clinic treatments (clinic isolation)", func(t *testing.T) {
		got, total, err := repo.FindHistoryByPetID(ctx, clinicA, pet.ID, nil, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total, "別 clinic の treatment は混入しない")
		require.Len(t, got, 3)
		for _, tr := range got {
			assert.NotEqual(t, "別clinic・投薬", tr.Content)
		}
	})

	t.Run("orders by medical_records.date DESC", func(t *testing.T) {
		got, _, err := repo.FindHistoryByPetID(ctx, clinicA, pet.ID, nil, 1, 100)
		require.NoError(t, err)
		require.Len(t, got, 3)
		// 先頭 2 件は新しいカルテ(6/3)、最後が古いカルテ(6/1)
		assert.Equal(t, mrNew.ID, got[0].MedicalRecordID)
		assert.Equal(t, mrNew.ID, got[1].MedicalRecordID)
		assert.Equal(t, mrOld.ID, got[2].MedicalRecordID)
		// preload された MedicalRecord.Date が降順
		require.NotNil(t, got[0].MedicalRecord)
		require.NotNil(t, got[2].MedicalRecord)
		assert.True(t, got[0].MedicalRecord.Date.After(got[2].MedicalRecord.Date))
	})

	t.Run("filters by item_type", func(t *testing.T) {
		medicine := model.TreatmentItemTypeMedicine
		got, total, err := repo.FindHistoryByPetID(ctx, clinicA, pet.ID, &medicine, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		require.Len(t, got, 2)
		for _, tr := range got {
			assert.Equal(t, model.TreatmentItemTypeMedicine, tr.ItemType)
		}
	})

	t.Run("returns empty for pet with no history", func(t *testing.T) {
		got, total, err := repo.FindHistoryByPetID(ctx, clinicA, uint64(999999), nil, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Empty(t, got)
	})
}
