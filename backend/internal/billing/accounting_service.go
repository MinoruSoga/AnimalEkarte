package billing

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
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
	// ChangeOverride は #188: お釣り直接上書きモード。true の場合、レジ実機の誤差吸収のため
	// change == received - amount の整合検証を緩和する（received >= amount・change >= 0 の下限ガードは維持）。
	// このフラグは検証専用で DB には永続化しない（payment_splits に列なし）。再編集時は保存済み ChangeAmount から
	// FE が上書き状態を再導出する（AccountingDetailModel.restoreChangeOverride）。
	ChangeOverride bool
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

// CorrectCreditPaymentInput は確定済み会計のクレジット（カード）金額を確定後に訂正する入力DTO（#189）。
// レジ実機（カルテ非連動）でのカード打ち間違いを、理由・権限・監査付きで訂正する専用フロー。
// 訂正対象は Method で指定したカード系内訳1件のみ。現金は #188（お釣り上書き）の管轄で本フロー対象外。
type CorrectCreditPaymentInput struct {
	ClinicID  uint64
	BillingID uint64
	StaffID   *uint64
	// Method は訂正対象のカード系支払い手段（credit_card | electronic_money）。
	Method model.PaymentMethod
	// Amount は訂正後の金額（1円以上）。カード内訳は受領額・お釣りを持たない（0固定）ため金額のみ受け取る。
	Amount int64
	// Reason は訂正理由（必須）。監査ログに記録する。
	Reason string
	// Memo は補足メモ（任意）。監査ログに記録する。
	Memo string
	// IsPostClose はハンドラがレジ締め済み期間判定を注入する（#211 M-2）。
	// true でも訂正は拒否しない（認可は post-close-edit 権限としてルートで既にゲート済み）。
	// 監査ログに post_close フラグを記録し WarnContext で可視化する、可視化専用フラグ。
	IsPostClose bool
}

// ClinicDailySummary は拠点別日次集計結果 (#86 段階3 論点4=2 拠点別集計)。
type ClinicDailySummary struct {
	ClinicID uint64
	Summary  *DailySummaryResult
}

type AccountingService interface {
	List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, search string, page, limit int) ([]model.Billing, int64, error)
	// ListForClinics は複数医院の会計を横断検索する (#86 段階3)。clinicIDs はハンドラ層で所属検証済みであること。
	ListForClinics(ctx context.Context, clinicIDs []uint64, petID, ownerID *uint64, status, startDate, endDate *string, search string, page, limit int) ([]model.Billing, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Billing, error)
	// GetByIDForClinics は複数医院スコープで会計を1件取得する (#86 詳細画面拠点横断)。clinicIDs はハンドラ層で所属検証済みであること。
	GetByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Billing, error)
	Create(ctx context.Context, input *CreateAccountingInput) (*model.Billing, error)
	// Complete は BUG-018: header/items/totals/payments/splits/reservation/audit を単一 tx で確定する。
	Complete(ctx context.Context, input *CompleteAccountingInput) (*CompleteAccountingResult, error)
	Update(ctx context.Context, input *UpdateAccountingInput) (*model.Billing, error)
	// CorrectCreditPayment は確定済み会計のクレジット（カード）金額を確定後に訂正する（#189）。
	// 確定済み（status=completed）かつ対象カード内訳が存在する場合のみ許可し、理由・監査を必須とする。
	CorrectCreditPayment(ctx context.Context, input *CorrectCreditPaymentInput) (*model.Billing, error)
	// BUG-371 / #118: 論理削除（status=cancelled）。actorID で監査ログを記録する。
	Cancel(ctx context.Context, clinicID, id uint64, actorID *uint64) error
	// BUG-370 / #120: 未納者一覧（startDate〜endDate の BETWEEN）
	ListUnpaidByBilling(ctx context.Context, clinicID uint64, startDate, endDate string, page, limit int) ([]model.Billing, int64, error)
	ListUnpaidByOwner(ctx context.Context, clinicID uint64, startDate, endDate string, page, limit int) ([]UnpaidOwnerAggregate, int64, UnpaidSummary, error)
	// #182: 会計画面表示用の飼主未納残高
	GetOwnerUnpaidBalance(ctx context.Context, clinicID, ownerID uint64) (OwnerUnpaidBalance, error)
	// #114: 月次未納繰越集計（前月繰越・当月未払い・次月繰越）
	GetMonthlyUnpaidCarryover(ctx context.Context, clinicID uint64, year, month, page, limit int) ([]MonthlyUnpaidOwnerPet, int64, MonthlyUnpaidSummary, error)
	// BUG-368: レジ締め日次集計
	GetDailySummary(ctx context.Context, clinicID uint64, dateStr string) (*DailySummaryResult, error)
	// GetDailySummaryForClinics は複数医院の拠点別日次集計を返す (#86 段階3 論点4=2)。
	GetDailySummaryForClinics(ctx context.Context, clinicIDs []uint64, dateStr string) ([]ClinicDailySummary, error)
}

// unbilledWriteGuard は BUG-013 write-time fail-closed 用（blocking unbilled warning の再集計）。
type unbilledWriteGuard interface {
	AssertNoBlockingUnbilled(ctx context.Context, clinicID, petID uint64) error
}

type accountingService struct {
	repo              AccountingRepository
	medicalRecordRepo billingMedicalRecordLocker
	hospRepo          billingHospitalizationFinder
	reservationRepo   accountingReservationRepository
	tagSyncSvc        cpmTagSyncer
	transactor        Transactor
	auditTx           billingAuditTxLogger
	payMethodRepo     PaymentMethodMasterRepository
	// closeRepo は W-013 締め後訂正台帳（cash_register_close_adjustments）書込用。
	// WithCashRegisterCloseRepository で注入。IsPostClose 経路では必須（欠落は fail-closed）。
	closeRepo CashRegisterCloseRepository
	// unbilledGuard は BUG-013: pet に blocking unbilled warning がある会計作成を拒否する。
	unbilledGuard unbilledWriteGuard
	// itemWriter / totalsWriter は BUG-018 Complete の ambient-tx 明細・合計 collaborator。
	itemWriter   completeItemWriter
	totalsWriter completeTotalsWriter
}

type accountingReservationRepository interface {
	sharedkernel.OwnerPetLinkVerifier
	CompleteForAccounting(ctx context.Context, clinicID uint64, medicalRecordID, ownerID, petID *uint64, scheduledDate time.Time) (int64, error)
}

type accountingServiceOption func(*accountingService)

// WithCashRegisterCloseRepository は締め後編集時の append-only adjustment 書込に使う close repo を配線する（W-013）。
func WithCashRegisterCloseRepository(repo CashRegisterCloseRepository) accountingServiceOption {
	return func(s *accountingService) {
		s.closeRepo = repo
	}
}

// WithUnbilledWriteGuard は BUG-013 の write-time 再集計ガードを配線する。
func WithUnbilledWriteGuard(guard unbilledWriteGuard) accountingServiceOption {
	return func(s *accountingService) {
		s.unbilledGuard = guard
	}
}

// auditTx は tx 内監査（#211/BE-refactor.md R1-2 fail-closed）の記録経路。会計は金銭データのため、
// 論理削除監査（Cancel）・クレジット訂正監査（CorrectCreditPayment）・締め後編集監査
// （Update / AuditActionBillingPostCloseEdit）をすべて ambient tx に参加させ、監査書込の失敗が
// 本体の書込もロールバックするようにする（3経路とも fail-closed 化済み。旧 auditSvc 経路は撤去した）。
// medicalRecordRepo / hospRepo / reservationRepo は AUD-002 の関連 FK clinic 所有・相互整合検証用。
// closeRepo は W-013 の締め後 adjustment 台帳書込用（WithCashRegisterCloseRepository で注入）。
func NewAccountingService(
	repo AccountingRepository,
	medicalRecordRepo billingMedicalRecordLocker,
	hospRepo billingHospitalizationFinder,
	reservationRepo accountingReservationRepository,
	tagSyncSvc cpmTagSyncer,
	transactor Transactor,
	auditTx billingAuditTxLogger,
	payMethodRepo PaymentMethodMasterRepository,
	opts ...accountingServiceOption,
) AccountingService {
	s := &accountingService{
		repo:              repo,
		medicalRecordRepo: medicalRecordRepo,
		hospRepo:          hospRepo,
		reservationRepo:   reservationRepo,
		tagSyncSvc:        tagSyncSvc,
		transactor:        transactor,
		auditTx:           auditTx,
		payMethodRepo:     payMethodRepo,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}
