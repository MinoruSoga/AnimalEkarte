package medicalrecord

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *medicalRecordService) CreateSubRecords(ctx context.Context, clinicID, recordID uint64, input CreateSubRecordsInput) {
	// 1. inquiry: 入力がある場合のみ upsert する。
	// 既存 appointment の再オープン時に空入力で既存問診を上書きしない。
	if hasInquirySubRecordInput(input) {
		skipInquiry := false
		if input.ChiefComplaintTypeID != nil {
			if _, err := s.chiefComplaintTypeRepo.FindByID(ctx, clinicID, *input.ChiefComplaintTypeID); err != nil {
				slog.WarnContext(ctx, "createSubRecords: failed to verify chief complaint type ownership; skipping inquiry upsert",
					slog.Uint64("medical_record_id", recordID),
					slog.Uint64("chief_complaint_type_id", *input.ChiefComplaintTypeID),
					slog.String("error", err.Error()))
				skipInquiry = true
			}
		}
		if !skipInquiry {
			inquiry := &model.Inquiry{
				MedicalRecordID: recordID,
			}
			if input.ChiefComplaintTypeID != nil {
				inquiry.ChiefComplaintTypeID = input.ChiefComplaintTypeID
			}
			if input.ChiefComplaint != nil {
				inquiry.ChiefComplaint = *input.ChiefComplaint
			}
			if input.Notes != nil {
				inquiry.Notes = *input.Notes
			}
			if _, err := s.inquiryRepo.SaveByMedicalRecordID(ctx, clinicID, inquiry); err != nil {
				slog.WarnContext(ctx, "createSubRecords: failed to upsert inquiry",
					slog.Uint64("medical_record_id", recordID),
					slog.String("error", err.Error()))
			}
		}
	}

	// 2. clinical_plan: 常に GetOrCreate で空レコードを確保し、フィールドがあれば更新
	plan, err := s.clinicalPlanRepo.FindByMedicalRecordID(ctx, clinicID, recordID)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			slog.WarnContext(ctx, "createSubRecords: failed to find clinical plan",
				slog.Uint64("medical_record_id", recordID),
				slog.String("error", err.Error()))
			return
		}
		plan = &model.ClinicalPlan{MedicalRecordID: recordID}
		if err := s.clinicalPlanRepo.Create(ctx, plan); err != nil {
			slog.WarnContext(ctx, "createSubRecords: failed to create clinical plan",
				slog.Uint64("medical_record_id", recordID),
				slog.String("error", err.Error()))
			return
		}
	}
	if input.Plan != nil || input.Assessment != nil || input.Diagnosis1CategoryID != nil || input.Diagnosis1NameID != nil ||
		input.Diagnosis2TypeID != nil || input.Diagnosis2NameID != nil {
		// MRC-14: clinical plan と同じ validateDiagnosisMasterFKs / assertDiagnosisNameBelongsToType を使う。
		if err := validateDiagnosisMasterFKs(
			ctx,
			clinicID,
			[]*uint64{input.Diagnosis1CategoryID, input.Diagnosis2TypeID},
			[]*uint64{input.Diagnosis1NameID, input.Diagnosis2NameID},
			s.diagTypeRepo,
			s.diagNameRepo,
		); err != nil {
			slog.WarnContext(ctx, "createSubRecords: failed to verify diagnosis FK ownership; skipping clinical plan update",
				slog.Uint64("medical_record_id", recordID),
				slog.String("error", err.Error()))
			return
		}
		// AUD-007: 第2診断 type↔name 整合（clinical plan Update と同契約）
		if err := assertDiagnosisNameBelongsToType(
			ctx, clinicID, input.Diagnosis2TypeID, input.Diagnosis2NameID, "diagnosis_2", s.diagNameRepo,
		); err != nil {
			slog.WarnContext(ctx, "createSubRecords: diagnosis_2 type/name mismatch; skipping clinical plan update",
				slog.Uint64("medical_record_id", recordID),
				slog.String("error", err.Error()))
			return
		}

		fields := map[string]any{}
		if input.Plan != nil {
			fields["treatment_policy"] = *input.Plan
		}
		if input.Assessment != nil {
			fields["diagnosis_details"] = *input.Assessment
		}
		if input.Diagnosis1CategoryID != nil {
			fields["diagnosis_type_id"] = *input.Diagnosis1CategoryID
		}
		if input.Diagnosis1NameID != nil {
			fields["diagnosis_name_id"] = *input.Diagnosis1NameID
		}
		if input.Diagnosis2TypeID != nil {
			fields["diagnosis_2_type_id"] = *input.Diagnosis2TypeID
		}
		if input.Diagnosis2NameID != nil {
			fields["diagnosis_2_name_id"] = *input.Diagnosis2NameID
		}
		if err := s.clinicalPlanRepo.Update(ctx, clinicID, plan.ID, fields, nil); err != nil {
			slog.WarnContext(ctx, "createSubRecords: failed to update clinical plan",
				slog.Uint64("medical_record_id", recordID),
				slog.String("error", err.Error()))
		}
	}
}

func hasInquirySubRecordInput(input CreateSubRecordsInput) bool {
	return input.ChiefComplaintTypeID != nil ||
		input.ChiefComplaint != nil ||
		input.Notes != nil
}

// AutoCreateFromReservation は予約ステータスが「受付済み」に変わったときカルテを best-effort で自動作成する。
// 同日同ペットのカルテが既に存在する場合はスキップする（重複防止）。
// LINE予約で owner_id / pet_id が未設定の場合は line_customer から補完を試みる（BUG-386）。
// 失敗してもメイン処理（予約更新）には影響しない。
