package medicalrecord

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *medicalRecordService) List(ctx context.Context, clinicIDs []uint64, filters MedicalRecordListFilters, page, limit int) ([]model.MedicalRecord, int64, error) {
	items, total, err := s.repo.FindAll(ctx, clinicIDs, filters, page, limit)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list medical records", "error", err)
		return nil, 0, apperrors.Wrap(err, "failed to list medical records")
	}
	return items, total, nil
}

func (s *medicalRecordService) GetByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get medical record", "error", err)
		return nil, apperrors.Wrap(err, "failed to get medical record")
	}
	return result, nil
}

func (s *medicalRecordService) GetByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.MedicalRecord, error) {
	result, err := s.repo.FindByIDForClinics(ctx, clinicIDs, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get medical record for clinics", "error", err)
		return nil, apperrors.Wrap(err, "failed to get medical record for clinics")
	}
	return result, nil
}

func (s *medicalRecordService) CountByPetID(ctx context.Context, clinicID, petID uint64) (int64, error) {
	count, err := s.repo.CountByPetID(ctx, clinicID, petID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to count medical records by pet", "error", err)
		return 0, apperrors.Wrap(err, "failed to count medical records by pet")
	}
	return count, nil
}

type medicalRecordCreateTxResult struct {
	record       *model.MedicalRecord
	existingHit  *model.MedicalRecord
	isFirstVisit bool
}

func (s *medicalRecordService) Create(ctx context.Context, clinicID uint64, input *CreateMedicalRecordInput) (*model.MedicalRecord, error) {
	// FEAT-382-2 supplement: whitelist validation for recommendation_reason（tx 外）
	if input != nil && input.RecommendationReason != nil {
		if _, ok := allowedRecommendationReasons[*input.RecommendationReason]; !ok {
			return nil, apperrors.WrapInvalidInput(
				"recommendation_reason must be one of: revisit, checkup, prevention, exam",
			)
		}
	}

	var result medicalRecordCreateTxResult
	if err := s.withTx(ctx, func(txCtx context.Context) error {
		var err error
		result, err = s.createMedicalRecordInTx(txCtx, clinicID, input)
		return err
	}); err != nil {
		return nil, err
	}
	if result.existingHit != nil {
		return result.existingHit, nil
	}

	s.runMedicalRecordCreatePostCommit(ctx, result)

	return result.record, nil
}

func (s *medicalRecordService) createMedicalRecordInTx(
	ctx context.Context,
	clinicID uint64,
	input *CreateMedicalRecordInput,
) (medicalRecordCreateTxResult, error) {
	// Validate request-derived context before appointment preparation can persist a backfill.
	if err := s.validateMedicalRecordOwnerPetLinks(ctx, clinicID, input.OwnerID, input.PetID); err != nil {
		return medicalRecordCreateTxResult{}, err
	}
	if err := s.validateMedicalRecordDoctor(ctx, clinicID, input.DoctorID); err != nil {
		return medicalRecordCreateTxResult{}, err
	}
	if err := s.applyAppointmentContextForCreate(ctx, clinicID, input); err != nil {
		return medicalRecordCreateTxResult{}, apperrors.Wrap(err, "failed to apply appointment context for medical record")
	}
	// AUD-008: Appointment 有無を問わず最終 Owner/Pet を clinic 所有・整合検証する。
	if err := s.validateMedicalRecordOwnerPetLinks(ctx, clinicID, input.OwnerID, input.PetID); err != nil {
		return medicalRecordCreateTxResult{}, err
	}
	if err := s.validateMedicalRecordDoctor(ctx, clinicID, input.DoctorID); err != nil {
		return medicalRecordCreateTxResult{}, err
	}
	// BUG-002: 最終 pet が確定したあとに死亡ガード（repo.Create 前・fail-closed）。
	if err := s.assertMedicalRecordPetNotDeceased(ctx, clinicID, input.PetID); err != nil {
		return medicalRecordCreateTxResult{}, err
	}

	built := buildMedicalRecordForCreate(clinicID, input)
	existing, err := s.findExistingRecordByAppointment(ctx, clinicID, built)
	if err != nil {
		return medicalRecordCreateTxResult{}, err
	}
	if existing != nil {
		return medicalRecordCreateTxResult{existingHit: existing}, nil
	}

	if built.RecordNo == "" {
		built.RecordNo = generateRecordNo(built.Date, built.ClinicID)
	}

	isFirstVisit := false
	// 初診判定: Reservation.VisitType 優先、AppointmentID なし or VisitType 空時は COUNT フォールバック（FEAT-383）
	if s.lstepDeliveryTrigger != nil && built.OwnerID != nil {
		visitType := s.resolveVisitTypeFromAppointment(ctx, built)
		switch {
		case visitType == string(model.VisitTypeFirst):
			isFirstVisit = true
		case visitType != "":
			isFirstVisit = false
		default:
			isFirstVisit = s.fallbackFirstVisitCheck(ctx, built.ClinicID, *built.OwnerID)
		}
	}

	if err := s.repo.Create(ctx, built); err != nil {
		slog.ErrorContext(ctx, "failed to create medical record", "error", err, "clinic_id", built.ClinicID)
		return medicalRecordCreateTxResult{}, apperrors.Wrap(err, "failed to create medical record")
	}

	return medicalRecordCreateTxResult{record: built, isFirstVisit: isFirstVisit}, nil
}

func (s *medicalRecordService) runMedicalRecordCreatePostCommit(ctx context.Context, result medicalRecordCreateTxResult) {
	record := result.record
	if record == nil {
		return
	}

	slog.InfoContext(ctx, "medical record created",
		slog.Uint64("record_id", record.ID),
		slog.Uint64("clinic_id", record.ClinicID))

	// 監査ログ: create（best-effort）
	if s.auditService != nil {
		newValue := extractMedicalRecordImportantFields(record)
		if err := s.auditService.LogMedicalRecordChange(ctx, record.ClinicID, record.EnteredBy, "create", record.ID, nil, newValue); err != nil {
			slog.ErrorContext(ctx, "audit log failed for medical record create", "error", err, "record_id", record.ID)
		}
	}

	// 初診ウェルカムトリガー（イベント駆動・非致命的）
	if result.isFirstVisit && record.OwnerID != nil {
		if err := s.lstepDeliveryTrigger.TriggerFirstVisitWelcome(ctx, record.ClinicID, *record.OwnerID); err != nil {
			slog.WarnContext(ctx, "first visit welcome trigger failed (non-fatal)", "owner_id", *record.OwnerID, "error", err)
		}
	}

	s.syncNextVisitTag(ctx, record.ClinicID, record)
}

func (s *medicalRecordService) Update(ctx context.Context, clinicID, id uint64, input UpdateMedicalRecordInput) (*model.MedicalRecord, error) {
	// 楽観的ロックチェックのため現在のレコードを取得
	existing, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get medical record", "error", err)
		return nil, apperrors.Wrap(err, "failed to get medical record")
	}

	// 確定済みカルテは更新不可
	if existing.Status == model.MedicalRecordStatusFinalized {
		slog.WarnContext(ctx, "attempted to update finalized medical record",
			slog.Uint64("record_id", id),
			slog.Uint64("clinic_id", clinicID))
		return nil, apperrors.WrapConflict("確定済みカルテは編集できません。訂正追記 (addendum) を使用してください")
	}
	if input.AppointmentID != nil && (existing.AppointmentID == nil || *input.AppointmentID != *existing.AppointmentID) {
		return nil, apperrors.WrapConflict("作成済みカルテの予約紐付けは変更できません。監査付きの専用再紐付け手順が必要です")
	}
	if existing.AppointmentID != nil && input.Date != nil {
		return nil, apperrors.WrapConflict("予約に紐づくカルテの診療日は変更できません。予約日時の訂正手順を使用してください")
	}

	// AUD-008: context変更時は最終状態でclinic所有・appointment整合を検証し、writeと同一txにする。
	needsLinkValidation := input.OwnerID != nil || input.PetID != nil
	needsContextValidation := needsLinkValidation || input.DoctorID != nil
	finalOwnerID, finalPetID := resolveFinalMedicalRecordOwnerPet(existing, input)
	finalDoctorID := resolveFinalMedicalRecordDoctor(existing, input)

	fields := buildMedicalRecordUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	// バージョンをインクリメント（input.Version 指定時はそれを起点にする。BE-refactor.md X-10:
	// version 一致確認は repo.Update の WHERE 述語に一本化したため、ここでの事前チェックは行わない）
	nextVersion := existing.Version + 1
	if input.Version != nil {
		nextVersion = *input.Version + 1
	}
	fields["version"] = nextVersion

	// 監査 diff は更新前に取得（finalize 検出のため）
	wasFinalized := existing.Status == model.MedicalRecordStatusFinalized
	isBecomingFinalized := input.Status != nil && *input.Status == model.MedicalRecordStatusFinalized && !wasFinalized

	var record *model.MedicalRecord
	if err := s.withTx(ctx, func(txCtx context.Context) error {
		updated, err := s.updateMedicalRecordInTx(
			txCtx,
			clinicID,
			id,
			existing,
			input,
			fields,
			finalOwnerID,
			finalPetID,
			finalDoctorID,
			needsLinkValidation,
			needsContextValidation,
			isBecomingFinalized,
		)
		if err != nil {
			return err
		}
		record = updated
		return nil
	}); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "medical record updated",
		slog.Uint64("record_id", id),
		slog.Uint64("clinic_id", clinicID))

	// 監査ログ: update（best-effort）。finalize は上の transaction 内で fail-closed に記録する。
	s.logMedicalRecordChangeBestEffort(ctx, clinicID, input.ActorID, "update", id, existing, record, "audit log failed for medical record update")

	if isBecomingFinalized {
		s.syncVisitCompletionTags(ctx, clinicID, record)
	}
	s.syncNextVisitTag(ctx, clinicID, record)

	return record, nil
}

func (s *medicalRecordService) Delete(ctx context.Context, clinicID, id uint64) error {
	var existing *model.MedicalRecord
	if err := s.withTx(ctx, func(txCtx context.Context) error {
		// appointments を先にロックし、予約側（LockAndFindByID → COUNT）と同じ順序にする。
		link, err := s.repo.LockLinkedAppointmentForUpdate(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to lock linked appointment for deletion")
		}
		if link != nil && link.Appointment != nil && link.Appointment.Status == model.ReservationStatusInConsultation {
			return wrapMedicalRecordDeleteConflict(
				medicalRecordDeleteStateConflict,
				"診療中の予約に紐づくカルテは削除できません",
			)
		}
		locked, err := s.repo.LockByIDForUpdate(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to lock medical record for deletion")
		}
		if locked.Status != model.MedicalRecordStatusDraft {
			return wrapMedicalRecordDeleteConflict(
				medicalRecordDeleteStateConflict,
				"確定済みまたは下書き以外の診療記録は削除できません",
			)
		}
		estimateCount, err := s.repo.CountEstimatesByMedicalRecordID(txCtx, clinicID, id)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to check estimate dependencies", "error", err, "id", id, "clinic_id", clinicID)
			return apperrors.Wrap(err, "failed to check estimate dependencies")
		}
		if estimateCount > 0 {
			return wrapMedicalRecordDeleteConflict(
				medicalRecordDeleteDependencyConflict,
				"この項目は使用中のため削除できません",
			)
		}
		if err := s.repo.Delete(txCtx, clinicID, id); err != nil {
			slog.ErrorContext(txCtx, "failed to delete medical record", "error", err, "id", id, "clinic_id", clinicID)
			return apperrors.Wrap(err, "failed to delete medical record")
		}
		existing = locked
		return nil
	}); err != nil {
		return err
	}
	oldValue := extractMedicalRecordImportantFields(existing)
	slog.InfoContext(ctx, "medical record deleted",
		slog.Uint64("record_id", id),
		slog.Uint64("clinic_id", clinicID))

	// 監査ログ: delete（best-effort, actorID=nil）
	if s.auditService != nil {
		if err := s.auditService.LogMedicalRecordChange(ctx, clinicID, nil, "delete", id, oldValue, nil); err != nil {
			slog.ErrorContext(ctx, "audit log failed for medical record delete", "error", err, "record_id", id)
		}
	}
	s.syncVisitCompletionTags(ctx, clinicID, existing)
	s.syncNextVisitTag(ctx, clinicID, existing)
	return nil
}

func (s *medicalRecordService) UpdateRecommendationReason(
	ctx context.Context, clinicID, id uint64, input UpdateRecommendationReasonInput,
) (*model.MedicalRecord, error) {
	// 1. whitelist validation (空文字は NULL として許容)
	if input.Reason != "" {
		if _, ok := allowedRecommendationReasons[input.Reason]; !ok {
			return nil, apperrors.WrapInvalidInput(
				"recommendation_reason must be one of: revisit, checkup, prevention, exam",
			)
		}
	}
	// 2. P1: existence + clinic_id check（audit diff 用にレコードを保持）
	existing, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find medical record")
	}
	// 確定済みカルテは推奨理由も更新不可（Update と同一方針。repo.Update 自体も
	// status=draft の WHERE で原子的に拒否するが、この事前チェックで友好的な
	// エラーメッセージと slog 警告を出す。SD-2 系ガード監査で発見された欠落）。
	if existing.Status == model.MedicalRecordStatusFinalized {
		slog.WarnContext(ctx, "attempted to update recommendation_reason on finalized medical record",
			slog.Uint64("record_id", id),
			slog.Uint64("clinic_id", clinicID))
		return nil, apperrors.WrapConflict("確定済みカルテの推奨理由は編集できません")
	}
	// 3. build update map
	var reasonValue any
	if input.Reason == "" {
		reasonValue = nil // SQL NULL
	} else {
		reasonValue = input.Reason
	}

	// バージョン照合なし（このメソッドの入力に version が存在しないため従来どおり）
	record, err := s.repo.Update(ctx, clinicID, id, map[string]any{
		colRecommendationReason: reasonValue,
	}, nil)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update recommendation_reason", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to update recommendation_reason")
	}
	slog.InfoContext(ctx, "recommendation_reason updated",
		slog.Uint64("record_id", id),
		slog.Uint64("clinic_id", clinicID),
		slog.String("reason", input.Reason))

	// 監査ログ: update（best-effort）
	s.logMedicalRecordChangeBestEffort(ctx, clinicID, nil, "update", id, existing, record, "audit log failed for recommendation_reason update")

	return record, nil
}

func (s *medicalRecordService) validateMedicalRecordUpdateLinks(
	ctx context.Context,
	clinicID uint64,
	ownerID, petID *uint64,
	needsLinkValidation, isBecomingFinalized bool,
) error {
	if needsLinkValidation {
		return s.validateMedicalRecordOwnerPetLinks(ctx, clinicID, ownerID, petID)
	}
	if isBecomingFinalized {
		return s.validateMedicalRecordSnapshotOwnerPetClinicRelations(ctx, clinicID, ownerID, petID)
	}
	return nil
}

func (s *medicalRecordService) logMedicalRecordChangeBestEffort(
	ctx context.Context,
	clinicID uint64,
	actorID *uint64,
	action string,
	id uint64,
	existing, record *model.MedicalRecord,
	failMsg string,
) {
	if s.auditService == nil {
		return
	}
	oldDiff, newDiff := diffMedicalRecordImportantFields(existing, record)
	if oldDiff == nil {
		return
	}
	if err := s.auditService.LogMedicalRecordChange(ctx, clinicID, actorID, action, id, oldDiff, newDiff); err != nil {
		slog.ErrorContext(ctx, failMsg, "error", err, "record_id", id)
	}
}
