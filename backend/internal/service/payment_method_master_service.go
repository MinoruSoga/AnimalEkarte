package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// CreatePaymentMethodInput は支払方法作成の入力
type CreatePaymentMethodInput struct {
	Name         string
	DisplayOrder int
}

// UpdatePaymentMethodInput は支払方法更新の入力
type UpdatePaymentMethodInput struct {
	Name         *string
	DisplayOrder *int
	IsActive     *bool
}

func buildPaymentMethodUpdateFields(input UpdatePaymentMethodInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields["name"] = *input.Name
	}
	if input.DisplayOrder != nil {
		fields["display_order"] = *input.DisplayOrder
	}
	if input.IsActive != nil {
		fields["is_active"] = *input.IsActive
	}
	return fields
}

// PaymentMethodMasterService は支払方法マスタのビジネスロジックインターフェース
type PaymentMethodMasterService interface {
	List(ctx context.Context, clinicID uint64) ([]model.PaymentMethodMaster, error)
	Create(ctx context.Context, clinicID uint64, input CreatePaymentMethodInput) (*model.PaymentMethodMaster, error)
	Update(ctx context.Context, clinicID, id uint64, input UpdatePaymentMethodInput) (*model.PaymentMethodMaster, error)
	Delete(ctx context.Context, clinicID, id uint64) error
}

type paymentMethodMasterService struct {
	repo repository.PaymentMethodMasterRepository
}

// NewPaymentMethodMasterService は PaymentMethodMasterService を初期化して返す
func NewPaymentMethodMasterService(repo repository.PaymentMethodMasterRepository) PaymentMethodMasterService {
	return &paymentMethodMasterService{repo: repo}
}

func (s *paymentMethodMasterService) List(ctx context.Context, clinicID uint64) ([]model.PaymentMethodMaster, error) {
	return s.repo.FindAll(ctx, clinicID)
}

func (s *paymentMethodMasterService) Create(ctx context.Context, clinicID uint64, input CreatePaymentMethodInput) (*model.PaymentMethodMaster, error) {
	slog.InfoContext(ctx, "creating payment method",
		slog.Uint64("clinic_id", clinicID),
		slog.String("name", input.Name))
	m := &model.PaymentMethodMaster{
		ClinicID:     clinicID,
		Name:         input.Name,
		DisplayOrder: input.DisplayOrder,
	}
	return s.repo.Create(ctx, m)
}

func (s *paymentMethodMasterService) Update(ctx context.Context, clinicID, id uint64, input UpdatePaymentMethodInput) (*model.PaymentMethodMaster, error) {
	fields := buildPaymentMethodUpdateFields(input)
	if len(fields) == 0 {
		return s.repo.FindByID(ctx, clinicID, id)
	}
	return s.repo.UpdateFields(ctx, clinicID, id, fields)
}

func (s *paymentMethodMasterService) Delete(ctx context.Context, clinicID, id uint64) error {
	count, err := s.repo.CountUsageByID(ctx, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check payment method usage")
	}
	if count > 0 {
		return apperrors.WrapConflict("この支払方法は使用中のため削除できません")
	}
	slog.InfoContext(ctx, "deleting payment method",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("id", id))
	return s.repo.Delete(ctx, clinicID, id)
}
