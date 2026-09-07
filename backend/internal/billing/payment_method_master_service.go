package billing

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// CreatePaymentMethodMasterInput は支払方法作成の入力
type CreatePaymentMethodMasterInput struct {
	Name         string
	DisplayOrder int
}

// UpdatePaymentMethodMasterInput は支払方法更新の入力
type UpdatePaymentMethodMasterInput struct {
	Name         *string
	DisplayOrder *int
	IsActive     *bool
}

const (
	colPaymentMethodName         = "name"
	colPaymentMethodDisplayOrder = "display_order"
	colPaymentMethodIsActive     = "is_active"
)

func buildPaymentMethodUpdate(input *UpdatePaymentMethodMasterInput) map[string]any {
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

// isSystemPaymentMethod は system_key を持つ予約済み支払方法（現金・クレジット等）かどうかを返す。
// system_key は immutable かつ編集 UI 非公開。システム行の無効化・削除は禁止する。
func isSystemPaymentMethod(m *model.PaymentMethodMaster) bool {
	return m != nil && m.SystemKey != nil && *m.SystemKey != ""
}

// PaymentMethodMasterService は支払方法マスタのビジネスロジックインターフェース
type PaymentMethodMasterService interface {
	List(ctx context.Context, clinicID uint64) ([]model.PaymentMethodMaster, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.PaymentMethodMaster, error)
	Create(ctx context.Context, clinicID uint64, input *CreatePaymentMethodMasterInput) (*model.PaymentMethodMaster, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdatePaymentMethodMasterInput) (*model.PaymentMethodMaster, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type paymentMethodMasterService struct {
	repo PaymentMethodMasterRepository
}

// NewPaymentMethodMasterService は PaymentMethodMasterService を初期化して返す
func NewPaymentMethodMasterService(repo PaymentMethodMasterRepository) PaymentMethodMasterService {
	return &paymentMethodMasterService{repo: repo}
}

func (s *paymentMethodMasterService) List(ctx context.Context, clinicID uint64) ([]model.PaymentMethodMaster, error) {
	items, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list payment methods")
	}
	return items, nil
}

func (s *paymentMethodMasterService) GetByID(ctx context.Context, clinicID, id uint64) (*model.PaymentMethodMaster, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get payment method")
	}
	return result, nil
}

func (s *paymentMethodMasterService) Create(ctx context.Context, clinicID uint64, input *CreatePaymentMethodMasterInput) (*model.PaymentMethodMaster, error) {
	if err := sharedkernel.ValidateRequiredName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate required name")
	}
	m := &model.PaymentMethodMaster{
		ClinicID:     clinicID,
		Name:         input.Name,
		DisplayOrder: input.DisplayOrder,
	}
	result, err := s.repo.Create(ctx, m)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to create payment method")
	}
	slog.InfoContext(ctx, "payment method created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("id", result.ID))
	return result, nil
}

func (s *paymentMethodMasterService) Update(ctx context.Context, clinicID, id uint64, input *UpdatePaymentMethodMasterInput) (*model.PaymentMethodMaster, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(sharedkernel.ErrMsgInputNotNil)
	}
	existing, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get payment method")
	}
	// システム標準行は無効化不可（名称・表示順の更新は許可）
	if isSystemPaymentMethod(existing) && input.IsActive != nil && !*input.IsActive {
		return nil, apperrors.WrapConflict("システム標準の支払方法は無効化できません")
	}
	if err := sharedkernel.ValidateOptionalName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate optional name")
	}
	fields := buildPaymentMethodUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(sharedkernel.ErrMsgAtLeastOneField)
	}
	result, err := s.repo.Update(ctx, clinicID, id, *input)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update payment method")
	}
	slog.InfoContext(ctx, "payment method updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("id", id))
	return result, nil
}

func (s *paymentMethodMasterService) Delete(ctx context.Context, clinicID, id uint64) error {
	existing, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to get payment method")
	}
	// システム標準行は使用有無に関わらず削除不可
	if isSystemPaymentMethod(existing) {
		return apperrors.WrapConflict("システム標準の支払方法は削除できません")
	}
	count, err := s.repo.CountUsageByPaymentMethodID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check payment method usage")
	}
	if count > 0 {
		return apperrors.WrapConflict("この支払方法は使用中のため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete payment method")
	}
	slog.InfoContext(ctx, "payment method deleted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("id", id))
	return nil
}

func (s *paymentMethodMasterService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(sharedkernel.ErrMsgIDsNotEmpty)
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		return apperrors.Wrap(err, "failed to reorder payment methods")
	}
	slog.InfoContext(ctx, "payment methods reordered",
		slog.Uint64("clinic_id", clinicID),
		slog.Int("count", len(ids)))
	return nil
}
