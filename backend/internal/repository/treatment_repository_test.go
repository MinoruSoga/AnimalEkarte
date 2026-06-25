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

// setupTreatmentHistoryTestDB は setupTestDB を拡張し pets / treatments / procedures 周りを整備する。
// FindHistoryByPetID は medical_records JOIN で clinic 隔離するため medical_records も使う。
// medical_records.pet_id は pets への FK があるため pets/animal_species/owners も用意する。
func setupTreatmentHistoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	if err := db.AutoMigrate(&model.AnimalSpecies{}, &model.Pet{}, &model.Procedure{}, &model.Treatment{}); err != nil {
		t.Fatalf("failed to migrate pets/treatments/procedures: %v", err)
	}
	db.Exec("TRUNCATE TABLE treatments CASCADE")
	db.Exec("TRUNCATE TABLE medical_records CASCADE")
	db.Exec("TRUNCATE TABLE procedures CASCADE")
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

func makeHistoryTreatmentWithProcedure(t *testing.T, db *gorm.DB, medicalRecordID, procedureID uint64, content string) {
	t.Helper()
	tr := &model.Treatment{
		MedicalRecordID: medicalRecordID,
		ProcedureID:     &procedureID,
		ItemType:        model.TreatmentItemTypeProcedure,
		Content:         content,
		SortOrder:       0,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(tr).Error)
}

func makeProcedure(t *testing.T, db *gorm.DB, clinicID uint64, name string, anesthesia model.AnesthesiaType, isSurgery bool) *model.Procedure {
	t.Helper()
	p := &model.Procedure{
		ClinicID:   clinicID,
		Name:       name,
		Anesthesia: anesthesia,
		IsSurgery:  isSurgery,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(p).Error)
	return p
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

	// clinic A, pet: 2 カルテ・3 treatment（投薬2 + 処置1）
	mrOld := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "A-OLD", older)
	mrNew := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "A-NEW", newer)
	makeHistoryTreatment(t, db, mrOld.ID, model.TreatmentItemTypeMedicine, "旧・投薬", 0)
	makeHistoryTreatment(t, db, mrNew.ID, model.TreatmentItemTypeMedicine, "新・投薬", 0)
	makeHistoryTreatment(t, db, mrNew.ID, model.TreatmentItemTypeProcedure, "新・処置", 1)

	// 同一 pet だが clinic B のカルテ（clinic_id フィルタが効くかを検証）→ 混入してはならない
	mrOther := makeHistoryMedicalRecord(t, db, clinicB, pet.ID, "B-1", newer)
	makeHistoryTreatment(t, db, mrOther.ID, model.TreatmentItemTypeMedicine, "別clinic・投薬", 0)

	t.Run("returns only same-clinic treatments (clinic isolation)", func(t *testing.T) {
		got, total, err := repo.FindHistoryByPetID(ctx, clinicA, pet.ID, model.PetTreatmentHistoryFilter{}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total, "別 clinic の treatment は混入しない")
		require.Len(t, got, 3)
		for _, tr := range got {
			assert.NotEqual(t, "別clinic・投薬", tr.Content)
		}
	})

	t.Run("orders by medical_records.date DESC", func(t *testing.T) {
		got, _, err := repo.FindHistoryByPetID(ctx, clinicA, pet.ID, model.PetTreatmentHistoryFilter{}, 1, 100)
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
		got, total, err := repo.FindHistoryByPetID(ctx, clinicA, pet.ID, model.PetTreatmentHistoryFilter{ItemType: &medicine}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		require.Len(t, got, 2)
		for _, tr := range got {
			assert.Equal(t, model.TreatmentItemTypeMedicine, tr.ItemType)
		}
	})

	t.Run("returns empty for pet with no history", func(t *testing.T) {
		got, total, err := repo.FindHistoryByPetID(ctx, clinicA, uint64(999999), model.PetTreatmentHistoryFilter{}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Empty(t, got)
	})
}

func TestTreatmentRepository_FindHistoryByPetID_ProcedureFilters(t *testing.T) {
	db := setupTreatmentHistoryTestDB(t)
	repo := NewTreatmentRepository(db)
	ctx := context.Background()

	const clinicA = uint64(1)

	owner := makeOwner(t, db, clinicA, "飼主B")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "タマ")

	date := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "PROC-1", date)

	// 麻酔あり処置（anesthesia != none）
	procAnesthesia := makeProcedure(t, db, clinicA, "麻酔処置", model.AnesthesiaTypeGeneral, false)
	// 手術処置（is_surgery = true）
	procSurgery := makeProcedure(t, db, clinicA, "手術", model.AnesthesiaTypeNone, true)
	// 麻酔なし・手術なし（通常処置）
	procNormal := makeProcedure(t, db, clinicA, "通常処置", model.AnesthesiaTypeNone, false)
	// 投薬（procedure_id なし）
	makeHistoryTreatment(t, db, mr.ID, model.TreatmentItemTypeMedicine, "投薬", 3)

	makeHistoryTreatmentWithProcedure(t, db, mr.ID, procAnesthesia.ID, "麻酔処置実施")
	makeHistoryTreatmentWithProcedure(t, db, mr.ID, procSurgery.ID, "手術実施")
	makeHistoryTreatmentWithProcedure(t, db, mr.ID, procNormal.ID, "通常処置実施")

	t.Run("AnesthesiaOnly returns only treatments linked to anesthetic procedures", func(t *testing.T) {
		got, total, err := repo.FindHistoryByPetID(ctx, clinicA, pet.ID, model.PetTreatmentHistoryFilter{AnesthesiaOnly: true}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total, "麻酔あり処置のみ")
		require.Len(t, got, 1)
		require.NotNil(t, got[0].Procedure)
		assert.Equal(t, model.AnesthesiaTypeGeneral, got[0].Procedure.Anesthesia)
		assert.Equal(t, "麻酔処置実施", got[0].Content)
	})

	t.Run("IsSurgery returns only treatments linked to surgery procedures", func(t *testing.T) {
		got, total, err := repo.FindHistoryByPetID(ctx, clinicA, pet.ID, model.PetTreatmentHistoryFilter{IsSurgery: true}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total, "手術処置のみ")
		require.Len(t, got, 1)
		require.NotNil(t, got[0].Procedure)
		assert.True(t, got[0].Procedure.IsSurgery)
		assert.Equal(t, "手術実施", got[0].Content)
	})

	t.Run("no filter returns all treatments including medicine", func(t *testing.T) {
		got, total, err := repo.FindHistoryByPetID(ctx, clinicA, pet.ID, model.PetTreatmentHistoryFilter{}, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(4), total, "全 4 件（投薬+麻酔+手術+通常）")
		assert.Len(t, got, 4)
	})
}
