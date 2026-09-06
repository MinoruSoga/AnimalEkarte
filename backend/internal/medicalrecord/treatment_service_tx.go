package medicalrecord

import (
	"context"
	"strconv"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *treatmentService) createTreatmentInTx(
	txCtx context.Context,
	clinicID, medicalRecordID uint64,
	input *CreateTreatmentInput,
	status model.TreatmentStatus,
) (*model.Treatment, error) {
	// テナント所有権 + 確定ロック検証（Update/Delete と対称化・healthcare review CRITICAL）。
	// treatments は自前 clinic_id を持たず medical_records 経由で隔離するため、所有権を Create でも明示検証する。
	// BE-refactor.md X-11/E-5: lockDraftMedicalRecord に行ロック+finalizedガードを集約。
	if err := lockDraftMedicalRecord(txCtx, s.medicalRecordRepo, clinicID, medicalRecordID,
		"failed to verify medical record ownership", "確定済みカルテには治療を追加できません"); err != nil {
		return nil, err
	}

	treatment := &model.Treatment{
		MedicalRecordID: medicalRecordID,
		ItemType:        input.ItemType,
		ConsultationID:  input.ConsultationID,
		ProcedureID:     input.ProcedureID,
		MedicineID:      input.MedicineID,
		InventoryID:     input.InventoryID,
		UnitPrice:       input.UnitPrice,
		Quantity:        input.Quantity,
		IsSelected:      input.IsSelected,
		Status:          status,
		Content:         input.Content,
		Memo:            input.Memo,
		AdminRoute:      input.AdminRoute,
		IsInsurance:     input.IsInsurance,
		DiscountRate:    input.DiscountRate,
		DiscountAmount:  input.DiscountAmount,
		SortOrder:       input.SortOrder,
	}

	// #201 B-2: per_weight 医薬の保存時 BE 再検証＋スナップショット値固定（C1/両上限/species 一致）。
	// TASK-377: reason-required では理由 validation → strict snapshot → audit を同一 tx で fail-closed。
	eval, derr := s.evaluateDoseForSave(txCtx, clinicID, medicalRecordID, input.ItemType, input.MedicineID, input.Quantity)
	if derr != nil {
		return nil, derr // species 不一致など fail-closed
	}
	var doseEval *SavedDoseEvaluation
	if eval != nil && eval.ExceedsCapSaved {
		return nil, apperrors.WrapInvalidInput("投与量がマスタで設定された絶対上限を超えているため保存できません")
	}
	if eval != nil {
		if err := applyDeviationReasonToEval(eval, input.DoseDeviationReason); err != nil {
			return nil, err
		}
		// reason-required の actor / audit 依存は treatment write 前に fail-closed する
		// （tx rollback に依存せず write ゼロを保証する）。
		if err := s.ensureDoseDeviationAuditReady(eval, input.ActorID); err != nil {
			return nil, err
		}
		doseEval = eval
		if err := applyDoseSnapshotToTreatment(txCtx, treatment, eval); err != nil {
			return nil, err
		}
	}

	// 1. Create Treatment
	if err := s.treatmentRepo.Create(txCtx, treatment); err != nil {
		return nil, apperrors.Wrap(err, "failed to create treatment")
	}

	// 2. Decrease Stock (if applicable)
	// BE-refactor.md FOLLOWUP-X14A: MedicineID を inventory id として DecreaseStock へ渡す
	// フォールバックは書込 IDOR（medicines.id と inventory_items.id の採番衝突で他クリニックの
	// 在庫を減算しうる）だったため廃止した。減算は InventoryID が明示指定された場合のみ行う。
	if input.InventoryID != nil && *input.InventoryID > 0 {
		if err := s.inventoryRepo.DecreaseStock(txCtx, clinicID, *input.InventoryID, input.Quantity); err != nil {
			return nil, apperrors.Wrap(err, "failed to decrease inventory stock")
		}
	}

	// #201 B-2 / BE-refactor.md R1-2: 逸脱（過量/過少/著しい上書き）を監査記録（fail-closed）。
	// tx 内で失敗すると treatment 作成・在庫減算ごとロールバックする。
	if doseEval != nil && input.MedicineID != nil {
		if err := s.auditDoseDeviationTx(txCtx, clinicID, input.ActorID, treatment.ID, *input.MedicineID, doseEval); err != nil {
			return nil, err
		}
	}

	return treatment, nil
}

func (s *treatmentService) updateTreatmentInTx(
	txCtx context.Context,
	clinicID, medicalRecordID, treatmentID uint64,
	input *UpdateTreatmentInput,
	doseRelevant bool,
) error {
	// テナント所有権 + 確定ロック検証（Create と対称化・BE-refactor.md H-8a）。
	// 冒頭の事前チェックは fast-fail として維持しつつ、tx 内でも lockDraftMedicalRecord
	// の行ロックで finalize と直列化し、チェック通過後〜Update 実行前に finalize が
	// 割り込むレースを防ぐ（BE-refactor.md E-5）。
	if err := lockDraftMedicalRecord(txCtx, s.medicalRecordRepo, clinicID, medicalRecordID,
		"failed to verify medical record ownership", "確定済みカルテの治療は編集できません"); err != nil {
		return err
	}

	// SEC-CS-F09: lock treatment row and recheck discount against the locked snapshot.
	// Handler early check uses a pre-TX GetByID; stale equality must not authorize overwrite.
	current, err := s.treatmentRepo.LockByIDForUpdate(txCtx, clinicID, treatmentID)
	if err != nil {
		return apperrors.Wrap(err, "failed to lock treatment for update")
	}
	if current.MedicalRecordID != medicalRecordID {
		return apperrors.WrapNotFound("treatment", strconv.FormatUint(treatmentID, 10))
	}
	effective := *input
	if err := applyTreatmentDiscountGuard(current, &effective); err != nil {
		return err
	}
	if len(buildTreatmentUpdate(&effective)) == 0 {
		return apperrors.WrapInvalidInput(errMsgAtLeastOneField)
	}

	var doseEval *SavedDoseEvaluation
	var doseMedicineID uint64
	if doseRelevant {
		// 親行 + treatment 行ロック取得後の locked snapshot で dose を評価する。
		effItemType, effMedicineID, effQty := effectiveDoseInputs(current, &effective)
		eval, derr := s.evaluateDoseForSave(txCtx, clinicID, medicalRecordID, effItemType, effMedicineID, effQty)
		if derr != nil {
			return derr // species 不一致など fail-closed
		}
		applied, err := s.applyLockedTreatmentDoseFields(txCtx, current, &effective, eval, effMedicineID)
		if err != nil {
			return err
		}
		doseEval = applied.eval
		doseMedicineID = applied.medicineID
	}
	if err := s.treatmentRepo.Update(txCtx, clinicID, treatmentID, effective); err != nil {
		return err
	}
	// #201 B-2 / BE-refactor.md R1-2: 逸脱監査を fail-closed 化。失敗すると Update ごとロールバックする。
	if doseEval != nil {
		return s.auditDoseDeviationTx(txCtx, clinicID, input.ActorID, treatmentID, doseMedicineID, doseEval)
	}
	return nil
}
