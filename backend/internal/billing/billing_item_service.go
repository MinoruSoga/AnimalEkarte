// Package billing provides billing item use cases.
package billing

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// --- DB column constants ---

const (
	colBillingItemUnitPrice             = "unit_price"
	colBillingItemQuantity              = "quantity"
	colBillingItemDiscountRate          = "discount_rate"
	colBillingItemDiscountAmount        = "discount_amount"
	colBillingItemTaxType               = "tax_type"
	colBillingItemTaxRate               = "tax_rate"
	colBillingItemIsInsuranceApplicable = "is_insurance_applicable"
	otherReasonMaxRunes                 = 500
)

// --- Input DTOs ---

// CreateBillingItemInput は billing_item 作成の入力DTO
type CreateBillingItemInput struct {
	ClinicID              uint64
	BillingID             uint64
	Category              string // Service 内で model.ItemCategory に変換
	Name                  string
	UnitPrice             int64
	Quantity              float64
	DiscountRate          float64
	DiscountAmount        int64
	TaxType               string // "" = デフォルト "excluded"
	TaxRate               float64
	IsInsuranceApplicable bool
	Source                string // "" = デフォルト "manual"
	OtherReason           *string
	CreatedBy             *uint64
	TreatmentID           *uint64
	VaccinationID         *uint64
	ExamID                *uint64
	AppointmentID         *uint64
	TrimmingCourseID      *uint64
	TrimmingOptionID      *uint64
	MerchandiseItemID     *uint64 // #81: 個別商品指定によるキャンペーンマッチング
	SortOrder             int
	// #115 / BUG-463: 締め後編集（handler がレジ締め済み判定を注入）
	StaffID         *uint64
	PostCloseReason *string
	IsPostClose     bool
}

// UpdateBillingItemInput は billing_item 更新の入力DTO（nil = 未指定）
type UpdateBillingItemInput struct {
	UnitPrice             *int64
	Quantity              *float64
	DiscountRate          *float64
	DiscountAmount        *int64
	TaxType               *model.TaxType
	TaxRate               *float64
	IsInsuranceApplicable *bool
	// #115 / BUG-463: 締め後編集（handler がレジ締め済み判定を注入）
	StaffID         *uint64
	PostCloseReason *string
	IsPostClose     bool
}

// DeleteBillingItemInput は明細削除の入力DTO。
// StaffID は vaccination claim 解放監査 actor（BUG-440）および締め後編集監査 actor。
// IsPostClose / PostCloseReason はレジ締め後削除ゲート（BUG-463 residual）。
// nil input は非締め後削除として扱う。
type DeleteBillingItemInput struct {
	StaffID         *uint64
	PostCloseReason *string
	IsPostClose     bool
}

func buildBillingItemUpdate(input *UpdateBillingItemInput) map[string]any {
	fields := make(map[string]any)
	if input.UnitPrice != nil {
		fields[colBillingItemUnitPrice] = *input.UnitPrice
	}
	if input.Quantity != nil {
		fields[colBillingItemQuantity] = *input.Quantity
	}
	if input.DiscountRate != nil {
		fields[colBillingItemDiscountRate] = *input.DiscountRate
	}
	if input.DiscountAmount != nil {
		fields[colBillingItemDiscountAmount] = *input.DiscountAmount
	}
	if input.TaxType != nil {
		fields[colBillingItemTaxType] = *input.TaxType
	}
	if input.TaxRate != nil {
		fields[colBillingItemTaxRate] = *input.TaxRate
	}
	if input.IsInsuranceApplicable != nil {
		fields[colBillingItemIsInsuranceApplicable] = *input.IsInsuranceApplicable
	}
	return fields
}

func validateCreateBillingItemInput(input *CreateBillingItemInput) error {
	if input == nil {
		return apperrors.WrapInvalidInput("請求明細は必須です")
	}
	if input.BillingID == 0 {
		return apperrors.WrapInvalidInput("請求IDは必須です")
	}
	if input.Name == "" {
		return apperrors.WrapInvalidInput("商品名は必須です")
	}
	if input.UnitPrice < 0 {
		return apperrors.WrapInvalidInput(sharedkernel.ErrMsgPriceZeroOrMore)
	}
	if input.Quantity <= 0 {
		return apperrors.WrapInvalidInput("数量は正の値である必要があります")
	}
	return nil
}

func resolveBillingItemDefaults(input *CreateBillingItemInput) (model.TaxType, float64, model.ItemSource, error) {
	// カテゴリバリデーション
	if err := validateItemCategory(input.Category); err != nil {
		return "", 0, "", apperrors.Wrap(err, "failed to validate item category")
	}

	// TaxType デフォルト設定とバリデーション
	taxType := model.TaxTypeExcluded
	if input.TaxType != "" {
		if err := sharedkernel.ValidateTaxType(input.TaxType); err != nil {
			return "", 0, "", apperrors.Wrap(err, "failed to validate tax type")
		}
		taxType = model.TaxType(input.TaxType)
	}

	// TaxRate デフォルト設定
	taxRate := sharedkernel.DefaultTaxRate
	if input.TaxRate > 0 {
		taxRate = input.TaxRate
	}

	// Source デフォルト設定とバリデーション
	source := model.ItemSourceManual
	if input.Source != "" {
		if err := validateItemSource(input.Source); err != nil {
			return "", 0, "", apperrors.Wrap(err, "failed to validate item source")
		}
		source = model.ItemSource(input.Source)
	}

	return taxType, taxRate, source, nil
}

func applyBillingItemOtherMetadata(input *CreateBillingItemInput, item *model.BillingItem) error {
	item.OtherReason = nil
	item.CreatedBy = nil
	if item.Source != model.ItemSourceManual ||
		item.Category != model.ItemCategoryOther ||
		input.MerchandiseItemID != nil {
		return nil
	}
	if input.OtherReason == nil {
		return apperrors.WrapInvalidInput("その他カテゴリの手入力明細は理由を入力してください")
	}
	trimmed := strings.TrimSpace(*input.OtherReason)
	if trimmed == "" {
		return apperrors.WrapInvalidInput("その他カテゴリの手入力明細は理由を入力してください")
	}
	if utf8.RuneCountInString(trimmed) > otherReasonMaxRunes {
		return apperrors.WrapInvalidInput("その他理由は500文字以内で入力してください")
	}
	if input.CreatedBy == nil || *input.CreatedBy == 0 {
		return apperrors.WrapInvalidInput("その他カテゴリの手入力明細は操作者が必要です")
	}
	item.OtherReason = &trimmed
	item.CreatedBy = input.CreatedBy
	return nil
}

// ---- BillingItemService ----

// BillingItemService は billing_items の CRUD とトータル再計算を担うインターフェース
type BillingItemService interface {
	CreateItem(ctx context.Context, input *CreateBillingItemInput) (*model.BillingItem, error)
	// CreateItemForComplete / RecalculateTotalsForComplete は BUG-018 Complete の ambient-tx collaborator。
	CreateItemForComplete(ctx context.Context, input *CreateBillingItemInput) (*model.BillingItem, error)
	RecalculateTotalsForComplete(ctx context.Context, clinicID, billingID uint64) (subtotal, taxTotal, totalAmount int64, err error)
	UpdateItem(ctx context.Context, clinicID, id uint64, input *UpdateBillingItemInput) (*model.BillingItem, error)
	// DeleteItem は明細を soft-delete する。
	// input.StaffID は vaccination claim 解放監査 actor（BUG-440）および締め後編集監査 actor。
	// input.IsPostClose 時は post_close_reason 必須 + 同 tx fail-closed 監査（BUG-463 residual）。
	// nil input は非締め後削除として扱う。
	DeleteItem(ctx context.Context, clinicID, id uint64, input *DeleteBillingItemInput) error
	GetUnbilledItems(ctx context.Context, clinicID, petID uint64) ([]model.BillingItem, error)
	// GetUnbilledItemDetails は未請求候補 items と typed blocking warnings を返す（BUG-013）。
	GetUnbilledItemDetails(ctx context.Context, clinicID, petID uint64) (*UnbilledDetails, error)
	// AssertNoBlockingUnbilled は pet に blocking unbilled warning がある場合 Conflict を返す（write-time fail-closed）。
	AssertNoBlockingUnbilled(ctx context.Context, clinicID, petID uint64) error
	// GetUngroupedSameDaySummary は同日同ペットの未会計対象化項目(診察/トリミング)の件数を返す(#77 取り残し警告)。
	GetUngroupedSameDaySummary(ctx context.Context, clinicID, petID uint64, date time.Time) (UngroupedSameDaySummary, error)
	// GetDiscountSuggestions は指定明細に適用可能な割引候補を返す（#81 Q-I スタッフ選択）。
	// campaignRepo 未配線の場合は飼主割引のみ。
	GetDiscountSuggestions(ctx context.Context, clinicID, itemID uint64) ([]DiscountSuggestion, error)
	// GetBilling は締め後編集判定用に請求ヘッダを返す（handler が ScheduledDate を参照）。
	GetBilling(ctx context.Context, clinicID, billingID uint64) (*model.Billing, error)
	// GetBillingForItem は明細 ID から親請求ヘッダを返す（締め後編集判定用）。
	GetBillingForItem(ctx context.Context, clinicID, itemID uint64) (*model.Billing, error)
}

// UngroupedSameDaySummary は #77 取り残し警告用の未会計対象化件数サマリ。
type UngroupedSameDaySummary struct {
	MedicalRecordCount int64
	TrimmingCount      int64
}

// Unbilled warning codes / sources (BUG-013). Keep payload minimal — no IDs, names, prices, SQL.
const (
	UnbilledWarningSourceVaccination               = "vaccination"
	UnbilledWarningCodeVaccinationMasterUnbillable = "vaccination_master_unbillable"
)

// UnbilledWarning は未請求集約の data-quality 警告（response 公開契約）。
type UnbilledWarning struct {
	Source   string `json:"source"`
	Code     string `json:"code"`
	Count    int    `json:"count"`
	Blocking bool   `json:"blocking"`
}

// UnbilledDetails は additive GET /billing-items/unbilled-details の結果。
type UnbilledDetails struct {
	Items    []model.BillingItem
	Warnings []UnbilledWarning
}

type billingItemService struct {
	repo               BillingItemRepository
	billingRepo        accountingBillingView
	treatmentRepo      treatmentBillingReader
	transactor         Transactor
	trimmingCourseRepo trimmingCourseFinder // X-4: クロステナント write 防止用の所有権検証
	trimmingOptionRepo trimmingOptionFinder // X-4: クロステナント write 防止用の所有権検証
	campaignRepo       CampaignRepository   // #81 段階2b: nil の場合は自動割引なし
	ownerRepo          billingOwnerReader   // #81 段階2b: 飼主割引取得用
	auditTx            billingAuditTxLogger // #115 / BUG-463: 締め後編集 fail-closed 監査
	// closeRepo は W-013 締め後明細変更の append-only adjustment 台帳用（任意 DI）。
	closeRepo CashRegisterCloseRepository
}

type unbilledTrimmingItemFinder interface {
	FindUnbilledTrimmingItemsByPetID(ctx context.Context, clinicID, petID uint64) ([]model.BillingItem, error)
}

type ungroupedTrimmingCounter interface {
	CountNonAccountingTrimmingByPetAndDate(ctx context.Context, clinicID, petID uint64, date time.Time) (int64, error)
}

// billingItemServiceOption は BillingItemService 構築時の任意依存注入。
type billingItemServiceOption func(*billingItemService)

// WithBillingItemAuditTx は締め後編集の fail-closed 監査 logger を配線する（#115 / BUG-463）。
func WithBillingItemAuditTx(auditTx billingAuditTxLogger) billingItemServiceOption {
	return func(s *billingItemService) {
		s.auditTx = auditTx
	}
}

// WithBillingItemCloseRepository は締め後明細変更時の cash_register_close_adjustments 追記に使う close repo を配線する（W-013）。
func WithBillingItemCloseRepository(repo CashRegisterCloseRepository) billingItemServiceOption {
	return func(s *billingItemService) {
		s.closeRepo = repo
	}
}

// NewBillingItemServiceWithCampaign は #81 段階2b: キャンペーン/飼主割引の自動適用を有効にした BillingItemService を返す。
func NewBillingItemServiceWithCampaign(repo BillingItemRepository, billingRepo accountingBillingView, treatmentRepo treatmentBillingReader, transactor Transactor, trimmingCourseRepo trimmingCourseFinder, trimmingOptionRepo trimmingOptionFinder, campaignRepo CampaignRepository, ownerRepo billingOwnerReader, opts ...billingItemServiceOption) BillingItemService {
	s := &billingItemService{repo: repo, billingRepo: billingRepo, treatmentRepo: treatmentRepo, transactor: transactor, trimmingCourseRepo: trimmingCourseRepo, trimmingOptionRepo: trimmingOptionRepo, campaignRepo: campaignRepo, ownerRepo: ownerRepo}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// rejectIfBillingFinalized は確定/取消済み会計への明細変更を Conflict で拒否する（BUG-463 / DeleteItem と同型）。
func rejectIfBillingFinalized(billing *model.Billing, action string) error {
	if billing == nil {
		return apperrors.WrapInternalServerError("locked billing is nil")
	}
	if billing.Status == model.BillingStatusCompleted ||
		billing.Status == model.BillingStatusCancelled {
		return apperrors.WrapConflict(fmt.Sprintf("確定済みまたは取消済みの会計明細は%sできません", action))
	}
	return nil
}

// rejectUnlessCompletedCorrectionAllowed は UpdateItem 用。
// cancelled は常に拒否。completed は post_close_reason 付きの訂正のみ許可（BUG-009）。
func rejectUnlessCompletedCorrectionAllowed(billing *model.Billing, action string, reason *string) error {
	if billing == nil {
		return apperrors.WrapInternalServerError("locked billing is nil")
	}
	if billing.Status == model.BillingStatusCancelled {
		return apperrors.WrapConflict(fmt.Sprintf("確定済みまたは取消済みの会計明細は%sできません", action))
	}
	if billing.Status == model.BillingStatusCompleted {
		if reason == nil || strings.TrimSpace(*reason) == "" {
			return apperrors.WrapInvalidInput("確定済み会計の明細修正には修正理由（post_close_reason）が必要です")
		}
		return nil
	}
	return nil
}

func requirePostCloseReason(isPostClose bool, reason *string) error {
	if isPostClose && (reason == nil || *reason == "") {
		return apperrors.WrapInvalidInput("レジ締め済み期間の会計編集には post_close_reason の入力が必要です")
	}
	return nil
}

// createItemAmbientOpts は CreateItem / Complete 共有の ambient tx 内明細作成オプション。
type createItemAmbientOpts struct {
	recalculate     bool // false のとき totals 再計算を呼び出し側に委ねる（Complete が最後に一括再計算）
	recordPostClose bool // false のとき締め後監査は command 単位で Complete が記録する
}
