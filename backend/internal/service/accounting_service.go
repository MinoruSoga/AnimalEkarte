package service

import (
	"context"
	"fmt"
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
	Subtotal          int64
	TaxTotal          int64
	TotalAmount       int64
	HasInsurance      bool
	Status            model.BillingStatus
	ScheduledDate     time.Time
	CompletedAt       *time.Time
	Memo              string
}

// PaymentSplitInput は支払い内訳1行の入力DTO（混在会計用）。
type PaymentSplitInput struct {
	Method          model.PaymentMethod
	PaymentMethodID *uint64
	Amount          int64
	ReceivedAmount  int64
	ChangeAmount    int64
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
	Subtotal          *int64
	TaxTotal          *int64
	TotalAmount       *int64
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
	// PaymentSplits: 混在支払い内訳（nil = 単一支払い、従来互換）
	PaymentSplits []PaymentSplitInput
}

// hasPaymentFields は UpdateAccountingInput に Payment 関連フィールドが含まれているか判定する。
func hasPaymentFields(input *UpdateAccountingInput) bool {
	return len(input.PaymentSplits) > 0 ||
		input.PaymentMethod != nil ||
		input.InsuranceRatio != nil ||
		input.InsuranceAmount != nil ||
		input.BillingAmount != nil ||
		input.ReceivedAmount != nil ||
		input.ChangeAmount != nil ||
		input.DiscountAmount != nil
}

// representativeMethod は splits から代表支払い手段を返す（legacy payments.method 用）。
// 優先順位: cash > credit_card > electronic_money
func representativeMethod(splits []PaymentSplitInput) model.PaymentMethod {
	for _, s := range splits {
		if s.Method == model.PaymentMethodCash {
			return model.PaymentMethodCash
		}
	}
	for _, s := range splits {
		if s.Method == model.PaymentMethodCreditCard {
			return model.PaymentMethodCreditCard
		}
	}
	return model.PaymentMethodElectronicMoney
}

// validatePaymentSplits は splits の整合性を検証する。
func validatePaymentSplits(splits []PaymentSplitInput, billingAmount *int64) error {
	if len(splits) == 0 {
		return nil
	}
	seen := make(map[model.PaymentMethod]bool, len(splits))
	var total int64
	for _, s := range splits {
		if seen[s.Method] {
			return apperrors.WrapInvalidInput(fmt.Sprintf("支払い手段 %s が重複しています", s.Method))
		}
		seen[s.Method] = true
		if s.Amount <= 0 {
			return apperrors.WrapInvalidInput("各支払い金額は1円以上でなければなりません")
		}
		total += s.Amount
		if s.Method == model.PaymentMethodCash {
			if s.ReceivedAmount < s.Amount {
				return apperrors.WrapInvalidInput("現金の預り金が不足しています")
			}
			if s.ChangeAmount != s.ReceivedAmount-s.Amount {
				return apperrors.WrapInvalidInput("お釣り計算が不正です")
			}
		}
	}
	if billingAmount != nil && total != *billingAmount {
		return apperrors.WrapInvalidInput(fmt.Sprintf("支払い内訳の合計（%d）が請求金額（%d）と一致しません", total, *billingAmount))
	}
	return nil
}

// buildPaymentFromInput は UpdateAccountingInput から Payment モデルを構築する。
// splits がある場合は代表支払い手段・受領額・お釣りを splits から導出する。
func buildPaymentFromInput(input *UpdateAccountingInput) *model.Payment {
	p := &model.Payment{
		BillingID: input.ID,
		PaidBy:    input.StaffID,
	}
	if input.Subtotal != nil {
		p.Subtotal = *input.Subtotal
	}
	if input.TaxTotal != nil {
		p.TaxTotal = *input.TaxTotal
	}
	if input.TotalAmount != nil {
		p.TotalAmount = *input.TotalAmount
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

	if len(input.PaymentSplits) > 0 {
		// splits から代表手段・受領額・お釣りを導出
		p.Method = representativeMethod(input.PaymentSplits)
		for _, s := range input.PaymentSplits {
			if s.Method == model.PaymentMethodCash {
				p.ReceivedAmount = s.ReceivedAmount
				p.ChangeAmount = s.ChangeAmount
				break
			}
		}
	} else {
		if input.ReceivedAmount != nil {
			p.ReceivedAmount = *input.ReceivedAmount
		}
		if input.ChangeAmount != nil {
			p.ChangeAmount = *input.ChangeAmount
		}
		if input.PaymentMethod != nil {
			p.Method = *input.PaymentMethod
		}
	}
	return p
}

// buildPaymentSplits は UpdateAccountingInput から PaymentSplit モデルのスライスを構築する。
// PaymentSplits が空の場合は単一支払いフィールドから1行を生成する（backward compat）。
func buildPaymentSplits(input *UpdateAccountingInput) []model.PaymentSplit {
	if len(input.PaymentSplits) > 0 {
		splits := make([]model.PaymentSplit, 0, len(input.PaymentSplits))
		for _, s := range input.PaymentSplits {
			splits = append(splits, model.PaymentSplit{
				ClinicID:        input.ClinicID,
				BillingID:       input.ID,
				Method:          s.Method,
				PaymentMethodID: s.PaymentMethodID,
				Amount:          s.Amount,
				ReceivedAmount:  s.ReceivedAmount,
				ChangeAmount:    s.ChangeAmount,
				PaidBy:          input.StaffID,
			})
		}
		return splits
	}
	// 単一支払い backward compat — BillingAmount が設定されている場合のみ生成
	if input.BillingAmount == nil || *input.BillingAmount <= 0 {
		return nil
	}
	method := model.PaymentMethodCash
	if input.PaymentMethod != nil {
		method = *input.PaymentMethod
	}
	var received, change int64
	if input.ReceivedAmount != nil {
		received = *input.ReceivedAmount
	}
	if input.ChangeAmount != nil {
		change = *input.ChangeAmount
	}
	return []model.PaymentSplit{
		{
			ClinicID:       input.ClinicID,
			BillingID:      input.ID,
			Method:         method,
			Amount:         *input.BillingAmount,
			ReceivedAmount: received,
			ChangeAmount:   change,
			PaidBy:         input.StaffID,
		},
	}
}

// buildAccountingUpdate は UpdateAccountingInput から nil でないフィールドのみ抽出する。
func buildAccountingUpdate(input *UpdateAccountingInput) map[string]any {
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
	// BUG-368: レジ締め日次集計
	GetDailySummary(ctx context.Context, clinicID uint64, dateStr string) (*repository.DailySummaryResult, error)
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
		slog.ErrorContext(ctx, "failed to list accounting", "error", err)
		return nil, 0, apperrors.Wrap(err, "failed to list accounting")
	}
	return result, total, nil
}

func (s *accountingService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Billing, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get accounting", "error", err)
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
		return nil, apperrors.WrapInvalidInput(ErrMsgPriceZeroOrMore)
	}
	if input.Subtotal+input.TaxTotal != input.TotalAmount {
		return nil, apperrors.WrapInvalidInput("小計と税額の合計が請求合計と一致しません")
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
		slog.ErrorContext(ctx, "failed to create accounting", "error", err)
		return nil, apperrors.Wrap(err, "failed to create accounting")
	}
	slog.InfoContext(ctx, "accounting created",
		slog.Uint64("billing_id", billing.ID),
		slog.Uint64("clinic_id", input.ClinicID))
	return billing, nil
}

func (s *accountingService) Update(ctx context.Context, input *UpdateAccountingInput) (*model.Billing, error) {
	if _, err := s.repo.FindByID(ctx, input.ClinicID, input.ID); err != nil {
		slog.ErrorContext(ctx, "failed to find accounting", "error", err)
		return nil, apperrors.Wrap(err, "failed to find accounting")
	}
	// BUG-142: 金額バリデーション
	if input.TotalAmount != nil && *input.TotalAmount < 0 {
		return nil, apperrors.WrapInvalidInput(ErrMsgPriceZeroOrMore)
	}
	// 混在会計バリデーション
	if err := validatePaymentSplits(input.PaymentSplits, input.BillingAmount); err != nil {
		return nil, err
	}
	fields := buildAccountingUpdate(input)
	if len(fields) == 0 && !hasPaymentFields(input) {
		return nil, apperrors.WrapInvalidInput("no fields to update")
	}

	// Billing 本体の更新
	if len(fields) > 0 {
		if _, err := s.repo.Update(ctx, input.ClinicID, input.ID, fields); err != nil {
			slog.ErrorContext(ctx, "failed to update accounting", "error", err)
			return nil, apperrors.Wrap(err, "failed to update accounting")
		}
	}

	// Payment upsert（支払フィールドが含まれている場合）
	if hasPaymentFields(input) {
		payment := buildPaymentFromInput(input)
		if err := s.repo.SavePayment(ctx, payment); err != nil {
			slog.ErrorContext(ctx, "failed to upsert payment", "error", err)
			return nil, apperrors.Wrap(err, "failed to upsert payment")
		}
		// payment_splits の更新（混在会計・backward compat 両対応）
		splits := buildPaymentSplits(input)
		if err := s.repo.SavePaymentSplits(ctx, splits); err != nil {
			slog.ErrorContext(ctx, "failed to save payment splits", "error", err)
			return nil, apperrors.Wrap(err, "failed to save payment splits")
		}
		slog.InfoContext(ctx, "payment upserted",
			slog.Uint64("clinic_id", input.ClinicID),
			slog.Uint64("billing_id", input.ID))
	}

	// 更新後のレコードを返す
	accounting, err := s.repo.FindByID(ctx, input.ClinicID, input.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to reload accounting after update", "error", err)
		return nil, apperrors.Wrap(err, "failed to reload accounting after update")
	}

	slog.InfoContext(ctx, "accounting updated",
		slog.Uint64("billing_id", accounting.ID),
		slog.Uint64("clinic_id", input.ClinicID))
	return accounting, nil
}

// BUG-370: 月末未納者一覧（会計単位）
func (s *accountingService) ListUnpaidByBilling(ctx context.Context, clinicID uint64, baseDate string, page, limit int) ([]model.Billing, int64, error) {
	result, total, err := s.repo.FindUnpaidByBilling(ctx, clinicID, baseDate, page, limit)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list unpaid billings", "error", err)
		return nil, 0, apperrors.Wrap(err, "failed to list unpaid billings")
	}
	return result, total, nil
}

// BUG-370: 月末未納者一覧（飼主単位集約）
func (s *accountingService) ListUnpaidByOwner(ctx context.Context, clinicID uint64, baseDate string, page, limit int) ([]repository.UnpaidOwnerAggregate, int64, repository.UnpaidSummary, error) {
	result, total, summary, err := s.repo.FindUnpaidByOwner(ctx, clinicID, baseDate, page, limit)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list unpaid by owner", "error", err)
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
	if _, err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
		return apperrors.Wrap(err, "failed to cancel accounting")
	}

	slog.InfoContext(ctx, "billing cancelled",
		slog.Uint64("billing_id", id),
		slog.Uint64("clinic_id", clinicID))

	return nil
}

// GetDailySummary は指定日のレジ締め集計を返す。BUG-368
func (s *accountingService) GetDailySummary(ctx context.Context, clinicID uint64, dateStr string) (*repository.DailySummaryResult, error) {
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	date, err := time.ParseInLocation("2006-01-02", dateStr, jst)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("date must be YYYY-MM-DD")
	}
	result, err := s.repo.GetDailySummary(ctx, clinicID, date)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get daily summary", "error", err)
		return nil, apperrors.Wrap(err, "failed to get daily summary")
	}
	return result, nil
}
