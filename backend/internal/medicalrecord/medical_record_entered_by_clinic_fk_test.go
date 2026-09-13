package medicalrecord

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

const enteredBySingleColumnMigrationPath = "../../migrations/002_medical_records_entered_by_staff_fk.sql"

func setupEnteredByClinicFKTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Company{}, &model.Clinic{}, &model.Account{},
		&model.Staff{}, &model.StaffClinicAssignment{},
		&model.AnimalSpecies{}, &model.Owner{}, &model.Pet{}, &model.MedicalRecord{},
	))
	testdb.Truncate(t, db,
		"medical_records", "pets", "owners", "animal_species",
		"staff_clinic_assignments", "staffs", "accounts",
	)
	ensureVaccinationTestClinics(t, db, 1, 2)
	return db
}

func applyLegacyEnteredByClinicCompositeFK(t *testing.T, db *gorm.DB) {
	t.Helper()
	// Warm shared test DBs may already have uq_staffs_id_clinic (and dependents).
	_ = db.Exec(`
		ALTER TABLE staffs
		    ADD CONSTRAINT uq_staffs_id_clinic UNIQUE (id, clinic_id)
	`).Error

	require.NoError(t, db.Exec(`
		ALTER TABLE medical_records
		    DROP CONSTRAINT IF EXISTS medical_records_entered_by_fkey,
		    DROP CONSTRAINT IF EXISTS fk_medical_records_entered_by,
		    DROP CONSTRAINT IF EXISTS fk_medical_records_entered_by_clinic
	`).Error)

	require.NoError(t, db.Exec(`
		ALTER TABLE medical_records
		    ADD CONSTRAINT fk_medical_records_entered_by_clinic
		        FOREIGN KEY (entered_by, clinic_id)
		        REFERENCES staffs (id, clinic_id)
		        ON DELETE RESTRICT
	`).Error)
}

func applyEnteredBySingleColumnMigration(t *testing.T, db *gorm.DB) {
	t.Helper()
	raw, err := os.ReadFile(enteredBySingleColumnMigrationPath)
	require.NoError(t, err, "002 migration file must exist for user-run make migrate")
	require.NoError(t, db.Exec(`
		ALTER TABLE medical_records
		    DROP CONSTRAINT IF EXISTS fk_medical_records_entered_by_clinic,
		    DROP CONSTRAINT IF EXISTS medical_records_entered_by_fkey,
		    DROP CONSTRAINT IF EXISTS fk_medical_records_entered_by
	`).Error)
	for _, stmt := range splitSQLStatements(string(raw)) {
		trimmed := strings.TrimSpace(stmt)
		if trimmed == "" {
			continue
		}
		require.NoError(t, db.Exec(stmt).Error, "apply 002 stmt: %s", stmt)
	}
}

func newEnteredByActorTestService(t *testing.T, db *gorm.DB) MedicalRecordService {
	t.Helper()
	return NewMedicalRecordServiceWithTxAudit(
		NewMedicalRecordRepository(db),
		nil, nil, nil, nil, nil, nil,
		reservationOwnerPet(db),
		nil, nil, nil,
		persistence.NewTransactor(db),
	)
}

func seedCrossClinicEnteredByActor(t *testing.T, db *gorm.DB) (
	clinicA, clinicB uint64,
	actor *model.Staff,
	ownerB *model.Owner,
	petB *model.Pet,
) {
	t.Helper()
	clinicA, clinicB = uint64(1), uint64(2)
	actor = makeMedicalRecordListStaff(t, db, clinicA, "兼務記録者", model.StaffTypeNurse)
	require.NoError(t, db.Create(&model.StaffClinicAssignment{
		StaffID: actor.ID, ClinicID: clinicA, IsMain: true,
	}).Error)
	require.NoError(t, db.Create(&model.StaffClinicAssignment{
		StaffID: actor.ID, ClinicID: clinicB, IsMain: false,
	}).Error)
	ownerB = makeTestOwner(t, db, clinicB, "B院飼主")
	petB = makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "B院ペット")
	return clinicA, clinicB, actor, ownerB, petB
}

func TestEnteredBy_LegacyCompositeFK_RejectsHomeClinicAActorOnClinicB(t *testing.T) {
	db := setupEnteredByClinicFKTestDB(t)
	applyLegacyEnteredByClinicCompositeFK(t, db)
	_, clinicB, actor, ownerB, petB := seedCrossClinicEnteredByActor(t, db)

	record := &model.MedicalRecord{
		ClinicID:  clinicB,
		RecordNo:  "MR-ENTERED-BY-COMPOSITE-RED",
		Date:      time.Now(),
		OwnerID:   &ownerB.ID,
		PetID:     &petB.ID,
		EnteredBy: &actor.ID,
	}
	err := db.WithContext(context.Background()).Create(record).Error
	require.Error(t, err, "composite entered_by×clinic FK must reject home-A actor on clinic B")
	assert.True(t, isFKConstraintErr(err), "want SQLSTATE 23503, got %v", err)
}

func TestEnteredBy_AssignedCreator_SucceedsAfterSingleColumnFKAndActorValidation(t *testing.T) {
	db := setupEnteredByClinicFKTestDB(t)
	applyLegacyEnteredByClinicCompositeFK(t, db)
	applyEnteredBySingleColumnMigration(t, db)

	_, clinicB, actor, ownerB, petB := seedCrossClinicEnteredByActor(t, db)
	svc := newEnteredByActorTestService(t, db)

	got, err := svc.Create(context.Background(), clinicB, &CreateMedicalRecordInput{
		Date:      time.Now(),
		VisitType: model.VisitTypeRevisit,
		OwnerID:   &ownerB.ID,
		PetID:     &petB.ID,
		EnteredBy: &actor.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	require.NotNil(t, got.EnteredBy)
	assert.Equal(t, actor.ID, *got.EnteredBy)
	assert.Equal(t, clinicB, got.ClinicID)
}

func TestEnteredBy_UnassignedActor_RejectedFailClosed(t *testing.T) {
	db := setupEnteredByClinicFKTestDB(t)
	applyEnteredBySingleColumnMigration(t, db)

	const clinicA, clinicB = uint64(1), uint64(2)
	actor := makeMedicalRecordListStaff(t, db, clinicA, "未所属記録者", model.StaffTypeNurse)
	require.NoError(t, db.Create(&model.StaffClinicAssignment{
		StaffID: actor.ID, ClinicID: clinicA, IsMain: true,
	}).Error)
	ownerB := makeTestOwner(t, db, clinicB, "未所属拒否飼主")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "未所属拒否ペット")

	svc := newEnteredByActorTestService(t, db)
	_, err := svc.Create(context.Background(), clinicB, &CreateMedicalRecordInput{
		Date:      time.Now(),
		VisitType: model.VisitTypeRevisit,
		OwnerID:   &ownerB.ID,
		PetID:     &petB.ID,
		EnteredBy: &actor.ID,
	})
	require.Error(t, err)
	assert.True(t,
		errors.Is(err, apperrors.ErrForbidden) ||
			apperrors.IsNotFound(err) ||
			apperrors.IsInvalidInput(err),
		"unassigned actor must fail closed without FK leak: %v", err)

	var count int64
	require.NoError(t, db.Model(&model.MedicalRecord{}).
		Where("clinic_id = ? AND entered_by = ?", clinicB, actor.ID).
		Count(&count).Error)
	assert.Zero(t, count, "no partial medical_records write")
}

func TestEnteredBy_SystemAdminWithoutAssignment_Allowed(t *testing.T) {
	db := setupEnteredByClinicFKTestDB(t)
	applyEnteredBySingleColumnMigration(t, db)

	const clinicB = uint64(2)
	account := &model.Account{
		Email:         "sysadmin-entered-by@example.test",
		PasswordHash:  "x",
		IsActive:      true,
		IsSystemAdmin: true,
	}
	require.NoError(t, db.Create(account).Error)
	actor := &model.Staff{
		ClinicID:  1,
		AccountID: &account.ID,
		Name:      "系統管理者記録者",
		StaffType: model.StaffTypeDoctor,
		IsActive:  true,
	}
	require.NoError(t, db.Create(actor).Error)
	ownerB := makeTestOwner(t, db, clinicB, "管理者飼主")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "管理者ペット")

	svc := newEnteredByActorTestService(t, db)
	got, err := svc.Create(context.Background(), clinicB, &CreateMedicalRecordInput{
		Date:                 time.Now(),
		VisitType:            model.VisitTypeRevisit,
		OwnerID:              &ownerB.ID,
		PetID:                &petB.ID,
		EnteredBy:            &actor.ID,
		EnteredBySystemAdmin: true,
	})
	require.NoError(t, err)
	require.NotNil(t, got.EnteredBy)
	assert.Equal(t, actor.ID, *got.EnteredBy)
}

func TestEnteredBy_AutoCreateNilActor_SkipsAssignmentGate(t *testing.T) {
	db := setupEnteredByClinicFKTestDB(t)
	applyEnteredBySingleColumnMigration(t, db)

	const clinicB = uint64(2)
	ownerB := makeTestOwner(t, db, clinicB, "自動作成飼主")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "自動作成ペット")
	svc := newEnteredByActorTestService(t, db)
	status := model.MedicalRecordStatusDraft
	got, err := svc.Create(context.Background(), clinicB, &CreateMedicalRecordInput{
		Date:      time.Now(),
		VisitType: model.VisitTypeRevisit,
		OwnerID:   &ownerB.ID,
		PetID:     &petB.ID,
		Status:    &status,
	})
	require.NoError(t, err)
	assert.Nil(t, got.EnteredBy)
}

func TestEnteredBy_HistoryRemainsReadableAfterAssignmentRemoved(t *testing.T) {
	db := setupEnteredByClinicFKTestDB(t)
	applyEnteredBySingleColumnMigration(t, db)

	_, clinicB, actor, ownerB, petB := seedCrossClinicEnteredByActor(t, db)
	repo := NewMedicalRecordRepository(db)
	svc := newEnteredByActorTestService(t, db)

	created, err := svc.Create(context.Background(), clinicB, &CreateMedicalRecordInput{
		Date:      time.Now(),
		VisitType: model.VisitTypeRevisit,
		OwnerID:   &ownerB.ID,
		PetID:     &petB.ID,
		EnteredBy: &actor.ID,
	})
	require.NoError(t, err)

	require.NoError(t, db.Where("staff_id = ? AND clinic_id = ?", actor.ID, clinicB).
		Delete(&model.StaffClinicAssignment{}).Error)

	got, err := repo.FindByID(context.Background(), clinicB, created.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, created.ID, got.ID)
	require.NotNil(t, got.EnteredBy)
	assert.Equal(t, actor.ID, *got.EnteredBy)

	list, total, err := repo.FindAll(context.Background(), []uint64{clinicB}, MedicalRecordListFilters{}, 1, 20)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(1))
	found := false
	for _, row := range list {
		if row.ID == created.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "list must keep history after assignment removal")
}

func TestEnteredBy_SameClinicCreate_StillWorks(t *testing.T) {
	db := setupEnteredByClinicFKTestDB(t)
	applyEnteredBySingleColumnMigration(t, db)

	const clinicA = uint64(1)
	actor := makeMedicalRecordListStaff(t, db, clinicA, "自院記録者", model.StaffTypeNurse)
	require.NoError(t, db.Create(&model.StaffClinicAssignment{
		StaffID: actor.ID, ClinicID: clinicA, IsMain: true,
	}).Error)
	owner := makeTestOwner(t, db, clinicA, "自院飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "自院ペット")
	svc := newEnteredByActorTestService(t, db)
	got, err := svc.Create(context.Background(), clinicA, &CreateMedicalRecordInput{
		Date:      time.Now(),
		VisitType: model.VisitTypeRevisit,
		OwnerID:   &owner.ID,
		PetID:     &pet.ID,
		EnteredBy: &actor.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, got.EnteredBy)
	assert.Equal(t, actor.ID, *got.EnteredBy)
}

func TestEnteredBy_MigrationFileExists_NextAfter001(t *testing.T) {
	path := filepath.Clean(enteredBySingleColumnMigrationPath)
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	body := string(raw)
	assert.Contains(t, body, "fk_medical_records_entered_by_clinic")
	assert.Contains(t, body, "fk_medical_records_entered_by")
	assert.Contains(t, body, "REFERENCES staffs (id)")
	assert.NotContains(t, body, "REFERENCES staffs (id, clinic_id)")
}
