package service

import (
	"context"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ReservationCustomerService は予約顧客のビジネスロジックインターフェース
type ReservationCustomerService interface {
	List(ctx context.Context, clinicID uint64) ([]model.ReservationCustomer, error)
	LinkOwner(ctx context.Context, clinicID, id uint64, ownerID *uint64) (*model.ReservationCustomer, error)
}

type reservationCustomerService struct {
	repo repository.ReservationCustomerRepository
}

func NewReservationCustomerService(repo repository.ReservationCustomerRepository) ReservationCustomerService {
	return &reservationCustomerService{repo: repo}
}

func (s *reservationCustomerService) List(ctx context.Context, clinicID uint64) ([]model.ReservationCustomer, error) {
	items, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list reservation customers")
	}
	return items, nil
}

func (s *reservationCustomerService) LinkOwner(ctx context.Context, clinicID, id uint64, ownerID *uint64) (*model.ReservationCustomer, error) {
	if err := s.repo.UpdateOwnerLink(ctx, clinicID, id, ownerID); err != nil {
		return nil, apperrors.Wrap(err, "failed to link owner to reservation customer")
	}
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get reservation customer")
	}
	return result, nil
}
