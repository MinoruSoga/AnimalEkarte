package medicalrecord

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func medicineFromCreateInput(
	clinicID uint64,
	input *CreateMedicineInput,
	calcType model.MedicineCalculationType,
	taxType model.TaxType,
	taxRate float64,
) *model.Medicine {
	medicine := &model.Medicine{
		ClinicID:            clinicID,
		Name:                input.Name,
		ParentID:            input.ParentID,
		Price:               input.Price,
		IsActive:            input.IsActive,
		Description:         input.Description,
		InventoryID:         input.InventoryID,
		DefaultQuantity:     input.DefaultQuantity,
		SortOrder:           input.SortOrder,
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
		return wrapMedicineNameConflict(txCtx, err, input.Name, clinicID, 0, "failed to create medicine")
	}
	// BUG-320: 薬品作成時に在庫アイテムを自動作成
	inventoryItem := &model.InventoryItem{
		ClinicID:      clinicID,
		Name:          medicine.Name,
		Category:      model.InventoryCategoryMedicine,
		Quantity:      0,
		Unit:          "錠", // デフォルト
		MinStockLevel: 0,
		Status:        model.InventoryStatusSufficient,
	}
	if err := s.inventoryRepo.Create(txCtx, clinicID, inventoryItem); err != nil {
		slog.ErrorContext(txCtx, "failed to create inventory item for medicine", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to create inventory item for medicine")
	}
	// MRC-02: persist inventory_id so delete/rename use id not fragile name matching.
	if inventoryItem.ID != 0 {
		invID := inventoryItem.ID
		medicine.InventoryID = &invID
		if _, err := s.repo.Update(txCtx, clinicID, medicine.ID, map[string]any{colMedicineInventoryID: invID}); err != nil {
			slog.ErrorContext(txCtx, "failed to link inventory item to medicine", "error", err, "clinic_id", clinicID, "medicine_id", medicine.ID)
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
