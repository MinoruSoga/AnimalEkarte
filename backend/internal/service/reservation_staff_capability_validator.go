package service

import (
	"context"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/repository"
)

func validateReservationStaffCapability(ctx context.Context, repo repository.ReservationStaffRepository, clinicID uint64, doctorID *uint64, reservationTypeID uint64) error {
	if repo == nil || doctorID == nil || *doctorID == 0 || reservationTypeID == 0 {
		return nil
	}
	if _, err := repo.FindByID(ctx, clinicID, *doctorID); err != nil {
		return apperrors.Wrap(err, "failed to verify reservation staff")
	}
	excluded, err := repo.FindAllExcludedReservationTypes(ctx, *doctorID)
	if err != nil {
		return apperrors.Wrap(err, "failed to get excluded reservation types")
	}
	for i := range excluded {
		if excluded[i].ReservationTypeID == reservationTypeID {
			return apperrors.WrapInvalidInput("選択した担当者はこの予約区分に対応していません")
		}
	}
	return nil
}
