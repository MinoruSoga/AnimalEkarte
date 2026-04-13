package service

import (
	"context"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// LineCustomerService は予約顧客のビジネスロジックインターフェース
type LineCustomerService interface {
	List(ctx context.Context, clinicID uint64) ([]model.LineCustomer, error)
	LinkOwner(ctx context.Context, clinicID, id uint64, ownerID *uint64) (*model.LineCustomer, error)
}

type lineCustomerService struct {
	repo repository.LineCustomerRepository
}

func NewLineCustomerService(repo repository.LineCustomerRepository) LineCustomerService {
	return &lineCustomerService{repo: repo}
}

func (s *lineCustomerService) List(ctx context.Context, clinicID uint64) ([]model.LineCustomer, error) {
	items, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list reservation customers")
	}
	return items, nil
}

func (s *lineCustomerService) LinkOwner(ctx context.Context, clinicID, id uint64, ownerID *uint64) (*model.LineCustomer, error) {
	if err := s.repo.UpdateOwnerLink(ctx, clinicID, id, ownerID); err != nil {
		return nil, apperrors.Wrap(err, "failed to link owner to reservation customer")
	}
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get reservation customer")
	}
	return result, nil
}
