package staff

import (
	"context"
	"errors"
	"fmt"
	"strings"

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

// FindOrCreateSyntheticAuditSentinelStaff は EMR-211 (b) で audit_logs の RESTRICT
// 参照先となる sentinel staff を (clinic_id, name) で find-or-create する。
// login 不能（account_id NULL）・StaffTypeResource・is_active=false・
// reservation_visible=false の sentinel 不変条件は write owner であるこの package 側で
// 立てる。呼び出し側が sentinel 作成区間の advisory lock を持つ前提。
// sentinel clinic は s09 接頭辞を持たない通常行であり小さい ID を取り得るため、
// s09 fixture 専用の rejectReservedClinicID ガードは適用しない。
func FindOrCreateSyntheticAuditSentinelStaff(ctx context.Context, db *gorm.DB, clinicID uint64, name string) (*model.Staff, error) {
	if clinicID == 0 {
		return nil, apperrors.WrapInvalidInput("clinic id is required")
	}
	if strings.TrimSpace(name) == "" {
		return nil, apperrors.WrapInvalidInput("sentinel staff name is required")
	}
	if db == nil {
		return nil, apperrors.WrapInvalidInput("db is required")
	}
	tx := persistence.DBOrTx(ctx, db)
	var staffRow model.Staff
	err := tx.Where("clinic_id = ? AND name = ?", clinicID, name).Take(&staffRow).Error
	switch {
	case err == nil:
	case errors.Is(err, gorm.ErrRecordNotFound):
		staffRow = model.Staff{
			ClinicID:  clinicID,
			AccountID: nil, // login 不能: sentinel にアカウントは持たせない
			Name:      name,
			StaffType: model.StaffTypeResource,
		}
		if createErr := tx.Create(&staffRow).Error; createErr != nil {
			return nil, apperrors.Wrap(createErr, "create audit sentinel staff")
		}
		// staffs.is_active / reservation_visible も default:true の zero 値 trap が
		// あるため明示 UPDATE で false にする。
		if updateErr := tx.Model(&model.Staff{}).Where("id = ?", staffRow.ID).
			Updates(map[string]any{"is_active": false, "reservation_visible": false}).Error; updateErr != nil {
			return nil, apperrors.Wrap(updateErr, "disable audit sentinel staff")
		}
		staffRow.IsActive = false
		staffRow.ReservationVisible = false
	default:
		return nil, apperrors.Wrap(err, "load audit sentinel staff")
	}
	return &staffRow, nil
}
