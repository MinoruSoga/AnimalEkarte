package medicalrecord

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
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
	// ActorID は監査用の認証済み staff（JSON 非公開。handler が JWT から注入）。
	// examination parent mutation と同型に必須（nil/0 は fail-closed）。
	ActorID *uint64
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
	// Delete は staff actor 必須。削除前値の audit を同一 DBOrTx で fail-closed に書く（BUG-010 residual）。
	Delete(ctx context.Context, clinicID, medicalRecordID uint64, actorID *uint64) error
}

type clinicalPlanService struct {
	repo         ClinicalPlanRepository
	medRec       medicalRecordFinder
	diagTypeRepo DiagnosisTypeRepository
	diagNameRepo DiagnosisNameRepository
	transactor   Transactor
	auditTx      AuditTxLogger
}

// NewClinicalPlanService はClinicalPlanServiceを初期化して返す。
// BUG-010 residual: transactor と auditTx は Update の write+audit 原子性に必須
// （MRA-01 / care_plan_item と同型の fail-closed）。nil の場合 Update は拒否する。
func NewClinicalPlanService(
	repo ClinicalPlanRepository,
	medRec medicalRecordFinder,
	diagTypeRepo DiagnosisTypeRepository,
	diagNameRepo DiagnosisNameRepository,
	transactor Transactor,
	auditTx AuditTxLogger,
) ClinicalPlanService {
	return &clinicalPlanService{
		repo:         repo,
		medRec:       medRec,
		diagTypeRepo: diagTypeRepo,
		diagNameRepo: diagNameRepo,
		transactor:   transactor,
		auditTx:      auditTx,
	}
}

func optionalDoubleUint64(v **uint64) *uint64 {
	if v == nil {
		return nil
	}
	return *v
}

// validateDiagnosisMasterFKs は診断 type/name マスタの clinic 所有権を単一経路で検証する
// （BE-refactor.md MRC-14 / X-09: clinical plan と CreateSubRecords の copy-paste drift 解消）。
func validateDiagnosisMasterFKs(
	ctx context.Context,
	clinicID uint64,
	typeIDs []*uint64,
	nameIDs []*uint64,
	diagTypeRepo DiagnosisTypeRepository,
	diagNameRepo DiagnosisNameRepository,
) error {
	findDiagType := func(actx context.Context, cid, mid uint64) error {
		_, err := diagTypeRepo.FindByID(actx, cid, mid)
		return err
	}
	for _, typeID := range typeIDs {
		if err := validateOwnedMasterFK(ctx, "diagnosis type", clinicID, typeID, findDiagType); err != nil {
			return err
		}
	}
	findDiagName := func(actx context.Context, cid, mid uint64) error {
		_, err := diagNameRepo.FindByID(actx, cid, mid)
		return err
	}
	for _, nameID := range nameIDs {
		if err := validateOwnedMasterFK(ctx, "diagnosis name", clinicID, nameID, findDiagName); err != nil {
			return err
		}
	}
	return nil
}

// assertDiagnosisNameBelongsToType は type↔name 整合を単一経路で検証する（AUD-007）。
// 片方のみの部分更新は許可し、両方そろった最終状態でのみ type↔name を強制する。
func assertDiagnosisNameBelongsToType(
	ctx context.Context,
	clinicID uint64,
	typeID, nameID *uint64,
	slot string,
	diagNameRepo DiagnosisNameRepository,
) error {
	if nameID == nil || typeID == nil {
		return nil
	}
	name, err := diagNameRepo.FindByID(ctx, clinicID, *nameID)
	if err != nil {
		return err
	}
	if name.DiagnosisTypeID != *typeID {
		return apperrors.WrapInvalidInput(slot + " name does not belong to the selected diagnosis type")
	}
	return nil
}

func (s *clinicalPlanService) validateDiagnosisFKs(ctx context.Context, clinicID uint64, input *UpdateClinicalPlanInput) error {
	return validateDiagnosisMasterFKs(
		ctx,
		clinicID,
		[]*uint64{input.DiagnosisTypeID, optionalDoubleUint64(input.Diagnosis2TypeID)},
		[]*uint64{input.DiagnosisNameID, optionalDoubleUint64(input.Diagnosis2NameID)},
		s.diagTypeRepo,
		s.diagNameRepo,
	)
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
	return assertDiagnosisNameBelongsToType(ctx, clinicID, type2, name2, "diagnosis_2", s.diagNameRepo)
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
// LockByIDForUpdate による行ロックは取らない。単純な存在確認+ステータス確認であり、
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

func clinicalPlanAuditValue(p *model.ClinicalPlan) map[string]any {
	if p == nil {
		return nil
	}
	return map[string]any{
		"id":                  p.ID,
		"medical_record_id":   p.MedicalRecordID,
		"physical_exam":       p.PhysicalExam,
		"diagnosis_type_id":   p.DiagnosisTypeID,
		"diagnosis_name_id":   p.DiagnosisNameID,
		"diagnosis_2_type_id": p.Diagnosis2TypeID,
		"diagnosis_2_name_id": p.Diagnosis2NameID,
		"diagnosis_details":   p.DiagnosisDetails,
		"treatment_policy":    p.TreatmentPolicy,
		"version":             p.Version,
	}
}

func (s *clinicalPlanService) auditUpdateTx(ctx context.Context, clinicID uint64, actorID *uint64, before, after *model.ClinicalPlan) error {
	// MRA-01 / BUG-010 residual: clinical legal fields require fail-closed audit in the same tx.
	if s.auditTx == nil {
		return apperrors.WrapInternalServerError("clinical plan audit dependency is required")
	}
	if actorID == nil || *actorID == 0 {
		return apperrors.WrapInvalidInput("authenticated staff actor is required for clinical plan mutation")
	}
	resourceID := uint64(0)
	if after != nil {
		resourceID = after.ID
	} else if before != nil {
		resourceID = before.ID
	}
	entry := &AuditEntry{
		ClinicID:   &clinicID,
		ActorID:    actorID,
		ActorType:  model.AuditActorTypeStaff,
		Action:     model.AuditActionClinicalPlanUpdate,
		Resource:   model.AuditResourceClinicalPlan,
		ResourceID: &resourceID,
		OldValue:   clinicalPlanAuditValue(before),
		NewValue:   clinicalPlanAuditValue(after),
		Metadata: map[string]any{
			"operation_type": "update",
		},
	}
	if err := s.auditTx.LogEntryTx(ctx, entry); err != nil {
		slog.ErrorContext(ctx, "failed to audit clinical plan update",
			"error", err,
			"clinic_id", clinicID,
			"clinical_plan_id", resourceID,
		)
		return apperrors.Wrap(err, "failed to write clinical plan update audit")
	}
	return nil
}

func (s *clinicalPlanService) auditDeleteTx(ctx context.Context, clinicID uint64, actorID *uint64, before *model.ClinicalPlan) error {
	// MRA-01 / BUG-010 residual: destructive clinical-plan delete requires fail-closed audit
	// with pre-delete field values in the same tx (care_plan_item.auditDeleteTx と同型).
	if s.auditTx == nil {
		return apperrors.WrapInternalServerError("clinical plan audit dependency is required")
	}
	if actorID == nil || *actorID == 0 {
		return apperrors.WrapInvalidInput("authenticated staff actor is required for clinical plan mutation")
	}
	resourceID := uint64(0)
	if before != nil {
		resourceID = before.ID
	}
	entry := &AuditEntry{
		ClinicID:   &clinicID,
		ActorID:    actorID,
		ActorType:  model.AuditActorTypeStaff,
		Action:     model.AuditActionClinicalPlanDelete,
		Resource:   model.AuditResourceClinicalPlan,
		ResourceID: &resourceID,
		OldValue:   clinicalPlanAuditValue(before),
		NewValue:   nil,
		Metadata: map[string]any{
			"operation_type": "delete",
		},
	}
	if err := s.auditTx.LogEntryTx(ctx, entry); err != nil {
		slog.ErrorContext(ctx, "failed to audit clinical plan delete",
			"error", err,
			"clinic_id", clinicID,
			"clinical_plan_id", resourceID,
		)
		return apperrors.Wrap(err, "failed to write clinical plan delete audit")
	}
	return nil
}

func (s *clinicalPlanService) withTx(ctx context.Context, fn func(context.Context) error) error {
	if s.transactor == nil {
		return apperrors.WrapInternalServerError("clinical plan transaction dependency is required")
	}
	return s.transactor.WithTx(ctx, fn)
}

func (s *clinicalPlanService) Update(ctx context.Context, clinicID, medicalRecordID uint64, input *UpdateClinicalPlanInput) (*model.ClinicalPlan, error) {
	// Fail-closed before any mutation: missing audit/tx deps must not allow silent writes.
	if s.auditTx == nil {
		return nil, apperrors.WrapInternalServerError("clinical plan audit dependency is required")
	}
	if s.transactor == nil {
		return nil, apperrors.WrapInternalServerError("clinical plan transaction dependency is required")
	}

	plan, err := s.GetOrCreate(ctx, clinicID, medicalRecordID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get or create clinical plan", "error", err)
		return nil, apperrors.Wrap(err, "failed to get or create clinical plan")
	}
	if err := s.assertParentDraft(ctx, clinicID, medicalRecordID, "確定済みカルテの所見・診断を編集できません"); err != nil {
		return nil, err
	}
	if input.ActorID == nil || *input.ActorID == 0 {
		return nil, apperrors.WrapInvalidInput("authenticated staff actor is required for clinical plan mutation")
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

	// Snapshot before mutation for audit OldValue (do not mutate plan pointer fields after).
	before := *plan

	var updated *model.ClinicalPlan
	if err := s.withTx(ctx, func(txCtx context.Context) error {
		if err := s.repo.Update(txCtx, clinicID, plan.ID, fields, input.Version); err != nil {
			slog.ErrorContext(txCtx, "failed to update clinical plan", "error", err)
			return apperrors.Wrap(err, "failed to update clinical plan")
		}
		reloaded, err := s.repo.FindByMedicalRecordID(txCtx, clinicID, medicalRecordID)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to get updated clinical plan", "error", err)
			return apperrors.Wrap(err, "failed to get updated clinical plan")
		}
		if err := s.auditUpdateTx(txCtx, clinicID, input.ActorID, &before, reloaded); err != nil {
			return err
		}
		updated = reloaded
		return nil
	}); err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "clinical_plan updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("clinical_plan_id", plan.ID),
		slog.Uint64("medical_record_id", medicalRecordID))
	return updated, nil
}

func (s *clinicalPlanService) Delete(ctx context.Context, clinicID, medicalRecordID uint64, actorID *uint64) error {
	// Fail-closed before any mutation: missing audit/tx deps must not allow silent deletes.
	if s.auditTx == nil {
		return apperrors.WrapInternalServerError("clinical plan audit dependency is required")
	}
	if s.transactor == nil {
		return apperrors.WrapInternalServerError("clinical plan transaction dependency is required")
	}

	plan, err := s.repo.FindByMedicalRecordID(ctx, clinicID, medicalRecordID)
	if err != nil {
		return apperrors.Wrap(err, "failed to get clinical plan")
	}
	if err := s.assertParentDraft(ctx, clinicID, medicalRecordID, "確定済みカルテの所見・診断は削除できません"); err != nil {
		return err
	}
	if actorID == nil || *actorID == 0 {
		return apperrors.WrapInvalidInput("authenticated staff actor is required for clinical plan mutation")
	}

	// Snapshot pre-delete values for audit OldValue (GORM soft-delete; durable recovery of
	// clinical field content for legal audit is this OldValue, not the tombstoned row alone).
	before := *plan
	if err := s.withTx(ctx, func(txCtx context.Context) error {
		if err := s.repo.Delete(txCtx, clinicID, plan.ID); err != nil {
			slog.ErrorContext(txCtx, "failed to delete clinical plan", "error", err, "clinic_id", clinicID, "clinical_plan_id", plan.ID)
			return apperrors.Wrap(err, "failed to delete clinical plan")
		}
		return s.auditDeleteTx(txCtx, clinicID, actorID, &before)
	}); err != nil {
		return err
	}

	slog.InfoContext(ctx, "clinical_plan deleted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("clinical_plan_id", plan.ID),
		slog.Uint64("medical_record_id", medicalRecordID))
	return nil
}

var _ ClinicalPlanService = (*clinicalPlanService)(nil)
