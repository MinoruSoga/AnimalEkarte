package medicalrecord

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

const maxLabImportRevertReasonRunes = 500

// RevertLabImportInput is the service input for compensating lab-import revert.
// Distinct from examination unconfirm (DEC-57 / TASK-032).
type RevertLabImportInput struct {
	ClinicID       uint64
	JobID          uuid.UUID
	ActorID        *uint64
	Reason         string
	IdempotencyKey uuid.UUID
}

// LabImportRevertService owns the persisted → reverted terminal compensation state machine.
type LabImportRevertService interface {
	Revert(ctx context.Context, input RevertLabImportInput) (*model.LabImportRevertResponse, error)
	// DetachDeviceJob retracts persisted device exams and returns the job to received.
	// It does not write reverted. Confirmed / usage receipts / finalized charts are 409.
	DetachDeviceJob(ctx context.Context, clinicID uint64, jobID uuid.UUID) error
}

type labImportRevertService struct {
	db          *gorm.DB
	tx          Transactor
	jobs        LabImportJobRepository
	events      LabImportEventRepository
	exams       ExaminationRepository
	usage       LabImportUsageReceiptRepository
	reverts     LabImportRevertReceiptRepository
	retractions LabImportRetractionRepository
	audit       LabAuditLogger // nil-safe
}

// NewLabImportRevertService constructs the compensating-revert service.
func NewLabImportRevertService(
	db *gorm.DB,
	tx Transactor,
	jobs LabImportJobRepository,
	events LabImportEventRepository,
	exams ExaminationRepository,
	usage LabImportUsageReceiptRepository,
	reverts LabImportRevertReceiptRepository,
	retractions LabImportRetractionRepository,
	audit LabAuditLogger,
) LabImportRevertService {
	return &labImportRevertService{
		db:          db,
		tx:          tx,
		jobs:        jobs,
		events:      events,
		exams:       exams,
		usage:       usage,
		reverts:     reverts,
		retractions: retractions,
		audit:       audit,
	}
}

func (s *labImportRevertService) Revert(ctx context.Context, input RevertLabImportInput) (*model.LabImportRevertResponse, error) {
	if input.ClinicID == 0 {
		return nil, apperrors.WrapInvalidInput("clinic_id is required")
	}
	if input.JobID == uuid.Nil {
		return nil, apperrors.WrapInvalidInput("job_id is required")
	}
	if input.IdempotencyKey == uuid.Nil {
		return nil, apperrors.WrapInvalidInput("Idempotency-Key must be a UUID")
	}
	reason := strings.TrimSpace(input.Reason)
	if reason == "" {
		return nil, apperrors.WrapInvalidInput("reason is required")
	}
	if utf8.RuneCountInString(reason) > maxLabImportRevertReasonRunes {
		return nil, apperrors.WrapInvalidInput("reason must be 500 characters or fewer")
	}

	requestHash := labImportRevertRequestHash(input.JobID, reason)
	var response *model.LabImportRevertResponse

	err := s.tx.WithTx(ctx, func(txCtx context.Context) error {
		// 1. Lock job status-independently.
		job, err := s.jobs.LockByIDForUpdate(txCtx, input.ClinicID, input.JobID)
		if err != nil {
			return err
		}

		// 2. Idempotency lookup under job lock.
		existing, findErr := s.reverts.LockByIdempotencyKey(txCtx, input.ClinicID, input.IdempotencyKey)
		if findErr == nil && existing != nil {
			if existing.JobID != input.JobID || existing.RequestHash != requestHash {
				return apperrors.WrapConflict("idempotency key already used with a different payload")
			}
			// same payload replay — do not append event/audit
			ids, parseErr := parseRetractedExamIDs(existing.RetractedExamIDs)
			if parseErr != nil {
				return apperrors.WrapInternalServerError("corrupt revert receipt")
			}
			response = &model.LabImportRevertResponse{
				JobID:            existing.JobID,
				Status:           existing.ResultStatus,
				RetractedExamIDs: ids,
				IdempotentReplay: true,
			}
			return nil
		}
		if findErr != nil && !apperrors.IsNotFound(findErr) {
			return findErr
		}

		// 3. Status gate: only persisted may revert. Already reverted without receipt is conflict.
		if job.Status == model.LabImportJobStatusReverted {
			return apperrors.WrapConflict("lab import job is already reverted")
		}
		if job.Status != model.LabImportJobStatusPersisted {
			return apperrors.WrapConflict(
				fmt.Sprintf("lab import job status %s cannot be reverted; only persisted jobs are eligible", job.Status),
			)
		}

		// 4. usage_tracking_started marker required (legacy / unknown → 409, zero write).
		hasTracking, err := s.events.HasEventType(txCtx, input.ClinicID, input.JobID, model.LabImportEventTypeUsageTrackingStarted)
		if err != nil {
			return err
		}
		if !hasTracking {
			return apperrors.WrapConflict("usage_unknown: lab import job has no usage_tracking_started marker; revert is refused")
		}

		// 5. Lock linked active exams in ID order.
		exams, err := s.lockLinkedExamsByJob(txCtx, input.ClinicID, input.JobID)
		if err != nil {
			return err
		}

		// 6. Lock usage receipts for the job.
		receipts, err := s.usage.LockByJobForUpdate(txCtx, input.ClinicID, input.JobID)
		if err != nil {
			return err
		}

		// 7. Conflict predicates — any hit is 409 with zero write (rollback).
		if err := s.assertRevertSafe(txCtx, input.ClinicID, input.JobID, exams, receipts); err != nil {
			return err
		}

		// 8. Retraction snapshots + conditional soft delete (parent only; child results kept).
		retractedIDs := make([]uint64, 0, len(exams))
		for _, exam := range exams {
			if err := s.retractExam(txCtx, input, reason, exam); err != nil {
				return err
			}
			retractedIDs = append(retractedIDs, exam.ID)
		}

		// 9. CAS job persisted → reverted.
		now := time.Now()
		affected, err := s.jobs.CompareAndSetStatus(
			txCtx, input.ClinicID, input.JobID,
			model.LabImportJobStatusPersisted, model.LabImportJobStatusReverted, &now,
		)
		if err != nil {
			return err
		}
		if affected != 1 {
			return apperrors.WrapConflict("lab import job status changed concurrently; revert refused")
		}

		// 10. Job event.
		from := model.LabImportJobStatusPersisted
		to := model.LabImportJobStatusReverted
		if err := s.events.Create(txCtx, &model.LabImportEvent{
			ClinicID:   input.ClinicID,
			JobID:      input.JobID,
			EventType:  model.LabImportEventTypeRevertRequested,
			FromStatus: &from,
			ToStatus:   &to,
		}); err != nil {
			return err
		}

		// 11. Revert receipt.
		idsJSON := MustJSON(retractedIDs)
		if err := s.reverts.Create(txCtx, &model.LabImportRevertReceipt{
			ClinicID:         input.ClinicID,
			JobID:            input.JobID,
			IdempotencyKey:   input.IdempotencyKey,
			RequestHash:      requestHash,
			Reason:           reason,
			ActorID:          input.ActorID,
			ResultStatus:     string(model.LabImportJobStatusReverted),
			RetractedExamIDs: idsJSON,
		}); err != nil {
			return err
		}

		// 12. Application audit — revert is not a commit success; record via dedicated path.
		if s.audit != nil {
			s.audit.LogRevertSucceeded(txCtx, input.ClinicID, input.ActorID, input.JobID, len(retractedIDs))
		}

		response = &model.LabImportRevertResponse{
			JobID:            input.JobID,
			Status:           string(model.LabImportJobStatusReverted),
			RetractedExamIDs: retractedIDs,
			IdempotentReplay: false,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (s *labImportRevertService) DetachDeviceJob(ctx context.Context, clinicID uint64, jobID uuid.UUID) error {
	if clinicID == 0 || jobID == uuid.Nil {
		return apperrors.WrapInvalidInput("job_id is required")
	}
	run := func(txCtx context.Context) error {
		job, err := s.jobs.LockByIDForUpdate(txCtx, clinicID, jobID)
		if err != nil {
			return err
		}
		if !isLabDeviceSourceType(string(job.SourceType)) {
			return apperrors.WrapInvalidInput("job is not a lab device source")
		}
		if job.Status != model.LabImportJobStatusPersisted {
			return nil
		}
		hasTracking, err := s.events.HasEventType(txCtx, clinicID, jobID, model.LabImportEventTypeUsageTrackingStarted)
		if err != nil {
			return err
		}
		if !hasTracking {
			return apperrors.WrapConflict("usage_unknown: lab import job has no usage_tracking_started marker; revert is refused")
		}
		exams, err := s.lockLinkedExamsByJob(txCtx, clinicID, jobID)
		if err != nil {
			return err
		}
		receipts, err := s.usage.LockByJobForUpdate(txCtx, clinicID, jobID)
		if err != nil {
			return err
		}
		if err := s.assertRevertSafe(txCtx, clinicID, jobID, exams, receipts); err != nil {
			return err
		}
		input := RevertLabImportInput{ClinicID: clinicID, JobID: jobID}
		for _, exam := range exams {
			if err := s.retractExam(txCtx, input, "検査受信の取り消し", exam); err != nil {
				return err
			}
		}
		affected, err := s.jobs.CompareAndSetStatus(
			txCtx, clinicID, jobID,
			model.LabImportJobStatusPersisted, model.LabImportJobStatusReceived, nil,
		)
		if err != nil {
			return err
		}
		if affected != 1 {
			return apperrors.WrapConflict("lab import job status changed concurrently; revert refused")
		}
		job.Status = model.LabImportJobStatusReceived
		job.PetID = nil
		job.FinishedAt = nil
		job.PersistedCount = 0
		return s.jobs.Update(txCtx, job)
	}
	if persistence.TxFromContext(ctx) != nil {
		return run(ctx)
	}
	return s.tx.WithTx(ctx, run)
}

func (s *labImportRevertService) lockLinkedExamsByJob(
	ctx context.Context,
	clinicID uint64,
	jobID uuid.UUID,
) ([]*model.Examination, error) {
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError("lab import exam lock requires an ambient transaction")
	}
	var exams []*model.Examination
	err := persistence.DBOrTx(ctx, s.db).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("clinic_id = ? AND job_id = ? AND deleted_at IS NULL", clinicID, jobID).
		Order("id ASC").
		Find(&exams).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "exam", jobID.String())
	}
	// Preload items after lock (same tx). Correlate grandchild exam_results via parent exams clinic.
	for i := range exams {
		var items []model.ExamResult
		if err := persistence.DBOrTx(ctx, s.db).
			Where(`exam_id = ? AND EXISTS (
				SELECT 1 FROM exams parent_exam
				WHERE parent_exam.id = exam_results.exam_id
				  AND parent_exam.clinic_id = ?
				  AND parent_exam.deleted_at IS NULL
			)`, exams[i].ID, clinicID).
			Order("sort_order ASC, id ASC").
			Find(&items).Error; err != nil {
			return nil, apperrors.FromGORM(err, "exam_item", fmt.Sprintf("%d", exams[i].ID))
		}
		exams[i].Items = items
	}
	return exams, nil
}

func (s *labImportRevertService) assertRevertSafe(
	ctx context.Context,
	clinicID uint64,
	jobID uuid.UUID,
	exams []*model.Examination,
	receipts []*model.LabImportUsageReceipt,
) error {
	// Downstream clinical use / manual mutation via receipts.
	for _, r := range receipts {
		switch r.UseKind {
		case model.LabImportUsageKindManualMutation:
			return apperrors.WrapConflict("manual_mutation: import-linked examination was manually mutated; revert refused")
		case model.LabImportUsageKindExaminationDetail,
			model.LabImportUsageKindExaminationItems,
			model.LabImportUsageKindLabReport,
			model.LabImportUsageKindPrintSnapshot:
			return apperrors.WrapConflict("downstream_use: import-linked examination was clinically used; revert refused")
		default:
			return apperrors.WrapConflict("usage_unknown: unrecognized usage receipt; revert refused")
		}
	}

	for _, exam := range exams {
		if exam.Status == model.ExaminationStatusConfirmed {
			return apperrors.WrapConflict("confirmed examination cannot be reverted via lab import compensation")
		}
		// Finalized medical record (FOR SHARE under ambient tx to close finalize TOCTOU).
		if exam.MedicalRecordID != nil {
			var mr model.MedicalRecord
			err := persistence.DBOrTx(ctx, s.db).
				Clauses(clause.Locking{Strength: "SHARE"}).
				Where("clinic_id = ? AND id = ?", clinicID, *exam.MedicalRecordID).
				First(&mr).Error
			if err == nil && mr.Status == model.MedicalRecordStatusFinalized {
				return apperrors.WrapConflict("finalized medical record blocks lab import revert")
			}
			if err != nil && !apperrors.IsNotFound(err) && err != gorm.ErrRecordNotFound {
				return apperrors.FromGORM(err, "medical_record", fmt.Sprintf("%d", *exam.MedicalRecordID))
			}
		}
		// Durable relation: medical_record_images.exam_id
		imgCount, err := CountMedicalRecordImagesByExam(ctx, s.db, clinicID, exam.ID)
		if err != nil {
			return err
		}
		if imgCount > 0 {
			return apperrors.WrapConflict("downstream_use: medical_record_images reference blocks lab import revert")
		}
		// Fail-closed relation checks: exam type clinic, pet-owner clinic, doctor assignment.
		if err := s.assertExamRelations(ctx, clinicID, exam); err != nil {
			return err
		}
	}
	_ = jobID
	return nil
}

func (s *labImportRevertService) assertExamRelations(ctx context.Context, clinicID uint64, exam *model.Examination) error {
	var examType model.ExaminationType
	if err := persistence.DBOrTx(ctx, s.db).
		Where("clinic_id = ? AND id = ? AND deleted_at IS NULL", clinicID, exam.ExamTypeID).
		First(&examType).Error; err != nil {
		if err == gorm.ErrRecordNotFound || apperrors.IsNotFound(err) {
			return apperrors.WrapConflict("exam type relation invalid or cross-clinic; revert refused")
		}
		return apperrors.FromGORM(err, "exam_type", fmt.Sprintf("%d", exam.ExamTypeID))
	}
	if exam.PetID != nil {
		var pet model.Pet
		if err := persistence.DBOrTx(ctx, s.db).
			Where("clinic_id = ? AND id = ? AND deleted_at IS NULL", clinicID, *exam.PetID).
			First(&pet).Error; err != nil {
			if err == gorm.ErrRecordNotFound || apperrors.IsNotFound(err) {
				return apperrors.WrapConflict("pet relation invalid or cross-clinic; revert refused")
			}
			return apperrors.FromGORM(err, "pet", fmt.Sprintf("%d", *exam.PetID))
		}
	}
	if exam.DoctorID != nil {
		var staff model.Staff
		if err := persistence.DBOrTx(ctx, s.db).
			Where("clinic_id = ? AND id = ? AND deleted_at IS NULL AND is_active = TRUE", clinicID, *exam.DoctorID).
			First(&staff).Error; err != nil {
			if err == gorm.ErrRecordNotFound || apperrors.IsNotFound(err) {
				return apperrors.WrapConflict("doctor assignment invalid or inactive; revert refused")
			}
			return apperrors.FromGORM(err, "staff", fmt.Sprintf("%d", *exam.DoctorID))
		}
	}
	return nil
}

func (s *labImportRevertService) retractExam(
	ctx context.Context,
	input RevertLabImportInput,
	reason string,
	exam *model.Examination,
) error {
	parentSnap := MustJSON(map[string]any{
		"id":                exam.ID,
		"clinic_id":         exam.ClinicID,
		"medical_record_id": exam.MedicalRecordID,
		"pet_id":            exam.PetID,
		"exam_type_id":      exam.ExamTypeID,
		"doctor_id":         exam.DoctorID,
		"job_id":            exam.JobID,
		"date":              exam.Date.Format(time.DateOnly),
		"machine":           exam.Machine,
		"status":            exam.Status,
		"result_summary":    exam.ResultSummary,
	})
	items := make([]model.LabImportExamRetractionItem, 0, len(exam.Items))
	for i, it := range exam.Items {
		items = append(items, model.LabImportExamRetractionItem{
			ItemSnapshot: MustJSON(it),
			SortOrder:    i,
		})
	}
	retraction := &model.LabImportExamRetraction{
		ClinicID:       input.ClinicID,
		JobID:          input.JobID,
		ExamID:         exam.ID,
		ActorID:        input.ActorID,
		Reason:         reason,
		ParentSnapshot: parentSnap,
	}
	if err := s.retractions.CreateWithItems(ctx, retraction, items); err != nil {
		return err
	}
	affected, err := SoftDeleteExamByJob(ctx, s.db, input.ClinicID, exam.ID, input.JobID)
	if err != nil {
		return err
	}
	if affected != 1 {
		return apperrors.WrapConflict("exam soft delete RowsAffected mismatch; revert refused")
	}
	// Child exam_results are intentionally NOT hard-deleted.
	slog.InfoContext(ctx, "lab import exam retracted",
		"clinic_id", input.ClinicID,
		"job_id", input.JobID.String(),
		"exam_id", exam.ID,
	)
	return nil
}

func labImportRevertRequestHash(jobID uuid.UUID, reason string) string {
	h := sha256.Sum256([]byte(jobID.String() + "\n" + reason))
	return hex.EncodeToString(h[:])
}

func parseRetractedExamIDs(raw string) ([]uint64, error) {
	if raw == "" {
		return []uint64{}, nil
	}
	var ids []uint64
	if err := json.Unmarshal([]byte(raw), &ids); err != nil {
		return nil, err
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids, nil
}
