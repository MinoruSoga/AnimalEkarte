package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ─── #201 B-2: 保存時 dose 再検証 ──────────────────────────────────────────────

// evaluateDoseForSave は medicine 明細の保存時 BE 再検証を行う（healthcare HIGH-3/HIGH-4）。
// (eval, nil)        : per_weight 対象で再検証成立（snapshot 永続化・逸脱 audit に使用）
// (nil, nil)         : 非対象（medicine でない/per_weight でない/species 解決不能/体重未記録/param 無し）→ 後方互換の手動入力
// (nil, err)         : species 不一致など fail-closed すべき安全違反、または読み出し失敗
// BE9-2D ④b: 保存 tx（Transactor.WithTx）の txCtx で呼ばれ、各 repo の dbOrTx が同一 tx で読む
// （読み出しと UPDATE を同一 tx に束ねるスナップショット TOCTOU 防止 = security review MEDIUM-1 を維持）。
func (s *treatmentService) evaluateDoseForSave(
	ctx context.Context,
	clinicID, medicalRecordID uint64,
	itemType model.TreatmentItemType, medicineID *uint64, submittedQty float64,
) (*SavedDoseEvaluation, error) {
	if itemType != model.TreatmentItemTypeMedicine || medicineID == nil {
		return nil, nil
	}
	medicine, err := s.medicineRepo.FindByID(ctx, clinicID, *medicineID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load medicine for dose revalidation", "error", err)
		return nil, apperrors.Wrap(err, "failed to load medicine for dose revalidation")
	}
	if medicine.CalculationType != model.MedicineCalculationTypePerWeight {
		return nil, nil // 後方互換: 手動 quantity（再検証なし）
	}

	// pet species を解決（medical_record → pet → animal_species）。マップ不能は fail-closed（手動）。
	mr, err := s.medicalRecordRepo.FindByID(ctx, clinicID, medicalRecordID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load medical record for dose revalidation", "error", err)
		return nil, apperrors.Wrap(err, "failed to load medical record for dose revalidation")
	}
	if mr.Pet == nil || mr.Pet.AnimalSpecies == nil {
		slog.InfoContext(ctx, "dose revalidation skipped: pet species unavailable", "medical_record_id", medicalRecordID)
		return nil, nil
	}
	species, ok := NormalizeDoseSpecies(mr.Pet.AnimalSpecies.Name)
	if !ok {
		slog.InfoContext(ctx, "dose revalidation skipped: species not mappable (fail-closed)", "species_name", mr.Pet.AnimalSpecies.Name)
		return nil, nil
	}

	// 体重を解決（来院近傍 vital・暫定確定(2026-07-15 PO)）。未記録は fail-closed（手動）。
	weightKg, weightSource, ok := s.resolveDoseWeight(ctx, clinicID, medicalRecordID)
	if !ok {
		slog.InfoContext(ctx, "dose revalidation skipped: weight unavailable", "medical_record_id", medicalRecordID)
		return nil, nil
	}

	// pet species で param を引く＝犬↔猫越境を構造的に排除（HIGH-4）。
	param, err := s.doseParamRepo.FindByMedicineAndSpecies(ctx, clinicID, *medicineID, species)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil, nil // この種の param 無し → 手動（fail-closed）
		}
		slog.ErrorContext(ctx, "failed to load dose param for revalidation", "error", err)
		return nil, apperrors.Wrap(err, "failed to load dose param")
	}
	// 防御アサート（HIGH-4）: 引いた species と param.Species は一致するはず。不一致＝安全のため保存中止。
	if param.Species != species {
		slog.ErrorContext(ctx, "dose param species mismatch (fail-closed)", "expected", string(species), "got", string(param.Species))
		return nil, apperrors.WrapConflict("投与量パラメータの種別が患者種と一致しません（安全のため保存を中止しました）")
	}

	eval := EvaluateSavedDose(SavedDoseInput{
		Medicine:     medicine,
		Param:        param,
		Species:      species,
		WeightKg:     weightKg,
		SubmittedQty: submittedQty,
	})
	eval.WeightSource = weightSource
	return &eval, nil
}

// resolveDoseWeight は来院カルテの vital から最新の体重(kg)を解決する（暫定確定(2026-07-15 PO)の安全側デフォルト）。
func (s *treatmentService) resolveDoseWeight(ctx context.Context, clinicID, medicalRecordID uint64) (weightKg float64, source string, ok bool) {
	vitals, err := s.vitalRepo.FindByMedicalRecordID(ctx, clinicID, medicalRecordID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load vitals for dose weight", "error", err)
		return 0, "", false
	}
	var chosen *model.VitalRecord
	for i := range vitals {
		v := &vitals[i]
		if v.Weight == nil || *v.Weight <= 0 {
			continue
		}
		if chosen == nil || v.RecordedAt.After(chosen.RecordedAt) {
			chosen = v
		}
	}
	if chosen == nil {
		return 0, "", false
	}
	kg, ok := NormalizeWeightToKg(*chosen.Weight, chosen.WeightUnit)
	if !ok {
		return 0, "", false
	}
	return kg, fmt.Sprintf("vital_records:%d", chosen.ID), true
}

// auditDoseDeviationTx は逸脱（過量/過少/著しい上書き）を audit_logs に fail-closed で記録する。
// BE9-2D ④b: 保存 tx（Transactor.WithTx）の txCtx で呼ばれ、AuditTxLogger.LogEntryTx（内部の
// CreateTx が dbOrTx で ctx の ambient tx に参加）が同一 tx へ書く。旧 repos.Transaction 機構では
// tx が ctx に伝播せず txRepos.Audit.CreateTx 直呼びが必要だったが、WithTx 化で medicine/dose-param
// と同じ共有監査カーネル経由（R1-2 (D1)・#211/refund パターン）に統一した。監査ログの内容は
// 旧実装と同一（LogEntryTx = buildAuditLog + validateAuditLog + CreateTx で処理列も等価）。
// エラーを返すと呼び出し元の WithTx が rollback し、treatment 書込ごと無効になる。
func (s *treatmentService) auditDoseDeviationTx(ctx context.Context, clinicID uint64, actorID *uint64, treatmentID, medicineID uint64, eval *SavedDoseEvaluation) error {
	if s.auditTx == nil || eval == nil || !eval.IsDeviation() {
		return nil
	}
	actorType := auditActorTypeFor(actorID)
	input := &AuditLogInput{
		ClinicID:   &clinicID,
		ActorID:    actorID,
		ActorType:  actorType,
		Action:     model.AuditActionTreatmentDoseDeviation,
		Resource:   model.AuditResourceTreatmentDose,
		ResourceID: &treatmentID,
		OldValue:   map[string]any{"computed_quantity": eval.Computed.Quantity, "upper_cap_mg": eval.UpperCapMg, "has_upper_cap": eval.HasUpperCap},
		NewValue:   map[string]any{"submitted_quantity": eval.Snapshot.SubmittedQty, "effective_mg": eval.SavedEffectiveMg, "exceeds_max": eval.ExceedsCapSaved, "below_min": eval.BelowMinSaved},
		Metadata:   map[string]any{"medicine_id": medicineID, "deviation_threshold_pending_operational": true},
	}
	if err := s.auditTx.LogEntryTx(ctx, input); err != nil {
		return apperrors.Wrap(err, "failed to write dose deviation audit log")
	}
	return nil
}
