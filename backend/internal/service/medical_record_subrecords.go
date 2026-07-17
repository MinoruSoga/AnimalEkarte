package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
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
	if input.Plan != nil || input.Assessment != nil || input.Diagnosis1CategoryID != nil || input.Diagnosis1NameID != nil {
		if err := s.validateCreateSubRecordDiagnosisFKs(ctx, clinicID, input); err != nil {
			slog.WarnContext(ctx, "createSubRecords: failed to verify diagnosis FK ownership; skipping clinical plan update",
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
		if err := s.clinicalPlanRepo.Update(ctx, clinicID, plan.ID, fields); err != nil {
			slog.WarnContext(ctx, "createSubRecords: failed to update clinical plan",
				slog.Uint64("medical_record_id", recordID),
				slog.String("error", err.Error()))
		}
	}
}

// validateCreateSubRecordDiagnosisFKs は clinicalPlanService.validateDiagnosisFKs と同型の
// clinic-scoped 所有権検証を CreateSubRecords の best-effort パスに複製したもの（最小差分の
// ため ClinicalPlanService は注入しない）。非 nil の診断 FK のみ検証する。
func (s *medicalRecordService) validateCreateSubRecordDiagnosisFKs(ctx context.Context, clinicID uint64, input CreateSubRecordsInput) error {
	findDiagType := func(actx context.Context, cid, mid uint64) error {
		_, err := s.diagTypeRepo.FindByID(actx, cid, mid)
		return err
	}
	for _, typeID := range []*uint64{input.Diagnosis1CategoryID, input.Diagnosis2TypeID} {
		if err := validateOwnedMasterFK(ctx, "diagnosis type", clinicID, typeID, findDiagType); err != nil {
			return err
		}
	}
	findDiagName := func(actx context.Context, cid, mid uint64) error {
		_, err := s.diagNameRepo.FindByID(actx, cid, mid)
		return err
	}
	for _, nameID := range []*uint64{input.Diagnosis1NameID, input.Diagnosis2NameID} {
		if err := validateOwnedMasterFK(ctx, "diagnosis name", clinicID, nameID, findDiagName); err != nil {
			return err
		}
	}
	return nil
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
