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
// shift_entries は staffs の子（shift_entry_breaks は CASCADE で追随）なので、
// write-owner であるこの package 内で staffs より先に同じ clinic スコープで消す。
// testdb の AutoMigrate スキーマには shift_entries が無い場合があるため存在確認する。
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
	var shiftEntriesExists bool
	if err := persistence.DBOrTx(ctx, db).Raw(
		"SELECT to_regclass('shift_entries') IS NOT NULL",
	).Scan(&shiftEntriesExists).Error; err != nil {
		return apperrors.Wrap(err, "check shift_entries existence")
	}
	if shiftEntriesExists {
		if err := persistence.DBOrTx(ctx, db).Exec("DELETE FROM shift_entries WHERE clinic_id = ?", clinicID).Error; err != nil {
			return apperrors.FromGORM(err, "shift_entry", fmt.Sprintf("clinic_id=%d", clinicID))
		}
	}
	if err := persistence.DBOrTx(ctx, db).Unscoped().Where("clinic_id = ?", clinicID).Delete(&model.Staff{}).Error; err != nil {
		return apperrors.FromGORM(err, "staff", fmt.Sprintf("clinic_id=%d", clinicID))
	}
	return nil
}
