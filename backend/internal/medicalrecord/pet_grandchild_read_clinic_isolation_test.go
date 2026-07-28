package medicalrecord

import (
	"context"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func setupPetGrandchildReadClinicIsolationDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.AnimalSpecies{},
		&model.Pet{},
		&model.Staff{},
		&model.Hospitalization{},
		&model.TreatmentPlan{},
		&model.Prescription{},
		&model.MedicalRecordAddendum{},
		&model.CheckupType{},
		&model.Checkup{},
		&model.CheckupTypeField{},
		&model.CheckupFieldResult{},
	))
	require.NoError(t, db.Exec(`
		TRUNCATE TABLE
			medical_record_addenda,
			checkup_field_results,
			treatment_plans,
			prescriptions,
			checkups,
			hospitalizations,
			staffs,
			pets,
			animal_species
		CASCADE
	`).Error)
	return db
}

func createGrandchildMedicalRecord(
	t *testing.T,
	db *gorm.DB,
	clinicID uint64,
	recordNo string,
	ownerID, petID *uint64,
) *model.MedicalRecord {
	t.Helper()
	record := &model.MedicalRecord{
		ClinicID: clinicID,
		RecordNo: recordNo,
		Date:     time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC),
		OwnerID:  ownerID,
		PetID:    petID,
		Status:   model.MedicalRecordStatusFinalized,
	}
	require.NoError(t, db.Create(record).Error)
	return record
}

func assertGrandchildNotFound(t *testing.T, isNil bool, err error) {
	t.Helper()
	assert.True(t, isNil)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "expected NotFound, got %v", err)
}

func TestPetGrandchildReadClinicIsolation(t *testing.T) {
	db := setupPetGrandchildReadClinicIsolationDB(t)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := testdb.MakeTestOwner(t, db, clinicA, "孫read飼主A")
	ownerOtherA := testdb.MakeTestOwner(t, db, clinicA, "孫read別飼主A")
	ownerB := testdb.MakeTestOwner(t, db, clinicB, "孫read飼主B")
	petOtherA := testdb.MakeSpeciesAndPet(t, db, clinicA, ownerOtherA.ID, "孫read別ペットA")
	petB := testdb.MakeSpeciesAndPet(t, db, clinicB, ownerB.ID, "孫readペットB")

	mrB := createGrandchildMedicalRecord(t, db, clinicB, "GRANDCHILD-MR-B", &ownerB.ID, &petB.ID)
	hospB := makeHospitalizationRec(t, db, clinicB, ownerB.ID, petB.ID, nil)
	staffA := makeDoctor(t, db, clinicA, "孫read医師A")
	addendumByMR := &model.MedicalRecordAddendum{
		ClinicID:        clinicA,
		MedicalRecordID: mrB.ID,
		AuthorUserID:    staffA.ID,
		AfterText:       "他院カルテ追記",
		Reason:          "親clinic相関テスト",
	}
	require.NoError(t, db.Create(addendumByMR).Error)

	treatmentByMR := makeTreatmentPlan(t, db, clinicA, &mrB.ID, nil, "他院カルテ治療計画", 1)
	treatmentByHospitalization := makeTreatmentPlan(t, db, clinicA, nil, &hospB.ID, "他院入院治療計画", 1)

	prescriptionByMR := &model.Prescription{
		ClinicID: clinicA, OwnerID: ownerA.ID, MedicalRecordID: &mrB.ID,
		PrescribedAt: time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC), DurationDays: 7,
	}
	require.NoError(t, db.Create(prescriptionByMR).Error)
	prescriptionByOwner := &model.Prescription{
		ClinicID: clinicA, OwnerID: ownerB.ID,
		PrescribedAt: time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC), DurationDays: 7,
	}
	require.NoError(t, db.Create(prescriptionByOwner).Error)
	prescriptionByPet := &model.Prescription{
		ClinicID: clinicA, OwnerID: ownerA.ID, PetID: &petB.ID,
		PrescribedAt: time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC), DurationDays: 7,
	}
	require.NoError(t, db.Create(prescriptionByPet).Error)
	prescriptionOwnerPetMismatch := &model.Prescription{
		ClinicID: clinicA, OwnerID: ownerA.ID, PetID: &petOtherA.ID,
		PrescribedAt: time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC), DurationDays: 7,
	}
	require.NoError(t, db.Create(prescriptionOwnerPetMismatch).Error)

	checkupTypeB := makeCheckupTypeMaster(t, db, clinicB, "孫read他院健診")
	checkupBID := makeCheckupRec(t, db, clinicB, mrB.ID, petB.ID, checkupTypeB.ID)
	checkupResult := &model.CheckupFieldResult{
		ClinicID: clinicA, CheckupID: checkupBID, FieldName: "他院健診結果",
		FieldType: model.CheckupFieldTypeText, ValueList: pq.StringArray{},
	}
	require.NoError(t, db.Create(checkupResult).Error)

	t.Run("treatment plans reject every non-null foreign-clinic parent", func(t *testing.T) {
		repo := NewTreatmentPlanRepository(db)

		byMR, err := repo.FindByMedicalRecordID(ctx, clinicA, mrB.ID)
		require.NoError(t, err)
		assert.Empty(t, byMR)

		byHospitalization, err := repo.FindByHospitalizationID(ctx, clinicA, hospB.ID)
		require.NoError(t, err)
		assert.Empty(t, byHospitalization)

		count, err := NewHospitalizationRepository(db).
			CountTreatmentPlansByHospitalizationID(ctx, clinicA, hospB.ID)
		require.NoError(t, err)
		assert.Zero(t, count)

		got, err := repo.FindByID(ctx, clinicA, treatmentByMR.ID)
		assertGrandchildNotFound(t, got == nil, err)
		got, err = repo.FindByID(ctx, clinicA, treatmentByHospitalization.ID)
		assertGrandchildNotFound(t, got == nil, err)
	})

	t.Run("checkup results reject a foreign-clinic checkup parent", func(t *testing.T) {
		got, err := NewCheckupFieldResultRepository(db).FindByCheckupID(ctx, clinicA, checkupBID)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("medical record addenda reject a foreign-clinic medical record parent by ID", func(t *testing.T) {
		got, err := NewMedicalRecordAddendumRepository(db).FindByID(ctx, clinicA, addendumByMR.ID)
		assertGrandchildNotFound(t, got == nil, err)
	})

	t.Run("prescriptions reject foreign parents and allow historical owner snapshots", func(t *testing.T) {
		repo := NewPrescriptionRepository(db)

		byMR, err := repo.FindByMedicalRecordID(ctx, clinicA, mrB.ID)
		require.NoError(t, err)
		assert.Empty(t, byMR)

		byOwner, err := repo.FindActiveByOwner(ctx, clinicA, ownerB.ID)
		require.NoError(t, err)
		assert.Empty(t, byOwner)

		got, err := repo.FindByID(ctx, clinicA, prescriptionByMR.ID)
		assertGrandchildNotFound(t, got == nil, err)
		got, err = repo.FindByID(ctx, clinicA, prescriptionByPet.ID)
		assertGrandchildNotFound(t, got == nil, err)
		got, err = repo.FindByID(ctx, clinicA, prescriptionOwnerPetMismatch.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, prescriptionOwnerPetMismatch.ID, got.ID, "snapshot owner may differ from the pet's current owner")
	})

}
