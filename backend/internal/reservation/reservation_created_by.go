package reservation

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// assertReservationCreatedBy validates the recording actor in the insert
// transaction. HTTP callers derive this ID from authentication, never JSON.
// Internal/LINE creation may have no staff actor. The database FK preserves
// identity; current authorization comes from assignments or the admin account.
func assertReservationCreatedBy(ctx context.Context, db *gorm.DB, clinicID uint64, staffID *uint64) error {
	if staffID == nil {
		return nil
	}
	if *staffID == 0 {
		return apperrors.WrapInvalidInput("created_by must be greater than zero")
	}
	if db == nil || persistence.TxFromContext(ctx) == nil {
		return apperrors.WrapInternalServerError("reservation actor validation requires an active transaction")
	}

	tx := persistence.DBOrTx(ctx, db)
	var staff model.Staff
	err := tx.Select("id", "account_id", "is_active").
		Clauses(clause.Locking{Strength: "SHARE"}).
		Where("id = ?", *staffID).First(&staff).Error
	if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && !staff.IsActive) {
		return reservationCreatorForbidden()
	}
	if err != nil {
		return apperrors.FromGORM(err, "reservation actor", "")
	}

	var assignment model.StaffClinicAssignment
	err = tx.Select("id", "staff_id", "clinic_id").
		Clauses(clause.Locking{Strength: "SHARE"}).
		Where("staff_id = ? AND clinic_id = ?", *staffID, clinicID).
		First(&assignment).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return apperrors.FromGORM(err, "reservation actor assignment", "")
	}
	if staff.AccountID == nil {
		return reservationCreatorForbidden()
	}

	// Verify the database authority and hold the account lock until commit so
	// concurrent admin revocation cannot invalidate the authorization we use.
	var account model.Account
	err = tx.Select("id").Clauses(clause.Locking{Strength: "SHARE"}).
		Where("id = ? AND is_active = TRUE AND is_system_admin = TRUE", *staff.AccountID).
		First(&account).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return reservationCreatorForbidden()
	}
	if err != nil {
		return apperrors.FromGORM(err, "reservation actor admin account", "")
	}
	return nil
}

func reservationCreatorForbidden() error {
	return apperrors.WrapForbidden("この医院の予約を登録する権限がありません")
}

// A creator is historical attribution, independent of current assignments or
// admin privileges. Only the public summary is loaded; the parent appointment
// remains clinic-scoped. Deleted staff retain their ID through created_by.
func reservationCreatedByStaffPreload(db *gorm.DB) *gorm.DB {
	return db.Select("id", "name").Where("staffs.deleted_at IS NULL")
}
