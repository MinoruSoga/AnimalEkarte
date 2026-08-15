package reservation

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- 職種紐付け ----

func (s *reservationTypeService) ListOccupations(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeOccupation, error) {
	if _, err := s.repo.FindByID(ctx, clinicID, reservationTypeID); err != nil {
		slog.ErrorContext(ctx, "failed to get reservation type", "error", err)
		return nil, apperrors.Wrap(err, "failed to get reservation type")
	}
	items, err := s.occupationRepo.FindAll(ctx, clinicID, reservationTypeID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list occupations", "error", err, "id", reservationTypeID, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to list occupations")
	}
	return items, nil
}

func (s *reservationTypeService) LinkOccupation(ctx context.Context, clinicID, reservationTypeID, occupationID uint64) (*model.ReservationTypeOccupation, error) {
	if _, err := s.repo.FindByID(ctx, clinicID, reservationTypeID); err != nil {
		slog.ErrorContext(ctx, "failed to get reservation type", "error", err)
		return nil, apperrors.Wrap(err, "failed to get reservation type")
	}
	if _, err := s.baseOccupationRepo.FindByID(ctx, clinicID, occupationID); err != nil {
		slog.ErrorContext(ctx, "occupation not found", "error", err)
		return nil, apperrors.Wrap(err, "occupation not found")
	}
	o := &model.ReservationTypeOccupation{
		ClinicID:          clinicID,
		ReservationTypeID: reservationTypeID,
		OccupationID:      occupationID,
	}
	if err := s.occupationRepo.Create(ctx, o); err != nil {
		slog.ErrorContext(ctx, "failed to link occupation", "error", err, "id", reservationTypeID, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to link occupation")
	}
	slog.InfoContext(ctx, "occupation linked",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("reservation_type_id", reservationTypeID),
		slog.Uint64("occupation_id", occupationID))
	result, err := s.occupationRepo.FindByID(ctx, clinicID, reservationTypeID, occupationID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get linked occupation", "error", err, "id", reservationTypeID, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get linked occupation")
	}
	return result, nil
}

func (s *reservationTypeService) UnlinkOccupation(ctx context.Context, clinicID, reservationTypeID, occupationID uint64) error {
	if _, err := s.occupationRepo.FindByID(ctx, clinicID, reservationTypeID, occupationID); err != nil {
		return apperrors.Wrap(err, "failed to find reservation type occupation")
	}
	if err := s.occupationRepo.Delete(ctx, clinicID, reservationTypeID, occupationID); err != nil {
		slog.ErrorContext(ctx, "failed to unlink occupation", "error", err, "id", reservationTypeID, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to unlink occupation")
	}
	slog.InfoContext(ctx, "occupation unlinked",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("reservation_type_id", reservationTypeID),
		slog.Uint64("occupation_id", occupationID))
	return nil
}
