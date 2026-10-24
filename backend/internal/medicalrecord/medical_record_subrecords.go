package medicalrecord

import (
	"context"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// CreateSubRecords creates inquiry / clinical_plan subrecords for a medical record.
// MRC-04 / DEC-32 / DEC-35: failures return error so HTTP Create cannot claim full
// clinical success with a silent empty chief complaint / diagnosis payload.
func (s *medicalRecordService) CreateSubRecords(ctx context.Context, clinicID, recordID uint64, input CreateSubRecordsInput) error {
	// 1. inquiry: 入力がある場合のみ upsert する。
	// 既存 appointment の再オープン時に空入力で既存問診を上書きしない。
	if err := s.upsertInquirySubRecord(ctx, clinicID, recordID, input); err != nil {
		return err
	}
	return s.ensureClinicalPlanSubRecord(ctx, clinicID, recordID, input)
}

func (s *medicalRecordService) upsertInquirySubRecord(
	ctx context.Context,
	clinicID, recordID uint64,
	input CreateSubRecordsInput,
) error {
	if !hasInquirySubRecordInput(input) {
		return nil
	}
	if err := s.assertChiefComplaintTypeForSubRecords(ctx, clinicID, input.ChiefComplaintTypeID); err != nil {
		return err
	}
	// EMR-87: 未送信フィールドは nil のまま渡して既存値を保持する。
	// ChiefComplaintTypeID が送られた場合のみ &&v（値セット）として渡す。
	fields := InquiryUpsertFields{
		MedicalRecordID: recordID,
		ChiefComplaint:  input.ChiefComplaint,
		Notes:           input.Notes,
	}
	if input.ChiefComplaintTypeID != nil {
		fields.ChiefComplaintTypeID = &input.ChiefComplaintTypeID
	}
	if _, err := s.inquiryRepo.SaveByMedicalRecordID(ctx, clinicID, fields); err != nil {
		return apperrors.Wrap(err, "failed to upsert medical record inquiry")
	}
	return nil
}

func (s *medicalRecordService) ensureClinicalPlanSubRecord(
	ctx context.Context,
	clinicID, recordID uint64,
	input CreateSubRecordsInput,
) error {
	// 2. clinical_plan: 常に GetOrCreate で空レコードを確保し、フィールドがあれば更新
	plan, err := s.clinicalPlanRepo.FindByMedicalRecordID(ctx, clinicID, recordID)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			return apperrors.Wrap(err, "failed to find medical record clinical plan")
		}
		// clinical_plans.clinic_id は NOT NULL（001）。trusted な clinicID を必ず設定する（EMR-495 Finding 5）。
		plan = &model.ClinicalPlan{ClinicID: clinicID, MedicalRecordID: recordID}
		if err := s.clinicalPlanRepo.Create(ctx, plan); err != nil {
			return apperrors.Wrap(err, "failed to create medical record clinical plan")
		}
	}
	if input.Plan == nil && input.Assessment == nil && input.Diagnoses == nil {
		return nil
	}
	// MRC-14 / EMR-495 A: clinical plan Update と同じ診断行検査（件数・slot・マスタ所有権・type↔name 整合）。
	if err := validateDiagnosisRows(ctx, clinicID, input.Diagnoses, s.diagTypeRepo, s.diagNameRepo); err != nil {
		return apperrors.Wrap(err, "failed to verify diagnosis masters for medical record subrecords")
	}

	cmd := clinicalPlanUpdateFromSubRecords(input)
	// EMR-495 F-01: plan 本文列だけでなく診断行だけの置換でも、plan の version（FE の楽観ロックトークン）を
	// 同じ tx で進める。clinical_plans にはストレージ側の version trigger が無く、診断行の storage version は
	// 行単位で plan の version を動かさない。進めないと、旧 version を持つ別の編集者の PATCH が新しい診断行を
	// 黙って上書きできる。expectedVersion は直前に読んだ version で、読込と書込の間の競合は Conflict になる。
	expectedVersion := plan.Version
	nextVersion := expectedVersion + 1
	cmd.persistVersion = &nextVersion
	return s.withOptionalTx(ctx, func(txCtx context.Context) error {
		if err := s.clinicalPlanRepo.Update(txCtx, clinicID, plan.ID, cmd, &expectedVersion); err != nil {
			return apperrors.Wrap(err, "failed to update medical record clinical plan")
		}
		if input.Diagnoses == nil {
			return nil
		}
		if err := s.clinicalPlanRepo.ReplaceDiagnoses(txCtx, clinicID, plan.ID, input.Diagnoses); err != nil {
			return apperrors.Wrap(err, "failed to replace medical record clinical plan diagnoses")
		}
		return nil
	})
}

// withOptionalTx は tx 依存がある構成（本番）では fn を同一 tx で実行する。
// 依存が無い構成（単体テストの最小組立）では ctx のまま実行する。
func (s *medicalRecordService) withOptionalTx(ctx context.Context, fn func(context.Context) error) error {
	if s.tx == nil {
		return fn(ctx)
	}
	return s.tx.WithTx(ctx, fn)
}

func (s *medicalRecordService) assertChiefComplaintTypeForSubRecords(
	ctx context.Context,
	clinicID uint64,
	typeID *uint64,
) error {
	if typeID == nil {
		return nil
	}
	if _, err := s.chiefComplaintTypeRepo.FindByID(ctx, clinicID, *typeID); err != nil {
		return apperrors.Wrap(err, "failed to verify chief complaint type for medical record subrecords")
	}
	return nil
}

func clinicalPlanUpdateFromSubRecords(input CreateSubRecordsInput) UpdateClinicalPlanInput {
	return UpdateClinicalPlanInput{
		TreatmentPolicy:  input.Plan,
		DiagnosisDetails: input.Assessment,
	}
}

func hasInquirySubRecordInput(input CreateSubRecordsInput) bool {
	return input.ChiefComplaintTypeID != nil ||
		input.ChiefComplaint != nil ||
		input.Notes != nil
}
