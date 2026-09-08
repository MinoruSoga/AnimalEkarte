package staff

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

func rejectReservedClinicID(clinicID uint64) error {
	if clinicID == 1 || clinicID == 2 {
		return apperrors.WrapInvalidInput(fmt.Sprintf("clinic_id %d is reserved", clinicID))
	}
	return nil
}

// CreateSyntheticClosingStaff は S09 合成 fixture の staffs 行を write owner 経由で作る。
func CreateSyntheticClosingStaff(ctx context.Context, db *gorm.DB, staff *model.Staff) error {
	if staff == nil {
		return apperrors.WrapInvalidInput("staff is required")
	}
	if err := rejectReservedClinicID(staff.ClinicID); err != nil {
		return err
	}
	if db == nil {
		return apperrors.WrapInvalidInput("db is required")
	}
	return NewRepository(db).Create(ctx, staff)
}

// UnscopedDeleteSyntheticClosingStaffs は合成 clinic の staffs だけを物理削除する。
func UnscopedDeleteSyntheticClosingStaffs(ctx context.Context, db *gorm.DB, clinicID uint64) error {
	if clinicID == 0 {
		return apperrors.WrapInvalidInput("clinic id is required")
	}
	if err := rejectReservedClinicID(clinicID); err != nil {
		return err
	}
	if db == nil {
		return apperrors.WrapInvalidInput("db is required")
	}
	if err := persistence.DBOrTx(ctx, db).Unscoped().Where("clinic_id = ?", clinicID).Delete(&model.Staff{}).Error; err != nil {
		return apperrors.FromGORM(err, "staff", fmt.Sprintf("clinic_id=%d", clinicID))
	}
	return nil
}
