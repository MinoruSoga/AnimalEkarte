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
	StaffID           *uint64
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
	// Payment フィールド（会計完了時に同時 upsert）
	PaymentMethod   *model.PaymentMethod
	InsuranceRatio  *float64
	InsuranceName   *string
	InsuranceAmount *int64
	DiscountAmount  *int64
	BillingAmount   *int64
	ReceivedAmount  *int64
	ChangeAmount    *int64
}

// buildBillingUpdateFields は UpdateAccountingInput から nil でないフィールドのみ抽出する。
func buildBillingUpdateFields(input *UpdateAccountingInput) map[string]any {
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
	List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Billing, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Billing, error)
	Create(ctx context.Context, input *CreateAccountingInput) (*model.Billing, error)
	Update(ctx context.Context, input *UpdateAccountingInput) (*model.Billing, error)
	// BUG-371: 論理削除（status=cancelled）。ハード削除（旧 Delete）の代替
	Cancel(ctx context.Context, clinicID, id uint64) error
	// BUG-370: 月末未納者一覧
	ListUnpaidByBilling(ctx context.Context, clinicID uint64, baseDate string, page, limit int) ([]model.Billing, int64, error)
	ListUnpaidByOwner(ctx context.Context, clinicID uint64, baseDate string, page, limit int) ([]repository.UnpaidOwnerAggregate, int64, repository.UnpaidSummary, error)
}

type accountingService struct {
	repo repository.AccountingRepository
}

func NewAccountingService(repo repository.AccountingRepository) AccountingService {
	return &accountingService{repo: repo}
}

func (s *accountingService) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Billing, int64, error) {
	result, total, err := s.repo.FindAll(ctx, clinicID, petID, ownerID, status, startDate, endDate, page, limit)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to list accounting")
	}
	return result, total, nil
}

func (s *accountingService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Billing, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get accounting")
	}
	return result, nil
}

func (s *accountingService) Create(ctx context.Context, input *CreateAccountingInput) (*model.Billing, error) {
	if input.ScheduledDate.IsZero() {
		return nil, apperrors.WrapInvalidInput("scheduled_date is required")
	}
	// BUG-142: 金額バリデーション
	if input.TotalAmount < 0 {
		return nil, apperrors.WrapInvalidInput("金額は0以上で指定してください")
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
		return nil, apperrors.Wrap(err, "failed to create accounting")
	}
	slog.InfoContext(ctx, "accounting created",
		slog.Uint64("billing_id", billing.ID),
		slog.Uint64("clinic_id", input.ClinicID))
	return billing, nil
}

func (s *accountingService) Update(ctx context.Context, input *UpdateAccountingInput) (*model.Billing, error) {
	// BUG-142: 金額バリデーション
	if input.TotalAmount != nil && *input.TotalAmount < 0 {
		return nil, apperrors.WrapInvalidInput("金額は0以上で指定してください")
	}
	fields := buildBillingUpdateFields(input)
	if len(fields) == 0 && !hasPaymentFields(input) {
		return nil, apperrors.WrapInvalidInput("no fields to update")
	}

	// Billing 本体の更新
	if len(fields) > 0 {
		if _, err := s.repo.UpdateFields(ctx, input.ClinicID, input.ID, fields); err != nil {
			return nil, apperrors.Wrap(err, "failed to update accounting")
		}
	}

	// Payment upsert（支払フィールドが含まれている場合）
	if hasPaymentFields(input) {
		payment := buildPaymentFromInput(input)
		if err := s.repo.UpsertPayment(ctx, payment); err != nil {
			return nil, apperrors.Wrap(err, "failed to upsert payment")
		}
		slog.InfoContext(ctx, "payment upserted",
			slog.Uint64("billing_id", input.ID))
	}

	// 更新後のレコードを返す
	billing, err := s.repo.FindByID(ctx, input.ClinicID, input.ID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to reload accounting after update")
	}

	slog.InfoContext(ctx, "accounting updated",
		slog.Uint64("billing_id", billing.ID),
		slog.Uint64("clinic_id", input.ClinicID))
	return billing, nil
}

// hasPaymentFields は UpdateAccountingInput に Payment 関連フィールドが含まれているか判定する。
func hasPaymentFields(input *UpdateAccountingInput) bool {
	return input.PaymentMethod != nil ||
		input.InsuranceRatio != nil ||
		input.InsuranceAmount != nil ||
		input.BillingAmount != nil ||
		input.ReceivedAmount != nil ||
		input.ChangeAmount != nil ||
		input.DiscountAmount != nil
}

// buildPaymentFromInput は UpdateAccountingInput から Payment モデルを構築する。
func buildPaymentFromInput(input *UpdateAccountingInput) *model.Payment {
	p := &model.Payment{
		BillingID: input.ID,
		PaidBy:    input.StaffID,
	}
	if input.Subtotal != nil {
		p.Subtotal = int64(*input.Subtotal)
	}
	if input.TaxTotal != nil {
		p.TaxTotal = int64(*input.TaxTotal)
	}
	if input.TotalAmount != nil {
		p.TotalAmount = int64(*input.TotalAmount)
	}
	if input.InsuranceName != nil {
		p.InsuranceName = *input.InsuranceName
	}
	if input.InsuranceRatio != nil {
		p.InsuranceRatio = *input.InsuranceRatio
	}
	if input.InsuranceAmount != nil {
		p.InsuranceAmount = *input.InsuranceAmount
	}
	if input.DiscountAmount != nil {
		p.DiscountAmount = *input.DiscountAmount
	}
	if input.BillingAmount != nil {
		p.BillingAmount = *input.BillingAmount
	}
	if input.ReceivedAmount != nil {
		p.ReceivedAmount = *input.ReceivedAmount
	}
	if input.ChangeAmount != nil {
		p.ChangeAmount = *input.ChangeAmount
	}
	if input.PaymentMethod != nil {
		p.Method = *input.PaymentMethod
	}
	return p
}

// BUG-370: 月末未納者一覧（会計単位）
func (s *accountingService) ListUnpaidByBilling(ctx context.Context, clinicID uint64, baseDate string, page, limit int) ([]model.Billing, int64, error) {
	result, total, err := s.repo.FindUnpaidByBilling(ctx, clinicID, baseDate, page, limit)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to list unpaid billings")
	}
	return result, total, nil
}

// BUG-370: 月末未納者一覧（飼主単位集約）
func (s *accountingService) ListUnpaidByOwner(ctx context.Context, clinicID uint64, baseDate string, page, limit int) ([]repository.UnpaidOwnerAggregate, int64, repository.UnpaidSummary, error) {
	result, total, summary, err := s.repo.FindUnpaidByOwner(ctx, clinicID, baseDate, page, limit)
	if err != nil {
		return nil, 0, summary, apperrors.Wrap(err, "failed to list unpaid by owner")
	}
	return result, total, summary, nil
}

// Cancel は会計を論理削除（status=cancelled）する。
// BUG-371: ハード削除の代替。監査性のため物理削除しない。
func (s *accountingService) Cancel(ctx context.Context, clinicID, id uint64) error {
	// 既存値取得
	existing, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to find accounting for cancel")
	}
	// 既に cancelled 状態なら二重キャンセル防止（AC-12）
	if existing.Status == model.BillingStatusCancelled {
		return apperrors.WrapConflict("既にキャンセル済みの会計です")
	}

	fields := map[string]any{
		"status": model.BillingStatusCancelled,
	}
	if _, err := s.repo.UpdateFields(ctx, clinicID, id, fields); err != nil {
		return apperrors.Wrap(err, "failed to cancel accounting")
	}

	slog.InfoContext(ctx, "billing cancelled",
		slog.Uint64("billing_id", id),
		slog.Uint64("clinic_id", clinicID))

	return nil
}
