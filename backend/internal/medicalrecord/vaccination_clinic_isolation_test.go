package medicalrecord

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

type vaccinationRelationVerifierStub struct {
	petOwners map[uint64]uint64
	doctors   map[uint64]bool
}

func (s *vaccinationRelationVerifierStub) AssertOwnerInClinic(_ context.Context, _, ownerID uint64) error {
	for _, id := range s.petOwners {
		if id == ownerID {
			return nil
		}
	}
	return apperrors.WrapNotFound("owner", "scoped")
}

func (s *vaccinationRelationVerifierStub) FindPetOwnerInClinic(_ context.Context, _, petID uint64) (uint64, error) {
	ownerID, ok := s.petOwners[petID]
	if !ok {
		return 0, apperrors.WrapNotFound("pet", "scoped")
	}
	return ownerID, nil
}

func (s *vaccinationRelationVerifierStub) AssertMedicalRecordDoctorInClinic(_ context.Context, _, doctorID uint64) error {
	if s.doctors[doctorID] {
		return nil
	}
	return apperrors.WrapNotFound("staff", "scoped")
}

type vaccinationMedicalRecordLockerStub struct {
	records map[uint64]*model.MedicalRecord
}

func (s *vaccinationMedicalRecordLockerStub) LockByIDForUpdate(_ context.Context, _, id uint64) (*model.MedicalRecord, error) {
	record, ok := s.records[id]
	if !ok {
		return nil, apperrors.WrapNotFound("medical_record", "scoped")
	}
	recordCopy := *record
	return &recordCopy, nil
}

type vaccinationTestTransactor struct{ db *gorm.DB }

func (t vaccinationTestTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	if t.db == nil {
		return fn(ctx)
	}
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(persistence.WithTxValue(ctx, tx))
	})
}

func ensureVaccinationTestClinics(t *testing.T, db *gorm.DB, clinicIDs ...uint64) {
	t.Helper()
	company := &model.Company{Name: "vaccination isolation test company"}
	require.NoError(t, db.WithContext(context.Background()).Create(company).Error)
	for _, clinicID := range clinicIDs {
		clinic := &model.Clinic{ID: clinicID, CompanyID: company.ID, Name: fmt.Sprintf("vaccination clinic %d", clinicID)}
		require.NoError(t, db.WithContext(context.Background()).Clauses(clause.OnConflict{DoNothing: true}).Create(clinic).Error)
	}
}

func TestVaccinationService_CreateRejectsCrossClinicRelations(t *testing.T) {
	const (
		clinicID       = uint64(1)
		petID          = uint64(10)
		ownerID        = uint64(20)
		medicalRecord  = uint64(30)
		otherRecord    = uint64(31)
		assignedDoctor = uint64(40)
		foreignDoctor  = uint64(41)
	)

	createCalls := 0
	repo := &mockVaccinationRepository{
		createFn: func(_ context.Context, vaccination *model.Vaccination) error {
			createCalls++
			vaccination.ID = 1
			return nil
		},
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccination, error) {
			return &model.Vaccination{ID: id}, nil
		},
	}
	verifier := &vaccinationRelationVerifierStub{
		petOwners: map[uint64]uint64{petID: ownerID},
		doctors:   map[uint64]bool{assignedDoctor: true},
	}
	locker := &vaccinationMedicalRecordLockerStub{records: map[uint64]*model.MedicalRecord{
		medicalRecord: {ID: medicalRecord, ClinicID: clinicID, PetID: ptrUint64(petID), OwnerID: ptrUint64(ownerID)},
		otherRecord:   {ID: otherRecord, ClinicID: clinicID, PetID: ptrUint64(petID + 1), OwnerID: ptrUint64(ownerID)},
	}}
	svc := NewVaccinationService(repo, okVaccineRepo(), nil, verifier, locker, vaccinationTestTransactor{})

	tests := []struct {
		name  string
		input CreateVaccinationInput
	}{
		{name: "foreign pet", input: CreateVaccinationInput{PetID: ptrUint64(999), VaccineID: 1, Date: time.Now()}},
		{name: "foreign medical record", input: CreateVaccinationInput{MedicalRecordID: ptrUint64(999), VaccineID: 1, Date: time.Now()}},
		{name: "same clinic different patient", input: CreateVaccinationInput{MedicalRecordID: ptrUint64(otherRecord), PetID: ptrUint64(petID), VaccineID: 1, Date: time.Now()}},
		{name: "unassigned doctor", input: CreateVaccinationInput{PetID: ptrUint64(petID), DoctorID: ptrUint64(foreignDoctor), VaccineID: 1, Date: time.Now()}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := createCalls
			got, err := svc.Create(context.Background(), clinicID, &tt.input)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.Equal(t, before, createCalls, "rejected relation must not write")
		})
	}

	got, err := svc.Create(context.Background(), clinicID, &CreateVaccinationInput{
		MedicalRecordID: ptrUint64(medicalRecord), PetID: ptrUint64(petID), VaccineID: 1,
		DoctorID: ptrUint64(assignedDoctor), Date: time.Now(),
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, 1, createCalls)

	missingDependencyService := NewVaccinationService(repo, okVaccineRepo(), nil, nil, nil, vaccinationTestTransactor{})
	got, err = missingDependencyService.Create(context.Background(), clinicID, &CreateVaccinationInput{VaccineID: 1, Date: time.Now()})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, 1, createCalls, "missing validation dependencies must fail before write")
}

func TestVaccinationService_UpdateRejectsCrossClinicRelations(t *testing.T) {
	const (
		clinicID      = uint64(1)
		vaccinationID = uint64(5)
		petID         = uint64(10)
		ownerID       = uint64(20)
		recordID      = uint64(30)
	)

	updateCalls := 0
	repo := &mockVaccinationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccination, error) {
			return &model.Vaccination{ID: id, ClinicID: clinicID, PetID: ptrUint64(petID), MedicalRecordID: ptrUint64(recordID), VaccineID: 1}, nil
		},
		updateFieldsFn: func(_ context.Context, _, id uint64, fields map[string]any) (*model.Vaccination, error) {
			updateCalls++
			return &model.Vaccination{ID: id, ClinicID: clinicID, PetID: ptrUint64(petID), MedicalRecordID: ptrUint64(recordID), VaccineID: 1, Remarks: fields["remarks"].(string)}, nil
		},
	}
	verifier := &vaccinationRelationVerifierStub{petOwners: map[uint64]uint64{petID: ownerID}}
	locker := &vaccinationMedicalRecordLockerStub{records: map[uint64]*model.MedicalRecord{
		recordID: {ID: recordID, ClinicID: clinicID, PetID: ptrUint64(petID), OwnerID: ptrUint64(ownerID)},
	}}
	svc := NewVaccinationService(repo, okVaccineRepo(), nil, verifier, locker, vaccinationTestTransactor{})

	foreignPetID := uint64(999)
	got, err := svc.Update(context.Background(), clinicID, vaccinationID, &UpdateVaccinationInput{PetID: &foreignPetID})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.Zero(t, updateCalls)

	otherPetID := uint64(11)
	verifier.petOwners[otherPetID] = ownerID
	got, err = svc.Update(context.Background(), clinicID, vaccinationID, &UpdateVaccinationInput{PetID: &otherPetID})
	require.Error(t, err, "pet-only PATCH must be checked against the existing medical record")
	assert.Nil(t, got)
	assert.Zero(t, updateCalls)

	remarks := "safe update"
	got, err = svc.Update(context.Background(), clinicID, vaccinationID, &UpdateVaccinationInput{Remarks: &remarks})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, 1, updateCalls)
}

func TestVaccinationRepository_RelationPreloadsAreClinicScoped(t *testing.T) {
	db := setupVaccinationRepoTestDB(t)
	repo := NewVaccinationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "scope owner A")
	ownerB := makeTestOwner(t, db, clinicB, "scope owner B")
	petA := makeVaccinationRepoTestPet(t, db, clinicA, ownerA.ID, "scope pet A")
	petB := makeVaccinationRepoTestPet(t, db, clinicB, ownerB.ID, "scope pet B")
	vaccineA := makeVaccineMaster(t, db, clinicA, "scope vaccine A")
	doctorB := makeDoctor(t, db, clinicB, "scope doctor B")
	assignedDoctor := makeDoctor(t, db, clinicB, "scope shared doctor")
	ensureVaccinationTestClinics(t, db, clinicA, clinicB)
	assignment := &model.StaffClinicAssignment{StaffID: assignedDoctor.ID, ClinicID: clinicA}
	require.NoError(t, db.WithContext(ctx).Create(assignment).Error)
	assignedRecord := &model.Vaccination{ClinicID: clinicA, PetID: &petA.ID, VaccineID: vaccineA.ID, DoctorID: &assignedDoctor.ID, Date: time.Now()}
	require.NoError(t, db.WithContext(ctx).Create(assignedRecord).Error)

	got, err := repo.FindByID(ctx, clinicA, assignedRecord.ID)
	require.NoError(t, err)
	require.NotNil(t, got.Doctor, "current active assignment must be visible")
	assert.Equal(t, assignedDoctor.ID, got.Doctor.ID)
	assignedResponse := toVaccinationResponse(got)
	require.NotNil(t, assignedResponse.DoctorID)
	assert.Equal(t, assignedDoctor.ID, *assignedResponse.DoctorID)
	require.NoError(t, db.WithContext(ctx).Model(&model.Staff{}).Where("id = ?", assignedDoctor.ID).Update("is_active", false).Error)
	got, err = repo.FindByID(ctx, clinicA, assignedRecord.ID)
	require.NoError(t, err)
	assert.Nil(t, got.Doctor, "inactive staff must not be preloaded")
	assert.Nil(t, toVaccinationResponse(got).DoctorID, "inactive staff ID must not be serialized")
	require.NoError(t, db.WithContext(ctx).Model(&model.Staff{}).Where("id = ?", assignedDoctor.ID).Update("is_active", true).Error)
	require.NoError(t, db.WithContext(ctx).Delete(assignment).Error)
	got, err = repo.FindByID(ctx, clinicA, assignedRecord.ID)
	require.NoError(t, err)
	assert.Nil(t, got.Doctor, "deleted clinic assignment must not be preloaded")
	assert.Nil(t, toVaccinationResponse(got).DoctorID, "unassigned staff ID must not be serialized")

	nurse := &model.Staff{ClinicID: clinicB, Name: "scope assigned nurse", StaffType: model.StaffTypeNurse, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(nurse).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.StaffClinicAssignment{StaffID: nurse.ID, ClinicID: clinicA}).Error)
	nurseRecord := &model.Vaccination{ClinicID: clinicA, PetID: &petA.ID, VaccineID: vaccineA.ID, DoctorID: &nurse.ID, Date: time.Now()}
	require.NoError(t, db.WithContext(ctx).Create(nurseRecord).Error)
	got, err = repo.FindByID(ctx, clinicA, nurseRecord.ID)
	require.NoError(t, err)
	assert.Nil(t, got.Doctor, "assigned non-doctor staff must not be preloaded as a doctor")
	assert.Nil(t, toVaccinationResponse(got).DoctorID)

	record := &model.Vaccination{ClinicID: clinicA, PetID: &petB.ID, VaccineID: vaccineA.ID, DoctorID: &doctorB.ID, Date: time.Now()}
	require.NoError(t, db.WithContext(ctx).Create(record).Error)

	got, err = repo.FindByID(ctx, clinicA, record.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Nil(t, got)

	pollutedPetA := makeVaccinationRepoTestPet(t, db, clinicA, ownerB.ID, "scope polluted pet A")
	pollutedOwnerRecord := &model.Vaccination{ClinicID: clinicA, PetID: &pollutedPetA.ID, VaccineID: vaccineA.ID, Date: time.Now()}
	require.NoError(t, db.WithContext(ctx).Create(pollutedOwnerRecord).Error)
	got, err = repo.FindByID(ctx, clinicA, pollutedOwnerRecord.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Nil(t, got)

	petA2 := makeVaccinationRepoTestPet(t, db, clinicA, ownerA.ID, "scope pet A2")
	foreignRecord := &model.MedicalRecord{
		ClinicID: clinicB, RecordNo: "VAC-SCOPE-MR-B", Date: time.Now(),
		OwnerID: &ownerB.ID, PetID: &petB.ID, Status: model.MedicalRecordStatusDraft,
	}
	require.NoError(t, db.WithContext(ctx).Create(foreignRecord).Error)
	foreignMedicalRecordVaccination := &model.Vaccination{
		ClinicID: clinicA, MedicalRecordID: &foreignRecord.ID, PetID: &petA2.ID,
		VaccineID: vaccineA.ID, Date: time.Now(),
	}
	require.NoError(t, db.WithContext(ctx).Create(foreignMedicalRecordVaccination).Error)
	got, err = repo.FindByID(ctx, clinicA, foreignMedicalRecordVaccination.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Nil(t, got)

	vaccineB := makeVaccineMaster(t, db, clinicB, "scope vaccine B")
	foreignVaccineRecord := &model.Vaccination{ClinicID: clinicA, PetID: &petA2.ID, VaccineID: vaccineB.ID, Date: time.Now()}
	require.NoError(t, db.WithContext(ctx).Create(foreignVaccineRecord).Error)
	got, err = repo.FindByID(ctx, clinicA, foreignVaccineRecord.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Nil(t, got)

	listed, total, err := repo.FindAll(ctx, clinicA, nil, nil, nil, nil, 1, 100)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total, "only the valid assigned-doctor and nurse rows remain")
	for _, item := range listed {
		switch item.ID {
		case record.ID, pollutedOwnerRecord.ID, foreignMedicalRecordVaccination.ID, foreignVaccineRecord.ID:
			t.Fatalf("polluted vaccination %d must be excluded from list/count", item.ID)
		}
	}
}

func TestVaccinationRepository_FindByID_AllowsHistoricalOwnerAfterPetTransfer(t *testing.T) {
	db := setupVaccinationRepoTestDB(t)
	repo := NewVaccinationRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	previousOwner := makeTestOwner(t, db, clinicID, "ワクチン譲渡前飼主")
	currentOwner := makeTestOwner(t, db, clinicID, "ワクチン譲渡後飼主")
	pet := makeVaccinationRepoTestPet(t, db, clinicID, previousOwner.ID, "ワクチン譲渡ペット")
	vaccine := makeVaccineMaster(t, db, clinicID, "譲渡後ワクチン")
	record := &model.MedicalRecord{
		ClinicID: clinicID, RecordNo: "MR-VAC-TRANSFER", Date: time.Now(),
		OwnerID: &previousOwner.ID, PetID: &pet.ID, Status: model.MedicalRecordStatusDraft,
	}
	require.NoError(t, db.Create(record).Error)
	vaccination := &model.Vaccination{
		ClinicID: clinicID, PetID: &pet.ID, MedicalRecordID: &record.ID,
		VaccineID: vaccine.ID, Date: time.Now(),
	}
	require.NoError(t, db.Create(vaccination).Error)
	require.NoError(t, db.Model(&model.Pet{}).Where("id = ?", pet.ID).Update("owner_id", currentOwner.ID).Error)

	got, err := repo.FindByID(ctx, clinicID, vaccination.ID)
	require.NoError(t, err)
	assert.Equal(t, vaccination.ID, got.ID)
	var persisted model.MedicalRecord
	require.NoError(t, db.First(&persisted, record.ID).Error)
	require.NotNil(t, persisted.OwnerID)
	assert.Equal(t, previousOwner.ID, *persisted.OwnerID)
}

func TestVaccinationRepository_FindByOwnerIsClinicScoped(t *testing.T) {
	db := setupVaccinationRepoTestDB(t)
	repo := NewVaccinationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "filter owner A")
	ownerB := makeTestOwner(t, db, clinicB, "filter owner B")
	petA := makeVaccinationRepoTestPet(t, db, clinicA, ownerA.ID, "filter pet A")
	petB := makeVaccinationRepoTestPet(t, db, clinicB, ownerB.ID, "filter pet B")
	vaccineA := makeVaccineMaster(t, db, clinicA, "filter vaccine A")
	normal := &model.Vaccination{ClinicID: clinicA, PetID: &petA.ID, VaccineID: vaccineA.ID, Date: time.Now()}
	require.NoError(t, db.WithContext(ctx).Create(normal).Error)
	polluted := &model.Vaccination{ClinicID: clinicA, PetID: &petB.ID, VaccineID: vaccineA.ID, Date: time.Now()}
	require.NoError(t, db.WithContext(ctx).Create(polluted).Error)

	for name, ownerID := range map[string]uint64{"foreign": ownerB.ID, "nonexistent": 999999} {
		t.Run(name, func(t *testing.T) {
			got, err := repo.FindByOwner(ctx, clinicA, ownerID)
			require.NoError(t, err)
			assert.Empty(t, got)
			listed, total, err := repo.FindAll(ctx, clinicA, nil, &ownerID, nil, nil, 1, 100)
			require.NoError(t, err)
			assert.Empty(t, listed)
			assert.Zero(t, total)
		})
	}

	got, err := repo.FindByOwner(ctx, clinicA, ownerA.ID)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, normal.ID, got[0].ID)
}

func TestVaccinationService_TagSyncRejectsCrossClinicOwner(t *testing.T) {
	const clinicA, clinicB = uint64(1), uint64(2)
	ownerID := uint64(10)
	petID := uint64(20)
	syncCalls := 0
	resyncCalls := 0
	tagSync := &mockLstepTagSyncService{
		syncVaccineTagFn: func(context.Context, uint64, uint64, uint64) error {
			syncCalls++
			return nil
		},
		resyncOwnerVaccineTagsFn: func(context.Context, uint64, uint64) error {
			resyncCalls++
			return nil
		},
	}
	svc := &vaccinationService{tagSyncSvc: tagSync}

	unsafe := []*model.Vaccination{
		{ID: 1, PetID: &petID, Pet: &model.Pet{ID: petID, ClinicID: clinicA, OwnerID: ownerID}},
		{ID: 2, PetID: &petID, Pet: &model.Pet{ID: petID, ClinicID: clinicA, OwnerID: ownerID, Owner: &model.Owner{ID: ownerID, ClinicID: clinicB}}},
		{ID: 3, PetID: &petID, Pet: &model.Pet{ID: petID, ClinicID: clinicA, OwnerID: ownerID, Owner: &model.Owner{ID: ownerID + 1, ClinicID: clinicA}}},
	}
	for _, vaccination := range unsafe {
		svc.syncVaccineTag(context.Background(), clinicA, vaccination)
		svc.resyncOwnerVaccineTags(context.Background(), clinicA, vaccination)
	}
	assert.Zero(t, syncCalls)
	assert.Zero(t, resyncCalls)

	safe := &model.Vaccination{ID: 4, PetID: &petID, Pet: &model.Pet{
		ID: petID, ClinicID: clinicA, OwnerID: ownerID,
		Owner: &model.Owner{ID: ownerID, ClinicID: clinicA},
	}}
	svc.syncVaccineTag(context.Background(), clinicA, safe)
	svc.resyncOwnerVaccineTags(context.Background(), clinicA, safe)
	assert.Equal(t, 1, syncCalls)
	assert.Equal(t, 1, resyncCalls)
}
