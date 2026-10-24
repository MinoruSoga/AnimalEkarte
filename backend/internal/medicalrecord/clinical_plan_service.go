package medicalrecord

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// UpdateClinicalPlanInput は診察所見・診断・治療方針更新の入力DTO（nil = 未送信フィールド）
// Diagnoses は EMR-495 A の診断行集合: nil=未送信（診断行は変更しない）、空スライス=全行クリア、
// 値あり=その集合へ置換（slot 1/2・各 slot 複数行可）。
type UpdateClinicalPlanInput struct {
	PhysicalExam     *string
	Diagnoses        *[]ClinicalPlanDiagnosisInput
	DiagnosisDetails *string
	TreatmentPolicy  *string
	// Version は楽観的ロック用（nil=照合スキップ）。buildClinicalPlanUpdate では扱わず、
	// Update メソッド内で nextVersion 計算と repo.Update への expectedVersion 受け渡しに使う
	// （medical_record_crud.go の Update と同型）。診断行だけの更新でも plan の version を進める。
	Version *int
	// ActorID は監査用の認証済み staff（JSON 非公開。handler が JWT から注入）。
	// examination parent mutation と同型に必須（nil/0 は fail-closed）。
	ActorID *uint64
	// persistVersion is set only by service after next-version calculation. HTTP/client must not set it.
	persistVersion *int
}

// hasChanges は plan 本文列または診断行の更新が要求されているかを返す。
func (i *UpdateClinicalPlanInput) hasChanges() bool {
	return i.PhysicalExam != nil || i.DiagnosisDetails != nil || i.TreatmentPolicy != nil || i.Diagnoses != nil
}

func buildClinicalPlanUpdate(input *UpdateClinicalPlanInput) map[string]any {
	fields := map[string]any{}
	if input.PhysicalExam != nil {
		fields["physical_exam"] = *input.PhysicalExam
	}
	if input.DiagnosisDetails != nil {
		fields["diagnosis_details"] = *input.DiagnosisDetails
	}
	if input.TreatmentPolicy != nil {
		fields["treatment_policy"] = *input.TreatmentPolicy
	}
	if input.persistVersion != nil {
		fields["version"] = *input.persistVersion
	}
	return fields
}

// ClinicalPlanService は診察所見・診断・治療方針のビジネスロジックインターフェース
type ClinicalPlanService interface {
	GetOrCreate(ctx context.Context, clinicID, medicalRecordID uint64) (*ClinicalPlanView, error)
	Update(ctx context.Context, clinicID, medicalRecordID uint64, input *UpdateClinicalPlanInput) (*ClinicalPlanView, error)
	// Delete は staff actor 必須。削除前値の audit を同一 DBOrTx で fail-closed に書く（BUG-010 residual）。
	Delete(ctx context.Context, clinicID, medicalRecordID uint64, actorID *uint64) error
}

type clinicalPlanService struct {
	repo         ClinicalPlanRepository
	medRec       clinicalPlanMedicalRecordRepository
	diagTypeRepo DiagnosisTypeRepository
	diagNameRepo DiagnosisNameRepository
	transactor   Transactor
	auditTx      AuditTxLogger
}

// clinicalPlanMedicalRecordRepository keeps the clinical-plan mutation path on the
// same parent-row locking contract as the other clinical child writers without
// widening the shared medicalRecordFinder dependency view.
type clinicalPlanMedicalRecordRepository interface {
	medicalRecordFinder
	medicalRecordLocker
}

// NewClinicalPlanService はClinicalPlanServiceを初期化して返す。
// BUG-010 residual: transactor と auditTx は Update の write+audit 原子性に必須
// （MRA-01 / care_plan_item と同型の fail-closed）。nil の場合 Update は拒否する。
func NewClinicalPlanService(
	repo ClinicalPlanRepository,
	medRec clinicalPlanMedicalRecordRepository,
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

func (s *clinicalPlanService) GetOrCreate(ctx context.Context, clinicID, medicalRecordID uint64) (*ClinicalPlanView, error) {
	plan, err := s.repo.FindByMedicalRecordID(ctx, clinicID, medicalRecordID)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			return nil, apperrors.Wrap(err, "failed to get clinical plan")
		}
		if _, ownerErr := s.medRec.FindByID(ctx, clinicID, medicalRecordID); ownerErr != nil {
			return nil, apperrors.Wrap(ownerErr, "failed to verify parent medical record")
		}
		// clinical_plans.clinic_id は NOT NULL（001）。trusted clinic を必ず設定する。
		plan = &model.ClinicalPlan{MedicalRecordID: medicalRecordID, ClinicID: clinicID}
		if err := s.repo.Create(ctx, plan); err != nil {
			return nil, apperrors.Wrap(err, "failed to create clinical plan")
		}
		slog.InfoContext(ctx, "clinical_plan created",
			slog.Uint64("clinic_id", clinicID),
			slog.Uint64("clinical_plan_id", plan.ID),
			slog.Uint64("medical_record_id", medicalRecordID))
	}
	return s.loadView(ctx, clinicID, plan)
}

// loadView は plan に生存中の診断行を添える。
func (s *clinicalPlanService) loadView(ctx context.Context, clinicID uint64, plan *model.ClinicalPlan) (*ClinicalPlanView, error) {
	rows, err := s.repo.ListDiagnoses(ctx, clinicID, plan.ID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get clinical plan diagnoses")
	}
	return &ClinicalPlanView{Plan: plan, Diagnoses: rows}, nil
}

func clinicalPlanAuditValue(v *ClinicalPlanView) map[string]any {
	if v == nil || v.Plan == nil {
		return nil
	}
	p := v.Plan
	diagnoses := make([]map[string]any, 0, len(v.Diagnoses))
	for _, d := range v.Diagnoses {
		diagnoses = append(diagnoses, map[string]any{
			"id":                      d.ID,
			"diagnosis_slot":          d.DiagnosisSlot,
			"diagnosis_type_id":       d.DiagnosisTypeID,
			"diagnosis_name_id":       d.DiagnosisNameID,
			"diagnosis_name_snapshot": d.DiagnosisNameSnapshot,
			"category_name_snapshot":  d.CategoryNameSnapshot,
			"sort_order":              d.SortOrder,
			"version":                 d.Version,
		})
	}
	return map[string]any{
		"id":                p.ID,
		"medical_record_id": p.MedicalRecordID,
		"physical_exam":     p.PhysicalExam,
		"diagnoses":         diagnoses,
		"diagnosis_details": p.DiagnosisDetails,
		"treatment_policy":  p.TreatmentPolicy,
		"version":           p.Version,
	}
}

func (s *clinicalPlanService) auditUpdateTx(ctx context.Context, clinicID uint64, actorID *uint64, before, after *ClinicalPlanView) error {
	// MRA-01 / BUG-010 residual: clinical legal fields require fail-closed audit in the same tx.
	if s.auditTx == nil {
		return apperrors.WrapInternalServerError("clinical plan audit dependency is required")
	}
	if actorID == nil || *actorID == 0 {
		return apperrors.WrapInvalidInput("authenticated staff actor is required for clinical plan mutation")
	}
	resourceID := uint64(0)
	if after != nil && after.Plan != nil {
		resourceID = after.Plan.ID
	} else if before != nil && before.Plan != nil {
		resourceID = before.Plan.ID
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

func (s *clinicalPlanService) auditDeleteTx(ctx context.Context, clinicID uint64, actorID *uint64, before *ClinicalPlanView) error {
	// MRA-01 / BUG-010 residual: destructive clinical-plan delete requires fail-closed audit
	// with pre-delete field values in the same tx (care_plan_item.auditDeleteTx と同型).
	if s.auditTx == nil {
		return apperrors.WrapInternalServerError("clinical plan audit dependency is required")
	}
	if actorID == nil || *actorID == 0 {
		return apperrors.WrapInvalidInput("authenticated staff actor is required for clinical plan mutation")
	}
	resourceID := uint64(0)
	if before != nil && before.Plan != nil {
		resourceID = before.Plan.ID
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

func (s *clinicalPlanService) Update(ctx context.Context, clinicID, medicalRecordID uint64, input *UpdateClinicalPlanInput) (*ClinicalPlanView, error) {
	// Fail-closed before any mutation: missing audit/tx deps must not allow silent writes.
	if s.auditTx == nil {
		return nil, apperrors.WrapInternalServerError("clinical plan audit dependency is required")
	}
	if s.transactor == nil {
		return nil, apperrors.WrapInternalServerError("clinical plan transaction dependency is required")
	}

	if input.ActorID == nil || *input.ActorID == 0 {
		return nil, apperrors.WrapInvalidInput("authenticated staff actor is required for clinical plan mutation")
	}
	if input.Diagnoses != nil {
		// 形（件数・slot）は DB に触れる前に検査する。
		if err := validateDiagnosisRowShape(*input.Diagnoses); err != nil {
			return nil, err
		}
	}

	var updated *ClinicalPlanView
	if err := s.withTx(ctx, func(txCtx context.Context) error {
		if err := lockDraftMedicalRecord(
			txCtx,
			s.medRec,
			clinicID,
			medicalRecordID,
			"failed to find medical record",
			"",
		); err != nil {
			return err
		}
		current, err := s.GetOrCreate(txCtx, clinicID, medicalRecordID)
		if err != nil {
			return apperrors.Wrap(err, "failed to get or create clinical plan")
		}
		if input.Diagnoses != nil {
			if err := validateDiagnosisRows(txCtx, clinicID, *input.Diagnoses, s.diagTypeRepo, s.diagNameRepo); err != nil {
				return err
			}
		}

		if !input.hasChanges() {
			updated = current
			return nil
		}
		plan := current.Plan
		// バージョンをインクリメント（input.Version 指定時はそれを起点にする。version 一致確認は
		// repo.Update の WHERE 述語に一本化したため、ここでの事前チェックは行わない。
		// medical_record_crud.go の Update と同型）。診断行だけの更新でも plan の version を進める
		// （FE の楽観ロックトークンを 1 本に保つ。EMR-495 Q9）。
		nextVersion := plan.Version + 1
		if input.Version != nil {
			nextVersion = *input.Version + 1
		}
		effective := *input
		effective.persistVersion = &nextVersion

		// Snapshot before mutation for audit OldValue (do not mutate plan pointer fields after).
		beforePlan := *plan
		before := &ClinicalPlanView{Plan: &beforePlan, Diagnoses: current.Diagnoses}
		if err := s.repo.Update(txCtx, clinicID, plan.ID, effective, input.Version); err != nil {
			return apperrors.Wrap(err, "failed to update clinical plan")
		}
		if input.Diagnoses != nil {
			if err := s.repo.ReplaceDiagnoses(txCtx, clinicID, plan.ID, *input.Diagnoses); err != nil {
				return apperrors.Wrap(err, "failed to replace clinical plan diagnoses")
			}
		}
		reloaded, err := s.repo.FindByMedicalRecordID(txCtx, clinicID, medicalRecordID)
		if err != nil {
			return apperrors.Wrap(err, "failed to get updated clinical plan")
		}
		after, err := s.loadView(txCtx, clinicID, reloaded)
		if err != nil {
			return err
		}
		if err := s.auditUpdateTx(txCtx, clinicID, input.ActorID, before, after); err != nil {
			return err
		}
		updated = after
		return nil
	}); err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "clinical_plan updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("clinical_plan_id", updated.Plan.ID),
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
	// Preserve the established clinical_plan NotFound contract before checking the
	// parent lifecycle. The transaction below re-reads the row after locking the
	// parent, so this preflight does not authorize the subsequent mutation.
	if _, err := s.repo.FindByMedicalRecordID(ctx, clinicID, medicalRecordID); err != nil {
		return apperrors.Wrap(err, "failed to get clinical plan")
	}

	if actorID == nil || *actorID == 0 {
		return apperrors.WrapInvalidInput("authenticated staff actor is required for clinical plan mutation")
	}

	var plan *model.ClinicalPlan
	if err := s.withTx(ctx, func(txCtx context.Context) error {
		if err := lockDraftMedicalRecord(
			txCtx,
			s.medRec,
			clinicID,
			medicalRecordID,
			"failed to find medical record",
			"",
		); err != nil {
			return err
		}
		currentPlan, err := s.repo.FindByMedicalRecordID(txCtx, clinicID, medicalRecordID)
		if err != nil {
			return apperrors.Wrap(err, "failed to get clinical plan")
		}
		plan = currentPlan
		before, err := s.loadView(txCtx, clinicID, currentPlan)
		if err != nil {
			return err
		}
		// Snapshot pre-delete values for audit OldValue (GORM soft-delete; durable recovery of
		// clinical field content for legal audit is this OldValue, not the tombstoned row alone).
		beforePlan := *currentPlan
		before.Plan = &beforePlan
		// 診断行は plan と同じ tx・同じ削除意味論（論理削除）で扱う。
		if err := s.repo.DeleteDiagnoses(txCtx, clinicID, currentPlan.ID); err != nil {
			return apperrors.Wrap(err, "failed to delete clinical plan diagnoses")
		}
		if err := s.repo.Delete(txCtx, clinicID, currentPlan.ID); err != nil {
			return apperrors.Wrap(err, "failed to delete clinical plan")
		}
		return s.auditDeleteTx(txCtx, clinicID, actorID, before)
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
