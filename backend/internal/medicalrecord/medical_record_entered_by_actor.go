package medicalrecord

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// enteredByActorGuard validates the authenticated recording actor inside the
// ambient create transaction (assignment lock or verified system admin).
type enteredByActorGuard interface {
	AssertEnteredByActor(ctx context.Context, clinicID, staffID uint64, systemAdminClaim bool) error
}

type gormEnteredByActorGuard struct {
	db *gorm.DB
}

func newGormEnteredByActorGuard(db *gorm.DB) enteredByActorGuard {
	return &gormEnteredByActorGuard{db: db}
}

func (g *gormEnteredByActorGuard) AssertEnteredByActor(
	ctx context.Context,
	clinicID, staffID uint64,
	systemAdminClaim bool,
) error {
	if g == nil || g.db == nil {
		return apperrors.WrapInternalServerError("entered_by actor validation dependency is required")
	}
	if staffID == 0 {
		return apperrors.WrapInvalidInput("entered_by must be greater than zero")
	}
	if persistence.TxFromContext(ctx) == nil {
		return apperrors.WrapInternalServerError("entered_by actor validation requires an active transaction")
	}

	var staff model.Staff
	if err := persistence.DBOrTx(ctx, g.db).
		Clauses(clause.Locking{Strength: "SHARE"}).
		Where("staffs.id = ? AND staffs.deleted_at IS NULL", staffID).
		First(&staff).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.WrapForbidden("記録者としてカルテを作成する権限がありません")
		}
		return apperrors.Wrap(err, "failed to verify entered_by staff")
	}
	if !staff.IsActive {
		return apperrors.WrapForbidden("記録者としてカルテを作成する権限がありません")
	}

	if systemAdminClaim {
		ok, err := g.isActiveSystemAdminStaff(ctx, staffID)
		if err != nil {
			return err
		}
		if !ok {
			return apperrors.WrapForbidden("記録者としてカルテを作成する権限がありません")
		}
		return nil
	}

	var assignment model.StaffClinicAssignment
	if err := persistence.DBOrTx(ctx, g.db).
		Clauses(clause.Locking{Strength: "SHARE"}).
		Where("staff_id = ? AND clinic_id = ?", staffID, clinicID).
		First(&assignment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.WrapForbidden("記録者としてカルテを作成する権限がありません")
		}
		return apperrors.Wrap(err, "failed to verify entered_by clinic assignment")
	}
	if assignment.StaffID != staffID || assignment.ClinicID != clinicID {
		return apperrors.WrapForbidden("記録者としてカルテを作成する権限がありません")
	}
	return nil
}

func (g *gormEnteredByActorGuard) isActiveSystemAdminStaff(ctx context.Context, staffID uint64) (bool, error) {
	var count int64
	err := persistence.DBOrTx(ctx, g.db).
		Table("staffs").
		Joins("INNER JOIN accounts ON accounts.id = staffs.account_id AND accounts.deleted_at IS NULL").
		Where("staffs.id = ?", staffID).
		Where("staffs.deleted_at IS NULL").
		Where("staffs.is_active = TRUE").
		Where("accounts.is_active = TRUE").
		Where("accounts.is_system_admin = TRUE").
		Count(&count).Error
	if err != nil {
		return false, apperrors.Wrap(err, "failed to verify system admin entered_by actor")
	}
	return count > 0, nil
}
