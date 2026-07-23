package reservation

import (
	"context"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ReservationStaffCapabilityView is the read-only capability required by staff validation.
type ReservationStaffCapabilityView interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Staff, error)
	SupportsReservationType(ctx context.Context, clinicID, staffID, reservationTypeID uint64) (bool, error)
}

func ValidateReservationStaffCapability(ctx context.Context, repo ReservationStaffCapabilityView, clinicID uint64, doctorID *uint64, reservationTypeID uint64) error {
	return validateReservationStaffCapability(ctx, repo, clinicID, doctorID, reservationTypeID, false)
}

// ValidateLineReservationStaffCapability applies the public LIFF boundary in addition to
// clinic assignment and reservation-type capability. Internal reservation workflows may assign
// inactive or hidden staff, so public availability is intentionally enforced only here.
func ValidateLineReservationStaffCapability(ctx context.Context, repo ReservationStaffCapabilityView, clinicID uint64, doctorID *uint64, reservationTypeID uint64) error {
	return validateReservationStaffCapability(ctx, repo, clinicID, doctorID, reservationTypeID, true)
}

func validateReservationStaffCapability(
	ctx context.Context,
	repo ReservationStaffCapabilityView,
	clinicID uint64,
	doctorID *uint64,
	reservationTypeID uint64,
	requireReservationVisible bool,
) error {
	if doctorID == nil || *doctorID == 0 {
		return nil
	}
	if repo == nil {
		return apperrors.WrapInternalServerError("reservation staff repository is required")
	}
	if reservationTypeID == 0 {
		return apperrors.WrapInternalServerError("reservation type is required to verify staff capability")
	}
	staff, err := repo.FindByID(ctx, clinicID, *doctorID)
	if err != nil {
		return apperrors.Wrap(err, "failed to verify reservation staff")
	}
	if requireReservationVisible && (!staff.IsActive || !staff.ReservationVisible) {
		return apperrors.WrapInvalidInput("選択した担当者はLINE予約では指定できません")
	}
	supports, err := repo.SupportsReservationType(ctx, clinicID, *doctorID, reservationTypeID)
	if err != nil {
		return apperrors.Wrap(err, "failed to get staff reservation capabilities")
	}
	if !supports {
		return apperrors.WrapInvalidInput("選択した担当者はこの予約区分に対応していません")
	}
	return nil
}
