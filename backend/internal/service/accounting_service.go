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
	// #115: 締め後編集フィールド
	PostCloseReason *string // 締め後編集理由（締め済み期間に編集する場合は必須）
	IsPostClose     bool    // ハンドラがレジ締め済み判定を注入する
}

// ClinicDailySummary は拠点別日次集計結果 (#86 段階3 論点4=2 拠点別集計)。
type ClinicDailySummary struct {
	ClinicID uint64
	Summary  *repository.DailySummaryResult
}

type AccountingService interface {
	List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Billing, int64, error)
	// ListForClinics は複数医院の会計を横断検索する (#86 段階3)。clinicIDs はハンドラ層で所属検証済みであること。
	ListForClinics(ctx context.Context, clinicIDs []uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Billing, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Billing, error)
	// GetByIDForClinics は複数医院スコープで会計を1件取得する (#86 詳細画面拠点横断)。clinicIDs はハンドラ層で所属検証済みであること。
	GetByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Billing, error)
	Create(ctx context.Context, input *CreateAccountingInput) (*model.Billing, error)
	Update(ctx context.Context, input *UpdateAccountingInput) (*model.Billing, error)
	// BUG-371 / #118: 論理削除（status=cancelled）。actorID で監査ログを記録する。
	Cancel(ctx context.Context, clinicID, id uint64, actorID *uint64) error
	// BUG-370 / #120: 未納者一覧（startDate〜endDate の BETWEEN）
	ListUnpaidByBilling(ctx context.Context, clinicID uint64, startDate, endDate string, page, limit int) ([]model.Billing, int64, error)
	ListUnpaidByOwner(ctx context.Context, clinicID uint64, startDate, endDate string, page, limit int) ([]repository.UnpaidOwnerAggregate, int64, repository.UnpaidSummary, error)
	// BUG-368: レジ締め日次集計
	GetDailySummary(ctx context.Context, clinicID uint64, dateStr string) (*repository.DailySummaryResult, error)
	// GetDailySummaryForClinics は複数医院の拠点別日次集計を返す (#86 段階3 論点4=2)。
	GetDailySummaryForClinics(ctx context.Context, clinicIDs []uint64, dateStr string) ([]ClinicDailySummary, error)
}

type accountingService struct {
	repo       repository.AccountingRepository
	tagSyncSvc LstepTagSyncService
	transactor repository.Transactor
	auditSvc   AuditService
}

func NewAccountingService(repo repository.AccountingRepository, tagSyncSvc LstepTagSyncService, transactor repository.Transactor, auditSvc AuditService) AccountingService {
	return &accountingService{repo: repo, tagSyncSvc: tagSyncSvc, transactor: transactor, auditSvc: auditSvc}
}
