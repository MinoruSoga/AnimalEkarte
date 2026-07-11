package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// UpdateClinicalPlanInput は診察所見・診断・治療方針更新の入力DTO（nil = 未送信フィールド）
type UpdateClinicalPlanInput struct {
	PhysicalExam         *string
	DiagnosisTypeID      *uint64
	DiagnosisNameID      *uint64
	Diagnosis2CategoryID *uint64
	Diagnosis2NameID     *uint64
	DiagnosisDetails     *string
	TreatmentPolicy      *string
}

func buildClinicalPlanUpdate(input *UpdateClinicalPlanInput) map[string]any {
	fields := map[string]any{}
	if input.PhysicalExam != nil {
		fields["physical_exam"] = *input.PhysicalExam
	}
	if input.DiagnosisTypeID != nil {
		fields["diagnosis_type_id"] = *input.DiagnosisTypeID
	}
	if input.DiagnosisNameID != nil {
		fields["diagnosis_name_id"] = *input.DiagnosisNameID
	}
	if input.Diagnosis2CategoryID != nil {
		fields["diagnosis_2_category_id"] = *input.Diagnosis2CategoryID
	}
	if input.Diagnosis2NameID != nil {
		fields["diagnosis_2_name_id"] = *input.Diagnosis2NameID
	}
	if input.DiagnosisDetails != nil {
		fields["diagnosis_details"] = *input.DiagnosisDetails
	}
	if input.TreatmentPolicy != nil {
		fields["treatment_policy"] = *input.TreatmentPolicy
	}
	return fields
}

// ClinicalPlanService は診察所見・診断・治療方針のビジネスロジックインターフェース
type ClinicalPlanService interface {
	GetOrCreate(ctx context.Context, clinicID, medicalRecordID uint64) (*model.ClinicalPlan, error)
	Update(ctx context.Context, clinicID, medicalRecordID uint64, input *UpdateClinicalPlanInput) (*model.ClinicalPlan, error)
	Delete(ctx context.Context, clinicID, medicalRecordID uint64) error
}

type clinicalPlanService struct {
	repo         repository.ClinicalPlanRepository
	medRec       repository.MedicalRecordRepository
	diagTypeRepo repository.DiagnosisTypeRepository
	diagNameRepo repository.DiagnosisNameRepository
}

// NewClinicalPlanService はClinicalPlanServiceを初期化して返す
func NewClinicalPlanService(repo repository.ClinicalPlanRepository, medRec repository.MedicalRecordRepository, diagTypeRepo repository.DiagnosisTypeRepository, diagNameRepo repository.DiagnosisNameRepository) ClinicalPlanService {
	return &clinicalPlanService{repo: repo, medRec: medRec, diagTypeRepo: diagTypeRepo, diagNameRepo: diagNameRepo}
}

// validateDiagnosisFKs は request 由来の clinic-scoped 診断マスタFK
// (diagnosis_type ×2 / diagnosis_name ×2) の所有権を検証する。別 clinic の診断分類を
// 患者の診断記録へ縫い付ける cross-tenant mislink を NotFound で遮断する。
func (s *clinicalPlanService) validateDiagnosisFKs(ctx context.Context, clinicID uint64, input *UpdateClinicalPlanInput) error {
	findDiagType := func(actx context.Context, cid, mid uint64) error {
		_, err := s.diagTypeRepo.FindByID(actx, cid, mid)
		return err
	}
	for _, typeID := range []*uint64{input.DiagnosisTypeID, input.Diagnosis2CategoryID} {
		if err := validateOwnedMasterFK(ctx, "diagnosis type", clinicID, typeID, findDiagType); err != nil {
			return err
		}
	}
	findDiagName := func(actx context.Context, cid, mid uint64) error {
		_, err := s.diagNameRepo.FindByID(actx, cid, mid)
		return err
	}
	for _, nameID := range []*uint64{input.DiagnosisNameID, input.Diagnosis2NameID} {
		if err := validateOwnedMasterFK(ctx, "diagnosis name", clinicID, nameID, findDiagName); err != nil {
			return err
		}
	}
	return nil
}

func (s *clinicalPlanService) GetOrCreate(ctx context.Context, clinicID, medicalRecordID uint64) (*model.ClinicalPlan, error) {
	plan, err := s.repo.FindByMedicalRecordID(ctx, clinicID, medicalRecordID)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			slog.ErrorContext(ctx, "failed to get clinical plan", "error", err)
			return nil, apperrors.Wrap(err, "failed to get clinical plan")
		}
		// テナント所有権検証（クロステナント write 防止）。
		// clinical_plans は自前 clinic_id を持たず medical_records 経由で隔離するため、
		// 自動作成前に親カルテの所有権を明示検証する（NotFound 分岐は越境ケースと重なる）。
		if _, ownerErr := s.medRec.FindByID(ctx, clinicID, medicalRecordID); ownerErr != nil {
			if !apperrors.IsNotFound(ownerErr) {
				slog.ErrorContext(ctx, "failed to verify parent medical record", "error", ownerErr)
			}
			return nil, apperrors.Wrap(ownerErr, "failed to verify parent medical record")
		}
		// 存在しない場合は空レコードを自動作成
		plan = &model.ClinicalPlan{
			MedicalRecordID: medicalRecordID,
		}
		if err := s.repo.Create(ctx, plan); err != nil {
			slog.ErrorContext(ctx, "failed to create clinical plan", "error", err)
			return nil, apperrors.Wrap(err, "failed to create clinical plan")
		}
		slog.InfoContext(ctx, "clinical_plan created",
			slog.Uint64("clinic_id", clinicID),
			slog.Uint64("clinical_plan_id", plan.ID),
			slog.Uint64("medical_record_id", medicalRecordID))
		return plan, nil
	}
	return plan, nil
}

func (s *clinicalPlanService) Update(ctx context.Context, clinicID, medicalRecordID uint64, input *UpdateClinicalPlanInput) (*model.ClinicalPlan, error) {
	plan, err := s.GetOrCreate(ctx, clinicID, medicalRecordID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get or create clinical plan", "error", err)
		return nil, apperrors.Wrap(err, "failed to get or create clinical plan")
	}
	// クロステナント write 防止: 貼り替え先の診断マスタFK所有権を検証する。
	if err := s.validateDiagnosisFKs(ctx, clinicID, input); err != nil {
		return nil, err
	}

	fields := buildClinicalPlanUpdate(input)
	if len(fields) == 0 {
		// 全フィールドが未指定の場合は no-op として現在のレコードをそのまま返す
		return plan, nil
	}
	if err := s.repo.Update(ctx, clinicID, plan.ID, fields); err != nil {
		slog.ErrorContext(ctx, "failed to update clinical plan", "error", err)
		return nil, apperrors.Wrap(err, "failed to update clinical plan")
	}
	slog.InfoContext(ctx, "clinical_plan updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("clinical_plan_id", plan.ID),
		slog.Uint64("medical_record_id", medicalRecordID))
	updated, err := s.repo.FindByMedicalRecordID(ctx, clinicID, medicalRecordID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get updated clinical plan", "error", err)
		return nil, apperrors.Wrap(err, "failed to get updated clinical plan")
	}
	return updated, nil
}

func (s *clinicalPlanService) Delete(ctx context.Context, clinicID, medicalRecordID uint64) error {
	plan, err := s.repo.FindByMedicalRecordID(ctx, clinicID, medicalRecordID)
	if err != nil {
		return apperrors.Wrap(err, "failed to get clinical plan")
	}
	if err := s.repo.Delete(ctx, clinicID, plan.ID); err != nil {
		slog.ErrorContext(ctx, "failed to delete clinical plan", "error", err, "clinic_id", clinicID, "clinical_plan_id", plan.ID)
		return apperrors.Wrap(err, "failed to delete clinical plan")
	}
	slog.InfoContext(ctx, "clinical_plan deleted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("clinical_plan_id", plan.ID),
		slog.Uint64("medical_record_id", medicalRecordID))
	return nil
}

var _ ClinicalPlanService = (*clinicalPlanService)(nil)
