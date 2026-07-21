package reservation

import (
	"context"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func ValidateReservationStaffCapability(ctx context.Context, repo ReservationStaffRepository, clinicID uint64, doctorID *uint64, reservationTypeID uint64) error {
	if repo == nil || doctorID == nil || *doctorID == 0 || reservationTypeID == 0 {
		return nil
	}
	if _, err := repo.FindByID(ctx, clinicID, *doctorID); err != nil {
		return apperrors.Wrap(err, "failed to verify reservation staff")
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
