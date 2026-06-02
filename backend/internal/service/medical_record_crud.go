package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *medicalRecordService) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.MedicalRecord, int64, error) {
	items, total, err := s.repo.FindAll(ctx, clinicID, petID, ownerID, startDate, endDate, page, limit)
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

func (s *medicalRecordService) CountByPetID(ctx context.Context, clinicID, petID uint64) (int64, error) {
	count, err := s.repo.CountByPetID(ctx, clinicID, petID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to count medical records by pet", "error", err)
		return 0, apperrors.Wrap(err, "failed to count medical records by pet")
	}
	return count, nil
}

func (s *medicalRecordService) Create(ctx context.Context, clinicID uint64, input *CreateMedicalRecordInput) (*model.MedicalRecord, error) {
	if err := s.applyAppointmentContextForCreate(ctx, clinicID, input); err != nil {
		return nil, apperrors.Wrap(err, "failed to apply appointment context for medical record")
	}

	record := buildMedicalRecordForCreate(clinicID, input)

	if existing := s.findExistingRecordByAppointment(ctx, clinicID, record); existing != nil {
		return existing, nil
	}

	// RecordNo が未設定の場合は service 層で自動生成する（handler 層に生成ロジックを置かない）
	if record.RecordNo == "" {
		record.RecordNo = generateRecordNo(record.Date, record.ClinicID)
	}

	// FEAT-382-2 supplement: whitelist validation for recommendation_reason
	if record.RecommendationReason != nil {
		if _, ok := allowedRecommendationReasons[*record.RecommendationReason]; !ok {
			return nil, apperrors.WrapInvalidInput(
				"recommendation_reason must be one of: revisit, checkup, prevention, exam",
			)
		}
	}

	// 初診判定: Reservation.VisitType 優先、AppointmentID なし or VisitType 空時は COUNT フォールバック（FEAT-383）
	var isFirstVisit bool
	if s.lstepDeliveryTrigger != nil && record.OwnerID != nil {
		visitType := s.resolveVisitTypeFromAppointment(ctx, record)
		switch {
		case visitType == string(model.VisitTypeFirst):
			isFirstVisit = true
		case visitType != "":
			isFirstVisit = false
		default:
			isFirstVisit = s.fallbackFirstVisitCheck(ctx, record.ClinicID, *record.OwnerID)
		}
	}

	if err := s.repo.Create(ctx, record); err != nil {
		slog.ErrorContext(ctx, "failed to create medical record", "error", err, "clinic_id", record.ClinicID)
		return nil, apperrors.Wrap(err, "failed to create medical record")
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

func (s *medicalRecordService) findExistingRecordByAppointment(
	ctx context.Context,
	clinicID uint64,
	record *model.MedicalRecord,
) *model.MedicalRecord {
	if record == nil || record.AppointmentID == nil || record.PetID == nil {
		return nil
	}
	dateStr := record.Date.Format("2006-01-02")
	records, _, err := s.repo.FindAll(ctx, clinicID, record.PetID, nil, &dateStr, &dateStr, 1, 50)
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

	// version が指定されている場合は一致確認
	if input.Version != nil && existing.Version != *input.Version {
		return nil, apperrors.WrapConflict("他のユーザーがこのカルテを変更しました。再読み込みしてください")
	}

	// owner_id 変更時: クリニック所属確認
	if input.OwnerID != nil {
		if _, err := s.ownerRepo.FindByID(ctx, clinicID, *input.OwnerID); err != nil {
			return nil, apperrors.WrapInvalidInput("owner not found in this clinic")
		}
	}

	// pet_id 変更時: クリニック所属確認
	if input.PetID != nil {
		if _, err := s.petRepo.FindByID(ctx, clinicID, *input.PetID); err != nil {
			return nil, apperrors.WrapInvalidInput("pet not found in this clinic")
		}
	}

	fields := buildMedicalRecordUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	// バージョンをインクリメント
	fields["version"] = existing.Version + 1

	// 監査 diff は更新前に取得（finalize 検出のため）
	wasFinalized := existing.Status == model.MedicalRecordStatusFinalized
	isBecomingFinalized := input.Status != nil && *input.Status == model.MedicalRecordStatusFinalized && !wasFinalized

	record, err := s.repo.Update(ctx, clinicID, id, fields)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update medical record", "error", err)
		return nil, apperrors.Wrap(err, "failed to update medical record")
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
	estimateCount, err := s.repo.CountEstimatesByMedicalRecordID(ctx, id)
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
	// 3. build update map
	var reasonValue any
	if input.Reason == "" {
		reasonValue = nil // SQL NULL
	} else {
		reasonValue = input.Reason
	}

	record, err := s.repo.Update(ctx, clinicID, id, map[string]any{
		colRecommendationReason: reasonValue,
	})
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
