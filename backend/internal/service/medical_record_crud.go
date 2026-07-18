package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

func (s *medicalRecordService) List(ctx context.Context, clinicIDs []uint64, filters repository.MedicalRecordListFilters, page, limit int) ([]model.MedicalRecord, int64, error) {
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

func (s *medicalRecordService) Create(ctx context.Context, clinicID uint64, input *CreateMedicalRecordInput) (*model.MedicalRecord, error) {
	// FEAT-382-2 supplement: whitelist validation for recommendation_reason（tx 外）
	if input != nil && input.RecommendationReason != nil {
		if _, ok := allowedRecommendationReasons[*input.RecommendationReason]; !ok {
			return nil, apperrors.WrapInvalidInput(
				"recommendation_reason must be one of: revisit, checkup, prevention, exam",
			)
		}
	}

	var (
		record       *model.MedicalRecord
		isFirstVisit bool
		existingHit  *model.MedicalRecord
	)
	if err := s.withTx(ctx, func(txCtx context.Context) error {
		if err := s.applyAppointmentContextForCreate(txCtx, clinicID, input); err != nil {
			return apperrors.Wrap(err, "failed to apply appointment context for medical record")
		}
		// AUD-008: Appointment 有無を問わず最終 Owner/Pet を clinic 所有・整合検証する。
		if err := s.validateMedicalRecordOwnerPetLinks(txCtx, clinicID, input.OwnerID, input.PetID); err != nil {
			return err
		}

		built := buildMedicalRecordForCreate(clinicID, input)
		if existing := s.findExistingRecordByAppointment(txCtx, clinicID, built); existing != nil {
			existingHit = existing
			return nil
		}

		if built.RecordNo == "" {
			built.RecordNo = generateRecordNo(built.Date, built.ClinicID)
		}

		// 初診判定: Reservation.VisitType 優先、AppointmentID なし or VisitType 空時は COUNT フォールバック（FEAT-383）
		if s.lstepDeliveryTrigger != nil && built.OwnerID != nil {
			visitType := s.resolveVisitTypeFromAppointment(txCtx, built)
			switch {
			case visitType == string(model.VisitTypeFirst):
				isFirstVisit = true
			case visitType != "":
				isFirstVisit = false
			default:
				isFirstVisit = s.fallbackFirstVisitCheck(txCtx, built.ClinicID, *built.OwnerID)
			}
		}

		if err := s.repo.Create(txCtx, built); err != nil {
			slog.ErrorContext(txCtx, "failed to create medical record", "error", err, "clinic_id", built.ClinicID)
			return apperrors.Wrap(err, "failed to create medical record")
		}
		record = built
		return nil
	}); err != nil {
		return nil, err
	}
	if existingHit != nil {
		return existingHit, nil
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
	if isFirstVisit && record.OwnerID != nil {
		if err := s.lstepDeliveryTrigger.TriggerFirstVisitWelcome(ctx, record.ClinicID, *record.OwnerID); err != nil {
			slog.WarnContext(ctx, "first visit welcome trigger failed (non-fatal)", "owner_id", *record.OwnerID, "error", err)
		}
	}

	s.syncNextVisitTag(ctx, record.ClinicID, record)

	return record, nil
}

func (s *medicalRecordService) applyAppointmentContextForCreate(
	ctx context.Context,
	clinicID uint64,
	input *CreateMedicalRecordInput,
) error {
	if input == nil || input.AppointmentID == nil || s.reservationRepo == nil {
		return nil
	}
	if input.PetID == nil {
		return nil
	}
	appt, err := s.reservationRepo.FindByID(ctx, clinicID, *input.AppointmentID)
	if err != nil {
		return apperrors.Wrap(err, "failed to get appointment for medical record")
	}
	if appt == nil {
		return apperrors.WrapNotFound("appointment", "")
	}

	fields := map[string]any{}
	if err := resolveAppointmentUint64("pet_id", appt.PetID, &input.PetID, fields); err != nil {
		return err
	}
	if err := resolveAppointmentUint64("owner_id", appt.OwnerID, &input.OwnerID, fields); err != nil {
		return err
	}
	if err := validateReservationOwnerPetLinks(ctx, s.reservationRepo, clinicID, input.OwnerID, input.PetID); err != nil {
		return err
	}
	if input.DoctorID == nil && appt.DoctorID != nil {
		input.DoctorID = appt.DoctorID
	} else if input.DoctorID != nil && appt.DoctorID == nil {
		fields["doctor_id"] = *input.DoctorID
	}
	if input.Date.IsZero() {
		input.Date = appt.StartTime
	}
	if input.VisitType == "" {
		input.VisitType = appt.VisitType
	}

	if len(fields) == 0 {
		return nil
	}
	if _, err := s.reservationRepo.Update(ctx, clinicID, *input.AppointmentID, fields); err != nil {
		return apperrors.Wrap(err, "failed to update appointment for medical record")
	}
	return nil
}

func resolveAppointmentUint64(
	field string,
	appointmentValue *uint64,
	inputValue **uint64,
	fields map[string]any,
) error {
	if appointmentValue != nil {
		if *inputValue != nil && **inputValue != *appointmentValue {
			return apperrors.WrapInvalidInput(field + " does not match appointment")
		}
		if *inputValue == nil {
			*inputValue = appointmentValue
		}
		return nil
	}
	if *inputValue != nil {
		fields[field] = **inputValue
	}
	return nil
}

// validateMedicalRecordOwnerPetLinks はカルテ本体の Owner/Pet clinic 所有と整合を検証する（AUD-008）。
// reservationRepo 経由で AUD-001 の validateReservationOwnerPetLinks を再利用する。
func (s *medicalRecordService) validateMedicalRecordOwnerPetLinks(
	ctx context.Context,
	clinicID uint64,
	ownerID, petID *uint64,
) error {
	if ownerID == nil && petID == nil {
		return nil
	}
	if s.reservationRepo == nil {
		return nil
	}
	return validateReservationOwnerPetLinks(ctx, s.reservationRepo, clinicID, ownerID, petID)
}

// resolveFinalMedicalRecordOwnerPet は PATCH 入力と現在値から最終 Owner/Pet を求める（AUD-008）。
func resolveFinalMedicalRecordOwnerPet(existing *model.MedicalRecord, input UpdateMedicalRecordInput) (ownerID, petID *uint64) {
	ownerID = existing.OwnerID
	petID = existing.PetID
	if input.OwnerID != nil {
		ownerID = input.OwnerID
	}
	if input.PetID != nil {
		petID = input.PetID
	}
	return ownerID, petID
}

func (s *medicalRecordService) findExistingRecordByAppointment(
	ctx context.Context,
	clinicID uint64,
	record *model.MedicalRecord,
) *model.MedicalRecord {
	if record == nil || record.AppointmentID == nil || record.PetID == nil {
		return nil
	}
	dateStr := record.Date.Format(time.DateOnly)
	records, _, err := s.repo.FindAll(ctx, []uint64{clinicID}, repository.MedicalRecordListFilters{
		PetID:     record.PetID,
		StartDate: &dateStr,
		EndDate:   &dateStr,
	}, 1, 50)
	if err != nil {
		slog.WarnContext(ctx, "failed to check existing medical record by appointment",
			slog.Uint64("appointment_id", *record.AppointmentID),
			slog.String("error", err.Error()))
		return nil
	}
	for i := range records {
		if records[i].AppointmentID != nil && *records[i].AppointmentID == *record.AppointmentID {
			slog.InfoContext(ctx, "medical record create skipped because appointment already has a record",
				slog.Uint64("appointment_id", *record.AppointmentID),
				slog.Uint64("record_id", records[i].ID))
			return &records[i]
		}
	}
	return nil
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

	// AUD-008: Owner/Pet 変更時は最終状態で clinic 所有・Owner-Pet 整合を検証し、write と同一 tx にする。
	needsLinkValidation := input.OwnerID != nil || input.PetID != nil

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
		if needsLinkValidation {
			finalOwnerID, finalPetID := resolveFinalMedicalRecordOwnerPet(existing, input)
			if err := s.validateMedicalRecordOwnerPetLinks(txCtx, clinicID, finalOwnerID, finalPetID); err != nil {
				return err
			}
		}
		updated, err := s.repo.Update(txCtx, clinicID, id, fields, input.Version)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to update medical record", "error", err)
			return apperrors.Wrap(err, "failed to update medical record")
		}
		record = updated
		return nil
	}); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "medical record updated",
		slog.Uint64("record_id", id),
		slog.Uint64("clinic_id", clinicID))

	// 監査ログ: update / finalize（best-effort）
	if s.auditService != nil {
		oldDiff, newDiff := diffMedicalRecordImportantFields(existing, record)
		if oldDiff != nil {
			if err := s.auditService.LogMedicalRecordChange(ctx, clinicID, input.ActorID, "update", id, oldDiff, newDiff); err != nil {
				slog.ErrorContext(ctx, "audit log failed for medical record update", "error", err, "record_id", id)
			}
		}
		if isBecomingFinalized {
			newValue := extractMedicalRecordImportantFields(record)
			if err := s.auditService.LogMedicalRecordChange(ctx, clinicID, input.ActorID, "finalize", id, nil, newValue); err != nil {
				slog.ErrorContext(ctx, "audit log failed for medical record finalize", "error", err, "record_id", id)
			}
		}
	}

	if isBecomingFinalized {
		s.syncVisitCompletionTags(ctx, clinicID, record)
	}
	s.syncNextVisitTag(ctx, clinicID, record)

	return record, nil
}

func (s *medicalRecordService) Delete(ctx context.Context, clinicID, id uint64) error {
	existing, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to find medical record")
	}
	// Prevent deletion of finalized medical records (data integrity/legal compliance)
	if existing.Status == model.MedicalRecordStatusFinalized {
		return apperrors.WrapConflict("確定済みの診療記録は削除できません")
	}
	estimateCount, err := s.repo.CountEstimatesByMedicalRecordID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check estimate dependencies", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to check estimate dependencies")
	}
	if estimateCount > 0 {
		return apperrors.WrapConflict("この項目は使用中のため削除できません")
	}
	oldValue := extractMedicalRecordImportantFields(existing)
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to delete medical record", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to delete medical record")
	}
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
	if s.auditService != nil {
		oldDiff, newDiff := diffMedicalRecordImportantFields(existing, record)
		if oldDiff != nil {
			if err := s.auditService.LogMedicalRecordChange(ctx, clinicID, nil, "update", id, oldDiff, newDiff); err != nil {
				slog.ErrorContext(ctx, "audit log failed for recommendation_reason update", "error", err, "record_id", id)
			}
		}
	}

	return record, nil
}
