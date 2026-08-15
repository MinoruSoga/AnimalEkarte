package reservation

import (
	"context"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ReservationStaffWriteGuard is the reservation-owned, consumer-minimal port
// used to protect appointment writes that reference staff. Its repository
// implementation locks the staff identity and active clinic assignment in
// FindByID, then locks the exact reservation-type capability in
// SupportsReservationType. Callers must invoke it inside their write
// transaction so those SHARE locks are held through commit.
type ReservationStaffWriteGuard interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Staff, error)
	SupportsReservationType(ctx context.Context, clinicID, staffID, reservationTypeID uint64) (bool, error)
}

func ValidateReservationStaffCapability(ctx context.Context, guard ReservationStaffWriteGuard, clinicID uint64, doctorID *uint64, reservationTypeID uint64) error {
	return validateReservationStaffCapability(ctx, guard, clinicID, doctorID, reservationTypeID, false)
}

// ValidateLineReservationStaffCapability applies the public LIFF boundary in addition to
// clinic assignment and reservation-type capability. Internal reservation workflows may assign
// inactive or hidden staff, so public availability is intentionally enforced only here.
func ValidateLineReservationStaffCapability(ctx context.Context, guard ReservationStaffWriteGuard, clinicID uint64, doctorID *uint64, reservationTypeID uint64) error {
	return validateReservationStaffCapability(ctx, guard, clinicID, doctorID, reservationTypeID, true)
}

func validateReservationStaffCapability(
	ctx context.Context,
	guard ReservationStaffWriteGuard,
	clinicID uint64,
	doctorID *uint64,
	reservationTypeID uint64,
	requireReservationVisible bool,
) error {
	if doctorID == nil || *doctorID == 0 {
		return nil
	}
	if guard == nil {
		return apperrors.WrapInternalServerError("reservation staff write guard is required")
	}
	if reservationTypeID == 0 {
		return apperrors.WrapInternalServerError("reservation type is required to verify staff capability")
	}
	staff, err := guard.FindByID(ctx, clinicID, *doctorID)
	if err != nil {
		return apperrors.Wrap(err, "failed to verify reservation staff")
	}
	supports, err := guard.SupportsReservationType(ctx, clinicID, *doctorID, reservationTypeID)
	if err != nil {
		return apperrors.Wrap(err, "failed to get staff reservation capabilities")
	}
	if requireReservationVisible && (!staff.IsActive || !staff.ReservationVisible) {
		return apperrors.WrapInvalidInput("選択した担当者はLINE予約では指定できません")
	}
	if !supports {
		return apperrors.WrapInvalidInput("選択した担当者はこの予約区分に対応していません")
	}
	return nil
}
