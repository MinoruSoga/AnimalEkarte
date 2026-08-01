package medicalrecord

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/reservation"
	"github.com/animal-ekarte/backend/internal/testdb"
)

type clinicalRelationWriteFixture struct {
	db                *gorm.DB
	clinicA           uint64
	clinicB           uint64
	petA              *model.Pet
	otherPetA         *model.Pet
	petB              *model.Pet
	recordA           *model.MedicalRecord
	otherPatientRecA  *model.MedicalRecord
	recordB           *model.MedicalRecord
	assignedDoctor    *model.Staff
	unassignedDoctor  *model.Staff
	checkupType       *model.CheckupType
	examinationType   *model.ExaminationType
	relationVerifier  ClinicalRelationVerifier
	writeTransactor   Transactor
	medicalRecordRepo MedicalRecordRepository
}

func setupClinicalRelationWriteFixture(t *testing.T) *clinicalRelationWriteFixture {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.Company{},
		&model.Clinic{},
		&model.Owner{},
		&model.AnimalSpecies{},
		&model.Pet{},
		&model.MedicalRecord{},
		&model.Staff{},
		&model.StaffClinicAssignment{},
		&model.CheckupType{},
		&model.Checkup{},
		&model.ExaminationType{},
		&model.Examination{},
	))
	db.Exec("TRUNCATE TABLE checkups, exams, checkup_types, exam_types, staff_clinic_assignments, staffs, medical_records, pets, animal_species, owners CASCADE")

	const clinicA, clinicB = uint64(1), uint64(2)
	ensureVaccinationTestClinics(t, db, clinicA, clinicB)
	ownerA := makeTestOwner(t, db, clinicA, "関係write医院A飼主")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "関係write医院A患者")
	otherPetA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "関係write医院A別患者")
	ownerB := makeTestOwner(t, db, clinicB, "関係write医院B飼主")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "関係write医院B患者")

	recordA := makeClinicalRelationRecord(t, db, clinicA, ownerA.ID, petA.ID, "MR-REL-WRITE-A")
	otherPatientRecA := makeClinicalRelationRecord(t, db, clinicA, ownerA.ID, otherPetA.ID, "MR-REL-WRITE-A-OTHER")
	recordB := makeClinicalRelationRecord(t, db, clinicB, ownerB.ID, petB.ID, "MR-REL-WRITE-B")

	assignedDoctor := makeDoctor(t, db, clinicB, "関係write有効担当医")
	require.NoError(t, db.Create(&model.StaffClinicAssignment{
		StaffID: assignedDoctor.ID, ClinicID: clinicA,
	}).Error)
	unassignedDoctor := makeDoctor(t, db, clinicA, "関係write未所属医師")

	checkupType := makeCheckupTypeMaster(t, db, clinicA, "関係write健診種別")
	examinationType := makeExamTypeMaster(t, db, clinicA, "関係write検査種別")
	relationVerifier := reservation.NewReservationRepository(db)

	return &clinicalRelationWriteFixture{
		db: db, clinicA: clinicA, clinicB: clinicB,
		petA: petA, otherPetA: otherPetA, petB: petB,
		recordA: recordA, otherPatientRecA: otherPatientRecA, recordB: recordB,
		assignedDoctor: assignedDoctor, unassignedDoctor: unassignedDoctor,
		checkupType: checkupType, examinationType: examinationType,
		relationVerifier: relationVerifier, writeTransactor: persistence.NewTransactor(db),
		medicalRecordRepo: NewMedicalRecordRepository(db),
	}
}

func makeClinicalRelationRecord(
	t *testing.T,
	db *gorm.DB,
	clinicID, ownerID, petID uint64,
	recordNo string,
) *model.MedicalRecord {
	t.Helper()
	record := &model.MedicalRecord{
		ClinicID: clinicID, RecordNo: recordNo, Date: time.Now(),
		OwnerID: &ownerID, PetID: &petID, Status: model.MedicalRecordStatusDraft,
	}
	require.NoError(t, db.Create(record).Error)
	return record
}

func TestDB_CheckupServiceRejectsPollutedClinicalRelations(t *testing.T) {
	fixture := setupClinicalRelationWriteFixture(t)
	ctx := context.Background()
	repo := NewCheckupRepository(fixture.db)
	service := NewCheckupService(
		repo,
		fixture.medicalRecordRepo,
		NewCheckupTypeRepository(fixture.db),
		nil,
		nil,
		CheckupWriteDependencies{
			RelationVerifier: fixture.relationVerifier,
			Transactor:       fixture.writeTransactor,
		},
	)

	tests := []struct {
		name            string
		medicalRecordID uint64
		petID           uint64
		doctorID        uint64
	}{
		{
			name:            "foreign medical record",
			medicalRecordID: fixture.recordB.ID,
			petID:           fixture.petB.ID, doctorID: fixture.assignedDoctor.ID,
		},
		{
			name:            "foreign pet",
			medicalRecordID: fixture.recordA.ID,
			petID:           fixture.petB.ID, doctorID: fixture.assignedDoctor.ID,
		},
		{
			name:            "different same-clinic patient",
			medicalRecordID: fixture.recordA.ID,
			petID:           fixture.otherPetA.ID, doctorID: fixture.assignedDoctor.ID,
		},
		{
			name:            "inactive or unassigned doctor",
			medicalRecordID: fixture.recordA.ID,
			petID:           fixture.petA.ID, doctorID: fixture.unassignedDoctor.ID,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := countClinicalRelationRows(t, fixture.db, &model.Checkup{}, fixture.clinicA)
			got, err := service.Create(ctx, tt.medicalRecordID, &CreateCheckupInput{
				ClinicID: fixture.clinicA, CheckupTypeID: fixture.checkupType.ID,
				PetID: &tt.petID, DoctorID: &tt.doctorID, Date: time.Now(),
			})
			require.Error(t, err)
			assert.Nil(t, got)
			assert.Equal(t, before, countClinicalRelationRows(t, fixture.db, &model.Checkup{}, fixture.clinicA))
		})
	}

	validPetID := fixture.petA.ID
	validDoctorID := fixture.assignedDoctor.ID
	created, err := service.Create(ctx, fixture.recordA.ID, &CreateCheckupInput{
		ClinicID: fixture.clinicA, CheckupTypeID: fixture.checkupType.ID,
		PetID: &validPetID, DoctorID: &validDoctorID, Date: time.Now(),
	})
	require.NoError(t, err)
	require.NotNil(t, created)

	otherPetID := fixture.otherPetA.ID
	got, err := service.Update(ctx, fixture.clinicA, fixture.recordA.ID, created.ID, &UpdateCheckupInput{
		PetID: &otherPetID,
	})
	require.Error(t, err)
	assert.Nil(t, got)

	unassignedDoctorID := fixture.unassignedDoctor.ID
	got, err = service.Update(ctx, fixture.clinicA, fixture.recordA.ID, created.ID, &UpdateCheckupInput{
		DoctorID: &unassignedDoctorID,
	})
	require.Error(t, err)
	assert.Nil(t, got)

	persisted, err := repo.FindByID(ctx, fixture.clinicA, created.ID)
	require.NoError(t, err)
	assert.Equal(t, fixture.petA.ID, *persisted.PetID)
	assert.Equal(t, fixture.assignedDoctor.ID, *persisted.DoctorID)
}

func TestDB_ExaminationServiceRejectsPollutedClinicalRelations(t *testing.T) {
	fixture := setupClinicalRelationWriteFixture(t)
	ctx := context.Background()
	repo := NewExaminationRepository(fixture.db)
	service := NewExaminationService(
		repo,
		fixture.medicalRecordRepo,
		NewExamTypeRepository(fixture.db),
		&mockAuditTxLogger{},
		fixture.writeTransactor,
		fixture.relationVerifier,
	)

	tests := []struct {
		name            string
		medicalRecordID uint64
		petID           uint64
		doctorID        uint64
	}{
		{
			name:            "foreign medical record",
			medicalRecordID: fixture.recordB.ID,
			petID:           fixture.petB.ID, doctorID: fixture.assignedDoctor.ID,
		},
		{
			name:            "foreign pet",
			medicalRecordID: fixture.recordA.ID,
			petID:           fixture.petB.ID, doctorID: fixture.assignedDoctor.ID,
		},
		{
			name:            "different same-clinic patient",
			medicalRecordID: fixture.recordA.ID,
			petID:           fixture.otherPetA.ID, doctorID: fixture.assignedDoctor.ID,
		},
		{
			name:            "inactive or unassigned doctor",
			medicalRecordID: fixture.recordA.ID,
			petID:           fixture.petA.ID, doctorID: fixture.unassignedDoctor.ID,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := countClinicalRelationRows(t, fixture.db, &model.Examination{}, fixture.clinicA)
			got, err := service.Create(ctx, fixture.clinicA, &CreateExaminationInput{
				MedicalRecordID: &tt.medicalRecordID, PetID: &tt.petID,
				ExamTypeID: fixture.examinationType.ID, DoctorID: &tt.doctorID, Date: time.Now(), ActorID: ptrUint64(1),
			})
			require.Error(t, err)
			assert.Nil(t, got)
			assert.Equal(t, before, countClinicalRelationRows(t, fixture.db, &model.Examination{}, fixture.clinicA))
		})
	}

	validRecordID := fixture.recordA.ID
	validPetID := fixture.petA.ID
	validDoctorID := fixture.assignedDoctor.ID
	created, err := service.Create(ctx, fixture.clinicA, &CreateExaminationInput{
		MedicalRecordID: &validRecordID, PetID: &validPetID,
		ExamTypeID: fixture.examinationType.ID, DoctorID: &validDoctorID, Date: time.Now(), ActorID: ptrUint64(1),
	})
	require.NoError(t, err)
	require.NotNil(t, created)

	otherPetID := fixture.otherPetA.ID
	got, err := service.Update(ctx, fixture.clinicA, created.ID, UpdateExaminationInput{
		PetID: &otherPetID, ActorID: ptrUint64(1),
	})
	require.Error(t, err)
	assert.Nil(t, got)

	otherPatientRecordID := fixture.otherPatientRecA.ID
	got, err = service.Update(ctx, fixture.clinicA, created.ID, UpdateExaminationInput{
		MedicalRecordID: &otherPatientRecordID, ActorID: ptrUint64(1),
	})
	require.Error(t, err)
	assert.Nil(t, got)

	unassignedDoctorID := fixture.unassignedDoctor.ID
	got, err = service.Update(ctx, fixture.clinicA, created.ID, UpdateExaminationInput{
		DoctorID: &unassignedDoctorID, ActorID: ptrUint64(1),
	})
	require.Error(t, err)
	assert.Nil(t, got)

	persisted, err := repo.FindByID(ctx, fixture.clinicA, created.ID)
	require.NoError(t, err)
	assert.Equal(t, fixture.recordA.ID, *persisted.MedicalRecordID)
	assert.Equal(t, fixture.petA.ID, *persisted.PetID)
	assert.Equal(t, fixture.assignedDoctor.ID, *persisted.DoctorID)
}

func countClinicalRelationRows(t *testing.T, db *gorm.DB, target any, clinicID uint64) int64 {
	t.Helper()
	var count int64
	require.NoError(t, db.Model(target).Where("clinic_id = ?", clinicID).Count(&count).Error)
	return count
}
