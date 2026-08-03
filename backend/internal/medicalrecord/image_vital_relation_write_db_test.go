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
	petdomain "github.com/animal-ekarte/backend/internal/pet"
	"github.com/animal-ekarte/backend/internal/reservation"
	staffdomain "github.com/animal-ekarte/backend/internal/staff"
	"github.com/animal-ekarte/backend/internal/testdb"
)

type imageVitalRelationDBFixture struct {
	db              *gorm.DB
	clinicA         uint64
	clinicB         uint64
	petA            *model.Pet
	petB            *model.Pet
	recordA         *model.MedicalRecord
	recordB         *model.MedicalRecord
	validStaff      *model.Staff
	foreignStaff    *model.Staff
	unassignedStaff *model.Staff
	validExam       *model.Examination
	foreignExam     *model.Examination
}

func setupImageVitalRelationDBFixture(t *testing.T) *imageVitalRelationDBFixture {
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
		&model.ExaminationType{},
		&model.Examination{},
		&model.MedicalRecordImage{},
		&model.VitalRecord{},
	))
	require.NoError(t, db.Exec(`
		TRUNCATE TABLE
			medical_record_images,
			vital_records,
			exams,
			exam_types,
			staff_clinic_assignments,
			staffs,
			medical_records,
			pets,
			animal_species,
			owners
		CASCADE
	`).Error)

	company := &model.Company{Name: "image-vital relation DB fixture"}
	require.NoError(t, db.Create(company).Error)
	clinicA := &model.Clinic{CompanyID: company.ID, Name: "image-vital clinic A"}
	clinicB := &model.Clinic{CompanyID: company.ID, Name: "image-vital clinic B"}
	require.NoError(t, db.Create(clinicA).Error)
	require.NoError(t, db.Create(clinicB).Error)

	species := &model.AnimalSpecies{Name: "image-vital relation species", IsActive: true}
	require.NoError(t, db.Create(species).Error)
	ownerA := &model.Owner{ClinicID: clinicA.ID, Name: "image-vital owner A"}
	ownerB := &model.Owner{ClinicID: clinicB.ID, Name: "image-vital owner B"}
	require.NoError(t, db.Create(ownerA).Error)
	require.NoError(t, db.Create(ownerB).Error)
	petA := &model.Pet{
		ClinicID: clinicA.ID, OwnerID: ownerA.ID,
		AnimalSpeciesID: species.ID, Name: "image-vital pet A",
	}
	petB := &model.Pet{
		ClinicID: clinicB.ID, OwnerID: ownerB.ID,
		AnimalSpeciesID: species.ID, Name: "image-vital pet B",
	}
	require.NoError(t, db.Create(petA).Error)
	require.NoError(t, db.Create(petB).Error)

	recordA := createImageVitalRelationMedicalRecord(
		t,
		db,
		clinicA.ID,
		ownerA.ID,
		petA.ID,
		"MR-IMAGE-VITAL-A",
	)
	recordB := createImageVitalRelationMedicalRecord(
		t,
		db,
		clinicB.ID,
		ownerB.ID,
		petB.ID,
		"MR-IMAGE-VITAL-B",
	)

	// Primary ClinicID is not the authorization source for multi-clinic staff.
	// validStaff deliberately belongs primarily to B and is actively assigned to A.
	validStaff := createImageVitalRelationStaff(t, db, clinicB.ID, "image-vital assigned staff")
	require.NoError(t, db.Create(&model.StaffClinicAssignment{
		StaffID: validStaff.ID, ClinicID: clinicA.ID,
	}).Error)
	foreignStaff := createImageVitalRelationStaff(t, db, clinicB.ID, "image-vital foreign staff")
	require.NoError(t, db.Create(&model.StaffClinicAssignment{
		StaffID: foreignStaff.ID, ClinicID: clinicB.ID,
	}).Error)
	unassignedStaff := createImageVitalRelationStaff(t, db, clinicA.ID, "image-vital unassigned staff")

	examTypeA := &model.ExaminationType{
		ClinicID: clinicA.ID, Name: "image-vital exam type A", IsActive: true,
	}
	examTypeB := &model.ExaminationType{
		ClinicID: clinicB.ID, Name: "image-vital exam type B", IsActive: true,
	}
	require.NoError(t, db.Create(examTypeA).Error)
	require.NoError(t, db.Create(examTypeB).Error)
	validExam := &model.Examination{
		ClinicID: clinicA.ID, MedicalRecordID: &recordA.ID, PetID: &petA.ID,
		ExamTypeID: examTypeA.ID, Date: time.Date(2026, 7, 24, 0, 0, 0, 0, time.UTC),
		Status: model.ExaminationStatusPending,
	}
	foreignExam := &model.Examination{
		ClinicID: clinicB.ID, MedicalRecordID: &recordB.ID, PetID: &petB.ID,
		ExamTypeID: examTypeB.ID, Date: time.Date(2026, 7, 24, 0, 0, 0, 0, time.UTC),
		Status: model.ExaminationStatusPending,
	}
	require.NoError(t, db.Create(validExam).Error)
	require.NoError(t, db.Create(foreignExam).Error)

	return &imageVitalRelationDBFixture{
		db: db, clinicA: clinicA.ID, clinicB: clinicB.ID,
		petA: petA, petB: petB, recordA: recordA, recordB: recordB,
		validStaff: validStaff, foreignStaff: foreignStaff,
		unassignedStaff: unassignedStaff, validExam: validExam, foreignExam: foreignExam,
	}
}

func createImageVitalRelationMedicalRecord(
	t *testing.T,
	db *gorm.DB,
	clinicID, ownerID, petID uint64,
	recordNo string,
) *model.MedicalRecord {
	t.Helper()
	record := &model.MedicalRecord{
		ClinicID: clinicID, RecordNo: recordNo,
		Date:    time.Date(2026, 7, 24, 0, 0, 0, 0, time.UTC),
		OwnerID: &ownerID, PetID: &petID, Status: model.MedicalRecordStatusDraft,
	}
	require.NoError(t, db.Create(record).Error)
	return record
}

func createImageVitalRelationStaff(
	t *testing.T,
	db *gorm.DB,
	clinicID uint64,
	name string,
) *model.Staff {
	t.Helper()
	staff := &model.Staff{
		ClinicID: clinicID, Name: name,
		IsActive: true, StaffType: model.StaffTypeDoctor,
	}
	require.NoError(t, db.Create(staff).Error)
	return staff
}

func TestDB_MedicalRecordImageServiceRejectsPollutedExamAndStaffRelations(t *testing.T) {
	fixture := setupImageVitalRelationDBFixture(t)
	ctx := context.Background()
	repo := NewMedicalRecordImageRepository(fixture.db)
	service := NewMedicalRecordImageServiceWithRelationValidation(
		repo,
		NewMedicalRecordRepository(fixture.db),
		petdomain.NewRepository(fixture.db),
		NewExaminationRepository(fixture.db),
		staffdomain.NewStaffRepository(fixture.db),
		staffdomain.NewStaffClinicAssignmentRepository(fixture.db),
		persistence.NewTransactor(fixture.db),
	)

	tests := []struct {
		name    string
		examID  *uint64
		staffID *uint64
	}{
		{
			name:   "foreign clinic examination",
			examID: &fixture.foreignExam.ID,
		},
		{
			name:    "foreign clinic staff assignment",
			examID:  &fixture.validExam.ID,
			staffID: &fixture.foreignStaff.ID,
		},
		{
			name:    "unassigned staff",
			examID:  &fixture.validExam.ID,
			staffID: &fixture.unassignedStaff.ID,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.Create(
				ctx,
				fixture.clinicA,
				fixture.recordA.ID,
				&CreateMedicalRecordImageInput{
					ImageURL:  "https://example.invalid/rejected.png",
					ImageType: model.MedicalImageTypeOther,
					ExamID:    tt.examID, StaffID: tt.staffID,
				},
			)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.Zero(
				t,
				countImageVitalRelationRows(
					t,
					fixture.db,
					&model.MedicalRecordImage{},
					"medical_record_id = ?",
					fixture.recordA.ID,
				),
				"rejected request-derived relations must write zero image rows",
			)
		})
	}

	validStaffID := fixture.validStaff.ID
	created, err := service.Create(
		ctx,
		fixture.clinicA,
		fixture.recordA.ID,
		&CreateMedicalRecordImageInput{
			ImageURL:  "https://example.invalid/accepted.png",
			ImageType: model.MedicalImageTypeOther,
			ExamID:    &fixture.validExam.ID, StaffID: &validStaffID,
		},
	)
	require.NoError(t, err)
	require.NotNil(t, created)
	require.NotNil(t, created.ExamID)
	require.NotNil(t, created.StaffID)
	assert.Equal(t, fixture.validExam.ID, *created.ExamID)
	assert.Equal(t, fixture.validStaff.ID, *created.StaffID)
	assert.EqualValues(
		t,
		1,
		countImageVitalRelationRows(
			t,
			fixture.db,
			&model.MedicalRecordImage{},
			"medical_record_id = ?",
			fixture.recordA.ID,
		),
	)
}

func TestDB_VitalServiceRejectsPollutedPatientRecordAndStaffRelations(t *testing.T) {
	fixture := setupImageVitalRelationDBFixture(t)
	ctx := context.Background()
	repo := NewVitalRepository(fixture.db)
	service := NewVitalServiceWithRelationValidation(
		repo,
		NewMedicalRecordRepository(fixture.db),
		okVitalAudit(),
		reservation.NewReservationRepository(fixture.db),
		staffdomain.NewStaffRepository(fixture.db),
		staffdomain.NewStaffClinicAssignmentRepository(fixture.db),
		persistence.NewTransactor(fixture.db),
	)
	temperature := 38.2

	createTests := []struct {
		name            string
		medicalRecordID uint64
		petID           uint64
		staffID         *uint64
	}{
		{
			name:            "foreign clinic medical record",
			medicalRecordID: fixture.recordB.ID,
			petID:           fixture.petB.ID, staffID: &fixture.validStaff.ID,
		},
		{
			name:            "cross-clinic pet",
			medicalRecordID: fixture.recordA.ID,
			petID:           fixture.petB.ID, staffID: &fixture.validStaff.ID,
		},
		{
			name:            "foreign clinic staff assignment",
			medicalRecordID: fixture.recordA.ID,
			petID:           fixture.petA.ID, staffID: &fixture.foreignStaff.ID,
		},
		{
			name:            "unassigned staff",
			medicalRecordID: fixture.recordA.ID,
			petID:           fixture.petA.ID, staffID: &fixture.unassignedStaff.ID,
		},
	}
	for _, tt := range createTests {
		t.Run("create "+tt.name, func(t *testing.T) {
			got, err := service.Create(ctx, tt.medicalRecordID, &CreateVitalInput{
				ClinicID: fixture.clinicA, PetID: tt.petID,
				RecordedAt: time.Date(2026, 7, 24, 9, 0, 0, 0, time.UTC),
				StaffID:    tt.staffID, Temperature: &temperature,
			})
			require.Error(t, err)
			assert.Nil(t, got)
			assert.Zero(
				t,
				countImageVitalRelationRows(
					t,
					fixture.db,
					&model.VitalRecord{},
					"clinic_id = ?",
					fixture.clinicA,
				),
				"rejected create must write zero vital rows",
			)
		})
	}

	validStaffID := fixture.validStaff.ID
	validVital, err := service.Create(ctx, fixture.recordA.ID, &CreateVitalInput{
		ClinicID: fixture.clinicA, PetID: fixture.petA.ID,
		RecordedAt: time.Date(2026, 7, 24, 10, 0, 0, 0, time.UTC),
		StaffID:    &validStaffID, Temperature: &temperature, Notes: "before update",
	})
	require.NoError(t, err)
	require.NotNil(t, validVital)
	assert.EqualValues(
		t,
		1,
		countImageVitalRelationRows(
			t,
			fixture.db,
			&model.VitalRecord{},
			"clinic_id = ?",
			fixture.clinicA,
		),
	)

	crossClinicPetVital := seedImageVitalPollutedVital(
		t,
		fixture,
		fixture.recordA.ID,
		fixture.petB.ID,
		"cross-clinic pet sentinel",
	)
	foreignRecordVital := seedImageVitalPollutedVital(
		t,
		fixture,
		fixture.recordB.ID,
		fixture.petB.ID,
		"foreign medical record sentinel",
	)
	rejectedNotes := "must not persist"
	updateTests := []struct {
		name            string
		vitalID         uint64
		medicalRecordID uint64
		input           *UpdateVitalInput
	}{
		{
			name:    "foreign clinic staff assignment",
			vitalID: validVital.ID, medicalRecordID: fixture.recordA.ID,
			input: &UpdateVitalInput{
				StaffID: &fixture.foreignStaff.ID, Notes: &rejectedNotes,
			},
		},
		{
			name:    "unassigned staff",
			vitalID: validVital.ID, medicalRecordID: fixture.recordA.ID,
			input: &UpdateVitalInput{
				StaffID: &fixture.unassignedStaff.ID, Notes: &rejectedNotes,
			},
		},
		{
			name:    "cross-clinic pet",
			vitalID: crossClinicPetVital.ID, medicalRecordID: fixture.recordA.ID,
			input: &UpdateVitalInput{Notes: &rejectedNotes},
		},
		{
			name:    "foreign clinic medical record",
			vitalID: foreignRecordVital.ID, medicalRecordID: fixture.recordB.ID,
			input: &UpdateVitalInput{Notes: &rejectedNotes},
		},
	}
	for _, tt := range updateTests {
		t.Run("update "+tt.name, func(t *testing.T) {
			before := loadImageVitalRawVital(t, fixture.db, tt.vitalID)
			got, updateErr := service.Update(
				ctx,
				fixture.clinicA,
				tt.medicalRecordID,
				tt.vitalID,
				tt.input,
			)
			require.Error(t, updateErr)
			assert.Nil(t, got)
			after := loadImageVitalRawVital(t, fixture.db, tt.vitalID)
			assert.Equal(t, before.Notes, after.Notes)
			assert.Equal(t, before.PetID, after.PetID)
			assert.Equal(t, uint64PointerValue(before.MedicalRecordID), uint64PointerValue(after.MedicalRecordID))
			assert.Equal(t, uint64PointerValue(before.StaffID), uint64PointerValue(after.StaffID))
		})
	}

	acceptedNotes := "accepted update"
	updated, err := service.Update(
		ctx,
		fixture.clinicA,
		fixture.recordA.ID,
		validVital.ID,
		&UpdateVitalInput{StaffID: &validStaffID, Notes: &acceptedNotes},
	)
	require.NoError(t, err)
	require.NotNil(t, updated)
	require.NotNil(t, updated.StaffID)
	assert.Equal(t, acceptedNotes, updated.Notes)
	assert.Equal(t, fixture.validStaff.ID, *updated.StaffID)
}

func seedImageVitalPollutedVital(
	t *testing.T,
	fixture *imageVitalRelationDBFixture,
	medicalRecordID, petID uint64,
	notes string,
) *model.VitalRecord {
	t.Helper()
	temperature := 37.8
	vital := &model.VitalRecord{
		ClinicID: fixture.clinicA, PetID: petID,
		MedicalRecordID: &medicalRecordID,
		RecordedAt:      time.Date(2026, 7, 24, 11, 0, 0, 0, time.UTC),
		Temperature:     &temperature, Notes: notes,
	}
	require.NoError(t, fixture.db.Create(vital).Error)
	return vital
}

func loadImageVitalRawVital(t *testing.T, db *gorm.DB, id uint64) *model.VitalRecord {
	t.Helper()
	var vital model.VitalRecord
	require.NoError(t, db.Where("id = ?", id).First(&vital).Error)
	return &vital
}

func countImageVitalRelationRows(
	t *testing.T,
	db *gorm.DB,
	target any,
	query string,
	args ...any,
) int64 {
	t.Helper()
	var count int64
	require.NoError(t, db.Model(target).Where(query, args...).Count(&count).Error)
	return count
}

func uint64PointerValue(value *uint64) uint64 {
	if value == nil {
		return 0
	}
	return *value
}
