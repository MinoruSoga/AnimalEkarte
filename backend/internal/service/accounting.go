package service

import (
	"context"

	"github.com/google/uuid"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// AccountingService 会計サービスインターフェース
type AccountingService interface {
	GetAllAccounting(ctx context.Context) ([]model.Accounting, error)
	GetAccountingByID(ctx context.Context, id string) (*model.Accounting, error)
	GetAccountingByPetID(ctx context.Context, petID string) ([]model.Accounting, error)
	GetAccountingByOwnerID(ctx context.Context, ownerID string) ([]model.Accounting, error)
	GetAccountingByStatus(ctx context.Context, status string) ([]model.Accounting, error)
	CreateAccounting(ctx context.Context, req *model.CreateAccountingRequest) (*model.Accounting, error)
	UpdateAccounting(ctx context.Context, id string, req *model.UpdateAccountingRequest) (*model.Accounting, error)
	DeleteAccounting(ctx context.Context, id string) error
}

// GetAllAccounting 全ての会計を取得
func (s *Service) GetAllAccounting(ctx context.Context) ([]model.Accounting, error) {
	return s.accountingRepo.GetAllAccounting(ctx)
}

// GetAccountingByID IDで会計を取得
func (s *Service) GetAccountingByID(ctx context.Context, id string) (*model.Accounting, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid accounting ID format")
	}

	acc, err := s.accountingRepo.GetAccountingByID(ctx, uid.String())
	if err != nil {
		return nil, err
	}

	if acc == nil {
		return nil, apperrors.WrapNotFound("accounting with id %s not found", id)
	}

	return acc, nil
}

// GetAccountingByPetID ペットIDで会計を取得
func (s *Service) GetAccountingByPetID(ctx context.Context, petID string) ([]model.Accounting, error) {
	uid, err := uuid.Parse(petID)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid pet ID format")
	}

	return s.accountingRepo.GetAccountingByPetID(ctx, uid.String())
}

// GetAccountingByOwnerID 飼い主IDで会計を取得
func (s *Service) GetAccountingByOwnerID(ctx context.Context, ownerID string) ([]model.Accounting, error) {
	uid, err := uuid.Parse(ownerID)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid owner ID format")
	}

	return s.accountingRepo.GetAccountingByOwnerID(ctx, uid.String())
}

// GetAccountingByStatus ステータスで会計を取得
func (s *Service) GetAccountingByStatus(ctx context.Context, status string) ([]model.Accounting, error) {
	return s.accountingRepo.GetAccountingByStatus(ctx, status)
}

// CreateAccounting 会計を作成
func (s *Service) CreateAccounting(ctx context.Context, req *model.CreateAccountingRequest) (*model.Accounting, error) {
	acc := &model.Accounting{
		ID:              uuid.New(),
		PetID:           req.PetID,
		OwnerID:         req.OwnerID,
		MedicalRecordID: req.MedicalRecordID,
		ScheduledDate:   req.ScheduledDate,
		Status:          "未収",
		Subtotal:        req.Subtotal,
		TaxTotal:        req.TaxTotal,
		TotalAmount:     req.TotalAmount,
		InsuranceName:   req.InsuranceName,
		InsuranceRatio:  req.InsuranceRatio,
		InsuranceAmount: req.InsuranceAmount,
		DiscountAmount:  req.DiscountAmount,
		BillingAmount:   req.BillingAmount,
		PaymentMethod:   req.PaymentMethod,
		Memo:            req.Memo,
	}

	if err := s.accountingRepo.CreateAccounting(ctx, acc); err != nil {
		return nil, err
	}

	return acc, nil
}

// UpdateAccounting 会計を更新
func (s *Service) UpdateAccounting(ctx context.Context, id string, req *model.UpdateAccountingRequest) (*model.Accounting, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid accounting ID format")
	}

	acc, err := s.accountingRepo.GetAccountingByID(ctx, uid.String())
	if err != nil {
		return nil, err
	}

	if acc == nil {
		return nil, apperrors.WrapNotFound("accounting with id %s not found", id)
	}

	// Update fields
	if req.Status != "" {
		acc.Status = req.Status
	}
	if req.Subtotal != nil {
		acc.Subtotal = req.Subtotal
	}
	if req.TaxTotal != nil {
		acc.TaxTotal = req.TaxTotal
	}
	if req.TotalAmount != nil {
		acc.TotalAmount = req.TotalAmount
	}
	if req.InsuranceName != "" {
		acc.InsuranceName = req.InsuranceName
	}
	if req.InsuranceRatio != nil {
		acc.InsuranceRatio = req.InsuranceRatio
	}
	if req.InsuranceAmount != nil {
		acc.InsuranceAmount = req.InsuranceAmount
	}
	if req.DiscountAmount != nil {
		acc.DiscountAmount = req.DiscountAmount
	}
	if req.BillingAmount != nil {
		acc.BillingAmount = req.BillingAmount
	}
	if req.ReceivedAmount != nil {
		acc.ReceivedAmount = req.ReceivedAmount
	}
	if req.ChangeAmount != nil {
		acc.ChangeAmount = req.ChangeAmount
	}
	if req.PaymentMethod != "" {
		acc.PaymentMethod = req.PaymentMethod
	}
	if req.CompletedAt != nil {
		acc.CompletedAt = req.CompletedAt
	}
	if req.Memo != "" {
		acc.Memo = req.Memo
	}

	if err := s.accountingRepo.UpdateAccounting(ctx, acc); err != nil {
		return nil, err
	}

	return acc, nil
}

// DeleteAccounting 会計を削除
func (s *Service) DeleteAccounting(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return apperrors.WrapInvalidInput("invalid accounting ID format")
	}

	return s.accountingRepo.DeleteAccounting(ctx, uid.String())
}
