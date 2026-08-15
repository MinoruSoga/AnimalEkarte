package medicalrecord

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/reservation"
)

type vaccinationReadbackFailureRepository struct {
	VaccinationRepository
}

func (r *vaccinationReadbackFailureRepository) FindByID(context.Context, uint64, uint64) (*model.Vaccination, error) {
	return nil, errors.New("forced readback failure")
}

type pauseFirstVaccinationUpdateRepository struct {
	VaccinationRepository
	reached chan struct{}
	release chan struct{}
	once    sync.Once
}

func (r *pauseFirstVaccinationUpdateRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Vaccination, error) {
	shouldPause := false
	r.once.Do(func() {
		shouldPause = true
		close(r.reached)
	})
	if shouldPause {
		select {
		case <-r.release:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return r.VaccinationRepository.Update(ctx, clinicID, id, fields)
}

func TestVaccinationService_CreateReadbackFailureRollsBack(t *testing.T) {
	db := setupVaccinationRepoTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)

	vaccine := makeVaccineMaster(t, db, clinicID, "rollback vaccine")
	realRepo := NewVaccinationRepository(db)
	repo := &vaccinationReadbackFailureRepository{VaccinationRepository: realRepo}
	svc := NewVaccinationService(
		repo,
		NewVaccineRepository(db),
		nil,
		&vaccinationRelationVerifierStub{},
		&vaccinationMedicalRecordLockerStub{},
		vaccinationTestTransactor{db: db},
	)

	got, err := svc.Create(ctx, clinicID, &CreateVaccinationInput{VaccineID: vaccine.ID, Date: time.Now()})
	require.Error(t, err)
	assert.Nil(t, got)

	var count int64
	require.NoError(t, db.WithContext(ctx).Model(&model.Vaccination{}).Where("clinic_id = ?", clinicID).Count(&count).Error)
	assert.Zero(t, count, "readback failure must roll back the preceding create")
}

func TestVaccinationDoctorVerifier_RejectsAssignedNonDoctor(t *testing.T) {
	db := setupVaccinationRepoTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)
	ensureVaccinationTestClinics(t, db, clinicID)

	doctor := &model.Staff{ClinicID: clinicID, Name: "assigned doctor", StaffType: model.StaffTypeDoctor, IsActive: true}
	nurse := &model.Staff{ClinicID: clinicID, Name: "assigned nurse", StaffType: model.StaffTypeNurse, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(doctor).Error)
	require.NoError(t, db.WithContext(ctx).Create(nurse).Error)
	for _, staffID := range []uint64{doctor.ID, nurse.ID} {
		require.NoError(t, db.WithContext(ctx).Create(&model.StaffClinicAssignment{StaffID: staffID, ClinicID: clinicID}).Error)
	}

	verifier := reservation.NewReservationRepository(db)
	require.NoError(t, verifier.AssertMedicalRecordDoctorInClinic(ctx, clinicID, doctor.ID))
	err := verifier.AssertMedicalRecordDoctorInClinic(ctx, clinicID, nurse.ID)
	require.Error(t, err)
}

func TestVaccinationService_RelationChangeWaitsForValidationTransaction(t *testing.T) {
	db := setupVaccinationRepoTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "lock owner A")
	ownerB := makeTestOwner(t, db, clinicB, "lock owner B")
	petA := makeVaccinationRepoTestPet(t, db, clinicA, ownerA.ID, "lock pet A")
	vaccineA := makeVaccineMaster(t, db, clinicA, "lock vaccine A")
	realRepo := NewVaccinationRepository(db)
	record := &model.Vaccination{ClinicID: clinicA, PetID: &petA.ID, VaccineID: vaccineA.ID, Date: time.Now()}
	require.NoError(t, realRepo.Create(ctx, record))

	pausingRepo := &pauseFirstVaccinationUpdateRepository{
		VaccinationRepository: realRepo,
		reached:               make(chan struct{}),
		release:               make(chan struct{}),
	}
	svc := NewVaccinationService(
		pausingRepo,
		NewVaccineRepository(db),
		nil,
		reservation.NewReservationRepository(db),
		NewMedicalRecordRepository(db),
		vaccinationTestTransactor{db: db},
	)

	type updateResult struct {
		vaccination *model.Vaccination
		err         error
	}
	serviceDone := make(chan updateResult, 1)
	remarks := "validated"
	go func() {
		vaccination, err := svc.Update(ctx, clinicA, record.ID, &UpdateVaccinationInput{Remarks: &remarks})
		serviceDone <- updateResult{vaccination: vaccination, err: err}
	}()

	select {
	case <-pausingRepo.reached:
	case <-ctx.Done():
		t.Fatal("service did not reach the post-validation update gate")
	}

	ownerChangeDone := make(chan error, 1)
	go func() {
		ownerChangeDone <- db.WithContext(ctx).Model(&model.Pet{}).Where("id = ?", petA.ID).Update("owner_id", ownerB.ID).Error
	}()

	select {
	case err := <-ownerChangeDone:
		t.Fatalf("pet-owner change bypassed the validation transaction lock: %v", err)
	case <-time.After(250 * time.Millisecond):
		// Expected: FindPetOwnerInClinic's FOR SHARE lock holds the pet relation until commit.
	}
	close(pausingRepo.release)

	var result updateResult
	select {
	case result = <-serviceDone:
	case <-ctx.Done():
		t.Fatal("vaccination update did not complete")
	}
	require.NoError(t, result.err)
	require.NotNil(t, result.vaccination)
	require.NotNil(t, result.vaccination.Pet)
	require.NotNil(t, result.vaccination.Pet.Owner)
	assert.Equal(t, ownerA.ID, result.vaccination.Pet.Owner.ID, "response must use the relation validated in the same transaction")

	select {
	case err := <-ownerChangeDone:
		require.NoError(t, err)
	case <-ctx.Done():
		t.Fatal("pet-owner change did not resume after vaccination transaction committed")
	}

	got, err := realRepo.FindByID(context.Background(), clinicA, record.ID)
	require.Error(t, err, "direct post-commit pollution must be contained by read scope")
	assert.Nil(t, got)
}

func TestVaccinationService_DoctorAssignmentDeletionWaitsForValidationTransaction(t *testing.T) {
	db := setupVaccinationRepoTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	const clinicID = uint64(1)
	ensureVaccinationTestClinics(t, db, clinicID)

	owner := makeTestOwner(t, db, clinicID, "doctor lock owner")
	pet := makeVaccinationRepoTestPet(t, db, clinicID, owner.ID, "doctor lock pet")
	vaccine := makeVaccineMaster(t, db, clinicID, "doctor lock vaccine")
	doctor := &model.Staff{ClinicID: clinicID, Name: "doctor lock doctor", StaffType: model.StaffTypeDoctor, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(doctor).Error)
	assignment := &model.StaffClinicAssignment{StaffID: doctor.ID, ClinicID: clinicID}
	require.NoError(t, db.WithContext(ctx).Create(assignment).Error)

	realRepo := NewVaccinationRepository(db)
	record := &model.Vaccination{
		ClinicID: clinicID, PetID: &pet.ID, VaccineID: vaccine.ID,
		DoctorID: &doctor.ID, Date: time.Now(),
	}
	require.NoError(t, realRepo.Create(ctx, record))
	pausingRepo := &pauseFirstVaccinationUpdateRepository{
		VaccinationRepository: realRepo,
		reached:               make(chan struct{}),
		release:               make(chan struct{}),
	}
	svc := NewVaccinationService(
		pausingRepo,
		NewVaccineRepository(db),
		nil,
		reservation.NewReservationRepository(db),
		NewMedicalRecordRepository(db),
		vaccinationTestTransactor{db: db},
	)

	type updateResult struct {
		vaccination *model.Vaccination
		err         error
	}
	serviceDone := make(chan updateResult, 1)
	remarks := "doctor validated"
	go func() {
		vaccination, err := svc.Update(ctx, clinicID, record.ID, &UpdateVaccinationInput{Remarks: &remarks})
		serviceDone <- updateResult{vaccination: vaccination, err: err}
	}()
	select {
	case <-pausingRepo.reached:
	case <-ctx.Done():
		t.Fatal("service did not reach the post-validation update gate")
	}

	assignmentDeleteDone := make(chan error, 1)
	go func() {
		assignmentDeleteDone <- db.WithContext(ctx).Delete(assignment).Error
	}()
	select {
	case err := <-assignmentDeleteDone:
		t.Fatalf("doctor assignment deletion bypassed the validation transaction lock: %v", err)
	case <-time.After(250 * time.Millisecond):
	}
	close(pausingRepo.release)

	var result updateResult
	select {
	case result = <-serviceDone:
	case <-ctx.Done():
		t.Fatal("vaccination update did not complete")
	}
	require.NoError(t, result.err)
	require.NotNil(t, result.vaccination)
	require.NotNil(t, result.vaccination.Doctor)
	response := toVaccinationResponse(result.vaccination)
	require.NotNil(t, response.DoctorID)
	assert.Equal(t, doctor.ID, *response.DoctorID)

	select {
	case err := <-assignmentDeleteDone:
		require.NoError(t, err)
	case <-ctx.Done():
		t.Fatal("doctor assignment deletion did not resume after vaccination transaction committed")
	}
	got, err := realRepo.FindByID(context.Background(), clinicID, record.ID)
	require.NoError(t, err)
	assert.Nil(t, got.Doctor)
	assert.Nil(t, toVaccinationResponse(got).DoctorID)
}

func TestVaccinationService_ConcurrentUpdatesSerialize(t *testing.T) {
	db := setupVaccinationRepoTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "concurrent owner")
	pet := makeVaccinationRepoTestPet(t, db, clinicID, owner.ID, "concurrent pet")
	vaccine := makeVaccineMaster(t, db, clinicID, "concurrent vaccine")
	realRepo := NewVaccinationRepository(db)
	record := &model.Vaccination{ClinicID: clinicID, PetID: &pet.ID, VaccineID: vaccine.ID, Date: time.Now()}
	require.NoError(t, realRepo.Create(ctx, record))

	pausingRepo := &pauseFirstVaccinationUpdateRepository{
		VaccinationRepository: realRepo,
		reached:               make(chan struct{}),
		release:               make(chan struct{}),
	}
	svc := NewVaccinationService(
		pausingRepo,
		NewVaccineRepository(db),
		nil,
		reservation.NewReservationRepository(db),
		NewMedicalRecordRepository(db),
		vaccinationTestTransactor{db: db},
	)

	firstDone := make(chan error, 1)
	firstRemarks := "first"
	go func() {
		_, err := svc.Update(ctx, clinicID, record.ID, &UpdateVaccinationInput{Remarks: &firstRemarks})
		firstDone <- err
	}()
	select {
	case <-pausingRepo.reached:
	case <-ctx.Done():
		t.Fatal("first update did not reach the lock gate")
	}

	secondDone := make(chan error, 1)
	secondRemarks := "second"
	go func() {
		_, err := svc.Update(ctx, clinicID, record.ID, &UpdateVaccinationInput{Remarks: &secondRemarks})
		secondDone <- err
	}()
	select {
	case err := <-secondDone:
		t.Fatalf("concurrent update bypassed the vaccination FOR UPDATE lock: %v", err)
	case <-time.After(250 * time.Millisecond):
	}
	close(pausingRepo.release)

	select {
	case err := <-firstDone:
		require.NoError(t, err)
	case <-ctx.Done():
		t.Fatal("first update did not complete")
	}
	select {
	case err := <-secondDone:
		require.NoError(t, err)
	case <-ctx.Done():
		t.Fatal("second update did not complete after the first committed")
	}

	got, err := realRepo.FindByID(context.Background(), clinicID, record.ID)
	require.NoError(t, err)
	assert.Equal(t, secondRemarks, got.Remarks)
}
