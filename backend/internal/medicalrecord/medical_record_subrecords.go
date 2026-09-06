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
	if err := s.assertChiefComplaintTypeForSubRecords(ctx, clinicID, recordID, input.ChiefComplaintTypeID); err != nil {
		return err
	}
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
		plan = &model.ClinicalPlan{MedicalRecordID: recordID}
		if err := s.clinicalPlanRepo.Create(ctx, plan); err != nil {
			return apperrors.Wrap(err, "failed to create medical record clinical plan")
		}
	}
	if input.Plan == nil && input.Assessment == nil && input.Diagnosis1CategoryID == nil && input.Diagnosis1NameID == nil &&
		input.Diagnosis2TypeID == nil && input.Diagnosis2NameID == nil {
		return nil
	}
	// MRC-14: clinical plan と同じ validateDiagnosisMasterFKs / assertDiagnosisNameBelongsToType を使う。
	if err := validateDiagnosisMasterFKs(
		ctx,
		clinicID,
		[]*uint64{input.Diagnosis1CategoryID, input.Diagnosis2TypeID},
		[]*uint64{input.Diagnosis1NameID, input.Diagnosis2NameID},
		s.diagTypeRepo,
		s.diagNameRepo,
	); err != nil {
		return apperrors.Wrap(err, "failed to verify diagnosis masters for medical record subrecords")
	}
	// AUD-007: 第2診断 type↔name 整合（clinical plan Update と同契約）
	if err := assertDiagnosisNameBelongsToType(
		ctx, clinicID, input.Diagnosis2TypeID, input.Diagnosis2NameID, "diagnosis_2", s.diagNameRepo,
	); err != nil {
		return err
	}

	cmd := clinicalPlanUpdateFromSubRecords(input)
	if err := s.clinicalPlanRepo.Update(ctx, clinicID, plan.ID, cmd, nil); err != nil {
		return apperrors.Wrap(err, "failed to update medical record clinical plan")
	}
	return nil
}

func (s *medicalRecordService) assertChiefComplaintTypeForSubRecords(
	ctx context.Context,
	clinicID, recordID uint64,
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
	cmd := UpdateClinicalPlanInput{
		TreatmentPolicy:  input.Plan,
		DiagnosisDetails: input.Assessment,
		DiagnosisTypeID:  input.Diagnosis1CategoryID,
		DiagnosisNameID:  input.Diagnosis1NameID,
	}
	if input.Diagnosis2TypeID != nil {
		id := input.Diagnosis2TypeID
		cmd.Diagnosis2TypeID = &id
	}
	if input.Diagnosis2NameID != nil {
		id := input.Diagnosis2NameID
		cmd.Diagnosis2NameID = &id
	}
	return cmd
}

func hasInquirySubRecordInput(input CreateSubRecordsInput) bool {
	return input.ChiefComplaintTypeID != nil ||
		input.ChiefComplaint != nil ||
		input.Notes != nil
}
