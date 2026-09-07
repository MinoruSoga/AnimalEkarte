package lstep

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// LineCustomerListResult carries list rows plus truncation metadata (G2F-05).
// Clients must not treat Items as the full clinic set when Truncated is true.
type LineCustomerListResult struct {
	Items     []model.LineCustomer
	Total     int64
	Limit     int
	Truncated bool
}

// LineCustomerService は予約顧客のビジネスロジックインターフェース
type LineCustomerService interface {
	List(ctx context.Context, clinicID uint64) (*LineCustomerListResult, error)
	LinkOwner(ctx context.Context, clinicID, id uint64, ownerID *uint64) (*model.LineCustomer, error)
}

type lineCustomerService struct {
	repo      LineCustomerRepository
	ownerRepo lstepOwnerRepo
}

func NewLineCustomerService(repo LineCustomerRepository, ownerRepo lstepOwnerRepo) LineCustomerService {
	return &lineCustomerService{repo: repo, ownerRepo: ownerRepo}
}

func (s *lineCustomerService) List(ctx context.Context, clinicID uint64) (*LineCustomerListResult, error) {
	items, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list reservation customers")
	}
	total, err := s.repo.CountAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to count reservation customers")
	}
	return &LineCustomerListResult{
		Items:     items,
		Total:     total,
		Limit:     lineCustomerListMax,
		Truncated: total > int64(lineCustomerListMax),
	}, nil
}

func (s *lineCustomerService) LinkOwner(ctx context.Context, clinicID, id uint64, ownerID *uint64) (*model.LineCustomer, error) {
	existing, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find line customer")
	}
	if ownerID != nil {
		if _, err := s.ownerRepo.FindByID(ctx, clinicID, *ownerID); err != nil {
			return nil, apperrors.Wrap(err, "owner not found")
		}
	}
	if err := s.repo.UpdateOwnerLink(ctx, clinicID, id, ownerID); err != nil {
		return nil, apperrors.Wrap(err, "failed to link owner to reservation customer")
	}
	slog.InfoContext(ctx, "line customer owner linked",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("line_customer_id", id),
	)
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		// G2A-01 / CODING_RULES.md:78 — write already committed; do not invert success into failure.
		slog.ErrorContext(ctx, "reload after line customer owner link failed; returning preloaded row with updated owner_id",
			"error", err, "clinic_id", clinicID, "line_customer_id", id)
		existing.OwnerID = ownerID
		if ownerID == nil {
			existing.Owner = nil
		}
		return existing, nil
	}
	return result, nil
}
