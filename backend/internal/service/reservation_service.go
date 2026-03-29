package service

import (
	"context"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

type ReservationService interface {
	List(ctx context.Context, clinicID uint64, page, limit int, date *time.Time, status *string, petID, ownerID *uint64) ([]model.ReservationAppointment, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.ReservationAppointment, error)
	Create(ctx context.Context, reservation *model.ReservationAppointment) error
	Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationInput) (*model.ReservationAppointment, error)
	Delete(ctx context.Context, clinicID, id uint64) error
}

type reservationService struct {
	repo repository.ReservationRepository
}

func NewReservationService(repo repository.ReservationRepository) ReservationService {
	return &reservationService{repo: repo}
}

func (s *reservationService) List(ctx context.Context, clinicID uint64, page, limit int, date *time.Time, status *string, petID, ownerID *uint64) ([]model.ReservationAppointment, int64, error) {
	return s.repo.FindAll(ctx, clinicID, page, limit, date, status, petID, ownerID)
}

func (s *reservationService) GetByID(ctx context.Context, clinicID, id uint64) (*model.ReservationAppointment, error) {
	return s.repo.FindByID(ctx, clinicID, id)
}

func (s *reservationService) Create(ctx context.Context, reservation *model.ReservationAppointment) error {
	if err := s.repo.Create(ctx, reservation); err != nil {
		return apperrors.Wrap(err, "failed to create reservation")
	}
	slog.InfoContext(ctx, "reservation created",
		slog.Uint64("reservation_id", reservation.ID),
		slog.Uint64("clinic_id", reservation.ClinicID))
	return nil
}

func (s *reservationService) Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationInput) (*model.ReservationAppointment, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("input must not be nil")
	}
	fields := buildReservationUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	reservation, err := s.repo.UpdateFields(ctx, clinicID, id, fields)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update reservation")
	}
	slog.InfoContext(ctx, "reservation updated",
		slog.Uint64("reservation_id", id),
		slog.Uint64("clinic_id", clinicID))
	return reservation, nil
}

// UpdateReservationInput は予約更新のサービス入力 DTO
type UpdateReservationInput struct {
	StartTime     *time.Time
	EndTime       *time.Time
	OwnerID       *uint64
	PetID         *uint64
	VisitType     *model.VisitType
	ServiceTypeID *uint64
	DoctorID      *uint64
	IsDesignated  *bool
	Status        *model.ReservationStatus
	Notes         *string
}

func buildReservationUpdateFields(input *UpdateReservationInput) map[string]any {
	fields := make(map[string]any)
	if input.StartTime != nil {
		fields["start_time"] = *input.StartTime
	}
	if input.EndTime != nil {
		fields["end_time"] = *input.EndTime
	}
	if input.OwnerID != nil {
		fields["owner_id"] = *input.OwnerID
	}
	if input.PetID != nil {
		fields["pet_id"] = *input.PetID
	}
	if input.VisitType != nil {
		fields["visit_type"] = *input.VisitType
	}
	if input.ServiceTypeID != nil {
		fields["service_type_id"] = *input.ServiceTypeID
	}
	if input.DoctorID != nil {
		fields["doctor_id"] = *input.DoctorID
	}
	if input.IsDesignated != nil {
		fields["is_designated"] = *input.IsDesignated
	}
	if input.Status != nil {
		fields["status"] = *input.Status
	}
	if input.Notes != nil {
		fields["notes"] = *input.Notes
	}
	return fields
}

func (s *reservationService) Delete(ctx context.Context, clinicID, id uint64) error {
	return s.repo.Delete(ctx, clinicID, id)
}
