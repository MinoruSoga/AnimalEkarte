package medicalrecord

import (
	"context"
)

// validateTreatmentMasterFKs は request 由来の clinic-scoped マスタFK
// (medicine/procedure/consultation/inventory) の所有権を検証する。treatments は自前 clinic_id を
// 持たず medical_record 経由で隔離されるため、これらのマスタが別 clinic のものでないことを
// write 前に明示確認しないと cross-tenant の mislink（#124/#125 同型）を作れてしまう。
// 別 clinic のマスタ参照は repo の FindByID が NotFound を返し write を遮断する。
// InventoryID は write 前の所有権検証で cross-clinic treatment link を拒否し、在庫減算時も
// DecreaseStock(ctx, clinicID, targetInvID, qty) の atomic predicate で同じ clinic scope を再適用する。
func (s *treatmentService) validateTreatmentMasterFKs(ctx context.Context, clinicID uint64, medicineID, procedureID, consultationID, inventoryID *uint64) error {
	if err := validateOwnedMasterFK(ctx, "medicine", clinicID, medicineID,
		func(actx context.Context, cid, mid uint64) error {
			_, err := s.medicineRepo.FindByID(actx, cid, mid)
			return err
		}); err != nil {
		return err
	}
	if err := validateOwnedMasterFK(ctx, "procedure", clinicID, procedureID,
		func(actx context.Context, cid, mid uint64) error {
			_, err := s.procedureRepo.FindByID(actx, cid, mid)
			return err
		}); err != nil {
		return err
	}
	if err := validateOwnedMasterFK(ctx, "consultation", clinicID, consultationID,
		func(actx context.Context, cid, mid uint64) error {
			_, err := s.consultationRepo.FindByID(actx, cid, mid)
			return err
		}); err != nil {
		return err
	}
	if err := validateOwnedMasterFK(ctx, "inventory", clinicID, inventoryID,
		func(actx context.Context, cid, mid uint64) error {
			_, err := s.inventoryRepo.FindByID(actx, cid, mid)
			return err
		}); err != nil {
		return err
	}
	return nil
}
