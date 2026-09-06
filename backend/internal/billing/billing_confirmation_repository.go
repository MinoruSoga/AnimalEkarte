package billing

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// BillingConfirmationRepository は会計医師確認のデータアクセスインターフェース
type BillingConfirmationRepository interface {
	FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) (*model.BillingConfirmation, error)
	Create(ctx context.Context, review *model.BillingConfirmation) error
	Update(ctx context.Context, clinicID, id uint64, cmd UpdateBillingConfirmationInput) error
	LockActiveStaffAssignment(ctx context.Context, clinicID, staffID uint64) error
}

type billingConfirmationRepository struct {
	db *gorm.DB
}

// NewBillingConfirmationRepository はBillingConfirmationRepositoryを初期化して返す
func NewBillingConfirmationRepository(db *gorm.DB) BillingConfirmationRepository {
	return &billingConfirmationRepository{db: db}
}

func (r *billingConfirmationRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) (*model.BillingConfirmation, error) {
	var review model.BillingConfirmation
	err := persistence.DBOrTx(ctx, r.db).
		Joins("JOIN medical_records ON medical_records.id = billing_confirmations.medical_record_id AND medical_records.deleted_at IS NULL").
		Where("medical_records.clinic_id = ? AND billing_confirmations.medical_record_id = ?", clinicID, medicalRecordID).
		First(&review).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "billing_confirmation", fmt.Sprintf("medical_record_id=%d", medicalRecordID))
	}
	return &review, nil
}

func (r *billingConfirmationRepository) Create(ctx context.Context, review *model.BillingConfirmation) error {
	if err := persistence.DBOrTx(ctx, r.db).Create(review).Error; err != nil {
		return apperrors.FromGORM(err, "billing_confirmation", "")
	}
	return nil
}

// Update は persistence.DBOrTx(ctx, r.db) で ambient tx に参加する（billingConfirmationService.Confirm/Return が
// SD-2/X-11 と同種の確定済みカルテガードのため LockByIDForUpdate の行ロックと同一 tx に束ねる）。
func (r *billingConfirmationRepository) Update(ctx context.Context, clinicID, id uint64, cmd UpdateBillingConfirmationInput) error {
	return r.update(ctx, clinicID, id, buildBillingConfirmationUpdate(cmd))
}

func (r *billingConfirmationRepository) update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	// Restrict update to rows belonging to this clinic via subquery on medical_records
	result := persistence.DBOrTx(ctx, r.db).
		Model(&model.BillingConfirmation{}).
		Where("id = ? AND medical_record_id IN (SELECT id FROM medical_records WHERE clinic_id = ?)", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "billing_confirmation", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("billing_confirmation", fmt.Sprintf("%d", id))
	}
	return nil
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
