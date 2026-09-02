package medicalrecord

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// CreateExaminationInput は検査作成の入力DTO
type CreateExaminationInput struct {
	MedicalRecordID *uint64
	PetID           *uint64
	ExamTypeID      uint64
	DoctorID        *uint64
	Date            time.Time
	ResultSummary   string
	Machine         string
	Status          model.ExaminationStatus
	Items           *[]UpsertExamItemInput
	ActorID         *uint64
}

// UpdateExaminationInput は検査更新のサービス入力 DTO
type UpdateExaminationInput struct {
	MedicalRecordID *uint64
	PetID           *uint64
	ExamTypeID      *uint64
	DoctorID        *uint64
	Date            *time.Time
	ResultSummary   *string
	Machine         *string
	Status          *model.ExaminationStatus
	Items           *[]UpsertExamItemInput
	ActorID         *uint64
}

type UnconfirmExaminationInput struct {
	Reason  string
	ActorID *uint64
}

// UpsertExamItemInput は検査項目（exam_results）の一括登録入力 DTO。
// status / is_abnormal はサーバ側で計算するため受け付けない（信頼境界はサーバ）。
type UpsertExamItemInput struct {
	ExamTypeFieldID *uint64
	Name            string
	InspectionValue string
	NormalValue     string
	Result          string
	Unit            string
	ReferenceValue  string
	SortOrder       int
}

// computeExamResultStatus は inspection_value を float としてパースし、
// ref_min / ref_max と比較して status と is_abnormal を導出する。
//
// 仕様:
//   - inspection_value が空・パース不能 → (normal, false)
//   - ref_min が指定され v < ref_min → (low, true)
//   - ref_max が指定され v > ref_max → (high, true)
//   - 範囲内 → (normal, false)
//   - ref_min == ref_max == nil → (normal, false)（比較できない）
func computeExamResultStatus(inspectionValue string, refMin, refMax *float64) (model.ExaminationResultStatus, bool) {
	assessment := assessExamResult(inspectionValue, refMin, refMax, nil, nil)
	return assessment.status, assessment.isAbnormal
}

func buildExaminationUpdate(input UpdateExaminationInput) map[string]any {
	fields := make(map[string]any)
	if input.MedicalRecordID != nil {
		fields["medical_record_id"] = *input.MedicalRecordID
	}
	if input.PetID != nil {
		fields["pet_id"] = *input.PetID
	}
	if input.ExamTypeID != nil {
		fields["exam_type_id"] = *input.ExamTypeID
	}
	if input.DoctorID != nil {
		fields["doctor_id"] = *input.DoctorID
	}
	if input.Date != nil {
		fields["date"] = *input.Date
	}
	if input.ResultSummary != nil {
		fields["result_summary"] = *input.ResultSummary
	}
	if input.Machine != nil {
		fields["machine"] = *input.Machine
	}
	if input.Status != nil {
		fields["status"] = *input.Status
	}
	return fields
}

func effectiveExaminationRelations(existing *model.Examination, input UpdateExaminationInput) (medicalRecordID, petID, doctorID *uint64) {
	medicalRecordID = existing.MedicalRecordID
	petID = existing.PetID
	doctorID = existing.DoctorID
	if input.MedicalRecordID != nil {
		medicalRecordID = input.MedicalRecordID
	}
	if input.PetID != nil {
		petID = input.PetID
	}
	if input.DoctorID != nil {
		doctorID = input.DoctorID
	}
	return medicalRecordID, petID, doctorID
}

type ExaminationService interface {
	List(ctx context.Context, clinicID uint64, petID, ownerID, medicalRecordID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Examination, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Examination, error)
	// GetPrintSnapshot returns a clinic-scoped atomic revision print DTO.
	// version nil uses the parent's current_revision_version (fail-closed if unset).
	GetPrintSnapshot(ctx context.Context, clinicID, examinationID uint64, version *uint64) (*ExaminationPrintSnapshot, error)
	Create(ctx context.Context, clinicID uint64, input *CreateExaminationInput) (*model.Examination, error)
	Update(ctx context.Context, clinicID, id uint64, input UpdateExaminationInput) (*model.Examination, error)
	Unconfirm(ctx context.Context, clinicID, id uint64, input UnconfirmExaminationInput) (*model.Examination, error)
	Delete(ctx context.Context, clinicID, id uint64, actorID *uint64) error
	ListItems(ctx context.Context, clinicID, examID uint64) ([]model.ExamResult, error)
	// ReplaceItems は検査項目を一括置換する（PUT セマンティクス）。actorID は監査ログ用の操作スタッフ ID
	// （nil = システム実行）。BE-refactor.md R1-2: 実削除が発生した場合は同一 tx 内で fail-closed 監査する。
	ReplaceItems(ctx context.Context, clinicID, examID uint64, actorID *uint64, inputs []UpsertExamItemInput) ([]model.ExamResult, error)
}

type examinationService struct {
	repo             ExaminationRepository
	medRec           medicalRecordLocker // lockDraftMedicalRecord のみ使用（⑦で narrow 化）
	examTypeRepo     ExamTypeRepository
	referenceRanges  ExamReferenceRangeResolver
	revisions        ExaminationRevisionRepository
	revisionWorkflow ExaminationRevisionWorkflowRepository
	auditTx          AuditTxLogger
	transactor       Transactor
	relations        ClinicalRelationVerifier
	petStatuses      examinationPetByIDInClinicFinder
	// usageTracker records import-linked clinical use / manual mutation receipts (TASK-032).
	// Nil-safe: treated as noop when unset (legacy tests / composition without lab receipts).
	usageTracker LabImportUsageTracker
}

// AttachLabImportUsageTracker wires TASK-032 usage receipt instrumentation onto an ExaminationService.
// No-op when svc is not the concrete examinationService or tracker is nil.
func AttachLabImportUsageTracker(svc ExaminationService, tracker LabImportUsageTracker) ExaminationService {
	if tracker == nil {
		return svc
	}
	if s, ok := svc.(*examinationService); ok {
		s.usageTracker = tracker
	}
	return svc
}

func (s *examinationService) usage() LabImportUsageTracker {
	if s.usageTracker != nil {
		return s.usageTracker
	}
	return noopLabImportUsageTracker{}
}

func NewExaminationService(
	repo ExaminationRepository,
	medRec medicalRecordLocker,
	examTypeRepo ExamTypeRepository,
	auditTx AuditTxLogger,
	transactor Transactor,
	relationVerifier ...ClinicalRelationVerifier,
) ExaminationService {
	var relations ClinicalRelationVerifier
	if len(relationVerifier) > 0 {
		relations = relationVerifier[0]
	} else {
		// A concrete dependency may intentionally implement both narrow views.
		// Production composition passes the verifier explicitly.
		relations, _ = medRec.(ClinicalRelationVerifier)
	}
	if transactor == nil {
		transactor, _ = medRec.(Transactor)
	}
	referenceRanges, _ := repo.(ExamReferenceRangeResolver)
	revisions, _ := repo.(ExaminationRevisionRepository)
	revisionWorkflow, _ := repo.(ExaminationRevisionWorkflowRepository)
	petStatuses, _ := relations.(examinationPetByIDInClinicFinder)
	return &examinationService{
		repo: repo, medRec: medRec, examTypeRepo: examTypeRepo, referenceRanges: referenceRanges,
		revisions: revisions, revisionWorkflow: revisionWorkflow,
		auditTx: auditTx, transactor: transactor, relations: relations, petStatuses: petStatuses,
	}
}

func (s *examinationService) List(ctx context.Context, clinicID uint64, petID, ownerID, medicalRecordID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Examination, int64, error) {
	items, total, err := s.repo.FindAll(ctx, clinicID, petID, ownerID, medicalRecordID, status, startDate, endDate, page, limit)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list examinations", "error", err)
		return nil, 0, apperrors.Wrap(err, "failed to list examinations")
	}
	return items, total, nil
}

func (s *examinationService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Examination, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get examination", "error", err)
		return nil, apperrors.Wrap(err, "failed to get examination")
	}
	// TASK-032: record usage receipt before returning clinical payload.
	if err := s.usage().RecordClinicalUse(ctx, clinicID, result, model.LabImportUsageKindExaminationDetail, nil); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *examinationService) Create(ctx context.Context, clinicID uint64, input *CreateExaminationInput) (*model.Examination, error) {
	if s.transactor == nil {
		return nil, apperrors.WrapInternalServerError("examination write transaction dependency is required")
	}
	if err := s.validateParentMutationAudit(input.ActorID); err != nil {
		return nil, err
	}
	targetStatus := input.Status
	if targetStatus == "" {
		targetStatus = model.ExaminationStatusPending
	}
	// A create request may carry status=confirmed. Persist an editable initial parent so item/range
	// validation and replacement cannot self-reject, then perform the confirmed transition last.
	initialStatus := targetStatus
	if targetStatus == model.ExaminationStatusConfirmed {
		initialStatus = model.ExaminationStatusPending
	}
	exam := &model.Examination{
		ClinicID:        clinicID,
		MedicalRecordID: input.MedicalRecordID,
		PetID:           input.PetID,
		ExamTypeID:      input.ExamTypeID,
		DoctorID:        input.DoctorID,
		Date:            input.Date,
		ResultSummary:   input.ResultSummary,
		Machine:         input.Machine,
		Status:          initialStatus,
	}

	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		record, err := lockOptionalDraftMedicalRecord(txCtx, s.medRec, clinicID, input.MedicalRecordID, "確定済みカルテに検査を追加できません")
		if err != nil {
			return err
		}
		if err := validateClinicalRelations(txCtx, s.relations, clinicID, record, input.PetID, input.DoctorID); err != nil {
			return err
		}
		petID := effectiveExaminationPetID(input.PetID, record)
		if err := validateExaminationPetNotDeceased(txCtx, s.petStatuses, clinicID, petID); err != nil {
			return err
		}

		// クロステナント write 防止: 別 clinic の exam_type を紐付けると、その exam_type が持つ
		// 異常値判定の基準値/単位（exam_type_fields）が検査記録に混入する（#124 同型）。所有権を検証する。
		if input.ExamTypeID != 0 {
			if _, err := s.examTypeRepo.FindByID(txCtx, clinicID, input.ExamTypeID); err != nil {
				return apperrors.Wrap(err, "failed to verify exam type ownership")
			}
		}

		if err := s.repo.Create(txCtx, exam); err != nil {
			slog.ErrorContext(txCtx, "failed to create examination", "error", err)
			return apperrors.Wrap(err, "failed to create examination")
		}
		if input.Items != nil {
			if _, err := s.replaceItemsTx(txCtx, clinicID, exam, input.ActorID, *input.Items); err != nil {
				return err
			}
		}
		if targetStatus == model.ExaminationStatusConfirmed {
			locked, err := s.repo.LockByIDForUpdate(txCtx, clinicID, exam.ID)
			if err != nil {
				return apperrors.Wrap(err, "failed to lock created examination for confirmation")
			}
			exam, err = s.confirmFirstRevisionTx(
				txCtx,
				clinicID,
				input.ActorID,
				locked,
				nil,
				model.AuditActionExaminationCreate,
				"create",
			)
			return err
		}
		// Create does not reload database-normalized columns (notably the PostgreSQL date value).
		// Audit and response data must describe the durable parent row, not the caller's timestamp.
		persisted, err := s.repo.FindByID(txCtx, clinicID, exam.ID)
		if err != nil {
			return apperrors.Wrap(err, "failed to reload created examination")
		}
		exam = persisted
		return s.logParentMutationTx(
			txCtx,
			clinicID,
			input.ActorID,
			model.AuditActionExaminationCreate,
			"create",
			nil,
			exam,
		)
	}); err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "examination created", slog.Uint64("clinic_id", clinicID), slog.Uint64("examination_id", exam.ID))
	return exam, nil
}

func (s *examinationService) Update(ctx context.Context, clinicID, id uint64, input UpdateExaminationInput) (*model.Examination, error) {
	if s.transactor == nil {
		return nil, apperrors.WrapInternalServerError("examination write transaction dependency is required")
	}
	fields := buildExaminationUpdate(input)
	if len(fields) == 0 && input.Items == nil {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	if err := s.validateParentMutationAudit(input.ActorID); err != nil {
		return nil, err
	}
	confirming := input.Status != nil && *input.Status == model.ExaminationStatusConfirmed
	if confirming {
		// DEC-53 requires confirmed to be the final write. Keep all other parent fields in this
		// first immutable map and write status only after item/range validation and replacement.
		delete(fields, "status")
	}

	var exam *model.Examination
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		updated, err := s.updateExaminationInTx(txCtx, clinicID, id, input, fields, confirming)
		if err != nil {
			return err
		}
		exam = updated
		return nil
	}); err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "examination updated", slog.Uint64("clinic_id", clinicID), slog.Uint64("examination_id", id))
	return exam, nil
}

func (s *examinationService) lockExaminationUpdateMedicalRecords(
	ctx context.Context,
	clinicID uint64,
	currentID, effectiveID *uint64,
) (*model.MedicalRecord, error) {
	ids := make([]uint64, 0, 2)
	if currentID != nil {
		ids = append(ids, *currentID)
	}
	if effectiveID != nil && (currentID == nil || *effectiveID != *currentID) {
		ids = append(ids, *effectiveID)
	}
	if len(ids) == 2 && ids[0] > ids[1] {
		ids[0], ids[1] = ids[1], ids[0]
	}

	var effective *model.MedicalRecord
	for _, recordID := range ids {
		record, err := lockClinicalMedicalRecord(ctx, s.medRec, clinicID, recordID)
		if err != nil {
			return nil, err
		}
		if record.Status == model.MedicalRecordStatusFinalized {
			return nil, apperrors.WrapConflict("確定済みカルテの検査は編集できません")
		}
		if effectiveID != nil && recordID == *effectiveID {
			effective = record
		}
	}
	return effective, nil
}

func (s *examinationService) Delete(ctx context.Context, clinicID, id uint64, actorID *uint64) error {
	if s.transactor == nil {
		return apperrors.WrapInternalServerError("examination write transaction dependency is required")
	}
	if err := s.validateParentMutationAudit(actorID); err != nil {
		return err
	}

	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		locked, err := s.repo.LockByIDForUpdate(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to lock examination")
		}
		if examinationFullyLocked(locked) || examinationResultsLocked(locked) {
			return errExaminationDeleteLocked(locked)
		}
		if locked.CurrentRevisionVersion != nil {
			return apperrors.WrapConflict("確定履歴のある検査は削除できません")
		}
		// HC-003 + BE-refactor.md H-8d: 親カルテが確定済みの場合は削除拒否。LockByIDForUpdate の
		// 行ロックで finalize と直列化し、確定と同時の検査削除が確定済みカルテに混入する競合を防ぐ
		// （Update :215-227 と対称・nil ガード込み）。
		if locked.MedicalRecordID != nil {
			if err := lockDraftMedicalRecord(txCtx, s.medRec, clinicID, *locked.MedicalRecordID,
				"failed to find medical record", "確定済みカルテの検査は削除できません"); err != nil {
				return err
			}
		}

		// FK依存チェック: 検査に紐付く検査明細が存在する場合は削除を拒否
		itemCount, err := s.repo.CountItemsByExamID(txCtx, clinicID, id)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to check examination item dependencies", "error", err, "id", id, "clinic_id", clinicID)
			return apperrors.Wrap(err, "failed to check examination item dependencies")
		}
		if itemCount > 0 {
			return apperrors.WrapConflict("検査結果が紐付いているため削除できません。先に検査結果を削除してください")
		}

		if err := s.repo.Delete(txCtx, clinicID, id); err != nil {
			slog.ErrorContext(txCtx, "failed to delete examination", "error", err, "id", id, "clinic_id", clinicID)
			return apperrors.Wrap(err, "failed to delete examination")
		}
		return s.logParentMutationTx(
			txCtx,
			clinicID,
			actorID,
			model.AuditActionExaminationDelete,
			"delete",
			locked,
			nil,
		)
	}); err != nil {
		return err
	}

	slog.InfoContext(ctx, "examination deleted",
		slog.Uint64("examination_id", id),
		slog.Uint64("clinic_id", clinicID))

	return nil
}
