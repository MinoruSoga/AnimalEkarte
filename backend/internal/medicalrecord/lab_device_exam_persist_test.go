package medicalrecord

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func setupLabDevicePersistTest(t *testing.T) (*gorm.DB, LabDeviceReceiveService, uint64) {
	t.Helper()
	db, _, finder := setupLabDeviceReceiveTest(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Owner{},
		&model.AnimalSpecies{},
		&model.Pet{},
		&model.MedicalRecord{},
		&model.MedicalRecordImage{},
		&model.LabImportEvent{},
		&model.ExaminationType{},
		&model.ExamTypeField{},
		&model.Examination{},
		&model.ExamResult{},
		&model.LabImportUsageReceipt{},
		&model.LabImportExamRetraction{},
		&model.LabImportExamRetractionItem{},
	))
	const clinicA, clinicB = uint64(9701), uint64(9702)
	clinicIDs := []uint64{clinicA, clinicB}
	require.NoError(t, db.Exec(
		`DELETE FROM exam_results WHERE exam_id IN (SELECT id FROM exams WHERE clinic_id IN ?)`,
		clinicIDs,
	).Error)
	require.NoError(t, db.Unscoped().Exec(`DELETE FROM exams WHERE clinic_id IN ?`, clinicIDs).Error)
	require.NoError(t, db.Exec(`DELETE FROM lab_import_exam_retraction_items WHERE clinic_id IN ?`, clinicIDs).Error)
	require.NoError(t, db.Exec(`DELETE FROM lab_import_exam_retractions WHERE clinic_id IN ?`, clinicIDs).Error)
	require.NoError(t, db.Exec(`DELETE FROM lab_import_usage_receipts WHERE clinic_id IN ?`, clinicIDs).Error)
	require.NoError(t, db.Exec(`DELETE FROM lab_import_events WHERE clinic_id IN ?`, clinicIDs).Error)
	require.NoError(t, db.Exec(`DELETE FROM exam_type_fields WHERE clinic_id IN ?`, clinicIDs).Error)
	require.NoError(t, db.Unscoped().Exec(`DELETE FROM exam_types WHERE clinic_id IN ?`, clinicIDs).Error)

	owner := makeTestOwner(t, db, clinicA, "persist-owner")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "タロウ")
	finder.pets[pet.ID] = pet

	tx := persistence.NewTransactor(db)
	masters := NewLabDeviceItemMasterService(NewLabDeviceItemMasterRepository(db))
	jobs := NewLabImportJobService(NewLabImportJobRepository(db), NewLabImportEventRepository(db))
	exams := NewLabImportExaminationService(
		NewExaminationRepository(db),
		NewLabImportDuplicateCheckerDB(db),
		NewExamTypeRepository(db),
		finder,
		stubMedicalRecordFinder{},
		tx,
	)
	revert := NewLabImportRevertService(
		db,
		tx,
		NewLabImportJobRepository(db),
		NewLabImportEventRepository(db),
		NewExaminationRepository(db),
		NewLabImportUsageReceiptRepository(db),
		NewLabImportRevertReceiptRepository(db),
		NewLabImportRetractionRepository(db),
		nil,
	)
	svc := NewLabDeviceReceiveService(
		NewLabDeviceReceiveRepository(db),
		masters,
		finder,
		tx,
		NewLabDeviceExamPersister(
			NewLabDeviceReceiveRepository(db),
			masters,
			exams,
			jobs,
			NewLabImportEventRepository(db),
			revert,
		),
	)
	return db, svc, pet.ID
}

type stubMedicalRecordFinder struct{}

func (stubMedicalRecordFinder) FindByID(context.Context, uint64, uint64) (*model.MedicalRecord, error) {
	return nil, apperrors.WrapNotFound("medical_record", "")
}

func seedMappedField(t *testing.T, db *gorm.DB, clinicID uint64, sourceType, code, fieldName string) uint64 {
	t.Helper()
	ctx := context.Background()
	masters := NewLabDeviceItemMasterService(NewLabDeviceItemMasterRepository(db))
	_, _, err := masters.EnsureDefaults(ctx, clinicID)
	require.NoError(t, err)
	et := &model.ExaminationType{ClinicID: clinicID, Name: fieldName + "-type", IsActive: true}
	require.NoError(t, db.Create(et).Error)
	field := &model.ExamTypeField{ClinicID: clinicID, ExamTypeID: et.ID, Name: fieldName, Unit: "ug/mL"}
	require.NoError(t, db.Create(field).Error)
	var row model.LabDeviceItemMaster
	require.NoError(t, db.Where("clinic_id = ? AND source_type = ? AND device_item_code = ?", clinicID, sourceType, code).
		First(&row).Error)
	_, err = masters.Update(ctx, clinicID, row.ID, UpdateLabDeviceItemMasterInput{
		DisplayName:     row.DisplayName,
		Unit:            row.Unit,
		ExamTypeFieldID: &field.ID,
		IsActive:        true,
	})
	require.NoError(t, err)
	return et.ID
}

func TestLabDeviceExamPersister_WaitPersistAndDetach(t *testing.T) {
	db, svc, petID := setupLabDevicePersistTest(t)
	ctx := context.Background()
	const clinicA = uint64(9701)
	examTypeID := seedMappedField(t, db, clinicA, "fuji_au10v", "vf-SAA", "SAA")

	_, err := svc.PutWait(ctx, clinicA, 7, petID)
	require.NoError(t, err)
	got, err := svc.ReceiveFrames(ctx, clinicA, synthFujiAU10V(), "AU10V")
	require.NoError(t, err)
	require.Len(t, got.Results, 1)
	assert.Equal(t, model.LabImportJobStatusPersisted, got.Results[0].Job.Status)
	require.NotNil(t, got.Results[0].Job.PetID)
	assert.Equal(t, petID, *got.Results[0].Job.PetID)

	var exams []model.Examination
	require.NoError(t, db.Where("clinic_id = ? AND job_id = ?", clinicA, got.Results[0].Job.JobID).Find(&exams).Error)
	require.Len(t, exams, 1)
	assert.Equal(t, examTypeID, exams[0].ExamTypeID)
	assert.Equal(t, petID, *exams[0].PetID)
	var results []model.ExamResult
	require.NoError(t, db.Where("exam_id = ?", exams[0].ID).Find(&results).Error)
	require.Len(t, results, 1)
	assert.Equal(t, "<3.75", results[0].InspectionValue)
	assert.NotNil(t, results[0].ExamTypeItemID)

	detached, err := svc.Detach(ctx, clinicA, got.Results[0].Job.JobID)
	require.NoError(t, err)
	assert.Nil(t, detached.PetID)
	assert.Equal(t, model.LabImportJobStatusReceived, detached.Status)

	var remaining []model.Examination
	require.NoError(t, db.Where("clinic_id = ? AND job_id = ?", clinicA, got.Results[0].Job.JobID).Find(&remaining).Error)
	assert.Empty(t, remaining)

	linked, err := svc.Attach(ctx, clinicA, got.Results[0].Job.JobID, petID)
	require.NoError(t, err)
	assert.Equal(t, model.LabImportJobStatusPersisted, linked.Status)
}

func TestLabDeviceExamPersister_UnmappedDoesNotWriteExam(t *testing.T) {
	db, svc, petID := setupLabDevicePersistTest(t)
	ctx := context.Background()
	const clinicA = uint64(9701)

	got, err := svc.ReceiveFrames(ctx, clinicA, synthFujiAU10V(), "AU10V")
	require.NoError(t, err)
	linked, err := svc.Attach(ctx, clinicA, got.Results[0].Job.JobID, petID)
	require.NoError(t, err)
	assert.Equal(t, model.LabImportJobStatusReceived, linked.Status)
	require.NotNil(t, linked.PetID)

	var count int64
	require.NoError(t, db.Model(&model.Examination{}).Where("clinic_id = ?", clinicA).Count(&count).Error)
	assert.Equal(t, int64(0), count)
}

func TestLabDeviceExamPersister_MultipleExamTypesNeedsReview(t *testing.T) {
	db, svc, petID := setupLabDevicePersistTest(t)
	ctx := context.Background()
	const clinicA = uint64(9701)
	seedMappedField(t, db, clinicA, "fuji_nx600", "Na-P", "Na")
	seedMappedField(t, db, clinicA, "fuji_nx600", "K-P", "K")

	_, err := svc.PutWait(ctx, clinicA, 7, petID)
	require.NoError(t, err)
	got, err := svc.ReceiveFrames(ctx, clinicA, synthFujiNX600(), "NX600")
	require.NoError(t, err)
	assert.Equal(t, model.LabImportJobStatusNeedsReview, got.Results[0].Job.Status)

	var count int64
	require.NoError(t, db.Model(&model.Examination{}).Where("clinic_id = ?", clinicA).Count(&count).Error)
	assert.Equal(t, int64(0), count)
}

func TestLabDeviceExamPersister_DetachBlockedAfterUsageReceipt(t *testing.T) {
	db, svc, petID := setupLabDevicePersistTest(t)
	ctx := context.Background()
	const clinicA = uint64(9701)
	seedMappedField(t, db, clinicA, "fuji_au10v", "vf-SAA", "SAA")
	_, err := svc.PutWait(ctx, clinicA, 7, petID)
	require.NoError(t, err)
	got, err := svc.ReceiveFrames(ctx, clinicA, synthFujiAU10V(), "AU10V")
	require.NoError(t, err)
	var exam model.Examination
	require.NoError(t, db.Where("clinic_id = ? AND job_id = ?", clinicA, got.Results[0].Job.JobID).First(&exam).Error)
	require.NoError(t, db.Create(&model.LabImportUsageReceipt{
		ClinicID: clinicA,
		JobID:    got.Results[0].Job.JobID,
		ExamID:   exam.ID,
		UseKind:  model.LabImportUsageKindExaminationDetail,
	}).Error)
	_, err = svc.Detach(ctx, clinicA, got.Results[0].Job.JobID)
	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
}
