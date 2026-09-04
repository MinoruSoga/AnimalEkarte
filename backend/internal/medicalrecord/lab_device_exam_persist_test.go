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
			tx,
		),
		nil,
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
	assert.Equal(t, model.LabImportJobStatusNeedsReview, linked.Status)
	require.NotNil(t, linked.PetID)

	var count int64
	require.NoError(t, db.Model(&model.Examination{}).Where("clinic_id = ?", clinicA).Count(&count).Error)
	assert.Equal(t, int64(0), count)
}

// T001: 複数 exam_type が混在する場合、保存拒否せず種別ごとに exam を1件ずつ作る。
func TestLabDeviceExamPersister_MultipleExamTypesPersistsBoth(t *testing.T) {
	db, svc, petID := setupLabDevicePersistTest(t)
	ctx := context.Background()
	const clinicA = uint64(9701)
	examTypeNaID := seedMappedField(t, db, clinicA, "fuji_nx600", "Na-P", "Na")
	examTypeKID := seedMappedField(t, db, clinicA, "fuji_nx600", "K-P", "K")

	_, err := svc.PutWait(ctx, clinicA, 7, petID)
	require.NoError(t, err)
	got, err := svc.ReceiveFrames(ctx, clinicA, synthFujiNX600(), "NX600")
	require.NoError(t, err)
	// Na-P → ExamType-Na、K-P → ExamType-K の2種別 → 2 exam 作成・persisted
	assert.Equal(t, model.LabImportJobStatusPersisted, got.Results[0].Job.Status)

	var exams []model.Examination
	require.NoError(t, db.Where("clinic_id = ? AND job_id = ?", clinicA, got.Results[0].Job.JobID).
		Order("exam_type_id ASC").Find(&exams).Error)
	require.Len(t, exams, 2, "exam_type ごとに1件ずつ、計2件の exam が作られること")

	examTypeIDs := []uint64{exams[0].ExamTypeID, exams[1].ExamTypeID}
	assert.Contains(t, examTypeIDs, examTypeNaID)
	assert.Contains(t, examTypeIDs, examTypeKID)
	for _, exam := range exams {
		assert.Equal(t, petID, *exam.PetID)
	}

	// detach でジョブ由来の exam がすべて取り消される。
	detached, err := svc.Detach(ctx, clinicA, got.Results[0].Job.JobID)
	require.NoError(t, err)
	assert.Equal(t, model.LabImportJobStatusReceived, detached.Status)
	assert.Nil(t, detached.PetID)

	var remaining []model.Examination
	require.NoError(t, db.Where("clinic_id = ? AND job_id = ?", clinicA, got.Results[0].Job.JobID).Find(&remaining).Error)
	assert.Empty(t, remaining, "detach 後は exam が0件になること")
}

// T001: VetLab（idexx_vetlab）で2種別の項目が混在する場合、2 exam が作られ detach で両方消える。
// synthIDEXXLongFrame の WBC と RBC をそれぞれ異なる exam_type にマッピングして確認する。
func TestLabDeviceExamPersister_VetLabMultiExamPersistAndDetach(t *testing.T) {
	db, svc, petID := setupLabDevicePersistTest(t)
	ctx := context.Background()
	const clinicA = uint64(9701)
	// idexx_vetlab カタログの実在コードを2つの異なる exam_type に割り当てる。
	examTypeCBCID := seedMappedField(t, db, clinicA, "idexx_vetlab", "WBC", "WBC")
	examTypeHemaID := seedMappedField(t, db, clinicA, "idexx_vetlab", "RBC", "RBC")

	_, err := svc.PutWait(ctx, clinicA, 7, petID)
	require.NoError(t, err)
	// synthIDEXXLongFrame は WBC と RBC を含む全 11 血球項目フレームを生成する。
	got, err := svc.ReceiveFrames(ctx, clinicA, synthIDEXXLongFrame(), "VetLab")
	require.NoError(t, err)
	require.Len(t, got.Results, 1)
	assert.Equal(t, model.LabImportJobStatusPersisted, got.Results[0].Job.Status)

	var exams []model.Examination
	require.NoError(t, db.Where("clinic_id = ? AND job_id = ?", clinicA, got.Results[0].Job.JobID).
		Order("exam_type_id ASC").Find(&exams).Error)
	require.Len(t, exams, 2, "VetLab 受信: WBC→ExamType-CBC, RBC→ExamType-Hema の2件")

	gotIDs := []uint64{exams[0].ExamTypeID, exams[1].ExamTypeID}
	assert.Contains(t, gotIDs, examTypeCBCID)
	assert.Contains(t, gotIDs, examTypeHemaID)

	// detach: ジョブ由来の 2 exam がすべて消える。
	detached, err := svc.Detach(ctx, clinicA, got.Results[0].Job.JobID)
	require.NoError(t, err)
	assert.Equal(t, model.LabImportJobStatusReceived, detached.Status)
	assert.Nil(t, detached.PetID)

	var remaining []model.Examination
	require.NoError(t, db.Where("clinic_id = ? AND job_id = ?", clinicA, got.Results[0].Job.JobID).Find(&remaining).Error)
	assert.Empty(t, remaining)

	// re-attach: 同じジョブを再び付けると 2 exam が再作成される。
	linked, err := svc.Attach(ctx, clinicA, got.Results[0].Job.JobID, petID)
	require.NoError(t, err)
	assert.Equal(t, model.LabImportJobStatusPersisted, linked.Status)
	var reattached []model.Examination
	require.NoError(t, db.Where("clinic_id = ? AND job_id = ?", clinicA, got.Results[0].Job.JobID).Find(&reattached).Error)
	assert.Len(t, reattached, 2)
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
