package billing

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// BillingConfirmationRepository は会計の操作者検証（有効な在籍スタッフの行ロック）のデータアクセスインターフェース。
// held for EMR-462: 旧 billing_confirmations 行（カルテ単位の pending / confirmed / returned）の読み書きは、置換先が
// 請求ごとの doctor_review_sets / doctor_review_records で状態の意味が異なるため退役した。
// 明細 writer（billingItemRepository）が操作者検証として共有する LockActiveStaffAssignment だけを残す。
type BillingConfirmationRepository interface {
	LockActiveStaffAssignment(ctx context.Context, clinicID, staffID uint64) error
}

type billingConfirmationRepository struct {
	db *gorm.DB
}

// NewBillingConfirmationRepository はBillingConfirmationRepositoryを初期化して返す
func NewBillingConfirmationRepository(db *gorm.DB) BillingConfirmationRepository {
	return &billingConfirmationRepository{db: db}
}

// LockActiveStaffAssignment verifies that the authenticated actor is an active,
// non-deleted staff member with an active assignment to clinicID. The shared
// row lock keeps both identity and assignment valid until the confirmation
// write commits.
func (r *billingConfirmationRepository) LockActiveStaffAssignment(ctx context.Context, clinicID, staffID uint64) error {
	tx := persistence.TxFromContext(ctx)
	if tx == nil {
		return apperrors.WrapInternalServerError("billing confirmation actor validation requires an active transaction")
	}
	if clinicID == 0 || staffID == 0 {
		return apperrors.WrapForbidden("active clinic assignment is required")
	}

	var assignment model.StaffClinicAssignment
	err := tx.WithContext(ctx).
		Model(&model.StaffClinicAssignment{}).
		Select("staff_clinic_assignments.*").
		Joins("JOIN staffs ON staffs.id = staff_clinic_assignments.staff_id AND staffs.deleted_at IS NULL AND staffs.is_active = TRUE").
		Where(
			"staff_clinic_assignments.staff_id = ? AND staff_clinic_assignments.clinic_id = ? AND staff_clinic_assignments.deleted_at IS NULL",
			staffID,
			clinicID,
		).
		Clauses(clause.Locking{Strength: "SHARE"}).
		First(&assignment).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperrors.WrapForbidden("active clinic assignment is required")
	}
	if err != nil {
		return apperrors.FromGORM(err, "staff_clinic_assignment", "actor validation")
	}
	return nil
}
