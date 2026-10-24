package lstep

import (
	"context"

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
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return nil, apperrors.Wrap(err, "failed to find line customer")
	}
	if ownerID != nil {
		if _, err := s.ownerRepo.FindByID(ctx, clinicID, *ownerID); err != nil {
			return nil, apperrors.Wrap(err, "owner not found")
		}
	}
	// EMR-504 / EMR-549 C1: line_customers.owner_id は DROP 済みで、置換先 owner_line_binding_states の
	// writer（EMR-469 W09）は未接続。成功を装わず固定 422 で拒否する（旧列へのフォールバックなし）。
	return nil, apperrors.WrapUnprocessableWithCode(
		apperrors.CodeLineBindingPending,
		"LINE顧客と飼い主の紐付けは現在変更できません（紐付け機能の準備中です）",
	)
}
