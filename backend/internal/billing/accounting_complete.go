package billing

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// CompleteAccountingItemInput は complete command の明細1行入力。
// FK 検証は CreateItemForComplete（ValidateCreateReferences）が tx 内で行う。
// 価格の master 再解決は既存 CreateItem と同型（source-linked 価格の server 上書きは別 packet）。
type CompleteAccountingItemInput struct {
	Category              string
	Name                  string
	UnitPrice             int64
	Quantity              float64
	DiscountRate          float64
	DiscountAmount        int64
	TaxType               string
	TaxRate               float64
	IsInsuranceApplicable bool
	Source                string
	OtherReason           *string
	MerchandiseItemID     *uint64
	TreatmentID           *uint64
	VaccinationID         *uint64
	ExamID                *uint64
	AppointmentID         *uint64
	TrimmingCourseID      *uint64
	TrimmingOptionID      *uint64
	SortOrder             int
}

// CompleteAccountingInput は POST /accountings/complete のサービス入力。
// clinic_id / actor は handler が認証 context から注入する。client total は受け取らない。
// held for EMR-462: 旧 command の形のまま退避している（wire は D12 で維持）。Complete が再配線されるまでは、入力の形を検証した後に固定エラーで拒否される。
type CompleteAccountingInput struct {
	ClinicID       uint64
	StaffID        *uint64
	IdempotencyKey string
	// BillingID は EMR-253: 既存 waiting billing の takeover 対象。
	// 指定時は新規 INSERT せず当該行を FOR UPDATE でロックして in-place 確定する。
	// waiting 以外の行・他スロットの行を指す場合は 409。
	BillingID         *uint64
	MedicalRecordID   *uint64
	HospitalizationID *uint64
	OwnerID           *uint64
	PetID             *uint64
	ScheduledDate     time.Time
	Memo              string
	HasInsurance      bool
	InsuranceRatio    *float64
	InsuranceName     *string
	InsuranceAmount   *int64
	DiscountAmount    *int64
	Items             []CompleteAccountingItemInput
	PaymentSplits     []PaymentSplitInput
	PostCloseReason   *string
	// ExpectedUnbilledRevision は EMR-196② の楽観ロック token。
	// PetID 指定時は必須（client が表示した unbilled-details の revision をそのまま返送）。
	// complete tx 内で再集計した集約版と不一致なら 409 UNBILLED_ITEMS_CHANGED。
	// digest（冪等 payload hash）には含めない — 版は操作内容ではなく前提条件であり、
	// 同一 key の retry が refetch 後も replay として扱われるようにするため。
	ExpectedUnbilledRevision string
	// IsPostClose は handler の候補 read（write 時に resolvePostCloseInTx で再評価）。
	IsPostClose bool
}

// CompleteAccountingResult は complete の結果。Created=true は 201、false は idempotent replay 200。
type CompleteAccountingResult struct {
	Accounting *model.Billing
	Created    bool
}

// AccountingCodeAlreadyCompleted は EMR-66 の二重確定 409 で返す安定エラーコード。
// medical_record_id / hospitalization_id のスロットに既存会計がある complete 確定に対応する。
const AccountingCodeAlreadyCompleted = "ACCOUNTING_ALREADY_COMPLETED"

// accountingAlreadyCompletedError は既に会計が存在するスロットへの complete を表す 409。
// Existing には既存会計を同梱し、handler が response body の accounting として返す。
// held for EMR-462: Complete が固定エラーで拒否する間は生成元が無い。handler の errors.As 分岐と handler test が参照するため、型と constructor を残す。
type accountingAlreadyCompletedError struct {
	appErr   *apperrors.AppError
	Existing *model.Billing
}

func (e *accountingAlreadyCompletedError) Error() string { return e.appErr.Error() }
func (e *accountingAlreadyCompletedError) Unwrap() error { return e.appErr }

func newAccountingAlreadyCompletedError(message string, existing *model.Billing) *accountingAlreadyCompletedError {
	return &accountingAlreadyCompletedError{
		appErr: &apperrors.AppError{
			Code:    AccountingCodeAlreadyCompleted,
			Message: message,
			Err:     apperrors.ErrConflict,
		},
		Existing: existing,
	}
}

// completeItemWriter は complete 用の ambient tx 明細 collaborator（WithTx を開始しない）。
// held for EMR-462: Complete は固定エラーで拒否するため現在は呼ばれない。composition の配線互換のためだけに残す。
type completeItemWriter interface {
	CreateItemForComplete(ctx context.Context, input *CreateBillingItemInput) (*model.BillingItem, error)
	// DeleteItemsForComplete は EMR-253: takeover 対象 waiting billing の既存明細を一括削除する。
	// request items が確定内容の権威のため、古い退院ケアプラン明細を残さない replace 戦略。
	DeleteItemsForComplete(ctx context.Context, clinicID, billingID uint64) error
}

// completeTotalsWriter は complete 用の合計算定 collaborator。billings のヘッダ合計列は廃止済みで、保存はしない。
// held for EMR-462: 算定結果の保存は snapshot writer が担う。Complete が固定エラーで拒否する間は呼ばれない。
type completeTotalsWriter interface {
	RecalculateTotalsForComplete(ctx context.Context, clinicID, billingID uint64) (subtotal, taxTotal, totalAmount int64, err error)
}

// WithCompleteItemWriter は complete 用の明細 collaborator を配線する。held for EMR-462（Complete 再配線まで未使用）。
func WithCompleteItemWriter(w completeItemWriter) accountingServiceOption {
	return func(s *accountingService) {
		s.itemWriter = w
	}
}

// WithCompleteTotalsWriter は complete 用の合計算定 collaborator を配線する。held for EMR-462（Complete 再配線まで未使用）。
func WithCompleteTotalsWriter(w completeTotalsWriter) accountingServiceOption {
	return func(s *accountingService) {
		s.totalsWriter = w
	}
}

// ValidateIdempotencyKeyUUID は Idempotency-Key が UUID であることを検証する。
func ValidateIdempotencyKeyUUID(key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return apperrors.WrapInvalidInput("Idempotency-Key header is required")
	}
	if _, err := uuid.Parse(key); err != nil {
		return apperrors.WrapInvalidInput("Idempotency-Key must be a valid UUID")
	}
	return nil
}

// Complete は POST /accountings/complete の確定 command。
// held for EMR-462: 旧実装は header・明細・合計・支払い・予約完了を単一 tx で書いていたが、置換先の保存
// （算定固定 snapshot・医師確認 set・受領と配賦・操作 receipt）の writer が未接続のため、書き込み前に固定エラー（422）で拒否する。
// header・明細・支払いのいずれも書かず、成功を装わない。入力の形（nil・Idempotency-Key・必須項目）の不正だけは従来どおり 400 を返す。
func (*accountingService) Complete(_ context.Context, input *CompleteAccountingInput) (*CompleteAccountingResult, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("complete input is required")
	}
	if err := ValidateIdempotencyKeyUUID(input.IdempotencyKey); err != nil {
		return nil, err
	}
	if input.ScheduledDate.IsZero() {
		return nil, apperrors.WrapInvalidInput("scheduled_date is required")
	}
	if len(input.Items) == 0 {
		return nil, apperrors.WrapInvalidInput("items are required")
	}
	return nil, apperrors.WrapFinanceConsumerPending("会計の確定は、算定固定・医師確認・受領の保存先が未接続のため利用できません")
}
