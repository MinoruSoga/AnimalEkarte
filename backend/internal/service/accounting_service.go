package service

import (
	"context"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// CreateAccountingInput は会計作成のサービス入力DTO。
type CreateAccountingInput struct {
	ClinicID          uint64
	MedicalRecordID   *uint64
	HospitalizationID *uint64
	OwnerID           *uint64
	PetID             *uint64
	Subtotal          int
	TaxTotal          int
	TotalAmount       int
	HasInsurance      bool
	Status            model.BillingStatus
	ScheduledDate     time.Time
	CompletedAt       *time.Time
	Memo              string
}

// UpdateAccountingInput は会計更新のサービス入力DTO。
// nil のフィールドは更新しない（GORM ゼロ値スキップ問題を回避）。
type UpdateAccountingInput struct {
	ID                uint64
	ClinicID          uint64
	MedicalRecordID   *uint64
	HospitalizationID *uint64
	OwnerID           *uint64
	PetID             *uint64
	Subtotal          *int
	TaxTotal          *int
	TotalAmount       *int
	HasInsurance      *bool
	Status            *model.BillingStatus
	ScheduledDate     *time.Time
	CompletedAt       *time.Time
	Memo              *string
}

// buildBillingUpdateFields は UpdateAccountingInput から nil でないフィールドのみ抽出する。
func buildBillingUpdateFields(input UpdateAccountingInput) map[string]any {
	fields := make(map[string]any)
	if input.MedicalRecordID != nil {
		fields["medical_record_id"] = *input.MedicalRecordID
	}
	if input.HospitalizationID != nil {
		fields["hospitalization_id"] = *input.HospitalizationID
	}
	if input.OwnerID != nil {
		fields["owner_id"] = *input.OwnerID
	}
	if input.PetID != nil {
		fields["pet_id"] = *input.PetID
	}
	if input.Subtotal != nil {
		fields["subtotal"] = *input.Subtotal
	}
	if input.TaxTotal != nil {
		fields["tax_total"] = *input.TaxTotal
	}
	if input.TotalAmount != nil {
		fields["total_amount"] = *input.TotalAmount
	}
	if input.HasInsurance != nil {
		fields["has_insurance"] = *input.HasInsurance
	}
	if input.Status != nil {
		fields["status"] = *input.Status
	}
	if input.ScheduledDate != nil {
		fields["scheduled_date"] = *input.ScheduledDate
	}
	if input.CompletedAt != nil {
		fields["completed_at"] = *input.CompletedAt
	}
	if input.Memo != nil {
		fields["memo"] = *input.Memo
	}
	return fields
}

type AccountingService interface {
	List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status *string, startDate, endDate *string, page, limit int) ([]model.Billing, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Billing, error)
	Create(ctx context.Context, input CreateAccountingInput) (*model.Billing, error)
	Update(ctx context.Context, input UpdateAccountingInput) (*model.Billing, error)
	Delete(ctx context.Context, clinicID, id uint64) error
}

type accountingService struct {
	repo repository.AccountingRepository
}

func NewAccountingService(repo repository.AccountingRepository) AccountingService {
	return &accountingService{repo: repo}
}

func (s *accountingService) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status *string, startDate, endDate *string, page, limit int) ([]model.Billing, int64, error) {
	return s.repo.FindAll(ctx, clinicID, petID, ownerID, status, startDate, endDate, page, limit)
}

func (s *accountingService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Billing, error) {
	return s.repo.FindByID(ctx, clinicID, id)
}

func (s *accountingService) Create(ctx context.Context, input CreateAccountingInput) (*model.Billing, error) {
	if input.ScheduledDate.IsZero() {
		return nil, apperrors.WrapInvalidInput("scheduled_date is required")
	}
	billing := &model.Billing{
		ClinicID:          input.ClinicID,
		MedicalRecordID:   input.MedicalRecordID,
		HospitalizationID: input.HospitalizationID,
		OwnerID:           input.OwnerID,
		PetID:             input.PetID,
		Subtotal:          input.Subtotal,
		TaxTotal:          input.TaxTotal,
		TotalAmount:       input.TotalAmount,
		HasInsurance:      input.HasInsurance,
		Status:            input.Status,
		ScheduledDate:     input.ScheduledDate,
		CompletedAt:       input.CompletedAt,
		Memo:              input.Memo,
	}
	if err := s.repo.Create(ctx, input.ClinicID, billing); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "accounting created",
		slog.Uint64("billing_id", billing.ID),
		slog.Uint64("clinic_id", input.ClinicID))
	return billing, nil
}

func (s *accountingService) Update(ctx context.Context, input UpdateAccountingInput) (*model.Billing, error) {
	fields := buildBillingUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("no fields to update")
	}
	billing, err := s.repo.UpdateFields(ctx, input.ClinicID, input.ID, fields)
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "accounting updated",
		slog.Uint64("billing_id", billing.ID),
		slog.Uint64("clinic_id", input.ClinicID))
	return billing, nil
}

func (s *accountingService) Delete(ctx context.Context, clinicID, id uint64) error {
	return s.repo.Delete(ctx, clinicID, id)
}
