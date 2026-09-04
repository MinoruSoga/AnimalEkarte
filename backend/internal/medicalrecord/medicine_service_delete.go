package medicalrecord

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *medicineService) deleteMedicineInTx(
	txCtx context.Context,
	clinicID, id uint64,
	m *model.Medicine,
) error {
	// Re-check usage inside tx to shrink MRC-07 TOCTOU window for medicine deletes.
	if m.ParentID != nil {
		usageCount, err := s.repo.CountUsageByMedicineID(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to re-check medicine usage")
		}
		if usageCount > 0 {
			return apperrors.WrapConflict("この薬剤は診療記録で使用中のため削除できません")
		}
	} else {
		count, err := s.repo.CountChildrenByParentID(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to re-count medicine children")
		}
		if count > 0 {
			return apperrors.WrapConflict(
				fmt.Sprintf("このカテゴリには%d件の薬剤が含まれています。先に薬剤を移動または削除してください", count),
			)
		}
	}
	if err := s.repo.Delete(txCtx, clinicID, id); err != nil {
		slog.ErrorContext(txCtx, "failed to delete medicine", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to delete medicine")
	}
	if m.InventoryID != nil && *m.InventoryID != 0 {
		err := s.inventoryRepo.Delete(txCtx, clinicID, *m.InventoryID)
		// NotFound is acceptable for already-orphaned inventory; other errors fail closed.
		if err != nil && !apperrors.IsNotFound(err) {
			slog.ErrorContext(txCtx, "failed to delete linked inventory for medicine", "error", err, "clinic_id", clinicID, "inventory_id", *m.InventoryID)
			return apperrors.Wrap(err, "failed to delete linked inventory for medicine")
		}
	} else if err := s.inventoryRepo.DeleteByNameAndMedicineCategory(txCtx, clinicID, m.Name); err != nil {
		slog.ErrorContext(txCtx, "failed to delete linked inventory for medicine by name", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to delete linked inventory for medicine")
	}
	resourceID := id
	if err := s.auditTx.LogEntryTx(txCtx, &AuditEntry{
		ClinicID:   &clinicID,
		ActorType:  auditActorTypeFor(nil),
		Action:     "medicine.delete",
		Resource:   model.AuditResourceMedicine,
		ResourceID: &resourceID,
		OldValue: map[string]any{
			"id":           m.ID,
			"name":         m.Name,
			"inventory_id": m.InventoryID,
		},
	}); err != nil {
		return apperrors.Wrap(err, "failed to audit medicine delete")
	}
	return nil
}
