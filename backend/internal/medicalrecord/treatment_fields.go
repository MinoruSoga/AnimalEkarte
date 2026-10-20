package medicalrecord

import (
	"maps"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/jackc/pgx/v5/pgtype"
)

// errMsgDoseAmountUnitPair は EMR-488: 薬剤3値の量・単位ペア不変条件違反時の拒否文言。
// DB の ck_administration_dose_pair / ck_prescription_total_pair と同一条件を service 境界で先に検証し、
// CHECK 違反（23514→500）ではなく 400 で拒否する。書込は全体不成立。
const errMsgDoseAmountUnitPair = "1回製剤量（処方総量）とその単位はペアで入力してください。単位は空白のみの値を許可しません"
const errMsgDoseAmountNegative = "1回製剤量・処方総量に負の値は指定できません"

// validateDoseAmountUnitPair は 量・単位 のペア不変条件（両 NULL または 量非NULL かつ trim 後非空の単位）
// を検証する純関数（001_init.sql の ck_administration_dose_pair / ck_prescription_total_pair と同一条件）。
// EMR-488: 単位の推測補完はしないため、片側のみの入力は保存せず 400 で拒否する。
// 量の負値も 400（臨床量として無意味。FE の入力ガードの裏返しとして wire 単体でも拒否する）。
func validateDoseAmountUnitPair(amount *model.ExactNumeric, unit *string) error {
	if amount != nil {
		if amount.NaN || amount.InfinityModifier != pgtype.Finite ||
			(amount.Int != nil && amount.Int.Sign() < 0) {
			return apperrors.WrapInvalidInput(errMsgDoseAmountNegative)
		}
	}
	if amount == nil && unit == nil {
		return nil
	}
	if amount != nil && unit != nil && strings.TrimSpace(*unit) != "" {
		return nil
	}
	return apperrors.WrapInvalidInput(errMsgDoseAmountUnitPair)
}

// validateDosePairsForCreate は create の3値ペア（1回製剤量・処方総量）を検証する。
func validateDosePairsForCreate(input *CreateTreatmentInput) error {
	if err := validateDoseAmountUnitPair(input.AdministrationDoseAmount, input.AdministrationDoseUnit); err != nil {
		return err
	}
	return validateDoseAmountUnitPair(input.PrescriptionTotalAmount, input.PrescriptionTotalUnit)
}

// validateDosePairsForUpdate は update 適用後の実効状態（nil 入力は既存値を維持）で
// 3値ペア不変条件を検証する。既存の有効ペアに対する片側のみの変更（単位維持で量だけ変更等）は許可する。
func validateDosePairsForUpdate(current *model.Treatment, input *UpdateTreatmentInput) error {
	effAdminAmount := current.AdministrationDoseAmount
	if input.AdministrationDoseAmount != nil {
		effAdminAmount = input.AdministrationDoseAmount
	}
	effAdminUnit := current.AdministrationDoseUnit
	if input.AdministrationDoseUnit != nil {
		effAdminUnit = input.AdministrationDoseUnit
	}
	if err := validateDoseAmountUnitPair(effAdminAmount, effAdminUnit); err != nil {
		return err
	}

	effTotalAmount := current.PrescriptionTotalAmount
	if input.PrescriptionTotalAmount != nil {
		effTotalAmount = input.PrescriptionTotalAmount
	}
	effTotalUnit := current.PrescriptionTotalUnit
	if input.PrescriptionTotalUnit != nil {
		effTotalUnit = input.PrescriptionTotalUnit
	}
	return validateDoseAmountUnitPair(effTotalAmount, effTotalUnit)
}

// applyTreatmentDiscountGuard rechecks discount fields against the FOR UPDATE locked row.
// If the request differs from the locked current value without DiscountEditAllowed → Forbidden.
// If equal and not allowed, omit discount columns so a stale PATCH cannot race-write.
func applyTreatmentDiscountGuard(locked *model.Treatment, input *UpdateTreatmentInput) error {
	if err := applyTreatmentDiscountRateGuard(locked, input); err != nil {
		return err
	}
	return applyTreatmentDiscountAmountGuard(locked, input)
}

func applyTreatmentDiscountRateGuard(locked *model.Treatment, input *UpdateTreatmentInput) error {
	if input.DiscountRate == nil {
		return nil
	}
	if !httpapi.FloatEquals(*input.DiscountRate, locked.DiscountRate) {
		if input.DiscountEditAllowed {
			return nil
		}
		return apperrors.WrapForbidden("割引フィールドの編集権限がありません")
	}
	if !input.DiscountEditAllowed {
		input.DiscountRate = nil
	}
	return nil
}

func applyTreatmentDiscountAmountGuard(locked *model.Treatment, input *UpdateTreatmentInput) error {
	if input.DiscountAmount == nil {
		return nil
	}
	if *input.DiscountAmount != locked.DiscountAmount {
		if input.DiscountEditAllowed {
			return nil
		}
		return apperrors.WrapForbidden("割引フィールドの編集権限がありません")
	}
	if !input.DiscountEditAllowed {
		input.DiscountAmount = nil
	}
	return nil
}

// GORMのzero-value問題（false/0/"" がスキップされる）を回避するために使用する。
func buildTreatmentUpdate(input *UpdateTreatmentInput) map[string]any {
	fields := map[string]any{}
	if input.ItemType != nil {
		fields["item_type"] = *input.ItemType
	}
	if input.ConsultationID != nil {
		fields["consultation_id"] = *input.ConsultationID
	}
	if input.ProcedureID != nil {
		fields["procedure_id"] = *input.ProcedureID
	}
	if input.MedicineID != nil {
		fields["medicine_id"] = *input.MedicineID
	}
	if input.InventoryID != nil {
		fields["inventory_id"] = *input.InventoryID
	}
	if input.UnitPrice != nil {
		fields["unit_price"] = *input.UnitPrice
	}
	if input.Quantity != nil {
		fields["quantity"] = *input.Quantity
	}
	if input.IsSelected != nil {
		fields["is_selected"] = *input.IsSelected
	}
	if input.Status != nil {
		fields["status"] = *input.Status
	}
	if input.Content != nil {
		fields["content"] = *input.Content
	}
	if input.Memo != nil {
		fields["memo"] = *input.Memo
	}
	if input.AdminRoute != nil {
		fields["admin_route"] = *input.AdminRoute
	}
	if input.IsInsurance != nil {
		fields["is_insurance"] = *input.IsInsurance
	}
	if input.DiscountRate != nil {
		fields["discount_rate"] = *input.DiscountRate
	}
	if input.DiscountAmount != nil {
		fields["discount_amount"] = *input.DiscountAmount
	}
	if input.SortOrder != nil {
		fields["sort_order"] = *input.SortOrder
	}
	// EMR-488: 薬剤3値。nil = 未送信 = 変更なし。送信時はその値（0 含む）をそのまま保存する。
	// persistDose（mg 計算値スナップショット）とは列が異なり、ここで混同・上書きしない。
	if input.AdministrationDoseAmount != nil {
		fields["administration_dose_amount"] = *input.AdministrationDoseAmount
	}
	if input.AdministrationDoseUnit != nil {
		fields["administration_dose_unit"] = *input.AdministrationDoseUnit
	}
	if input.AdministrationInstructions != nil {
		fields["administration_instructions"] = *input.AdministrationInstructions
	}
	if input.PrescriptionTotalAmount != nil {
		fields["prescription_total_amount"] = *input.PrescriptionTotalAmount
	}
	if input.PrescriptionTotalUnit != nil {
		fields["prescription_total_unit"] = *input.PrescriptionTotalUnit
	}
	if len(input.persistDose) > 0 {
		maps.Copy(fields, input.persistDose)
	}
	return fields
}

// effectiveDoseInputs は existing と input から dose 再評価に使う実効値（item_type/medicine_id/
// quantity）を算出する純関数（BE-refactor.md E-3）。
func effectiveDoseInputs(existing *model.Treatment, input *UpdateTreatmentInput) (effItemType model.TreatmentItemType, effMedicineID *uint64, effQty float64) {
	effItemType = existing.ItemType
	if input.ItemType != nil {
		effItemType = *input.ItemType
	}
	effMedicineID = existing.MedicineID
	if input.MedicineID != nil {
		effMedicineID = input.MedicineID
	}
	effQty = existing.Quantity
	if input.Quantity != nil {
		effQty = *input.Quantity
	}
	return effItemType, effMedicineID, effQty
}

func validateTreatmentItemType(t model.TreatmentItemType) error {
	switch t {
	case model.TreatmentItemTypeConsultation,
		model.TreatmentItemTypeProcedure,
		model.TreatmentItemTypeMedicine,
		model.TreatmentItemTypeOther:
		return nil
	}
	return apperrors.WrapInvalidInput("invalid item_type: " + string(t))
}

func parseTreatmentStatus(s string) (model.TreatmentStatus, error) {
	switch model.TreatmentStatus(s) {
	case model.TreatmentStatusPending,
		model.TreatmentStatusCompleted,
		model.TreatmentStatusNotApplicable:
		return model.TreatmentStatus(s), nil
	}
	return "", apperrors.WrapInvalidInput("invalid status: " + s)
}
