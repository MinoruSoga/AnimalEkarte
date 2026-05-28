package service

import (
	"context"
	"time"

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
	repo       repository.AccountingRepository
	tagSyncSvc LstepTagSyncService
}

func NewAccountingService(repo repository.AccountingRepository, tagSyncSvc LstepTagSyncService) AccountingService {
	return &accountingService{repo: repo, tagSyncSvc: tagSyncSvc}
}
