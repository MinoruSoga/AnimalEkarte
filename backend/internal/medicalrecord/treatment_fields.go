package medicalrecord

import (
	"maps"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

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
