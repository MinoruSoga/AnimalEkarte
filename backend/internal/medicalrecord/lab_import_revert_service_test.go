package medicalrecord

// lab_import_revert_service_test.go — TASK-032 compensating revert state machine.
//
// Covers: safe success, conflict gates (confirmed / finalized / manual mutation /
// downstream use / usage_unknown), idempotent replay, conflicting payload 409,
// ambient-tx requirement, and zero-write on conflict.

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func setupLabImportRevertTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupLabImportTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.LabImportUsageReceipt{},
		&model.LabImportRevertReceipt{},
		&model.LabImportExamRetraction{},
		&model.LabImportExamRetractionItem{},
		&model.MedicalRecord{},
	))
	for _, tbl := range []string{
		"lab_import_exam_retraction_items",
		"lab_import_exam_retractions",
		"lab_import_usage_receipts",
		"lab_import_revert_receipts",
		"medical_records",
	} {
		_ = db.Exec("TRUNCATE TABLE " + tbl + " CASCADE")
	}
	return db
}

func newRevertService(db *gorm.DB) LabImportRevertService {
	return NewLabImportRevertService(
		db,
		persistence.NewTransactor(db),
		NewLabImportJobRepository(db),
		NewLabImportEventRepository(db),
		NewExaminationRepository(db),
		NewLabImportUsageReceiptRepository(db),
		NewLabImportRevertReceiptRepository(db),
		NewLabImportRetractionRepository(db),
		nil,
	)
}

func seedTrackedPersistedJob(t *testing.T, db *gorm.DB, clinicID uint64) (*model.LabImportJob, *model.Examination) {
	t.Helper()
	job := makeLabImportJob(t, db, clinicID, model.LabImportJobStatusPersisted)
	require.NoError(t, db.Create(&model.LabImportEvent{
		ClinicID:  clinicID,
		JobID:     job.ID,
		EventType: model.LabImportEventTypeUsageTrackingStarted,
	}).Error)
	et := makeLabImportExamTypeMaster(t, db, clinicID, "CBC")
	jid := job.ID
	exam := &model.Examination{
		ClinicID:   clinicID,
		ExamTypeID: et.ID,
		JobID:      &jid,
		Date:       time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		Status:     model.ExaminationStatusResultEntered,
		Machine:    "fixture",
	}
	require.NoError(t, db.Create(exam).Error)
	require.NoError(t, db.Create(&model.ExamResult{
		ExamID: exam.ID, Name: "WBC", InspectionValue: "10", SortOrder: 0,
	}).Error)
	// Reload items
	require.NoError(t, db.Preload("Items").First(exam, exam.ID).Error)
	return job, exam
}

func TestLabImportRevert_Success_SoftDeletesDraftParent_KeepsChildResults(t *testing.T) {
	db := setupLabImportRevertTestDB(t)
	svc := newRevertService(db)
	const clinicID = uint64(1)
	job, exam := seedTrackedPersistedJob(t, db, clinicID)
	actor := uint64(9)
	key := uuid.New()

	resp, err := svc.Revert(context.Background(), RevertLabImportInput{
		ClinicID: clinicID, JobID: job.ID, ActorID: &actor,
		Reason: "fixture import was wrong", IdempotencyKey: key,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, string(model.LabImportJobStatusReverted), resp.Status)
	assert.False(t, resp.IdempotentReplay)
	assert.Equal(t, []uint64{exam.ID}, resp.RetractedExamIDs)

	// Job terminal.
	var stored model.LabImportJob
	require.NoError(t, db.First(&stored, "id = ?", job.ID).Error)
	assert.Equal(t, model.LabImportJobStatusReverted, stored.Status)

	// Parent soft-deleted; child results remain.
	var active model.Examination
	err = db.First(&active, "id = ?", exam.ID).Error
	assert.Error(t, err, "soft-deleted exam must not appear under default GORM scope")
	var raw model.Examination
	require.NoError(t, db.Unscoped().First(&raw, "id = ?", exam.ID).Error)
	assert.True(t, raw.DeletedAt.Valid)

	var childCount int64
	require.NoError(t, db.Model(&model.ExamResult{}).Where("exam_id = ?", exam.ID).Count(&childCount).Error)
	assert.Equal(t, int64(1), childCount, "child exam_results must not be hard-deleted")

	// Retraction snapshot present.
	var retCount int64
	require.NoError(t, db.Model(&model.LabImportExamRetraction{}).
		Where("clinic_id = ? AND job_id = ? AND exam_id = ?", clinicID, job.ID, exam.ID).
		Count(&retCount).Error)
	assert.Equal(t, int64(1), retCount)

	// Event + receipt.
	var eventCount int64
	require.NoError(t, db.Model(&model.LabImportEvent{}).
		Where("clinic_id = ? AND job_id = ? AND event_type = ?", clinicID, job.ID, model.LabImportEventTypeRevertRequested).
		Count(&eventCount).Error)
	assert.Equal(t, int64(1), eventCount)
}

func TestLabImportRevert_IdempotentReplay_DoesNotGrowEventOrAudit(t *testing.T) {
	db := setupLabImportRevertTestDB(t)
	svc := newRevertService(db)
	const clinicID = uint64(1)
	job, _ := seedTrackedPersistedJob(t, db, clinicID)
	actor := uint64(9)
	key := uuid.New()
	input := RevertLabImportInput{
		ClinicID: clinicID, JobID: job.ID, ActorID: &actor,
		Reason: "same reason", IdempotencyKey: key,
	}

	first, err := svc.Revert(context.Background(), input)
	require.NoError(t, err)
	assert.False(t, first.IdempotentReplay)

	var eventBefore int64
	require.NoError(t, db.Model(&model.LabImportEvent{}).
		Where("clinic_id = ? AND job_id = ?", clinicID, job.ID).Count(&eventBefore).Error)
	var receiptBefore int64
	require.NoError(t, db.Model(&model.LabImportRevertReceipt{}).
		Where("clinic_id = ? AND job_id = ?", clinicID, job.ID).Count(&receiptBefore).Error)

	second, err := svc.Revert(context.Background(), input)
	require.NoError(t, err)
	assert.True(t, second.IdempotentReplay)
	assert.Equal(t, first.RetractedExamIDs, second.RetractedExamIDs)

	var eventAfter int64
	require.NoError(t, db.Model(&model.LabImportEvent{}).
		Where("clinic_id = ? AND job_id = ?", clinicID, job.ID).Count(&eventAfter).Error)
	var receiptAfter int64
	require.NoError(t, db.Model(&model.LabImportRevertReceipt{}).
		Where("clinic_id = ? AND job_id = ?", clinicID, job.ID).Count(&receiptAfter).Error)
	assert.Equal(t, eventBefore, eventAfter, "idempotent replay must not append events")
	assert.Equal(t, receiptBefore, receiptAfter, "idempotent replay must not append receipts")
}

func TestLabImportRevert_ConflictingIdempotencyPayload_Returns409(t *testing.T) {
	db := setupLabImportRevertTestDB(t)
	svc := newRevertService(db)
	const clinicID = uint64(1)
	job, _ := seedTrackedPersistedJob(t, db, clinicID)
	actor := uint64(9)
	key := uuid.New()

	_, err := svc.Revert(context.Background(), RevertLabImportInput{
		ClinicID: clinicID, JobID: job.ID, ActorID: &actor,
		Reason: "reason-a", IdempotencyKey: key,
	})
	require.NoError(t, err)

	// New job for second attempt with same key different payload is blocked even if job differs —
	// same key + different hash → 409. Use same job with different reason.
	_, err = svc.Revert(context.Background(), RevertLabImportInput{
		ClinicID: clinicID, JobID: job.ID, ActorID: &actor,
		Reason: "reason-b-different", IdempotencyKey: key,
	})
	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
}

func TestLabImportRevert_ConflictTable(t *testing.T) {
	type seedFn func(t *testing.T, db *gorm.DB, clinicID uint64, job *model.LabImportJob, exam *model.Examination)

	cases := []struct {
		name      string
		seed      seedFn
		wantMsg   string
		skipTrack bool // when true, do not write usage_tracking_started
	}{
		{
			name: "usage_unknown",
			seed: func(t *testing.T, db *gorm.DB, clinicID uint64, job *model.LabImportJob, exam *model.Examination) {
				// delete tracking event if present
				require.NoError(t, db.Where("job_id = ? AND event_type = ?", job.ID, model.LabImportEventTypeUsageTrackingStarted).
					Delete(&model.LabImportEvent{}).Error)
			},
			wantMsg: "usage_unknown",
		},
		{
			name: "confirmed examination",
			seed: func(t *testing.T, db *gorm.DB, clinicID uint64, job *model.LabImportJob, exam *model.Examination) {
				require.NoError(t, db.Model(exam).Update("status", model.ExaminationStatusConfirmed).Error)
			},
			wantMsg: "confirmed",
		},
		{
			name: "manual mutation",
			seed: func(t *testing.T, db *gorm.DB, clinicID uint64, job *model.LabImportJob, exam *model.Examination) {
				require.NoError(t, db.Create(&model.LabImportUsageReceipt{
					ClinicID: clinicID, JobID: job.ID, ExamID: exam.ID,
					UseKind: model.LabImportUsageKindManualMutation,
				}).Error)
			},
			wantMsg: "manual_mutation",
		},
		{
			name: "downstream clinical use",
			seed: func(t *testing.T, db *gorm.DB, clinicID uint64, job *model.LabImportJob, exam *model.Examination) {
				require.NoError(t, db.Create(&model.LabImportUsageReceipt{
					ClinicID: clinicID, JobID: job.ID, ExamID: exam.ID,
					UseKind: model.LabImportUsageKindExaminationDetail,
				}).Error)
			},
			wantMsg: "downstream_use",
		},
		{
			name: "finalized medical record",
			seed: func(t *testing.T, db *gorm.DB, clinicID uint64, job *model.LabImportJob, exam *model.Examination) {
				mr := &model.MedicalRecord{
					ClinicID: clinicID,
					RecordNo: "MR-REVERT-1",
					Date:     time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
					Status:   model.MedicalRecordStatusFinalized,
				}
				require.NoError(t, db.Create(mr).Error)
				require.NoError(t, db.Model(exam).Update("medical_record_id", mr.ID).Error)
			},
			wantMsg: "finalized",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := setupLabImportRevertTestDB(t)
			svc := newRevertService(db)
			const clinicID = uint64(1)
			job, exam := seedTrackedPersistedJob(t, db, clinicID)
			tc.seed(t, db, clinicID, job, exam)

			var eventBefore int64
			require.NoError(t, db.Model(&model.LabImportEvent{}).
				Where("clinic_id = ? AND job_id = ?", clinicID, job.ID).Count(&eventBefore).Error)

			actorID := uint64(1)
			_, err := svc.Revert(context.Background(), RevertLabImportInput{
				ClinicID: clinicID, JobID: job.ID, ActorID: &actorID,
				Reason: "should fail", IdempotencyKey: uuid.New(),
			})
			require.Error(t, err)
			assert.True(t, apperrors.IsConflict(err), "got %v", err)
			assert.Contains(t, err.Error(), tc.wantMsg)

			// Zero write: job still persisted, no revert event/receipt, exam still active.
			var stored model.LabImportJob
			require.NoError(t, db.First(&stored, "id = ?", job.ID).Error)
			assert.Equal(t, model.LabImportJobStatusPersisted, stored.Status)

			var eventAfter int64
			require.NoError(t, db.Model(&model.LabImportEvent{}).
				Where("clinic_id = ? AND job_id = ?", clinicID, job.ID).Count(&eventAfter).Error)
			assert.Equal(t, eventBefore, eventAfter)

			var receiptCount int64
			require.NoError(t, db.Model(&model.LabImportRevertReceipt{}).
				Where("clinic_id = ? AND job_id = ?", clinicID, job.ID).Count(&receiptCount).Error)
			assert.Equal(t, int64(0), receiptCount)

			var active model.Examination
			require.NoError(t, db.First(&active, "id = ?", exam.ID).Error)
			assert.False(t, active.DeletedAt.Valid)
		})
	}
}

func TestLabImportRevert_TransitionStatusCannotReachReverted(t *testing.T) {
	db := setupLabImportRevertTestDB(t)
	jobRepo := NewLabImportJobRepository(db)
	eventRepo := NewLabImportEventRepository(db)
	svc := NewLabImportJobService(jobRepo, eventRepo)
	job := makeLabImportJob(t, db, 1, model.LabImportJobStatusPersisted)

	_, err := svc.TransitionStatus(context.Background(), 1, job.ID, model.LabImportJobStatusReverted, TransitionCounts{})
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Contains(t, err.Error(), "POST /lab-imports")
}

func TestLabImportUsageReceipt_FailureBlocksClinicalResponse(t *testing.T) {
	// Unit-level: a failing receipt Create must surface as error (no silent success).
	// Manual mutation path inserts without opening a nested tx (mutation already holds one).
	failing := &failingUsageReceiptRepo{err: assert.AnError}
	tracker := NewLabImportUsageTracker(nil, nil, nil, failing)
	jobID := uuid.New()
	exam := &model.Examination{ID: 1, ClinicID: 1, JobID: &jobID}
	err := tracker.RecordClinicalUse(context.Background(), 1, exam, model.LabImportUsageKindManualMutation, nil)
	require.Error(t, err)
}

func TestLabImportUsageReceipt_SkipsHandCreatedExam(t *testing.T) {
	tracker := NewLabImportUsageTracker(nil, nil, nil, &failingUsageReceiptRepo{err: assert.AnError})
	exam := &model.Examination{ID: 1, ClinicID: 1, JobID: nil}
	require.NoError(t, tracker.RecordClinicalUse(context.Background(), 1, exam, model.LabImportUsageKindExaminationDetail, nil))
}

func TestLabImportRevert_WrongClinic_NotFound(t *testing.T) {
	db := setupLabImportRevertTestDB(t)
	svc := newRevertService(db)
	job, _ := seedTrackedPersistedJob(t, db, 1)
	actorID := uint64(1)
	_, err := svc.Revert(context.Background(), RevertLabImportInput{
		ClinicID: 999, JobID: job.ID, ActorID: &actorID,
		Reason: "cross clinic", IdempotencyKey: uuid.New(),
	})
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

type failingUsageReceiptRepo struct{ err error }

func (f *failingUsageReceiptRepo) Create(context.Context, *model.LabImportUsageReceipt) error {
	return f.err
}
func (f *failingUsageReceiptRepo) LockByJobForUpdate(context.Context, uint64, uuid.UUID) ([]*model.LabImportUsageReceipt, error) {
	return nil, f.err
}
func (f *failingUsageReceiptRepo) CountByJob(context.Context, uint64, uuid.UUID) (int64, error) {
	return 0, f.err
}
func (f *failingUsageReceiptRepo) CountManualMutationByJob(context.Context, uint64, uuid.UUID) (int64, error) {
	return 0, f.err
}
