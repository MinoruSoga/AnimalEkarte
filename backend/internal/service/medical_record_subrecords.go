package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *medicalRecordService) CreateSubRecords(ctx context.Context, clinicID, recordID uint64, input CreateSubRecordsInput) {
	// 1. inquiry: フィールドの有無に関わらず常に upsert（空でも OK）
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
	if input.Plan != nil || input.Assessment != nil || input.Diagnosis1CategoryID != nil || input.Diagnosis1NameID != nil {
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
		if input.Diagnosis2CategoryID != nil {
			fields["diagnosis_2_category_id"] = *input.Diagnosis2CategoryID
		}
		if input.Diagnosis2NameID != nil {
			fields["diagnosis_2_name_id"] = *input.Diagnosis2NameID
		}
		if err := s.clinicalPlanRepo.Update(ctx, clinicID, plan.ID, fields); err != nil {
			slog.WarnContext(ctx, "createSubRecords: failed to update clinical plan",
				slog.Uint64("medical_record_id", recordID),
				slog.String("error", err.Error()))
		}
	}
}

// AutoCreateFromReservation は予約ステータスが「受付済み」に変わったときカルテを best-effort で自動作成する。
// 同日同ペットのカルテが既に存在する場合はスキップする（重複防止）。
// LINE予約で owner_id / pet_id が未設定の場合は line_customer から補完を試みる（BUG-386）。
// 失敗してもメイン処理（予約更新）には影響しない。
