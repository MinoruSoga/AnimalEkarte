package billing

import (
	"fmt"
	"sort"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func lockBillingClinic(db *gorm.DB, billingID uint64) (uint64, error) {
	var ref struct {
		ClinicID uint64
	}
	if err := db.
		Table("billings").
		Select("clinic_id").
		Where("id = ? AND deleted_at IS NULL", billingID).
		// SavePayment uses the billing row as the serialization point for its
		// one-payment-per-billing upsert. Refund and split writes use the same
		// lock order so concurrent accounting writes cannot race that upsert.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Take(&ref).Error; err != nil {
		return 0, apperrors.FromGORM(
			err,
			"billing",
			fmt.Sprintf("%d", billingID),
		)
	}
	return ref.ClinicID, nil
}

// lockActiveBillingStaffs validates request-derived actor FKs and keeps their
// active/same-clinic state stable until the surrounding write commits.
func lockActiveBillingStaffs(
	db *gorm.DB,
	clinicID uint64,
	staffIDs []uint64,
) error {
	unique := make(map[uint64]struct{}, len(staffIDs))
	for _, staffID := range staffIDs {
		if staffID == 0 {
			return apperrors.WrapForbidden(
				"active staff in the billing clinic is required",
			)
		}
		unique[staffID] = struct{}{}
	}
	if len(unique) == 0 {
		return nil
	}

	ids := make([]uint64, 0, len(unique))
	for staffID := range unique {
		ids = append(ids, staffID)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	var found []uint64
	if err := db.
		Table("staff_clinic_assignments").
		Select("staff_clinic_assignments.staff_id").
		Joins(
			"JOIN staffs ON staffs.id = staff_clinic_assignments.staff_id"+
				" AND staffs.is_active = TRUE"+
				" AND staffs.deleted_at IS NULL",
		).
		Where(
			"staff_clinic_assignments.staff_id IN ?"+
				" AND staff_clinic_assignments.clinic_id = ?"+
				" AND staff_clinic_assignments.deleted_at IS NULL",
			ids,
			clinicID,
		).
		Order("staff_clinic_assignments.staff_id ASC").
		Clauses(clause.Locking{Strength: "SHARE"}).
		Find(&found).Error; err != nil {
		return apperrors.FromGORM(err, "staff_clinic_assignment", "")
	}
	if len(found) != len(ids) {
		return apperrors.WrapForbidden(
			"active staff in the billing clinic is required",
		)
	}
	return nil
}
