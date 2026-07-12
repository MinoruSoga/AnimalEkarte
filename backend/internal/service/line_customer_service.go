package service

import (
	"context"
	"log/slog"

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
		slog.ErrorContext(ctx, "failed to list reservation customers", "error", err)
		return nil, apperrors.Wrap(err, "failed to list reservation customers")
	}
	return items, nil
}

func (s *lineCustomerService) LinkOwner(ctx context.Context, clinicID, id uint64, ownerID *uint64) (*model.LineCustomer, error) {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to find line customer", "error", err)
		return nil, apperrors.Wrap(err, "failed to find line customer")
	}
	if err := s.repo.UpdateOwnerLink(ctx, clinicID, id, ownerID); err != nil {
		slog.ErrorContext(ctx, "failed to link owner to reservation customer", "error", err)
		return nil, apperrors.Wrap(err, "failed to link owner to reservation customer")
	}
	slog.InfoContext(ctx, "line customer owner linked",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("line_customer_id", id),
	)
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get reservation customer", "error", err)
		return nil, apperrors.Wrap(err, "failed to get reservation customer")
	}
	return result, nil
}
