package medicalrecord

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// LabImportUsageTracker records append-only usage / manual-mutation receipts for
// import-linked examinations. Clinical payload responses must not be returned when
// Record fails (TASK-032).
type LabImportUsageTracker interface {
	// RecordClinicalUse writes a usage receipt when exam is import-linked (job_id set).
	// No-op when job_id is nil (hand-created exam). Failure must block the clinical response.
	RecordClinicalUse(ctx context.Context, clinicID uint64, exam *model.Examination, kind model.LabImportUsageKind, actorID *uint64) error
	// RecordManualMutation writes a manual_mutation receipt inside the mutation transaction.
	// No-op when job_id is nil.
	RecordManualMutation(ctx context.Context, clinicID uint64, exam *model.Examination, actorID *uint64) error
}

type labImportUsageTracker struct {
	db       *gorm.DB
	tx       Transactor
	jobs     LabImportJobRepository
	receipts LabImportUsageReceiptRepository
}

// NewLabImportUsageTracker constructs a tracker that serializes clinical-use receipts
// with the job/exam row so compensating revert cannot race a concurrent read.
func NewLabImportUsageTracker(
	db *gorm.DB,
	tx Transactor,
	jobs LabImportJobRepository,
	receipts LabImportUsageReceiptRepository,
) LabImportUsageTracker {
	return &labImportUsageTracker{db: db, tx: tx, jobs: jobs, receipts: receipts}
}

func (t *labImportUsageTracker) RecordClinicalUse(
	ctx context.Context,
	clinicID uint64,
	exam *model.Examination,
	kind model.LabImportUsageKind,
	actorID *uint64,
) error {
	if exam == nil || exam.JobID == nil {
		return nil
	}
	if clinicID == 0 || exam.ClinicID != clinicID {
		return apperrors.WrapInternalServerError("usage receipt clinic scope mismatch")
	}
	// Manual mutation is recorded inside an existing mutation tx; clinical reads open a short tx.
	if kind == model.LabImportUsageKindManualMutation {
		return t.insertReceipt(ctx, clinicID, exam, kind, actorID)
	}
	if t.tx == nil {
		return t.insertReceipt(ctx, clinicID, exam, kind, actorID)
	}
	return t.tx.WithTx(ctx, func(txCtx context.Context) error {
		// Serialize with compensating revert: lock job then exam (must still be active + job persisted).
		job, err := t.jobs.LockByIDForUpdate(txCtx, clinicID, *exam.JobID)
		if err != nil {
			return err
		}
		if job.Status != model.LabImportJobStatusPersisted {
			return apperrors.WrapConflict("lab import job is not persisted; clinical use receipt refused")
		}
		var locked model.Examination
		if err := persistence.DBOrTx(txCtx, t.db).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("clinic_id = ? AND id = ? AND job_id = ? AND deleted_at IS NULL", clinicID, exam.ID, *exam.JobID).
			First(&locked).Error; err != nil {
			return apperrors.FromGORM(err, "exam", "")
		}
		return t.insertReceipt(txCtx, clinicID, &locked, kind, actorID)
	})
}

func (t *labImportUsageTracker) insertReceipt(
	ctx context.Context,
	clinicID uint64,
	exam *model.Examination,
	kind model.LabImportUsageKind,
	actorID *uint64,
) error {
	receipt := &model.LabImportUsageReceipt{
		ClinicID: clinicID,
		JobID:    *exam.JobID,
		ExamID:   exam.ID,
		UseKind:  kind,
		ActorID:  actorID,
	}
	if err := t.receipts.Create(ctx, receipt); err != nil {
		return apperrors.Wrap(err, "failed to record lab import usage receipt")
	}
	return nil
}

func (t *labImportUsageTracker) RecordManualMutation(
	ctx context.Context,
	clinicID uint64,
	exam *model.Examination,
	actorID *uint64,
) error {
	return t.RecordClinicalUse(ctx, clinicID, exam, model.LabImportUsageKindManualMutation, actorID)
}

// noopLabImportUsageTracker is used when receipts are not wired (tests / legacy composition).
type noopLabImportUsageTracker struct{}

func (noopLabImportUsageTracker) RecordClinicalUse(context.Context, uint64, *model.Examination, model.LabImportUsageKind, *uint64) error {
	return nil
}
func (noopLabImportUsageTracker) RecordManualMutation(context.Context, uint64, *model.Examination, *uint64) error {
	return nil
}
