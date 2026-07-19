package service

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

const (
	colClinicalPlanDiagnosis2TypeID = "diagnosis_2_type_id"
	colClinicalPlanDiagnosis2NameID = "diagnosis_2_name_id"
)

// UpdateClinicalPlanInput は診察所見・診断・治療方針更新の入力DTO（nil = 未送信フィールド）
// Diagnosis2TypeID / Diagnosis2NameID は **uint64: nil=未送信, &nil=NULLクリア, &&v=セット。
type UpdateClinicalPlanInput struct {
	PhysicalExam     *string
	DiagnosisTypeID  *uint64
	DiagnosisNameID  *uint64
	Diagnosis2TypeID **uint64
	Diagnosis2NameID **uint64
	DiagnosisDetails *string
	TreatmentPolicy  *string
	// Version は楽観的ロック用（nil=照合スキップ）。buildClinicalPlanUpdate では扱わず、
	// Update メソッド内で nextVersion 計算と repo.Update への expectedVersion 受け渡しに使う
	// （medical_record_crud.go の Update と同型）。
	Version *int
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
	if input.Diagnosis2TypeID != nil {
		fields[colClinicalPlanDiagnosis2TypeID] = *input.Diagnosis2TypeID
	}
	if input.Diagnosis2NameID != nil {
		fields[colClinicalPlanDiagnosis2NameID] = *input.Diagnosis2NameID
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

func optionalDoubleUint64(v **uint64) *uint64 {
	if v == nil {
		return nil
	}
	return *v
}

func (s *clinicalPlanService) validateDiagnosisFKs(ctx context.Context, clinicID uint64, input *UpdateClinicalPlanInput) error {
	findDiagType := func(actx context.Context, cid, mid uint64) error {
		_, err := s.diagTypeRepo.FindByID(actx, cid, mid)
		return err
	}
	for _, typeID := range []*uint64{input.DiagnosisTypeID, optionalDoubleUint64(input.Diagnosis2TypeID)} {
		if err := validateOwnedMasterFK(ctx, "diagnosis type", clinicID, typeID, findDiagType); err != nil {
			return err
		}
	}
	findDiagName := func(actx context.Context, cid, mid uint64) error {
		_, err := s.diagNameRepo.FindByID(actx, cid, mid)
		return err
	}
	for _, nameID := range []*uint64{input.DiagnosisNameID, optionalDoubleUint64(input.Diagnosis2NameID)} {
		if err := validateOwnedMasterFK(ctx, "diagnosis name", clinicID, nameID, findDiagName); err != nil {
			return err
		}
	}
	return nil
}

func (s *clinicalPlanService) validateDiagnosisTypeNameConsistency(ctx context.Context, clinicID uint64, plan *model.ClinicalPlan, input *UpdateClinicalPlanInput) error {
	// AUD-007: 第2診断スロットの type↔name 整合のみ強制（第1は既存部分PATCH互換を維持）。
	type2 := plan.Diagnosis2TypeID
	if input.Diagnosis2TypeID != nil {
		type2 = *input.Diagnosis2TypeID
	}
	name2 := plan.Diagnosis2NameID
	if input.Diagnosis2NameID != nil {
		name2 = *input.Diagnosis2NameID
	}
	return s.assertNameBelongsToType(ctx, clinicID, type2, name2, "diagnosis_2")
}

func (s *clinicalPlanService) assertNameBelongsToType(ctx context.Context, clinicID uint64, typeID, nameID *uint64, slot string) error {
	// 片方のみの部分PATCHは従来どおり許可。両方そろった最終状態でのみ type↔name 整合を強制する。
	if nameID == nil || typeID == nil {
		return nil
	}
	name, err := s.diagNameRepo.FindByID(ctx, clinicID, *nameID)
	if err != nil {
		return err
	}
	if name.DiagnosisTypeID != *typeID {
		return apperrors.WrapInvalidInput(slot + " name does not belong to the selected diagnosis type")
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
		if _, ownerErr := s.medRec.FindByID(ctx, clinicID, medicalRecordID); ownerErr != nil {
			if !apperrors.IsNotFound(ownerErr) {
				slog.ErrorContext(ctx, "failed to verify parent medical record", "error", ownerErr)
			}
			return nil, apperrors.Wrap(ownerErr, "failed to verify parent medical record")
		}
		plan = &model.ClinicalPlan{MedicalRecordID: medicalRecordID}
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

// assertParentDraft は親カルテが draft であることを検証する（SD-2 系ガード監査で発見された欠落）。
// examination/vital 等の子エンティティが使う lockDraftMedicalRecord とは異なり
// LockByIDForUpdate による行ロックは取らない（clinicalPlanService は Transactor を持たない設計上の
// 制約。cross_tenant_master_fk_write_test.go の既存呼び出しが repository.Transactor 無しの
// 旧コンストラクタシグネチャに依存しており変更できない）。単純な存在確認+ステータス確認であり、
// このチェックと後続の書込の間のレースは clinical_plan_repository.go の Update/Delete が
// medical_records.status='draft' を WHERE に含めることで原子的に閉じる。
func (s *clinicalPlanService) assertParentDraft(ctx context.Context, clinicID, medicalRecordID uint64, conflictMsg string) error {
	parent, err := s.medRec.FindByID(ctx, clinicID, medicalRecordID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find medical record", "error", err)
		return apperrors.Wrap(err, "failed to find medical record")
	}
	if parent.Status == model.MedicalRecordStatusFinalized {
		return apperrors.WrapConflict(conflictMsg)
	}
	return nil
}

func (s *clinicalPlanService) Update(ctx context.Context, clinicID, medicalRecordID uint64, input *UpdateClinicalPlanInput) (*model.ClinicalPlan, error) {
	plan, err := s.GetOrCreate(ctx, clinicID, medicalRecordID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get or create clinical plan", "error", err)
		return nil, apperrors.Wrap(err, "failed to get or create clinical plan")
	}
	if err := s.assertParentDraft(ctx, clinicID, medicalRecordID, "確定済みカルテの所見・診断を編集できません"); err != nil {
		return nil, err
	}
	if err := s.validateDiagnosisFKs(ctx, clinicID, input); err != nil {
		return nil, err
	}
	if err := s.validateDiagnosisTypeNameConsistency(ctx, clinicID, plan, input); err != nil {
		return nil, err
	}

	fields := buildClinicalPlanUpdate(input)
	if len(fields) == 0 {
		return plan, nil
	}
	// バージョンをインクリメント（input.Version 指定時はそれを起点にする。version 一致確認は
	// repo.Update の WHERE 述語に一本化したため、ここでの事前チェックは行わない。
	// medical_record_crud.go の Update と同型）
	nextVersion := plan.Version + 1
	if input.Version != nil {
		nextVersion = *input.Version + 1
	}
	fields["version"] = nextVersion
	if err := s.repo.Update(ctx, clinicID, plan.ID, fields, input.Version); err != nil {
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
	if err := s.assertParentDraft(ctx, clinicID, medicalRecordID, "確定済みカルテの所見・診断は削除できません"); err != nil {
		return err
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
