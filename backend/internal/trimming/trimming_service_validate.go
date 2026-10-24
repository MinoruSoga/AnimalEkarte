package trimming

import (
	"context"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *trimmingService) validateTrimmingReservationType(ctx context.Context, clinicID, reservationTypeID uint64, requireActive bool) error {
	if s.reservationType == nil {
		return apperrors.WrapInternalServerError("reservation type repository is required")
	}

	reservationType, err := s.reservationType.FindByID(ctx, clinicID, reservationTypeID)
	if err != nil {
		return apperrors.Wrap(err, "failed to get trimming reservation type")
	}
	if !reservationType.HasBusinessKind(model.ReservationBusinessKindTrimming) {
		return apperrors.WrapInvalidInput("reservation_type_id must be a trimming reservation type")
	}
	if requireActive && !reservationType.IsActive {
		return apperrors.WrapInvalidInput("reservation_type_id references an inactive trimming reservation type")
	}
	return nil
}

func (s *trimmingService) requireBookingConstraintDependencies() error {
	if s.unavailableTime == nil {
		return apperrors.WrapInternalServerError("reservation type unavailable-time repository is required")
	}
	if s.reservation == nil {
		return apperrors.WrapInternalServerError("reservation repository is required")
	}
	if s.reservationType == nil {
		return apperrors.WrapInternalServerError("reservation type repository is required")
	}
	return nil
}
