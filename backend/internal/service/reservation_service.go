package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

type ReservationService interface {
	List(ctx context.Context, clinicID uint64, page, limit int, date *time.Time, status *string, petID, ownerID *uint64) ([]model.ReservationAppointment, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.ReservationAppointment, error)
	Create(ctx context.Context, reservation *model.ReservationAppointment) error
	Update(ctx context.Context, reservation *model.ReservationAppointment) error
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
		return fmt.Errorf("failed to create reservation: %w", err)
	}
	slog.InfoContext(ctx, "reservation created",
		slog.Uint64("reservation_id", reservation.ID),
		slog.Uint64("clinic_id", reservation.ClinicID))
	return nil
}

func (s *reservationService) Update(ctx context.Context, reservation *model.ReservationAppointment) error {
	if err := s.repo.Update(ctx, reservation); err != nil {
		return fmt.Errorf("failed to update reservation: %w", err)
	}
	slog.InfoContext(ctx, "reservation updated",
		slog.Uint64("reservation_id", reservation.ID),
		slog.Uint64("clinic_id", reservation.ClinicID))
	return nil
}

func (s *reservationService) Delete(ctx context.Context, clinicID, id uint64) err
r {
	return s.repo.Delete(ctx, clinicID, id)
}
