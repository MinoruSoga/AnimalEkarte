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

const (
	colPaymentMethodName         = "name"
	colPaymentMethodDisplayOrder = "display_order"
	colPaymentMethodIsActive     = "is_active"
)

func buildPaymentMethodUpdate(input *UpdatePaymentMethodInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colPaymentMethodName] = *input.Name
	}
	if input.DisplayOrder != nil {
		fields[colPaymentMethodDisplayOrder] = *input.DisplayOrder
	}
	if input.IsActive != nil {
		fields[colPaymentMethodIsActive] = *input.IsActive
	}
	return fields
}

// PaymentMethodMasterService は支払方法マスタのビジネスロジックインターフェース
type PaymentMethodMasterService interface {
	List(ctx context.Context, clinicID uint64) ([]model.PaymentMethodMaster, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.PaymentMethodMaster, error)
	Create(ctx context.Context, clinicID uint64, input *CreatePaymentMethodInput) (*model.PaymentMethodMaster, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdatePaymentMethodInput) (*model.PaymentMethodMaster, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type paymentMethodMasterService struct {
	repo repository.PaymentMethodMasterRepository
}

// NewPaymentMethodMasterService は PaymentMethodMasterService を初期化して返す
func NewPaymentMethodMasterService(repo repository.PaymentMethodMasterRepository) PaymentMethodMasterService {
	return &paymentMethodMasterService{repo: repo}
}

func (s *paymentMethodMasterService) List(ctx context.Context, clinicID uint64) ([]model.PaymentMethodMaster, error) {
	items, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list payment methods", "error", err)
		return nil, apperrors.Wrap(err, "failed to list payment methods")
	}
	return items, nil
}

func (s *paymentMethodMasterService) GetByID(ctx context.Context, clinicID, id uint64) (*model.PaymentMethodMaster, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get payment method", "error", err)
		return nil, apperrors.Wrap(err, "failed to get payment method")
	}
	return result, nil
}

func (s *paymentMethodMasterService) Create(ctx context.Context, clinicID uint64, input *CreatePaymentMethodInput) (*model.PaymentMethodMaster, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, err
	}
	m := &model.PaymentMethodMaster{
		ClinicID:     clinicID,
		Name:         input.Name,
		DisplayOrder: input.DisplayOrder,
	}
	result, err := s.repo.Create(ctx, m)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create payment method", "error", err)
		return nil, apperrors.Wrap(err, "failed to create payment method")
	}
	slog.InfoContext(ctx, "payment method created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("id", result.ID))
	return result, nil
}

func (s *paymentMethodMasterService) Update(ctx context.Context, clinicID, id uint64, input *UpdatePaymentMethodInput) (*model.PaymentMethodMaster, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(ErrMsgInputNotNil)
	}
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to get payment method", "error", err)
		return nil, apperrors.Wrap(err, "failed to get payment method")
	}
	if err := validateOptionalName(input.Name); err != nil {
		return nil, err
	}
	fields := buildPaymentMethodUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)
	}
	result, err := s.repo.Update(ctx, clinicID, id, fields)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update payment method", "error", err)
		return nil, apperrors.Wrap(err, "failed to update payment method")
	}
	slog.InfoContext(ctx, "payment method updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("id", id))
	return result, nil
}

func (s *paymentMethodMasterService) Delete(ctx context.Context, clinicID, id uint64) error {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to get payment method", "error", err)
		return apperrors.Wrap(err, "failed to get payment method")
	}
	count, err := s.repo.CountUsageByPaymentMethodID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check payment method usage", "error", err)
		return apperrors.Wrap(err, "failed to check payment method usage")
	}
	if count > 0 {
		return apperrors.WrapConflict("この支払方法は使用中のため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to delete payment method", "error", err)
		return apperrors.Wrap(err, "failed to delete payment method")
	}
	slog.InfoContext(ctx, "payment method deleted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("id", id))
	return nil
}

func (s *paymentMethodMasterService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(ErrMsgIDsNotEmpty)
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		slog.ErrorContext(ctx, "failed to reorder payment methods", "error", err)
		return apperrors.Wrap(err, "failed to reorder payment methods")
	}
	slog.InfoContext(ctx, "payment methods reordered",
		slog.Uint64("clinic_id", clinicID),
		slog.Int("count", len(ids)))
	return nil
}
