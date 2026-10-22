package medicalrecord

import (
	"context"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func medicineFromCreateInput(
	clinicID uint64,
	input *CreateMedicineInput,
	calcType model.MedicineCalculationType,
	taxType model.TaxType,
	taxRate model.ExactNumeric,
) *model.Medicine {
	// model.SortOrder は *int（#729）。request の int をそのまま保存する（EMR-447）。
	sortOrder := input.SortOrder
	medicine := &model.Medicine{
		ClinicID:            clinicID,
		Name:                input.Name,
		ParentID:            input.ParentID,
		Price:               input.Price,
		IsActive:            input.IsActive,
		Description:         input.Description,
		InventoryID:         input.InventoryID,
		DefaultQuantity:     input.DefaultQuantity,
		SortOrder:           &sortOrder,
		TaxType:             taxType,
		TaxRate:             taxRate,
		IsNonInsurance:      input.IsNonInsurance,
		CalculationType:     calcType,
		Strength:            input.Strength,
		FrequencyPerDay:     input.FrequencyPerDay,
		DefaultDurationDays: input.DefaultDurationDays,
	}
	if input.DosageForm != nil && *input.DosageForm != "" {
		df := model.DosageForm(*input.DosageForm)
		medicine.DosageForm = &df
	}
	if input.MedicineUnit != nil && *input.MedicineUnit != "" {
		mu := model.MedicineUnit(*input.MedicineUnit)
		medicine.MedicineUnit = &mu
	}
	return medicine
}

func (s *medicineService) createMedicineInTx(
	txCtx context.Context,
	clinicID uint64,
	input *CreateMedicineInput,
	medicine *model.Medicine,
	calcType model.MedicineCalculationType,
) error {
	if err := s.repo.Create(txCtx, medicine); err != nil {
		return wrapMedicineNameConflict(err, input.Name, "failed to create medicine")
	}
	// BUG-320: 薬品作成時に在庫アイテムを自動作成
	// #729: Quantity/MinStockLevel は nullable pointer、在庫 status は保存せず読み取り時に導出する。
	// 既存どおり数量0・最小在庫0を明示保存する（EMR-447 で model 型へ追従）。
	quantity := int64(0)
	minStockLevel := 0
	inventoryItem := &model.InventoryItem{
		ClinicID:      clinicID,
		Name:          medicine.Name,
		Category:      model.InventoryCategoryMedicine,
		Quantity:      &quantity,
		Unit:          "錠", // デフォルト
		MinStockLevel: &minStockLevel,
	}
	if err := s.inventoryRepo.Create(txCtx, clinicID, inventoryItem); err != nil {
		return apperrors.Wrap(err, "failed to create inventory item for medicine")
	}
	// MRC-02: persist inventory_id so delete/rename use id not fragile name matching.
	if inventoryItem.ID != 0 {
		invID := inventoryItem.ID
		medicine.InventoryID = &invID
		if _, err := s.repo.Update(txCtx, clinicID, medicine.ID, UpdateMedicineInput{InventoryID: &invID}); err != nil {
			return apperrors.Wrap(err, "failed to link inventory item to medicine")
		}
	}
	// #201 B-2: per_weight 有効化は安全クリティカル設定変更 → 監査（fail-closed）。
	if calcType == model.MedicineCalculationTypePerWeight {
		if err := s.auditPerWeightEnableTx(txCtx, clinicID, input.ActorID, medicine.ID, nil, medicine); err != nil {
			return err
		}
	}
	return nil
}
