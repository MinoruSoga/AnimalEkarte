package service

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

type ReservationService interface {
	List(ctx context.Context, clinicID uint64, page, limit int, date *time.Time, status *string, petID *uint64, ownerID *uint64) ([]model.ReservationAppointment, int64, error)
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

func (s *reservationService) List(ctx context.Context, clinicID uint64, page, limit int, date *time.Time, status *string, petID *uint64, ownerID *uint64) ([]model.ReservationAppointment, int64, error) {
	return s.repo.FindAll(ctx, clinicID, page, limit, date, status, petID, ownerID)
}

func (s *reservationService) GetByID(ctx context.Context, clinicID, id uint64) (*model.ReservationAppointment, error) {
	return s.repo.FindByID(ctx, clinicID, id)
}

func (s *reservationService) Create(ctx context.Context, reservation *model.ReservationAppointment) error {
	return s.repo.Create(ctx, reservation)
}

func (s *reservationService) Update(ctx context.Context, reservation *model.ReservationAppointment) error {
	return s.repo.Update(ctx, reservation)
}

func (s *reservationService) Delete(ctx context.Context, clinicID, id uint64) error {
	return s.repo.Delete(ctx, clinicID, id)
}
