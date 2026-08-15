package medicalrecord

// medicine_dose_param_service.go — #201 B-2c: 薬剤 × 種の投与量パラメータ authoring CRUD。
//
// 製品軸（medicines.calculation_type/strength）は medicine_service が、種軸（dog/cat の mg/kg）は本 service が扱う。
// 医療安全ガード（authoring 経路）:
//   - 親 medicine の所有権/存在（clinicScope FindByID）。別 clinic → NotFound(404・IDOR 遮断)。
//   - dose param は per_weight 専用。calculation_type=none の薬剤への設定を拒否（誤 per_weight 保存防止）。
//   - ValidateMedicineDoseConfig で親 medicine の単位/含量整合を再検証（互換しない unit・strength 欠落を拒否）。
//   - ValidateMedicineDoseParamInput で行検証（default-deny / range / 上限必須 / 丸めペア）を DB CHECK と二重化。
// 作成/更新/削除は audit_logs に before/after で記録する（既存 AuditService・dose param 用 action 定数を再利用）。

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// medicine_dose_params の更新対象列（PUT 全列置換セマンティクス）。
const (
	colDoseParamDoseBasis       = "dose_basis"
	colDoseParamDosePerKg       = "dose_per_kg"
	colDoseParamMinMgPerKg      = "min_mg_per_kg"
	colDoseParamMaxMgPerKg      = "max_mg_per_kg"
	colDoseParamAbsoluteMaxDose = "absolute_max_dose"
	colDoseParamRoundingStep    = "rounding_step"
	colDoseParamRoundingMode    = "rounding_mode"
	colDoseParamNotes           = "notes"
)

// buildDoseParamReplaceFields は upsert(更新経路)の全列置換 map を構築する（PUT セマンティクス）。
// 省略された optional は NULL を明示置換する（map[string]any の untyped nil を GORM が NULL として書く）。
// species は自然キー（URL path）であり更新対象外。
func buildDoseParamReplaceFields(in *MedicineDoseParamInput) map[string]any {
	fields := map[string]any{
		colDoseParamDoseBasis:       string(in.DoseBasis),
		colDoseParamDosePerKg:       in.DosePerKg,
		colDoseParamNotes:           in.Notes,
		colDoseParamMinMgPerKg:      floatOrNil(in.MinMgPerKg),
		colDoseParamMaxMgPerKg:      floatOrNil(in.MaxMgPerKg),
		colDoseParamAbsoluteMaxDose: floatOrNil(in.AbsoluteMaxDose),
		colDoseParamRoundingStep:    floatOrNil(in.RoundingStep),
	}
	if in.RoundingMode != nil {
		fields[colDoseParamRoundingMode] = string(*in.RoundingMode)
	} else {
		fields[colDoseParamRoundingMode] = nil
	}
	return fields
}

// floatOrNil は *float64 を GORM update map 用の値へ変換する（nil → untyped nil = SQL NULL）。
func floatOrNil(p *float64) any {
	if p == nil {
		return nil
	}
	return *p
}

// MedicineDoseParamService は種軸 dose パラメータの authoring CRUD。
type MedicineDoseParamService interface {
	List(ctx context.Context, clinicID, medicineID uint64) ([]model.MedicineDoseParam, error)
	Upsert(ctx context.Context, clinicID, medicineID uint64, input *MedicineDoseParamInput, actorID *uint64) (*model.MedicineDoseParam, error)
	Delete(ctx context.Context, clinicID, medicineID uint64, species model.MedicineDoseSpecies, actorID *uint64) error
}

type medicineDoseParamService struct {
	repo       MedicineDoseParamRepository
	medRepo    MedicineRepository // 親 medicine の所有権/存在検証（clinicScope）
	transactor Transactor
	auditTx    AuditTxLogger // nil 可（後方互換）。dose param 変更の監査記録（fail-closed）。
}

// NewMedicineDoseParamService は Transactor/AuditTxLogger を注入する。
// BE-refactor.md R1-2 (D1): dose param の Create/Update/Delete と監査を単一 tx に束ね、
// 監査書込の失敗が書込自体もロールバックするようにする（fail-closed。#211/refund パターン踏襲）。
// transactor は必須（Upsert/Delete が s.transactor.WithTx を無条件に呼ぶため、nil 注入は panic）。
// 本番配線（service.go）とテストは常に非 nil の Transactor を注入する。
func NewMedicineDoseParamService(repo MedicineDoseParamRepository, medRepo MedicineRepository, transactor Transactor, auditTx AuditTxLogger) MedicineDoseParamService {
	return &medicineDoseParamService{repo: repo, medRepo: medRepo, transactor: transactor, auditTx: auditTx}
}

func (s *medicineDoseParamService) List(ctx context.Context, clinicID, medicineID uint64) ([]model.MedicineDoseParam, error) {
	// P1: 親 medicine の所有権/存在確認。別 clinic → NotFound(404) で越境遮断。
	if _, err := s.medRepo.FindByID(ctx, clinicID, medicineID); err != nil {
		return nil, apperrors.Wrap(err, "failed to find medicine")
	}
	params, err := s.repo.FindByMedicineID(ctx, clinicID, medicineID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list medicine dose params", "error", err, "medicine_id", medicineID, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to list medicine dose params")
	}
	return params, nil
}

func (s *medicineDoseParamService) Upsert(ctx context.Context, clinicID, medicineID uint64, input *MedicineDoseParamInput, actorID *uint64) (*model.MedicineDoseParam, error) {
	// 医療安全ガード③: 行検証（default-deny / range / 上限必須 / 丸めペア）は DB 非依存なので先に実行。
	if err := ValidateMedicineDoseParamInput(input); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate dose param input")
	}

	// MRC-03: parent per_weight / ownership / existing row checks share the write transaction
	// so concurrent medicine type flips cannot pass a stale pre-tx validation.
	var result *model.MedicineDoseParam
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		med, err := s.medRepo.FindByID(txCtx, clinicID, medicineID)
		if err != nil {
			return apperrors.Wrap(err, "failed to find medicine")
		}
		if med.CalculationType != model.MedicineCalculationTypePerWeight {
			return apperrors.WrapInvalidInput("この薬剤は per_weight 計算ではないため投与量パラメータを設定できません")
		}
		if err := ValidateMedicineDoseConfig(med.CalculationType, med.MedicineUnit, med.Strength, med.FrequencyPerDay, med.DefaultDurationDays); err != nil {
			return apperrors.Wrap(err, "failed to validate dose config")
		}
		existing, findErr := s.repo.FindByMedicineAndSpecies(txCtx, clinicID, medicineID, input.Species)
		switch {
		case findErr == nil:
			fields := buildDoseParamReplaceFields(input)
			updated, txErr := s.repo.Update(txCtx, clinicID, existing.ID, fields)
			if txErr != nil {
				slog.ErrorContext(txCtx, "failed to update medicine dose param", "error", txErr, "id", existing.ID, "clinic_id", clinicID)
				return apperrors.Wrap(txErr, "failed to update medicine dose param")
			}
			if err := s.auditChangeTx(txCtx, clinicID, actorID, model.AuditActionMedicineDoseParamUpsert, medicineID, updated.ID, existing, updated); err != nil {
				return err
			}
			result = updated
			return nil
		case apperrors.IsNotFound(findErr):
			param := &model.MedicineDoseParam{
				ClinicID:        clinicID,
				MedicineID:      medicineID,
				Species:         input.Species,
				DoseBasis:       input.DoseBasis,
				DosePerKg:       input.DosePerKg,
				MinMgPerKg:      input.MinMgPerKg,
				MaxMgPerKg:      input.MaxMgPerKg,
				AbsoluteMaxDose: input.AbsoluteMaxDose,
				RoundingStep:    input.RoundingStep,
				RoundingMode:    input.RoundingMode,
				Notes:           input.Notes,
			}
			if txErr := s.repo.Create(txCtx, clinicID, param); txErr != nil {
				slog.ErrorContext(txCtx, "failed to create medicine dose param", "error", txErr, "medicine_id", medicineID, "clinic_id", clinicID)
				return apperrors.Wrap(txErr, "failed to create medicine dose param")
			}
			if err := s.auditChangeTx(txCtx, clinicID, actorID, model.AuditActionMedicineDoseParamUpsert, medicineID, param.ID, nil, param); err != nil {
				return err
			}
			result = param
			return nil
		default:
			slog.ErrorContext(txCtx, "failed to lookup medicine dose param", "error", findErr, "medicine_id", medicineID, "clinic_id", clinicID)
			return apperrors.Wrap(findErr, "failed to lookup medicine dose param")
		}
	}); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *medicineDoseParamService) Delete(ctx context.Context, clinicID, medicineID uint64, species model.MedicineDoseSpecies, actorID *uint64) error {
	if !model.ValidMedicineDoseSpecies(species) {
		return apperrors.WrapInvalidInput("species は dog または cat である必要があります")
	}
	// MRC-03: ownership + row lookup re-checked inside the delete transaction.
	return s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if _, err := s.medRepo.FindByID(txCtx, clinicID, medicineID); err != nil {
			return apperrors.Wrap(err, "failed to find medicine")
		}
		existing, err := s.repo.FindByMedicineAndSpecies(txCtx, clinicID, medicineID, species)
		if err != nil {
			return apperrors.Wrap(err, "failed to find medicine dose param")
		}
		if txErr := s.repo.Delete(txCtx, clinicID, existing.ID); txErr != nil {
			slog.ErrorContext(txCtx, "failed to delete medicine dose param", "error", txErr, "id", existing.ID, "clinic_id", clinicID)
			return apperrors.Wrap(txErr, "failed to delete medicine dose param")
		}
		return s.auditChangeTx(txCtx, clinicID, actorID, model.AuditActionMedicineDoseParamDelete, medicineID, existing.ID, existing, nil)
	})
}

// auditChangeTx は dose param 変更（upsert/delete）を fail-closed で監査記録する。
// before/after は doseParamAuditValue でスナップショットし、metadata に medicine_id/species を含める。
// BE-refactor.md R1-2: ambient tx に参加する LogEntryTx を使う。失敗時は呼び出し元の WithTx が
// rollback し、dose param の書込自体も無効になる（#211/refund パターン踏襲）。
func (s *medicineDoseParamService) auditChangeTx(ctx context.Context, clinicID uint64, actorID *uint64, action string, medicineID, paramID uint64, before, after *model.MedicineDoseParam) error {
	if s.auditTx == nil {
		return nil
	}
	actorType := auditActorTypeFor(actorID)
	// metadata の species は before/after のいずれか存在する方から取る。
	species := ""
	switch {
	case after != nil:
		species = string(after.Species)
	case before != nil:
		species = string(before.Species)
	}
	resourceID := paramID
	input := &AuditEntry{
		ClinicID:   &clinicID,
		ActorID:    actorID,
		ActorType:  actorType,
		Action:     action,
		Resource:   model.AuditResourceMedicineDoseParam,
		ResourceID: &resourceID,
		OldValue:   doseParamAuditValue(before),
		NewValue:   doseParamAuditValue(after),
		Metadata:   map[string]any{"medicine_id": medicineID, "species": species},
	}
	if err := s.auditTx.LogEntryTx(ctx, input); err != nil {
		return apperrors.Wrap(err, "failed to audit medicine dose param change")
	}
	return nil
}

// doseParamAuditValue は audit before/after 用に dose param の主要値をスナップショットする（nil → nil）。
func doseParamAuditValue(p *model.MedicineDoseParam) map[string]any {
	if p == nil {
		return nil
	}
	v := map[string]any{
		"species":     string(p.Species),
		"dose_basis":  string(p.DoseBasis),
		"dose_per_kg": p.DosePerKg,
	}
	if p.MinMgPerKg != nil {
		v["min_mg_per_kg"] = *p.MinMgPerKg
	}
	if p.MaxMgPerKg != nil {
		v["max_mg_per_kg"] = *p.MaxMgPerKg
	}
	if p.AbsoluteMaxDose != nil {
		v["absolute_max_dose"] = *p.AbsoluteMaxDose
	}
	if p.RoundingStep != nil {
		v["rounding_step"] = *p.RoundingStep
	}
	if p.RoundingMode != nil {
		v["rounding_mode"] = string(*p.RoundingMode)
	}
	return v
}
