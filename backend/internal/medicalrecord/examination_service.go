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
	List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Examination, int64, error)
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

func (s *examinationService) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Examination, int64, error) {
	items, total, err := s.repo.FindAll(ctx, clinicID, petID, ownerID, status, startDate, endDate, page, limit)
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
		var record *model.MedicalRecord
		if input.MedicalRecordID != nil {
			var err error
			record, err = lockClinicalMedicalRecord(txCtx, s.medRec, clinicID, *input.MedicalRecordID)
			if err != nil {
				return err
			}
			if record.Status == model.MedicalRecordStatusFinalized {
				return apperrors.WrapConflict("確定済みカルテに検査を追加できません")
			}
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
		locked, err := s.repo.LockByIDForUpdate(txCtx, clinicID, id)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to lock examination", "error", err)
			return apperrors.Wrap(err, "failed to lock examination")
		}
		if locked.Status == model.ExaminationStatusConfirmed {
			return apperrors.WrapConflict("確定済みの検査は編集できません")
		}
		before := *locked
		revisioned := locked.CurrentRevisionVersion != nil
		petChanged := examinationOptionalIDChanged(locked.PetID, input.PetID)
		medicalRecordChanged := examinationOptionalIDChanged(locked.MedicalRecordID, input.MedicalRecordID)
		if revisioned && (petChanged || medicalRecordChanged) {
			return apperrors.WrapConflict("revision history exists; examination patient relation cannot be changed")
		}
		if revisioned && s.revisionWorkflow == nil {
			return apperrors.WrapInternalServerError("examination revision workflow repository capability is required")
		}
		if confirming && revisioned {
			if len(fields) != 0 || input.Items != nil {
				return apperrors.WrapConflict("save working examination changes before reconfirming")
			}
			exam, err = s.reconfirmRevisionTx(txCtx, clinicID, input.ActorID, locked)
			return err
		}

		medicalRecordID, petID, doctorID := effectiveExaminationRelations(locked, input)
		record, err := s.lockExaminationUpdateMedicalRecords(
			txCtx,
			clinicID,
			locked.MedicalRecordID,
			medicalRecordID,
		)
		if err != nil {
			return err
		}
		if err := validateClinicalRelations(txCtx, s.relations, clinicID, record, petID, doctorID); err != nil {
			return err
		}
		if petChanged || medicalRecordChanged {
			targetPetID := effectiveExaminationPetID(petID, record)
			if err := validateExaminationPetNotDeceased(txCtx, s.petStatuses, clinicID, targetPetID); err != nil {
				return err
			}
		}

		// クロステナント write 防止: 貼り替え先 exam_type が caller の clinic に属することを検証する。
		if err := validateOwnedMasterFK(txCtx, "exam type", clinicID, input.ExamTypeID,
			func(actx context.Context, cid, mid uint64) error {
				_, err := s.examTypeRepo.FindByID(actx, cid, mid)
				return err
			}); err != nil {
			return err
		}

		itemsToReplace := input.Items
		if petChanged && itemsToReplace == nil {
			existingItems, err := s.repo.FindAllItemsByExamID(txCtx, clinicID, id)
			if err != nil {
				return apperrors.Wrap(err, "failed to load examination items for patient reassessment")
			}
			reassessedInputs := examinationItemsToUpsertInputs(existingItems)
			itemsToReplace = &reassessedInputs
		}

		exam = locked
		if len(fields) > 0 {
			updated, err := s.repo.Update(txCtx, clinicID, id, fields)
			if err != nil {
				slog.ErrorContext(txCtx, "failed to update examination", "error", err)
				return apperrors.Wrap(err, "failed to update examination")
			}
			exam = updated
		}
		if itemsToReplace != nil {
			if _, err := s.replaceItemsTx(txCtx, clinicID, exam, input.ActorID, *itemsToReplace); err != nil {
				return err
			}
		}
		if confirming {
			exam, err = s.confirmFirstRevisionTx(
				txCtx,
				clinicID,
				input.ActorID,
				exam,
				&before,
				model.AuditActionExaminationConfirm,
				"confirm",
			)
			return err
		}
		if revisioned {
			exam, err = s.appendWorkingRevisionTx(txCtx, clinicID, input.ActorID, &before, exam, examinationWorkingUpdateReason)
			return err
		}
		return s.logParentMutationTx(
			txCtx,
			clinicID,
			input.ActorID,
			model.AuditActionExaminationUpdate,
			"update",
			&before,
			exam,
		)
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

// ListItems は検査項目一覧を返す。clinic_id 隔離は repository の JOIN 条件で保証する。
// 親 exam の存在確認は FindByID で先行する（404 を返すため）。
func (s *examinationService) ListItems(ctx context.Context, clinicID, examID uint64) ([]model.ExamResult, error) {
	if _, err := s.repo.FindByID(ctx, clinicID, examID); err != nil {
		return nil, apperrors.Wrap(err, "failed to find examination")
	}
	items, err := s.repo.FindAllItemsByExamID(ctx, clinicID, examID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list examination items", "error", err, "exam_id", examID, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to list examination items")
	}
	return items, nil
}

// ReplaceItems は検査項目を一括置換する（PUT セマンティクス）。
//
// 仕様:
//  1. 親 exam の存在を FindByID で確認（P1）
//  2. 親 exam が confirmed の場合は Conflict (409) で拒否
//  3. 各 input の inspection_value とサーバで解決した基準値から status / is_abnormal を導出
//  4. repository の ReplaceItemsByExamID（トランザクション内で全削除→一括挿入）に委譲
//  5. 実削除が発生した場合（deletedCount > 0）は同一 tx 内で監査ログを書き込む。監査書込が失敗したら
//     tx を rollback する（best-effort ではなく fail-closed。BE-refactor.md R1-2・#211 と同方針）。
func (s *examinationService) ReplaceItems(ctx context.Context, clinicID, examID uint64, actorID *uint64, inputs []UpsertExamItemInput) ([]model.ExamResult, error) {
	if s.transactor == nil {
		return nil, apperrors.WrapInternalServerError("examination write transaction dependency is required")
	}

	// #211/R1-2 tx 内監査による原子的置換: スナップショット読取→削除/挿入→削除監査 を単一トランザクションで
	// 実行する。監査書込が失敗したら tx 全体を rollback し、削除・挿入も巻き戻す（監査なしの検査結果削除を
	// 許さない＝fail-closed）。exam_results は hard-delete のため old_value が唯一の耐久記録であり、
	// 旧コードでは「置換 commit 後に監査を書く」経路自体が存在しなかった（audit_tx_inventory_lint_test.go
	// が発見した無監査ギャップ）。スナップショットも同一 tx 内で取得し TOCTOU 窓を作らない。
	var saved []model.ExamResult
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		locked, err := s.repo.LockByIDForUpdate(txCtx, clinicID, examID)
		if err != nil {
			return apperrors.Wrap(err, "failed to lock examination")
		}
		if locked.Status == model.ExaminationStatusConfirmed {
			return apperrors.WrapConflict("確定済みの検査は編集できません")
		}
		before := *locked
		revisioned := locked.CurrentRevisionVersion != nil
		if revisioned {
			if s.revisionWorkflow == nil {
				return apperrors.WrapInternalServerError("examination revision workflow repository capability is required")
			}
			if err := s.validateParentMutationAudit(actorID); err != nil {
				return err
			}
		}
		if locked.MedicalRecordID != nil {
			if err := lockDraftMedicalRecord(
				txCtx,
				s.medRec,
				clinicID,
				*locked.MedicalRecordID,
				"failed to find medical record",
				"確定済みカルテの検査結果は編集できません",
			); err != nil {
				return err
			}
		}

		replaced, err := s.replaceItemsTx(txCtx, clinicID, locked, actorID, inputs)
		if err != nil {
			return err
		}
		saved = replaced
		if revisioned {
			_, err = s.appendWorkingRevisionTx(
				txCtx,
				clinicID,
				actorID,
				&before,
				locked,
				examinationWorkingItemsReason,
			)
			return err
		}
		return nil
	}); err != nil {
		return nil, apperrors.Wrap(err, "failed to replace examination items in transaction")
	}

	slog.InfoContext(ctx, "examination items replaced",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("examination_id", examID),
		slog.Int("item_count", len(saved)),
	)
	return saved, nil
}

// replaceItemsTx validates and replaces examination results inside the caller-owned transaction.
// Create, Update, and the split PUT endpoint all share this tail so result persistence and
// deletion audit use the same transaction as the parent mutation.
func (s *examinationService) replaceItemsTx(
	ctx context.Context,
	clinicID uint64,
	exam *model.Examination,
	actorID *uint64,
	inputs []UpsertExamItemInput,
) ([]model.ExamResult, error) {
	fieldIDs := make([]uint64, 0, len(inputs))
	fieldIDSet := make(map[uint64]struct{}, len(inputs))
	for _, in := range inputs {
		if in.ExamTypeFieldID != nil {
			if _, exists := fieldIDSet[*in.ExamTypeFieldID]; !exists {
				fieldIDSet[*in.ExamTypeFieldID] = struct{}{}
				fieldIDs = append(fieldIDs, *in.ExamTypeFieldID)
			}
		}
	}

	// #124 防止: request の exam_type_field が caller の clinic に属する、ロック済み検査の
	// 検査種別フィールドであることを同じ transaction 内で検証する。
	if len(fieldIDs) > 0 {
		examType, err := s.examTypeRepo.FindByID(ctx, clinicID, exam.ExamTypeID)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to verify exam type ownership")
		}
		validFieldIDs := make(map[uint64]struct{}, len(examType.Items))
		for i := range examType.Items {
			validFieldIDs[examType.Items[i].ID] = struct{}{}
		}
		for _, in := range inputs {
			if in.ExamTypeFieldID != nil {
				if _, ok := validFieldIDs[*in.ExamTypeFieldID]; !ok {
					return nil, apperrors.WrapInvalidInput("exam_type_field が当該検査種別に属していません（別クリニック/別種別の項目は紐付けできません）")
				}
			}
		}
	}

	resolvedRanges := make(map[uint64]model.ExamReferenceRange, len(fieldIDs))
	if len(fieldIDs) > 0 {
		if exam.PetID == nil {
			return nil, apperrors.WrapInvalidInput("基準値を解決するには検査対象のペットが必要です")
		}
		if s.referenceRanges == nil {
			return nil, apperrors.WrapInternalServerError("examination reference range resolver is required")
		}
		animalSpeciesID, err := s.referenceRanges.FindAnimalSpeciesID(ctx, clinicID, exam.ID)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to resolve examination animal species")
		}
		resolvedRanges, err = s.referenceRanges.ResolveByFieldIDs(ctx, clinicID, animalSpeciesID, fieldIDs)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to resolve examination reference ranges")
		}
	}

	items := make([]model.ExamResult, 0, len(inputs))
	for _, in := range inputs {
		var refMin, refMax *float64
		var qualitativeMin, qualitativeMax *string
		if in.ExamTypeFieldID != nil {
			if referenceRange, ok := resolvedRanges[*in.ExamTypeFieldID]; ok {
				refMin = cloneOptionalFloat64(referenceRange.RefMin)
				refMax = cloneOptionalFloat64(referenceRange.RefMax)
				qualitativeMin = cloneOptionalString(referenceRange.QualitativeMin)
				qualitativeMax = cloneOptionalString(referenceRange.QualitativeMax)
			}
		}
		assessment := assessExamResult(
			in.InspectionValue,
			refMin,
			refMax,
			qualitativeMin,
			qualitativeMax,
		)
		items = append(items, model.ExamResult{
			ExamID:          exam.ID,
			ExamTypeItemID:  in.ExamTypeFieldID,
			Name:            in.Name,
			InspectionValue: in.InspectionValue,
			NormalValue:     in.NormalValue,
			Result:          in.Result,
			Unit:            in.Unit,
			ReferenceValue:  in.ReferenceValue,
			RefMin:          refMin,
			RefMax:          refMax,
			QualitativeMin:  qualitativeMin,
			QualitativeMax:  qualitativeMax,
			IsAbnormal:      assessment.isAbnormal,
			Status:          assessment.status,
			SortOrder:       in.SortOrder,
		})
	}

	before, err := s.repo.FindAllItemsByExamID(ctx, clinicID, exam.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to snapshot existing examination items before replace", "error", err, "exam_id", exam.ID, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to load existing examination items")
	}

	saved, deletedCount, err := s.repo.ReplaceItemsByExamID(ctx, clinicID, exam.ID, items)
	if err != nil {
		slog.ErrorContext(ctx, "failed to replace examination items", "error", err, "exam_id", exam.ID, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to replace examination items")
	}

	// 実際に削除が発生した場合のみ監査する（純粋な新規挿入は削除を伴わない）。ゲートはスナップショット
	// 件数でなく DELETE の実削除数（deletedCount）に基づく（#211 security MEDIUM-1 と同方針: 並行 INSERT
	// 競合下でスナップショット 0 件でも実削除>0 を取りこぼさない）。監査書込失敗は tx を rollback する。
	if err := logReplaceDeletionTx(ctx, s.auditTx, clinicID, actorID, deletedCount,
		model.AuditActionExamResultReplace, model.AuditResourceExamResult, exam.ID,
		extractExamResultsAudit(before), extractExamResultsAudit(saved),
		map[string]any{
			"exam_id":       exam.ID,
			"deleted_count": deletedCount,
			"new_count":     len(saved),
		},
		"audit log failed for examination items replace; rolling back deletion",
		"failed to write examination items deletion audit", "exam_id"); err != nil {
		return nil, err
	}
	return saved, nil
}

func cloneOptionalFloat64(value *float64) *float64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

// extractExamResultsAudit は監査ログの old_value/new_value に格納する検査結果値のスナップショットを構築する。
// 飼主/患者の識別情報は含まず、行 ID・フィールド定義参照・検査値のみを記録する
// （extractCheckupFieldResultsAudit と同方針）。
func extractExamResultsAudit(results []model.ExamResult) []map[string]any {
	if len(results) == 0 {
		return nil
	}
	out := make([]map[string]any, 0, len(results))
	for i := range results {
		r := results[i]
		// is_assessed is derived, not stored — reuse assessExamResult so audit
		// provenance matches API assessment rules (bounds, fail-closed cases).
		assessment := assessExamResult(
			r.InspectionValue,
			r.RefMin,
			r.RefMax,
			r.QualitativeMin,
			r.QualitativeMax,
		)
		out = append(out, map[string]any{
			"id":                 r.ID,
			"exam_type_field_id": r.ExamTypeItemID,
			"name":               r.Name,
			"inspection_value":   r.InspectionValue,
			"ref_min":            r.RefMin,
			"ref_max":            r.RefMax,
			"qualitative_min":    r.QualitativeMin,
			"qualitative_max":    r.QualitativeMax,
			"is_assessed":        assessment.isAssessed,
			"is_abnormal":        r.IsAbnormal,
			"status":             string(r.Status),
		})
	}
	return out
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
		if locked.Status == model.ExaminationStatusConfirmed {
			return apperrors.WrapConflict("確定済みの検査は削除できません")
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
